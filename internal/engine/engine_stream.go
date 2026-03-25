package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
)

type StreamChunk struct {
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   string `json:"error,omitempty"`
}

func NewStreamChunk(content string) StreamChunk {
	return StreamChunk{Content: content, Done: false}
}

func NewStreamChunkDone() StreamChunk {
	return StreamChunk{Done: true}
}

func NewStreamChunkError(err error) StreamChunk {
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

type EngineStream struct {
	dad *Engine
}

func (my *EngineStream) ProcessMessage(ctx context.Context, sessionId, message string) <-chan StreamChunk {
	ch := make(chan StreamChunk, 100)

	loom.Go(func(later loom.Later) {
		defer close(ch)

		if sessionId == "" {
			ch <- NewStreamChunkError(fmt.Errorf("session ID cannot be empty"))
			return
		}

		if message == "" {
			ch <- NewStreamChunkError(fmt.Errorf("message cannot be empty"))
			return
		}

		var dad = my.dad
		dad.mu.RLock()
		session, exists := dad.sessions[sessionId]
		dad.mu.RUnlock()

		if !exists {
			ch <- NewStreamChunkError(fmt.Errorf("session not found: %s", sessionId))
			return
		}

		logo.Info("[Stream Session:", sessionId, "] User:", message)

		if err := session.AddMessage("user", message); err != nil {
			ch <- NewStreamChunkError(fmt.Errorf("failed to save user message: %w", err))
			return
		}

		hasLLM := (dad.pluginManager != nil && dad.llmPlugin != nil)
		if !hasLLM {
			ch <- NewStreamChunk(fmt.Sprintf("Echo: %s", message))
			ch <- NewStreamChunkDone()
			return
		}

		logo.Info("[Stream Session:", sessionId, "] Starting streaming")

		var fullContent strings.Builder
		streamCh, err := my.callLLM(ctx, session)
		if err != nil {
			ch <- NewStreamChunkError(fmt.Errorf("failed to start streaming: %w", err))
			return
		}

		for chunk := range streamCh {
			fullContent.WriteString(chunk)
			ch <- NewStreamChunk(chunk)
		}

		response := fullContent.String()
		if err := session.AddMessage("assistant", response); err != nil {
			ch <- NewStreamChunkError(fmt.Errorf("failed to save assistant response: %w", err))
			return
		}

		logo.Info("[Stream Session:", sessionId, "] Streaming completed, response length:", len(response))
		ch <- NewStreamChunkDone()
	})

	return ch
}

func (my *EngineStream) callLLM(ctx context.Context, session *Session) (<-chan string, error) {
	var systemPrompt = my.dad.buildSystemPrompt()
	var messages = session.AsBamlMessages()

	var chunks = streamChat(ctx, systemPrompt, messages)

	ch := make(chan string, 100)
	go func() {
		defer close(ch)
		for chunk := range chunks {
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
