package smartpromptopt

import (
	"context"
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
)

func TestPreHook_EmptyRequest(t *testing.T) {
	cfg := Config{
		Enabled: true,
		TokenBudget: 100,
	}
	p, err := NewSmartPromptOpt(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	// Test with nil request
	ctx := context.Background()
	result, sc, err := p.PreHook(&ctx, nil)
	if err != nil {
		t.Errorf("unexpected error for nil request: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result for nil request, got %v", result)
	}
	if sc != nil {
		t.Errorf("expected nil short-circuit for nil request, got %v", sc)
	}

	// Test with empty messages
	emptyMsgs := []schemas.BifrostMessage{}
	req := &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o",
		Input: schemas.RequestInput{
			ChatCompletionInput: &emptyMsgs,
		},
	}
	
	_, sc, err = p.PreHook(&ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sc == nil {
		t.Error("expected short-circuit for empty request")
	}
	if sc.Error.StatusCode == nil || *sc.Error.StatusCode != 400 {
		t.Errorf("expected status code 400, got %v", sc.Error.StatusCode)
	}
}

func TestPreHook_DisabledPlugin(t *testing.T) {
	cfg := Config{
		Enabled: false, // Plugin disabled
		TokenBudget: 100,
	}
	p, err := NewSmartPromptOpt(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	content := "Test message"
	msgs := []schemas.BifrostMessage{
		{Role: schemas.ModelChatMessageRoleUser, Content: schemas.MessageContent{ContentStr: &content}},
	}
	req := &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o",
		Input: schemas.RequestInput{
			ChatCompletionInput: &msgs,
		},
	}
	
	ctx := context.Background()
	result, sc, err := p.PreHook(&ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sc != nil {
		t.Errorf("unexpected short-circuit: %v", sc)
	}
	if result != req {
		t.Error("expected request to be returned unchanged when plugin is disabled")
	}
}

func TestPreHook_ContentBlocks(t *testing.T) {
	cfg := Config{
		Enabled: true,
		TokenBudget: 100,
		StrengthenInstructions: true,
		InstructionPrefix: "Be concise",
	}
	p, err := NewSmartPromptOpt(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	// Test with content blocks
	text1 := "  Hello  "
	text2 := "  World  "
	msgs := []schemas.BifrostMessage{
		{
			Role: schemas.ModelChatMessageRoleUser,
			Content: schemas.MessageContent{
				ContentBlocks: &[]schemas.ContentBlock{
					{Type: schemas.ContentBlockTypeText, Text: &text1},
					{Type: schemas.ContentBlockTypeText, Text: &text2},
				},
			},
		},
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
		t.Errorf("unexpected error: %v", err)
	}
	if sc != nil {
		t.Errorf("unexpected short-circuit: %v", sc)
	}

	// Check that content blocks were normalized
	msgsAfter := *req.Input.ChatCompletionInput
	if len(msgsAfter) != 2 { // System message + user message
		t.Fatalf("expected 2 messages, got %d", len(msgsAfter))
	}
	
	userMsg := msgsAfter[1]
	if userMsg.Content.ContentBlocks == nil {
		t.Fatal("expected content blocks to be preserved")
	}
	
	blocks := *userMsg.Content.ContentBlocks
	if len(blocks) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(blocks))
	}
	
	// Check normalized text
	if blocks[0].Text == nil || *blocks[0].Text != "Hello" {
		t.Errorf("expected first block to be normalized to 'Hello', got %v", blocks[0].Text)
	}
	if blocks[1].Text == nil || *blocks[1].Text != "World" {
		t.Errorf("expected second block to be normalized to 'World', got %v", blocks[1].Text)
	}
}

func TestPreHook_ProviderDefaults(t *testing.T) {
	temp := 0.2
	topP := 0.9
	cfg := Config{
		Enabled: true,
		ProviderDefaults: map[string]ModelDefaults{
			"openai:gpt-4o": {
				Temperature: &temp,
				TopP:        &topP,
			},
		},
	}
	p, err := NewSmartPromptOpt(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating plugin: %v", err)
	}

	content := "Test"
	msgs := []schemas.BifrostMessage{
		{Role: schemas.ModelChatMessageRoleUser, Content: schemas.MessageContent{ContentStr: &content}},
	}
	req := &schemas.BifrostRequest{
		Provider: schemas.OpenAI,
		Model:    "gpt-4o",
		Input: schemas.RequestInput{
			ChatCompletionInput: &msgs,
		},
		// No params set initially
	}
	
	ctx := context.Background()
	_, sc, err := p.PreHook(&ctx, req)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if sc != nil {
		t.Errorf("unexpected short-circuit: %v", sc)
	}

	// Check that defaults were applied
	if req.Params == nil {
		t.Fatal("expected params to be initialized")
	}
	if req.Params.Temperature == nil || *req.Params.Temperature != temp {
		t.Errorf("expected temperature %v, got %v", temp, req.Params.Temperature)
	}
	if req.Params.TopP == nil || *req.Params.TopP != topP {
		t.Errorf("expected topP %v, got %v", topP, req.Params.TopP)
	}
}
