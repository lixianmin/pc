# BAML 集成实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 LLM 调用和 Tool 执行从插件进程迁移到主进程，使用 BAML 统一管理 LLM 调用和 Prompt

**Architecture:** 
- BAML 内置到主进程，替代 LLM 插件
- Tool 从独立进程迁移到 `internal/tool/` 内置
- Channel 插件保持独立进程

**Tech Stack:** Go 1.25+, BAML, github.com/boundaryml/baml

---

## 文件结构

### 新增文件

```
baml_src/
├── clients.baml              # LLM client 配置
├── chat.baml                 # 对话 function
├── generator.baml            # Go 代码生成配置

internal/llm/
├── client.go                 # BAML 客户端封装
├── client_test.go            # 单元测试

internal/tool/
├── executor.go               # Tool 执行器
├── executor_test.go          # 测试
├── shell.go                  # Shell 工具
├── shell_test.go             # 测试
├── file.go                   # 文件读写工具
├── file_test.go              # 测试
├── http.go                   # HTTP 请求工具
├── http_test.go              # 测试
├── search.go                 # 搜索工具
├── note.go                   # 笔记工具
```

### 修改文件

```
internal/engine/engine.go     # 使用 BAML 调用
internal/engine/stream.go     # 使用 BAML streaming
internal/plugin/plugin_manager.go  # 移除 LLM/Tool 相关
Makefile                      # 添加 BAML generate
```

### 删除文件

```
examples/plugins/llm/         # 阶段 1.6 删除
examples/plugins/tool/        # 阶段 2.5 删除
```

---

## Chunk 1: BAML 基础设施

### Task 1.1: 安装 BAML CLI

**Files:**
- 无文件变更

- [ ] **Step 1: 安装 BAML CLI**

```bash
go install github.com/boundaryml/baml/baml-cli@latest
```

- [ ] **Step 2: 验证安装**

Run: `baml-cli version`
Expected: 显示版本号（如 `baml-cli version 0.x.x`）

- [ ] **Step 3: 安装 Go 运行时依赖**

```bash
go get github.com/boundaryml/baml
```

- [ ] **Step 4: 安装 goimports（BAML 生成需要）**

```bash
go install golang.org/x/tools/cmd/goimports@latest
```

**验证**:
```bash
baml-cli version
go list -m github.com/boundaryml/baml
```

**回滚**:
```bash
go clean -cache
```

---

### Task 1.2: 创建 baml_src 目录结构

**Files:**
- Create: `baml_src/clients.baml`
- Create: `baml_src/chat.baml`
- Create: `baml_src/generator.baml`

- [ ] **Step 1: 创建 baml_src 目录**

```bash
mkdir -p baml_src
```

- [ ] **Step 2: 创建 clients.baml**

```baml
// baml_src/clients.baml
// LLM Client 配置

client<llm> gpt-4o {
  provider openai
  options {
    model "gpt-4o-mini"
    api_key env.OPENAI_API_KEY
  }
}

client<llm> claude {
  provider anthropic
  options {
    model "claude-3-5-sonnet-20241022"
    api_key env.ANTHROPIC_API_KEY
  }
}

client<llm> ollama {
  provider ollama
  options {
    model "llama3"
    base_url env.OLLAMA_BASE_URL
  }
}
```

- [ ] **Step 3: 创建 chat.baml**

```baml
// baml_src/chat.baml
// 主对话 function

class Message {
  role string @description("user, assistant, or system")
  content string
}

function Chat(
  messages: Message[],
  system_prompt: string
) -> string {
  client "gpt-4o"
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

    {{#for msg in messages}}
    {{ _.role(msg.role) }}
    {{ msg.content }}
    {{/for}}
  "#
}
```

- [ ] **Step 4: 创建 generator.baml**

```baml
// baml_src/generator.baml
// Go 代码生成配置

generator go {
  target go
  output_dir "../baml_client"
  version "0.203.1"
}
```

- [ ] **Step 5: 生成 baml_client**

```bash
baml-cli generate
```

**验证**:
```bash
ls baml_client/
# 期望: client.go, types.go 等文件
```

