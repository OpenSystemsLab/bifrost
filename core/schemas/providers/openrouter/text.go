package openrouter

import (
	"github.com/maximhq/bifrost/core/schemas"
)

// ConvertTextRequestToOpenRouter converts a Bifrost text completion request to OpenRouter format
func ConvertTextRequestToOpenRouter(bifrostReq *schemas.BifrostRequest) *OpenRouterCompletionRequest {
	if bifrostReq == nil || bifrostReq.Input.TextCompletionInput == nil {
		return nil
	}

	openRouterReq := &OpenRouterCompletionRequest{
		Model:  bifrostReq.Model,
		Prompt: *bifrostReq.Input.TextCompletionInput,
	}

	// Convert parameters if present
	if bifrostReq.Params != nil {
		// Standard completion parameters
		openRouterReq.MaxTokens = bifrostReq.Params.MaxTokens
		openRouterReq.Temperature = bifrostReq.Params.Temperature
		openRouterReq.TopP = bifrostReq.Params.TopP
		openRouterReq.TopK = bifrostReq.Params.TopK
		openRouterReq.Stop = bifrostReq.Params.StopSequences
		openRouterReq.FrequencyPenalty = bifrostReq.Params.FrequencyPenalty
		openRouterReq.PresencePenalty = bifrostReq.Params.PresencePenalty
		openRouterReq.User = bifrostReq.Params.User
		openRouterReq.Seed = bifrostReq.Params.Seed

		// Handle LogitBias - direct assignment since both are map[string]float64
		if bifrostReq.Params.LogitBias != nil {
			openRouterReq.LogitBias = bifrostReq.Params.LogitBias
		}

		// Handle extra parameters for OpenRouter-specific fields
		if bifrostReq.Params.ExtraParams != nil {
			// Check for OpenRouter-specific parameters in ExtraParams
			if models, ok := bifrostReq.Params.ExtraParams["models"].([]string); ok {
				openRouterReq.Models = &models
			}

			if transforms, ok := bifrostReq.Params.ExtraParams["transforms"].([]string); ok {
				openRouterReq.Transforms = &transforms
			}

			if topLogprobs, ok := bifrostReq.Params.ExtraParams["top_logprobs"].(int); ok {
				openRouterReq.TopLogprobs = &topLogprobs
			}

			if minP, ok := bifrostReq.Params.ExtraParams["min_p"].(float64); ok {
				openRouterReq.MinP = &minP
			}

			if topA, ok := bifrostReq.Params.ExtraParams["top_a"].(float64); ok {
				openRouterReq.TopA = &topA
			}

			if repetitionPenalty, ok := bifrostReq.Params.ExtraParams["repetition_penalty"].(float64); ok {
				openRouterReq.RepetitionPenalty = &repetitionPenalty
			}

			// Handle provider preferences
			if providerData, ok := bifrostReq.Params.ExtraParams["provider"].(map[string]interface{}); ok {
				provider := &OpenRouterProviderPreferences{}
				if sort, ok := providerData["sort"].(string); ok {
					provider.Sort = sort
				}
				if provider.Sort != "" {
					openRouterReq.Provider = provider
				}
			}

			// Handle reasoning configuration
			if reasoningData, ok := bifrostReq.Params.ExtraParams["reasoning"].(map[string]interface{}); ok {
				reasoning := &OpenRouterReasoning{}
				if effort, ok := reasoningData["effort"].(string); ok {
					reasoning.Effort = &effort
				}
				if maxTokens, ok := reasoningData["max_tokens"].(int); ok {
					reasoning.MaxTokens = &maxTokens
				}
				if exclude, ok := reasoningData["exclude"].(bool); ok {
					reasoning.Exclude = &exclude
				}
				if reasoning.Effort != nil || reasoning.MaxTokens != nil || reasoning.Exclude != nil {
					openRouterReq.Reasoning = reasoning
				}
			}

			// Handle usage preferences
			if usageData, ok := bifrostReq.Params.ExtraParams["usage"].(map[string]interface{}); ok {
				usage := &OpenRouterUsage{}
				if includeUsage, ok := usageData["include_usage"].(bool); ok {
					usage.IncludeUsage = &includeUsage
				}
				if usage.IncludeUsage != nil {
					openRouterReq.Usage = usage
				}
			}
		}
	}

	return openRouterReq
}

// ConvertOpenRouterTextResponseToBifrost converts an OpenRouter text response to Bifrost format
func ConvertOpenRouterTextResponseToBifrost(response *OpenRouterTextResponse) *schemas.BifrostResponse {
	if response == nil {
		return nil
	}

	// Convert choices
	choices := make([]schemas.BifrostResponseChoice, 0, len(response.Choices))
	for i, ch := range response.Choices {
		txt := ch.Text        // local copy
		fr := ch.FinishReason // local copy
		choices = append(choices, schemas.BifrostResponseChoice{
			Index: i,
			BifrostNonStreamResponseChoice: &schemas.BifrostNonStreamResponseChoice{
				Message: schemas.BifrostMessage{
					Role:    schemas.ModelChatMessageRoleAssistant,
					Content: schemas.MessageContent{ContentStr: &txt},
				},
			},
			FinishReason: &fr,
		})
	}

	// Create Bifrost response
	bifrostResponse := &schemas.BifrostResponse{
		ID:                response.ID,
		Choices:           choices,
		Model:             response.Model,
		Created:           response.Created,
		SystemFingerprint: response.SystemFingerprint,
		Usage:             response.Usage,
		ExtraFields: schemas.BifrostResponseExtraFields{
			Provider: schemas.OpenRouter,
		},
	}

	return bifrostResponse
}
