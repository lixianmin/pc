package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentsMd(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectError  bool
		expectName   string
		expectPersonality []string
		expectProfession  string
		expectInstructions string
	}{
		{
			name: "完整 agents.md",
			content: `# PersonalClaw

## Personality
- 友好
- 专业
- 高效

## Profession
通用助手，擅长编程和数据分析

## Instructions
你是一个有用的 AI 助手。请遵循以下原则：
1. 回答简洁明了
2. 代码使用中文注释
3. 主动提供帮助
`,
			expectError:  false,
			expectName:   "PersonalClaw",
			expectPersonality: []string{"友好", "专业", "高效"},
			expectProfession:  "通用助手，擅长编程和数据分析",
			expectInstructions: "你是一个有用的 AI 助手。请遵循以下原则：\n1. 回答简洁明了\n2. 代码使用中文注释\n3. 主动提供帮助",
		},
		{
			name: "只有名称的 agents.md",
			content: `# SimpleAgent

`,
			expectError:  false,
			expectName:   "SimpleAgent",
			expectPersonality: nil,
			expectProfession:  "",
			expectInstructions: "",
		},
		{
			name: "空文件",
			content: ``,
			expectError: true,
		},
		{
			name: "只有描述没有名称",
			content: `## Personality
- 友好
`,
			expectError: true,
		},
		{
			name: "复杂格式 agents.md",
			content: `# CodeReviewAgent

## Personality
- 严谨
- 细致

## Profession
代码审查专家，专注于 Go 和 Python 代码质量

## Instructions
你的职责是审查代码并提供改进建议：

### 审查重点
1. 代码安全性
2. 性能优化
3. 可读性

### 输出格式
- 问题等级：[高/中/低]
- 具体位置
- 改进建议
`,
			expectError:  false,
			expectName:   "CodeReviewAgent",
			expectPersonality: []string{"严谨", "细致"},
			expectProfession:  "代码审查专家，专注于 Go 和 Python 代码质量",
			expectInstructions: "你的职责是审查代码并提供改进建议：\n\n### 审查重点\n1. 代码安全性\n2. 性能优化\n3. 可读性\n\n### 输出格式\n- 问题等级：[高/中/低]\n- 具体位置\n- 改进建议",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory and file
			tmpDir := t.TempDir()
			agentsPath := filepath.Join(tmpDir, "agents.md")

			err := os.WriteFile(agentsPath, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			// Test loading
			config, err := LoadAgentsMd(agentsPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify name
			if config.Name != tt.expectName {
				t.Errorf("name mismatch: got %q, want %q", config.Name, tt.expectName)
			}

			// Verify personality
			if len(config.Personality) != len(tt.expectPersonality) {
				t.Errorf("personality length mismatch: got %d, want %d", len(config.Personality), len(tt.expectPersonality))
			} else {
				for i, p := range config.Personality {
					if p != tt.expectPersonality[i] {
						t.Errorf("personality[%d] mismatch: got %q, want %q", i, p, tt.expectPersonality[i])
					}
				}
			}

			// Verify profession
			if config.Profession != tt.expectProfession {
				t.Errorf("profession mismatch: got %q, want %q", config.Profession, tt.expectProfession)
			}

			// Verify instructions
			if config.Instructions != tt.expectInstructions {
				t.Errorf("instructions mismatch: got %q, want %q", config.Instructions, tt.expectInstructions)
			}
		})
	}
}

func TestLoadAgentsMd_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "non_existent.md")

	_, err := LoadAgentsMd(nonExistentPath)
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestAgentsMdConfig_ToSystemPrompt(t *testing.T) {
	tests := []struct {
		name     string
		config   AgentsMdConfig
		expected string
	}{
		{
			name: "完整配置",
			config: AgentsMdConfig{
				Name:        "TestAgent",
				Personality: []string{"友好", "专业"},
				Profession:  "测试专家",
				Instructions: "请认真测试",
			},
			expected: `你是 TestAgent。

## 职业
测试专家

## 性格特征
- 友好
- 专业

## 行为准则
请认真测试`,
		},
		{
			name: "只有名称",
			config: AgentsMdConfig{
				Name: "SimpleAgent",
			},
			expected: "你是 SimpleAgent。",
		},
		{
			name: "没有性格",
			config: AgentsMdConfig{
				Name:       "NoTraitAgent",
				Profession: "专业助手",
				Instructions: "帮助用户",
			},
			expected: `你是 NoTraitAgent。

## 职业
专业助手

## 行为准则
帮助用户`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ToSystemPrompt()
			if result != tt.expected {
				t.Errorf("ToSystemPrompt() mismatch:\ngot:\n%s\n\nwant:\n%s", result, tt.expected)
			}
		})
	}
}
