package aiengine_test

import (
	"fmt"
	"log"

	"github.com/masa23/aiengine-go"
)

// ExampleChatCompletion demonstrates how to use the chat completion API.
func Example_chatCompletion() {
	// Create a request
	req := &aiengine.ChatCompletionRequest{
		Model: "gpt-oss-120b",
		Messages: []aiengine.ChatCompletionRequestMessage{
			&aiengine.ChatCompletionRequestUserMessage{
				Role:    aiengine.ChatCompletionMessageRoleTypeUser,
				Content: "Hello, how are you?",
			},
		},
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		log.Fatal(err)
	}

	// Show what the request looks like
	fmt.Printf("Model: %s\n", req.Model)
	fmt.Printf("Message count: %d\n", len(req.Messages))
	// Output: Model: gpt-oss-120b
	// Message count: 1
}
