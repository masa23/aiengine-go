package aiengine_test

import (
	"fmt"

	"github.com/masa23/aiengine-go"
)

// ExampleTranscription demonstrates how to use the audio transcription API.
func Example_transcription() {
	transReq := &aiengine.TranscriptionRequest{
		File:  "/path/to/your/audio.mp3",
		Model: "whisper-large-v3-turbo",
	}

	// Show what the request looks like
	fmt.Printf("File: %s\n", transReq.File)
	fmt.Printf("Model: %s\n", transReq.Model)
	// Output: File: /path/to/your/audio.mp3
	// Model: whisper-large-v3-turbo
}
