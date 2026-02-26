package main

import (
	"context"
	"fmt"
	"log"

	"github.com/masa23/aiengine-go"
)

// Example 1: ChatCompletion APIを使ったチャット
func main() {
	// クライアントの初期化（APIキーは環境変数 SAKURA_AI_ENGINE_API_KEY から取得）
	client, err := aiengine.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// チャットリクエストの作成
	req := &aiengine.ChatCompletionRequest{
		Model: "gpt-oss-120b", // モデル名
		Messages: []aiengine.ChatCompletionRequestMessage{
			&aiengine.ChatCompletionRequestUserMessage{
				Role:    aiengine.ChatCompletionMessageRoleTypeUser,
				Content: "こんにちは。日本語で自己紹介をお願いします。",
			},
		},
	}

	// バリデーション
	if err := req.Validate(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// チャットの実行
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create chat completion: %v", err)
	}

	// 応答の表示
	if len(resp.Choices) > 0 {
		fmt.Printf("AIの応答:\n%s\n", resp.Choices[0].Message.Content)
	} else {
		fmt.Println("応答がありません")
	}
}
