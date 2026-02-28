package agent

import (
	"fmt"
	"os"
	"strings"
)

// AgentsMdConfig represents the configuration loaded from agents.md file.
type AgentsMdConfig struct {
	Name         string   // Agent name (from first heading)
	Personality  []string // Personality traits (from ## Personality section)
	Profession   string   // Profession description (from ## Profession section)
	Instructions string   // Behavior instructions (from ## Instructions section)
}

// LoadAgentsMd loads and parses an agents.md file.
// The file format is:
//
//	# Agent Name
//	## Personality
//	- Trait 1
//	- Trait 2
//	## Profession
//	Profession description
//	## Instructions
//	Behavior instructions...
func LoadAgentsMd(path string) (*AgentsMdConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read agents.md: %w", err)
	}

	content := string(data)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("agents.md is empty")
	}

	config := &AgentsMdConfig{}

	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionContent []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Parse name from first level-1 heading
		if i == 0 && strings.HasPrefix(trimmed, "# ") {
			config.Name = strings.TrimSpace(trimmed[2:])
			continue
		}

		// Check for section headers
		if strings.HasPrefix(trimmed, "## ") {
			// Save previous section content if any
			if currentSection != "" && len(sectionContent) > 0 {
				saveSection(config, currentSection, sectionContent)
			}
			currentSection = strings.TrimSpace(trimmed[3:])
			sectionContent = nil
			continue
		}

		// Collect content for current section
		if currentSection != "" {
			sectionContent = append(sectionContent, line)
		}
	}

	// Save last section
	if currentSection != "" && len(sectionContent) > 0 {
		saveSection(config, currentSection, sectionContent)
	}

	// Validate: name is required
	if config.Name == "" {
		return nil, fmt.Errorf("agents.md must have a name (first line should be '# AgentName')")
	}

	return config, nil
}

// saveSection saves content to the appropriate field based on section name.
func saveSection(config *AgentsMdConfig, section string, content []string) {
	switch strings.ToLower(section) {
	case "personality":
		config.Personality = parsePersonality(content)
	case "profession":
		config.Profession = strings.TrimSpace(strings.Join(content, "\n"))
	case "instructions":
		config.Instructions = strings.TrimSpace(strings.Join(content, "\n"))
	}
}

// parsePersonality parses personality traits from bullet list.
func parsePersonality(lines []string) []string {
	var traits []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Match bullet list items: "- trait" or "* trait"
		if strings.HasPrefix(trimmed, "- ") {
			trait := strings.TrimSpace(trimmed[2:])
			if trait != "" {
				traits = append(traits, trait)
			}
		} else if strings.HasPrefix(trimmed, "* ") {
			trait := strings.TrimSpace(trimmed[2:])
			if trait != "" {
				traits = append(traits, trait)
			}
		}
	}
	return traits
}

// ToSystemPrompt converts the agents.md config to a system prompt string.
func (c *AgentsMdConfig) ToSystemPrompt() string {
	var parts []string

	// Name
	parts = append(parts, fmt.Sprintf("你是 %s。", c.Name))

	// Profession
	if c.Profession != "" {
		parts = append(parts, "", "## 职业", c.Profession)
	}

	// Personality
	if len(c.Personality) > 0 {
		parts = append(parts, "", "## 性格特征")
		for _, trait := range c.Personality {
			parts = append(parts, "- "+trait)
		}
	}

	// Instructions
	if c.Instructions != "" {
		parts = append(parts, "", "## 行为准则", c.Instructions)
	}

	return strings.Join(parts, "\n")
}
