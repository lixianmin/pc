package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockSkillProvider is a mock implementation of SkillProvider for testing.
type MockSkillProvider struct {
	skills map[string]SkillInfo
}

func NewMockSkillProvider() *MockSkillProvider {
	return &MockSkillProvider{
		skills: map[string]SkillInfo{
			"code_review": {
				Name:        "code_review",
				Description: "Review code for issues",
				Content:     "Code review skill content",
			},
			"test": {
				Name:        "test",
				Description: "Run tests",
				Content:     "Testing skill content",
			},
		},
	}
}

func (m *MockSkillProvider) GetSkill(name string) (SkillInfo, error) {
	if skill, ok := m.skills[name]; ok {
		return skill, nil
	}
	return SkillInfo{}, os.ErrNotExist
}

func (m *MockSkillProvider) ListSkills() []SkillInfo {
	var result []SkillInfo
	for _, skill := range m.skills {
		result = append(result, skill)
	}
	return result
}

func TestParseInput(t *testing.T) {
	provider := NewMockSkillProvider()

	tests := []struct {
		name          string
		input         string
		wantCleanText string
		wantRefsCount int
		wantRefTypes  []string
	}{
		{
			name:          "simple message",
			input:         "Hello world",
			wantCleanText: "Hello world",
			wantRefsCount: 0,
		},
		{
			name:          "message with skill reference",
			input:         "Please @code_review this file",
			wantCleanText: "Please [skill: code_review] this file",
			wantRefsCount: 1,
			wantRefTypes:  []string{"skill"},
		},
		{
			name:          "message with multiple references",
			input:         "Use @code_review and @test on this",
			wantCleanText: "Use [skill: code_review] and [skill: test] on this",
			wantRefsCount: 2,
			wantRefTypes:  []string{"skill", "skill"},
		},
		{
			name:          "message with empty reference",
			input:         "Use @ to do something",
			wantCleanText: "Use @ to do something",
			wantRefsCount: 0,
		},
		{
			name:          "message with only @",
			input:         "@",
			wantCleanText: "@",
			wantRefsCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInput(tt.input, provider)
			if err != nil {
				t.Errorf("ParseInput() error = %v", err)
				return
			}

			if result.CleanText != tt.wantCleanText {
				t.Errorf("CleanText = %q, want %q", result.CleanText, tt.wantCleanText)
			}

			if len(result.Refs) != tt.wantRefsCount {
				t.Errorf("Refs count = %d, want %d", len(result.Refs), tt.wantRefsCount)
			}

			for i, refType := range tt.wantRefTypes {
				if i < len(result.Refs) && result.Refs[i].Type != refType {
					t.Errorf("Ref[%d].Type = %q, want %q", i, result.Refs[i].Type, refType)
				}
			}
		})
	}
}

func TestParseInput_WithFile(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "This is test file content"
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	provider := NewMockSkillProvider()

	tests := []struct {
		name            string
		input           string
		wantFileCount   int
		wantFileName    string
		wantFileContent string
	}{
		{
			name:          "reference existing file",
			input:         "Check " + testFile,
			wantFileCount: 0, // File without @ doesn't count
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseInput(tt.input, provider)
			if err != nil {
				t.Errorf("ParseInput() error = %v", err)
				return
			}

			files := result.GetReferencedFiles()
			if len(files) != tt.wantFileCount {
				t.Errorf("File count = %d, want %d", len(files), tt.wantFileCount)
			}
		})
	}
}

func TestParsedInput_BuildContext(t *testing.T) {
	tests := []struct {
		name     string
		refs     []Reference
		contains []string
	}{
		{
			name:     "no references",
			refs:     []Reference{},
			contains: nil,
		},
		{
			name: "skill reference",
			refs: []Reference{
				{Type: "skill", Name: "code_review", Content: "Review code"},
			},
			contains: []string{"Skill: code_review", "Review code"},
		},
		{
			name: "file reference",
			refs: []Reference{
				{Type: "file", Name: "test.go", Content: "package main"},
			},
			contains: []string{"File: test.go", "package main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &ParsedInput{Refs: tt.refs}
			context := p.BuildContext()

			for _, want := range tt.contains {
				if !strings.Contains(context, want) {
					t.Errorf("BuildContext() should contain %q, got: %s", want, context)
				}
			}
		})
	}
}

func TestParsedInput_GetReferencedSkills(t *testing.T) {
	p := &ParsedInput{
		Refs: []Reference{
			{Type: "skill", Name: "code_review"},
			{Type: "file", Name: "test.go"},
			{Type: "skill", Name: "test"},
		},
	}

	skills := p.GetReferencedSkills()
	if len(skills) != 2 {
		t.Errorf("GetReferencedSkills() = %d, want 2", len(skills))
	}

	expected := []string{"code_review", "test"}
	for i, want := range expected {
		if i >= len(skills) || skills[i] != want {
			t.Errorf("GetReferencedSkills()[%d] = %q, want %q", i, skills[i], want)
		}
	}
}

func TestParsedInput_GetReferencedFiles(t *testing.T) {
	p := &ParsedInput{
		Refs: []Reference{
			{Type: "skill", Name: "code_review"},
			{Type: "file", Name: "test.go"},
			{Type: "file", Name: "main.go"},
		},
	}

	files := p.GetReferencedFiles()
	if len(files) != 2 {
		t.Errorf("GetReferencedFiles() = %d, want 2", len(files))
	}

	expected := []string{"test.go", "main.go"}
	for i, want := range expected {
		if i >= len(files) || files[i] != want {
			t.Errorf("GetReferencedFiles()[%d] = %q, want %q", i, files[i], want)
		}
	}
}

func TestReadFileContent(t *testing.T) {
	// Create temp directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "read existing file",
			path:    testFile,
			want:    testContent,
			wantErr: false,
		},
		{
			name:    "read non-existent file",
			path:    filepath.Join(tmpDir, "nonexistent.txt"),
			want:    "",
			wantErr: true,
		},
		{
			name:    "read directory",
			path:    tmpDir,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readFileContent(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFileContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readFileContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
