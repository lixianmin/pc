package agent

import (
	"fmt"
	"strings"

	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/pkg/types"
)

// ToolInfo represents information about an available tool/plugin.
type ToolInfo struct {
	Name         string            // Tool name
	Description  string            // Tool description
	Type         string            // Tool type (llm, channel, tool, search)
	ParamsSchema map[string]string // Parameter name -> description
}

// ToolExample represents an example tool usage.
type ToolExample struct {
	Name        string
	Description string
	Params      string
	Output      string
}

// ToolInfoFromPlugin creates ToolInfo from a Plugin.
func ToolInfoFromPlugin(p *types.Plugin) ToolInfo {
	return ToolInfo{
		Name:         p.Name,
		Description:  fmt.Sprintf("%s plugin", p.Type),
		Type:         string(p.Type),
		ParamsSchema: make(map[string]string),
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

// Build constructs the complete system prompt with ReAct tool usage guide.
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

	// Tools section with ReAct guide
	if len(b.tools) > 0 {
		parts = append(parts, "", b.buildToolUsageGuide())
	}

	return strings.Join(parts, "\n")
}

// buildToolUsageGuide builds the tool usage guide section for ReAct loop.
func (b *SystemPromptBuilder) buildToolUsageGuide() string {
	var parts []string

	// Header
	parts = append(parts, "## 工具使用指南")
	parts = append(parts, "")
	parts = append(parts, "当你需要获取外部信息或执行操作时，可以使用以下工具。不要告诉用户'你无法'或'你没有权限'，而是主动使用合适的工具！")
	parts = append(parts, "")

	// Available tools list
	parts = append(parts, "### 可用工具")
	for _, t := range b.tools {
		parts = append(parts, b.formatToolDescription(t))
	}

	// Tool call format
	parts = append(parts, "")
	parts = append(parts, "### 工具调用格式")
	parts = append(parts, "")
	parts = append(parts, "当需要使用工具时，请使用以下 XML 格式：")
	parts = append(parts, "")
	parts = append(parts, "<tool_call>")
	parts = append(parts, "<name>工具名</name>")
	parts = append(parts, "<params>{\"参数名\": \"参数值\"}</params>")
	parts = append(parts, "</tool_call>")
	parts = append(parts, "")

	// Examples
	parts = append(parts, "### 示例")
	parts = append(parts, "")
	parts = append(parts, b.buildToolExamples())

	// ReAct workflow
	parts = append(parts, "")
	parts = append(parts, "### 工作流程 (ReAct)")
	parts = append(parts, "")
	parts = append(parts, "1. **思考 (Think)**: 分析用户需求，判断是否需要使用工具")
	parts = append(parts, "2. **行动 (Act)**: 如需工具，输出 `<tool_call>` 标签")
	parts = append(parts, "3. **观察 (Observe)**: 系统将执行工具并返回结果")
	parts = append(parts, "4. **回复 (Respond)**: 基于工具结果生成最终回复")
	parts = append(parts, "")
	parts = append(parts, "如果需要多个步骤，可以重复 1-3 步直到任务完成。")

	return strings.Join(parts, "\n")
}

// formatToolDescription formats a single tool description.
func (b *SystemPromptBuilder) formatToolDescription(t ToolInfo) string {
	var parts []string

	// Tool name and description
	if t.Description != "" {
		parts = append(parts, fmt.Sprintf("- **%s**: %s", t.Name, t.Description))
	} else {
		parts = append(parts, fmt.Sprintf("- **%s**", t.Name))
	}

	// Parameters
	if len(t.ParamsSchema) > 0 {
		var params []string
		for paramName, paramDesc := range t.ParamsSchema {
			params = append(params, fmt.Sprintf("    - `%s`: %s", paramName, paramDesc))
		}
		parts = append(parts, strings.Join(params, "\n"))
	}

	return strings.Join(parts, "\n")
}

// buildToolExamples builds example tool usage section.
func (b *SystemPromptBuilder) buildToolExamples() string {
	// Find a shell-like or file tool for the example
	var hasShellTool bool
	for _, t := range b.tools {
		if t.Type == "tool" {
			hasShellTool = true
			break
		}
	}

	var examples []string

	if hasShellTool {
		examples = append(examples, "**示例 1: 执行 shell 命令**")
		examples = append(examples, "")
		examples = append(examples, "用户: 列出主目录的文件")
		examples = append(examples, "")
		examples = append(examples, "你的回复:")
		examples = append(examples, "<thinking>")
		examples = append(examples, "用户想要查看主目录的文件，我可以使用 shell 工具执行 ls 命令。")
		examples = append(examples, "</thinking>")
		examples = append(examples, "")
		examples = append(examples, "<tool_call>")
		examples = append(examples, "<name>shell</name>")
		examples = append(examples, "<params>{\"command\": \"ls ~\", \"description\": \"List home directory files\"}</params>")
		examples = append(examples, "</tool_call>")
		examples = append(examples, "")
		examples = append(examples, "系统返回工具执行结果后，你回复:")
		examples = append(examples, "您的主目录包含以下文件：Desktop、Documents、Downloads...")
	}

	// Generic example
	examples = append(examples, "")
	examples = append(examples, "**示例 2: 通用格式**")
	examples = append(examples, "")
	examples = append(examples, "<tool_call>")
	examples = append(examples, "<name>工具名称</name>")
	examples = append(examples, "<params>{\"key\": \"value\", \"key2\": \"value2\"}</params>")
	examples = append(examples, "</tool_call>")

	return strings.Join(examples, "\n")
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
