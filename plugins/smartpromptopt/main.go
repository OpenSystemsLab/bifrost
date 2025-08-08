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
	cfg      Config
	pc       PineconeClient // optional, nil if disabled
	logger   schemas.Logger // optional, may be nil
	estimator TokenEstimator

	// ring buffer for last N debug decisions (optional)
	ring *RingBuffer[string]
}

func NewSmartPromptOpt(cfg Config) (*SmartPromptOptPlugin, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid smart_prompt_opt config: %w", err)
	}

	var pc PineconeClient
	if cfg.Pinecone.Enabled {
		pc = NewPineconeSDKClient(cfg.Pinecone)
	}

	p := &SmartPromptOptPlugin{
		cfg:       cfg,
		pc:        pc,
		logger:    cfg.Logger,
		estimator: DefaultEstimator{},
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
		p.logger.Error(fmt.Errorf(msg))
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
			msgs[i].Content = strings.TrimSpace(collapseWhitespace(msgs[i].Content))
		}
		// Trim history
		if p.cfg.MaxHistoryTurns > 0 && len(msgs) > p.cfg.MaxHistoryTurns {
			msgs = msgs[len(msgs)-p.cfg.MaxHistoryTurns:]
			summary = append(summary, "history_trim")
		}
		req.Input.ChatCompletionInput = &msgs
		if len(msgs) > 0 { summary = append(summary, "chat") }
	}
	return strings.Join(summary, ",")
}

func (p *SmartPromptOptPlugin) applyInstructionPrefix(req *schemas.BifrostRequest) bool {
	prefix := strings.TrimSpace(p.cfg.InstructionPrefix)
	if prefix == "" { return false }
	// Avoid double-applying if already present at start of first system/user message
	already := false
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		if len(msgs) > 0 {
			first := msgs[0].Content
			if strings.HasPrefix(first, prefix) { already = true }
		}
		if !already {
			sys := schemas.BifrostMessage{Role: schemas.ModelChatMessageRoleSystem, Content: prefix}
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
	if !ok { return }
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
	if p.pc == nil { return "", nil }
	// Collect text
	var text string
	if req.Input.TextCompletionInput != nil {
		text = *req.Input.TextCompletionInput
	}
	if req.Input.ChatCompletionInput != nil {
		for _, m := range *req.Input.ChatCompletionInput {
			text += "\n" + m.Content
		}
	}
	text = strings.TrimSpace(text)
	if text == "" { return "", nil }

	chunks := ChunkByRune(text, p.cfg.SemanticCompression.ChunkSize)
	if p.cfg.Pinecone.UpsertOnLargeInputs {
		_ = p.pc.UpsertTexts(ctx, chunks, map[string]string{"source": "smart_prompt_opt"})
	}
	q := p.pc.QueryByText(ctx, text, p.cfg.Pinecone.TopK)
	if len(q) == 0 { return "", nil }
	// Summarize matches as compact bullets
	b := &strings.Builder{}
	b.WriteString("Context Summary (retrieved):\n")
	for i, m := range q {
		b.WriteString(fmt.Sprintf("- [%d] %s\n", i+1, Truncate(m.Text, 400)))
	}
	return b.String(), map[string]any{"hits": len(q)}
}

func (p *SmartPromptOptPlugin) injectSummary(req *schemas.BifrostRequest, summary string, meta map[string]any) {
	msg := schemas.BifrostMessage{Role: schemas.ModelChatMessageRoleSystem, Content: summary}
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		req.Input.ChatCompletionInput = &append([]schemas.BifrostMessage{msg}, msgs...)
		return
	}
	// convert text to chat form when necessary
	if req.Input.TextCompletionInput != nil {
		t := *req.Input.TextCompletionInput
		msgs := []schemas.BifrostMessage{msg, schemas.BifrostMessage{Role: schemas.ModelChatMessageRoleUser, Content: t}}
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
		if pp == "" { continue }
		if _, ok := seen[pp]; ok { continue }
		seen[pp] = struct{}{}
		out = append(out, pp)
	}
	return strings.Join(out, "\n\n")
}

func isEmptyRequest(req *schemas.BifrostRequest) bool {
	if req == nil { return true }
	if req.Input.TextCompletionInput != nil {
		if strings.TrimSpace(*req.Input.TextCompletionInput) == "" { return true }
	}
	if req.Input.ChatCompletionInput != nil {
		msgs := *req.Input.ChatCompletionInput
		if len(msgs) == 0 { return true }
		allEmpty := true
		for _, m := range msgs {
			if strings.TrimSpace(m.Content) != "" { allEmpty = false; break }
		}
		return allEmpty
	}
	return false
}

func getCorrelationID(ctx context.Context) string {
	// placeholder: if context has a value, derive from it; else timestamp
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