**回滚**:
```bash
rm -rf baml_src/ baml_client/
```

---

### Task 1.3: 创建 LLM 封装层

**Files:**
- Create: `internal/llm/client.go`
- Create: `internal/llm/client_test.go`

- [ ] **Step 1: 创建 internal/llm 目录**

```bash
mkdir -p internal/llm
```

- [ ] **Step 2: 写 client.go**

```go
// internal/llm/client.go
package llm

import (
	"context"
	"fmt"

	b "github.com/lixianmin/pc/baml_client"
)

type Message struct {
	Role    string
	Content string
}

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Chat(ctx context.Context, messages []Message, systemPrompt string) (string, error) {
	var bamlMessages []b.Message
	for _, m := range messages {
		bamlMessages = append(bamlMessages, b.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	result, err := b.Chat(ctx, bamlMessages, systemPrompt)
	if err != nil {
		return "", fmt.Errorf("BAML Chat failed: %w", err)
	}

	return result, nil
}
```

- [ ] **Step 3: 写 client_test.go**

```go
// internal/llm/client_test.go
package llm

import (
	"context"
	"os"
	"testing"
)

// TEMP: 验证 BAML 连接，阶段 3 删除
func TestBAMLConnection(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	client := NewClient()
	ctx := context.Background()

	result, err := client.Chat(ctx, []Message{
		{Role: "user", Content: "Say 'hello' in one word"},
	}, "You are a helpful assistant.")

	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}

	if result == "" {
		t.Fatal("Expected non-empty result")
	}

	t.Logf("Response: %s", result)
}
```

- [ ] **Step 4: 运行测试验证**

Run: `go test ./internal/llm/ -v -run TestBAMLConnection`
Expected: 测试通过（需要设置 OPENAI_API_KEY）

- [ ] **Step 5: 提交**

```bash
git add baml_src/ internal/llm/
git commit -m "feat(llm): add BAML integration with basic Chat function"
```

**验证**:
```bash
go test ./internal/llm/ -v
```

**回滚**:
```bash
rm -rf internal/llm/
git checkout baml_src/
```

---

## Chunk 2: Engine 集成 BAML

### Task 2.1: 修改 Engine 使用 BAML

**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: 读取当前 callLLM 方法**

Run: 查看当前 `internal/engine/engine.go` 中的 `callLLM` 方法
Expected: 理解当前通过 pluginManager.CallPlugin 调用

- [ ] **Step 2: 添加 LLM client 到 Engine 结构体**

在 `internal/engine/engine.go` 中修改：

```go
import (
    // 新增
    "github.com/lixianmin/pc/internal/llm"
)

type Engine struct {
    // 保留
    pluginManager  *plugin.PluginManager
    mockCaller     PluginCaller
    sessions       map[string]*Session
    mu             sync.RWMutex
    llmPlugin      *types.Plugin  // 保留，阶段 3 删除
    systemPrompt   string
    maxIterations  int
    toolTimeout    time.Duration
    llmCallback    llmCallback
    promptRecorder *debug.PromptRecorder
    taskManager    *task.Manager
    decomposer     *task.Decomposer
    taskEnabled    bool
    skillManager   *skill.SkillManager

    // 新增
    llmClient      *llm.Client
    useBAML        bool  // TEMP: 用于切换，阶段 3 删除
}
```

- [ ] **Step 3: 初始化 LLM client**

在 `NewEngine` 中添加：

```go
func NewEngine(pm *plugin.PluginManager) *Engine {
    return &Engine{
        pluginManager: pm,
        sessions:      make(map[string]*Session),
        maxIterations: 10,
        toolTimeout:   30 * time.Second,
        taskManager:   task.NewManager(""),
        decomposer:    task.NewDecomposer(),
        // 新增
        llmClient:     llm.NewClient(),
        useBAML:       false,  // TEMP: 默认使用插件，阶段 3 改为 true
    }
}
```

- [ ] **Step 4: 修改 callLLM 方法**

