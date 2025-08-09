package smartpromptopt

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/maximhq/bifrost/core/schemas"
)

const (
	PluginName = "smart_prompt_opt"
)

// SmartPromptOptPlugin implements input optimization and optional Pinecone-backed semantic compression.
// Focus: prompt improvement and cost reduction via PreHook.
// PostHook is minimal (metrics hook point).
type SmartPromptOptPlugin struct {
	cfg       Config
	pc        *PineconeSDKClient // optional, nil if disabled
	logger    schemas.Logger     // optional, may be nil
	estimator TokenEstimator

	// ring buffer for last N debug decisions (optional)
	ring *RingBuffer[string]
}

func NewSmartPromptOpt(cfg Config) (*SmartPromptOptPlugin, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid smart_prompt_opt config: %w", err)
	}

	var pc *PineconeSDKClient
	if cfg.Pinecone.Enabled {
		var err error
		pc, err = NewPineconeSDKClient(cfg.Pinecone)
		if err != nil {
			return nil, fmt.Errorf("failed to create pinecone client: %w", err)
		}
	}

	p := &SmartPromptOptPlugin{
		cfg:       cfg,
		pc:        pc,
		logger:    cfg.Logger,
		estimator: NewDefaultEstimator(0), // Use default chars-per-token ratio
	}
	if cfg.AdvancedLogging.RingSize > 0 {
		p.ring = NewRingBuffer[string](cfg.AdvancedLogging.RingSize)
	}
	return p, nil
}

func (p *SmartPromptOptPlugin) GetName() string { return PluginName }

func (p *SmartPromptOptPlugin) PreHook(ctx *context.Context, req *schemas.BifrostRequest) (*schemas.BifrostRequest, *schemas.PluginShortCircuit, error) {
	if !p.cfg.Enabled || req == nil {
		return req, nil, nil
	}

	start := time.Now()
	corrID := getCorrelationID(*ctx)
	log := p.logf
	log(schemas.LogLevelDebug, corrID, "prehook_start provider=%s model=%s", req.Provider, req.Model)
	p.debugRecord("start:" + req.Model)

	// 1) Normalize input
	normSummary := p.normalizeRequest(req)
	log(schemas.LogLevelDebug, corrID, "normalized %s", normSummary)

	// validate empty after normalization
	if isEmptyRequest(req) {
		msg := "empty input after normalization"
		log(schemas.LogLevelWarn, corrID, msg)
		status := 400
		return req, &schemas.PluginShortCircuit{Error: &schemas.BifrostError{StatusCode: &status, Error: schemas.ErrorField{Message: msg}}}, nil
	}

	// 2) Strengthen instructions
	if p.cfg.StrengthenInstructions {
		applied := p.applyInstructionPrefix(req)
		log(schemas.LogLevelDebug, corrID, "instruction_prefix_applied=%t", applied)
	}

	// 3) Apply provider defaults if missing
	p.applyProviderDefaults(req)

	// 4) Budget enforcement and optional Pinecone retrieval/compression
	estTokens := p.estimateTokens(req)
	log(schemas.LogLevelInfo, corrID, "estimated_tokens=%d budget=%d", estTokens, p.cfg.TokenBudget)

	if p.cfg.TokenBudget > 0 && estTokens > p.cfg.TokenBudget {
		if p.cfg.SemanticCompression.Enabled && p.pc != nil {
			ctxTimeout, cancel := context.WithTimeout(*ctx, time.Duration(p.cfg.Pinecone.TimeoutMs)*time.Millisecond)
			defer cancel()
			sum, meta := p.compressWithPinecone(ctxTimeout, req)
			if sum != "" {
				p.injectSummary(req, sum, meta)
				log(schemas.LogLevelInfo, corrID, "compression_applied via pinecone length=%d", len(sum))
			} else {
				log(schemas.LogLevelWarn, corrID, "compression_skipped_or_failed")
			}
		} else {
			trimmed := p.trimHeuristically(req)
			log(schemas.LogLevelInfo, corrID, "trim_applied=%t", trimmed)
		}
	}

	lat := time.Since(start)
	log(schemas.LogLevelDebug, corrID, "prehook_end latency_ms=%d", lat.Milliseconds())
	return req, nil, nil
}

