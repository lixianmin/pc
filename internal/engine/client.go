package engine

import (
	"context"

	baml "github.com/lixianmin/pc/baml_client"
	"github.com/lixianmin/pc/baml_client/types"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (my *Client) Chat(ctx context.Context, messages []types.Message, systemPrompt string) (ChatResult, error) {
	return baml.Chat(ctx, messages, systemPrompt)
}

func (my *Client) StreamChat(ctx context.Context, messages []types.Message, systemPrompt string) <-chan StreamChunk {
	stream, err := baml.Stream.Chat(ctx, messages, systemPrompt)
	ch := make(chan StreamChunk, 100)

	if err != nil {
		ch <- StreamChunk{Error: err.Error()}
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)
		for chunk := range stream {
			if chunk.IsError {
				ch <- StreamChunk{Error: chunk.Error.Error()}
				return
			}
			if chunk.IsFinal && chunk.Final() != nil {
				result := chunk.Final()
				if result.IsChatResponse() {
					chatResp := result.AsChatResponse()
					content := ""
					if chatResp != nil {
						content = chatResp.Content
					}
					ch <- StreamChunk{Content: content, Done: true}
				} else {
					ch <- StreamChunk{Content: "[Tool call]", Done: true}
				}
				return
			}
			if chunk.Stream() != nil {
				s := chunk.Stream()
				if s.IsChatResponse() {
					chatResp := s.AsChatResponse()
					if chatResp != nil && chatResp.Content != nil {
						ch <- StreamChunk{Content: *chatResp.Content}
					}
				}
			}
		}
	}()

	return ch
}
