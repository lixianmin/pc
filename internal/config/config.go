package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SystemPromptConfig holds system prompt configuration.
// Deprecated: Use AgentPath instead. Kept for backward compatibility.
type SystemPromptConfig struct {
	File string `yaml:"file,omitempty"` // Path to agents.md file
}

// GetFile returns the system prompt file path, or default if empty.
// Deprecated: Use GetAgentPath instead.
func (s SystemPromptConfig) GetFile() string {
	if s.File != "" {
		return s.File
	}
	return ""
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
// Deprecated: Agent configuration is now loaded from agent.md file specified by AgentPath.
type AgentConfig struct {
	Name        string   `yaml:"name,omitempty"`        // Agent name (deprecated, use agent.md)
	Profession  string   `yaml:"profession,omitempty"`  // Agent profession (deprecated, use agent.md)
	Personality []string `yaml:"personality,omitempty"` // Personality traits (deprecated, use agent.md)
}

// Config holds the complete configuration.
type Config struct {
	AgentPath    string             `yaml:"agent_path,omitempty"`    // Path to agent.md file
	Workspace    string             `yaml:"workspace"`               // Workspace directory
	Log          LogConfig          `yaml:"log"`                     // Logging configuration
	SkillsDir    string             `yaml:"skills_dir"`              // Skills directory
	PluginsDir   string             `yaml:"plugins_dir"`             // Plugins directory
	LLMTimeout   int                `yaml:"llm_timeout,omitempty"`   // LLM request timeout in seconds (default: 120)
	SystemPrompt SystemPromptConfig `yaml:"system_prompt,omitempty"` // Deprecated: Use AgentPath instead
	Agent        AgentConfig        `yaml:"agent,omitempty"`         // Deprecated: Use agent.md file instead
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		AgentPath:  "~/.pc/agent.md",
		Workspace:  "~/workspace",
		SkillsDir:  "~/.pc/skills",
		PluginsDir: "~/.pc/plugins",
		LLMTimeout: 120,
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
		cfg.AgentPath = expandPath(cfg.AgentPath)
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
	cfg.AgentPath = expandPath(cfg.AgentPath)

	return cfg, nil
}

// Validate validates the configuration.
func (my *Config) Validate() error {
	// Note: Agent configuration is now loaded from agent.md file
	// The agent.md file existence is validated at runtime by the agent manager

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
// For example: PC_AGENT_PATH, PC_LOG_LEVEL.
func applyEnvOverrides(cfg *Config) {
	// Agent path (path to agent.md file)
	if v := os.Getenv("PC_AGENT_PATH"); v != "" {
		cfg.AgentPath = v
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

	// System prompt file (deprecated, use PC_AGENT_PATH instead)
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

// GetAgentPath returns the path to the agent.md file.
func (my *Config) GetAgentPath() string {
	// For backward compatibility: if AgentPath is not set but SystemPrompt.File is,
	// use SystemPrompt.File as fallback
	if my.AgentPath == "" && my.SystemPrompt.File != "" {
		return my.SystemPrompt.File
	}
	// Default to ~/.pc/agent.md if neither is set
	if my.AgentPath == "" {
		return "~/.pc/agent.md"
	}
	return my.AgentPath
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

// GetLLMTimeout returns the LLM request timeout in seconds.
func (my *Config) GetLLMTimeout() int {
	if my.LLMTimeout <= 0 {
		return 120
	}
	return my.LLMTimeout
}
