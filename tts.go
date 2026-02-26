// Package aiengine provides a client library for interacting with the Sakura AI Engine API.
//
// This package includes clients and data structures for accessing text-to-speech (TTS) functionality.
// The TTS functionality converts text to audio using the VOICEVOX-compatible interface.
//
// Basic usage:
//
//	req := &aiengine.TtsAudioQueryRequest{
//	    Text: "こんにちは。",
//	    Speaker: 1,
//	}
//	query, err := client.CreateAudioQuery(context.Background(), req)
//
//	audioData, err := client.SynthesizeTtsSpeech(context.Background(), &aiengine.TtsSynthesisRequest{
//	    Speaker: 1,
//	    Query: query,
//	})
package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"unicode/utf8"
)

// TtsAccentPhrase represents an accent phrase in the TTS output.
// An accent phrase is a unit of speech with a specific intonation pattern.
type TtsAccentPhrase struct {
	// Moras is the sequence of moras (syllables) in the accent phrase
	Moras []json.RawMessage `json:"moras"`
	// Accent is the accent position within the phrase
	Accent int `json:"accent"`
	// PauseMora is the silent mora (pause) between phrases, if any
	PauseMora *json.RawMessage `json:"pause_mora,omitempty"`
	// IsInterrogative indicates whether the phrase is a question
	IsInterrogative bool `json:"is_interrogative"`
}

// TtsAudioQuery represents the audio query response for TTS synthesis.
// This structure contains all the parameters needed for speech synthesis.
type TtsAudioQuery struct {
	// AccentPhrases is the list of accent phrases in the text
	AccentPhrases []TtsAccentPhrase `json:"accent_phrases"`
	// SpeedScale controls the overall speech speed (1.0 is normal)
	SpeedScale float64 `json:"speedScale"`
	// PitchScale controls the overall pitch (0.0 is normal)
	PitchScale float64 `json:"pitchScale"`
	// IntonationScale controls the overall intonation (1.0 is normal)
	IntonationScale float64 `json:"intonationScale"`
	// VolumeScale controls the overall volume (1.0 is normal)
	VolumeScale float64 `json:"volumeScale"`
	// PrePhonemeLength is the silence duration before speech in seconds
	PrePhonemeLength float64 `json:"prePhonemeLength"`
	// PostPhonemeLength is the silence duration after speech in seconds
	PostPhonemeLength float64 `json:"postPhonemeLength"`
	// PauseLength is the pause duration for punctuation marks (null = ignored)
	PauseLength *float64 `json:"pauseLength,omitempty"`
	// PauseLengthScale is the pause duration multiplier for punctuation marks
	PauseLengthScale float64 `json:"pauseLengthScale"`
	// OutputSamplingRate is the audio output sampling rate in Hz
	OutputSamplingRate int `json:"outputSamplingRate"`
	// OutputStereo indicates whether to output stereo audio
	OutputStereo bool `json:"outputStereo"`
	// Kana is the reading of the text in AquesTalk notation (read-only)
	Kana string `json:"kana"`
}

// TtsAudioQueryRequest represents the request structure for creating an audio query.
type TtsAudioQueryRequest struct {
	// Text is the text to convert to speech (1-1000 characters)
	Text string
	// Speaker is the speaker/style ID
	Speaker int
	// EnableKatakanaEnglish enables katakana English (default: true)
	EnableKatakanaEnglish *bool
	// CoreVersion is the TTS core version (currently ignored if specified)
	CoreVersion string
}

// Validate checks if the TtsAudioQueryRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *TtsAudioQueryRequest) Validate() error {
	if r.Text == "" {
		return &ValidationError{
			Field:   "Text",
			Message: "text is required",
		}
	}

	if utf8.RuneCountInString(r.Text) < 1 || utf8.RuneCountInString(r.Text) > 1000 {
		return &ValidationError{
			Field:   "Text",
			Message: "text length must be between 1 and 1000 characters",
		}
	}

	if r.Speaker < 0 {
		return &ValidationError{
			Field:   "Speaker",
			Message: "speaker must be >= 0",
		}
	}

	return nil
}

// CreateAudioQuery creates an audio query for TTS synthesis.
// This endpoint generates a JSON query that can be used with the synthesis endpoint.
//
// Typical workflow:
//  1. Call CreateAudioQuery to create the query
//  2. Optionally modify the query parameters (speed, pitch, etc.)
//  3. Call SynthesizeTtsSpeech with the query to generate audio
//
// The API provides a VOICEVOX Engine API compatible interface.
// Official specification: https://voicevox.github.io/voicevox_engine/api/
func (c *Client) CreateAudioQuery(ctx context.Context, req *TtsAudioQueryRequest) (*TtsAudioQuery, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/tts/v1/audio_query", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Set("text", req.Text)
	params.Set("speaker", fmt.Sprintf("%d", req.Speaker))
	if req.EnableKatakanaEnglish != nil {
		params.Set("enable_katakana_english", fmt.Sprintf("%t", *req.EnableKatakanaEnglish))
	}
	if req.CoreVersion != "" {
		params.Set("core_version", req.CoreVersion)
	}

	// Create request
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	respBody, err := c.doRequest(httpReq)
	if err != nil {
		return nil, err
	}

	// Parse response
	var query TtsAudioQuery
	if err := json.Unmarshal(respBody, &query); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &query, nil
}