func (p *SmartPromptOptPlugin) PostHook(ctx *context.Context, result *schemas.BifrostResponse, err *schemas.BifrostError) (*schemas.BifrostResponse, *schemas.BifrostError, error) {
	// Minimal for now. Could log usage for adaptive budgets in future.
	return result, err, nil
}

func (p *SmartPromptOptPlugin) Cleanup() error { return nil }

// ===== Helpers =====

func (p *SmartPromptOptPlugin) logf(level schemas.LogLevel, corrID string, format string, args ...any) {
	if p.logger == nil {
		return
	}
	msg := fmt.Sprintf("[%s] "+format, append([]any{corrID}, args...)...)
	switch level {
	case schemas.LogLevelDebug:
		p.logger.Debug(msg)
	case schemas.LogLevelInfo:
		p.logger.Info(msg)
	case schemas.LogLevelWarn:
		p.logger.Warn(msg)
	case schemas.LogLevelError:
		p.logger.Error(fmt.Errorf("%s", msg))
	}
}

func (p *SmartPromptOptPlugin) debugRecord(s string) {
	if p.ring != nil {
		p.ring.Add(s)
	}
}

func (p *SmartPromptOptPlugin) normalizeRequest(req *schemas.BifrostRequest) string {
	summary := []string{}
	// Text completion
	if req.Input.TextCompletionInput != nil {
		t := strings.TrimSpace(*req.Input.TextCompletionInput)
		t = collapseWhitespace(t)
		t = dedupeParagraphs(t)
		req.Input.TextCompletionInput = &t
		summary = append(summary, "text")
	}
	// Chat messages
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		for i := range msgs {
			// Handle content normalization based on type
			if msgs[i].Content.ContentStr != nil {
				// Normalize string content
				normalized := strings.TrimSpace(collapseWhitespace(*msgs[i].Content.ContentStr))
				msgs[i].Content.ContentStr = &normalized
			} else if msgs[i].Content.ContentBlocks != nil {
				// Normalize text blocks
				for j, block := range *msgs[i].Content.ContentBlocks {
					if block.Text != nil {
						normalized := strings.TrimSpace(collapseWhitespace(*block.Text))
						(*msgs[i].Content.ContentBlocks)[j].Text = &normalized
					}
				}
			}
		}
		// Trim history
		if p.cfg.MaxHistoryTurns > 0 && len(msgs) > p.cfg.MaxHistoryTurns {
			msgs = msgs[len(msgs)-p.cfg.MaxHistoryTurns:]
			summary = append(summary, "history_trim")
		}
		req.Input.ChatCompletionInput = &msgs
		if len(msgs) > 0 {
			summary = append(summary, "chat")
		}
	}
	return strings.Join(summary, ",")
}

func (p *SmartPromptOptPlugin) applyInstructionPrefix(req *schemas.BifrostRequest) bool {
	prefix := strings.TrimSpace(p.cfg.InstructionPrefix)
	if prefix == "" {
		return false
	}
	// Avoid double-applying if already present at start of first system/user message
	already := false
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		if len(msgs) > 0 {
			// Check if first message already has the prefix
			if msgs[0].Content.ContentStr != nil && strings.HasPrefix(*msgs[0].Content.ContentStr, prefix) {
				already = true
			}
		}
		if !already {
			sys := schemas.BifrostMessage{
				Role: schemas.ModelChatMessageRoleSystem,
				Content: schemas.MessageContent{
					ContentStr: &prefix,
				},
			}
			msgs = append([]schemas.BifrostMessage{sys}, msgs...)
			req.Input.ChatCompletionInput = &msgs
			return true
		}
	} else if req.Input.TextCompletionInput != nil {
		t := *req.Input.TextCompletionInput
		if !strings.HasPrefix(t, prefix) {
			t = prefix + "\n\n" + t
			req.Input.TextCompletionInput = &t
			return true
		}
	}
	return false
}

func (p *SmartPromptOptPlugin) applyProviderDefaults(req *schemas.BifrostRequest) {
	if req.Params == nil {
		req.Params = &schemas.ModelParameters{}
	}
	md, ok := p.cfg.ProviderDefaults[string(req.Provider)+":"+req.Model]
	if !ok {
		return
	}
	if req.Params.Temperature == nil && md.Temperature != nil {
		req.Params.Temperature = md.Temperature
	}
	if req.Params.TopP == nil && md.TopP != nil {
		req.Params.TopP = md.TopP
	}
}

