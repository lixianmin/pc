package engine

import "sync"

type Registry struct {
	mu    sync.RWMutex
	tools map[string]BuiltinTool
}

func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]BuiltinTool),
	}
}

func (my *Registry) Register(tool BuiltinTool) {
	my.mu.Lock()
	defer my.mu.Unlock()
	if _, exists := my.tools[tool.Name()]; exists {
		return
	}
	my.tools[tool.Name()] = tool
}

func (my *Registry) Get(name string) (BuiltinTool, bool) {
	my.mu.RLock()
	defer my.mu.RUnlock()
	tool, exists := my.tools[name]
	return tool, exists
}

func (my *Registry) List() map[string]BuiltinTool {
	my.mu.RLock()
	defer my.mu.RUnlock()
	result := make(map[string]BuiltinTool, len(my.tools))
	for k, v := range my.tools {
		result[k] = v
	}
	return result
}
