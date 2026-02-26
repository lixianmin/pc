package types

// AgentState represents the current state of an agent.
type AgentState string

const (
	// AgentStateIdle is the idle state (no active task).
	AgentStateIdle AgentState = "idle"
	// AgentStateWorking is the working state (executing a task).
	AgentStateWorking AgentState = "working"
	// AgentStateWaiting is the waiting state (awaiting response).
	AgentStateWaiting AgentState = "waiting"
	// AgentStateError is the error state (encountered an error).
	AgentStateError AgentState = "error"
)

// Agent represents an AI agent.
type Agent struct {
	Name        string     `json:"name"`        // Agent name
	Profession  string     `json:"profession"`  // Agent profession (optional)
	Personality []string   `json:"personality"` // Personality traits (optional)
	Skills      []Skill    `json:"skills"`      // Skills list
	Memory      *Memory    `json:"memory"`      // Agent memory
	State       AgentState `json:"state"`       // Current state
}

// Memory represents agent memory.
type Memory struct {
	ShortTerm []Message    `json:"short_term"` // Short-term memory (recent messages)
	LongTerm  []MemoryItem `json:"long_term"`  // Long-term memory (facts, preferences)
}

// Message represents a message in conversation.
type Message struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // Message content
	Ts      int64  `json:"ts"`      // Timestamp (Unix milliseconds)
}

// MemoryItem represents an item in long-term memory.
type MemoryItem struct {
	Key      string `json:"key"`        // Memory key
	Value    any    `json:"value"`      // Memory value
	UpdateAt int64  `json:"update_at"`  // Last update time (Unix milliseconds)
	ExpiresAt int64 `json:"expires_at"` // Expiration time (0 means never expires, Unix milliseconds)
}

// Skill represents an agent skill.
type Skill struct {
	Name        string            `json:"name"`        // Skill name
	Description string            `json:"description"` // Skill description
	Tools       []string          `json:"tools"`       // Required tools
	Metadata    map[string]string `json:"metadata"`    // Additional metadata
}

// TaskState represents the state of a task.
type TaskState string

const (
	// TaskStatePending is the pending state (not started).
	TaskStatePending TaskState = "PENDING"
	// TaskStateInProgress is the in-progress state (executing).
	TaskStateInProgress TaskState = "IN_PROGRESS"
	// TaskStateCompleted is the completed state (finished).
	TaskStateCompleted TaskState = "COMPLETED"
	// TaskStateFailed is the failed state (error occurred).
	TaskStateFailed TaskState = "FAILED"
)

// Task represents a task in the agent's task list.
type Task struct {
	ID         string     `json:"id"`          // Task ID
	Title      string     `json:"title"`       // Task title
	State      TaskState  `json:"state"`       // Task state
	Steps      []TaskStep `json:"steps"`       // Task steps
	CreateAt   int64      `json:"create_at"`   // Creation time (Unix milliseconds)
	UpdateAt   int64      `json:"update_at"`   // Last update time (Unix milliseconds)
	CompleteAt int64      `json:"complete_at"` // Completion time (Unix milliseconds)
}

// TaskStep represents a step in a task.
type TaskStep struct {
	Description string `json:"description"` // Step description
	Done        bool   `json:"done"`        // Whether step is completed
}

// HeartbeatTask represents a periodic (heartbeat) task.
type HeartbeatTask struct {
	Schedule    string `json:"schedule"`    // Cron schedule
	Description string `json:"description"` // Task description
	OnFailure   string `json:"on_failure"`  // Action on failure ("notify" or "ignore")
	Enabled     bool   `json:"enabled"`     // Whether task is enabled
}

// PluginMetadata represents plugin metadata.
type PluginMetadata struct {
	Name    string     `json:"name"`    // Plugin name
	Type    PluginType `json:"type"`    // Plugin type
	Enabled bool       `json:"enabled"` // Whether plugin is enabled
	Version string     `json:"version"` // Plugin version
	Entry   string     `json:"entry"`   // Entry point (executable path)
}
