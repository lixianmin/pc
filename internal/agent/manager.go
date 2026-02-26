package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/lixianmin/pc/internal/config"
	"github.com/lixianmin/pc/internal/logger"
	"github.com/lixianmin/pc/pkg/types"
)

// Manager manages agent lifecycle and state.
type Manager struct {
	config     *config.Config
	agent      *types.Agent
	configPath string
	statePath  string
	mu         sync.RWMutex
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

// Close saves the agent state and cleans up resources.
func (my *Manager) Close() error {
	if err := my.SaveState(); err != nil {
		return err
	}

	// Note: we don't close config or logger here as they are managed externally

	return nil
}
