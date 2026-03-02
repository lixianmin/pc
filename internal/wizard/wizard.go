package wizard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lixianmin/pc/internal/config"
	"gopkg.in/yaml.v3"
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
	fmt.Println()

	// Ask for user's name (how agent should address user)
	userName := promptUserName()
	fmt.Println()

	// Ask for agent name (how user should address agent)
	agentName, err := promptAgentName()
	if err != nil {
		return fmt.Errorf("failed to get agent name: %w", err)
	}
	fmt.Println()

	// Ask for profession
	profession := promptProfession()
	fmt.Println()

	// Ask for personality
	personality := promptPersonality()
	fmt.Println()

	// Ask for workspace
	workspace := promptWorkspace()
	fmt.Println()

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

	// Get PC directory for saving agent.md
	pcDir := filepath.Dir(configPath)

	// Generate agent.md file
	agentMdPath := filepath.Join(pcDir, "agent.md")
	if err := generateAgentMd(pcDir, agentName, userName, profession, personality); err != nil {
		return fmt.Errorf("failed to generate agent.md: %w", err)
	}

	// Create config with agent_path
	cfg := config.DefaultConfig()
	cfg.AgentPath = agentMdPath
	cfg.Workspace = workspace

	// Save config
	if err := SaveConfig(cfg, configPath); err != nil {
		return err
	}

	fmt.Println("Configuration saved to:", configPath)
	fmt.Println("Agent definition saved to:", agentMdPath)
	fmt.Println("You can now run 'pc' to start your assistant.")

	return nil
}

// generateExamplePlugins generates example plugins in the plugins directory.
func generateExamplePlugins(pluginDir string) error {
	fmt.Println("Generating example plugins...")

	// Generate glm-llm plugin
	if err := generateLLMPlugin(pluginDir); err != nil {
		return err
	}

	// Generate telegram-bot plugin
	if err := generateTelegramPlugin(pluginDir); err != nil {
		return err
	}

	fmt.Println("Example plugins generated:")
	fmt.Printf("  - glm-llm (LLM plugin)\n")
	fmt.Printf("  - telegram-bot (Channel plugin)\n")

	return nil
}

// pluginMeta represents plugin metadata for YAML serialization
type pluginMeta struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Enabled bool   `yaml:"enabled"`
	Version string `yaml:"version"`
	Entry   string `yaml:"entry"`
}

// llmConfig represents LLM plugin configuration
type llmConfig struct {
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	BaseURL string `yaml:"base_url"`
}

// telegramConfig represents Telegram plugin configuration
type telegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
}

// generateLLMPlugin generates the GLM LLM example plugin
func generateLLMPlugin(pluginDir string) error {
	llmPluginDir := filepath.Join(pluginDir, "llm", "glm-llm")
	if err := CreateDirectory(llmPluginDir); err != nil {
		return err
	}

	llmConfigDir := filepath.Join(llmPluginDir, "config")
	if err := CreateDirectory(llmConfigDir); err != nil {
		return err
	}

	// Create and marshal plugin.yml
	meta := pluginMeta{
		Name:    "glm-llm",
		Type:    "llm",
		Enabled: true,
		Version: "1.0.0",
		Entry:   "./bin/glm-llm",
	}

	metaBytes, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal glm-llm plugin meta: %w", err)
	}

	if err := os.WriteFile(filepath.Join(llmPluginDir, "plugin.yml"), metaBytes, 0644); err != nil {
		return fmt.Errorf("failed to write glm-llm plugin.yml: %w", err)
	}

	// Create and marshal config.yml
	cfg := llmConfig{
		APIKey:  "your-glm-api-key-here",
		Model:   "glm-4.7",
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
	}

	configBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal glm-llm config: %w", err)
	}

	if err := os.WriteFile(filepath.Join(llmConfigDir, "config.yml"), configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write glm-llm config.yml: %w", err)
	}

	return nil
}

