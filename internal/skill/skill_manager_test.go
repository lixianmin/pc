package skill

import (
	"os"
	"testing"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "create new skill manager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSkillManager()
			if got == nil {
				t.Error("NewManager() returned nil")
			}
		})
	}
}

func TestLoadSkills(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
	}{
		{
			name: "load skills from directory",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				// Create a simple skill file with multiple steps
				skillContent := `# Test Skill
A simple test skill.

## Steps
1. Do step one
2. Do step two
3. Do step three
4. Do step four
`
				skillPath := tmpDir + "/test_skill.md"
				_ = os.WriteFile(skillPath, []byte(skillContent), 0644)
				return tmpDir
			},
			wantErr: false,
		},
		{
			name: "load from non-existent directory",
			setup: func(t *testing.T) string {
				return "/non-existent-dir"
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSkillManager()
			skillDir := tt.setup(t)

			skills, err := m.LoadSkills(skillDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSkills() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(skills) == 0 {
				t.Error("LoadSkills() returned empty skills")
			}
		})
	}
}

func TestGetSkill(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*SkillManager)
		skillName string
		wantErr   bool
	}{
		{
			name: "get existing skill",
			setup: func(m *SkillManager) {
				m.skills = []Skill{
					{Name: "test-skill", Description: "Test", Steps: []string{"step1", "step2"}, FilePath: "/path/to/skill.md"},
				}
			},
			skillName: "test-skill",
			wantErr:   false,
		},
		{
			name: "get non-existent skill",
			setup: func(m *SkillManager) {
				m.skills = []Skill{}
			},
			skillName: "non-existent",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSkillManager()
			tt.setup(m)

			skill, err := m.GetSkill(tt.skillName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSkill() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && skill == nil {
				t.Error("GetSkill() returned nil")
			}
		})
	}
}

func TestListSkills(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*SkillManager)
		wantCount int
	}{
		{
			name: "list skills with entries",
			setup: func(m *SkillManager) {
				m.skills = []Skill{
					{Name: "skill1", Description: "First skill", Steps: []string{"a"}, FilePath: "/path1"},
					{Name: "skill2", Description: "Second skill", Steps: []string{"b"}, FilePath: "/path2"},
				}
			},
			wantCount: 2,
		},
		{
			name: "list empty skills",
			setup: func(m *SkillManager) {
				m.skills = []Skill{}
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewSkillManager()
			tt.setup(m)

			skills := m.ListSkills()
			if len(skills) != tt.wantCount {
				t.Errorf("ListSkills() count = %v, want %v", len(skills), tt.wantCount)
			}
		})
	}
}
