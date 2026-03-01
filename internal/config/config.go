package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SystemPromptConfig holds system prompt configuration.
type SystemPromptConfig struct {
	File string `yaml:"file"` // Path to agents.md file
}

// GetFile returns the system prompt file path, or default if empty.
func (s SystemPromptConfig) GetFile() string {
	if s.File != "" {
		return s.File
	}
	return "~/.pc/agents.md"
}

// LogLevel represents the logging level.
type LogLevel string

const (
	// DebugLevel is the most verbose logging level.
	DebugLevel LogLevel = "debug"
	// InfoLevel is the default logging level.
	InfoLevel LogLevel = "info"
	// WarnLevel is for warning messages.
	WarnLevel LogLevel = "warn"
	// ErrorLevel is for error messages.
	ErrorLevel LogLevel = "error"
)

// LogConfig holds logging configuration.
type LogConfig struct {
	Level  LogLevel `yaml:"level"`  // Logging level (debug, info, warn, error)
	Output string   `yaml:"output"` // Output destination (stdout, file, or path)
}

// AgentConfig holds agent-specific configuration.
type AgentConfig struct {
	Name        string   `yaml:"name"`        // Agent name
	Profession  string   `yaml:"profession"`  // Agent profession (optional)
	Personality []string `yaml:"personality"` // Personality traits (optional)
}

// Config holds the complete configuration.
type Config struct {
	Agent        AgentConfig        `yaml:"agent"`         // Agent configuration
	Workspace    string             `yaml:"workspace"`     // Workspace directory
	Log          LogConfig          `yaml:"log"`           // Logging configuration
	SkillsDir    string             `yaml:"skills_dir"`    // Skills directory
	PluginsDir   string             `yaml:"plugins_dir"`   // Plugins directory
	SystemPrompt SystemPromptConfig `yaml:"system_prompt"` // System prompt configuration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Agent: AgentConfig{
			Name:        "PersonalClaw",
			Profession:  "通用助手",
			Personality: []string{"友好", "专业"},
		},
		Workspace:  "~/workspace",
		SkillsDir:  "~/.pc/skills",
		PluginsDir: "~/.pc/plugins",
		SystemPrompt: SystemPromptConfig{
			File: "~/.pc/agents.md",
		},
		Log: LogConfig{
			Level:  InfoLevel,
			Output: "stdout",
		},
	}
}

// Load loads configuration from a YAML file.
// If the file doesn't exist, it returns a default config.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Apply environment variable overrides even when file doesn't exist
		applyEnvOverrides(cfg)
		cfg.Workspace = expandPath(cfg.Workspace)
		cfg.SkillsDir = expandPath(cfg.SkillsDir)
		cfg.PluginsDir = expandPath(cfg.PluginsDir)
		cfg.SystemPrompt.File = expandPath(cfg.SystemPrompt.File)
		return cfg, nil
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)

	// Expand paths with ~
	cfg.Workspace = expandPath(cfg.Workspace)
	cfg.SkillsDir = expandPath(cfg.SkillsDir)
	cfg.PluginsDir = expandPath(cfg.PluginsDir)
	cfg.SystemPrompt.File = expandPath(cfg.SystemPrompt.File)

	return cfg, nil
}

// Validate validates the configuration.
func (my *Config) Validate() error {
	if my.Agent.Name == "" {
		return fmt.Errorf("agent.name is required")
	}

	if my.Workspace == "" {
		return fmt.Errorf("workspace is required")
	}

	if my.SkillsDir == "" {
		return fmt.Errorf("skills_dir is required")
	}

	if my.PluginsDir == "" {
		return fmt.Errorf("plugins_dir is required")
	}

	// Validate log level
	switch my.Log.Level {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel:
		// Valid
	default:
		return fmt.Errorf("invalid log level: %s", my.Log.Level)
	}

	return nil
}

// Save saves the configuration to a YAML file.
func (my *Config) Save(path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(my)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write file
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// applyEnvOverrides applies environment variable overrides to the config.
// Environment variables should be in the format PC_<FIELD>_<SUBFIELD>.
// For example: PC_AGENT_NAME, PC_LOG_LEVEL.
func applyEnvOverrides(cfg *Config) {
	// Agent name
	if v := os.Getenv("PC_AGENT_NAME"); v != "" {
		cfg.Agent.Name = v
	}

	// Agent profession
	if v := os.Getenv("PC_AGENT_PROFESSION"); v != "" {
		cfg.Agent.Profession = v
	}

	// Workspace
	if v := os.Getenv("PC_WORKSPACE"); v != "" {
		cfg.Workspace = v
	}

	// Log level
	if v := os.Getenv("PC_LOG_LEVEL"); v != "" {
		cfg.Log.Level = LogLevel(v)
	}

	// Log output
	if v := os.Getenv("PC_LOG_OUTPUT"); v != "" {
		cfg.Log.Output = v
	}

	// Skills dir
	if v := os.Getenv("PC_SKILLS_DIR"); v != "" {
		cfg.SkillsDir = v
	}

	// Plugins dir
	if v := os.Getenv("PC_PLUGINS_DIR"); v != "" {
		cfg.PluginsDir = v
	}

	// System prompt file
	if v := os.Getenv("PC_SYSTEM_PROMPT_FILE"); v != "" {
		cfg.SystemPrompt.File = v
	}
}

// expandPath expands a path with ~ to the user's home directory.
func expandPath(path string) string {
	if path == "~" || path == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~/" {
			return home
		}
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// GetAgentName returns the agent name.
func (my *Config) GetAgentName() string {
	return my.Agent.Name
}

// GetAgentProfession returns the agent profession.
func (my *Config) GetAgentProfession() string {
	return my.Agent.Profession
}

// GetAgentPersonality returns the agent personality traits.
func (my *Config) GetAgentPersonality() []string {
	return my.Agent.Personality
}

// GetWorkspace returns the workspace directory.
func (my *Config) GetWorkspace() string {
	return my.Workspace
}

// GetSkillsDir returns the skills directory.
func (my *Config) GetSkillsDir() string {
	return my.SkillsDir
}

// GetPluginsDir returns the plugins directory.
func (my *Config) GetPluginsDir() string {
	return my.PluginsDir
}

// GetLogLevel returns the log level.
func (my *Config) GetLogLevel() LogLevel {
	return my.Log.Level
}

// GetLogOutput returns the log output destination.
func (my *Config) GetLogOutput() string {
	return my.Log.Output
}

// GetSystemPromptFile returns the system prompt file path.
func (my *Config) GetSystemPromptFile() string {
	return my.SystemPrompt.GetFile()
}
