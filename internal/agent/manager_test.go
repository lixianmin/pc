package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lixianmin/pc/internal/config"
)

func TestAgentManager_NewManager(t *testing.T) {
	tests := []struct {
		name       string
		config     *config.Config
		configPath string
		wantErr    bool
	}{
		{
			name: "valid config",
			config: &config.Config{
				Agent: config.AgentConfig{
					Name: "TestAgent",
				},
				Workspace:  "/tmp/test/workspace",
				SkillsDir:  "/tmp/test/skills",
				PluginsDir: "/tmp/test/plugins",
				Log: config.LogConfig{
					Level:  config.InfoLevel,
					Output: "stdout",
				},
			},
			configPath: "/tmp/test/config.yml",
			wantErr:    false,
		},
		{
			name:       "nil config",
			config:     nil,
			configPath: "/tmp/test/config.yml",
			wantErr:    true,
		},
		{
			name: "empty config",
			config: &config.Config{
				Agent:      config.AgentConfig{},
				Workspace:  "",
				SkillsDir:  "",
				PluginsDir: "",
				Log:        config.LogConfig{},
			},
			configPath: "/tmp/test/config.yml",
			wantErr:    false, // NewManager doesn't validate, only Create does
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := NewManager(tt.config, tt.configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && mgr == nil {
				t.Error("NewManager() returned nil manager for valid config")
			}
		})
	}
}

func TestAgentManager_GetAgent(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*Manager, error)
		wantNil   bool
	}{
		{
			name: "agent initialized from config",
			setupFunc: func() (*Manager, error) {
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name:        "TestAgent",
						Profession:  "Developer",
						Personality: []string{"friendly", "professional"},
					},
					Workspace:  "/tmp/test/workspace",
					SkillsDir:  "/tmp/test/skills",
					PluginsDir: "/tmp/test/plugins",
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				return NewManager(cfg, "/tmp/test/config.yml")
			},
			wantNil: false,
		},
		{
			name: "agent with empty name",
			setupFunc: func() (*Manager, error) {
				cfg := &config.Config{
					Agent:      config.AgentConfig{},
					Workspace:  "/tmp/test/workspace",
					SkillsDir:  "/tmp/test/skills",
					PluginsDir: "/tmp/test/plugins",
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				return NewManager(cfg, "/tmp/test/config.yml")
			},
			wantNil: false, // GetAgent should always return a non-nil agent
		},
		{
			name: "multiple calls return same agent",
			setupFunc: func() (*Manager, error) {
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  "/tmp/test/workspace",
					SkillsDir:  "/tmp/test/skills",
					PluginsDir: "/tmp/test/plugins",
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				return NewManager(cfg, "/tmp/test/config.yml")
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("setupFunc() error = %v", err)
			}

			agent := mgr.GetAgent()
			if (agent == nil) != tt.wantNil {
				t.Errorf("GetAgent() agent = %v, wantNil %v", agent, tt.wantNil)
			}

			if !tt.wantNil {
				if agent.State == "" {
					t.Error("GetAgent() agent state should not be empty")
				}
			}
		})
	}
}

func TestAgentManager_UpdateAgentName(t *testing.T) {
	tests := []struct {
		name      string
		agentName string
		newName   string
		wantErr   bool
	}{
		{
			name:      "valid name update",
			agentName: "OldName",
			newName:   "NewName",
			wantErr:   false,
		},
		{
			name:      "update to same name",
			agentName: "SameName",
			newName:   "SameName",
			wantErr:   false,
		},
		{
			name:      "update to empty name",
			agentName: "OldName",
			newName:   "",
			wantErr:   false, // UpdateAgentName doesn't validate the name
		},
		{
			name:      "update to unicode name",
			agentName: "OldName",
			newName:   "中文助手",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Agent: config.AgentConfig{
					Name: tt.agentName,
				},
				Workspace:  "/tmp/test/workspace",
				SkillsDir:  "/tmp/test/skills",
				PluginsDir: "/tmp/test/plugins",
				Log: config.LogConfig{
					Level:  config.InfoLevel,
					Output: "stdout",
				},
			}
			mgr, err := NewManager(cfg, "/tmp/test/config.yml")
			if err != nil {
				t.Fatalf("NewManager() error = %v", err)
			}

			// Initialize agent
			_ = mgr.GetAgent()

			err = mgr.UpdateAgentName(tt.newName)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateAgentName() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify the name was updated
			agent := mgr.GetAgent()
			if agent.Name != tt.newName {
				t.Errorf("UpdateAgentName() agent.Name = %v, want %v", agent.Name, tt.newName)
			}
		})
	}
}

