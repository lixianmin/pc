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
	LoadSkills(skillDir string) ([]Skill, error)
	GetSkill(name string) *Skill
	ListSkills() []Skill
}
