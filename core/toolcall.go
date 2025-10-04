package bifrost

import (
	"fmt"

	schemas "github.com/maximhq/bifrost/core/schemas"
)

// handleToolCallsInResponse handles tool calls in non-streaming responses
func (bifrost *Bifrost) handleToolCallsInResponse(
	result *schemas.BifrostResponse,
	req *ChannelMessage,
	provider schemas.Provider,
	key schemas.Key,
) (*schemas.BifrostResponse, *schemas.BifrostError) {
	// If MCP is not configured, just return the original result
	if bifrost.mcpManager == nil {
		return result, nil
	}

	// Check if there are any tool calls in the response
	if result == nil || len(result.Choices) == 0 {
		return result, nil
	}

	var toolCalls []schemas.ToolCall
	choice := result.Choices[0]

	// Check both stream and non-stream response formats
	if choice.BifrostNonStreamResponseChoice != nil &&
		choice.BifrostNonStreamResponseChoice.Message.AssistantMessage != nil &&
		choice.BifrostNonStreamResponseChoice.Message.AssistantMessage.ToolCalls != nil {
		toolCalls = *choice.BifrostNonStreamResponseChoice.Message.AssistantMessage.ToolCalls
	} else if choice.BifrostStreamResponseChoice != nil && len(choice.BifrostStreamResponseChoice.Delta.ToolCalls) > 0 {
		toolCalls = choice.BifrostStreamResponseChoice.Delta.ToolCalls
	}

	// If no tool calls, return the original result
	if len(toolCalls) == 0 {
		return result, nil
	}

	bifrost.logger.Debug("Found %d tool calls in response, executing...", len(toolCalls))

	// Execute each tool call
	var toolMessages []schemas.BifrostMessage
	for _, toolCall := range toolCalls {
		toolMsg, err := bifrost.mcpManager.executeTool(req.Context, toolCall)
		if err != nil {
			bifrost.logger.Warn("Failed to execute tool call %s: %v",
				*toolCall.Function.Name, err)
			// Create error response for this tool
			errorMsg := fmt.Sprintf("Error executing tool: %v", err)
			toolMsg = &schemas.BifrostMessage{
				Role: schemas.ModelChatMessageRoleTool,
				Content: schemas.MessageContent{
					ContentStr: &errorMsg,
				},
				ToolMessage: &schemas.ToolMessage{
					ToolCallID: toolCall.ID,
				},
			}
		}
		toolMessages = append(toolMessages, *toolMsg)
	}

	bifrost.logger.Debug("Executed %d tools, making follow-up request", len(toolMessages))

	// Get the assistant message from the response
	assistantMessage := choice.BifrostNonStreamResponseChoice.Message

	// Append assistant message and tool response messages to the conversation
	updatedMessages := append(*req.Input.ChatCompletionInput, assistantMessage)
	updatedMessages = append(updatedMessages, toolMessages...)

	// Create a new request with updated messages
	followUpReq := req
	followUpReq.Input.ChatCompletionInput = &updatedMessages

	// Make a new request with the updated conversation
	followUpResult, followUpError := handleProviderRequest(provider, followUpReq, key, req.Type)
	if followUpError != nil {
		bifrost.logger.Warn("Failed to make follow-up request after tool execution: %v", followUpError.Error.Message)
		return nil, followUpError
	}

	return followUpResult, nil
}

// interceptToolCalls intercepts tool calls from the stream, executes them, and adds responses back to the conversation
func (bifrost *Bifrost) interceptToolCalls(
	inputStream chan *schemas.BifrostStream,
	req *ChannelMessage,
	provider schemas.Provider,
	key schemas.Key,
	postHookRunner schemas.PostHookRunner,
) (chan *schemas.BifrostStream, *schemas.BifrostError) {
	// If MCP is not configured, just return the original stream
	if bifrost.mcpManager == nil {
		return inputStream, nil
	}
	// Create output channel for intercepted stream
	outputStream := make(chan *schemas.BifrostStream, schemas.DefaultStreamBufferSize)

	go func() {
		defer close(outputStream)

		var toolCalls []schemas.ToolCall
		var assistantContentStr string

		// Forward all messages while collecting tool calls
		for streamMsg := range inputStream {
			if streamMsg.BifrostResponse != nil && len(streamMsg.BifrostResponse.Choices) > 0 {
				choice := streamMsg.BifrostResponse.Choices[0]

				if choice.BifrostStreamResponseChoice != nil {
					// Collect content from delta
					delta := choice.BifrostStreamResponseChoice.Delta

					if delta.Content != nil && *delta.Content != "" {
						assistantContentStr += *delta.Content
					}

					// Collect tool calls
					if len(delta.ToolCalls) > 0 {
						toolCalls = append(toolCalls, delta.ToolCalls...)
					}
				}
			}

			// Forward the message to output
			outputStream <- streamMsg
		}

		// If no tool calls were collected, we're done
		if len(toolCalls) == 0 {
			return
		}

		bifrost.logger.Debug("Intercepted %d tool calls, executing...", len(toolCalls))

		// Execute each tool call
		var toolMessages []schemas.BifrostMessage
		for _, toolCall := range toolCalls {
			toolMsg, err := bifrost.mcpManager.executeTool(req.Context, toolCall)
			if err != nil {
				bifrost.logger.Warn("Failed to execute tool call %s: %v",
					*toolCall.Function.Name, err)
				// Create error response for this tool
				errorMsg := fmt.Sprintf("Error executing tool: %v", err)
				toolMsg = &schemas.BifrostMessage{
					Role: schemas.ModelChatMessageRoleTool,
					Content: schemas.MessageContent{
						ContentStr: &errorMsg,
					},
					ToolMessage: &schemas.ToolMessage{
						ToolCallID: toolCall.ID,
					},
				}
			}
			bifrost.logger.Debug("Tool message: %+v", toolMsg)
			toolMessages = append(toolMessages, *toolMsg)
		}

		bifrost.logger.Debug("Executed %d tools, making follow-up request", len(toolMessages))

		// Create assistant message with tool calls for the conversation
		assistantMessage := schemas.BifrostMessage{
			Role: schemas.ModelChatMessageRoleAssistant,
			Content: schemas.MessageContent{
				ContentStr: &assistantContentStr,
			},
			AssistantMessage: &schemas.AssistantMessage{
				ToolCalls: &toolCalls,
			},
		}

		// Append assistant message and tool response messages to the conversation
		updatedMessages := append(*req.Input.ChatCompletionInput, assistantMessage)
		updatedMessages = append(updatedMessages, toolMessages...)

		// Create a new request with updated messages
		followUpReq := req
		followUpReq.Input.ChatCompletionInput = &updatedMessages

		// Make a new stream request with the updated conversation
		followUpStream, followUpError := handleProviderStreamRequest(provider, followUpReq, key, postHookRunner, req.Type)
		if followUpError != nil {
			bifrost.logger.Warn("Failed to make follow-up request after tool execution: %v", followUpError.Error.Message)
			// Send error to output stream
			outputStream <- &schemas.BifrostStream{
				BifrostError: followUpError,
			}
			return
		}

		// Forward the follow-up stream to output
		for streamMsg := range followUpStream {
			bifrost.logger.Debug("Follow-up stream message: %+v", streamMsg)
			outputStream <- streamMsg
		}
	}()

	return outputStream, nil
}