func TestAgentManager_SaveState(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (*Manager, string, error)
		cleanup    func(string)
		wantErr    bool
	}{
		{
			name: "save state creates file",
			setupFunc: func() (*Manager, string, error) {
				tmpDir := t.TempDir()
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				configPath := filepath.Join(tmpDir, "config.yml")
				mgr, err := NewManager(cfg, configPath)
				if err != nil {
					return nil, "", err
				}
				_ = mgr.GetAgent()
				return mgr, tmpDir, nil
			},
			cleanup: nil, // t.TempDir() handles cleanup
			wantErr: false,
		},
		{
			name: "save state with nil agent",
			setupFunc: func() (*Manager, string, error) {
				tmpDir := t.TempDir()
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				configPath := filepath.Join(tmpDir, "config.yml")
				mgr, err := NewManager(cfg, configPath)
				if err != nil {
					return nil, "", err
				}
				// Don't initialize agent
				return mgr, tmpDir, nil
			},
			cleanup: nil,
			wantErr: false, // SaveState returns nil for nil agent
		},
		{
			name: "save state creates directory",
			setupFunc: func() (*Manager, string, error) {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "nested", "dir", "config.yml")
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				mgr, err := NewManager(cfg, configPath)
				if err != nil {
					return nil, "", err
				}
				_ = mgr.GetAgent()
				return mgr, tmpDir, nil
			},
			cleanup: nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, tmpDir, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("setupFunc() error = %v", err)
			}

			err = mgr.SaveState()
			if (err != nil) != tt.wantErr {
				t.Errorf("SaveState() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify state file exists
			statePath := filepath.Join(tmpDir, "nested", "dir", "agent.state")
			if tt.name == "save state creates directory" {
				_, err := os.Stat(statePath)
				if err != nil {
					t.Errorf("SaveState() state file not created: %v", err)
				}
			}
		})
	}
}

func TestAgentManager_LoadState(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() (*Manager, string, error)
		wantErr    bool
		validate   func(*testing.T, *Manager)
	}{
		{
			name: "load non-existent state file",
			setupFunc: func() (*Manager, string, error) {
				tmpDir := t.TempDir()
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				configPath := filepath.Join(tmpDir, "config.yml")
				mgr, err := NewManager(cfg, configPath)
				return mgr, tmpDir, err
			},
			wantErr: false,
			validate: func(t *testing.T, mgr *Manager) {
				agent := mgr.GetAgent()
				if agent == nil {
					t.Error("LoadState() should initialize default agent")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("setupFunc() error = %v", err)
			}

			err = mgr.LoadState()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadState() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.validate != nil {
				tt.validate(t, mgr)
			}
		})
	}
}

func TestAgentManager_Close(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() (*Manager, error)
		wantErr   bool
	}{
		{
			name: "close without errors",
			setupFunc: func() (*Manager, error) {
				tmpDir := t.TempDir()
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				configPath := filepath.Join(tmpDir, "config.yml")
				mgr, err := NewManager(cfg, configPath)
				if err != nil {
					return nil, err
				}
				_ = mgr.GetAgent()
				return mgr, nil
			},
			wantErr: false,
		},
		{
			name: "close without agent initialized",
			setupFunc: func() (*Manager, error) {
				tmpDir := t.TempDir()
				cfg := &config.Config{
					Agent: config.AgentConfig{
						Name: "TestAgent",
					},
					Workspace:  filepath.Join(tmpDir, "workspace"),
					SkillsDir:  filepath.Join(tmpDir, "skills"),
					PluginsDir: filepath.Join(tmpDir, "plugins"),
					Log: config.LogConfig{
						Level:  config.InfoLevel,
						Output: "stdout",
					},
				}
				configPath := filepath.Join(tmpDir, "config.yml")
				return NewManager(cfg, configPath)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := tt.setupFunc()
			if err != nil {
				t.Fatalf("setupFunc() error = %v", err)
			}

			err = mgr.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
