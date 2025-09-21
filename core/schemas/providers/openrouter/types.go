package openrouter

import "github.com/maximhq/bifrost/core/schemas"

// OpenRouter request structures

// OpenRouterProviderPreferences represents provider routing preferences
type OpenRouterProviderPreferences struct {
	Sort string `json:"sort,omitempty"` // Sort preference (e.g., price, throughput)
}

// OpenRouterReasoning represents reasoning/thinking token configuration
type OpenRouterReasoning struct {
	Effort    *string `json:"effort,omitempty"`     // "high", "medium", "low" - OpenAI-style reasoning effort
	MaxTokens *int    `json:"max_tokens,omitempty"` // Non-OpenAI-style reasoning effort setting
	Exclude   *bool   `json:"exclude,omitempty"`    // Whether to exclude reasoning from the response (defaults to false)
}

// OpenRouterUsage represents usage information preferences
type OpenRouterUsage struct {
	IncludeUsage *bool `json:"include_usage,omitempty"` // Whether to include usage information in the response
}

// OpenRouterCompletionRequest represents a completion request to OpenRouter API
type OpenRouterCompletionRequest struct {
	// Required fields
	Model  string `json:"model"`  // The model ID to use
	Prompt string `json:"prompt"` // The text prompt to complete

	// OpenRouter-specific fields
	Models     *[]string                      `json:"models,omitempty"`     // Alternate list of models for routing overrides
	Provider   *OpenRouterProviderPreferences `json:"provider,omitempty"`   // Preferences for provider routing
	Reasoning  *OpenRouterReasoning           `json:"reasoning,omitempty"`  // Configuration for reasoning tokens
	Usage      *OpenRouterUsage               `json:"usage,omitempty"`      // Whether to include usage information
	Transforms *[]string                      `json:"transforms,omitempty"` // List of prompt transforms (OpenRouter-only)

	// Standard completion parameters
	Stream            *bool              `json:"stream,omitempty"`             // Enable streaming of results (defaults to false)
	MaxTokens         *int               `json:"max_tokens,omitempty"`         // Maximum number of tokens (range: [1, context_length))
	Temperature       *float64           `json:"temperature,omitempty"`        // Sampling temperature (range: [0, 2])
	Seed              *int               `json:"seed,omitempty"`               // Seed for deterministic outputs
	TopP              *float64           `json:"top_p,omitempty"`              // Top-p sampling value (range: (0, 1])
	TopK              *int               `json:"top_k,omitempty"`              // Top-k sampling value (range: [1, Infinity))
	FrequencyPenalty  *float64           `json:"frequency_penalty,omitempty"`  // Frequency penalty (range: [-2, 2])
	PresencePenalty   *float64           `json:"presence_penalty,omitempty"`   // Presence penalty (range: [-2, 2])
	RepetitionPenalty *float64           `json:"repetition_penalty,omitempty"` // Repetition penalty (range: (0, 2])
	LogitBias         map[string]float64 `json:"logit_bias,omitempty"`         // Mapping of token IDs to bias values
	TopLogprobs       *int               `json:"top_logprobs,omitempty"`       // Number of top log probabilities to return
	MinP              *float64           `json:"min_p,omitempty"`              // Minimum probability threshold (range: [0, 1])
	TopA              *float64           `json:"top_a,omitempty"`              // Alternate top sampling parameter (range: [0, 1])
	User              *string            `json:"user,omitempty"`               // A stable identifier for end-users
	Stop              *[]string          `json:"stop,omitempty"`               // Stop sequences
}

// OpenRouter response structures

// OpenRouterTextResponse represents the response from OpenRouter text completion API
type OpenRouterTextResponse struct {
	ID                string                 `json:"id"`
	Model             string                 `json:"model"`
	Created           int                    `json:"created"`
	SystemFingerprint *string                `json:"system_fingerprint"`
	Choices           []OpenRouterTextChoice `json:"choices"`
	Usage             *schemas.LLMUsage      `json:"usage"`
}

// OpenRouterTextChoice represents a choice in the OpenRouter text completion response
type OpenRouterTextChoice struct {
	Text         string `json:"text"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}
