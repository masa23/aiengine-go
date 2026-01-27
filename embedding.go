// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing embedding functionality.
// The embedding functionality generates vector representations from text.
//
// Basic usage:
//
//	req := &aiengine.EmbeddingRequest{
//	    Input: "Your text here",
//	    Model: "multilingual-e5-large",
//	}
//	resp, err := client.CreateEmbeddings(context.Background(), req)
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// EmbeddingResponse represents the response structure for embedding vectors.
type EmbeddingResponse struct {
	// Model is the model used for embedding
	Model string `json:"model"`
	// Data contains the embedding data
	Data []struct {
		// Index is the index of the embedding
		Index int `json:"index"`
		// Object is the object type
		Object string `json:"object"`
		// Embedding is the embedding vector
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

// TextInput represents a single text input.
type TextInput string

// TextInputs represents multiple text inputs.
type TextInputs []string

// EmbeddingRequest represents the request structure for embedding vectors.
type EmbeddingRequest struct {
	// Model is the model to use for embedding
	Model string `json:"model"`
	// Input is the text input (TextInput or TextInputs)
	Input interface{} `json:"input"`
}

// Validate checks if the EmbeddingRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *EmbeddingRequest) Validate() error {
	if r.Model == "" {
		return &ValidationError{
			Field:   "Model",
			Message: "model is required",
		}
	}

	if r.Input == nil {
		return &ValidationError{
			Field:   "Input",
			Message: "input is required",
		}
	}

	// Check if input is empty based on its type
	switch v := r.Input.(type) {
	case TextInput:
		if string(v) == "" {
			return &ValidationError{
				Field:   "Input",
				Message: "input text is empty",
			}
		}
	case TextInputs:
		if len(v) == 0 {
			return &ValidationError{
				Field:   "Input",
				Message: "input texts array is empty",
			}
		}
		for i, text := range v {
			if text == "" {
				return &ValidationError{
					Field:   "Input",
					Message: fmt.Sprintf("input text at index %d is empty", i),
				}
			}
		}
	case string:
		if v == "" {
			return &ValidationError{
				Field:   "Input",
				Message: "input text is empty",
			}
		}
	case []string:
		if len(v) == 0 {
			return &ValidationError{
				Field:   "Input",
				Message: "input texts array is empty",
			}
		}
		for i, text := range v {
			if text == "" {
				return &ValidationError{
					Field:   "Input",
					Message: fmt.Sprintf("input text at index %d is empty", i),
				}
			}
		}
	}

	return nil
}

// CreateEmbeddings generates embedding vectors from text.
// This method sends an embedding request to the Sakura AI Engine API and returns the response.
//
// Parameters:
//   - ctx: The context for the request
//   - req: The embedding request
//
// Returns:
//
//	*EmbeddingResponse: The embedding response
//	error: An error if the request fails
func (c *Client) CreateEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	url := fmt.Sprintf("%s/v1/embeddings", c.baseURL)

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

	var resp EmbeddingResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
