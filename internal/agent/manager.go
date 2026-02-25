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
		statePath: filepath.Join(filepath.Dir(configPath), "agent.state"),
		mu:         sync.RWMutex{},
	}, nil
}

// LoadConfig loads configuration and initializes agent.
func (m *Manager) LoadConfig() error {
	// Load configuration
	cfg, err := config.Load(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg

	return nil
}

// GetAgent returns the current agent.
func (m *Manager) GetAgent() *types.Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.agent == nil {
		// Initialize agent if not already initialized
		m.agent = &types.Agent{
			Name:       m.config.GetAgentName(),
			Profession: m.config.GetAgentProfession(),
			Personality: m.config.GetAgentPersonality(),
			State:      types.AgentStateIdle,
		}
	}

	return m.agent
}

// SaveState saves the current agent state.
func (m *Manager) SaveState() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.agent == nil {
		return nil
	}

	state, err := m.agentToState(m.agent)
	if err != nil {
		return fmt.Errorf("failed to serialize agent state: %w", err)
	}

	// Create state directory if it doesn't exist
	stateDir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Write state to file
	if err := os.WriteFile(m.statePath, state, 0644); err != nil {
		return fmt.Errorf("failed to write agent state: %w", err)
	}

	if l := logger.Get(); l != nil {
		l.Info("Agent state saved")
	}
	return nil
}

// LoadState restores the agent state from file.
func (m *Manager) LoadState() error {
	// Read state file
	data, err := os.ReadFile(m.statePath)
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

	m.mu.Lock()
	defer m.mu.Unlock()

	// Deserialize agent from state
	agent, err := m.stateToAgent(data)
	if err != nil {
		return fmt.Errorf("failed to deserialize agent state: %w", err)
	}

	m.agent = agent
	if l := logger.Get(); l != nil {
		l.Info("Agent state restored")
	}

	return nil
}

// agentToState converts agent to serializable state format.
func (m *Manager) agentToState(agent *types.Agent) ([]byte, error) {
	// Use simple format: name, state
	// For now, just serialize to JSON manually
	state := map[string]any{
		"name":     agent.Name,
		"state":    string(agent.State),
	}

	data, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal agent state: %w", err)
	}

	return data, nil
}

// stateToAgent converts serialized state back to agent.
func (m *Manager) stateToAgent(data []byte) (*types.Agent, error) {
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
func (m *Manager) UpdateAgentName(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.agent == nil {
		return fmt.Errorf("agent not initialized")
	}

	m.agent.Name = name
	return nil
}

// Close saves the agent state and cleans up resources.
func (m *Manager) Close() error {
	if err := m.SaveState(); err != nil {
		return err
	}

	// Note: we don't close config or logger here as they are managed externally

	return nil
}
