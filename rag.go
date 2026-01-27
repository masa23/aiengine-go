// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing RAG (Retrieval-Augmented Generation) functionality.
// The RAG functionality enables document uploading, searching, and chatting.
//
// Basic usage:
//
//	docs, err := client.ListDocuments(context.Background(), "", "", "", 0, 10)
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	urlpkg "net/url"
	"os"
	"path/filepath"
)

// Document represents the structure for a document.
type Document struct {
	// ID is the unique identifier for the document
	ID string `json:"id"`
	// CreatedAt is the timestamp when the document was created
	CreatedAt string `json:"created_at"`
	// Status is the current status of the document
	Status string `json:"status"`
	// Name is the name of the document
	Name string `json:"name"`
	// Model is the model used for the document
	Model string `json:"model"`
	// ChunkCount is the number of chunks in the document
	ChunkCount int `json:"chunk_count"`
	// Tags are the tags associated with the document
	Tags []string `json:"tags"`
	// Content is the content of the document
	Content string `json:"content"`
	// ErrorMessage is the error message if any
	ErrorMessage string `json:"error_message"`
}

// DocumentList represents the structure for a document list.
type DocumentList struct {
	// ID is the unique identifier for the document
	ID string `json:"id"`
	// CreatedAt is the timestamp when the document was created
	CreatedAt string `json:"created_at"`
	// Status is the current status of the document
	Status string `json:"status"`
	// Name is the name of the document
	Name string `json:"name"`
	// Model is the model used for the document
	Model string `json:"model"`
	// ChunkCount is the number of chunks in the document
	ChunkCount int `json:"chunk_count"`
	// Tags are the tags associated with the document
	Tags []string `json:"tags"`
	// ErrorMessage is the error message if any
	ErrorMessage string `json:"error_message"`
}

// PaginatedDocumentList represents the structure for a paginated document list.
type PaginatedDocumentList struct {
	// Meta contains pagination metadata
	Meta struct {
		// Page is the current page number
		Page int `json:"page"`
		// PageSize is the number of items per page
		PageSize int `json:"page_size"`
		// TotalPages is the total number of pages
		TotalPages int `json:"total_pages"`
		// Count is the total number of items
		Count int `json:"count"`
		// Next is the URL for the next page
		Next string `json:"next"`
		// Previous is the URL for the previous page
		Previous string `json:"previous"`
	} `json:"meta"`
	// Results contains the list of documents
	Results []DocumentList `json:"results"`
}

// DocumentChunk represents the structure for a document chunk.
type DocumentChunk struct {
	// Document is the parent document
	Document Document `json:"document"`
	// ChunkIndex is the index of the chunk within the document
	ChunkIndex int `json:"chunk_index"`
	// Content is the content of the chunk
	Content string `json:"content"`
	// Metadata is the metadata associated with the chunk
	Metadata string `json:"metadata"`
}

// PaginatedDocumentChunkList represents the structure for a paginated document chunk list.
type PaginatedDocumentChunkList struct {
	// Meta contains pagination metadata
	Meta struct {
		// Page is the current page number
		Page int `json:"page"`
		// PageSize is the number of items per page
		PageSize int `json:"page_size"`
		// TotalPages is the total number of pages
		TotalPages int `json:"total_pages"`
		// Count is the total number of items
		Count int `json:"count"`
		// Next is the URL for the next page
		Next string `json:"next"`
		// Previous is the URL for the previous page
		Previous string `json:"previous"`
	} `json:"meta"`
	// Results contains the list of document chunks
	Results []DocumentChunk `json:"results"`
}

// QueryResultChunkMetadata represents the type for query result chunk metadata.
type QueryResultChunkMetadata map[string]string

// QueryResultChunk represents the structure for a query result chunk.
type QueryResultChunk struct {
	// Document is the parent document
	Document Document `json:"document"`
	// ChunkIndex is the index of the chunk within the document
	ChunkIndex int `json:"chunk_index"`
	// Distance is the distance score of the chunk
	Distance float64 `json:"distance"`
	// Content is the content of the chunk
	Content string `json:"content"`
	// Metadata is the metadata associated with the chunk
	Metadata QueryResultChunkMetadata `json:"metadata"`
}

// QueryResultDocument represents the structure for a query result document.
type QueryResultDocument struct {
	// ID is the unique identifier for the document
	ID string `json:"id"`
	// Name is the name of the document
	Name string `json:"name"`
	// Tags are the tags associated with the document
	Tags []string `json:"tags"`
	// Model is the model used for the document
	Model string `json:"model"`
	// Distance is the distance score of the document
	Distance float64 `json:"distance"`
	// Content is the content of the document
	Content string `json:"content"`
}

// ChatResultSourceMetadata represents the type for chat result source metadata.
type ChatResultSourceMetadata map[string]string

