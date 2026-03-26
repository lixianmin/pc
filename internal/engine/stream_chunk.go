package engine

import (
	"fmt"

	"github.com/lixianmin/pc/pkg/protocol"
)

type StreamChunk struct {
	Type    protocol.StreamChunkType `json:"type"`
	Content string                   `json:"content,omitempty"`
	Tool    string                   `json:"tool,omitempty"`
	Meta    any                      `json:"meta,omitempty"`
	Done    bool                     `json:"done"`
	Error   string                   `json:"error,omitempty"`
}

func NewStreamChunk(typ protocol.StreamChunkType, content string) StreamChunk {
	return StreamChunk{Type: typ, Content: content}
}

func NewStreamChunkToolCall(tool, content string) StreamChunk {
	return StreamChunk{Type: protocol.ChunkTypeToolCall, Tool: tool, Content: content}
}

func NewStreamChunkToolResult(tool, content string) StreamChunk {
	return StreamChunk{Type: protocol.ChunkTypeToolResult, Tool: tool, Content: content}
}

func NewStreamChunkDone() StreamChunk {
	return StreamChunk{Type: protocol.ChunkTypeDone, Done: true}
}

func NewStreamChunkError(err error) StreamChunk {
	return StreamChunk{Type: protocol.ChunkTypeError, Error: err.Error(), Done: true}
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
