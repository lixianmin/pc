package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/llm"
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
	return my.callLLMStreamViaBAML(ctx, session, message)
}

func (my *Engine) callLLMStreamViaBAML(ctx context.Context, session *Session, message string) (<-chan string, error) {
	history := session.GetMessages()
	systemPrompt := my.buildDynamicSystemPrompt()

	var messages []llm.Message
	for _, msg := range history {
		messages = append(messages, llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	messages = append(messages, llm.Message{
		Role:    "user",
		Content: message,
	})

	stream := my.llmClient.StreamChat(ctx, messages, systemPrompt)

	ch := make(chan string, 100)
	go func() {
		defer close(ch)
		for chunk := range stream {
			if chunk.Error != nil {
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

func (my *Engine) callLLMStreamViaPlugin(ctx context.Context, session *Session, message string) (<-chan string, error) {
	history := session.GetMessages()
	systemPrompt := my.buildDynamicSystemPrompt()

	capacity := len(history) + 1
	if systemPrompt != "" {
		capacity++
	}
	messages := make([]map[string]string, 0, capacity)

	if systemPrompt != "" {
		messages = append(messages, map[string]string{
			"role":    "system",
			"content": systemPrompt,
		})
	}

	for _, msg := range history {
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": message,
	})

	params := map[string]any{
		"messages": messages,
	}

	result, err := my.pluginManager.CallPlugin(my.llmPlugin, "stream", params)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected stream response format")
	}

	chunks, ok := resultMap["chunks"].([]any)
	if !ok {
		return nil, fmt.Errorf("stream response missing chunks")
	}

	ch := make(chan string, len(chunks))
	go func() {
		defer close(ch)
		for _, chunk := range chunks {
			chunkMap, ok := chunk.(map[string]any)
			if !ok {
				continue
			}
			content, _ := chunkMap["content"].(string)
			if content != "" {
				ch <- content
			}
		}
	}()

	return ch, nil
}
