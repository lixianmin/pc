package llm

import (
	"context"
	"os"
	"testing"
)

// TEMP: 验证 BAML 连接，阶段 3 删除
func TestBAMLConnection(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client := NewClient()
	ctx := context.Background()

	result, err := client.Chat(ctx, []Message{
		{Role: "user", Content: "Say 'hello' in one word"},
	}, "You are a helpful assistant.")

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	t.Logf("Response: %s", result)
}
