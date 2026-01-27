// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing various AI services
// such as chat completions, embeddings, and RAG (Retrieval-Augmented Generation) functionality.
//
// Basic usage:
//
//	client := aiengine.NewClient("your-api-key")
//	// Or create from environment variables
//	client, err := aiengine.NewClientFromEnv()
//
// For chat completions:
//
//	req := &aiengine.ChatCompletionRequest{
//	    Model: "gpt-oss-120b",
//	    Messages: []aiengine.ChatMessage{
//	        {Role: "user", Content: "Hello!"},
//	    },
//	}
//	resp, err := client.CreateChatCompletion(context.Background(), req)
//
// For embeddings:
//
//	req := &aiengine.EmbeddingRequest{
//	    Input: "Your text here",
//	    Model: "multilingual-e5-large",
//	}
//	resp, err := client.CreateEmbedding(context.Background(), req)
//
// For RAG operations:
//
//	docs, err := client.ListDocuments(context.Background(), "", "", "", 0, 10)
package aiengine

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// APIError represents a custom error type for API errors.
// It contains detailed information about the error including status code, message, and request ID.
type APIError struct {
	// StatusCode is the HTTP status code returned by the API
	StatusCode int
	// Code is the API-specific error code (if available)
	Code int `json:"code"`
	// Message is the human-readable error message
	Message string `json:"message"`
	// RequestID is the unique identifier for the request (from response headers)
	RequestID string
	// RawBody contains the raw response body
	RawBody []byte
}

// Error returns the formatted error message for the APIError.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("API error %d (request ID: %s): %s", e.StatusCode, e.RequestID, e.Message)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// ValidationError represents a custom error type for request validation errors.
// It contains information about which field failed validation and why.
type ValidationError struct {
	// Field is the name of the field that failed validation
	Field string `json:"field"`
	// Message is the human-readable validation error message
	Message string `json:"message"`
}

// Error returns the formatted error message for the ValidationError.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// Client is the client for the Sakura AI Engine API.
// It handles authentication, HTTP requests, and response processing for all API operations.
type Client struct {
	// apiKey is the API key used for authentication
	apiKey string
	// baseURL is the base URL for the API endpoints
	baseURL string
	// httpClient is the HTTP client used for making requests
	httpClient *http.Client
	// userAgent is the User-Agent header value
	userAgent string
	// maxRetries is the maximum number of retries for failed requests
	maxRetries int
	// retryBackoff is the base backoff duration for retries
	retryBackoff time.Duration
}

// ClientOption is a function that configures a Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPClient sets the HTTP client for the Client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithTimeout sets the timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithMaxRetries sets the maximum number of retries for failed requests.
func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
	}
}

// WithRetryBackoff sets the base backoff duration for retries.
func WithRetryBackoff(backoff time.Duration) ClientOption {
	return func(c *Client) {
		c.retryBackoff = backoff
	}
}

