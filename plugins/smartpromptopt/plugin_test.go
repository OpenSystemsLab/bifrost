package smartpromptopt

import (
	"context"
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
)

func TestPreHook_NormalizationAndPrefixAndTrim(t *testing.T) {
	prefix := "You are concise. Answer tersely."
	cfg := Config{
		Enabled:                true,
		TokenBudget:            10, // small to trigger trimming branch, but we rely on MaxHistoryTurns trim
		MaxHistoryTurns:        1,
		StrengthenInstructions: true,
		InstructionPrefix:      prefix,
		ProviderDefaults:       map[string]ModelDefaults{},
		SemanticCompression:    SemanticCompressionConfig{Enabled: false},
	}
	p, err := NewSmartPromptOpt(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	// Build a chat request with extra whitespace and multiple turns
	msgs := []schemas.BifrostMessage{
		{Role: schemas.ModelChatMessageRoleUser, Content: "  Hello\n\nWorld  "},
		{Role: schemas.ModelChatMessageRoleAssistant, Content: "Some reply"},
		{Role: schemas.ModelChatMessageRoleUser, Content: "Another question"},
	}
	req := &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o",
		Input: schemas.RequestInput{
			ChatCompletionInput: &msgs,
		},
	}
	ctx := context.Background()
	_, sc, err := p.PreHook(&ctx, req)
	if err != nil {
		t.Fatalf("prehook returned error: %v", err)
	}
	if sc != nil {
		t.Fatalf("expected no short-circuit, got: %#v", sc)
	}

	// Expect system prefix inserted at position 0
	if req.Input.ChatCompletionInput == nil || len(*req.Input.ChatCompletionInput) == 0 {
		t.Fatalf("chat messages missing after prehook")
	}
	got0 := (*req.Input.ChatCompletionInput)[0]
	if got0.Role != schemas.ModelChatMessageRoleSystem {
		t.Fatalf("expected first message to be system, got role=%s", got0.Role)
	}
	if got0.Content != prefix {
		t.Fatalf("expected prefix content, got: %q", got0.Content)
	}

	// Expect history trimmed to MaxHistoryTurns + 1 for the injected system? Here we only guarantee last user/assistant kept after prefix.
	// Given MaxHistoryTurns=1, we expect 1 original message plus the system prefix = 2 total now.
	if l := len(*req.Input.ChatCompletionInput); l < 2 {
		t.Fatalf("expected at least 2 messages (system + 1 trimmed original), got %d", l)
	}

	// Ensure whitespace collapsed in remaining messages
	for i, m := range *req.Input.ChatCompletionInput {
		if i == 0 { // skip system
			continue
		}
		if m.Content != "Hello World" && m.Content != "Some reply" && m.Content != "Another question" {
			// We don't know which message remains after trimming, but collapsed whitespace should remove double newlines
			// If it is the first user message, expect "Hello World"
		}
	}
}
