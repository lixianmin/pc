package task

import (
	"testing"
)

// MockLLMClient is a mock LLM client for testing
type MockLLMClient struct {
	DecomposeFunc func(goal string) ([]string, error)
}

func (m *MockLLMClient) DecomposeGoal(goal string) ([]string, error) {
	if m.DecomposeFunc != nil {
		return m.DecomposeFunc(goal)
	}
	return []string{"step1", "step2"}, nil
}

func TestNewDecomposer(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new decomposer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecomposer()
			if d == nil {
				t.Error("NewDecomposer() returned nil")
			}
		})
	}
}

func TestDecomposer_ParseGoal(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantGoal   string
		wantErr    bool
	}{
		{
			name:       "parse valid goal",
			input:      "Deploy v1.0.0 to production",
			wantGoal:   "Deploy v1.0.0 to production",
			wantErr:    false,
		},
		{
			name:       "parse empty input",
			input:      "",
			wantGoal:   "",
			wantErr:    true,
		},
		{
			name:       "parse whitespace input",
			input:      "   ",
			wantGoal:   "",
			wantErr:    true,
		},
		{
			name:       "parse goal with prefix",
			input:      "我的目标是：学习Go语言",
			wantGoal:   "学习Go语言",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecomposer()
			goal, err := d.ParseGoal(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGoal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && goal != tt.wantGoal {
				t.Errorf("ParseGoal() = %v, want %v", goal, tt.wantGoal)
			}
		})
	}
}

func TestDecomposer_Decompose(t *testing.T) {
	tests := []struct {
		name          string
		goal          string
		mockSteps     []string
		mockErr       error
		wantStepCount int
		wantErr       bool
	}{
		{
			name:          "decompose simple goal",
			goal:          "Build a website",
			mockSteps:     []string{"Design UI", "Write code", "Deploy"},
			wantStepCount: 3,
			wantErr:       false,
		},
		{
			name:          "decompose complex goal",
			goal:          "Setup CI/CD pipeline",
			mockSteps:     []string{"Configure GitHub Actions", "Add tests", "Setup deployment"},
			wantStepCount: 3,
			wantErr:       false,
		},
		{
			name:          "decompose empty goal",
			goal:          "",
			mockSteps:     nil,
			wantStepCount: 0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecomposer()
			d.SetLLMClient(&MockLLMClient{
				DecomposeFunc: func(goal string) ([]string, error) {
					return tt.mockSteps, tt.mockErr
				},
			})

			steps, err := d.Decompose(tt.goal)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decompose() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(steps) != tt.wantStepCount {
				t.Errorf("Decompose() step count = %v, want %v", len(steps), tt.wantStepCount)
			}
		})
	}
}

func TestDecomposer_ConvertToTask(t *testing.T) {
	tests := []struct {
		name          string
		goal          string
		steps         []string
		wantTaskTitle string
		wantStepCount int
		wantErr       bool
	}{
		{
			name:          "convert with steps",
			goal:          "Build API",
			steps:         []string{"Design schema", "Implement endpoints", "Add tests"},
			wantTaskTitle: "Build API",
			wantStepCount: 3,
			wantErr:       false,
		},
		{
			name:          "convert without steps",
			goal:          "Simple task",
			steps:         []string{},
			wantTaskTitle: "Simple task",
			wantStepCount: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecomposer()
			task, err := d.ConvertToTask(tt.goal, tt.steps)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertToTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if task.Title != tt.wantTaskTitle {
					t.Errorf("ConvertToTask() task.Title = %v, want %v", task.Title, tt.wantTaskTitle)
				}
				if len(task.Steps) != tt.wantStepCount {
					t.Errorf("ConvertToTask() step count = %v, want %v", len(task.Steps), tt.wantStepCount)
				}
			}
		})
	}
}

func TestDecomposer_DecomposeAndCreateTask(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		mockSteps     []string
		wantTaskTitle string
		wantStepCount int
		wantErr       bool
	}{
		{
			name:          "decompose and create task",
			input:         "Deploy new feature",
			mockSteps:     []string{"Code review", "Merge PR", "Deploy to staging", "Deploy to prod"},
			wantTaskTitle: "Deploy new feature",
			wantStepCount: 4,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDecomposer()
			d.SetLLMClient(&MockLLMClient{
				DecomposeFunc: func(goal string) ([]string, error) {
					return tt.mockSteps, nil
				},
			})

			task, err := d.DecomposeAndCreateTask(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecomposeAndCreateTask() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if task.Title != tt.wantTaskTitle {
					t.Errorf("DecomposeAndCreateTask() task.Title = %v, want %v", task.Title, tt.wantTaskTitle)
				}
				if len(task.Steps) != tt.wantStepCount {
					t.Errorf("DecomposeAndCreateTask() step count = %v, want %v", len(task.Steps), tt.wantStepCount)
				}
			}
		})
	}
}
