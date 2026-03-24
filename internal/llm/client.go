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

type StreamChunk struct {
	Content string
	Done    bool
	Error   error
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

func (c *Client) StreamChat(ctx context.Context, messages []Message, systemPrompt string) <-chan StreamChunk {
	var bamlMessages []types.Message
	for _, m := range messages {
		bamlMessages = append(bamlMessages, types.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	stream, err := baml.Stream.Chat(ctx, bamlMessages, systemPrompt)
	if err != nil {
		ch := make(chan StreamChunk, 1)
		ch <- StreamChunk{Error: err, Done: true}
		close(ch)
		return ch
	}

	ch := make(chan StreamChunk, 100)
	go func() {
		defer close(ch)
		for chunk := range stream {
			if chunk.IsError {
				ch <- StreamChunk{Error: chunk.Error, Done: true}
				return
			}
			if chunk.IsFinal && chunk.Final() != nil {
				ch <- StreamChunk{Content: *chunk.Final(), Done: true}
				return
			}
			if chunk.Stream() != nil {
				ch <- StreamChunk{Content: *chunk.Stream()}
			}
		}
	}()

	return ch
}
