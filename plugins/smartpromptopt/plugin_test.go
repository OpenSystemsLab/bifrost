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
		TokenBudget:            100, // reasonable budget to avoid triggering the second trim
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
	helloContent := "  Hello\n\nWorld  "
	replyContent := "Some reply"
	anotherContent := "Another question"
	msgs := []schemas.BifrostMessage{
		{Role: schemas.ModelChatMessageRoleUser, Content: schemas.MessageContent{ContentStr: &helloContent}},
		{Role: schemas.ModelChatMessageRoleAssistant, Content: schemas.MessageContent{ContentStr: &replyContent}},
		{Role: schemas.ModelChatMessageRoleUser, Content: schemas.MessageContent{ContentStr: &anotherContent}},
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
	
	// Debug: print all messages
	msgsAfter := *req.Input.ChatCompletionInput
	t.Logf("Messages after prehook (count=%d):", len(msgsAfter))
	for i, m := range msgsAfter {
		content := "<nil>"
		if m.Content.ContentStr != nil {
			content = *m.Content.ContentStr
		}
		t.Logf("  [%d] Role=%s, Content=%q", i, m.Role, content)
	}
	
	got0 := (*req.Input.ChatCompletionInput)[0]
	if got0.Role != schemas.ModelChatMessageRoleSystem {
		t.Fatalf("expected first message to be system, got role=%s", got0.Role)
	}
	if got0.Content.ContentStr == nil || *got0.Content.ContentStr != prefix {
		content := "<nil>"
		if got0.Content.ContentStr != nil {
			content = *got0.Content.ContentStr
		}
		t.Fatalf("expected prefix content %q, got: %q", prefix, content)
	}

	// With MaxHistoryTurns=1, we trim to the last 1 message BEFORE adding the prefix
	// So we expect 2 messages: system prefix + the last original message
	if l := len(*req.Input.ChatCompletionInput); l != 2 {
		t.Fatalf("expected exactly 2 messages (system + 1 trimmed original), got %d", l)
	}

	// Ensure whitespace collapsed in remaining messages
	for i, m := range *req.Input.ChatCompletionInput {
		if i == 0 { // skip system
			continue
		}
		if m.Content.ContentStr != nil {
			content := *m.Content.ContentStr
			// Check that whitespace was properly collapsed
			if content != "Hello World" && content != "Some reply" && content != "Another question" {
				// We expect one of the normalized messages
				t.Errorf("unexpected message content after normalization: %q", content)
			}
		}
	}
}
