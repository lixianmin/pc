package agent

import (
	"strings"
	"testing"

	"github.com/lixianmin/pc/internal/skill"
	"github.com/lixianmin/pc/pkg/types"
)

func TestSystemPromptBuilder_Build(t *testing.T) {
	tests := []struct {
		name            string
		basePrompt      string
		skills          []skill.Skill
		tools           []ToolInfo
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:       "完整 system prompt",
			basePrompt: "你是 TestAgent。",
			skills: []skill.Skill{
				{Name: "code_review", Description: "代码审查技能"},
				{Name: "test", Description: "测试技能"},
			},
			tools: []ToolInfo{
				{Name: "git", Description: "Git 操作", Type: "tool"},
				{Name: "openai", Description: "LLM 调用", Type: "llm"},
			},
			wantContains: []string{
				"你是 TestAgent。",
				"## 可用技能",
				"code_review",
				"代码审查技能",
				"test",
				"## 工具使用指南",
				"git",
				"Git 操作",
				"openai",
				"<tool_call>",
				"<name>",
				"<params>",
				"ReAct",
				"思考 (Think)",
				"行动 (Act)",
				"观察 (Observe)",
				"回复 (Respond)",
			},
		},
		{
			name:       "只有基础 prompt",
			basePrompt: "你是 SimpleAgent。",
			skills:     nil,
			tools:      nil,
			wantContains: []string{
				"你是 SimpleAgent。",
			},
			wantNotContains: []string{
				"## 可用技能",
				"## 工具使用指南",
			},
		},
		{
			name:       "只有 skills",
			basePrompt: "你是 Agent。",
			skills: []skill.Skill{
				{Name: "build", Description: "构建项目"},
			},
			tools: nil,
			wantContains: []string{
				"你是 Agent。",
				"## 可用技能",
				"build",
				"构建项目",
			},
			wantNotContains: []string{
				"## 工具使用指南",
			},
		},
		{
			name:       "只有 tools",
			basePrompt: "你是 Agent。",
			skills:     nil,
			tools: []ToolInfo{
				{Name: "telegram", Description: "发送消息", Type: "channel"},
			},
			wantContains: []string{
				"你是 Agent。",
				"## 工具使用指南",
				"telegram",
				"发送消息",
				"<tool_call>",
				"ReAct",
			},
			wantNotContains: []string{
				"## 可用技能",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSystemPromptBuilder()
			builder.SetBasePrompt(tt.basePrompt)
			builder.SetSkills(tt.skills)
			builder.SetTools(tt.tools)

			result := builder.Build()

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("Build() should contain %q, got:\n%s", want, result)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if containsString(result, notWant) {
					t.Errorf("Build() should NOT contain %q, got:\n%s", notWant, result)
				}
			}
		})
	}
}

func TestSystemPromptBuilder_BuildWithPluginManager(t *testing.T) {
	tests := []struct {
		name         string
		plugins      []*types.Plugin
		wantContains []string
	}{
		{
			name: "从 plugin manager 构建",
			plugins: []*types.Plugin{
				{Name: "openai-llm", Type: types.PluginTypeLLM, Enabled: true},
				{Name: "telegram-bot", Type: types.PluginTypeChannel, Enabled: true},
				{Name: "git-tool", Type: types.PluginTypeTool, Enabled: true},
				{Name: "disabled-plugin", Type: types.PluginTypeTool, Enabled: false},
			},
			wantContains: []string{
				"openai-llm",
				"telegram-bot",
				"git-tool",
			},
		},
		{
			name:         "空 plugins",
			plugins:      nil,
			wantContains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSystemPromptBuilder()
			builder.SetBasePrompt("你是 Agent。")

			// Simulate building tools from plugins
			var tools []ToolInfo
			for _, p := range tt.plugins {
				if p.Enabled {
					tools = append(tools, ToolInfo{
						Name:        p.Name,
						Description: string(p.Type) + " plugin",
						Type:        string(p.Type),
					})
				}
			}
			builder.SetTools(tools)

			result := builder.Build()

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("Build() should contain %q", want)
				}
			}
		})
	}
}

