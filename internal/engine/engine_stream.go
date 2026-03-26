package engine

import (
	"context"
	"fmt"

	"github.com/lixianmin/got/loom"
	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/baml_client"
	"github.com/lixianmin/pc/pkg/protocol"
)

type StreamChunkType = protocol.StreamChunkType

const (
	ChunkTypeThinking   = protocol.ChunkTypeThinking
	ChunkTypeToolCall   = protocol.ChunkTypeToolCall
	ChunkTypeToolResult = protocol.ChunkTypeToolResult
	ChunkTypeResponse   = protocol.ChunkTypeResponse
	ChunkTypeDone       = protocol.ChunkTypeDone
	ChunkTypeError      = protocol.ChunkTypeError
)

type StreamChunk struct {
	Type    StreamChunkType `json:"type"`
	Content string          `json:"content,omitempty"`
	Tool    string          `json:"tool,omitempty"`
	Meta    any             `json:"meta,omitempty"`
	Done    bool            `json:"done"`
	Error   string          `json:"error,omitempty"`
}

func NewStreamChunk(typ StreamChunkType, content string) StreamChunk {
	return StreamChunk{Type: typ, Content: content}
}

func NewStreamChunkToolCall(tool, content string) StreamChunk {
	return StreamChunk{Type: ChunkTypeToolCall, Tool: tool, Content: content}
}

func NewStreamChunkToolResult(tool, content string) StreamChunk {
	return StreamChunk{Type: ChunkTypeToolResult, Tool: tool, Content: content}
}

func NewStreamChunkDone() StreamChunk {
	return StreamChunk{Type: ChunkTypeDone, Done: true}
}

func NewStreamChunkError(err error) StreamChunk {
	return StreamChunk{Type: ChunkTypeError, Error: err.Error(), Done: true}
}

func (my StreamChunk) String() string {
	if my.Error != "" {
		return fmt.Sprintf("StreamChunk{Error: %s}", my.Error)
	}
	if my.Done {
		return "StreamChunk{Done: true}"
	}
	return fmt.Sprintf("StreamChunk{Type: %s, Content: %q}", my.Type, my.Content)
}

type EngineStream struct {
	dad *Engine
}

func (my *EngineStream) ProcessMessage(ctx context.Context, session *Session, message string) <-chan StreamChunk {
	ch := make(chan StreamChunk, 100)

	loom.Go(func(later loom.Later) {
		defer close(ch)

		if session == nil {
			ch <- NewStreamChunkError(fmt.Errorf("session cannot be nil"))
			return
		}

		if message == "" {
			ch <- NewStreamChunkError(fmt.Errorf("message cannot be empty"))
			return
		}

		logo.Info("[Stream Session:", session.Id, "] User:", message)
		session.AddMessage("user", message)

		var finalResponse string
		var err error
		finalResponse, err = my.reactLoop(ctx, session, ch)
		if err != nil {
			ch <- NewStreamChunkError(err)
			return
		}

		logo.Info("[Stream Session:", session.Id, "] Completed, response length:", len(finalResponse))
		ch <- NewStreamChunkDone()
	})

	return ch
}

func (my *EngineStream) reactLoop(ctx context.Context, session *Session, ch chan<- StreamChunk) (string, error) {
	const maxIterations = 10

	for iteration := 0; iteration < maxIterations; iteration++ {
		ch <- NewStreamChunk(ChunkTypeThinking, fmt.Sprintf("Thinking (iteration %d)...", iteration+1))

		result, err := my.callLLM(ctx, session)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if chatResp := result.AsChatResponse(); chatResp != nil {
			ch <- NewStreamChunk(ChunkTypeResponse, chatResp.Content)
			session.AddMessage("assistant", chatResp.Content)
			return chatResp.Content, nil
		}

		_, toolResult, err := my.executeToolCall(result, ch)
		if err != nil {
			session.AddMessage("assistant", fmt.Sprintf("Tool error: %s", err))
			continue
		}

		session.AddMessage("assistant", toolResult)
	}

	return "", fmt.Errorf("max iterations (%d) exceeded", maxIterations)
}

func (my *EngineStream) executeToolCall(result *ChatResult, ch chan<- StreamChunk) (string, string, error) {
	if bash := result.AsBash(); bash != nil {
		ch <- NewStreamChunkToolCall("bash", bash.Command)
		output, err := my.dad.useTool(context.Background(), result)
		if err != nil {
			ch <- NewStreamChunkToolResult("bash", fmt.Sprintf("Error: %s", err))
			return "bash", "", err
		}
		ch <- NewStreamChunkToolResult("bash", truncateOutput(output, 500))
		return "bash", output, nil
	}

	if edit := result.AsEdit(); edit != nil {
		desc := fmt.Sprintf("Edit %s", edit.FilePath)
		ch <- NewStreamChunkToolCall("edit", desc)
		output, err := my.dad.useTool(context.Background(), result)
		if err != nil {
			ch <- NewStreamChunkToolResult("edit", fmt.Sprintf("Error: %s", err))
			return "edit", "", err
		}
		ch <- NewStreamChunkToolResult("edit", output)
		return "edit", output, nil
	}

	if read := result.AsRead(); read != nil {
		desc := fmt.Sprintf("Read %s", read.FilePath)
		ch <- NewStreamChunkToolCall("read", desc)
		output, err := my.dad.useTool(context.Background(), result)
		if err != nil {
			ch <- NewStreamChunkToolResult("read", fmt.Sprintf("Error: %s", err))
			return "read", "", err
		}
		ch <- NewStreamChunkToolResult("read", truncateOutput(output, 500))
		return "read", output, nil
	}

	if write := result.AsWrite(); write != nil {
		desc := fmt.Sprintf("Write %s", write.FilePath)
		ch <- NewStreamChunkToolCall("write", desc)
		output, err := my.dad.useTool(context.Background(), result)
		if err != nil {
			ch <- NewStreamChunkToolResult("write", fmt.Sprintf("Error: %s", err))
			return "write", "", err
		}
		ch <- NewStreamChunkToolResult("write", output)
		return "write", output, nil
	}

	return "", "", fmt.Errorf("unknown tool type")
}

func (my *EngineStream) callLLM(ctx context.Context, session *Session) (*ChatResult, error) {
	var systemPrompt = my.dad.buildSystemPrompt()
	var messages = session.AsBamlMessages()
	my.dad.printPrompt(systemPrompt, messages)

	result, err := baml_client.Chat(ctx, systemPrompt, messages)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
