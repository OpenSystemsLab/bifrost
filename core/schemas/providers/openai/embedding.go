package openai

import (
	"github.com/bytedance/sonic"
	"github.com/maximhq/bifrost/core/schemas"
)

// ConvertEmbeddingRequestToBifrost converts an OpenAI embedding request to Bifrost format
func (r *OpenAIEmbeddingRequest) ConvertEmbeddingRequestToBifrost() *schemas.BifrostRequest {
	provider, model := schemas.ParseModelString(r.Model, schemas.OpenAI)

	// Create embedding input
	embeddingInput := &schemas.EmbeddingInput{}

	// Cleaner coercion: marshal input and try to unmarshal into supported shapes
	if raw, err := sonic.Marshal(r.Input); err == nil {
		// 1) string
		var s string
		if err := sonic.Unmarshal(raw, &s); err == nil {
			embeddingInput.Text = &s
		} else {
			// 2) []string
			var ss []string
			if err := sonic.Unmarshal(raw, &ss); err == nil {
				embeddingInput.Texts = ss
			} else {
				// 3) []int
				var i []int
				if err := sonic.Unmarshal(raw, &i); err == nil {
					embeddingInput.Embedding = i
				} else {
					// 4) [][]int
					var ii [][]int
					if err := sonic.Unmarshal(raw, &ii); err == nil {
						embeddingInput.Embeddings = ii
					}
				}
			}
		}
	}

	bifrostReq := &schemas.BifrostRequest{
		Provider: provider,
		Model:    model,
		Input: schemas.RequestInput{
			EmbeddingInput: embeddingInput,
		},
	}

	// Convert parameters first
	params := r.convertEmbeddingParameters()

	// Map parameters
	bifrostReq.Params = filterParams(provider, params)

	return bifrostReq
}

// ConvertEmbeddingResponseToOpenAI converts a Bifrost embedding response to OpenAI format
func ConvertEmbeddingResponseToOpenAI(bifrostResp *schemas.BifrostResponse) *OpenAIEmbeddingResponse {
	if bifrostResp == nil || bifrostResp.Data == nil {
		return nil
	}

	return &OpenAIEmbeddingResponse{
		Object:            "list",
		Data:              bifrostResp.Data,
		Model:             bifrostResp.Model,
		Usage:             bifrostResp.Usage,
		ServiceTier:       bifrostResp.ServiceTier,
		SystemFingerprint: bifrostResp.SystemFingerprint,
	}
}

// ConvertEmbeddingRequestToOpenAI converts a Bifrost embedding request to OpenAI format
func ConvertEmbeddingRequestToOpenAI(bifrostReq *schemas.BifrostRequest) *OpenAIEmbeddingRequest {
	if bifrostReq == nil || bifrostReq.Input.EmbeddingInput == nil {
		return nil
	}

	embeddingInput := bifrostReq.Input.EmbeddingInput
	params := bifrostReq.Params

	openaiReq := &OpenAIEmbeddingRequest{
		Model: bifrostReq.Model,
	}

	// Set input - convert to interface{} for flexibility
	if len(embeddingInput.Texts) == 1 {
		openaiReq.Input = embeddingInput.Texts[0] // Single string
	} else {
		openaiReq.Input = embeddingInput.Texts // Array of strings
	}

	// Map parameters
	if params != nil {
		openaiReq.EncodingFormat = params.EncodingFormat
		openaiReq.Dimensions = params.Dimensions
		openaiReq.User = params.User
	}

	return openaiReq
}