// NewClient creates a new Client instance with the provided API key and options.
// The client will use the default base URL (https://api.ai.sakura.ad.jp) and
// an HTTP client with a 60-second timeout if no options are provided.
//
// Parameters:
//   - apiKey: The API key for authenticating with the Sakura AI Engine API
//   - opts: Optional ClientOption functions to configure the client
//
// Returns:
//
//	*Client: A new Client instance
func NewClient(apiKey string, opts ...ClientOption) *Client {
	c := &Client{
		apiKey:       apiKey,
		baseURL:      "https://api.ai.sakura.ad.jp",
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		userAgent:    "aiengine-go/v0.1.0", // Default version, will be updated
		maxRetries:   3,                    // Default retry count
		retryBackoff: 1 * time.Second,      // Default backoff
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	// Set User-Agent with Go version info
	if c.userAgent == "aiengine-go/v0.1.0" {
		c.userAgent = fmt.Sprintf("aiengine-go/v0.1.0 (go/%s; %s/%s)",
			runtime.Version()[2:], runtime.GOOS, runtime.GOARCH)
	}

	return c
}

// NewClientFromEnv creates a client using environment variables.
// This function reads the API key and optional base URL from environment variables.
//
// Environment Variables:
//   - SAKURA_AI_ENGINE_API_KEY (required): The API key for authenticating with the API
//   - SAKURA_AI_ENGINE_BASE_URL (optional): The base URL for the API. Defaults to https://api.ai.sakura.ad.jp
//
// Returns:
//
//	*Client: A new Client instance
//	error: An error if the required API key environment variable is not set
func NewClientFromEnv(opts ...ClientOption) (*Client, error) {
	apiKey := os.Getenv("SAKURA_AI_ENGINE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("SAKURA_AI_ENGINE_API_KEY is not set")
	}

	// Create client with options
	c := NewClient(apiKey, opts...)

	// Override base URL if set in environment
	if base := os.Getenv("SAKURA_AI_ENGINE_BASE_URL"); base != "" {
		c.baseURL = base
	}

	return c, nil
}

// setAuthorizationHeader sets the authorization header on the request.
// This method adds the Bearer token authentication header using the client's API key.
//
// Parameters:
//   - req: The HTTP request to set the authorization header on
func (c *Client) setAuthorizationHeader(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
}

// setUserAgentHeader sets the User-Agent header on the request.
func (c *Client) setUserAgentHeader(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
}

// shouldRetry determines if a request should be retried based on the response status code.
func (c *Client) shouldRetry(statusCode int) bool {
	switch statusCode {
	case 429, 503, 504: // Too Many Requests, Service Unavailable, Gateway Timeout
		return true
	default:
		return false
	}
}

// parseRetryAfter parses the Retry-After header value and returns the duration.
func (c *Client) parseRetryAfter(retryAfter string) time.Duration {
	// First try to parse as integer (seconds)
	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Then try to parse as HTTP-date
	if date, err := http.ParseTime(retryAfter); err == nil {
		return time.Until(date)
	}

	// Default to 1 second if parsing fails
	return 1 * time.Second
}

// doRequest executes an HTTP request with proper authentication and error handling.
// This method sets the authorization header, executes the request, and handles the response.
// If the response status indicates an error, it attempts to parse the response as an APIError.
//
// Parameters:
//   - req: The HTTP request to execute
//
// Returns:
//
//	[]byte: The response body as bytes
//	error: An error if the request fails or the response indicates an error
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	c.setAuthorizationHeader(req)
	c.setUserAgentHeader(req)

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Clone the request for retry attempts
		clonedReq := req.Clone(req.Context())
		clonedReq.Header = req.Header.Clone()

		resp, err := c.httpClient.Do(clonedReq)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			if attempt < c.maxRetries {
				// Exponential backoff with jitter
				backoff := c.retryBackoff * time.Duration(1<<uint(attempt))
				time.Sleep(backoff + time.Duration(rand.Int63n(int64(100*time.Millisecond))))
				continue
			}
			return nil, lastErr
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			if attempt < c.maxRetries {
				// Exponential backoff with jitter
				backoff := c.retryBackoff * time.Duration(1<<uint(attempt))
				time.Sleep(backoff + time.Duration(rand.Int63n(int64(100*time.Millisecond))))
				continue
			}
			return nil, lastErr
		}

		// Check if we should retry based on status code
		if c.shouldRetry(resp.StatusCode) && attempt < c.maxRetries {
			// Parse Retry-After header if present
			retryAfter := resp.Header.Get("Retry-After")
			if retryAfter != "" {
				delay := c.parseRetryAfter(retryAfter)
				time.Sleep(delay)
			} else {
				// Exponential backoff with jitter
				backoff := c.retryBackoff * time.Duration(1<<uint(attempt))
				time.Sleep(backoff + time.Duration(rand.Int63n(int64(100*time.Millisecond))))
			}
			continue
		}

		// If successful (status 2xx), return the response
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return body, nil
		}

		// Handle error responses
		// Extract request ID from response headers
		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = resp.Header.Get("Request-ID")
		}

		// Create APIError with status code and request ID
		apiErr := APIError{
			StatusCode: resp.StatusCode,
			RequestID:  requestID,
			RawBody:    body,
		}

		// Try to parse API error in various formats
		var errorData map[string]interface{}
		if err := json.Unmarshal(body, &errorData); err == nil {
			// Handle different error formats
			if code, ok := errorData["code"]; ok {
				if codeNum, ok := code.(float64); ok {
					apiErr.Code = int(codeNum)
				}
			}

			if message, ok := errorData["message"]; ok {
				if messageStr, ok := message.(string); ok {
					apiErr.Message = messageStr
				}
			} else if errorMsg, ok := errorData["error"]; ok {
				// Handle {"error": {...}} format
				if errorObj, ok := errorMsg.(map[string]interface{}); ok {
					if message, ok := errorObj["message"]; ok {
						if messageStr, ok := message.(string); ok {
							apiErr.Message = messageStr
						}
					}
					if code, ok := errorObj["code"]; ok {
						if codeNum, ok := code.(float64); ok {
							apiErr.Code = int(codeNum)
						}
					}
				} else if errorStr, ok := errorMsg.(string); ok {
					// Handle {"error": "message"} format
					apiErr.Message = errorStr
				}
			}
		}

		// If we couldn't parse a specific message, use the raw body
		if apiErr.Message == "" {
			apiErr.Message = string(body)
		}

		lastErr = &apiErr
	}

	return nil, lastErr
}

// FindDocumentsByName searches for documents by name.
// This method is a convenience function that calls ListDocuments with the name parameter.
//
// Parameters:
//   - ctx: The context for the request
//   - name: The name of the documents to search for
//
// Returns:
//
//	[]DocumentList: A slice of DocumentList items matching the name
//	error: An error if the request fails
func (c *Client) FindDocumentsByName(ctx context.Context, name string) ([]DocumentList, error) {
	list, err := c.ListDocuments(ctx, "", name, "", 0, 0)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

// DeleteDocumentsByName deletes all documents with the specified name.
// This method first finds all documents matching the name, then attempts to delete each one.
// If deletion of any document fails, it logs a warning but continues deleting other documents.
//
// Parameters:
//   - ctx: The context for the request
//   - name: The name of the documents to delete
//
// Returns:
//
//	error: An error if finding the documents fails, nil otherwise (individual deletion errors are logged)
func (c *Client) DeleteDocumentsByName(ctx context.Context, name string) error {
	documents, err := c.FindDocumentsByName(ctx, name)
	if err != nil {
		return err
	}

	for _, doc := range documents {
		err := c.DeleteDocument(ctx, doc.ID)
		if err != nil {
			// Log the error but continue deleting other documents
			fmt.Printf("Warning: Failed to delete document %s: %v\n", doc.ID, err)
		}
	}

	return nil
}
