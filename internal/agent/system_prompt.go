package agent

import (
	"fmt"
	"strings"

	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/pkg/types"
)

// ToolInfo represents information about an available tool/plugin.
type ToolInfo struct {
	Name        string // Tool name
	Description string // Tool description
	Type        string // Tool type (llm, channel, tool, search)
}

// ToolInfoFromPlugin creates ToolInfo from a Plugin.
func ToolInfoFromPlugin(p *types.Plugin) ToolInfo {
	return ToolInfo{
		Name:        p.Name,
		Description: fmt.Sprintf("%s plugin", p.Type),
		Type:        string(p.Type),
	}
}

// SystemPromptBuilder builds a complete system prompt for LLM.
type SystemPromptBuilder struct {
	basePrompt string
	skills     []skill.Skill
	tools      []ToolInfo
}

// NewSystemPromptBuilder creates a new system prompt builder.
func NewSystemPromptBuilder() *SystemPromptBuilder {
	return &SystemPromptBuilder{
		skills: []skill.Skill{},
		tools:  []ToolInfo{},
	}
}

// SetBasePrompt sets the base system prompt (from agents.md).
func (b *SystemPromptBuilder) SetBasePrompt(prompt string) *SystemPromptBuilder {
	b.basePrompt = prompt
	return b
}

// SetSkills sets the available skills.
func (b *SystemPromptBuilder) SetSkills(skills []skill.Skill) *SystemPromptBuilder {
	b.skills = skills
	return b
}

// SetTools sets the available tools.
func (b *SystemPromptBuilder) SetTools(tools []ToolInfo) *SystemPromptBuilder {
	b.tools = tools
	return b
}

// Build constructs the complete system prompt.
// Format:
//
//	{basePrompt}
//
//	## 可用技能
//	- skill_name: description
//
//	## 可用工具
//	- tool_name (type): description
func (b *SystemPromptBuilder) Build() string {
	var parts []string

	// Base prompt
	if b.basePrompt != "" {
		parts = append(parts, b.basePrompt)
	}

	// Skills section
	if len(b.skills) > 0 {
		parts = append(parts, "", "## 可用技能")
		for _, s := range b.skills {
			if s.Description != "" {
				parts = append(parts, fmt.Sprintf("- %s: %s", s.Name, s.Description))
			} else {
				parts = append(parts, fmt.Sprintf("- %s", s.Name))
			}
		}
	}

	// Tools section
	if len(b.tools) > 0 {
		parts = append(parts, "", "## 可用工具")
		for _, t := range b.tools {
			if t.Description != "" {
				parts = append(parts, fmt.Sprintf("- %s (%s): %s", t.Name, t.Type, t.Description))
			} else {
				parts = append(parts, fmt.Sprintf("- %s (%s)", t.Name, t.Type))
			}
		}
	}

	return strings.Join(parts, "\n")
}

// BuildDefault creates a default system prompt when agents.md is not available.
func BuildDefault(agentName string, skills []skill.Skill, plugins []*types.Plugin) string {
	builder := NewSystemPromptBuilder()
	builder.SetBasePrompt(fmt.Sprintf("你是 %s。", agentName))
	builder.SetSkills(skills)

	// Convert plugins to tools
	var tools []ToolInfo
	for _, p := range plugins {
		if p.Enabled {
			tools = append(tools, ToolInfoFromPlugin(p))
		}
	}
	builder.SetTools(tools)

	return builder.Build()
}
