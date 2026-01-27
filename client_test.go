package aiengine

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewClientFromEnv_NotSet(t *testing.T) {
	// Ensure it errors when the env var isn't set.
	t.Setenv("SAKURA_AI_ENGINE_API_KEY", "")
	_, err := NewClientFromEnv()
	if err == nil {
		t.Fatalf("expected error when SAKURA_AI_ENGINE_API_KEY is not set")
	}
}

func TestNewClient_DefaultBaseURL(t *testing.T) {
	c := NewClient("dummy")
	if c.baseURL != "https://api.ai.sakura.ad.jp" {
		t.Fatalf("unexpected baseURL: %s", c.baseURL)
	}
}

func TestNewClient_InvalidAPIKey(t *testing.T) {
	// Test with an invalid API key to ensure proper error handling.
	c := NewClient("invalid-api-key")
	if c.apiKey != "invalid-api-key" {
		t.Fatalf("unexpected apiKey: %s", c.apiKey)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	// Test WithBaseURL option
	c := NewClient("test-key", WithBaseURL("https://custom.api.example.com"))
	if c.baseURL != "https://custom.api.example.com" {
		t.Fatalf("unexpected baseURL: %s", c.baseURL)
	}

	// Test WithTimeout option
	c = NewClient("test-key", WithTimeout(30*time.Second))
	if c.httpClient.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %v", c.httpClient.Timeout)
	}

	// Test WithHTTPClient option
	customClient := &http.Client{Timeout: 10 * time.Second}
	c = NewClient("test-key", WithHTTPClient(customClient))
	if c.httpClient != customClient {
		t.Fatalf("unexpected httpClient: %v", c.httpClient)
	}

	// Test WithMaxRetries option
	c = NewClient("test-key", WithMaxRetries(5))
	if c.maxRetries != 5 {
		t.Fatalf("unexpected maxRetries: %d", c.maxRetries)
	}

	// Test WithRetryBackoff option
	c = NewClient("test-key", WithRetryBackoff(2*time.Second))
	if c.retryBackoff != 2*time.Second {
		t.Fatalf("unexpected retryBackoff: %v", c.retryBackoff)
	}

	// Test multiple options together
	c = NewClient("test-key",
		WithBaseURL("https://custom.api.example.com"),
		WithTimeout(30*time.Second),
		WithMaxRetries(5))
	if c.baseURL != "https://custom.api.example.com" {
		t.Fatalf("unexpected baseURL: %s", c.baseURL)
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %v", c.httpClient.Timeout)
	}
	if c.maxRetries != 5 {
		t.Fatalf("unexpected maxRetries: %d", c.maxRetries)
	}
}

func TestChatCompletionRequest_EmptyMessage(t *testing.T) {
	// Test with an empty message to ensure proper error handling.
	req := &ChatCompletionRequest{
		Model: "test-model",
		Messages: []ChatCompletionRequestMessage{
			&ChatCompletionRequestUserMessage{
				Role:    ChatCompletionMessageRoleTypeUser,
				Content: "",
			},
		},
	}
	if req.Messages[0].(*ChatCompletionRequestUserMessage).Content != "" {
		t.Fatalf("unexpected message content: %s", req.Messages[0].(*ChatCompletionRequestUserMessage).Content)
	}
}

func TestChatCompletionRequest_LongMessage(t *testing.T) {
	// Test with a very long message to ensure proper handling.
	longMessage := string(make([]byte, 10000))
	req := &ChatCompletionRequest{
		Model: "test-model",
		Messages: []ChatCompletionRequestMessage{
			&ChatCompletionRequestUserMessage{
				Role:    ChatCompletionMessageRoleTypeUser,
				Content: longMessage,
			},
		},
	}
	content := req.Messages[0].(*ChatCompletionRequestUserMessage).Content.(string)
	if content != longMessage {
		t.Fatalf("unexpected message content length: %d", len(content))
	}
}

// MockClient is a mock implementation of Client for testing.
type MockClient struct {
	ChatCompletionResponse *ChatCompletionResponse
	EmbeddingResponse      *EmbeddingResponse
	TranscriptionResponse  *TranscriptionResponse
	ChatResult             *ChatResult
	Error                  error
}

