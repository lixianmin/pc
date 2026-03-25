package engine

import (
	"context"

	"github.com/lixianmin/pc/baml_client"
	"github.com/lixianmin/pc/baml_client/types"
)

func streamChat(ctx context.Context, systemPrompt string, messages []types.Message) <-chan StreamChunk {
	stream, err := baml_client.Stream.Chat(ctx, systemPrompt, messages)
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
