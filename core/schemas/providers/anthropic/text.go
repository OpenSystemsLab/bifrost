package anthropic

import (
	"fmt"

	"github.com/maximhq/bifrost/core/schemas"
)

// ConvertTextRequestToAnthropic converts a Bifrost text completion request to Anthropic format
func ConvertTextRequestToAnthropic(bifrostReq *schemas.BifrostRequest) *AnthropicTextRequest {
	anthropicReq := &AnthropicTextRequest{
		Model:             bifrostReq.Model,
		Prompt:            fmt.Sprintf("\n\nHuman: %s\n\nAssistant:", *bifrostReq.Input.TextCompletionInput),
		MaxTokensToSample: 4096, // Default value
	}

	// Convert parameters
	if bifrostReq.Params != nil {
		if bifrostReq.Params.MaxTokens != nil {
			anthropicReq.MaxTokensToSample = *bifrostReq.Params.MaxTokens
		}
		anthropicReq.Temperature = bifrostReq.Params.Temperature
		anthropicReq.TopP = bifrostReq.Params.TopP
		anthropicReq.TopK = bifrostReq.Params.TopK
		anthropicReq.StopSequences = bifrostReq.Params.StopSequences
	}

	return anthropicReq
}
