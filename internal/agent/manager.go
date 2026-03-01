package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lixianmin/pc/internal/config"
	"github.com/lixianmin/pc/internal/logger"
	"github.com/lixianmin/pc/internal/plugin"
	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/pkg/types"
)

// Manager manages agent lifecycle and state.
type Manager struct {
	config         *config.Config
	agent          *types.Agent
	configPath     string
	statePath      string
	agentsMdConfig *AgentsMdConfig
	skillManager   skill.ISkillManager
	pluginManager  *plugin.PluginManager
	mu             sync.RWMutex
}

// NewManager creates a new agent manager.
func NewManager(cfg *config.Config, configPath string) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	return &Manager{
		config:     cfg,
		configPath: configPath,
		statePath:  filepath.Join(filepath.Dir(configPath), "agent.state"),
		mu:         sync.RWMutex{},
	}, nil
}

// LoadConfig loads configuration and initializes agent.
func (my *Manager) LoadConfig() error {
	// Load configuration
	cfg, err := config.Load(my.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	my.config = cfg

	return nil
}

// GetAgent returns the current agent.
func (my *Manager) GetAgent() *types.Agent {
	my.mu.RLock()
	defer my.mu.RUnlock()

	if my.agent == nil {
		// Initialize agent if not already initialized
		my.agent = &types.Agent{
			Name:        my.config.GetAgentName(),
			Profession:  my.config.GetAgentProfession(),
			Personality: my.config.GetAgentPersonality(),
			State:       types.AgentStateIdle,
		}
	}

	return my.agent
}

// SaveState saves the current agent state.
func (my *Manager) SaveState() error {
	my.mu.RLock()
	defer my.mu.RUnlock()

	if my.agent == nil {
		return nil
	}

	state, err := my.agentToState(my.agent)
	if err != nil {
		return fmt.Errorf("failed to serialize agent state: %w", err)
	}

	// Create state directory if it doesn't exist
	stateDir := filepath.Dir(my.statePath)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write state to file
	if err := os.WriteFile(my.statePath, state, 0644); err != nil {
		return fmt.Errorf("failed to write agent state: %w", err)
	}

	if l := logger.Get(); l != nil {
		l.Info("Agent state saved")
	}
	return nil
}

// LoadState restores the agent state from file.
func (my *Manager) LoadState() error {
	// Read state file
	data, err := os.ReadFile(my.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// State file doesn't exist, initialize default agent
			if l := logger.Get(); l != nil {
				l.Info("No agent state file found, using default state")
			}
			return nil
		}
		return fmt.Errorf("failed to read agent state: %w", err)
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	// Deserialize agent from state
	agent, err := my.stateToAgent(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize agent state: %w", err)
	}

	my.agent = agent
	if l := logger.Get(); l != nil {
		l.Info("Agent state restored")
	}

	return nil
}

// agentToState converts agent to serializable state format.
func (my *Manager) agentToState(agent *types.Agent) ([]byte, error) {
	// Use simple format: name, state
	// For now, just serialize to JSON manually
	state := map[string]any{
		"name":  agent.Name,
		"state": string(agent.State),
	}

	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent state: %w", err)
	}

	return data, nil
}

// stateToAgent converts serialized state back to agent.
func (my *Manager) stateToAgent(data []byte) (*types.Agent, error) {
	var state map[string]any
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent state: %w", err)
	}

	agent := &types.Agent{
		Name:  state["name"].(string),
		State: types.AgentState(state["state"].(string)),
	}

	return agent, nil
}

// UpdateAgentName updates the agent name.
func (my *Manager) UpdateAgentName(name string) error {
	my.mu.Lock()
	defer my.mu.Unlock()

	if my.agent == nil {
		return fmt.Errorf("agent not initialized")
	}

	my.agent.Name = name
	return nil
}