```go
func (my *Engine) callLLM(ctx context.Context, session *Session, message string) (string, error) {
    startTime := time.Now()

    systemPrompt := my.buildDynamicSystemPrompt()
    if my.systemPrompt != "" {
        systemPrompt = my.systemPrompt + "\n\n" + systemPrompt
    }

    // TEMP: 根据 useBAML 切换调用方式，阶段 3 删除 if 分支
    if my.useBAML {
        return my.callLLMViaBAML(ctx, session, systemPrompt, message)
    }
    // 保留原有插件调用逻辑
    return my.callLLMViaPlugin(ctx, session, systemPrompt, message)
}

// 新增: BAML 调用方法
func (my *Engine) callLLMViaBAML(ctx context.Context, session *Session, systemPrompt string, message string) (string, error) {
    var messages []llm.Message
    for _, msg := range session.Messages {
        messages = append(messages, llm.Message{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }
    messages = append(messages, llm.Message{
        Role:    "user",
        Content: message,
    })

    content, err := my.llmClient.Chat(ctx, messages, systemPrompt)
    if err != nil {
        return "", err
    }

    logo.Info("[Engine.callLLMViaBAML] LLM call completed, response length:", len(content))
    return content, nil
}

// 保留: 原有插件调用方法（重命名）
func (my *Engine) callLLMViaPlugin(ctx context.Context, session *Session, systemPrompt string, message string) (string, error) {
    // 原有 callLLM 的实现代码
    // ... 保持不变 ...
}
```

- [ ] **Step 5: 添加 SetUseBAML 方法**

```go
// TEMP: 用于测试切换，阶段 3 删除
func (my *Engine) SetUseBAML(useBAML bool) {
    my.useBAML = useBAML
}
```

- [ ] **Step 6: 编译验证**

Run: `go build ./...`
Expected: 编译成功

- [ ] **Step 7: 提交**

```bash
git add internal/engine/engine.go
git commit -m "feat(engine): add BAML LLM call support with toggle"
```

**验证**:
```bash
go build ./...
go test ./internal/engine/ -v -run TestEngine
```

**回滚**:
```bash
git checkout internal/engine/engine.go
```

---

### Task 2.2: 添加 BAML 集成测试

**Files:**
- Create: `internal/engine/baml_integration_test.go`

- [ ] **Step 1: 创建集成测试文件**

```go
// internal/engine/baml_integration_test.go
//go:build integration
// +build integration

package engine

import (
	"context"
	"os"
	"testing"

	"github.com/lixianmin/pc/internal/llm"
)

// TEMP: 验证 Engine BAML 集成，阶段 3 删除
func TestEngineWithBAML(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}

	engine := NewEngine(nil)
	engine.SetUseBAML(true)
	engine.SetSystemPrompt("You are a helpful assistant.")

	ctx := context.Background()
	session := engine.GetOrCreateSession("test-session")

	response, err := engine.ProcessMessage(ctx, "test-session", "Say hello")
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if response == "" {
		t.Fatal("Expected non-empty response")
	}

	t.Logf("Response: %s", response)
}
```

- [ ] **Step 2: 运行集成测试**

Run: `go test ./internal/engine/ -v -tags=integration -run TestEngineWithBAML`
Expected: 测试通过

- [ ] **Step 3: 提交**

```bash
git add internal/engine/baml_integration_test.go
git commit -m "test(engine): add BAML integration test"
```

---

### Task 2.3: 添加 Streaming 支持

**Files:**
- Modify: `internal/engine/stream.go`

- [ ] **Step 1: 读取当前 stream.go**

查看 `internal/engine/stream.go` 中的 `callLLMStream` 方法

- [ ] **Step 2: 更新 chat.baml 添加 stream function**

在 `baml_src/chat.baml` 中添加：

```baml
function StreamChat(
  messages: Message[],
  system_prompt: string
) -> string {
  client "gpt-4o"
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

    {{#for msg in messages}}
    {{ _.role(msg.role) }}
    {{ msg.content }}
    {{/for}}
  "#
}
```

- [ ] **Step 3: 重新生成 baml_client**

```bash
baml-cli generate
```

- [ ] **Step 4: 更新 internal/llm/client.go**

添加 Stream 方法：

