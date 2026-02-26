package task

import (
	"testing"
)

func TestNewDecisionEngine(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new decision engine",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			if d == nil {
				t.Error("NewDecisionEngine() returned nil")
			}
		})
	}
}

func TestDecisionEngine_Decide(t *testing.T) {
	tests := []struct {
		name        string
		context     Context
		wantAction  ActionType
		wantTool    string
		wantErr     bool
	}{
		{
			name: "decide to respond directly",
			context: Context{
				UserMessage: "Hello",
				History:     []Message{},
			},
			wantAction: ActionRespond,
			wantTool:   "",
			wantErr:    false,
		},
		{
			name: "decide to use tool",
			context: Context{
				UserMessage: "What is the weather today?",
				History:     []Message{},
			},
			wantAction: ActionUseTool,
			wantTool:   "weather",
			wantErr:    false,
		},
		{
			name: "decide to use skill",
			context: Context{
				UserMessage: "Deploy my application",
				History:     []Message{},
			},
			wantAction: ActionUseSkill,
			wantTool:   "deploy",
			wantErr:    false,
		},
		{
			name: "decide to wait for clarification",
			context: Context{
				UserMessage: "",
				History:     []Message{},
			},
			wantAction: ActionWait,
			wantTool:   "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			decision, err := d.Decide(tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decide() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if decision.Action != tt.wantAction {
					t.Errorf("Decide() Action = %v, want %v", decision.Action, tt.wantAction)
				}
			}
		})
	}
}

func TestDecisionEngine_NeedsTool(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		wantNeed bool
		wantTool string
	}{
		{
			name:     "needs weather tool",
			message:  "What is the weather like?",
			wantNeed: true,
			wantTool: "weather",
		},
		{
			name:     "needs search tool",
			message:  "Search for Go programming tutorials",
			wantNeed: true,
			wantTool: "search",
		},
		{
			name:     "needs file tool",
			message:  "Read the config file",
			wantNeed: true,
			wantTool: "file",
		},
		{
			name:     "does not need tool",
			message:  "Hello, how are you?",
			wantNeed: false,
			wantTool: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			need, tool := d.NeedsTool(tt.message)
			if need != tt.wantNeed {
				t.Errorf("NeedsTool() need = %v, want %v", need, tt.wantNeed)
			}
			if need && tool != tt.wantTool {
				t.Errorf("NeedsTool() tool = %v, want %v", tool, tt.wantTool)
			}
		})
	}
}

func TestDecisionEngine_SelectSkill(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantSkill string
		wantFound bool
	}{
		{
			name:      "select deploy skill",
			message:   "Deploy the application to production",
			wantSkill: "deploy",
			wantFound: true,
		},
		{
			name:      "select test skill",
			message:   "Run all tests",
			wantSkill: "test",
			wantFound: true,
		},
		{
			name:      "select build skill",
			message:   "Build the project",
			wantSkill: "build",
			wantFound: true,
		},
		{
			name:      "no skill found",
			message:   "What is your name?",
			wantSkill: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			skill, found := d.SelectSkill(tt.message)
			if found != tt.wantFound {
				t.Errorf("SelectSkill() found = %v, want %v", found, tt.wantFound)
			}
			if found && skill != tt.wantSkill {
				t.Errorf("SelectSkill() skill = %v, want %v", skill, tt.wantSkill)
			}
		})
	}
}

func TestDecisionEngine_CalculateConfidence(t *testing.T) {
	tests := []struct {
		name            string
		context         Context
		decision        Decision
		wantConfidence  float64
		wantThreshold   float64
	}{
		{
			name: "high confidence for direct response",
			context: Context{
				UserMessage: "Hello",
				History:     []Message{},
			},
			decision: Decision{
				Action: ActionRespond,
			},
			wantConfidence: 0.9,
			wantThreshold:  0.7,
		},
		{
			name: "low confidence for empty message",
			context: Context{
				UserMessage: "",
				History:     []Message{},
			},
			decision: Decision{
				Action: ActionWait,
			},
			wantConfidence: 0.3,
			wantThreshold:  0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			confidence := d.CalculateConfidence(tt.context, tt.decision)
			if confidence < 0 || confidence > 1 {
				t.Errorf("CalculateConfidence() = %v, want value between 0 and 1", confidence)
			}
		})
	}
}
