package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name    string
		want    string
		got     string
		notZero bool
	}{
		{"agent name", "PersonalClaw", cfg.Agent.Name, true},
		{"agent profession", "通用助手", cfg.Agent.Profession, true},
		{"workspace", "~/workspace", cfg.Workspace, true},
		{"skills dir", "~/.pc/skills", cfg.SkillsDir, true},
		{"plugins dir", "~/.pc/plugins", cfg.PluginsDir, true},
		{"log level", string(InfoLevel), string(cfg.Log.Level), true},
		{"log output", "stdout", cfg.Log.Output, true},
		{"personality", "", "", false}, // Check not empty
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.notZero {
				if tt.want != tt.got {
					t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
				}
			} else {
				if len(cfg.Agent.Personality) == 0 {
					t.Errorf("%s should not be empty", tt.name)
				}
			}
		})
	}
}

func TestLoadAndSave(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Create a config
	cfg := &Config{
		Agent: AgentConfig{
			Name:        "TestAgent",
			Profession:  "Test Profession",
			Personality: []string{"Friendly", "Helpful"},
		},
		Workspace:  "~/test/workspace",
		SkillsDir:  "~/.pc/test/skills",
		PluginsDir: "~/.pc/test/plugins",
		Log: LogConfig{
			Level:  DebugLevel,
			Output: "/tmp/test.log",
		},
	}

	// Save config
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load config
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify loaded config matches saved config
	if loaded.Agent.Name != cfg.Agent.Name {
		t.Errorf("Agent.Name = %v, want %v", loaded.Agent.Name, cfg.Agent.Name)
	}
	if loaded.Agent.Profession != cfg.Agent.Profession {
		t.Errorf("Agent.Profession = %v, want %v", loaded.Agent.Profession, cfg.Agent.Profession)
	}
	if loaded.Log.Level != cfg.Log.Level {
		t.Errorf("Log.Level = %v, want %v", loaded.Log.Level, cfg.Log.Level)
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.yml")

	// Load non-existent file should return default config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify it's a default config
	if cfg.Agent.Name != "PersonalClaw" {
		t.Errorf("Agent.Name = %v, want PersonalClaw", cfg.Agent.Name)
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yml")

	// Write invalid YAML
	err := os.WriteFile(configPath, []byte("invalid: yaml: content: [unclosed"), 0644)
	if err != nil {
		t.Fatalf("failed to write invalid YAML: %v", err)
	}

	// Load should return error
	_, err = Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  InfoLevel,
					Output: "stdout",
				},
			},
			wantErr: false,
		},
		{
			name: "missing agent name",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  InfoLevel,
					Output: "stdout",
				},
			},
			wantErr: true,
		},
		{
			name: "missing workspace",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  InfoLevel,
					Output: "stdout",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  LogLevel("invalid"),
					Output: "stdout",
				},
			},
			wantErr: true,
		},
		{
			name: "valid debug level",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  DebugLevel,
					Output: "stdout",
				},
			},
			wantErr: false,
		},
		{
			name: "valid warn level",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  WarnLevel,
					Output: "stdout",
				},
			},
			wantErr: false,
		},
		{
			name: "valid error level",
			cfg: &Config{
				Agent: AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  ErrorLevel,
					Output: "stdout",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envKey   string
		envValue string
		checkFn  func(*Config, string)
	}{
		{
			name:     "PC_AGENT_NAME",
			envKey:   "PC_AGENT_NAME",
			envValue: "EnvAgent",
			checkFn: func(cfg *Config, v string) {
				if cfg.Agent.Name != v {
					t.Errorf("Agent.Name = %v, want %v", cfg.Agent.Name, v)
				}
			},
		},
		{
			name:     "PC_AGENT_PROFESSION",
			envKey:   "PC_AGENT_PROFESSION",
			envValue: "EnvProfession",
			checkFn: func(cfg *Config, v string) {
				if cfg.Agent.Profession != v {
					t.Errorf("Agent.Profession = %v, want %v", cfg.Agent.Profession, v)
				}
			},
		},
		{
			name:     "PC_WORKSPACE",
			envKey:   "PC_WORKSPACE",
			envValue: "/env/workspace",
			checkFn: func(cfg *Config, v string) {
				if cfg.Workspace != v {
					t.Errorf("Workspace = %v, want %v", cfg.Workspace, v)
				}
			},
		},
		{
			name:     "PC_LOG_LEVEL",
			envKey:   "PC_LOG_LEVEL",
			envValue: "debug",
			checkFn: func(cfg *Config, v string) {
				if string(cfg.Log.Level) != v {
					t.Errorf("Log.Level = %v, want %v", cfg.Log.Level, v)
				}
			},
		},
		{
			name:     "PC_LOG_OUTPUT",
			envKey:   "PC_LOG_OUTPUT",
			envValue: "/tmp/env.log",
			checkFn: func(cfg *Config, v string) {
				if cfg.Log.Output != v {
					t.Errorf("Log.Output = %v, want %v", cfg.Log.Output, v)
				}
			},
		},
		{
			name:     "PC_SKILLS_DIR",
			envKey:   "PC_SKILLS_DIR",
			envValue: "/env/skills",
			checkFn: func(cfg *Config, v string) {
				if cfg.SkillsDir != v {
					t.Errorf("SkillsDir = %v, want %v", cfg.SkillsDir, v)
				}
			},
		},
		{
			name:     "PC_PLUGINS_DIR",
			envKey:   "PC_PLUGINS_DIR",
			envValue: "/env/plugins",
			checkFn: func(cfg *Config, v string) {
				if cfg.PluginsDir != v {
					t.Errorf("PluginsDir = %v, want %v", cfg.PluginsDir, v)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			t.Setenv(tt.envKey, tt.envValue)

			// Load config (should apply env override)
			cfg, err := Load(filepath.Join(t.TempDir(), "config.yml"))
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Check override was applied
			tt.checkFn(cfg, tt.envValue)
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot get home directory")
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "expand ~ at start",
			path: "~/test",
			want: filepath.Join(home, "test"),
		},
		{
			name: "expand ~ in middle",
			path: "/path/to/~/test",
			want: "/path/to/~/test", // Should not expand
		},
		{
			name: "absolute path",
			path: "/absolute/path",
			want: "/absolute/path",
		},
		{
			name: "relative path",
			path: "relative/path",
			want: "relative/path",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
		{
			name: "just ~",
			path: "~",
			want: home,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if got != tt.want {
				t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetters(t *testing.T) {
	cfg := &Config{
		Agent: AgentConfig{
			Name:        "TestAgent",
			Profession:  "TestProfession",
			Personality: []string{"Friendly"},
		},
		Workspace:  "/workspace",
		SkillsDir:  "/skills",
		PluginsDir: "/plugins",
		Log: LogConfig{
			Level:  DebugLevel,
			Output: "/log",
		},
	}

	tests := []struct {
		name string
		got  any
		want any
	}{
		{"GetAgentName", cfg.GetAgentName(), "TestAgent"},
		{"GetAgentProfession", cfg.GetAgentProfession(), "TestProfession"},
		// Handle slice separately
		{"GetAgentPersonality", nil, nil},
		{"GetWorkspace", cfg.GetWorkspace(), "/workspace"},
		{"GetSkillsDir", cfg.GetSkillsDir(), "/skills"},
		{"GetPluginsDir", cfg.GetPluginsDir(), "/plugins"},
		{"GetLogLevel", cfg.GetLogLevel(), DebugLevel},
		{"GetLogOutput", cfg.GetLogOutput(), "/log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "GetAgentPersonality" {
				// Handle slice comparison separately
				got := cfg.GetAgentPersonality()
				want := []string{"Friendly"}
				if len(got) != len(want) {
					t.Errorf("%s() length = %v, want %v", tt.name, len(got), len(want))
				} else if len(got) > 0 && got[0] != want[0] {
					t.Errorf("%s() = %v, want %v", tt.name, got, want)
				}
			} else {
				if tt.got != tt.want {
					t.Errorf("%s() = %v, want %v", tt.name, tt.got, tt.want)
				}
			}
		})
	}
}

func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "nested", "config.yml")

	cfg := DefaultConfig()

	// Save should create all parent directories
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save() did not create the config file")
	}
}

func TestPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "partial.yml")

	// Write partial config (only agent name)
	yamlContent := `
agent:
  name: PartialAgent
`
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	if err != nil {
		t.Fatalf("failed to write partial config: %v", err)
	}

	// Load config
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify partial values are loaded
	if cfg.Agent.Name != "PartialAgent" {
		t.Errorf("Agent.Name = %v, want PartialAgent", cfg.Agent.Name)
	}

	// Verify other values use defaults
	if cfg.Agent.Profession != "通用助手" {
		t.Errorf("Agent.Profession = %v, want 通用助手", cfg.Agent.Profession)
	}
}