```go
func (c *Client) StreamChat(ctx context.Context, messages []Message, systemPrompt string) (<-chan StreamChunk, error) {
    var bamlMessages []b.Message
    for _, m := range messages {
        bamlMessages = append(bamlMessages, b.Message{
            Role:    m.Role,
            Content: m.Content,
        })
    }

    stream, err := b.StreamChat(ctx, bamlMessages, systemPrompt)
    if err != nil {
        return nil, fmt.Errorf("BAML StreamChat failed: %w", err)
    }

    ch := make(chan StreamChunk)
    go func() {
        defer close(ch)
        for chunk := range stream {
            if chunk.IsError {
                ch <- StreamChunk{Error: chunk.Error}
                return
            }
            if chunk.Stream() != nil {
                ch <- StreamChunk{Content: *chunk.Stream()}
            }
        }
    }()

    return ch, nil
}

type StreamChunk struct {
    Content string
    Error   error
}
```

- [ ] **Step 5: 修改 stream.go**

添加 BAML streaming 支持（保留原有逻辑）

- [ ] **Step 6: 编译验证**

Run: `go build ./...`

- [ ] **Step 7: 提交**

```bash
git add baml_src/ internal/llm/ internal/engine/stream.go
git commit -m "feat(engine): add BAML streaming support"
```

---

### Task 2.4: Gateway 集成 BAML

**Files:**
- Modify: `internal/gateway/gateway.go`

- [ ] **Step 1: 在 Gateway 中启用 BAML**

在 `internal/gateway/gateway.go` 的 Engine 初始化处添加：

```go
engine.SetUseBAML(true)  // TEMP: 阶段 3 移除此行，useBAML 默认为 true
```

- [ ] **Step 2: 重新编译并测试**

```bash
make build
./pc gateway restart
./pc tui
# 输入: Hello
# 期望: 收到正常响应
```

- [ ] **Step 3: 提交**

```bash
git add internal/gateway/gateway.go
git commit -m "feat(gateway): enable BAML for LLM calls"
```

**验证**:
```bash
./pc gateway restart
./pc tui
# 发送消息测试
```

**回滚**:
```bash
git checkout internal/gateway/gateway.go
./pc gateway restart
```

---

### Task 2.5: 删除 LLM 插件

**Files:**
- Delete: `examples/plugins/llm/`

- [ ] **Step 1: 确认 BAML 工作正常**

```bash
./pc tui
# 发送多条消息测试
# 测试流式输出
# 测试工具调用
```

- [ ] **Step 2: 删除 LLM 插件目录**

```bash
rm -rf examples/plugins/llm/
```

- [ ] **Step 3: 清理 Engine 中的 llmPlugin 字段**

从 `internal/engine/engine.go` 中删除：
```go
llmPlugin      *types.Plugin  // 删除此行
```

删除 `callLLMViaPlugin` 方法

删除 `useBAML` 字段，让 BAML 成为唯一调用方式

- [ ] **Step 4: 运行测试**

Run: `make test`
Expected: 所有测试通过

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "refactor: remove LLM plugin, use BAML as primary LLM provider"
```

**验证**:
```bash
make test
./pc gateway restart
./pc tui
```

**回滚**:
```bash
git checkout examples/plugins/llm/
git checkout internal/engine/engine.go
```

---

## Chunk 3: Tool 内置迁移

### Task 3.1: 创建 Tool 目录结构

**Files:**
- Create: `internal/tool/executor.go`
- Create: `internal/tool/executor_test.go`

- [ ] **Step 1: 创建目录**

```bash
mkdir -p internal/tool
```

- [ ] **Step 2: 写 executor.go 接口**

```go
// internal/tool/executor.go
package tool

import (
	"context"
	"fmt"
	"sync"
)

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]any) (any, error)
}

type Executor struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewExecutor() *Executor {
	return &Executor{
		tools: make(map[string]Tool),
	}
}

func (e *Executor) Register(tool Tool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.tools[tool.Name()] = tool
}

func (e *Executor) Execute(ctx context.Context, name string, params map[string]any) (any, error) {
	e.mu.RLock()
	tool, ok := e.tools[name]
	e.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", name)
	}

	return tool.Execute(ctx, params)
}

