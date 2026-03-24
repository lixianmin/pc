package llm

import (
	"context"
	"fmt"

	baml "github.com/lixianmin/pc/baml_client/baml_client"
	"github.com/lixianmin/pc/baml_client/baml_client/types"
)

type Message struct {
	Role    string
	Content string
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Chat(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	var bamlMessages []types.Message
	for _, m := range messages {
		bamlMessages = append(bamlMessages, types.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	result, err := baml.Chat(ctx, bamlMessages, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("BAML Chat failed: %w", err)
	}

	return result, nil
}
