package aiengine_test

import (
	"fmt"
	"log"

	"github.com/masa23/aiengine-go"
)

// ExampleEmbedding demonstrates how to use the embeddings API.
func Example_embedding() {
	req := &aiengine.EmbeddingRequest{
		Model: "multilingual-e5-large",
		Input: "This is a test for embedding.",
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		log.Fatal(err)
	}

	// Show what the request looks like
	fmt.Printf("Model: %s\n", req.Model)
	fmt.Printf("Input: %v\n", req.Input)
	// Output: Model: multilingual-e5-large
	// Input: This is a test for embedding.
}
