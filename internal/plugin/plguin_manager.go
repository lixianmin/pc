package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lixianmin/pc/pkg/types"
	"gopkg.in/yaml.v3"
)

// PluginManager manages plugin lifecycle and invocation.
type PluginManager struct {
	plugins    map[string]*types.Plugin
	pluginsDir string
	protocols  map[string]any // Plugin protocol instances
	mu         sync.RWMutex
}

// NewPluginManager creates a new plugin manager.
func NewPluginManager(pluginsDir string) (*PluginManager, error) {
	return &PluginManager{
		plugins:    make(map[string]*types.Plugin),
		pluginsDir: pluginsDir,
		protocols:  make(map[string]any),
		mu:         sync.RWMutex{},
	}, nil
}

// Discover scans the plugins directory and loads all plugins.
func (my *PluginManager) Discover() ([]*types.Plugin, error) {
	my.mu.Lock()
	defer my.mu.Unlock()

	// Check if plugins directory exists
	if _, err := os.Stat(my.pluginsDir); os.IsNotExist(err) {
		return nil, nil
	}

	var plugins []*types.Plugin

	// Scan for plugin types
	pluginTypes := []string{
		string(types.PluginTypeLLM),
		string(types.PluginTypeSearch),
		string(types.PluginTypeChannel),
		string(types.PluginTypeTool),
	}

	for _, pType := range pluginTypes {
		typeDir := filepath.Join(my.pluginsDir, pType)
		entries, err := os.ReadDir(typeDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("failed to read plugin directory %s: %w", typeDir, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pluginPath := filepath.Join(typeDir, entry.Name())
			plugin, err := my.loadPlugin(pluginPath)
			if err != nil {
				// Log but continue loading other plugins
				continue
			}
			if plugin != nil {
				plugins = append(plugins, plugin)
				my.plugins[plugin.Name] = plugin
			}
		}
	}

	return plugins, nil
}

// LoadPlugin loads a plugin from the given path.
func (my *PluginManager) LoadPlugin(path string) (*types.Plugin, error) {
	return my.loadPlugin(path)
}

// loadPlugin loads plugin metadata from plugin.yml.
func (my *PluginManager) loadPlugin(path string) (*types.Plugin, error) {
	pluginYmlPath := filepath.Join(path, "plugin.yml")

	data, err := os.ReadFile(pluginYmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.yml: %w", err)
	}

	var meta struct {
		Name    string `yaml:"name"`
		Type    string `yaml:"type"`
		Enabled bool   `yaml:"enabled"`
		Version string `yaml:"version"`
		Entry   string `yaml:"entry"`
	}

	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.yml: %w", err)
	}

	// Validate required fields
	if meta.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}
	// Use default type if not specified
	if meta.Type == "" {
		meta.Type = string(types.PluginTypeTool)
	}
	if meta.Entry == "" {
		return nil, fmt.Errorf("plugin entry is required")
	}

	plugin := &types.Plugin{
		Name:    meta.Name,
		Type:    types.PluginTypeFromString(meta.Type),
		Enabled: meta.Enabled,
		Version: meta.Version,
		Entry:   meta.Entry,
		Path:    path,
	}

	return plugin, nil
}

// CallPlugin invokes a method on a plugin.
func (my *PluginManager) CallPlugin(plugin *types.Plugin, method string, params any) (any, error) {
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}

	// TODO: Implement actual plugin invocation via stdio protocol
	// For now, return an error indicating not implemented
	return nil, fmt.Errorf("plugin invocation not yet implemented")
}

// GetPlugin returns a plugin by name.
func (my *PluginManager) GetPlugin(name string) (*types.Plugin, error) {
	my.mu.RLock()
	defer my.mu.RUnlock()

	plugin, ok := my.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// ListPlugins returns all loaded plugins.
func (my *PluginManager) ListPlugins() []*types.Plugin {
	my.mu.RLock()
	defer my.mu.RUnlock()

	plugins := make([]*types.Plugin, 0, len(my.plugins))
	for _, plugin := range my.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// EnablePlugin enables a plugin.
func (my *PluginManager) EnablePlugin(name string) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	plugin, ok := my.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Enabled = true
	return nil
}

// DisablePlugin disables a plugin.
func (my *PluginManager) DisablePlugin(name string) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	plugin, ok := my.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Enabled = false
	return nil
}

// Close cleans up all plugin resources.
func (my *PluginManager) Close() error {
	my.mu.Lock()
	defer my.mu.Unlock()

	// Close all plugin protocols
	for name, proto := range my.protocols {
		// TODO: Properly close plugin protocol connections
		_ = name
		_ = proto
	}

	my.protocols = make(map[string]any)
	return nil
}
