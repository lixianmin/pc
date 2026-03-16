package plugin

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/lixianmin/logo"
	"github.com/lixianmin/pc/pkg/protocol"
	"github.com/lixianmin/pc/pkg/types"
	"gopkg.in/yaml.v3"
)

// PluginManager manages plugin lifecycle and invocation.
type PluginManager struct {
	plugins    map[string]*types.Plugin
	pluginsDir string
	protocols  map[string]*protocol.StdioProtocol // Plugin protocol instances
	mu         sync.RWMutex
}

// NewPluginManager creates a new plugin manager.
func NewPluginManager(pluginsDir string) (*PluginManager, error) {
	return &PluginManager{
		plugins:    make(map[string]*types.Plugin),
		pluginsDir: pluginsDir,
		protocols:  make(map[string]*protocol.StdioProtocol),
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
			plugin, err := my.LoadPlugin(pluginPath)
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

// ensurePluginStarted ensures a plugin is started and connected.
func (my *PluginManager) ensurePluginStarted(plugin *types.Plugin) (*protocol.StdioProtocol, error) {
	// Check if already connected
	if proto, ok := my.protocols[plugin.Name]; ok {
		return proto, nil
	}

	// Create new protocol connection
	entryPath := filepath.Join(plugin.Path, plugin.Entry)
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		// Try with just the entry path as-is
		entryPath = plugin.Entry
		if filepath.IsLocal(entryPath) {
			entryPath = filepath.Join(plugin.Path, entryPath)
		}
	}

	// Check if entry exists
	if _, err := os.Stat(entryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin entry not found: %s", entryPath)
	}

	logo.Info("Starting plugin:", plugin.Name, "path:", entryPath)
	cmd := exec.Command(entryPath)
	proto := protocol.NewStdioProtocol(cmd)

	if err := proto.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	my.protocols[plugin.Name] = proto

	// Initialize plugin with config
	if err := my.initializePlugin(plugin, proto); err != nil {
		proto.Close()
		delete(my.protocols, plugin.Name)
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	logo.Info("Plugin started and initialized successfully:", plugin.Name)
	return proto, nil
}

// initializePlugin sends initialize request with config to plugin
func (my *PluginManager) initializePlugin(plugin *types.Plugin, proto *protocol.StdioProtocol) error {
	// Read plugin config
	configPath := filepath.Join(plugin.Path, "config.yml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		// Config is optional
		logo.Info("No config file found for plugin:", plugin.Name)
		return nil
	}

	// Parse config as generic map
	var config map[string]any
	if err := yaml.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if len(config) == 0 {
		return nil
	}

	logo.Info("Initializing plugin:", plugin.Name, "with config")

	// Send initialize request
	result, err := proto.Call("initialize", config)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	logo.Info("Plugin initialized:", plugin.Name, "result:", result)
	return nil
}

// CallPlugin invokes a method on a plugin.
func (my *PluginManager) CallPlugin(plugin *types.Plugin, method string, params any) (any, error) {
	if plugin == nil {
		return nil, fmt.Errorf("plugin is nil")
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	// Ensure plugin is started
	proto, err := my.ensurePluginStarted(plugin)
	if err != nil {
		return nil, err
	}

	// Call the method
	logo.Debug("Calling plugin:", plugin.Name, "method:", method)
	return proto.Call(method, params)
}

// StartPlugin pre-starts a plugin without calling any method.
func (my *PluginManager) StartPlugin(plugin *types.Plugin) error {
	if plugin == nil {
		return fmt.Errorf("plugin is nil")
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	_, err := my.ensurePluginStarted(plugin)
	return err
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

// GetPluginsByType returns all plugins of a specific type.
func (my *PluginManager) GetPluginsByType(pluginType types.PluginType) []*types.Plugin {
	my.mu.RLock()
	defer my.mu.RUnlock()

	var plugins []*types.Plugin
	for _, plugin := range my.plugins {
		if plugin.Type == pluginType {
			plugins = append(plugins, plugin)
		}
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
	for _, proto := range my.protocols {
		if proto != nil {
			proto.Close()
		}
	}

	my.protocols = make(map[string]*protocol.StdioProtocol)
	return nil
}
