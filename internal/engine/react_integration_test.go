package engine

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/lixianmin/pc/pkg/types"
)

type MockPluginManager struct {
	plugins []*types.Plugin
	calls   []struct {
		name   string
		method string
		params any
	}
	responses map[string]any
	errors    map[string]error
}

func NewMockPluginManager() *MockPluginManager {
	return &MockPluginManager{
		plugins:   make([]*types.Plugin, 0),
		responses: make(map[string]any),
		errors:    make(map[string]error),
	}
}

func (m *MockPluginManager) ListPlugins() []*types.Plugin {
	return m.plugins
}

func (m *MockPluginManager) CallPlugin(plugin *types.Plugin, method string, params any) (any, error) {
	m.calls = append(m.calls, struct {
		name   string
		method string
		params any
	}{name: plugin.Name, method: method, params: params})

	key := plugin.Name + "." + method
	if err, ok := m.errors[key]; ok {
		return nil, err
	}

	if resp, ok := m.responses[key]; ok {
		return resp, nil
	}

	return map[string]any{"status": "ok"}, nil
}
func (m *MockPluginManager) AddPlugin(name, pluginType string) {
	m.plugins = append(m.plugins, &types.Plugin{
		Name: name,
		Type: types.PluginTypeFromString(pluginType),
	})
}

func (m *MockPluginManager) SetResponse(pluginName, method string, response any) {
	key := pluginName + "." + method
	m.responses[key] = response
}

func (m *MockPluginManager) SetError(pluginName, method string, err error) {
	key := pluginName + "." + method
	m.errors[key] = err
}

func (m *MockPluginManager) GetCalls() []struct {
	name   string
	method string
	params any
} {
	return m.calls
}

func (m *MockPluginManager) ResetCalls() {
	m.calls = nil
}

type MockLLMPlugin struct {
	responses []string
	callIndex int
}

func NewMockLLMPlugin(responses []string) *MockLLMPlugin {
	return &MockLLMPlugin{
		responses: responses,
		callIndex: 0,
	}
}

