package smartpromptopt

import (
	"testing"

	"github.com/maximhq/bifrost/core/schemas"
)

func TestDefaultEstimator(t *testing.T) {
	tests := []struct {
		name          string
		charsPerToken int
		input         string
		expectedTokens int
	}{
		{
			name:          "default ratio",
			charsPerToken: 0, // Will use default (4)
			input:         "Hello, world!", // 13 chars
			expectedTokens: 4, // (13 + 3) / 4 = 4
		},
		{
			name:          "custom ratio",
			charsPerToken: 3,
			input:         "Hello, world!", // 13 chars
			expectedTokens: 5, // (13 + 2) / 3 = 5
		},
		{
			name:          "empty input",
			charsPerToken: 4,
			input:         "",
			expectedTokens: 0,
		},
		{
			name:          "unicode characters",
			charsPerToken: 4,
			input:         "Hello 世界", // 8 runes
			expectedTokens: 2, // (8 + 3) / 4 = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := NewDefaultEstimator(tt.charsPerToken)
			
			req := &schemas.BifrostRequest{
				Input: schemas.RequestInput{
					TextCompletionInput: &tt.input,
				},
			}
			
			got := estimator.Estimate(req)
			if got != tt.expectedTokens {
				t.Errorf("Estimate() = %d, want %d", got, tt.expectedTokens)
			}
		})
	}
}

func TestCountMessageCharacters(t *testing.T) {
	content := "Hello, world!"
	msg := schemas.BifrostMessage{
		Role: schemas.ModelChatMessageRoleUser,
		Content: schemas.MessageContent{
			ContentStr: &content,
		},
	}
	
	chars := countMessageCharacters(msg)
	if chars != 13 {
		t.Errorf("countMessageCharacters() = %d, want 13", chars)
	}
	
	// Test with content blocks
	text1 := "Hello"
	text2 := "World"
	msgWithBlocks := schemas.BifrostMessage{
		Role: schemas.ModelChatMessageRoleUser,
		Content: schemas.MessageContent{
			ContentBlocks: &[]schemas.ContentBlock{
				{Type: schemas.ContentBlockTypeText, Text: &text1},
				{Type: schemas.ContentBlockTypeText, Text: &text2},
			},
		},
	}
	
	chars = countMessageCharacters(msgWithBlocks)
	if chars != 10 { // "Hello" + "World" = 10
		t.Errorf("countMessageCharacters() with blocks = %d, want 10", chars)
	}
}
