package skill

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
	FilePath    string   `json:"file_path"`
}

// ISkillManager is the interface for skill management.
type ISkillManager interface {
	// LoadSkills loads all skills from the skill directory.
	LoadSkills(skillDir string) ([]Skill, error)

	// GetSkill returns a skill by name.
	GetSkill(name string) (*Skill, error)

	// ListSkills returns all loaded skills.
	ListSkills() []Skill
}

// SkillManager is the skill manager implementation.
type SkillManager struct {
	skills []Skill
}

// NewSkillManager creates a new skill manager.
func NewSkillManager() *SkillManager {
	return &SkillManager{
		skills: []Skill{},
	}
}

// LoadSkills loads all skills from the skill directory.
func (my *SkillManager) LoadSkills(skillDir string) ([]Skill, error) {
	// Check if directory exists
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("skill directory not found: %s", skillDir)
	}

	// Read all .md files in the directory
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill directory: %w", err)
	}

	var skills []Skill
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Parse skill file
		skillFilePath := filepath.Join(skillDir, entry.Name())
		skill, err := parseSkillFile(skillFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse skill file %s: %w", skillFilePath, err)
		}

		skills = append(skills, skill)
	}

	my.skills = skills
	return skills, nil
}

// parseSkillFile parses a skill markdown file.
func parseSkillFile(filePath string) (Skill, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Skill{}, fmt.Errorf("failed to read skill file: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	var skill Skill
	skill.FilePath = filePath

	// Parse skill name (first heading)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			skill.Name = strings.TrimSpace(line[2:])
		}
		// Parse description (line after name, before ## Steps)
		if strings.HasPrefix(line, "## Steps") {
			break
		}
		if skill.Name != "" && !strings.HasPrefix(line, "#") && line != "" {
			if skill.Description == "" {
				skill.Description = line
			}
		}
	}

	// Parse steps
	inSteps := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## Steps") {
			inSteps = true
			continue
		}
		if inSteps {
			// Check for numbered list items (e.g., "1.", "2.", etc.)
			if len(line) > 2 && line[0] >= '0' && line[0] <= '9' && line[1] == '.' {
				step := strings.TrimSpace(line[2:])
				if step != "" {
					skill.Steps = append(skill.Steps, step)
				}
			}
		}
	}

	return skill, nil
}

// GetSkill returns a skill by name.
func (my *SkillManager) GetSkill(name string) (*Skill, error) {
	for _, skill := range my.skills {
		if skill.Name == name {
			return &skill, nil
		}
	}
	return nil, fmt.Errorf("skill not found: %s", name)
}

// ListSkills returns all loaded skills.
func (my *SkillManager) ListSkills() []Skill {
	return my.skills
}