// SpeechRequest represents the request structure for the Create speech API.
type SpeechRequest struct {
	// Model is the TTS model identifier (e.g., "zundamon")
	Model string
	// Input is the text to synthesize (up to ~1000 characters)
	Input string
	// Voice is the speaker/style (e.g., "normal")
	Voice string
	// Instructions is additional instructions (e.g., tone of speech)
	// Note: Currently ignored if specified
	Instructions string
	// ResponseFormat is the output format
	// Note: Currently always returns wav regardless of specification
	ResponseFormat string
	// StreamFormat is the stream format
	// Note: Currently ignored if specified
	StreamFormat string
}

// Validate checks if the SpeechRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *SpeechRequest) Validate() error {
	if r.Model == "" {
		return &ValidationError{
			Field:   "Model",
			Message: "model is required",
		}
	}

	if r.Input == "" {
		return &ValidationError{
			Field:   "Input",
			Message: "input is required",
		}
	}

	if len(r.Input) < 1 || len(r.Input) > 1000 {
		return &ValidationError{
			Field:   "Input",
			Message: "input length must be between 1 and 1000 characters",
		}
	}

	return nil
}

// CreateSpeech generates speech audio from text.
// This is the text-to-speech endpoint that directly converts text to audio.
//
// Required parameters: input, model
// Optional parameters: voice, instructions, response_format, stream_format
//
// Note:
// - instructions is currently ignored if specified
// - response_format is currently ignored (always returns wav)
// - streaming is not supported yet
//
// Returns:
//   - []byte: The audio data in WAV format
//   - error: An error if the request fails
func (c *Client) CreateSpeech(ctx context.Context, req *SpeechRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/v1/audio/speech", c.baseURL)

	// Build request body as JSON
	reqBody := struct {
		Model          string `json:"model"`
		Input          string `json:"input"`
		Voice          string `json:"voice,omitempty"`
		Instructions   string `json:"instructions,omitempty"`
		ResponseFormat string `json:"response_format,omitempty"`
		StreamFormat   string `json:"stream_format,omitempty"`
	}{
		Model:          req.Model,
		Input:          req.Input,
		Voice:          req.Voice,
		Instructions:   req.Instructions,
		ResponseFormat: req.ResponseFormat,
		StreamFormat:   req.StreamFormat,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request directly to handle binary response
	c.setAuthorizationHeader(httpReq)
	c.setUserAgentHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, &APIError{
			StatusCode: httpResp.StatusCode,
			Message:    httpResp.Status,
			RawBody:    body,
		}
	}

	// Read binary audio data
	audioData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return audioData, nil
}

// TtsSynthesisRequest represents the request structure for TTS synthesis.
type TtsSynthesisRequest struct {
	// Speaker is the speaker/style ID
	Speaker int
	// EnableInterrogativeUpspeak automatically adjusts tone for questions (default: true)
	EnableInterrogativeUpspeak *bool
	// CoreVersion is the TTS core version (currently ignored if specified)
	CoreVersion string
	// Query is the audio query to synthesize (required)
	Query *TtsAudioQuery
}

// Validate checks if the TtsSynthesisRequest is valid.
// This method validates that required fields are present and have valid values.
//
// Returns:
//   - error: A ValidationError if the request is invalid, nil otherwise
func (r *TtsSynthesisRequest) Validate() error {
	if r.Speaker < 0 {
		return &ValidationError{
			Field:   "Speaker",
			Message: "speaker must be >= 0",
		}
	}

	if r.Query == nil {
		return &ValidationError{
			Field:   "Query",
			Message: "query is required",
		}
	}

	return nil
}

// SynthesizeTtsSpeech generates speech audio from an audio query.
// This endpoint takes a JSON audio query and returns audio data in WAV format.
//
// The audio query should first be created using CreateAudioQuery.
// This API provides a VOICEVOX Engine API compatible interface.
// Official specification: https://voicevox.github.io/voicevox_engine/api/
//
// Returns:
//   - []byte: The audio data in WAV format
//   - error: An error if the request fails
func (c *Client) SynthesizeTtsSpeech(ctx context.Context, req *TtsSynthesisRequest) ([]byte, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("%s/tts/v1/synthesis", c.baseURL)

	// Build query parameters
	params := url.Values{}
	params.Set("speaker", fmt.Sprintf("%d", req.Speaker))
	if req.EnableInterrogativeUpspeak != nil {
		params.Set("enable_interrogative_upspeak", fmt.Sprintf("%t", *req.EnableInterrogativeUpspeak))
	}
	if req.CoreVersion != "" {
		params.Set("core_version", req.CoreVersion)
	}

	// Marshal query to JSON
	queryJSON, err := json.Marshal(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Create request
	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())
	httpReq, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request directly to handle binary response
	c.setAuthorizationHeader(httpReq)
	c.setUserAgentHeader(httpReq)

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		body, _ := io.ReadAll(httpResp.Body)
		return nil, &APIError{
			StatusCode: httpResp.StatusCode,
			Message:    httpResp.Status,
			RawBody:    body,
		}
	}

	// Read binary audio data
	audioData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return audioData, nil
}
