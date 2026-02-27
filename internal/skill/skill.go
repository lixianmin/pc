package skill

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