// LoadAgentsMd loads the agents.md file from the given path.
// If the file doesn't exist, it returns nil without error.
func (my *Manager) LoadAgentsMd(path string) error {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if l := logger.Get(); l != nil {
			l.Info("No agents.md file found at:", path)
		}
		return nil
	}

	cfg, err := LoadAgentsMd(path)
	if err != nil {
		return fmt.Errorf("failed to load agents.md: %w", err)
	}

	my.mu.Lock()
	defer my.mu.Unlock()

	my.agentsMdConfig = cfg

	if l := logger.Get(); l != nil {
		l.Info("Loaded agents.md:", cfg.Name)
	}

	return nil
}

// GetAgentsMdConfig returns the loaded agents.md config.
// Returns nil if agents.md was not loaded.
func (my *Manager) GetAgentsMdConfig() *AgentsMdConfig {
	my.mu.RLock()
	defer my.mu.RUnlock()

	return my.agentsMdConfig
}

// GetSystemPromptBase returns the base system prompt from agents.md.
// If agents.md was not loaded, returns a default system prompt.
func (my *Manager) GetSystemPromptBase() string {
	my.mu.RLock()
	defer my.mu.RUnlock()

	if my.agentsMdConfig != nil {
		return my.agentsMdConfig.ToSystemPrompt()
	}

	// Return default system prompt based on config
	if my.config != nil {
		return fmt.Sprintf("你是 %s。", my.config.GetAgentName())
	}

	return "你是一个 AI 助手。"
}

// Close saves the agent state and cleans up resources.
func (my *Manager) Close() error {
	if err := my.SaveState(); err != nil {
		return err
	}

	// Note: we don't close config or logger here as they are managed externally

	return nil
}

// SetSkillManager sets the skill manager for system prompt building.
func (my *Manager) SetSkillManager(sm skill.ISkillManager) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.skillManager = sm
}

// SetPluginManager sets the plugin manager for system prompt building.
func (my *Manager) SetPluginManager(pm *plugin.PluginManager) {
	my.mu.Lock()
	defer my.mu.Unlock()
	my.pluginManager = pm
}

// GetSystemPromptBuilder creates and returns a SystemPromptBuilder with current state.
func (my *Manager) GetSystemPromptBuilder() *SystemPromptBuilder {
	my.mu.RLock()
	defer my.mu.RUnlock()

	builder := NewSystemPromptBuilder()

	// Set base prompt from agents.md or default
	basePrompt := my.getSystemPromptBaseLocked()
	builder.SetBasePrompt(basePrompt)

	// Set skills if skill manager is available
	if my.skillManager != nil {
		skills := my.skillManager.ListSkills()
		builder.SetSkills(skills)
	}

	// Set tools if plugin manager is available
	if my.pluginManager != nil {
		plugins := my.pluginManager.ListPlugins()
		tools := make([]ToolInfo, 0, len(plugins))
		for _, p := range plugins {
			if p.Enabled {
				tools = append(tools, ToolInfoFromPlugin(p))
			}
		}
		builder.SetTools(tools)
	}

	return builder
}

// BuildSystemPrompt builds and returns the complete system prompt.
func (my *Manager) BuildSystemPrompt() string {
	return my.GetSystemPromptBuilder().Build()
}

// getSystemPromptBaseLocked returns the base system prompt (must be called with lock held).
func (my *Manager) getSystemPromptBaseLocked() string {
	// Try to read from agents.md file first (using simplified method)
	if my.config != nil {
		agentsMdPath := my.config.GetSystemPromptFile()
		content, err := ReadAgentsMdContent(agentsMdPath)
		if err == nil && content != "" {
			return content
		}
	}

	// Fall back to parsed agents.md config
	if my.agentsMdConfig != nil {
		return my.agentsMdConfig.ToSystemPrompt()
	}

	// Return default system prompt based on config
	if my.config != nil {
		return fmt.Sprintf("你是 %s。", my.config.GetAgentName())
	}

	return "你是一个 AI 助手。"
}