// ChatResultSource represents the structure for a chat result source.
type ChatResultSource struct {
	// Document is the parent document
	Document Document `json:"document"`
	// ChunkIndex is the index of the chunk within the document
	ChunkIndex int `json:"chunk_index"`
	// Distance is the distance score of the source
	Distance float64 `json:"distance"`
	// Content is the content of the source
	Content string `json:"content"`
	// Metadata is the metadata associated with the source
	Metadata ChatResultSourceMetadata `json:"metadata"`
}

// ChatResult represents the structure for a chat result.
type ChatResult struct {
	// Answer is the chat response
	Answer string `json:"answer"`
	// Sources are the sources used to generate the answer
	Sources []ChatResultSource `json:"sources"`
}

// DocumentUpload represents the structure for a document upload.
type DocumentUpload struct {
	// ID is the unique identifier for the upload
	ID string `json:"id"`
	// Status is the current status of the upload
	Status string `json:"status"`
	// Content is the content of the document
	Content string `json:"content"`
	// File is the path to the file being uploaded
	File string `json:"file"`
	// Name is the name of the document
	Name string `json:"name"`
	// Tags are the tags associated with the document
	Tags []string `json:"tags"`
	// Model is the model used for the document
	Model string `json:"model"`
}

// Validate checks if the DocumentUpload is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *DocumentUpload) Validate() error {
	if r.Name == "" {
		return &ValidationError{
			Field:   "Name",
			Message: "name is required",
		}
	}

	if r.File == "" {
		return &ValidationError{
			Field:   "File",
			Message: "file is required",
		}
	}

	return nil
}

// ChatRequest represents the structure for a chat request.
type ChatRequest struct {
	// Model is the model to use for the document search
	Model string `json:"model,omitempty"`
	// ChatModel is the model to use for chat completion
	ChatModel string `json:"chat_model"`
	// Query is the query text
	Query string `json:"query"`
	// Prompt is the prompt template
	Prompt string `json:"prompt,omitempty"`
	// Tags are the tags to filter documents
	Tags []string `json:"tags,omitempty"`
	// TopK is the number of top results to return
	TopK int `json:"top_k,omitempty"`
	// Threshold is the similarity threshold
	Threshold float64 `json:"threshold,omitempty"`
	// UseFullContent indicates whether to use full document content
	UseFullContent bool `json:"use_full_content,omitempty"`
}

// Validate checks if the ChatRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *ChatRequest) Validate() error {
	if r.ChatModel == "" {
		return &ValidationError{
			Field:   "ChatModel",
			Message: "chat_model is required",
		}
	}

	if r.Query == "" {
		return &ValidationError{
			Field:   "Query",
			Message: "query is required",
		}
	}

	return nil
}

// QueryRequest represents the structure for a query request.
type QueryRequest struct {
	// Model is the model to use for the query
	Model string `json:"model,omitempty"`
	// Query is the query text
	Query string `json:"query"`
	// Tags are the tags to filter documents
	Tags []string `json:"tags,omitempty"`
	// TopK is the number of top results to return
	TopK int `json:"top_k,omitempty"`
	// Threshold is the similarity threshold
	Threshold float64 `json:"threshold,omitempty"`
}

// QueryResultChunkList represents the structure for a query result chunk list.
type QueryResultChunkList struct {
	// Results contains the list of query result chunks
	Results []QueryResultChunk `json:"results"`
}

