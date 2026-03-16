package engine

import "fmt"

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
