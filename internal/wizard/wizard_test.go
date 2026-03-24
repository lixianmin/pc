package wizard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lixianmin/pc/internal/config"
)

func TestCheckConfigExists(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() string
		wantExists bool
	}{
		{
			name: "config exists",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				_ = os.WriteFile(configPath, []byte("test"), 0644)
				return configPath
			},
			wantExists: true,
		},
		{
			name: "config does not exist",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yml")
				return configPath
			},
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setupFunc()

			exists := ConfigExists(configPath)

			if exists != tt.wantExists {
				t.Errorf("ConfigExists() = %v, want %v", exists, tt.wantExists)
			}
		})
	}
}

func TestCreateDirectory(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create new directory",
			path:    filepath.Join(t.TempDir(), "new-dir"),
			wantErr: false,
		},
		{
			name:    "create nested directory",
			path:    filepath.Join(t.TempDir(), "parent", "child", "grandchild"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateDirectory(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateDirectory() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify directory was created
			if !tt.wantErr {
				if _, err := os.Stat(tt.path); os.IsNotExist(err) {
					t.Errorf("CreateDirectory() directory not created")
				}
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name: "save valid config",
			cfg: &config.Config{
				Agent: config.AgentConfig{
					Name: "TestAgent",
				},
				Workspace: "/tmp/workspace",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")

			err := SaveConfig(tt.cfg, configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify config was saved
			if !tt.wantErr {
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("SaveConfig() config not saved")
				}
			}
		})
	}
}

func TestValidateAgentName(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		wantErr   bool
	}{
		{
			name:      "valid agent name",
			agentName: "PersonalClaw",
			wantErr:   false,
		},
		{
			name:      "valid name with spaces",
			agentName: "My Agent",
			wantErr:   false,
		},
		{
			name:      "valid unicode name",
			agentName: "我的助手",
			wantErr:   false,
		},
		{
			name:      "empty name",
			agentName: "",
			wantErr:   true,
		},
		{
			name:      "name with only spaces",
			agentName: "   ",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAgentName(tt.agentName)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAgentName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetDefaultWorkspace(t *testing.T) {
	tests := []struct {
		name     string
		homeDir  string
		wantPath string
	}{
		{
			name:     "default workspace",
			homeDir:  "/home/user",
			wantPath: "/home/user/workspace",
		},
		{
			name:     "custom home",
			homeDir:  "/custom/home",
			wantPath: filepath.Join("/custom/home", "workspace"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set HOME environment variable for testing
			t.Setenv("HOME", tt.homeDir)

			workspace := GetDefaultWorkspace()

			if workspace != tt.wantPath {
				t.Errorf("GetDefaultWorkspace() = %v, want %v", workspace, tt.wantPath)
			}
		})
	}
}

func TestGenerateConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		forceDir string
		wantPath string
	}{
		{
			name:     "default config path",
			forceDir: "",
			wantPath: filepath.Join(os.Getenv("HOME"), ".pc", "config.yml"),
		},
		{
			name:     "custom config path",
			forceDir: "/custom/config/dir",
			wantPath: filepath.Join("/custom/config/dir", "config.yml"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save HOME and set custom dir
			home := os.Getenv("HOME")
			t.Setenv("HOME", home)

			configPath := GenerateConfigPath(tt.forceDir)

			if configPath != tt.wantPath {
				t.Errorf("GenerateConfigPath() = %v, want %v", configPath, tt.wantPath)
			}
		})
	}
}

func TestGeneratePluginDir(t *testing.T) {
	tests := []struct {
		name     string
		forceDir string
		wantPath string
	}{
		{
			name:     "default plugin dir",
			forceDir: "",
			wantPath: filepath.Join(os.Getenv("HOME"), ".pc", "plugins"),
		},
		{
			name:     "custom plugin dir",
			forceDir: "/custom/config/dir",
			wantPath: filepath.Join("/custom/config/dir", "plugins"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save HOME and set custom dir
			home := os.Getenv("HOME")
			t.Setenv("HOME", home)

			pluginDir := GeneratePluginDir(tt.forceDir)

			if pluginDir != tt.wantPath {
				t.Errorf("GeneratePluginDir() = %v, want %v", pluginDir, tt.wantPath)
			}
		})
	}
}

func TestGenerateExamplePlugins(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "generate example plugins",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			pluginDir := filepath.Join(tmpDir, "plugins")

			err := generateExamplePlugins(pluginDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateExamplePlugins() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check glm-llm plugin
			llmPluginDir := filepath.Join(pluginDir, "llm", "glm-llm")
			llmPluginYml := filepath.Join(llmPluginDir, "plugin.yml")
			llmConfigYml := filepath.Join(llmPluginDir, "config", "config.yml")

			if _, err := os.Stat(llmPluginYml); os.IsNotExist(err) {
				t.Error("glm-llm plugin.yml not created")
			}
			if _, err := os.Stat(llmConfigYml); os.IsNotExist(err) {
				t.Error("glm-llm config.yml not created")
			}

			// Check telegram-bot plugin
			telegramPluginDir := filepath.Join(pluginDir, "channel", "telegram-bot")
			telegramPluginYml := filepath.Join(telegramPluginDir, "plugin.yml")
			telegramConfigYml := filepath.Join(telegramPluginDir, "config", "config.yml")

			if _, err := os.Stat(telegramPluginYml); os.IsNotExist(err) {
				t.Error("telegram-bot plugin.yml not created")
			}
			if _, err := os.Stat(telegramConfigYml); os.IsNotExist(err) {
				t.Error("telegram-bot config.yml not created")
			}

			// Verify content
			if !tt.wantErr {
				data, _ := os.ReadFile(llmConfigYml)
				if !strings.Contains(string(data), "glm-4.7") {
					t.Error("glm-llm config.yml missing glm-4.7 model")
				}

				data, _ = os.ReadFile(telegramConfigYml)
				if !strings.Contains(string(data), "bot_token") {
					t.Error("telegram-bot config.yml missing bot_token field")
				}
			}
		})
	}
}