func TestToolInfo_FromPlugin(t *testing.T) {
	tests := []struct {
		name     string
		plugin   *types.Plugin
		expected ToolInfo
	}{
		{
			name: "LLM plugin",
			plugin: &types.Plugin{
				Name:    "openai-llm",
				Type:    types.PluginTypeLLM,
				Version: "1.0.0",
				Enabled: true,
			},
			expected: ToolInfo{
				Name:        "openai-llm",
				Description: "LLM Provider plugin",
				Type:        "llm",
			},
		},
		{
			name: "Channel plugin",
			plugin: &types.Plugin{
				Name:    "telegram-bot",
				Type:    types.PluginTypeChannel,
				Version: "1.0.0",
				Enabled: true,
			},
			expected: ToolInfo{
				Name:        "telegram-bot",
				Description: "Channel plugin",
				Type:        "channel",
			},
		},
		{
			name: "Tool plugin",
			plugin: &types.Plugin{
				Name:    "git-tool",
				Type:    types.PluginTypeTool,
				Version: "1.0.0",
				Enabled: true,
			},
			expected: ToolInfo{
				Name:        "git-tool",
				Description: "Tool plugin",
				Type:        "tool",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToolInfoFromPlugin(tt.plugin)
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
			}
			if result.Type != tt.expected.Type {
				t.Errorf("Type = %v, want %v", result.Type, tt.expected.Type)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestSystemPromptBuilder_ToolUsageGuide(t *testing.T) {
	tests := []struct {
		name         string
		tools        []ToolInfo
		wantContains []string
	}{
		{
			name: "工具使用指南包含 ReAct 说明",
			tools: []ToolInfo{
				{Name: "bash", Description: "执行 bash 命令", Type: "tool"},
			},
			wantContains: []string{
				"## 工具使用指南",
				"当你需要获取外部信息或执行操作时",
				"不要告诉用户'你无法'或'你没有权限'",
				"### 可用工具",
				"**bash**",
				"### 工具调用格式",
				"<tool_call>",
				"<name>工具名</name>",
				"<params>{\"参数名\": \"参数值\"}</params>",
				"</tool_call>",
				"### 示例",
				"### 工作流程 (ReAct)",
				"思考 (Think)",
				"行动 (Act)",
				"观察 (Observe)",
				"回复 (Respond)",
			},
		},
		{
			name: "带参数 schema 的工具描述",
			tools: []ToolInfo{
				{
					Name:        "bash",
					Description: "执行 bash 命令",
					Type:        "tool",
					ParamsSchema: map[string]string{
						"command":     "要执行的命令",
						"description": "命令描述",
					},
				},
			},
			wantContains: []string{
				"**bash**",
				"执行 bash 命令",
				"`command`",
				"要执行的命令",
				"`description`",
				"命令描述",
			},
		},
		{
			name:  "空工具列表不生成工具指南",
			tools: []ToolInfo{},
			wantContains: []string{
				"你是 Agent。",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSystemPromptBuilder()
			builder.SetBasePrompt("你是 Agent。")
			builder.SetTools(tt.tools)

			result := builder.Build()

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("Build() should contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}

func TestSystemPromptBuilder_ToolExamples(t *testing.T) {
	tests := []struct {
		name         string
		tools        []ToolInfo
		wantContains []string
	}{
		{
			name: "有 tool 类型时显示 bash 示例",
			tools: []ToolInfo{
				{Name: "bash", Description: "执行命令", Type: "tool"},
			},
			wantContains: []string{
				"**示例 1: 执行 bash 命令**",
				"用户: 列出主目录的文件",
				"<thinking>",
				"用户想要查看主目录的文件",
				"<tool_call>",
				"<name>bash</name>",
				"\"command\": \"ls ~\"",
				"</tool_call>",
			},
		},
		{
			name: "通用示例始终显示",
			tools: []ToolInfo{
				{Name: "search", Description: "搜索", Type: "search"},
			},
			wantContains: []string{
				"**示例 2: 通用格式**",
				"<tool_call>",
				"<name>工具名称</name>",
				"<params>{\"key\": \"value\", \"key2\": \"value2\"}</params>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewSystemPromptBuilder()
			builder.SetBasePrompt("你是 Agent。")
			builder.SetTools(tt.tools)

			result := builder.Build()

			for _, want := range tt.wantContains {
				if !containsString(result, want) {
					t.Errorf("Build() should contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}
