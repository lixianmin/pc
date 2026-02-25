package skill

import (
	"testing"
)

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new executor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewExecutor()
			if got == nil {
				t.Error("NewExecutor() returned nil")
			}
		})
	}
}

func TestExecuteSkill(t *testing.T) {
	tests := []struct {
		name    string
		skill   *Skill
		context map[string]any
		wantErr bool
	}{
		{
			name: "execute simple skill",
			skill: &Skill{
				Name:        "test-skill",
				Description: "Test",
				Steps:       []string{"step1", "step2"},
			},
			context: map[string]any{},
			wantErr: false,
		},
		{
			name: "execute skill with context",
			skill: &Skill{
				Name:        "context-skill",
				Description: "Test",
				Steps:       []string{"step1"},
			},
			context: map[string]any{"key": "value"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor()
			result, err := e.ExecuteSkill(tt.skill, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteSkill() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("ExecuteSkill() returned nil result")
			}
		})
	}
}

func TestExecuteStep(t *testing.T) {
	tests := []struct {
		name    string
		step    ExecutionStep
		context map[string]any
		wantErr bool
	}{
		{
			name: "execute action step",
			step: ExecutionStep{
				Type:    "action",
				Command: "test command",
			},
			context: map[string]any{},
			wantErr: false,
		},
		{
			name: "execute condition step",
			step: ExecutionStep{
				Type:      "condition",
				Condition: "true",
			},
			context: map[string]any{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor()
			result, err := e.ExecuteStep(tt.step, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteStep() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("ExecuteStep() returned nil result")
			}
		})
	}
}

func TestCallTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		params   map[string]any
		wantErr  bool
	}{
		{
			name:     "call existing tool",
			toolName: "test-tool",
			params:   map[string]any{"param": "value"},
			wantErr:  false,
		},
		{
			name:     "call non-existent tool",
			toolName: "non-existent",
			params:   map[string]any{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewExecutor()
			result, err := e.CallTool(tt.toolName, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("CallTool() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && result == nil {
				t.Error("CallTool() returned nil result")
			}
		})
	}
}
