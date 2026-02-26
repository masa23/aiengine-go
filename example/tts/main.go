package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/masa23/aiengine-go"
)

// Example 1: CreateSpeech APIを使った簡易的な音声合成
func main() {
	// クライアントの初期化（APIキーは環境変数 SAKURA_AI_ENGINE_API_KEY から取得）
	client, err := aiengine.NewClientFromEnv()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// 音声合成リクエストの作成
	req := &aiengine.SpeechRequest{
		Model: "zundamon",          // 音声モデル（例: zundamon）
		Input: "こんにちは。今日はいい天気ですね。", // 合成するテキスト
		Voice: "normal",            // 音声スタイル（例: normal）
	}

	// バリデーション
	if err := req.Validate(); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	// 音声合成の実行
	audioData, err := client.CreateSpeech(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create speech: %v", err)
	}

	// 音声ファイルの保存
	outputFile := "output.wav"
	if err := os.WriteFile(outputFile, audioData, 0644); err != nil {
		log.Fatalf("Failed to write audio file: %v", err)
	}

	fmt.Printf("音声ファイルを保存しました: %s (%d bytes)\n", outputFile, len(audioData))
}
