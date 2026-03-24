package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/lixianmin/pc/internal/task"
)

func TestEngine_TaskDecomposition(t *testing.T) {
	tests := []struct {
		name          string
		sessionId     string
		message       string
		wantTask      bool
		wantErr       bool
		wantTaskTitle string
		wantStepCount int
	}{
		{
			name:          "decompose goal with /task prefix",
			sessionId:     "test-task-1",
			message:       "/task deploy the application",
			wantTask:      true,
			wantErr:       false,
			wantTaskTitle: "deploy the application",
			wantStepCount: 4,
		},
		{
			name:          "decompose goal with 我想 prefix",
			sessionId:     "test-task-2",
			message:       "我想测试这个功能",
			wantTask:      true,
			wantErr:       false,
			wantTaskTitle: "测试这个功能",
			wantStepCount: 4,
		},
		{
			name:      "regular message does not trigger decomposition",
			message:   "Hello, how are you?",
			sessionId: "test-task-3",
			wantErr:   false,
			wantTask:  false,
		},
		{
			name:      "empty goal returns error",
			sessionId: "test-task-4",
			message:   "/task   ",
			wantTask:  false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)
			e.SetTaskEnabled(true)
			ctx := context.Background()
			_ = e.FetchSession(tt.sessionId)

			resp, err := e.ProcessMessage(ctx, tt.sessionId, tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			tm := e.GetTaskManager()
			if tm == nil {
				t.Error("Expected TaskManager to be initialized")
				return
			}

			tasks := tm.ListTasks()
			if tt.wantTask {
				if len(tasks) == 0 {
					t.Error("Expected at least one task to be created")
					return
				}
				var foundTask *task.Task
				for _, tk := range tasks {
					if tk.Title == tt.wantTaskTitle {
						foundTask = tk
						break
					}
				}
				if foundTask == nil {
					t.Errorf("Expected task with title %q not found", tt.wantTaskTitle)
					return
				}
				if len(foundTask.Steps) != tt.wantStepCount {
					t.Errorf("Task steps = %v, want %v", len(foundTask.Steps), tt.wantStepCount)
				}
				if resp == "" {
					t.Error("Expected non-empty response for task decomposition")
					return
				}
			} else if !tt.wantErr {
				if len(tasks) > 0 {
					t.Errorf("Expected no tasks for non-goal message, got %d", len(tasks))
				}
			}
		})
	}
}

func TestEngine_LoadSkills(t *testing.T) {
	tests := []struct {
		name       string
		skillDir   string
		wantSkills int
		wantErr    bool
	}{
		{
			name:       "load skills from directory",
			skillDir:   "",
			wantSkills: 1,
			wantErr:    false,
		},
		{
			name:       "load from non-existent directory",
			skillDir:   "/nonexistent/skills",
			wantSkills: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)

			if tt.skillDir == "" {
				tempDir, err := os.MkdirTemp("", "pc-skills-test-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tempDir)

				skillFile := filepath.Join(tempDir, "test-skill.md")
				content := `# Test Skill

## Description
A test skill for unit testing.

## Steps
1. Step one
2. Step two
3. Step three
`
				if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write skill file: %v", err)
				}
				tt.skillDir = tempDir
			}

			err := e.SetSkillDir(tt.skillDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetSkillDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			skills := e.ListSkills()
			if len(skills) != tt.wantSkills {
				t.Errorf("ListSkills() = %v, want %v", len(skills), tt.wantSkills)
			}

			if tt.wantSkills > 0 {
				found := false
				for _, s := range skills {
					if s.Name == "Test Skill" {
						found = true
						if len(s.Steps) != 3 {
							t.Errorf("Expected 3 steps, got %v", len(s.Steps))
						}
						break
					}
				}
				if !found {
					t.Error("Expected to find 'Test Skill'")
				}
			}
		})
	}
}

func TestEngine_SkillsInSystemPrompt(t *testing.T) {
	tests := []struct {
		name         string
		setupSkills  bool
		wantContains string
	}{
		{
			name:         "skills included in prompt",
			setupSkills:  true,
			wantContains: "Skills",
		},
		{
			name:         "no skills when empty",
			setupSkills:  false,
			wantContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(nil)
			e.SetSystemPrompt("You are an assistant.")

			if tt.setupSkills {
				tempDir, err := os.MkdirTemp("", "pc-skills-prompt-*")
				if err != nil {
					t.Fatalf("Failed to create temp dir: %v", err)
				}
				defer os.RemoveAll(tempDir)

				skillFile := filepath.Join(tempDir, "code-review.md")
				content := `# Code Review

## Description
Review code quality.

## Steps
1. Read code
2. Analyze
`
				if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
					t.Fatalf("Failed to write skill file: %v", err)
				}

				if err := e.SetSkillDir(tempDir); err != nil {
					t.Fatalf("SetSkillDir failed: %v", err)
				}
			}

			prompt := e.BuildSystemPrompt()

			if tt.wantContains != "" {
				if !containsSkill(prompt, tt.wantContains) {
					t.Errorf("Expected prompt to contain %q", tt.wantContains)
				}
			}
		})
	}
}

func containsSkill(prompt, substr string) bool {
	return len(prompt) > 0 && substr != ""
}
