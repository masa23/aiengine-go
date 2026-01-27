// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing chat completion functionality.
// The chat completion functionality generates AI responses to user messages.
//
// Basic usage:
//
//	req := &aiengine.ChatCompletionRequest{
//	    Model: "gpt-oss-120b",
//	    Messages: []aiengine.ChatMessage{
//	        {Role: "user", Content: "Hello!"},
//	    },
//	}
//	resp, err := client.CreateChatCompletion(context.Background(), req)
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ChatCompletionMessageRoleType represents the role of a chat message.
type ChatCompletionMessageRoleType string

const (
	// ChatCompletionMessageRoleTypeDeveloper represents a developer message role
	ChatCompletionMessageRoleTypeDeveloper ChatCompletionMessageRoleType = "developer"
	// ChatCompletionMessageRoleTypeSystem represents a system message role
	ChatCompletionMessageRoleTypeSystem ChatCompletionMessageRoleType = "system"
	// ChatCompletionMessageRoleTypeUser represents a user message role
	ChatCompletionMessageRoleTypeUser ChatCompletionMessageRoleType = "user"
	// ChatCompletionMessageRoleTypeAssistant represents an assistant message role
	ChatCompletionMessageRoleTypeAssistant ChatCompletionMessageRoleType = "assistant"
	// ChatCompletionMessageRoleTypeTool represents a tool message role
	ChatCompletionMessageRoleTypeTool ChatCompletionMessageRoleType = "tool"
)

// ChatCompletionRequestMessageContentPartText represents the structure for text content.
type ChatCompletionRequestMessageContentPartText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ChatCompletionRequestMessageContentPartImage represents the structure for image content.
type ChatCompletionRequestMessageContentPartImage struct {
	Type     string `json:"type"`
	ImageURL struct {
		URL string `json:"url"`
	} `json:"image_url"`
}

// ChatCompletionRequestUserMessageContentPart is the interface for user message content parts.
type ChatCompletionRequestUserMessageContentPart interface {
	isChatCompletionRequestUserMessageContentPart()
}

func (ChatCompletionRequestMessageContentPartText) isChatCompletionRequestUserMessageContentPart()  {}
func (ChatCompletionRequestMessageContentPartImage) isChatCompletionRequestUserMessageContentPart() {}

// ChatCompletionRequestSystemMessageContentPart is the interface for system message content parts.
type ChatCompletionRequestSystemMessageContentPart interface {
	isChatCompletionRequestSystemMessageContentPart()
}

func (ChatCompletionRequestMessageContentPartText) isChatCompletionRequestSystemMessageContentPart() {
}

// ChatCompletionRequestAssistantMessageContentPart is the interface for assistant message content parts.
type ChatCompletionRequestAssistantMessageContentPart interface {
	isChatCompletionRequestAssistantMessageContentPart()
}

func (ChatCompletionRequestMessageContentPartText) isChatCompletionRequestAssistantMessageContentPart() {
}

// ChatCompletionRequestToolMessageContentPart is the interface for tool message content parts.
type ChatCompletionRequestToolMessageContentPart interface {
	isChatCompletionRequestToolMessageContentPart()
}

func (ChatCompletionRequestMessageContentPartText) isChatCompletionRequestToolMessageContentPart() {}

// ChatCompletionRequestDeveloperMessageContent represents the content type for developer messages.
type ChatCompletionRequestDeveloperMessageContent interface{}

// ChatCompletionRequestDeveloperMessage represents the structure for developer messages.
type ChatCompletionRequestDeveloperMessage struct {
	// Content is the message content (string or []ChatCompletionRequestMessageContentPartText)
	Content ChatCompletionRequestDeveloperMessageContent `json:"content"`
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
}

// ChatCompletionRequestSystemMessageContent represents the content type for system messages.
type ChatCompletionRequestSystemMessageContent interface{}

// ChatCompletionRequestSystemMessage represents the structure for system messages.
type ChatCompletionRequestSystemMessage struct {
	// Content is the message content (string or []ChatCompletionRequestSystemMessageContentPart)
	Content ChatCompletionRequestSystemMessageContent `json:"content"`
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
}

// ChatCompletionRequestUserMessageContent represents the content type for user messages.
type ChatCompletionRequestUserMessageContent interface{}

// ChatCompletionRequestUserMessage represents the structure for user messages.
type ChatCompletionRequestUserMessage struct {
	// Content is the message content (string or []ChatCompletionRequestUserMessageContentPart)
	Content ChatCompletionRequestUserMessageContent `json:"content"`
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
}

// ChatCompletionRequestAssistantMessageContent represents the content type for assistant messages.
type ChatCompletionRequestAssistantMessageContent interface{}

