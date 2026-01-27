package aiengine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"testing"
	"time"
)

func integrationClientOrSkip(t *testing.T) *SakuraClient {
	t.Helper()

	apiKey := os.Getenv("SAKURA_AI_ENGINE_API_KEY")
	if apiKey == "" {
		t.Skip("SAKURA_AI_ENGINE_API_KEY is not set; skipping integration test")
	}

	// Create client with options for better reliability
	c := NewSakuraClient(apiKey,
		WithMaxRetries(3),
		WithRetryBackoff(1*time.Second))

	if base := os.Getenv("SAKURA_AI_ENGINE_BASE_URL"); base != "" {
		c.baseURL = base
	}
	return c
}

func envOrSkip(t *testing.T, key string) string {
	t.Helper()
	v := os.Getenv(key)
	if v == "" {
		t.Skip(key + " is not set; skipping integration test")
	}
	return v
}

func randomTag() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return "aiengine-go-test-" + hex.EncodeToString(b)
}

func TestIntegration_ChatCompletions(t *testing.T) {
	c := integrationClientOrSkip(t)
	chatModel := os.Getenv("SAKURA_AI_ENGINE_CHAT_MODEL")
	if chatModel == "" {
		chatModel = "gpt-oss-120b"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := &ChatCompletionRequest{
		Model: chatModel,
		Messages: []ChatCompletionRequestMessage{
			&ChatCompletionRequestUserMessage{
				Role:    ChatCompletionMessageRoleTypeUser,
				Content: "こんにちは。1+1は？",
			},
		},
	}

	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("CreateChatCompletion: %v", err)
	}
	if resp == nil || len(resp.Choices) == 0 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestIntegration_Embeddings(t *testing.T) {
	c := integrationClientOrSkip(t)
	embedModel := os.Getenv("SAKURA_AI_ENGINE_EMBEDDING_MODEL")
	if embedModel == "" {
		embedModel = "multilingual-e5-large"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req := &EmbeddingRequest{
		Model: embedModel,
		Input: "これはテストです。",
	}

	resp, err := c.CreateEmbeddings(ctx, req)
	if err != nil {
		t.Fatalf("CreateEmbeddings: %v", err)
	}
	if resp == nil || len(resp.Data) == 0 || len(resp.Data[0].Embedding) == 0 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestIntegration_Transcription(t *testing.T) {
	c := integrationClientOrSkip(t)
	audioPath := envOrSkip(t, "SAKURA_AI_ENGINE_AUDIO_FILE")
	model := os.Getenv("SAKURA_AI_ENGINE_TRANSCRIPTION_MODEL")
	if model == "" {
		model = "whisper-large-v3-turbo"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req := &TranscriptionRequest{
		File:     audioPath,
		Model:    model,
		Language: "ja",
	}

	resp, err := c.CreateTranscription(ctx, req)
	if err != nil {
		t.Fatalf("CreateTranscription: %v", err)
	}
	if resp == nil || resp.Text == "" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestIntegration_RAG_Upload_Query_Chat_Delete(t *testing.T) {
	c := integrationClientOrSkip(t)

	// RAG embedding model for indexing documents
	ragEmbedModel := os.Getenv("SAKURA_AI_ENGINE_RAG_EMBEDDING_MODEL")
	if ragEmbedModel == "" {
		ragEmbedModel = "multilingual-e5-large"
	}

	chatModel := envOrSkip(t, "SAKURA_AI_ENGINE_CHAT_MODEL")
	tag := randomTag()

	// Create a temp text file to upload.
	tmp, err := os.CreateTemp("", "aiengine-go-doc-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer os.Remove(tmp.Name())

	docContent := "これはGoクライアントのRAG統合テストです。合言葉: SAKURA_RAG_TEST_TOKEN。\n"
	if _, err := tmp.WriteString(docContent); err != nil {
		t.Fatalf("WriteString: %v", err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// Upload
	up, err := c.UploadDocument(ctx, &DocumentUpload{
		File:  tmp.Name(),
		Name:  "aiengine-go integration test",
		Tags:  []string{tag},
		Model: ragEmbedModel,
	})
	if err != nil {
		t.Fatalf("UploadDocument: %v", err)
	}
	if up == nil || up.ID == "" {
		t.Fatalf("unexpected upload response: %#v", up)
	}

	// Ensure document is deleted even if test fails
	defer func() {
		// Delete all documents with the name "aiengine-go integration test"
		err := c.DeleteDocumentsByName(context.Background(), "aiengine-go integration test")
		if err != nil {
			t.Logf("Warning: Failed to delete documents by name: %v", err)
		} else {
			t.Log("Successfully deleted documents by name")
		}
	}()

	// Poll until available (or error)
	deadline := time.Now().Add(60 * time.Second)
	for {
		if time.Now().After(deadline) {
			t.Fatalf("document not available in time: id=%s status=%s", up.ID, up.Status)
		}
		d, err := c.GetDocument(ctx, up.ID)
		if err != nil {
			t.Fatalf("GetDocument: %v", err)
		}
		if d.Status == "available" {
			break
		}
		if d.Status == "error" {
			t.Fatalf("document status=error: %#v", d)
		}
		time.Sleep(5 * time.Second) // Poll every 5 seconds
	}

	// Query (should retrieve something)
	qr, err := c.QueryDocuments(ctx, &QueryRequest{
		Model: ragEmbedModel,
		Query: "合言葉は？",
		Tags:  []string{tag},
		TopK:  3,
	})
	if err != nil {
		t.Fatalf("QueryDocuments: %v", err)
	}
	if qr == nil {
		t.Fatalf("unexpected query response: %#v", qr)
	}
	if len(qr.Results) == 0 {
		t.Fatalf("unexpected query response: %#v", qr)
	}

	// Chat with documents
	cr, err := c.ChatWithDocuments(ctx, &ChatRequest{
		Model:     ragEmbedModel,
		ChatModel: chatModel,
		Query:     "アップロードした文書の合言葉をそのまま答えてください。",
		Tags:      []string{tag},
		TopK:      3,
	})
	if err != nil {
		t.Fatalf("ChatWithDocuments: %v", err)
	}
	if cr == nil || cr.Answer == "" {
		t.Fatalf("unexpected chat response: %#v", cr)
	}
}
