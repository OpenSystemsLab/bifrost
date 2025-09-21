package openai

import (
	"github.com/maximhq/bifrost/core/schemas"
)

// ConvertTextRequestToOpenAI converts a Bifrost text completion request to OpenAI format
func ConvertTextRequestToOpenAI(bifrostReq *schemas.BifrostRequest) *OpenAITextCompletionRequest {
	if bifrostReq == nil || bifrostReq.Input.TextCompletionInput == nil {
		return nil
	}

	openaiReq := &OpenAITextCompletionRequest{
		Model:           bifrostReq.Model,
		Prompt:          *bifrostReq.Input.TextCompletionInput,
		ModelParameters: bifrostReq.Params, // Directly embed the parameters
	}

	// Handle OpenAI-specific parameters from ExtraParams
	if bifrostReq.Params != nil && bifrostReq.Params.ExtraParams != nil {
		// Log probabilities (int version, different from LogProbs bool)
		if logprobs, ok := bifrostReq.Params.ExtraParams["logprobs"].(int); ok {
			openaiReq.Logprobs = &logprobs
		}

		// Echo prompt
		if echo, ok := bifrostReq.Params.ExtraParams["echo"].(bool); ok {
			openaiReq.Echo = &echo
		}

		// Best of
		if bestOf, ok := bifrostReq.Params.ExtraParams["best_of"].(int); ok {
			openaiReq.BestOf = &bestOf
		}

		// Suffix
		if suffix, ok := bifrostReq.Params.ExtraParams["suffix"].(string); ok {
			openaiReq.Suffix = &suffix
		}
	}

	return openaiReq
}

// ConvertOpenAITextResponseToBifrost converts an OpenAI text completion response to Bifrost format
func ConvertOpenAITextResponseToBifrost(response *OpenAITextCompletionResponse, model string, providerName schemas.ModelProvider) *schemas.BifrostResponse {
	if response == nil {
		return nil
	}

	// Convert choices
	choices := make([]schemas.BifrostResponseChoice, 0, len(response.Choices))
	for i, choice := range response.Choices {
		// Create a copy of the text to avoid pointer issues
		textCopy := choice.Text

		bifrostChoice := schemas.BifrostResponseChoice{
			Index: i,
			BifrostNonStreamResponseChoice: &schemas.BifrostNonStreamResponseChoice{
				Message: schemas.BifrostMessage{
					Role: schemas.ModelChatMessageRoleAssistant,
					Content: schemas.MessageContent{
						ContentStr: &textCopy,
					},
				},
			},
			FinishReason: choice.FinishReason,
		}

		// Add log probabilities if available
		if choice.Logprobs != nil {
			bifrostChoice.BifrostNonStreamResponseChoice.LogProbs = &schemas.LogProbs{
				Text: *choice.Logprobs,
			}
		}

		choices = append(choices, bifrostChoice)
	}

	// Create the Bifrost response
	bifrostResponse := &schemas.BifrostResponse{
		ID:      response.ID,
		Object:  "list", // Standard Bifrost object type for completions
		Choices: choices,
		Model:   model,
		Created: response.Created,
		ExtraFields: schemas.BifrostResponseExtraFields{
			Provider: providerName,
		},
	}

	// Set system fingerprint
	if response.SystemFingerprint != nil {
		bifrostResponse.SystemFingerprint = response.SystemFingerprint
	}

	// Set usage information
	if response.Usage != nil {
		// Create a copy to avoid pointer issues
		usageCopy := *response.Usage
		bifrostResponse.Usage = &usageCopy
	}

	return bifrostResponse
}