// ChatCompletionRequestAssistantMessage represents the structure for assistant messages.
type ChatCompletionRequestAssistantMessage struct {
	// Content is the message content (string or []ChatCompletionRequestAssistantMessageContentPart)
	Content ChatCompletionRequestAssistantMessageContent `json:"content,omitempty"`
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
}

// ChatCompletionRequestToolMessageContent represents the content type for tool messages.
type ChatCompletionRequestToolMessageContent interface{}

// ChatCompletionRequestToolMessage represents the structure for tool messages.
type ChatCompletionRequestToolMessage struct {
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
	// Content is the message content (string or []ChatCompletionRequestToolMessageContentPart)
	Content ChatCompletionRequestToolMessageContent `json:"content"`
	// ToolCallID is the ID of the tool call
	ToolCallID string `json:"tool_call_id"`
}

// ChatCompletionRequestMessage is the interface for chat messages.
type ChatCompletionRequestMessage interface {
	isChatCompletionRequestMessage()
}

func (ChatCompletionRequestDeveloperMessage) isChatCompletionRequestMessage() {}
func (ChatCompletionRequestSystemMessage) isChatCompletionRequestMessage()    {}
func (ChatCompletionRequestUserMessage) isChatCompletionRequestMessage()      {}
func (ChatCompletionRequestAssistantMessage) isChatCompletionRequestMessage() {}
func (ChatCompletionRequestToolMessage) isChatCompletionRequestMessage()      {}

// FunctionObject represents the structure for function objects.
type FunctionObject struct {
	// Description is the function description
	Description string `json:"description,omitempty"`
	// Name is the function name
	Name string `json:"name"`
	// Parameters are the function parameters
	Parameters interface{} `json:"parameters,omitempty"`
}

// ChatCompletionNamedToolChoice represents the structure for specific tool choices.
type ChatCompletionNamedToolChoice struct {
	// Type is the tool type
	Type string `json:"type"`
	// Function is the function object
	Function FunctionObject `json:"function"`
}

// ChatCompletionToolChoiceOption represents the type for tool choice options.
type ChatCompletionToolChoiceOption interface{}

// ChatCompletionTool represents the structure for tools.
type ChatCompletionTool struct {
	// Type is the tool type
	Type string `json:"type"`
	// Function is the function object
	Function FunctionObject `json:"function"`
}

// ChatCompletionRequest represents the structure for chat completion requests.
type ChatCompletionRequest struct {
	// Model is the model to use for completion
	Model string `json:"model"`
	// Messages is the list of messages in the conversation
	Messages []ChatCompletionRequestMessage `json:"messages"`
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature is the sampling temperature
	Temperature float64 `json:"temperature,omitempty"`
	// ToolChoice is the tool choice option
	ToolChoice ChatCompletionToolChoiceOption `json:"tool_choice,omitempty"`
	// Tools is the list of tools available
	Tools []ChatCompletionTool `json:"tools,omitempty"`
	// Stream indicates whether to stream the response
	Stream bool `json:"stream,omitempty"`
}

// Validate checks if the ChatCompletionRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *ChatCompletionRequest) Validate() error {
	if r.Model == "" {
		return &ValidationError{
			Field:   "Model",
			Message: "model is required",
		}
	}

	if len(r.Messages) == 0 {
		return &ValidationError{
			Field:   "Messages",
			Message: "at least one message is required",
		}
	}

	return nil
}

// ChatCompletionResponseChoiceMessage represents the structure for chat completion response choice messages.
type ChatCompletionResponseChoiceMessage struct {
	// Role is the message role
	Role ChatCompletionMessageRoleType `json:"role"`
	// Content is the message content
	Content string `json:"content"`
}

// ChatCompletionResponse represents the structure for chat completion responses.
type ChatCompletionResponse struct {
	// ID is the unique identifier for the response
	ID string `json:"id"`
	// Object is the object type
	Object string `json:"object"`
	// Created is the timestamp when the response was created
	Created int64 `json:"created"`
	// Model is the model used for completion
	Model string `json:"model"`
	// Choices is the list of response choices
	Choices []struct {
		// Index is the index of the choice
		Index int `json:"index"`
		// Message is the response message
		Message ChatCompletionResponseChoiceMessage `json:"message"`
		// FinishReason is the reason the completion finished
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	// Usage contains token usage information
	Usage struct {
		// PromptTokens is the number of tokens in the prompt
		PromptTokens int `json:"prompt_tokens"`
		// CompletionTokens is the number of tokens in the completion
		CompletionTokens int `json:"completion_tokens"`
		// TotalTokens is the total number of tokens
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

// CreateChatCompletion creates a chat completion.
// This method sends a chat completion request to the Sakura AI Engine API and returns the response.
//
// Parameters:
//   - ctx: The context for the request
//   - req: The chat completion request
//
// Returns:
//
//	*ChatCompletionResponse: The chat completion response
//	error: An error if the request fails
func (c *SakuraClient) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	url := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