func (e *Executor) ListTools() []ToolInfo {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var tools []ToolInfo
	for name, tool := range e.tools {
		tools = append(tools, ToolInfo{
			Name:        name,
			Description: tool.Description(),
		})
	}
	return tools
}

type ToolInfo struct {
	Name        string
	Description string
}
```

- [ ] **Step 3: 写 executor_test.go**

```go
// internal/tool/executor_test.go
package tool

import (
	"context"
	"testing"
)

type mockTool struct {
	name string
}

func (m *mockTool) Name() string { return m.name }
func (m *mockTool) Description() string { return "mock tool" }
func (m *mockTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	return "mock result", nil
}

func TestExecutor_Register(t *testing.T) {
	e := NewExecutor()
	tool := &mockTool{name: "test"}

	e.Register(tool)

	if len(e.tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(e.tools))
	}
}

func TestExecutor_Execute(t *testing.T) {
	e := NewExecutor()
	tool := &mockTool{name: "test"}
	e.Register(tool)

	result, err := e.Execute(context.Background(), "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "mock result" {
		t.Errorf("expected 'mock result', got %v", result)
	}
}

func TestExecutor_Execute_UnknownTool(t *testing.T) {
	e := NewExecutor()

	_, err := e.Execute(context.Background(), "unknown", nil)
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}
```

- [ ] **Step 4: 运行测试**

Run: `go test ./internal/tool/ -v`
Expected: 测试通过

- [ ] **Step 5: 提交**

```bash
git add internal/tool/
git commit -m "feat(tool): add ToolExecutor interface"
```

---

### Task 3.2: 迁移 Shell Tool

**Files:**
- Create: `internal/tool/shell.go`
- Create: `internal/tool/shell_test.go`

- [ ] **Step 1: 写 shell.go**

从 `examples/plugins/tool/shell/shell.go` 迁移代码：

```go
// internal/tool/shell.go
package tool

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type ShellConfig struct {
	Timeout          time.Duration
	ConfirmDangerous bool
	AllowedCommands  []string
	BlockedCommands  []string
}

type ShellTool struct {
	config ShellConfig
}

func NewShellTool(config ShellConfig) *ShellTool {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &ShellTool{config: config}
}

func (t *ShellTool) Name() string {
	return "shell"
}

func (t *ShellTool) Description() string {
	return "Execute shell commands"
}

type ShellParams struct {
	Command    string `json:"command"`
	Timeout    int    `json:"timeout,omitempty"`
	WorkingDir string `json:"working_dir,omitempty"`
}

type ShellResult struct {
	Stdout     string `json:"stdout"`
	Stderr     string `json:"stderr"`
	ExitCode   int    `json:"exit_code"`
	DurationMs int64  `json:"duration_ms"`
}

func (t *ShellTool) Execute(ctx context.Context, params map[string]any) (any, error) {
	command, _ := params["command"].(string)
	if command == "" {
		return nil, fmt.Errorf("command is required")
	}

	if err := t.validateCommand(command); err != nil {
		return nil, err
	}

	timeout := t.config.Timeout
	if timeoutMs, ok := params["timeout"].(int); ok && timeoutMs > 0 {
		timeout = time.Duration(timeoutMs) * time.Millisecond
	}

	workingDir, _ := params["working_dir"].(string)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := &ShellResult{
		Stdout:     stdout.String(),
		Stderr:     stderr.String(),
		DurationMs: duration.Milliseconds(),
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		return result, fmt.Errorf("command timed out after %v", timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
		return result, nil
	}

	result.ExitCode = 0
	return result, nil
}

func (t *ShellTool) validateCommand(command string) error {
	if len(t.config.AllowedCommands) > 0 {
		allowed := false
		for _, prefix := range t.config.AllowedCommands {
			if strings.HasPrefix(command, prefix) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("command not in allowed list")
		}
	}

	for _, blocked := range t.config.BlockedCommands {
		if strings.Contains(command, blocked) {
			return fmt.Errorf("command contains blocked pattern: %s", blocked)
		}
	}

	return nil
}
```

- [ ] **Step 2: 写 shell_test.go**

```go
// internal/tool/shell_test.go
package tool

import (
	"context"
	"testing"
)

func TestShellTool_Name(t *testing.T) {
	tool := NewShellTool(ShellConfig{})
	if tool.Name() != "shell" {
		t.Errorf("expected 'shell', got %s", tool.Name())
	}
}

func TestShellTool_Execute_Echo(t *testing.T) {
	tool := NewShellTool(ShellConfig{})
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "echo hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shellResult := result.(*ShellResult)
	if shellResult.Stdout != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", shellResult.Stdout)
	}
	if shellResult.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", shellResult.ExitCode)
	}
}

