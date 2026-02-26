# aiengine (Go client)

[![Go Reference](https://pkg.go.dev/badge/github.com/masa23/aiengine-go.svg)](https://pkg.go.dev/github.com/masa23/aiengine-go)

このライブラリは、さくらのAI Engine APIのGo言語用クライアントです。チャット補完、埋め込みベクトル、RAG（Retrieval-Augmented Generation）、音声書き起こし、テキスト読み上げ（TTS）などの機能を利用できます。

注意: このライブラリはさくらのAI Engine API用のクライアントです。一部のインターフェースはOpenAI APIと類似していますが、完全な互換性は保証されません。モデル名やパラメータについては、AI Engineの仕様に合わせてください。

AI Engineを使って作成されているため、細かくは動作テストしていません。

## インストール

```bash
go get github.com/masa23/aiengine-go
```

## 使用例

### クライアントの初期化

```go
import "github.com/masa23/aiengine-go"

client := aiengine.NewClient("your-api-key")
```

または、オプションを使用して初期化:

```go
client := aiengine.NewClient("your-api-key",
    aiengine.WithBaseURL("https://custom.api.example.com"),
    aiengine.WithTimeout(30*time.Second),
    aiengine.WithMaxRetries(5))
```

または、環境変数から初期化:

```go
client, err := aiengine.NewClientFromEnv()
if err != nil {
    log.Fatal(err)
}
```

または、オプション付きで環境変数から初期化:

```go
client, err := aiengine.NewClientFromEnv(
    aiengine.WithTimeout(30*time.Second),
    aiengine.WithMaxRetries(5))
if err != nil {
    log.Fatal(err)
}
```

### チャット補完

```go
req := &aiengine.ChatCompletionRequest{
    Model: "gpt-oss-120b",
    Messages: []aiengine.ChatCompletionRequestMessage{
        &aiengine.ChatCompletionRequestUserMessage{
            Role:    aiengine.ChatCompletionMessageRoleTypeUser,
            Content: "こんにちは、元気ですか？",
        },
    },
}

resp, err := client.CreateChatCompletion(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message)
```

### 埋め込みベクトル

```go
req := &aiengine.EmbeddingRequest{
    Model: "multilingual-e5-large",
    Input: "これは埋め込みのテストです。",
}

resp, err := client.CreateEmbeddings(context.Background(), req)
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Data[0].Embedding)
```

### RAG (Retrieval-Augmented Generation)

#### ドキュメントのアップロード

```go
uploadReq := &aiengine.DocumentUpload{
    Name: "test-document",
    File: "path/to/your/document.txt",
    Model: "multilingual-e5-large",
}

uploadResp, err := client.UploadDocument(context.Background(), uploadReq)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Document uploaded with ID:", uploadResp.ID)
```

#### ドキュメントとのチャット

```go
chatReq := &aiengine.ChatRequest{
    ChatModel: "gpt-oss-120b",
    Query:     "ドキュメントの内容について教えてください。",
    Tags:      []string{"test-document"},
}

chatResp, err := client.ChatWithDocuments(context.Background(), chatReq)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Answer:", chatResp.Answer)
```

### 音声書き起こし

```go
transReq := &aiengine.TranscriptionRequest{
    File: "path/to/your/audio.mp3",
    Model: "whisper-large-v3-turbo",
}

transResp, err := client.CreateTranscription(context.Background(), transReq)
if err != nil {
    log.Fatal(err)
}

fmt.Println("Transcription:", transResp.Text)
```

### テキスト読み上げ (TTS)

#### シンプルなTTS（Create Speech）

```go
speechReq := &aiengine.SpeechRequest{
    Model: "zundamon",
    Voice: "normal",
    Input: "こんにちは。",
}

audioData, err := client.CreateSpeech(context.Background(), speechReq)
if err != nil {
    log.Fatal(err)
}

// 音声データをファイルに保存
err = os.WriteFile("output.wav", audioData, 0644)
if err != nil {
    log.Fatal(err)
}
```

#### VOICEVOX形式のTTS（Audio Query + Synthesis）

```go
// ステップ1: 音声合成用クエリを作成
queryReq := &aiengine.TtsAudioQueryRequest{
    Text: "こんにちは。",
    Speaker: 3, // さくらのAI Engineの話者ID（例: 3はずんだもん）
}

query, err := client.CreateAudioQuery(context.Background(), queryReq)
if err != nil {
    log.Fatal(err)
}

// ステップ2: クエリパラメータを調整（オプション）
query.SpeedScale = 1.2      // 話速
query.PitchScale = 0.5      // 音高
query.VolumeScale = 1.0     // 音量

// ステップ3: 音声を合成
synthesisReq := &aiengine.TtsSynthesisRequest{
    Speaker: 1,
    Query:   query,
}

audioData, err := client.SynthesizeTtsSpeech(context.Background(), synthesisReq)
if err != nil {
    log.Fatal(err)
}

// 音声データをファイルに保存
err = os.WriteFile("output.wav", audioData, 0644)
if err != nil {
    log.Fatal(err)
}
```

## Example プログラム

### チャット (Chat)

チャット補完を試すには、環境変数 `SAKURA_AI_ENGINE_API_KEY` を設定して以下を実行します:

```bash
cd example/chat
go run main.go
```

### 音声合成 (TTS)

音声合成を試すには、環境変数 `SAKURA_AI_ENGINE_API_KEY` を設定して以下を実行します:

```bash
cd example/tts
go run main.go
```

## クライアントオプション

クライアントは以下のオプションをサポートしています:

- `WithBaseURL(string)`: APIのベースURLを設定します
- `WithHTTPClient(*http.Client)`: カスタムHTTPクライアントを設定します
- `WithTimeout(time.Duration)`: リクエストタイムアウトを設定します
- `WithMaxRetries(int)`: リトライ回数を設定します（デフォルト: 3）
- `WithRetryBackoff(time.Duration)`: リトライ時のバックオフ時間を設定します（デフォルト: 1秒）

## リトライ機能

クライアントは自動的にリトライ機能を実装しており、以下のステータスコードの場合にリトライを行います:

- 429 (Too Many Requests)
- 503 (Service Unavailable)
- 504 (Gateway Timeout)

`Retry-After`ヘッダーが存在する場合はその値を尊重し、そうでない場合は指数関数的バックオフでリトライします。

## 環境変数

Integration tests run only when required environment variables are set.

### 必須

- `SAKURA_AI_ENGINE_API_KEY` (統合テストに必要)

### オプション

- `SAKURA_AI_ENGINE_BASE_URL` (デフォルト: https://api.ai.sakura.ad.jp)

### テストごとのオプション

- `SAKURA_AI_ENGINE_CHAT_MODEL`
- `SAKURA_AI_ENGINE_EMBEDDING_MODEL`
- `SAKURA_AI_ENGINE_RAG_EMBEDDING_MODEL`
- `SAKURA_AI_ENGINE_AUDIO_FILE` (音声書き起こしのための音声ファイルパス)
- `SAKURA_AI_ENGINE_TTS_SPEAKER` (TTSの話者ID、デフォルト: 3 - ずんだもん)

## テスト

テストを実行するには、上記の環境変数を設定してください。

```bash
go test -v ./...
```

## ライセンス

MIT License