func (p *SmartPromptOptPlugin) estimateTokens(req *schemas.BifrostRequest) int {
	return p.estimator.Estimate(req)
}

func (p *SmartPromptOptPlugin) compressWithPinecone(ctx context.Context, req *schemas.BifrostRequest) (summary string, meta map[string]any) {
	if p.pc == nil {
		return "", nil
	}

	// Collect all text from the request
	text := collectRequestText(req)
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil
	}

	chunks := ChunkByRune(text, p.cfg.SemanticCompression.ChunkSize)
	if p.cfg.Pinecone.UpsertOnLargeInputs {
		_ = p.pc.UpsertTexts(ctx, chunks, map[string]string{"source": "smart_prompt_opt"})
	}
	q, err := p.pc.QueryByText(ctx, text, p.cfg.Pinecone.TopK)
	if err != nil {
		return "", nil
	}
	if len(q) == 0 {
		return "", nil
	}
	// Summarize matches as compact bullets
	b := &strings.Builder{}
	b.WriteString("Context Summary (retrieved):\n")
	for i, m := range q {
		b.WriteString(fmt.Sprintf("- [%d] %s\n", i+1, Truncate(m.Text, 400)))
	}
	return b.String(), map[string]any{"hits": len(q)}
}

func (p *SmartPromptOptPlugin) injectSummary(req *schemas.BifrostRequest, summary string, meta map[string]any) {
	msg := schemas.BifrostMessage{
		Role: schemas.ModelChatMessageRoleSystem,
		Content: schemas.MessageContent{
			ContentStr: &summary,
		},
	}
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		// Create a new slice with the summary message first
		newMsgs := append([]schemas.BifrostMessage{msg}, msgs...)
		req.Input.ChatCompletionInput = &newMsgs
		return
	}
	// convert text to chat form when necessary
	if req.Input.TextCompletionInput != nil {
		t := *req.Input.TextCompletionInput
		userMsg := schemas.BifrostMessage{
			Role: schemas.ModelChatMessageRoleUser,
			Content: schemas.MessageContent{
				ContentStr: &t,
			},
		}
		msgs := []schemas.BifrostMessage{msg, userMsg}
		req.Input.ChatCompletionInput = &msgs
		req.Input.TextCompletionInput = nil
	}
}

func (p *SmartPromptOptPlugin) trimHeuristically(req *schemas.BifrostRequest) bool {
	trimmed := false
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		if len(msgs) > p.cfg.MaxHistoryTurns && p.cfg.MaxHistoryTurns > 0 {
			msgs = msgs[len(msgs)-p.cfg.MaxHistoryTurns:]
			trimmed = true
		}
		req.Input.ChatCompletionInput = &msgs
	}
	return trimmed
}

// Utility functions

func collapseWhitespace(s string) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(strings.TrimSpace(s), " ")
}

func dedupeParagraphs(s string) string {
	seen := map[string]struct{}{}
	parts := strings.Split(s, "\n\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if pp == "" {
			continue
		}
		if _, ok := seen[pp]; ok {
			continue
		}
		seen[pp] = struct{}{}
		out = append(out, pp)
	}
	return strings.Join(out, "\n\n")
}

func isEmptyRequest(req *schemas.BifrostRequest) bool {
	if req == nil {
		return true
	}
	if req.Input.TextCompletionInput != nil {
		if strings.TrimSpace(*req.Input.TextCompletionInput) == "" {
			return true
		}
	}
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		if len(msgs) == 0 {
			return true
		}
		allEmpty := true
		for _, m := range msgs {
			// Check if message has non-empty content
			if m.Content.ContentStr != nil && strings.TrimSpace(*m.Content.ContentStr) != "" {
				allEmpty = false
				break
			} else if m.Content.ContentBlocks != nil {
				// Check if any text block has content
				for _, block := range *m.Content.ContentBlocks {
					if block.Text != nil && strings.TrimSpace(*block.Text) != "" {
						allEmpty = false
						break
					}
				}
				if !allEmpty {
					break
				}
			}
		}
		return allEmpty
	}
	return false
}

func getCorrelationID(ctx context.Context) string {
	// placeholder: if context has a value, derive from it; else timestamp
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