func TestShellTool_Execute_MissingCommand(t *testing.T) {
	tool := NewShellTool(ShellConfig{})
	_, err := tool.Execute(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing command")
	}
}

// TEMP: 验证工具工作，阶段 3 删除
func TestShellTool_Execute_Complex(t *testing.T) {
	tool := NewShellTool(ShellConfig{})
	result, err := tool.Execute(context.Background(), map[string]any{
		"command": "ls -la | head -5",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	shellResult := result.(*ShellResult)
	t.Logf("Stdout: %s", shellResult.Stdout)
	t.Logf("ExitCode: %d", shellResult.ExitCode)
}
```

- [ ] **Step 3: 运行测试**

Run: `go test ./internal/tool/ -v -run TestShell`
Expected: 测试通过

- [ ] **Step 4: 提交**

```bash
git add internal/tool/shell.go internal/tool/shell_test.go
git commit -m "feat(tool): migrate shell tool to builtin"
```

---

### Task 3.3: 迁移其他 Tool

**Files:**
- Create: `internal/tool/file.go`
- Create: `internal/tool/file_test.go`
- Create: `internal/tool/http.go`
- Create: `internal/tool/http_test.go`
- Create: `internal/tool/search.go`
- Create: `internal/tool/note.go`

- [ ] **Step 1: 迁移 file tool**

从 `examples/plugins/tool/file/file.go` 迁移到 `internal/tool/file.go`

- [ ] **Step 2: 迁移 http tool**

从 `examples/plugins/tool/http/http.go` 迁移到 `internal/tool/http.go`

- [ ] **Step 3: 迁移 search tool**

从 `examples/plugins/tool/search/search.go` 迁移到 `internal/tool/search.go`

- [ ] **Step 4: 迁移 note tool**

从 `examples/plugins/tool/note/note.go` 迁移到 `internal/tool/note.go`

- [ ] **Step 5: 运行所有 tool 测试**

Run: `go test ./internal/tool/ -v`
Expected: 所有测试通过

- [ ] **Step 6: 提交**

```bash
git add internal/tool/
git commit -m "feat(tool): migrate file, http, search, note tools to builtin"
```

---

### Task 3.4: Engine 集成 ToolExecutor

**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: 添加 ToolExecutor 到 Engine**

```go
import (
    "github.com/lixianmin/pc/internal/tool"
)

type Engine struct {
    // ... existing fields ...

    // 新增
    toolExecutor *tool.Executor
}

func NewEngine(pm *plugin.PluginManager) *Engine {
    executor := tool.NewExecutor()
    executor.Register(tool.NewShellTool(tool.ShellConfig{Timeout: 30 * time.Second}))
    executor.Register(tool.NewFileTool(tool.FileConfig{}))
    executor.Register(tool.NewHTTPTool(tool.HTTPConfig{}))
    executor.Register(tool.NewSearchTool())
    executor.Register(tool.NewNoteTool())

    return &Engine{
        // ... existing fields ...
        toolExecutor: executor,
    }
}
```

- [ ] **Step 2: 修改工具调用逻辑**

在 `internal/engine/react.go` 或相关位置修改工具调用：

```go
func (my *Engine) executeToolCall(toolName string, params map[string]any) (any, error) {
    // TEMP: 记录工具调用，阶段 3 删除
    logo.Info("[TEMP] Tool called:", toolName, "params:", params)

    return my.toolExecutor.Execute(context.Background(), toolName, params)
}
```

- [ ] **Step 3: 编译验证**

Run: `go build ./...`

- [ ] **Step 4: 提交**

```bash
git add internal/engine/engine.go
git commit -m "feat(engine): integrate builtin ToolExecutor"
```

---

### Task 3.5: 删除 Tool 插件

**Files:**
- Delete: `examples/plugins/tool/`

- [ ] **Step 1: 确认内置 Tool 工作正常**

```bash
./pc tui
# 测试工具调用
# 输入: List files in current directory
# 期望: Agent 调用 shell 工具，返回结果
```

- [ ] **Step 2: 删除 Tool 插件目录**

```bash
rm -rf examples/plugins/tool/
```

- [ ] **Step 3: 运行测试**

Run: `make test`

- [ ] **Step 4: 提交**

```bash
git add -A
git commit -m "refactor: remove tool plugins, use builtin tools"
```

---

## Chunk 4: 清理和文档

### Task 4.1: 清理临时代码

**Files:**
- Modify: `internal/llm/client_test.go`
- Modify: `internal/engine/engine.go`
- Modify: `internal/tool/shell_test.go`

- [ ] **Step 1: 查找临时代码**

```bash
grep -r "// TEMP:" internal/
grep -r "// TODO(temp):" internal/
```

- [ ] **Step 2: 删除临时代码**

删除所有标记为 `// TEMP:` 或 `// TODO(temp):` 的代码：
- `internal/llm/client_test.go` 中的 `TestBAMLConnection`
- `internal/engine/engine.go` 中的 `useBAML` 字段和 `SetUseBAML` 方法
- `internal/engine/engine.go` 中的 `callLLMViaPlugin` 方法
- `internal/engine/baml_integration_test.go` (整个文件)
- `internal/tool/shell_test.go` 中的 `TestShellTool_Execute_Complex`
- `internal/engine/react.go` 中的临时日志

- [ ] **Step 3: 删除 integration test 文件**

```bash
rm internal/engine/baml_integration_test.go
```

- [ ] **Step 4: 验证没有临时代码**

```bash
grep -r "// TEMP:" internal/ && echo "ERROR: temp code found" || echo "OK"
```

- [ ] **Step 5: 提交**

```bash
git add -A
git commit -m "chore: remove temporary code from BAML migration"
```

---

### Task 4.2: 清理 PluginManager

**Files:**
- Modify: `internal/plugin/plugin_manager.go`

- [ ] **Step 1: 移除 LLM/Tool 相关代码**

从 `PluginManager` 中移除：
- `llmTimeout` 字段
- `SetLLMTimeout` 方法
- LLM 和 Tool 类型的插件扫描逻辑

保留：
- Channel 插件相关代码

- [ ] **Step 2: 运行测试**

Run: `make test`

- [ ] **Step 3: 提交**

```bash
git add internal/plugin/
git commit -m "refactor(plugin): remove LLM/Tool from PluginManager"
```

---

### Task 4.3: 更新 Makefile

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: 更新 Makefile**

```makefile
.PHONY: all generate build build-plugins test clean run fmt vet lint help

BINARY_NAME=pc
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

all: fmt vet test build build-plugins

generate:
	@echo "Generating BAML client..."
	baml-cli generate

build: generate
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/pc

build-plugins:
	@echo "Building channel plugins..."
	@mkdir -p $(HOME)/.pc/plugins/channel/telegram/bin
	@cd examples/plugins/channel/telegram/cmd/telegram-bot && \
		go build -o $(HOME)/.pc/plugins/channel/telegram/bin/telegram-bot .
	@echo "Channel plugins built."

test: generate
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@rm -rf baml_client/

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

fmt:
	@echo "Formatting code..."
	go fmt ./...

vet:
	@echo "Vetting code..."
	go vet ./...

lint:
	@echo "Linting..."
	@golangci-lint run ./... || echo "golangci-lint not installed"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

help:
	@echo "Usage:"
	@echo "  make generate      - Generate BAML client"
	@echo "  make build         - Build the binary (includes generate)"
	@echo "  make build-plugins - Build channel plugins only"
	@echo "  make test          - Run tests"
	@echo "  make all           - fmt, vet, test, build, build-plugins"
```

- [ ] **Step 2: 测试构建**

```bash
make clean
make all
```

- [ ] **Step 3: 提交**

```bash
git add Makefile
git commit -m "chore: update Makefile for BAML integration"
```

---

### Task 4.4: 更新架构文档

**Files:**
- Modify: `notes/02.arch.md`

- [ ] **Step 1: 更新架构图**

在 `notes/02.arch.md` 中更新架构图，反映 BAML 集成后的架构

- [ ] **Step 2: 更新插件类型说明**

更新插件类型列表，说明 LLM 和 Tool 已内置

- [ ] **Step 3: 添加 BAML 相关说明**

添加 BAML 配置和使用说明

- [ ] **Step 4: 提交**

```bash
git add notes/02.arch.md
git commit -m "docs: update architecture for BAML integration"
```

---

### Task 4.5: 归档旧文档

**Files:**
- Create: `docs/superpowers/archive/`
- Move: 旧文档到 archive 目录

- [ ] **Step 1: 创建 archive 目录**

```bash
mkdir -p docs/superpowers/archive/specs
mkdir -p docs/superpowers/archive/plans
```

- [ ] **Step 2: 归档 specs**

```bash
mv docs/superpowers/specs/2026-03-16-architecture-review-design.md docs/superpowers/archive/specs/
mv docs/superpowers/specs/2026-03-19-builtin-tools-design.md docs/superpowers/archive/specs/
```

- [ ] **Step 3: 归档 plans**

```bash
mv docs/superpowers/plans/2026-03-16-session-consolidation.md docs/superpowers/archive/plans/
mv docs/superpowers/plans/2026-03-16-streaming-support.md docs/superpowers/archive/plans/
mv docs/superpowers/plans/2026-03-19-builtin-tools-implementation.md docs/superpowers/archive/plans/
mv docs/superpowers/plans/2026-03-19-m5-task-integration.md docs/superpowers/archive/plans/
mv docs/superpowers/plans/2026-03-19-skill-integration.md docs/superpowers/archive/plans/
```

- [ ] **Step 4: 提交**

```bash
git add docs/superpowers/
git commit -m "docs: archive completed design and plan documents"
```

---

### Task 4.6: 最终验证

**Files:**
- 无文件变更

- [ ] **Step 1: 完整构建测试**

```bash
make clean
make all
```

- [ ] **Step 2: 运行所有测试**

```bash
make test
```

- [ ] **Step 3: 功能测试**

```bash
./pc gateway restart
./pc tui
# 测试对话
# 测试工具调用
# 测试流式输出
```

- [ ] **Step 4: 检查临时代码**

```bash
grep -r "// TEMP:" internal/ && echo "ERROR: temp code found" || echo "OK"
```

- [ ] **Step 5: 最终提交**

```bash
git add -A
git commit -m "feat: complete BAML integration - LLM and Tool builtin migration"
```

---

## 成功标准

- [ ] `make test` 全部通过
- [ ] `pc tui` 可以正常对话
- [ ] ReAct 循环正常工作（工具调用）
- [ ] 流式输出正常
- [ ] Channel 插件（Telegram）正常工作
- [ ] 无临时代码残留
- [ ] 文档已更新

---

## 验证检查点汇总

| 检查点 | 命令 | 期望结果 |
|-------|------|---------|
| BAML 安装 | `baml-cli version` | 显示版本号 |
| BAML 生成 | `baml-cli generate` | 生成 baml_client/ |
| LLM 单元测试 | `go test ./internal/llm/ -v` | 测试通过 |
| Tool 单元测试 | `go test ./internal/tool/ -v` | 测试通过 |
| Engine 测试 | `go test ./internal/engine/ -v` | 测试通过 |
| 完整测试 | `make test` | 全部通过 |
| 构建 | `make build` | 构建成功 |
| TUI 对话 | `./pc tui` | 正常响应 |
| 工具调用 | TUI 中输入 "list files" | 返回文件列表 |
| 无临时代码 | `grep -r "// TEMP:" internal/` | 无输出 |
