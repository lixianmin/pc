package engine

import (
	"context"
	"fmt"

	baml "github.com/lixianmin/pc/baml_client"
	"github.com/lixianmin/pc/baml_client/types"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (my *Client) Chat(ctx context.Context, messages []types.Message, systemPrompt string) (string, error) {
	result, err := baml.Chat(ctx, messages, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("BAML Chat failed: %w", err)
	}

	return result, nil
}

func (my *Client) StreamChat(ctx context.Context, messages []types.Message, systemPrompt string) <-chan StreamChunk {
	stream, err := baml.Stream.Chat(ctx, messages, systemPrompt)
	if err != nil {
		ch := make(chan StreamChunk, 1)
		ch <- NewStreamError(err)
		close(ch)
		return ch
	}

	ch := make(chan StreamChunk, 100)
	go func() {
		defer close(ch)
		for chunk := range stream {
			if chunk.IsError {
				ch <- NewStreamError(chunk.Error)
				return
			}
			if chunk.IsFinal && chunk.Final() != nil {
				ch <- StreamChunk{Content: *chunk.Final(), Done: true}
				return
			}
			if chunk.Stream() != nil {
				ch <- NewStreamChunk(*chunk.Stream())
			}
		}
	}()

	return ch
}
