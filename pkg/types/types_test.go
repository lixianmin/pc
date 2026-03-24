package types

import (
	"encoding/json"
	"testing"
)

func TestPluginTypeFromString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want PluginType
	}{
		{"llm", "llm", PluginTypeLLM},
		{"search", "search", PluginTypeSearch},
		{"channel", "channel", PluginTypeChannel},
		{"tool", "tool", PluginTypeTool},
		{"unknown", "unknown", PluginTypeTool},
		{"", "", PluginTypeTool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PluginTypeFromString(tt.s)
			if got != tt.want {
				t.Errorf("PluginTypeFromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPluginTypeString(t *testing.T) {
	tests := []struct {
		name string
		p    PluginType
		want string
	}{
		{"LLM", PluginTypeLLM, "llm"},
		{"Search", PluginTypeSearch, "search"},
		{"Channel", PluginTypeChannel, "channel"},
		{"Tool", PluginTypeTool, "tool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPluginTypeIsValid(t *testing.T) {
	tests := []struct {
		name string
		p    PluginType
		want bool
	}{
		{"LLM", PluginTypeLLM, true},
		{"Search", PluginTypeSearch, true},
		{"Channel", PluginTypeChannel, true},
		{"Tool", PluginTypeTool, true},
		{"Invalid", PluginType("invalid"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAgentJSON(t *testing.T) {
	agent := Agent{
		Name:        "TestAgent",
		Profession:  "TestProfession",
		Personality: []string{"Friendly", "Helpful"},
		Skills: []Skill{
			{
				Name:        "test-skill",
				Description: "Test skill description",
				Tools:       []string{"tool1", "tool2"},
			},
		},
		Memory: &Memory{
			ShortTerm: []Message{
				{
					Role:    "user",
					Content: "Hello",
					Ts:      1234567890000,
				},
			},
		},
		State: AgentStateIdle,
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Agent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != agent.Name {
		t.Errorf("Decoded Name = %v, want %v", decoded.Name, agent.Name)
	}

	if decoded.State != agent.State {
		t.Errorf("Decoded State = %v, want %v", decoded.State, agent.State)
	}
}

func TestSkillJSON(t *testing.T) {
	skill := Skill{
		Name:        "test-skill",
		Description: "Test skill",
		Tools:       []string{"tool1", "tool2"},
		Metadata: map[string]string{
			"author": "test",
		},
	}

	data, err := json.Marshal(skill)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Skill
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != skill.Name {
		t.Errorf("Decoded Name = %v, want %v", decoded.Name, skill.Name)
	}

	if len(decoded.Tools) != len(skill.Tools) {
		t.Errorf("Decoded Tools length = %v, want %v", len(decoded.Tools), len(skill.Tools))
	}
}

func TestTaskJSON(t *testing.T) {
	task := Task{
		ID:         "task-001",
		Title:      "Test Task",
		State:      TaskStateInProgress,
		CreateAt:   1234567890000,
		UpdateAt:   1234567891000,
		CompleteAt: 0,
		Steps: []TaskStep{
			{Description: "Step 1", Done: true},
			{Description: "Step 2", Done: false},
		},
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Task
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("Decoded ID = %v, want %v", decoded.ID, task.ID)
	}

	if decoded.State != task.State {
		t.Errorf("Decoded State = %v, want %v", decoded.State, task.State)
	}

	if len(decoded.Steps) != len(task.Steps) {
		t.Errorf("Decoded Steps length = %v, want %v", len(decoded.Steps), len(task.Steps))
	}
}

func TestHeartbeatTaskJSON(t *testing.T) {
	ht := HeartbeatTask{
		Schedule:    "0 2 * * *",
		Description: "Daily backup",
		OnFailure:   "notify",
		Enabled:     true,
	}

	data, err := json.Marshal(ht)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded HeartbeatTask
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Schedule != ht.Schedule {
		t.Errorf("Decoded Schedule = %v, want %v", decoded.Schedule, ht.Schedule)
	}
}

func TestPluginMetadataJSON(t *testing.T) {
	pm := PluginMetadata{
		Name:    "test-plugin",
		Type:    PluginTypeLLM,
		Enabled: true,
		Version: "1.0.0",
		Entry:   "./bin/test",
	}

	data, err := json.Marshal(pm)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded PluginMetadata
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Name != pm.Name {
		t.Errorf("Decoded Name = %v, want %v", decoded.Name, pm.Name)
	}

	if decoded.Type != pm.Type {
		t.Errorf("Decoded Type = %v, want %v", decoded.Type, pm.Type)
	}
}

func TestAgentStates(t *testing.T) {
	tests := []struct {
		name  string
		state AgentState
	}{
		{"Idle", AgentStateIdle},
		{"Working", AgentStateWorking},
		{"Waiting", AgentStateWaiting},
		{"Error", AgentStateError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.state == "" {
				t.Errorf("AgentState %s is empty", tt.name)
			}
		})
	}
}

func TestTaskStates(t *testing.T) {
	tests := []struct {
		name  string
		state TaskState
	}{
		{"Pending", TaskStatePending},
		{"InProgress", TaskStateInProgress},
		{"Completed", TaskStateCompleted},
		{"Failed", TaskStateFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.state == "" {
				t.Errorf("TaskState %s is empty", tt.name)
			}
		})
	}
}

func TestMemoryJSON(t *testing.T) {
	memory := Memory{
		ShortTerm: []Message{
			{
				Role:    "user",
				Content: "Hello",
				Ts:      1234567890000,
			},
			{
				Role:    "assistant",
				Content: "Hi there!",
				Ts:      1234567891000,
			},
		},
		LongTerm: []MemoryItem{
			{
				Key:       "user_name",
				Value:     "Alice",
				UpdateAt:  1234567890000,
				ExpiresAt: 0,
			},
		},
	}

	data, err := json.Marshal(memory)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Memory
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if len(decoded.ShortTerm) != len(memory.ShortTerm) {
		t.Errorf("Decoded ShortTerm length = %v, want %v", len(decoded.ShortTerm), len(memory.ShortTerm))
	}

	if len(decoded.LongTerm) != len(memory.LongTerm) {
		t.Errorf("Decoded LongTerm length = %v, want %v", len(decoded.LongTerm), len(memory.LongTerm))
	}
}

func TestAgentWithoutMemory(t *testing.T) {
	agent := Agent{
		Name:       "TestAgent",
		Profession: "TestProfession",
		State:      AgentStateIdle,
	}

	data, err := json.Marshal(agent)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Agent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Memory != nil {
		t.Errorf("Decoded Memory should be nil, got %v", decoded.Memory)
	}
}

func TestPluginTypeConstantValues(t *testing.T) {
	tests := []struct {
		name  string
		value PluginType
	}{
		{"LLM constant", PluginTypeLLM},
		{"Search constant", PluginTypeSearch},
		{"Channel constant", PluginTypeChannel},
		{"Tool constant", PluginTypeTool},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value == "" {
				t.Errorf("PluginType constant %s is empty", tt.name)
			}
		})
	}
}

func TestTaskStepJSON(t *testing.T) {
	step := TaskStep{
		Description: "Test step",
		Done:        false,
	}

	data, err := json.Marshal(step)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded TaskStep
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Description != step.Description {
		t.Errorf("Decoded Description = %v, want %v", decoded.Description, step.Description)
	}

	if decoded.Done != step.Done {
		t.Errorf("Decoded Done = %v, want %v", decoded.Done, step.Done)
	}
}

func TestMessageJSON(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
		Ts:      1234567890000,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Message
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Role != msg.Role {
		t.Errorf("Decoded Role = %v, want %v", decoded.Role, msg.Role)
	}

	if decoded.Content != msg.Content {
		t.Errorf("Decoded Content = %v, want %v", decoded.Content, msg.Content)
	}

	if decoded.Ts != msg.Ts {
		t.Errorf("Decoded Ts = %v, want %v", decoded.Ts, msg.Ts)
	}
}

func TestMemoryItemJSON(t *testing.T) {
	item := MemoryItem{
		Key:       "test_key",
		Value:     "test_value",
		UpdateAt:  1234567890000,
		ExpiresAt: 12345678900,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded MemoryItem
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.Key != item.Key {
		t.Errorf("Decoded Key = %v, want %v", decoded.Key, item.Key)
	}

	if decoded.Value != item.Value {
		t.Errorf("Decoded Value = %v, want %v", decoded.Value, item.Value)
	}
}
