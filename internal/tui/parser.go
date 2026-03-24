package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ParsedInput represents the result of parsing user input.
type ParsedInput struct {
	Original  string      // Original input
	CleanText string      // Text without @ references
	Refs      []Reference // List of references found
}

// Reference represents a reference in the input (@skillname or @filepath).
type Reference struct {
	Type    string // "skill" or "file"
	Raw     string // Raw text (e.g., "@skillname")
	Name    string // Name without @ (e.g., "skillname")
	Content string // Content to inject
}

// ParseInput parses user input and extracts @ references.
func ParseInput(input string, skillProvider SkillProvider) (*ParsedInput, error) {
	result := &ParsedInput{
		Original: input,
		Refs:     make([]Reference, 0),
	}

	words := strings.Fields(input)
	var cleanWords []string

	for _, word := range words {
		if strings.HasPrefix(word, "@") {
			ref := parseReference(word, skillProvider)
			if ref != nil {
				result.Refs = append(result.Refs, *ref)
				// Add the content as context instead of the raw reference
				cleanWords = append(cleanWords, fmt.Sprintf("[%s: %s]", ref.Type, ref.Name))
				continue
			}
		}
		cleanWords = append(cleanWords, word)
	}

	result.CleanText = strings.Join(cleanWords, " ")
	return result, nil
}

// SkillProvider is an interface for looking up skills.
type SkillProvider interface {
	GetSkill(name string) (SkillInfo, error)
	ListSkills() []SkillInfo
}

// SkillInfo represents skill information.
type SkillInfo struct {
	Name        string
	Description string
	Content     string // Full content if available
}

// parseReference parses a single @ reference.
func parseReference(word string, provider SkillProvider) *Reference {
	name := word[1:] // Remove @ prefix
	if name == "" {
		return nil
	}

	// Try to parse as skill first
	if provider != nil {
		skill, err := provider.GetSkill(name)
		if err == nil {
			return &Reference{
				Type:    "skill",
				Raw:     word,
				Name:    name,
				Content: skill.Content,
			}
		}
	}

	// Try to parse as file path
	// Expand ~ to home directory
	if strings.HasPrefix(name, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			name = filepath.Join(home, name[1:])
		}
	}

	// Check if file exists and read it
	content, err := readFileContent(name)
	if err == nil {
		return &Reference{
			Type:    "file",
			Raw:     word,
			Name:    word[1:], // Keep original name for display
			Content: content,
		}
	}

	// If not a valid skill or file, return nil
	return nil
}

// readFileContent reads the content of a file.
func readFileContent(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", fmt.Errorf("path is a directory: %s", path)
	}

	// Limit file size to 100KB for safety
	if info.Size() > 100*1024 {
		return "", fmt.Errorf("file too large: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// BuildContext builds the context string from references.
func (my *ParsedInput) BuildContext() string {
	if len(my.Refs) == 0 {
		return ""
	}

	var parts []string
	for _, ref := range my.Refs {
		switch ref.Type {
		case "skill":
			parts = append(parts, fmt.Sprintf("## Skill: %s\n%s", ref.Name, ref.Content))
		case "file":
			parts = append(parts, fmt.Sprintf("## File: %s\n```\n%s\n```", ref.Name, ref.Content))
		}
	}

	return "\n\n---\nContext:\n" + strings.Join(parts, "\n\n")
}

// GetReferencedSkills returns the names of referenced skills.
func (my *ParsedInput) GetReferencedSkills() []string {
	var skills []string
	for _, ref := range my.Refs {
		if ref.Type == "skill" {
			skills = append(skills, ref.Name)
		}
	}
	return skills
}

// GetReferencedFiles returns the paths of referenced files.
func (my *ParsedInput) GetReferencedFiles() []string {
	var files []string
	for _, ref := range my.Refs {
		if ref.Type == "file" {
			files = append(files, ref.Name)
		}
	}
	return files
}
