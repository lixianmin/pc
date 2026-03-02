package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/lixianmin/logo"
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
	pluginManager PluginManager
	timeout       time.Duration
}

// NewToolExecutor creates a new tool executor.
func NewToolExecutor(pm PluginManager) *ToolExecutor {
	return &ToolExecutor{
		pluginManager: pm,
		timeout:       30 * time.Second, // Default timeout
	}
}

// SetTimeout sets the execution timeout for tool calls.
func (e *ToolExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// Execute executes a single tool call.
// It finds the corresponding tool plugin and invokes it.
func (e *ToolExecutor) Execute(ctx context.Context, call ToolCall) ToolResult {
	if e.pluginManager == nil {
		return ToolResult{
			Name:  call.Name,
			Error: fmt.Errorf("plugin manager not initialized"),
		}
	}

	// Find the tool plugin
	plugin, err := e.findToolPlugin(call.Name)
	if err != nil {
		return ToolResult{
			Name:  call.Name,
			Error: fmt.Errorf("tool not found: %w", err),
		}
	}

	// Create timeout context if not provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), e.timeout)
		defer cancel()
	}

	// Execute the tool call
	logo.Info("Executing tool:", call.Name, "params:", call.Params)

	resultChan := make(chan struct {
		result any
		err    error
	}, 1)

	go func() {
		result, err := e.pluginManager.CallPlugin(plugin, "call", call.Params)
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
		return e.formatResult(call.Name, res.result)
	}
}

// ExecuteMultiple executes multiple tool calls in sequence.
func (e *ToolExecutor) ExecuteMultiple(ctx context.Context, calls []ToolCall) []ToolResult {
	results := make([]ToolResult, 0, len(calls))

	for _, call := range calls {
		result := e.Execute(ctx, call)
		results = append(results, result)
	}

	return results
}

// findToolPlugin finds a tool plugin by name.
func (e *ToolExecutor) findToolPlugin(name string) (*types.Plugin, error) {
	plugins := e.pluginManager.ListPlugins()

	for _, p := range plugins {
		if p.Name == name && p.Type == types.PluginTypeTool && p.Enabled {
			return p, nil
		}
	}

	return nil, fmt.Errorf("tool plugin '%s' not found or not enabled", name)
}

// formatResult formats the plugin result into a ToolResult.
func (e *ToolExecutor) formatResult(name string, result any) ToolResult {
	// Try to extract output from result
	switch v := result.(type) {
	case map[string]interface{}:
		// Check for standard output fields
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
		// If no standard field, format the whole map
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

// IsToolAvailable checks if a tool plugin is available.
func (e *ToolExecutor) IsToolAvailable(name string) bool {
	if e.pluginManager == nil {
		return false
	}

	_, err := e.findToolPlugin(name)
	return err == nil
}

// ListAvailableTools returns a list of available tool names.
func (e *ToolExecutor) ListAvailableTools() []string {
	if e.pluginManager == nil {
		return nil
	}

	plugins := e.pluginManager.ListPlugins()
	var tools []string

	for _, p := range plugins {
		if p.Type == types.PluginTypeTool && p.Enabled {
			tools = append(tools, p.Name)
		}
	}

	return tools
}
