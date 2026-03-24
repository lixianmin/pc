package tool

import (
	"context"
	"fmt"
	"sync"
)

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]any) (any, error)
}

type Executor struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewExecutor() *Executor {
	return &Executor{
		tools: make(map[string]Tool),
	}
}

func (e *Executor) Register(tool Tool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tools[tool.Name()] = tool
}

func (e *Executor) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	e.mu.RLock()
	tool, ok := e.tools[name]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	return tool.Execute(ctx, params)
}

func (e *Executor) ListTools() []ToolInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var tools []ToolInfo
	for name, tool := range e.tools {
		tools = append(tools, ToolInfo{
			Name:        name,
			Description: tool.Description(),
		})
	}
	return tools
}

type ToolInfo struct {
	Name        string
	Description string
}
