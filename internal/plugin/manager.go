package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lixianmin/pc/pkg/types"
	"gopkg.in/yaml.v3"
)

// Manager manages plugin lifecycle and invocation.
type Manager struct {
	plugins    map[string]*types.Plugin
	pluginsDir string
	protocols  map[string]interface{} // Plugin protocol instances
	mu         sync.RWMutex
}

// NewManager creates a new plugin manager.
func NewManager(pluginsDir string) (*Manager, error) {
	return &Manager{
		plugins:    make(map[string]*types.Plugin),
		pluginsDir: pluginsDir,
		protocols:  make(map[string]interface{}),
		mu:         sync.RWMutex{},
	}, nil
}

// Discover scans the plugins directory and loads all plugins.
func (m *Manager) Discover() ([]*types.Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if plugins directory exists
	if _, err := os.Stat(m.pluginsDir); os.IsNotExist(err) {
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
		typeDir := filepath.Join(m.pluginsDir, pType)
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
			plugin, err := m.loadPlugin(pluginPath)
			if err != nil {
				// Log but continue loading other plugins
				continue
			}
			if plugin != nil {
				plugins = append(plugins, plugin)
				m.plugins[plugin.Name] = plugin
			}
		}
	}

	return plugins, nil
}

// LoadPlugin loads a plugin from the given path.
func (m *Manager) LoadPlugin(path string) (*types.Plugin, error) {
	return m.loadPlugin(path)
}

// loadPlugin loads plugin metadata from plugin.yml.
func (m *Manager) loadPlugin(path string) (*types.Plugin, error) {
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
func (m *Manager) CallPlugin(plugin *types.Plugin, method string, params any) (any, error) {
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}

	// TODO: Implement actual plugin invocation via stdio protocol
	// For now, return an error indicating not implemented
	return nil, fmt.Errorf("plugin invocation not yet implemented")
}

// GetPlugin returns a plugin by name.
func (m *Manager) GetPlugin(name string) (*types.Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, ok := m.plugins[name]
	if !ok {
		return nil, fmt.Errorf("plugin %s not found", name)
	}

	return plugin, nil
}

// ListPlugins returns all loaded plugins.
func (m *Manager) ListPlugins() []*types.Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*types.Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// EnablePlugin enables a plugin.
func (m *Manager) EnablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Enabled = true
	return nil
}

// DisablePlugin disables a plugin.
func (m *Manager) DisablePlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, ok := m.plugins[name]
	if !ok {
		return fmt.Errorf("plugin %s not found", name)
	}

	plugin.Enabled = false
	return nil
}

// Close cleans up all plugin resources.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all plugin protocols
	for name, proto := range m.protocols {
		// TODO: Properly close plugin protocol connections
		_ = name
		_ = proto
	}

	m.protocols = make(map[string]interface{})
	return nil
}
