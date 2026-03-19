package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/internal/engine/builtin"
	"github.com/lixianmin/pc/pkg/types"
)

// PluginManager defines the interface for plugin management.
// This interface is satisfied by *plugin.PluginManager.
type PluginManager interface {
	ListPlugins() []*types.Plugin
	CallPlugin(plugin *types.Plugin, method string, params any) (any, error)
}

// ToolExecutor executes tool calls by invoking tool plugins.
type ToolExecutor struct {
	pluginManager   PluginManager
	builtinRegistry *Registry
	timeout         time.Duration
}

// NewToolExecutor creates a new tool executor.
func NewToolExecutor(pm PluginManager) *ToolExecutor {
	registry := NewRegistry()
	registry.Register(builtin.NewBashTool())
	registry.Register(builtin.NewReadTool())
	registry.Register(builtin.NewWriteTool())
	registry.Register(builtin.NewEditTool())
	registry.Register(builtin.NewGlobTool())
	registry.Register(builtin.NewGrepTool())

	return &ToolExecutor{
		pluginManager:   pm,
		builtinRegistry: registry,
		timeout:         30 * time.Second,
	}
}

// SetTimeout sets the execution timeout for tool calls.
func (my *ToolExecutor) SetTimeout(timeout time.Duration) {
	my.timeout = timeout
}

// Execute executes a single tool call.
// It checks built-in tools first, then falls back to plugin tools.
func (my *ToolExecutor) Execute(ctx context.Context, call ToolCall) ToolResult {
	if my.builtinRegistry != nil {
		if tool, ok := my.builtinRegistry.Get(call.Name); ok {
			return my.executeBuiltin(ctx, call, tool)
		}
	}

	if my.pluginManager == nil {
		return ToolResult{
			Name:  call.Name,
			Error: fmt.Errorf("plugin manager not initialized"),
		}
	}

	plugin, err := my.findToolPlugin(call.Name)
	if err != nil {
		return ToolResult{
			Name:  call.Name,
			Error: fmt.Errorf("tool not found: %w", err),
		}
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), my.timeout)
		defer cancel()
	}

	logo.Info("Executing tool:", call.Name, "params:", call.Params)

	resultChan := make(chan struct {
		result any
		err    error
	}, 1)

	go func() {
		result, err := my.pluginManager.CallPlugin(plugin, "call", call.Params)
		resultChan <- struct {
			result any
			err    error
		}{result, err}
	}()

	select {
	case <-ctx.Done():
		return ToolResult{
			Name:  call.Name,
			Error: fmt.Errorf("tool execution timeout: %w", ctx.Err()),
		}
	case res := <-resultChan:
		if res.err != nil {
			return ToolResult{
				Name:  call.Name,
				Error: fmt.Errorf("tool execution failed: %w", res.err),
			}
		}
		return my.formatResult(call.Name, res.result)
	}
}

func (my *ToolExecutor) executeBuiltin(ctx context.Context, call ToolCall, tool BuiltinTool) ToolResult {
	logo.Info("Executing builtin tool:", call.Name, "params:", call.Params)

	result, err := tool.Execute(ctx, call.Params)
	if err != nil {
		return ToolResult{
			Name:  call.Name,
			Error: err,
		}
	}

	return ToolResult{
		Name:   call.Name,
		Output: result,
	}
}

// ExecuteMultiple executes multiple tool calls in sequence.
func (my *ToolExecutor) ExecuteMultiple(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, 0, len(calls))

	for _, call := range calls {
		result := my.Execute(ctx, call)
		results = append(results, result)
	}

	return results
}

func (my *ToolExecutor) findToolPlugin(name string) (*types.Plugin, error) {
	if my.pluginManager == nil {
		return nil, fmt.Errorf("plugin manager not initialized")
	}

	plugins := my.pluginManager.ListPlugins()

	for _, p := range plugins {
		if p.Name == name && p.Type == types.PluginTypeTool && p.Enabled {
			return p, nil
		}
	}

	return nil, fmt.Errorf("tool plugin '%s' not found or not enabled", name)
}

func (my *ToolExecutor) formatResult(name string, result any) ToolResult {
	switch v := result.(type) {
	case map[string]interface{}:
		if output, ok := v["output"].(string); ok {
			return ToolResult{
				Name:   name,
				Output: output,
			}
		}
		if stdout, ok := v["stdout"].(string); ok {
			return ToolResult{
				Name:   name,
				Output: stdout,
			}
		}
		if content, ok := v["content"].(string); ok {
			return ToolResult{
				Name:   name,
				Output: content,
			}
		}
		return ToolResult{
			Name:   name,
			Output: fmt.Sprintf("%v", v),
		}
	case string:
		return ToolResult{
			Name:   name,
			Output: v,
		}
	default:
		return ToolResult{
			Name:   name,
			Output: fmt.Sprintf("%v", v),
		}
	}
}

func (my *ToolExecutor) IsToolAvailable(name string) bool {
	if my.builtinRegistry != nil {
		if _, ok := my.builtinRegistry.Get(name); ok {
			return true
		}
	}

	if my.pluginManager == nil {
		return false
	}

	_, err := my.findToolPlugin(name)
	return err == nil
}

func (my *ToolExecutor) ListAvailableTools() []string {
	var tools []string

	if my.builtinRegistry != nil {
		for name := range my.builtinRegistry.List() {
			tools = append(tools, name)
		}
	}

	if my.pluginManager != nil {
		plugins := my.pluginManager.ListPlugins()
		for _, p := range plugins {
			if p.Type == types.PluginTypeTool && p.Enabled {
				tools = append(tools, p.Name)
			}
		}
	}

	return tools
}
