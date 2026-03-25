package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/baml_client/types"
)

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   string `json:"error,omitempty"`
}

func NewStreamChunk(content string) StreamChunk {
	return StreamChunk{Content: content, Done: false}
}

func NewStreamDone() StreamChunk {
	return StreamChunk{Done: true}
}

func NewStreamError(err error) StreamChunk {
	return StreamChunk{Error: err.Error(), Done: true}
}

func (my StreamChunk) String() string {
	if my.Error != "" {
		return fmt.Sprintf("StreamChunk{Error: %s}", my.Error)
	}
	if my.Done {
		return "StreamChunk{Done: true}"
	}
	return fmt.Sprintf("StreamChunk{Content: %q}", my.Content)
}

func (my *Engine) ProcessMessageStream(ctx context.Context, sessionId, message string) <-chan StreamChunk {
	ch := make(chan StreamChunk, 100)

	loom.Go(func(later loom.Later) {
		defer close(ch)

		if sessionId == "" {
			ch <- NewStreamError(fmt.Errorf("session ID cannot be empty"))
			return
		}

		if message == "" {
			ch <- NewStreamError(fmt.Errorf("message cannot be empty"))
			return
		}

		my.mu.RLock()
		session, exists := my.sessions[sessionId]
		my.mu.RUnlock()

		if !exists {
			ch <- NewStreamError(fmt.Errorf("session not found: %s", sessionId))
			return
		}

		logo.Info("[Stream Session:", sessionId, "] User:", message)

		if err := session.AddMessage("user", message); err != nil {
			ch <- NewStreamError(fmt.Errorf("failed to save user message: %w", err))
			return
		}

		hasLLM := my.llmClient != nil || (my.pluginManager != nil && my.llmPlugin != nil)
		if !hasLLM {
			ch <- NewStreamChunk(fmt.Sprintf("Echo: %s", message))
			ch <- NewStreamDone()
			return
		}

		logo.Info("[Stream Session:", sessionId, "] Starting streaming")

		var fullContent strings.Builder
		streamCh, err := my.callLLMStream(ctx, session, message)
		if err != nil {
			ch <- NewStreamError(fmt.Errorf("failed to start streaming: %w", err))
			return
		}

		for chunk := range streamCh {
			fullContent.WriteString(chunk)
			ch <- NewStreamChunk(chunk)
		}

		response := fullContent.String()
		if err := session.AddMessage("assistant", response); err != nil {
			ch <- NewStreamError(fmt.Errorf("failed to save assistant response: %w", err))
			return
		}

		logo.Info("[Stream Session:", sessionId, "] Streaming completed, response length:", len(response))
		ch <- NewStreamDone()
	})

	return ch
}

func (my *Engine) callLLMStream(ctx context.Context, session *Session, message string) (<-chan string, error) {
	history := session.GetMessages()
	systemPrompt := my.buildSystemPrompt()

	var messages []types.Message
	for _, msg := range history {
		messages = append(messages, types.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, types.Message{
		Role:    "user",
		Content: message,
	})

	stream := my.llmClient.StreamChat(ctx, messages, systemPrompt)

	ch := make(chan string, 100)
	go func() {
		defer close(ch)
		for chunk := range stream {
			if chunk.Error != "" {
				logo.Error("[Engine.callLLMStreamViaBAML] Stream error:", chunk.Error)
				return
			}
			if chunk.Done {
				return
			}
			if chunk.Content != "" {
				ch <- chunk.Content
			}
		}
	}()

	return ch, nil
}