func (m *MockLLMPlugin) GetNextResponse() string {
	if m.callIndex >= len(m.responses) {
		return "I don't know how to help with that."
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return resp
}

func TestReActLoop_SingleToolCall(t *testing.T) {
	tests := []struct {
		name              string
		userMessage       string
		llmResponses      []string
		toolName          string
		toolResponse      any
		expectedToolCalls int
		validateResponse  func(t *testing.T, response string)
	}{
		{
			name:        "single_shell_tool_call",
			userMessage: "列出当前目录的文件",
			llmResponses: []string{
				`我来帮你列出当前目录的文件。

<invoke>
<name>shell</name>
<params>{"command": "ls -la"}</params>
</invoke>`,
				`当前目录包含以下内容：
- Documents (文件夹)
- Downloads (文件夹)
- Desktop (文件夹)
- .bashrc (配置文件)
- .zshrc (配置文件)`,
			},
			toolName:          "shell",
			toolResponse:      map[string]any{"output": "total 128\ndrwxr-xr-x  20 user  staff   640 Mar  1 10:00 .\n-rw-r--r--   1 user  staff  1024 Mar  1 09:00 .bashrc"},
			expectedToolCalls: 1,
			validateResponse: func(t *testing.T, response string) {
				if !strings.Contains(response, "Documents") && !strings.Contains(response, "文件") {
					t.Errorf("Response should mention files or directories, got: %s", response)
				}
			},
		},
		{
			name:        "file_read_tool_call",
			userMessage: "读取 /etc/hosts 文件的内容",
			llmResponses: []string{
				`我需要读取 /etc/hosts 文件来查看主机名映射。

<invoke>
<name>file</name>
<params>{"action": "read", "path": "/etc/hosts"}</params>
</invoke>`,
				`/etc/hosts 文件内容如下：

127.0.0.1       localhost
255.255.255.255 broadcasthost
::1             localhost

这是系统的主机名映射文件。`,
			},
			toolName:          "file",
			toolResponse:      map[string]any{"content": "127.0.0.1       localhost\n255.255.255.255 broadcasthost\n::1             localhost"},
			expectedToolCalls: 1,
			validateResponse: func(t *testing.T, response string) {
				if !strings.Contains(response, "localhost") {
					t.Errorf("Response should contain 'localhost', got: %s", response)
				}
			},
		},
		{
			name:        "no_tool_call_direct_response",
			userMessage: "你好，今天天气怎么样？",
			llmResponses: []string{
				`你好！我是一个 AI 助手，无法获取实时天气信息。建议你查看天气预报应用或网站来了解今天的天气情况。`,
			},
			toolName:          "",
			toolResponse:      nil,
			expectedToolCalls: 0,
			validateResponse: func(t *testing.T, response string) {
				if !strings.Contains(response, "你好") && !strings.Contains(response, "天气") {
					t.Errorf("Response should be a greeting about weather, got: %s", response)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := NewMockPluginManager()

			mockPM.AddPlugin("mock-llm", "llm")

			if tt.toolName != "" {
				mockPM.AddPlugin(tt.toolName, "tool")
				if tt.toolResponse != nil {
					mockPM.SetResponse(tt.toolName, "call", tt.toolResponse)
				}
			}

			engine := NewEngineWithMock(mockPM)
			engine.SetMaxIterations(10)

			ctx := context.Background()
			sessionID := "test-react-" + tt.name
			engine.CreateSession(sessionID)

			responseIndex := 0
			engine.SetLLMCallback(func(ctx context.Context, session *Session, systemPrompt string) (string, error) {
				if responseIndex >= len(tt.llmResponses) {
					return "I cannot help with that.", nil
				}
				resp := tt.llmResponses[responseIndex]
				responseIndex++
				return resp, nil
			})

			response, err := engine.ProcessMessage(ctx, sessionID, tt.userMessage)
			if err != nil {
				t.Fatalf("ProcessMessage() error = %v", err)
			}

			calls := mockPM.GetCalls()
			toolCalls := 0
			for _, call := range calls {
				if call.method == "call" {
					toolCalls++
				}
			}

			if toolCalls != tt.expectedToolCalls {
				t.Errorf("Expected %d tool calls, got %d", tt.expectedToolCalls, toolCalls)
			}

			if tt.validateResponse != nil {
				tt.validateResponse(t, response)
			}

			t.Logf("Test %s passed: toolCalls=%d, response=%s", tt.name, toolCalls, response)
		})
	}
}

func TestReActLoop_MultipleToolCalls(t *testing.T) {
	tests := []struct {
		name              string
		userMessage       string
		llmResponses      []string
		toolResponses     map[string]map[string]any
		expectedToolCalls []string
	}{
		{
			name:        "two_sequential_tools",
			userMessage: "找出所有 .go 文件并统计代码行数",
			llmResponses: []string{
				`我来帮你找出所有 .go 文件。

<invoke>
<name>shell</name>
<params>{"command": "find . -name '*.go'"}</params>
</invoke>`,
				`找到了 3 个 Go 文件，现在统计代码行数。

<invoke>
<name>shell</name>
<params>{"command": "find . -name '*.go' | xargs wc -l"}</params>
</invoke>`,
				`统计完成！

找到 3 个 Go 文件，总共 450 行代码：
- main.go: 150 行
- utils.go: 200 行  
- handler.go: 100 行`,
			},
			toolResponses: map[string]map[string]any{
				"shell.call": {"output": "./main.go\n./utils.go\n./handler.go"},
			},
			expectedToolCalls: []string{"shell", "shell"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := NewMockPluginManager()
			mockPM.AddPlugin("mock-llm", "llm")
			mockPM.AddPlugin("shell", "tool")

			for key, resp := range tt.toolResponses {
				mockPM.SetResponse("shell", "call", resp)
				_ = key
			}

			engine := NewEngineWithMock(mockPM)
			engine.SetMaxIterations(10)

			ctx := context.Background()
			sessionID := "test-react-multi-" + tt.name
			engine.CreateSession(sessionID)

			responseIndex := 0
			engine.SetLLMCallback(func(ctx context.Context, session *Session, systemPrompt string) (string, error) {
				if responseIndex >= len(tt.llmResponses) {
					return "Task completed.", nil
				}
				resp := tt.llmResponses[responseIndex]
				responseIndex++
				return resp, nil
			})

			response, err := engine.ProcessMessage(ctx, sessionID, tt.userMessage)
			if err != nil {
				t.Fatalf("ProcessMessage() error = %v", err)
			}

			calls := mockPM.GetCalls()
			actualToolCalls := make([]string, 0)
			for _, call := range calls {
				if call.method == "call" {
					actualToolCalls = append(actualToolCalls, call.name)
				}
			}

			if len(actualToolCalls) != len(tt.expectedToolCalls) {
				t.Errorf("Expected %d tool calls, got %d: %v",
					len(tt.expectedToolCalls), len(actualToolCalls), actualToolCalls)
			}

			for i, expected := range tt.expectedToolCalls {
				if i >= len(actualToolCalls) || actualToolCalls[i] != expected {
					t.Errorf("Tool call %d: expected %s, got %s",
						i, expected, actualToolCalls[i])
				}
			}

			t.Logf("Test %s passed: toolCalls=%v, response=%s", tt.name, actualToolCalls, response)
		})
	}
}

func TestReActLoop_MaxIterations(t *testing.T) {
	mockPM := NewMockPluginManager()
	mockPM.AddPlugin("mock-llm", "llm")
	mockPM.AddPlugin("shell", "tool")

	engine := NewEngineWithMock(mockPM)
	engine.SetMaxIterations(3)

	ctx := context.Background()
	sessionID := "test-max-iterations"
	engine.CreateSession(sessionID)

	callCount := 0
	engine.SetLLMCallback(func(ctx context.Context, session *Session, systemPrompt string) (string, error) {
		callCount++
		return fmt.Sprintf(`第 %d 次调用工具。

<invoke>
<name>shell</name>
<params>{"command": "echo test"}</params>
</invoke>`, callCount), nil
	})

	_, err := engine.ProcessMessage(ctx, sessionID, "无限循环测试")

	if err == nil {
		t.Error("Expected error when exceeding max iterations")
	}

	if !strings.Contains(err.Error(), "maximum iterations") {
		t.Errorf("Expected max iterations error, got: %v", err)
	}

	if callCount > 3 {
		t.Errorf("Should stop after max iterations, but called LLM %d times", callCount)
	}

	t.Logf("Test passed: stopped after %d iterations with error: %v", callCount, err)
}

func TestReActLoop_ToolExecutionError(t *testing.T) {
	tests := []struct {
		name         string
		userMessage  string
		llmResponses []string
		toolError    error
		validate     func(t *testing.T, response string, err error)
	}{
		{
			name:        "tool_execution_failure",
			userMessage: "删除系统文件",
			llmResponses: []string{
				`我需要执行删除命令。

<invoke>
<name>shell</name>
<params>{"command": "rm -rf /important"}</params>
</invoke>`,
				`抱歉，删除文件失败了。错误信息：permission denied。

我不能删除系统文件，这可能需要管理员权限，而且删除系统文件是危险的操作。`,
			},
			toolError: fmt.Errorf("permission denied"),
			validate: func(t *testing.T, response string, err error) {
				if err != nil {
					t.Errorf("Should handle tool error gracefully, got: %v", err)
				}
				if !strings.Contains(response, "失败") && !strings.Contains(response, "错误") {
					t.Errorf("Response should mention the failure, got: %s", response)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPM := NewMockPluginManager()
			mockPM.AddPlugin("mock-llm", "llm")
			mockPM.AddPlugin("shell", "tool")
			mockPM.SetError("shell", "call", tt.toolError)

			engine := NewEngineWithMock(mockPM)
			engine.SetMaxIterations(10)

			ctx := context.Background()
			sessionID := "test-tool-error-" + tt.name
			engine.CreateSession(sessionID)

			responseIndex := 0
			engine.SetLLMCallback(func(ctx context.Context, session *Session, systemPrompt string) (string, error) {
				if responseIndex >= len(tt.llmResponses) {
					return "Error handled.", nil
				}
				resp := tt.llmResponses[responseIndex]
				responseIndex++
				return resp, nil
			})

			response, err := engine.ProcessMessage(ctx, sessionID, tt.userMessage)

			if tt.validate != nil {
				tt.validate(t, response, err)
			}

			t.Logf("Test %s passed: response=%s", tt.name, response)
		})
	}
}

func TestReActLoop_Timeout(t *testing.T) {
	mockPM := NewMockPluginManager()
	mockPM.AddPlugin("mock-llm", "llm")
	mockPM.AddPlugin("shell", "tool")

	engine := NewEngineWithMock(mockPM)
	engine.SetMaxIterations(10)
	engine.SetToolTimeout(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	sessionID := "test-timeout"
	engine.CreateSession(sessionID)

	engine.SetLLMCallback(func(ctx context.Context, session *Session, systemPrompt string) (string, error) {
		time.Sleep(200 * time.Millisecond)
		return "Delayed response", nil
	})

	_, err := engine.ProcessMessage(ctx, sessionID, "timeout test")

	if err == nil {
		t.Error("Expected timeout error")
	}

	t.Logf("Test passed: got expected timeout error: %v", err)
}