func TestAgentPersonality(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "personality.yml")

	tests := []struct {
		name        string
		personality []string
	}{
		{"single trait", []string{"Friendly"}},
		{"multiple traits", []string{"Friendly", "Professional", "Helpful"}},
		{"empty traits", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config with personality
			cfg := &Config{
				Agent: AgentConfig{
					Name:        "TestAgent",
					Personality: tt.personality,
				},
				Workspace:  "~/workspace",
				SkillsDir:  "~/.pc/skills",
				PluginsDir: "~/.pc/plugins",
				Log: LogConfig{
					Level:  InfoLevel,
					Output: "stdout",
				},
			}

			// Save and load
			err := cfg.Save(configPath)
			if err != nil {
				t.Fatalf("Save() error = %v", err)
			}

			loaded, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Verify personality
			if len(loaded.Agent.Personality) != len(tt.personality) {
				t.Errorf("Personality length = %v, want %v", len(loaded.Agent.Personality), len(tt.personality))
			}
		})
	}
}

func TestYAMLContent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "content.yml")

	// Create a config and save it
	cfg := DefaultConfig()
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Read the YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	// Verify it's valid YAML (should contain expected fields)
	content := string(data)
	expectedFields := []string{"agent:", "name:", "workspace:", "log:", "level:"}
	for _, field := range expectedFields {
		if !contains(content, field) {
			t.Errorf("YAML does not contain expected field: %s\nContent:\n%s", field, content)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func TestConfigStructTags(t *testing.T) {
	// Verify that Config struct can be marshaled to YAML
	cfg := DefaultConfig()

	// This test just ensures the struct can be marshaled
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	// Verify output is not empty
	if len(data) == 0 {
		t.Error("Marshal produced empty output")
	}
}

// TestSystemPromptFile tests the system_prompt.file configuration (M9-003)
func TestSystemPromptFile(t *testing.T) {
	tests := []struct {
		name         string
		configFile   string
		wantDefault  string
		wantCustom   string
		envValue     string
		wantEnvValue string
	}{
		{
			name:        "default value",
			configFile:  "",
			wantDefault: "~/.pc/agents.md",
		},
		{
			name:       "custom path in config",
			configFile: "system_prompt:\n  file: /custom/path/agents.md\n",
			wantCustom: "/custom/path/agents.md",
		},
		{
			name:         "env override",
			configFile:   "",
			envValue:     "/env/path/agents.md",
			wantEnvValue: "/env/path/agents.md",
		},
		{
			name:       "relative path",
			configFile: "system_prompt:\n  file: ./agents.md\n",
			wantCustom: "./agents.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yml")

			// Set env if needed
			if tt.envValue != "" {
				t.Setenv("PC_SYSTEM_PROMPT_FILE", tt.envValue)
			}

			// Write config file if content provided
			if tt.configFile != "" {
				err := os.WriteFile(configPath, []byte(tt.configFile), 0644)
				if err != nil {
					t.Fatalf("failed to write config: %v", err)
				}
			}

			// Load config
			cfg, err := Load(configPath)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			// Check expected value
			var want string
			switch {
			case tt.wantEnvValue != "":
				want = tt.wantEnvValue
			case tt.wantCustom != "":
				want = tt.wantCustom
			default:
				want = tt.wantDefault
			}

			// Path should be expanded (if starts with ~)
			if strings.HasPrefix(want, "~") {
				home, _ := os.UserHomeDir()
				want = expandPath(want)
				_ = home
			}

			got := cfg.GetSystemPromptFile()
			if got != want {
				t.Errorf("GetSystemPromptFile() = %v, want %v", got, want)
			}
		})
	}
}

