package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lixianmin/pc/pkg/types"
)

func TestPluginManager_NewManager(t *testing.T) {
	tests := []struct {
		name       string
		pluginsDir string
		wantErr    bool
	}{
		{
			name:       "valid plugins directory",
			pluginsDir: "/tmp/test/plugins",
			wantErr:    false,
		},
		{
			name:       "empty plugins directory",
			pluginsDir: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewPluginManager(tt.pluginsDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mgr == nil {
				t.Error("NewManager() returned nil for valid config")
			}
		})
	}
}

func TestPluginManager_DiscoverPlugins(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() string
		wantCount int
		wantErr   bool
	}{
		{
			name: "discover with valid plugin structure",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				pluginsDir := filepath.Join(tmpDir, "plugins")

				// Create test plugin structure
				llmDir := filepath.Join(pluginsDir, "llm", "openai")
				if err := os.MkdirAll(llmDir, 0755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				// Create plugin.yml
				pluginYml := `name: openai-llm
type: llm
enabled: true
version: 1.0.0
entry: ./bin/openai-llm
`
				if err := os.WriteFile(filepath.Join(llmDir, "plugin.yml"), []byte(pluginYml), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				return pluginsDir
			},
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "discover with multiple plugins",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				pluginsDir := filepath.Join(tmpDir, "plugins")

				// Create first plugin
				llmDir := filepath.Join(pluginsDir, "llm", "openai")
				if err := os.MkdirAll(llmDir, 0755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				pluginYml1 := `name: openai-llm
type: llm
enabled: true
version: 1.0.0
entry: ./bin/openai-llm
`
				if err := os.WriteFile(filepath.Join(llmDir, "plugin.yml"), []byte(pluginYml1), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				// Create second plugin
				channelDir := filepath.Join(pluginsDir, "channel", "telegram")
				if err := os.MkdirAll(channelDir, 0755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}
				pluginYml2 := `name: telegram-bot
type: channel
enabled: true
version: 1.0.0
entry: ./bin/telegram-bot
`
				if err := os.WriteFile(filepath.Join(channelDir, "plugin.yml"), []byte(pluginYml2), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				return pluginsDir
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "discover with invalid plugin.yml",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				pluginsDir := filepath.Join(tmpDir, "plugins")

				llmDir := filepath.Join(pluginsDir, "llm", "openai")
				if err := os.MkdirAll(llmDir, 0755); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				// Invalid plugin.yml (missing required fields)
				pluginYml := `name: openai-llm
enabled: true
`
				if err := os.WriteFile(filepath.Join(llmDir, "plugin.yml"), []byte(pluginYml), 0644); err != nil {
					t.Fatalf("setup failed: %v", err)
				}

				return pluginsDir
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "discover with non-existent directory",
			setupFunc: func() string {
				return "/non/existent/plugins"
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginsDir := tt.setupFunc()

			mgr, err := NewPluginManager(pluginsDir)
			if err != nil {
				t.Fatalf("NewManager() failed: %v", err)
			}

			plugins, err := mgr.Discover()
			if (err != nil) != tt.wantErr {
				t.Errorf("Discover() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(plugins) != tt.wantCount {
				t.Errorf("Discover() count = %d, want %d", len(plugins), tt.wantCount)
			}
		})
	}
}

func TestPluginManager_LoadPlugin(t *testing.T) {
	tests := []struct {
		name      string
		pluginYml string
		wantName  string
		wantType  types.PluginType
		wantErr   bool
	}{
		{
			name: "load valid LLM plugin",
			pluginYml: `name: openai-llm
type: llm
enabled: true
version: 1.0.0
entry: ./bin/openai-llm
`,
			wantName: "openai-llm",
			wantType: types.PluginTypeLLM,
			wantErr:  false,
		},
		{
			name: "load valid Channel plugin",
			pluginYml: `name: telegram-bot
type: channel
enabled: true
version: 1.0.0
entry: ./bin/telegram-bot
`,
			wantName: "telegram-bot",
			wantType: types.PluginTypeChannel,
			wantErr:  false,
		},
		{
			name: "load valid Tool plugin",
			pluginYml: `name: linux-shell
type: tool
enabled: true
version: 1.0.0
entry: ./bin/shell
`,
			wantName: "linux-shell",
			wantType: types.PluginTypeTool,
			wantErr:  false,
		},
		{
			name: "load plugin with missing type",
			pluginYml: `name: test-plugin
enabled: true
version: 1.0.0
entry: ./bin/test
`,
			wantName: "test-plugin",
			wantType: types.PluginTypeTool, // default
			wantErr:  false,
		},
		{
			name:      "load plugin with invalid YAML",
			pluginYml: `invalid yaml: [unclosed`,
			wantName:  "",
			wantType:  "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pluginDir := filepath.Join(tmpDir, "plugin")

			if err := os.MkdirAll(pluginDir, 0755); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Write plugin.yml
			if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yml"), []byte(tt.pluginYml), 0644); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			mgr, err := NewPluginManager("/tmp/plugins")
			if err != nil {
				t.Fatalf("NewManager() failed: %v", err)
			}

			plugin, err := mgr.LoadPlugin(pluginDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadPlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if plugin == nil {
					t.Error("LoadPlugin() returned nil plugin for valid config")
				}
				if plugin != nil && plugin.Name != tt.wantName {
					t.Errorf("LoadPlugin() name = %v, want %v", plugin.Name, tt.wantName)
				}
				if plugin != nil && plugin.Type != tt.wantType {
					t.Errorf("LoadPlugin() type = %v, want %v", plugin.Type, tt.wantType)
				}
			}
		})
	}
}

func TestPluginManager_CallPlugin(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*PluginManager, *types.Plugin)
		method    string
		params    any
		wantErr   bool
	}{
		{
			name: "call plugin with no plugin loaded",
			setupFunc: func() (*PluginManager, *types.Plugin) {
				tmpDir := t.TempDir()
				mgr, _ := NewPluginManager(filepath.Join(tmpDir, "plugins"))
				return mgr, nil
			},
			method:  "test",
			params:  map[string]any{"key": "value"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, plugin := tt.setupFunc()

			_, err := mgr.CallPlugin(plugin, tt.method, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CallPlugin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
