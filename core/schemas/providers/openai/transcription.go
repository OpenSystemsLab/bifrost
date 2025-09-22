package openai

import "github.com/maximhq/bifrost/core/schemas"

// ConvertTranscriptionRequestToBifrost converts an OpenAI transcription request to Bifrost format
func (r *OpenAITranscriptionRequest) ConvertTranscriptionRequestToBifrost() *schemas.BifrostRequest {
	provider, model := schemas.ParseModelString(r.Model, schemas.OpenAI)

	// Create transcription input
	transcriptionInput := &schemas.TranscriptionInput{
		File: r.File,
	}

	// Set optional fields
	if r.Language != nil {
		transcriptionInput.Language = r.Language
	}
	if r.Prompt != nil {
		transcriptionInput.Prompt = r.Prompt
	}
	if r.ResponseFormat != nil {
		transcriptionInput.ResponseFormat = r.ResponseFormat
	}

	bifrostReq := &schemas.BifrostRequest{
		Provider: provider,
		Model:    model,
		Input: schemas.RequestInput{
			TranscriptionInput: transcriptionInput,
		},
	}

	// Convert parameters first
	params := r.convertTranscriptionParameters()

	// Map parameters
	bifrostReq.Params = filterParams(provider, params)

	return bifrostReq
}

// ConvertTranscriptionResponseToOpenAI converts a Bifrost transcription response to OpenAI format
func ConvertTranscriptionResponseToOpenAI(bifrostResp *schemas.BifrostResponse) *schemas.BifrostTranscribe {
	if bifrostResp == nil || bifrostResp.Transcribe == nil {
		return nil
	}
	return bifrostResp.Transcribe
}

// ConvertTranscriptionRequestToOpenAI converts a Bifrost transcription request to OpenAI format
func ConvertTranscriptionRequestToOpenAI(bifrostReq *schemas.BifrostRequest) *OpenAITranscriptionRequest {
	if bifrostReq == nil || bifrostReq.Input.TranscriptionInput == nil {
		return nil
	}

	transcriptionInput := bifrostReq.Input.TranscriptionInput
	params := bifrostReq.Params

	openaiReq := &OpenAITranscriptionRequest{
		Model: bifrostReq.Model,
		File:  transcriptionInput.File,
	}

	// Set optional fields
	openaiReq.Language = transcriptionInput.Language
	openaiReq.Prompt = transcriptionInput.Prompt
	openaiReq.ResponseFormat = transcriptionInput.ResponseFormat

	// Map parameters
	if params != nil && params.ExtraParams != nil {
		if temperature, ok := params.ExtraParams["temperature"].(float64); ok {
			openaiReq.Temperature = &temperature
		}
		if include, ok := params.ExtraParams["include"].([]string); ok {
			openaiReq.Include = include
		}
		if timestampGranularities, ok := params.ExtraParams["timestamp_granularities"].([]string); ok {
			openaiReq.TimestampGranularities = timestampGranularities
		}
		if stream, ok := params.ExtraParams["stream"].(bool); ok {
			openaiReq.Stream = &stream
		}
	}

	return openaiReq
}

