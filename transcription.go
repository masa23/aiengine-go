// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing audio transcription functionality.
// The transcription functionality converts audio files to text.
//
// Basic usage:
//
//	req := &aiengine.TranscriptionRequest{
//	    File: "path/to/your/audio.mp3",
//	    Model: "whisper-large-v3-turbo",
//	}
//	resp, err := client.CreateTranscription(context.Background(), req)
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// TranscriptionResponse represents the response structure for audio transcription.
type TranscriptionResponse struct {
	// Model is the model used for transcription
	Model string `json:"model"`
	// Text is the transcribed text
	Text string `json:"text"`
}

// TranscriptionRequest represents the request structure for audio transcription.
type TranscriptionRequest struct {
	// File is the path to the audio file to transcribe
	File string
	// Model is the model to use for transcription
	Model string
	// Language is the language of the audio (optional)
	Language string
	// Prompt is the prompt to guide the transcription (optional)
	Prompt string
	// Temperature is the sampling temperature (optional)
	Temperature float64
	// Stream indicates whether to stream the response (optional)
	Stream bool
}

// Validate checks if the TranscriptionRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *TranscriptionRequest) Validate() error {
	if r.File == "" {
		return &ValidationError{
			Field:   "File",
			Message: "file is required",
		}
	}

	// Check if file exists
	if _, err := os.Stat(r.File); os.IsNotExist(err) {
		return &ValidationError{
			Field:   "File",
			Message: fmt.Sprintf("file does not exist: %s", r.File),
		}
	}

	return nil
}

// CreateTranscription は音声ファイルの書き起こしを行います
func (c *Client) CreateTranscription(ctx context.Context, req *TranscriptionRequest) (*TranscriptionResponse, error) {
	url := fmt.Sprintf("%s/v1/audio/transcriptions", c.baseURL)

	// ファイルを開く
	file, err := os.Open(req.File)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// マルチパートフォームを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// ファイルフィールドを追加
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(req.File))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// ファイル内容をコピー
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// その他のフィールドを追加
	if req.Model != "" {
		_ = writer.WriteField("model", req.Model)
	}
	if req.Language != "" {
		_ = writer.WriteField("language", req.Language)
	}
	if req.Prompt != "" {
		_ = writer.WriteField("prompt", req.Prompt)
	}
	if req.Temperature != 0 {
		_ = writer.WriteField("temperature", fmt.Sprintf("%f", req.Temperature))
	}
	if req.Stream {
		_ = writer.WriteField("stream", "true")
	}

	// マルチパートフォームを閉じる
	err = writer.Close()
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
	var resp TranscriptionResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}
