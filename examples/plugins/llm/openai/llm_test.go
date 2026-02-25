package openai

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lixianmin/pc/pkg/types"
	"github.com/lixianmin/pc/examples/plugins/shared"
)

func TestCreatePluginStructure(t *testing.T) {
	tests := []struct {
		name    string
		baseDir string
		wantErr bool
	}{
		{
			name:    "create complete plugin structure",
			baseDir: t.TempDir(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginDir := filepath.Join(tt.baseDir, "plugins", "llm", "openai")
			configDir := filepath.Join(pluginDir, "config")

			err := shared.CreatePluginStructure(pluginDir, configDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePluginStructure() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
					t.Error("Plugin directory not created")
				}
				if _, err := os.Stat(configDir); os.IsNotExist(err) {
					t.Error("Config directory not created")
				}
			}
		})
	}
}

func TestGeneratePluginYML(t *testing.T) {
	tests := []struct {
		name     string
		plugin  types.Plugin
		wantYml string
	}{
		{
			name: "generate valid plugin yml",
			plugin: types.Plugin{
				Name:    "openai-llm",
				Type:    types.PluginTypeLLM,
				Enabled: true,
				Version: "1.0.0",
				Entry:   "./bin/openai-llm",
			},
			wantYml: "name: openai-llm\ntype: llm\nenabled: true\nversion: 1.0.0\nentry: ./bin/openai-llm\n",
		},
		{
			name: "generate channel plugin yml",
			plugin: types.Plugin{
				Name:    "telegram-bot",
				Type:    types.PluginTypeChannel,
				Enabled: true,
				Version: "1.0.0",
				Entry:   "./bin/telegram-bot",
			},
			wantYml: "name: telegram-bot\ntype: channel\nenabled: true\nversion: 1.0.0\nentry: ./bin/telegram-bot\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			ymlPath := filepath.Join(tmpDir, "plugin.yml")

			err := shared.GeneratePluginYML(&tt.plugin, ymlPath)
			if err != nil {
				t.Errorf("GeneratePluginYML() error = %v", err)
			}

			if err == nil {
				data, _ := os.ReadFile(ymlPath)
				if string(data) != tt.wantYml {
					t.Errorf("GeneratePluginYML() content mismatch")
				}
			}
		})
	}
}

func TestGenerateConfigTemplate(t *testing.T) {
	tests := []struct {
		name       string
		pluginType string
		wantKeys   []string
	}{
		{
			name:       "generate LLM config template",
			pluginType: "llm",
			wantKeys:   []string{"api_key", "model", "base_url"},
		},
		{
			name:       "generate Channel config template",
			pluginType: "channel",
			wantKeys:   []string{"bot_token", "chat_id"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := shared.GenerateConfigTemplate(tt.pluginType)

			for _, key := range tt.wantKeys {
				found := false
				for _, line := range strings.Split(template, "\n") {
					if shared.ContainsKey(line, key) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GenerateConfigTemplate() missing key %s", key)
				}
			}
		})
	}
}
