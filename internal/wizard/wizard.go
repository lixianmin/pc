package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lixianmin/pc/internal/config"
)

// defaultPCDir is the default .pc directory.
const defaultPCDir = ".pc"

// ConfigExists checks if the config file exists.
func ConfigExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// CreateDirectory creates a directory if it doesn't exist.
func CreateDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// SaveConfig saves the configuration to a file.
func SaveConfig(cfg *config.Config, path string) error {
	if err := cfg.Save(path); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// ValidateAgentName validates the agent name.
func ValidateAgentName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("agent name cannot be empty")
	}
	// Check if name is only whitespace
	if name == "" {
		return fmt.Errorf("agent name cannot be only whitespace")
	}
	// Simple validation: not empty and not only whitespace
	return nil
}

// GetDefaultWorkspace returns the default workspace path.
func GetDefaultWorkspace() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "workspace"
	}
	return filepath.Join(home, "workspace")
}

// GenerateConfigPath generates the config file path.
func GenerateConfigPath(forceDir string) string {
	if forceDir != "" {
		return filepath.Join(forceDir, "config.yml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "config.yml")
	}
	return filepath.Join(home, defaultPCDir, "config.yml")
}

// GeneratePluginDir generates the plugins directory path.
func GeneratePluginDir(forceDir string) string {
	if forceDir != "" {
		return filepath.Join(forceDir, "plugins")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", "plugins")
	}
	return filepath.Join(home, defaultPCDir, "plugins")
}

// RunWizard runs the initialization wizard.
func RunWizard() error {
	configPath := GenerateConfigPath("")

	// Check if config already exists
	if ConfigExists(configPath) {
		fmt.Println("Config file already exists at:", configPath)
		return nil
	}

	fmt.Println("Welcome to PersonalClaw!")
	fmt.Println("Let's set up your AI assistant.")

	// Ask for agent name
	agentName, err := promptAgentName()
	if err != nil {
		return fmt.Errorf("failed to get agent name: %w", err)
	}

	// Ask for profession (optional)
	profession := promptProfession()

	// Ask for personality (optional)
	personality := promptPersonality()

	// Ask for workspace
	workspace := promptWorkspace()

	// Create directories
	pluginDir := GeneratePluginDir("")
	workspaceDir := filepath.Dir(workspace)
	if workspaceDir == "." {
		workspaceDir = workspace
	}
	if err := CreateDirectory(pluginDir); err != nil {
		return fmt.Errorf("failed to create plugins directory: %w", err)
	}

	// Create workspace directory if it doesn't exist
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		if err := CreateDirectory(workspace); err != nil {
			return fmt.Errorf("failed to create workspace directory: %w", err)
		}
	}

	// Generate example plugins
	if err := generateExamplePlugins(pluginDir); err != nil {
		return fmt.Errorf("failed to generate example plugins: %w", err)
	}

	// Create config
	cfg := config.DefaultConfig()
	cfg.Agent.Name = agentName
	if profession != "" {
		cfg.Agent.Profession = profession
	}
	if len(personality) > 0 {
		cfg.Agent.Personality = personality
	}
	cfg.Workspace = workspace

	// Save config
	if err := SaveConfig(cfg, configPath); err != nil {
		return err
	}

	fmt.Println("Configuration saved to:", configPath)
	fmt.Println("You can now run 'pc' to start your assistant.")

	return nil
}

// generateExamplePlugins generates example plugins in the plugins directory.
func generateExamplePlugins(pluginDir string) error {
	fmt.Println("Generating example plugins...")

	// Generate glm-llm plugin
	llmPluginDir := filepath.Join(pluginDir, "llm", "glm-llm")
	if err := CreateDirectory(llmPluginDir); err != nil {
		return err
	}

	llmConfigDir := filepath.Join(llmPluginDir, "config")
	if err := CreateDirectory(llmConfigDir); err != nil {
		return err
	}

	llmPluginYml := `name: glm-llm
type: llm
enabled: true
version: 1.0.0
entry: ./bin/glm-llm
`
	if err := os.WriteFile(filepath.Join(llmPluginDir, "plugin.yml"), []byte(llmPluginYml), 0644); err != nil {
		return fmt.Errorf("failed to write glm-llm plugin.yml: %w", err)
	}

	llmConfigYml := `# GLM API Configuration
api_key: your-glm-api-key-here
model: glm-4.7
base_url: https://open.bigmodel.cn/api/coding/paas/v4
`
	if err := os.WriteFile(filepath.Join(llmConfigDir, "config.yml"), []byte(llmConfigYml), 0644); err != nil {
		return fmt.Errorf("failed to write glm-llm config.yml: %w", err)
	}

	// Generate telegram-bot plugin
	telegramPluginDir := filepath.Join(pluginDir, "channel", "telegram-bot")
	if err := CreateDirectory(telegramPluginDir); err != nil {
		return err
	}

	telegramConfigDir := filepath.Join(telegramPluginDir, "config")
	if err := CreateDirectory(telegramConfigDir); err != nil {
		return err
	}

	telegramPluginYml := `name: telegram-bot
type: channel
enabled: true
version: 1.0.0
entry: ./bin/telegram-bot
`
	if err := os.WriteFile(filepath.Join(telegramPluginDir, "plugin.yml"), []byte(telegramPluginYml), 0644); err != nil {
		return fmt.Errorf("failed to write telegram-bot plugin.yml: %w", err)
	}

	telegramConfigYml := `# Telegram Bot Configuration
bot_token: your-bot-token-here
chat_id: your-chat-id-here
`
	if err := os.WriteFile(filepath.Join(telegramConfigDir, "config.yml"), []byte(telegramConfigYml), 0644); err != nil {
		return fmt.Errorf("failed to write telegram-bot config.yml: %w", err)
	}

	fmt.Println("Example plugins generated:")
	fmt.Printf("  - glm-llm (LLM plugin)\n")
	fmt.Printf("  - telegram-bot (Channel plugin)\n")

	return nil
}

func promptAgentName() (string, error) {
	defaultName := "PersonalClaw"
	for {
		fmt.Printf("Agent name [%s]: ", defaultName)
		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		if input == "" {
			input = defaultName
		}

		if err := ValidateAgentName(input); err != nil {
			fmt.Println("Error:", err)
			continue
		}
		return input, nil
	}
}

func promptProfession() string {
	defaultProfession := "通用助手"
	fmt.Printf("Profession [%s] (optional, press Enter to skip): ", defaultProfession)
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultProfession
	}
	return input
}

func promptPersonality() []string {
	fmt.Println("Personality traits (optional, press Enter to skip):")
	fmt.Println("  Example: 友好, 专业, 幽默")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		return []string{"友好", "专业"}
	}

	// Split by comma
	traits := []string{}
	for _, trait := range strings.Split(input, ",") {
		trait = strings.TrimSpace(trait)
		if trait != "" {
			traits = append(traits, trait)
		}
	}

	return traits
}

func promptWorkspace() string {
	defaultWorkspace := GetDefaultWorkspace()
	for {
		fmt.Printf("Workspace [%s]: ", defaultWorkspace)
		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(input)

		if input == "" {
			input = defaultWorkspace
		}

		// Expand ~ in path
		if strings.HasPrefix(input, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Warning: Failed to expand ~, using as-is")
				continue
			}
			input = filepath.Join(home, input[2:])
		} else if input == "~" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Println("Warning: Failed to expand ~, using as-is")
				continue
			}
			input = home
		}

		return input
	}
}