// ListDocuments lists documents.
// This method retrieves a paginated list of documents with optional filtering by model, name, and tag.
//
// Parameters:
//   - ctx: The context for the request
//   - model: Filter documents by model (optional)
//   - name: Filter documents by name (optional)
//   - tag: Filter documents by tag (optional)
//   - page: Page number (optional, default: 1)
//   - pageSize: Number of items per page (optional, default: 10)
//
// Returns:
//
//	*PaginatedDocumentList: The paginated list of documents
//	error: An error if the request fails
func (c *SakuraClient) ListDocuments(ctx context.Context, model, name, tag string, page, pageSize int) (*PaginatedDocumentList, error) {
	url := fmt.Sprintf("%s/v1/documents/", c.baseURL)

	// Build query parameters
	q := urlpkg.Values{}
	if model != "" {
		q.Set("model", model)
	}
	if name != "" {
		q.Set("name", name)
	}
	if tag != "" {
		q.Set("tag", tag)
	}
	if page > 0 {
		q.Set("page", fmt.Sprintf("%d", page))
	}
	if pageSize > 0 {
		q.Set("page_size", fmt.Sprintf("%d", pageSize))
	}
	if len(q) > 0 {
		url += "?" + q.Encode()
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp PaginatedDocumentList
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// GetDocument retrieves a document by its ID.
// This method fetches the details of a specific document using its unique identifier.
//
// Parameters:
//   - ctx: The context for the request
//   - id: The unique identifier of the document to retrieve
//
// Returns:
//   - *Document: The retrieved document
//   - error: An error if the request fails
func (c *SakuraClient) GetDocument(ctx context.Context, id string) (*Document, error) {
	url := fmt.Sprintf("%s/v1/documents/%s/", c.baseURL, id)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp Document
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// UpdateDocument updates an existing document.
// This method updates the properties of a document identified by its ID.
//
// Parameters:
//   - ctx: The context for the request
//   - id: The unique identifier of the document to update
//   - doc: The document object containing updated information
//
// Returns:
//   - *Document: The updated document
//   - error: An error if the request fails
func (c *SakuraClient) UpdateDocument(ctx context.Context, id string, doc *Document) (*Document, error) {
	url := fmt.Sprintf("%s/v1/documents/%s/", c.baseURL, id)

	body, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp Document
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// DeleteDocument deletes a document by its ID.
// This method removes a document from the system using its unique identifier.
//
// Parameters:
//   - ctx: The context for the request
//   - id: The unique identifier of the document to delete
//
// Returns:
//   - error: An error if the request fails
func (c *SakuraClient) DeleteDocument(ctx context.Context, id string) error {
	url := fmt.Sprintf("%s/v1/documents/%s/", c.baseURL, id)

	httpReq, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	_, err = c.doRequest(httpReq)
	if err != nil {
		return err
	}

	return nil
}

// ListDocumentChunks lists the chunks of a document.
// This method retrieves a paginated list of chunks for a specific document.
//
// Parameters:
//   - ctx: The context for the request
//   - documentID: The unique identifier of the document
//   - page: Page number (optional, default: 1)
//   - pageSize: Number of items per page (optional, default: 10)
//
// Returns:
//   - *PaginatedDocumentChunkList: The paginated list of document chunks
//   - error: An error if the request fails
func (c *SakuraClient) ListDocumentChunks(ctx context.Context, documentID string, page, pageSize int) (*PaginatedDocumentChunkList, error) {
	url := fmt.Sprintf("%s/v1/documents/%s/chunks/", c.baseURL, documentID)

	// クエリパラメータを構築
	params := ""
	if page > 0 {
		params += fmt.Sprintf("page=%d&", page)
	}
	if pageSize > 0 {
		params += fmt.Sprintf("page_size=%d&", pageSize)
	}

	// 末尾の&を削除
	if len(params) > 0 {
		params = params[:len(params)-1]
		url += "?" + params
	}

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp PaginatedDocumentChunkList
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// GetDocumentChunk retrieves a specific chunk of a document.
// This method fetches a single chunk from a document using the document ID and chunk index.
//
// Parameters:
//   - ctx: The context for the request
//   - documentID: The unique identifier of the document
//   - index: The index of the chunk to retrieve
//
// Returns:
//   - *DocumentChunk: The retrieved document chunk
//   - error: An error if the request fails
func (c *SakuraClient) GetDocumentChunk(ctx context.Context, documentID string, index int) (*DocumentChunk, error) {
	url := fmt.Sprintf("%s/v1/documents/%s/chunks/%d/", c.baseURL, documentID, index)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var resp DocumentChunk
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// UploadDocument はドキュメントをアップロードします
func (c *SakuraClient) UploadDocument(ctx context.Context, uploadReq *DocumentUpload) (*DocumentUpload, error) {
	url := fmt.Sprintf("%s/v1/documents/upload/", c.baseURL)

	// マルチパートフォームを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// ファイルフィールドを追加
	if uploadReq.File != "" {
		fileWriter, err := writer.CreateFormFile("file", filepath.Base(uploadReq.File))
		if err != nil {
			return nil, fmt.Errorf("failed to create form file: %w", err)
		}

		// ファイルを開いて内容をコピー
		file, err := os.Open(uploadReq.File)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return nil, fmt.Errorf("failed to copy file: %w", err)
		}
	}

	// その他のフィールドを追加
	if uploadReq.Name != "" {
		_ = writer.WriteField("name", uploadReq.Name)
	}
	if len(uploadReq.Tags) > 0 {
		for _, t := range uploadReq.Tags {
			_ = writer.WriteField("tags", t)
		}
	}
	if uploadReq.Model != "" {
		_ = writer.WriteField("model", uploadReq.Model)
	}

	// マルチパートフォームを閉じる
	err := writer.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	// リクエストを作成
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	// リクエストを実行
	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	// レスポンスを解析
	var resp DocumentUpload
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// QueryDocuments はドキュメントをクエリします
func (c *SakuraClient) QueryDocuments(ctx context.Context, queryReq *QueryRequest) (*QueryResultChunkList, error) {
	url := fmt.Sprintf("%s/v1/documents/query/", c.baseURL)

	body, err := json.Marshal(queryReq)
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

	var resp QueryResultChunkList
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

// ChatWithDocuments はドキュメントとチャットします
func (c *SakuraClient) ChatWithDocuments(ctx context.Context, chatReq *ChatRequest) (*ChatResult, error) {
	url := fmt.Sprintf("%s/v1/documents/chat/", c.baseURL)

	body, err := json.Marshal(chatReq)
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

	var resp ChatResult
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
