package smartpromptopt

import (
	"unicode/utf8"

	"github.com/maximhq/bifrost/core/schemas"
)

type TokenEstimator interface {
	Estimate(req *schemas.BifrostRequest) int
}

// DefaultEstimator provides a basic character-based token estimation
type DefaultEstimator struct{
	CharsPerToken int
}

// NewDefaultEstimator creates a new estimator with configurable chars-per-token ratio
func NewDefaultEstimator(charsPerToken int) TokenEstimator {
	if charsPerToken <= 0 {
		charsPerToken = DefaultTokensPerChar
	}
	return DefaultEstimator{CharsPerToken: charsPerToken}
}

func (e DefaultEstimator) Estimate(req *schemas.BifrostRequest) int {
	if req == nil { 
		return 0 
	}
	chars := countRequestCharacters(req)
	if chars <= 0 { 
		return 0 
	}
	// Round up division
	return (chars + e.CharsPerToken - 1) / e.CharsPerToken
}

// countRequestCharacters counts total characters in a request
func countRequestCharacters(req *schemas.BifrostRequest) int {
	chars := 0
	
	// Count text completion input
	if req.Input.TextCompletionInput != nil {
		chars += utf8.RuneCountInString(*req.Input.TextCompletionInput)
	}
	
	// Count chat messages
	if req.Input.ChatCompletionInput != nil {
		for _, m := range *req.Input.ChatCompletionInput {
			chars += countMessageCharacters(m)
		}
	}
	
	return chars
}

// countMessageCharacters counts characters in a single message
func countMessageCharacters(msg schemas.BifrostMessage) int {
	chars := 0
	
	// Count string content
	if msg.Content.ContentStr != nil {
		chars += utf8.RuneCountInString(*msg.Content.ContentStr)
	}
	
	// Count content blocks
	if msg.Content.ContentBlocks != nil {
		for _, block := range *msg.Content.ContentBlocks {
			if block.Text != nil {
				chars += utf8.RuneCountInString(*block.Text)
			}
		}
	}
	
	return chars
}
