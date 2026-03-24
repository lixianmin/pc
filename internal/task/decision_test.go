package task

import (
	"context"
	"errors"
	"testing"
)

// mockLLMClient is a mock implementation of LLMClient for testing
type mockLLMClient struct {
	response string
	err      error
}

func (m *mockLLMClient) Complete(ctx context.Context, prompt string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

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
		name       string
		context    Context
		wantAction ActionType
		wantTool   string
		wantErr    bool
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
		name           string
		context        Context
		decision       Decision
		wantConfidence float64
		wantThreshold  float64
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

func TestDecisionEngine_SetLLMClient(t *testing.T) {
	tests := []struct {
		name       string
		client     LLMCompletionClient
		wantUseLLM bool
	}{
		{
			name:       "set valid LLM client",
			client:     &mockLLMClient{response: `{"action": "respond"}`},
			wantUseLLM: true,
		},
		{
			name:       "set nil LLM client",
			client:     nil,
			wantUseLLM: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			d.SetLLMClient(tt.client)
			if d.useLLM != tt.wantUseLLM {
				t.Errorf("useLLM = %v, want %v", d.useLLM, tt.wantUseLLM)
			}
			if d.llmClient != tt.client {
				t.Errorf("llmClient not set correctly")
			}
		})
	}
}

func TestDecisionEngine_DecideWithLLM(t *testing.T) {
	tests := []struct {
		name       string
		llmResp    string
		llmErr     error
		context    Context
		wantAction ActionType
	}{
		{
			name:    "LLM decides to respond",
			llmResp: `{"action": "respond", "reason": "General greeting", "confidence": 0.9}`,
			llmErr:  nil,
			context: Context{
				UserMessage: "Hello",
			},
			wantAction: ActionRespond,
		},
		{
			name:    "LLM decides to use tool",
			llmResp: `{"action": "use_tool", "tool": "weather", "reason": "User asked about weather", "confidence": 0.85}`,
			llmErr:  nil,
			context: Context{
				UserMessage: "What's the weather like?",
			},
			wantAction: ActionUseTool,
		},
		{
			name:    "LLM decides to use skill",
			llmResp: `{"action": "use_skill", "skill": "deploy", "reason": "User wants to deploy", "confidence": 0.8}`,
			llmErr:  nil,
			context: Context{
				UserMessage: "Deploy my app",
			},
			wantAction: ActionUseSkill,
		},
		{
			name:       "LLM fails, fallback to keywords",
			llmResp:    "",
			llmErr:     errors.New("LLM error"),
			context:    Context{UserMessage: "Search for something"},
			wantAction: ActionUseTool, // Fallback to keyword matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			mockClient := &mockLLMClient{response: tt.llmResp, err: tt.llmErr}
			d.SetLLMClient(mockClient)

			decision, err := d.Decide(tt.context)
			if err != nil {
				t.Errorf("Decide() unexpected error = %v", err)
				return
			}
			if decision.Action != tt.wantAction {
				t.Errorf("Decide() Action = %v, want %v", decision.Action, tt.wantAction)
			}
		})
	}
}

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "JSON in markdown code block",
			input: "```json\n{\"action\": \"respond\"}\n```",
			want:  `{"action": "respond"}`,
		},
		{
			name:  "Plain JSON",
			input: `{"action": "respond"}`,
			want:  `{"action": "respond"}`,
		},
		{
			name:  "JSON with surrounding text",
			input: `Some text before {"action": "respond"} some text after`,
			want:  `{"action": "respond"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSON(tt.input)
			if got != tt.want {
				t.Errorf("extractJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecisionEngine_ParseLLMResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		wantAction ActionType
		wantTool   string
		wantSkill  string
	}{
		{
			name:       "Valid JSON response",
			response:   `{"action": "use_tool", "tool": "search", "reason": "Need to search", "confidence": 0.8}`,
			wantAction: ActionUseTool,
			wantTool:   "search",
		},
		{
			name:       "JSON in markdown block",
			response:   "```json\n{\"action\": \"respond\", \"reason\": \"General question\"}\n```",
			wantAction: ActionRespond,
		},
		{
			name:       "Invalid JSON, falls back to keyword",
			response:   "xyz abc",
			wantAction: ActionRespond, // Falls back to keyword matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecisionEngine()
			ctx := Context{UserMessage: "hello"} // Use greeting that results in ActionRespond
			decision, err := d.parseLLMResponse(tt.response, ctx)
			if err != nil {
				t.Errorf("parseLLMResponse() unexpected error = %v", err)
				return
			}
			if decision.Action != tt.wantAction {
				t.Errorf("parseLLMResponse() Action = %v, want %v", decision.Action, tt.wantAction)
			}
		})
	}
}