// generateTelegramPlugin generates the Telegram bot example plugin
func generateTelegramPlugin(pluginDir string) error {
	telegramPluginDir := filepath.Join(pluginDir, "channel", "telegram-bot")
	if err := CreateDirectory(telegramPluginDir); err != nil {
		return err
	}

	telegramConfigDir := filepath.Join(telegramPluginDir, "config")
	if err := CreateDirectory(telegramConfigDir); err != nil {
		return err
	}

	// Create and marshal plugin.yml
	meta := pluginMeta{
		Name:    "telegram-bot",
		Type:    "channel",
		Enabled: true,
		Version: "1.0.0",
		Entry:   "./bin/telegram-bot",
	}

	metaBytes, err := yaml.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram-bot plugin meta: %w", err)
	}

	if err := os.WriteFile(filepath.Join(telegramPluginDir, "plugin.yml"), metaBytes, 0644); err != nil {
		return fmt.Errorf("failed to write telegram-bot plugin.yml: %w", err)
	}

	// Create and marshal config.yml
	cfg := telegramConfig{
		BotToken: "your-bot-token-here",
		ChatID:   "your-chat-id-here",
	}

	configBytes, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram-bot config: %w", err)
	}

	if err := os.WriteFile(filepath.Join(telegramConfigDir, "config.yml"), configBytes, 0644); err != nil {
		return fmt.Errorf("failed to write telegram-bot config.yml: %w", err)
	}

	return nil
}

// generateAgentMd generates the agent.md file with agent configuration.
func generateAgentMd(pcDir string, agentName, userName, profession string, personality []string) error {
	agentMdPath := filepath.Join(pcDir, "agent.md")

	// Build the markdown content
	var content strings.Builder

	// Title (Agent Name)
	content.WriteString("# ")
	content.WriteString(agentName)
	content.WriteString("\n\n")

	// User section (who the agent is talking to)
	content.WriteString("## User\n")
	content.WriteString("你的用户是 **")
	content.WriteString(userName)
	content.WriteString("**。请用这个名字称呼用户。\n\n")

	// Profession section
	content.WriteString("## Profession\n")
	content.WriteString(profession)
	content.WriteString("\n\n")

	// Personality section
	content.WriteString("## Personality\n")
	for _, trait := range personality {
		content.WriteString("- ")
		content.WriteString(trait)
		content.WriteString("\n")
	}
	content.WriteString("\n")

	// Instructions section with personalized greeting
	content.WriteString("## Instructions\n")
	content.WriteString("你是 **")
	content.WriteString(agentName)
	content.WriteString("**，一位")
	content.WriteString(profession)
	content.WriteString("。你的用户是 **")
	content.WriteString(userName)
	content.WriteString("**。\n\n")
	content.WriteString("在与用户交流时，请：\n")
	content.WriteString("1. 用友好、专业的语气回应\n")
	content.WriteString("2. 适时使用用户的名字来建立亲切感\n")
	content.WriteString("3. 根据你的性格特点调整回应风格\n")
	content.WriteString("4. 主动提供帮助，预判用户需求\n")

	// Write to file
	if err := os.WriteFile(agentMdPath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write agent.md: %w", err)
	}

	return nil
}

func promptUserName() string {
	defaultName := "主人"
	fmt.Printf("您的名字 [%s] (Agent将用这个名字称呼您): ", defaultName)
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		input = defaultName
	}
	return input
}

func promptAgentName() (string, error) {
	defaultName := "PersonalClaw"
	for {
		fmt.Printf("Agent的名字 [%s] (您将用这个名字称呼Agent): ", defaultName)
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
	fmt.Printf("Agent的职业/人设 [%s] (如：编程助手、生活管家、学习伙伴): ", defaultProfession)
	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultProfession
	}
	return input
}

func promptPersonality() []string {
	fmt.Println("Agent的性格特点 (可选，用逗号分隔，直接回车使用默认值):")
	fmt.Println("  示例: 友好, 专业, 幽默, 严谨, 活泼")

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
