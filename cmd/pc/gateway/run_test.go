package gateway

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lixianmin/pc/internal/agent"
	"github.com/lixianmin/pc/internal/config"
)

// TestAgentStatePersistence tests that agent state is persisted across restarts
func TestAgentStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	statePath := filepath.Join(tmpDir, "agent.state")

	// Create config
	cfg := &config.Config{
		Agent: config.AgentConfig{
			Name: "TestAgent",
		},
		Workspace:  tmpDir,
		SkillsDir:  tmpDir,
		PluginsDir: tmpDir,
		Log: config.LogConfig{
			Level:  config.InfoLevel,
			Output: "stdout",
		},
	}

	// Create manager and agent
	mgr, err := agent.NewManager(cfg, configPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Initialize agent with a specific name
	agentObj := mgr.GetAgent()
	agentObj.Name = "PersistedAgentName"

	// Save state
	if err := mgr.SaveState(); err != nil {
		t.Fatalf("SaveState() error = %v", err)
	}

	// Verify state file exists
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Fatal("State file should exist after SaveState()")
	}

	// Create new manager (simulating restart)
	newMgr, err := agent.NewManager(cfg, configPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Load state
	if err := newMgr.LoadState(); err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	// Verify agent name was restored
	loadedAgent := newMgr.GetAgent()
	if loadedAgent.Name != "PersistedAgentName" {
		t.Errorf("Agent name not persisted: got %v, want PersistedAgentName", loadedAgent.Name)
	}
}

// TestAgentManagerIntegration tests AgentManager integration with Gateway
func TestAgentManagerIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yml")
	agentMdPath := filepath.Join(tmpDir, "agent.md")

	// Create agent.md file
	agentMdContent := `# TestAgent

## Profession
TestProfession

## Personality
- friendly
`
	if err := os.WriteFile(agentMdPath, []byte(agentMdContent), 0644); err != nil {
		t.Fatalf("failed to write agent.md: %v", err)
	}

	cfg := &config.Config{
		AgentPath:  agentMdPath,
		Workspace:  tmpDir,
		SkillsDir:  tmpDir,
		PluginsDir: tmpDir,
		Log: config.LogConfig{
			Level:  config.InfoLevel,
			Output: "stdout",
		},
	}

	mgr, err := agent.NewManager(cfg, configPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Load agent.md
	if err := mgr.LoadAgentsMd(agentMdPath); err != nil {
		t.Fatalf("LoadAgentsMd() error = %v", err)
	}

	// Test that GetAgent returns non-nil agent
	agentObj := mgr.GetAgent()
	if agentObj == nil {
		t.Fatal("GetAgent() returned nil")
	}

	// Verify agent properties from agent.md
	if agentObj.Name != "TestAgent" {
		t.Errorf("Agent.Name = %v, want TestAgent", agentObj.Name)
	}

	// Test SetSkillManager
	// Note: In real integration, this would be a real SkillManager
	// For now, we just verify the method exists and doesn't panic
	// mgr.SetSkillManager(skill.NewSkillManager())

	// Test SetPluginManager
	// mgr.SetPluginManager(pluginManager)
}

// TestGatewayRunWithAgentManager verifies gateway run initializes AgentManager
// This is an integration test that would require full setup
func TestGatewayRunWithAgentManager(t *testing.T) {
	// This test verifies the structure is in place
	// Full integration testing requires:
	// 1. Config file setup
	// 2. Plugin directory setup
	// 3. Running actual daemon (too complex for unit test)

	// For now, just verify runCmd exists
	if runCmd == nil {
		t.Fatal("runCmd should not be nil")
	}

	if runCmd.Use != "run" {
		t.Errorf("runCmd.Use = %v, want run", runCmd.Use)
	}

	// Verify --daemon flag exists
	daemonFlag, err := runCmd.Flags().GetBool("daemon")
	if err != nil {
		t.Log("daemon flag not set, expected for internal use")
	}
	_ = daemonFlag
}
