package engine

import (
	"context"

	baml "github.com/lixianmin/pc/baml_client"
	"github.com/lixianmin/pc/baml_client/types"
)

type ChatResult struct {
	types.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool
}

func (my *ChatResult) IsChatResponse() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsChatResponse()
}

func (my *ChatResult) AsChatResponse() *types.ChatResponse {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsChatResponse()
}

func (my *ChatResult) IsBashTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsBashTool()
}

func (my *ChatResult) AsBashTool() *types.BashTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsBashTool()
}

func (my *ChatResult) IsReadTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsReadTool()
}

func (my *ChatResult) AsReadTool() *types.ReadTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsReadTool()
}

func (my *ChatResult) IsWriteTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsWriteTool()
}

func (my *ChatResult) AsWriteTool() *types.WriteTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsWriteTool()
}

func (my *ChatResult) IsEditTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsEditTool()
}

func (my *ChatResult) AsEditTool() *types.EditTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsEditTool()
}

func (my *ChatResult) IsWebSearchTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsWebSearchTool()
}

func (my *ChatResult) AsWebSearchTool() *types.WebSearchTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsWebSearchTool()
}

func (my *ChatResult) IsWebFetchTool() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsWebFetchTool()
}

func (my *ChatResult) AsWebFetchTool() *types.WebFetchTool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsWebFetchTool()
}

func (my *ChatResult) IsUseSkill() bool {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.IsUseSkill()
}

func (my *ChatResult) AsUseSkill() *types.UseSkill {
	return my.Union8BashToolOrChatResponseOrEditToolOrReadToolOrUseSkillOrWebFetchToolOrWebSearchToolOrWriteTool.AsUseSkill()
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (my *Client) Chat(ctx context.Context, messages []types.Message, systemPrompt string) (*ChatResult, error) {
	result, err := baml.Chat(ctx, messages, systemPrompt)
	if err != nil {
		return nil, err
	}

	return &ChatResult{result}, nil
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
