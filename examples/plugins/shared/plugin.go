package shared

import (
	"fmt"
	"os"
	"strings"

	"github.com/lixianmin/pc/pkg/types"
)

// CreatePluginStructure creates the directory structure for a plugin.
func CreatePluginStructure(pluginDir, configDir string) error {
	// Create plugin directory
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	return nil
}

// GeneratePluginYML generates the plugin.yml file.
func GeneratePluginYML(plugin *types.Plugin, ymlPath string) error {
	yml := fmt.Sprintf("name: %s\ntype: %s\nenabled: %v\nversion: %s\nentry: %s\n",
		plugin.Name, plugin.Type, plugin.Enabled, plugin.Version, plugin.Entry)

	if err := os.WriteFile(ymlPath, []byte(yml), 0644); err != nil {
		return fmt.Errorf("failed to write plugin.yml: %w", err)
	}

	return nil
}

// GenerateConfigTemplate generates a config template file.
func GenerateConfigTemplate(pluginType string) string {
	if pluginType == "llm" {
		return "# OpenAI API Configuration\napi_key: your-api-key-here\nmodel: gpt-4\nbase_url: https://api.openai.com/v1\n"
	}

	if pluginType == "channel" {
		return "# Telegram Bot Configuration\nbot_token: your-bot-token-here\nchat_id: your-chat-id-here\n"
	}

	return "# Configuration Template\n"
}

// ContainsKey checks if a line contains a key in comment format.
func ContainsKey(line, key string) bool {
	// Check if line starts with "# " (comment)
	if len(line) > 2 && line[0] == '#' && line[1] == ' ' {
		// Skip if it's a comment line
		return false
	}

	// Check if line is a key-value pair
	keyWithColon := key + ":"
	if strings.Contains(line, keyWithColon) {
		// Find position of the colon
		idx := strings.Index(line, keyWithColon)
		colonPos := idx + len(key)
		// Check if there's a space after the colon
		if colonPos < len(line)-1 && line[colonPos+1] == ' ' {
			return true
		}
	}
	return false
}
