package aiengine_test

import (
	"fmt"
	"log"

	"github.com/masa23/aiengine-go"
)

// ExampleRAG demonstrates how to use the RAG (Retrieval-Augmented Generation) functionality.
func Example_rag() {
	// Upload document
	uploadReq := &aiengine.DocumentUpload{
		Name:  "example-document",
		File:  "/path/to/your/document.txt",
		Model: "multilingual-e5-large",
	}

	// Validate the request
	if err := uploadReq.Validate(); err != nil {
		log.Fatal(err)
	}

	// Chat with documents
	chatReq := &aiengine.ChatRequest{
		ChatModel: "gpt-oss-120b",
		Query:     "What is in the document?",
	}

	// Validate the request
	if err := chatReq.Validate(); err != nil {
		log.Fatal(err)
	}

	// Show what the requests look like
	fmt.Printf("Upload document: %s\n", uploadReq.Name)
	fmt.Printf("Chat query: %s\n", chatReq.Query)
	// Output: Upload document: example-document
	// Chat query: What is in the document?
}