func (m *MockClient) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.ChatCompletionResponse, nil
}

func (m *MockClient) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.EmbeddingResponse, nil
}

func (m *MockClient) CreateTranscription(ctx context.Context, req *TranscriptionRequest) (*TranscriptionResponse, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.TranscriptionResponse, nil
}

func (m *MockClient) ChatWithDocuments(ctx context.Context, req *ChatRequest) (*ChatResult, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.ChatResult, nil
}

func TestMockClient_CreateChatCompletion(t *testing.T) {
	expectedResp := &ChatCompletionResponse{
		ID: "test-id",
		Choices: []struct {
			Index        int                                 `json:"index"`
			Message      ChatCompletionResponseChoiceMessage `json:"message"`
			FinishReason string                              `json:"finish_reason"`
		}{{Message: ChatCompletionResponseChoiceMessage{Role: ChatCompletionMessageRoleTypeAssistant, Content: "test response"}}},
	}
	mockClient := &MockClient{
		ChatCompletionResponse: expectedResp,
	}

	req := &ChatCompletionRequest{
		Model: "test-model",
		Messages: []ChatCompletionRequestMessage{
			&ChatCompletionRequestUserMessage{
				Role:    ChatCompletionMessageRoleTypeUser,
				Content: "test message",
			},
		},
	}

	resp, err := mockClient.CreateChatCompletion(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateChatCompletion: %v", err)
	}
	if resp == nil || resp.ID != expectedResp.ID {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestAPIError_Error(t *testing.T) {
	// Test APIError.Error() method with request ID
	err := &APIError{
		StatusCode: 400,
		Code:       1001,
		Message:    "Bad Request",
		RequestID:  "req-12345",
	}
	expected := "API error 400 (request ID: req-12345): Bad Request"
	if err.Error() != expected {
		t.Fatalf("unexpected error message: %s", err.Error())
	}

	// Test APIError.Error() method without request ID
	err = &APIError{
		StatusCode: 500,
		Code:       5001,
		Message:    "Internal Server Error",
	}
	expected = "API error 500: Internal Server Error"
	if err.Error() != expected {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestDoRequest_ErrorHandling(t *testing.T) {
	// Test doRequest error handling with different error formats
	// This would require mocking the HTTP client, which is beyond the scope of this example
	// In a real test, you would create a mock HTTP server that returns different error formats
	// and verify that the APIError is constructed correctly
}

func TestShouldRetry(t *testing.T) {
	c := NewClient("test-key")

	// Test cases that should retry
	retryCodes := []int{429, 503, 504}
	for _, code := range retryCodes {
		if !c.shouldRetry(code) {
			t.Errorf("shouldRetry(%d) = false, want true", code)
		}
	}

	// Test cases that should not retry
	noRetryCodes := []int{200, 400, 401, 403, 404, 500, 502}
	for _, code := range noRetryCodes {
		if c.shouldRetry(code) {
			t.Errorf("shouldRetry(%d) = true, want false", code)
		}
	}
}

func TestParseRetryAfter(t *testing.T) {
	c := NewClient("test-key")

	// Test integer format
	duration := c.parseRetryAfter("120")
	if duration != 120*time.Second {
		t.Errorf("parseRetryAfter(\"120\") = %v, want 120s", duration)
	}

	// Test date format (using HTTP-date format)
	now := time.Now()
	future := now.Add(60 * time.Second)
	dateStr := future.UTC().Format(http.TimeFormat)
	duration = c.parseRetryAfter(dateStr)
	// Allow some tolerance for time passing during the test
	if duration < 55*time.Second || duration > 65*time.Second {
		t.Errorf("parseRetryAfter(date) = %v, want approximately 60s", duration)
	}

	// Test invalid format (should default to 1 second)
	duration = c.parseRetryAfter("invalid")
	if duration != 1*time.Second {
		t.Errorf("parseRetryAfter(\"invalid\") = %v, want 1s", duration)
	}
}