// TestSystemPromptFileSaveLoad tests saving and loading system_prompt.file
func TestSystemPromptFileSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")

	// Create config with custom system_prompt.file
	cfg := &Config{
		Agent: AgentConfig{
			Name: "TestAgent",
		},
		Workspace:  "~/workspace",
		SkillsDir:  "~/.pc/skills",
		PluginsDir: "~/.pc/plugins",
		SystemPrompt: SystemPromptConfig{
			File: "/custom/agents.md",
		},
		Log: LogConfig{
			Level:  InfoLevel,
			Output: "stdout",
		},
	}

	// Save
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify
	if loaded.SystemPrompt.File != cfg.SystemPrompt.File {
		t.Errorf("SystemPrompt.File = %v, want %v", loaded.SystemPrompt.File, cfg.SystemPrompt.File)
	}
}

// TestSystemPromptFileGetter tests the getter method
func TestSystemPromptFileGetter(t *testing.T) {
	cfg := &Config{SystemPrompt: SystemPromptConfig{File: "/test/agents.md"}}
	if got := cfg.GetSystemPromptFile(); got != "/test/agents.md" {
		t.Errorf("GetSystemPromptFile() = %v, want /test/agents.md", got)
	}
}

// TestDefaultSystemPromptFile tests that default config has correct default
func TestDefaultSystemPromptFile(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.SystemPrompt.File != "~/.pc/agents.md" {
		t.Errorf("Default SystemPrompt.File = %v, want ~/.pc/agents.md", cfg.SystemPrompt.File)
	}
}
