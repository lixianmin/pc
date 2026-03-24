# BAML 集成设计文档

## 背景

当前 PC 项目中，LLM 调用和 Tool 执行都是通过独立插件进程实现的：

- LLM 插件（如 OpenAI）通过 stdio 协议与主进程通信
- Tool 插件（shell、file、http 等）也是独立进程

这种架构的**问题**：

1. 通信开销：JSON 序列化/反序列化
2. 资源消耗：每个插件都是独立进程
3. Prompt 管理分散：散落在 Go 代码中，难以维护
4. 调试复杂：多进程调试困难

## 目标

1. **统一 LLM 调用**：使用 BAML 管理所有 LLM 调用
2. **统一 Prompt 管理**：用 `.baml` 文件定义 Prompt、Agent 行为
3. **内置 Tool**：将 Tool 从独立进程迁移到主进程
4. **保留 Channel 插件**：Telegram 等 Channel 插件保持独立进程

## 架构变更

### Before（当前）

```
┌─────────────────────────────────────────────────────────┐
│ 主进程 (pc)                                              │
│  ┌─────────┐  ┌──────────────┐                          │
│  │ Engine  │  │ PluginManager│                          │
│  └────┬────┘  └──────┬───────┘                          │
└───────┼──────────────┼───────────────────────────────────┘
        │              │
        │ stdio        │ stdio
        ▼              ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│ LLM Plugin    │ │ Tool Plugins  │ │ Channel Plugin│
│ (独立进程)    │ │ (独立进程)    │ │ (独立进程)    │
└───────────────┘ └───────────────┘ └───────────────┘
```

### After（目标）

```
┌─────────────────────────────────────────────────────────┐
│ 主进程 (pc)                                              │
│  ┌─────────┐  ┌──────────────┐  ┌────────────────┐     │
│  │ Engine  │  │ ToolExecutor │  │ PluginManager  │     │
│  └────┬────┘  │ (内置)       │  │ (Channel only) │     │
│       │       └──────────────┘  └───────┬────────┘     │
│       │                                    │            │
│  ┌────▼────┐                               │ stdio      │
│  │baml_    │                               ▼            │
│  │client   │                        ┌───────────────┐   │
│  │(内置)   │                        │ Channel Plugin│   │
│  └─────────┘                        │ (独立进程)    │   │
│                                     └───────────────┘   │
└─────────────────────────────────────────────────────────┘
        │
        │ HTTP
        ▼
┌───────────────┐
│ LLM API       │
│ (外部服务)    │
└───────────────┘
```

## 文件结构变更

### 新增

```
pc/
├── baml_src/                    # BAML 源文件
│   ├── chat.baml                # 主对话 function
│   ├── react.baml               # ReAct 循环 function
│   ├── tools.baml               # Tool 调用 function（可选）
│   ├── clients.baml             # LLM client 配置
│   └── generator.baml           # Go 代码生成配置
│
├── baml_client/                 # 自动生成（不要手动编辑）
│   ├── client.go
│   ├── types.go
│   └── ...
│
├── internal/
│   ├── llm/                     # LLM 调用封装
│   │   ├── client.go            # BAML 客户端封装
│   │   └── client_test.go
│   │
│   └── tool/                    # 内置 Tool 实现
│       ├── executor.go          # Tool 执行器接口
│       ├── shell.go             # Shell 工具
│       ├── file.go              # 文件读写工具
│       ├── http.go              # HTTP 请求工具
│       ├── search.go            # 搜索工具
│       └── note.go              # 笔记工具
```

### 删除

```
examples/plugins/llm/            # LLM 插件目录（整体删除）
examples/plugins/tool/           # Tool 插件目录（整体删除）
```

### 保留

```
examples/plugins/channel/        # Channel 插件（Telegram 等）
pkg/protocol/                    # Channel 仍需要协议定义
```

## BAML 文件示例

### clients.baml - LLM Client 配置

```baml
// OpenAI GPT-4
client<llm> gpt-4o {
  provider openai
  options {
    model "gpt-4o-mini"
    api_key env.OPENAI_API_KEY
  }
}

// Anthropic Claude
client<llm> claude {
  provider anthropic
  options {
    model "claude-3-5-sonnet-20241022"
    api_key env.ANTHROPIC_API_KEY
  }
}

// Ollama 本地模型
client<llm> ollama {
  provider ollama
  options {
    model "llama3"
    base_url "http://localhost:11434"
  }
}

// Fallback 策略
client<llm> primary {
  provider openai
  retry_policy {
    max_retries 3
  }
  fallback ["claude", "ollama"]
}
```

### chat.baml - 主对话 Function

```baml
function Chat(
  messages: Message[],
  system_prompt: string
) -> string {
  client "primary"
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

    {{#for msg in messages}}
    {{ _.role(msg.role) }}
    {{ msg.content }}
    {{/for}}
  "#
}

class Message {
  role string @description("user, assistant, or system")
  content string
}
```

### react.baml - ReAct 循环

```baml
class ToolCall {
  name string @description("Tool name to call")
  params map<string, any> @description("Parameters for the tool")
}

class ReActResponse {
  thinking string @description("Internal reasoning")
  tool_call ToolCall? @description("Optional tool call")
  response string @description("Final response to user")
}

function ReAct(
  user_message: string,
  system_prompt: string,
  available_tools: string
) -> ReActResponse {
  client "primary"
  prompt #"
    {{ _.role("system") }}
    {{ system_prompt }}

    Available tools:
    {{ available_tools }}

    {{ _.role("user") }}
    {{ user_message }}

    Think step by step. If you need to use a tool, output a tool_call.
    Otherwise, provide a response directly.
  "#
}
```

## 代码变更

### Engine 变更

```go
// internal/engine/engine.go

import (
    b "github.com/lixianmin/pc/baml_client"
    "github.com/lixianmin/pc/internal/tool"
)

type Engine struct {
    // 删除
    // pluginManager  *plugin.PluginManager
    // llmPlugin      *types.Plugin

    // 新增
    toolExecutor *tool.Executor

    // 保留
    sessions      map[string]*Session
    systemPrompt  string
    skillManager  *skill.SkillManager
}

func (my *Engine) callLLM(ctx context.Context, session *Session, systemPrompt string) (string, error) {
    // 构建 BAML 消息格式
    var messages []types.Message
    for _, msg := range session.Messages {
        messages = append(messages, types.Message{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }

    // 调用 BAML
    result, err := b.Chat(ctx, messages, systemPrompt)
    if err != nil {
        return "", fmt.Errorf("LLM call failed: %w", err)
    }
    return result, nil
}

func (my *Engine) executeToolCall(name string, params map[string]any) (any, error) {
    return my.toolExecutor.Execute(name, params)
}
```

### Tool Executor 实现

```go
// internal/tool/executor.go

type Executor struct {
    shell  *ShellTool
    file   *FileTool
    http   *HTTPTool
    search *SearchTool
    note   *NoteTool
}

func NewExecutor() *Executor {
    return &Executor{
        shell:  NewShellTool(DefaultShellConfig),
        file:   NewFileTool(DefaultFileConfig),
        http:   NewHTTPTool(DefaultHTTPConfig),
        search: NewSearchTool(),
        note:   NewNoteTool(),
    }
}

func (e *Executor) Execute(name string, params map[string]any) (any, error) {
    switch name {
    case "shell":
        return e.shell.Execute(params)
    case "file_read":
        return e.file.Read(params)
    case "file_write":
        return e.file.Write(params)
    case "http_get":
        return e.http.Get(params)
    case "http_post":
        return e.http.Post(params)
    case "search":
        return e.search.Execute(params)
    case "note":
        return e.note.Execute(params)
    default:
        return nil, fmt.Errorf("unknown tool: %s", name)
    }
}
```

## 迁移原则

### 每步可验证

每个迁移步骤必须：
1. **有明确的验证方法** - 如何确认这一步成功
2. **可独立验证** - 不依赖后续步骤
3. **可回滚** - 如果失败，可以安全回退

### 临时代码规范

迁移过程中允许写临时验证代码，但必须：
1. **标记清晰** - 使用 `// TODO(temp):` 或 `// TEMP:` 注释
2. **有删除计划** - 在哪个阶段删除
3. **最后清理** - 迁移完成后统一删除

### 验证检查点

| 阶段 | 验证命令 | 期望结果 |
|-----|---------|---------|
| 阶段 1.1 安装 BAML | `baml-cli version` | 显示版本号 |
| 阶段 1.2 创建 .baml | `baml-cli generate` | 生成 baml_client/ |
| 阶段 1.3 单元测试 | `go test ./internal/llm/` | 测试通过 |
| 阶段 1.4 集成测试 | `pc tui` + 发送消息 | 正常响应 |
| 阶段 2.1 Tool 迁移 | `go test ./internal/tool/` | 测试通过 |
| 阶段 2.2 ReAct 测试 | `pc tui` + 工具调用 | 工具执行成功 |
| 阶段 3 清理 | `make test` | 全部测试通过 |

---

## 迁移步骤

### 阶段 1：引入 BAML，替换 LLM 调用

#### 步骤 1.1：安装 BAML CLI

```bash
go install github.com/boundaryml/baml/baml-cli@latest
baml-cli init
```

**验证**：
```bash
baml-cli version
# 期望：显示版本号
```

**回滚**：
```bash
rm -rf baml_src/
go clean -cache
```

---

#### 步骤 1.2：创建 .baml 文件

创建 `baml_src/` 目录和基础文件：
- `clients.baml` - LLM client 配置
- `chat.baml` - 对话 function
- `generator.baml` - 代码生成配置

**验证**：
```bash
baml-cli generate
# 期望：生成 baml_client/ 目录，无错误
ls baml_client/
# 期望：看到 client.go, types.go 等文件
```

**回滚**：
```bash
rm -rf baml_src/ baml_client/
```

---

#### 步骤 1.3：创建 LLM 封装层

创建 `internal/llm/client.go`，封装 BAML 调用。

**验证**：
```bash
go test ./internal/llm/ -v
# 期望：测试通过（需要设置 OPENAI_API_KEY）
```

**临时代码**（标记 `// TEMP:`）：
```go
// TEMP: 验证 BAML 调用是否正常，阶段 3 删除
func TestBAMLConnection(t *testing.T) {
    ctx := context.Background()
    result, err := b.Chat(ctx, []types.Message{{Role: "user", Content: "Hello"}}, "You are helpful.")
    if err != nil {
        t.Fatalf("BAML call failed: %v", err)
    }
    t.Logf("Response: %s", result)
}
```

**回滚**：
```bash
rm -rf internal/llm/
git checkout internal/engine/engine.go
```

---

#### 步骤 1.4：Engine 集成 BAML

修改 `Engine.callLLM()` 使用 BAML，保留旧代码作为注释。

**验证**：
```bash
# 启动 gateway
./pc gateway start

# 启动 TUI 发送消息
./pc tui
# 输入：Hello
# 期望：收到正常响应

# 检查日志
tail -f ~/.pc/logs/pc.log
# 期望：看到 BAML 调用日志，无错误
```

**回滚**：
```bash
git checkout internal/engine/engine.go
git checkout internal/engine/stream.go
./pc gateway restart
```

---

#### 步骤 1.5：流式输出集成

修改 `Engine.callLLMStream()` 使用 BAML streaming。

**验证**：
```bash
./pc tui
# 输入：Tell me a long story
# 期望：看到流式输出，逐字显示
```

**回滚**：
```bash
git checkout internal/engine/stream.go
./pc gateway restart
```

---

#### 步骤 1.6：删除 LLM 插件

确认 BAML 工作正常后，删除 LLM 插件目录。

**验证**：
```bash
./pc tui
# 输入：Hello
# 期望：正常响应

make test
# 期望：所有测试通过
```

**回滚**：
```bash
git checkout examples/plugins/llm/
```

---

### 阶段 2：内置 Tool

#### 步骤 2.1：创建 Tool 目录结构

创建 `internal/tool/` 目录。

**验证**：
```bash
ls internal/tool/
# 期望：看到 executor.go, shell.go, file.go 等
```

---

#### 步骤 2.2：迁移 Shell Tool

从 `examples/plugins/tool/shell/` 迁移代码。

**验证**：
```bash
go test ./internal/tool/ -run TestShell -v
# 期望：测试通过
```

**临时代码**：
```go
// TEMP: 验证 shell 工具，阶段 3 删除
func TestShellExecute(t *testing.T) {
    shell := NewShellTool(DefaultShellConfig)
    result, err := shell.Execute(map[string]any{"command": "echo hello"})
    if err != nil {
        t.Fatalf("Shell execute failed: %v", err)
    }
    t.Logf("Result: %+v", result)
}
```

---

#### 步骤 2.3：迁移其他 Tool

依次迁移 file, http, search, note 工具。

**验证**：
```bash
go test ./internal/tool/ -v
# 期望：所有工具测试通过
```

---

#### 步骤 2.4：Engine 集成 Tool Executor

修改 `Engine.executeToolCall()` 使用内置 Executor。

**验证**：
```bash
./pc tui
# 输入：List files in current directory
# 期望：Agent 调用 shell 工具，返回结果
```

**临时代码**：
```go
// TEMP: 记录工具调用，阶段 3 删除
logo.Info("[TEMP] Tool called:", name, "params:", params)
```

---

#### 步骤 2.5：删除 Tool 插件

确认内置 Tool 工作正常后，删除 Tool 插件目录。

**验证**：
```bash
./pc tui
# 测试各种工具调用
make test
# 期望：所有测试通过
```

---

### 阶段 3：清理

#### 步骤 3.1：清理临时代码

删除所有 `// TEMP:` 标记的代码。

```bash
# 查找临时代码
grep -r "// TEMP:" internal/
grep -r "// TODO(temp):" internal/

# 确认后删除
```

**验证**：
```bash
# 确认没有临时代码
grep -r "// TEMP:" internal/ && echo "ERROR: temp code found" || echo "OK"
make test
```

---

#### 步骤 3.2：清理 PluginManager

移除 PluginManager 中 LLM/Tool 相关代码。

**验证**：
```bash
make test
./pc gateway restart
./pc tui
# 完整功能测试
```

---

#### 步骤 3.3：更新 Makefile

更新 Makefile 添加 BAML generate，移除 LLM/Tool 插件构建。

**验证**：
```bash
make clean
make all
# 期望：构建成功

./pc gateway restart
./pc tui
# 期望：功能正常
```

---

#### 步骤 3.4：更新文档

1. 更新 `notes/02.arch.md`
2. 创建 `docs/superpowers/archive/`
3. 移动归档文档

**验证**：
```bash
# 检查文档一致性
cat notes/02.arch.md | grep -i "baml"
# 期望：包含 BAML 相关描述
```

### 文档归档清单

以下文档在迁移完成后归档到 `docs/superpowers/archive/`：

**specs/ 归档**：
- `2026-03-16-architecture-review-design.md` - 已实施
- `2026-03-19-builtin-tools-design.md` - 被 BAML 设计取代

**plans/ 归档**：
- `2026-03-16-session-consolidation.md` - 已完成
- `2026-03-16-streaming-support.md` - 已完成
- `2026-03-19-builtin-tools-implementation.md` - 被 BAML 迁移取代
- `2026-03-19-m5-task-integration.md` - 已完成
- `2026-03-19-skill-integration.md` - 已完成

## 配置管理

### 纯 BAML 方式

LLM 配置完全由 `.baml` 文件 + 环境变量管理，不在 `config.yml` 中配置。

### 环境变量

```bash
# ~/.pc/.env 或系统环境变量
OPENAI_API_KEY=sk-xxx
ANTHROPIC_API_KEY=sk-xxx

# Ollama 本地模型（可选）
OLLAMA_BASE_URL=http://localhost:11434
```

### 切换 LLM Provider

修改 `.baml` 文件中的 client 引用：

```baml
// 在 chat.baml 中切换 client
function Chat(...) -> string {
  client "gpt-4o"    // 改为 "claude" 或 "ollama" 即可切换
  prompt #"... "#
}
```

### config.yml（保持不变）

```yaml
# Agent 配置
agent:
  name: PersonalClaw
  profession: 通用助手
  personality:
    - 友好
    - 专业

# Channel 插件
plugins:
  channel:
    telegram:
      enabled: true
```

## Makefile 变更

```makefile
.PHONY: generate build build-plugins test clean

# BAML 代码生成
generate:
	baml-cli generate

# 主程序构建（包含 BAML 生成）
build: generate
	go build -o bin/pc ./cmd/pc

# Channel 插件构建（LLM/Tool 已内置，不再需要单独构建）
build-plugins:
	@echo "Building channel plugins..."
	@mkdir -p $(HOME)/.pc/plugins/channel/telegram/bin
	@cd examples/plugins/channel/telegram/cmd/telegram-bot && \
		go build -o $(HOME)/.pc/plugins/channel/telegram/bin/telegram-bot .
	@echo "Channel plugins built."

# 完整构建：主程序 + 插件
all: generate build build-plugins

test: generate
	go test ./...

dev: generate
	go run ./cmd/pc

clean:
	rm -f bin/pc
	rm -rf baml_client/
```

**变更说明**：
- `make build` 自动执行 `baml-cli generate`
- `build-plugins` 只构建 Channel 插件
- `make all` 构建主程序和 Channel 插件
- LLM/Tool 插件构建步骤删除（已内置）

## 二期优化项

以下细节在迁移稳定后再调整：

### Tool 命名规范化

| 当前命名 | 建议命名 | 说明 |
|---------|---------|------|
| `shell` | `bash` | 更准确反映实际执行环境 |
| `file_read` | `read` | 简化命名 |
| `file_write` | `write` | 简化命名 |
| `http_get` | 待定 | 可能统一为 `http` |
| `http_post` | 待定 | 可能统一为 `http` |

> 注意：命名变更需要同步更新 BAML 中的工具定义和 ReAct 提示词。

---

## 风险与缓解

| 风险 | 缓解措施 |
|-----|---------|
| BAML 不稳定或有 bug | 分阶段迁移，每阶段可验证 |
| 性能不如预期 | BENCHMARK 测试，必要时回滚 |
| Tool 迁移遗漏功能 | 对照现有插件逐个验证 |
| Channel 插件受影响 | Channel 保持不变，独立测试 |

## 成功标准

1. 所有现有功能正常工作
2. `pc tui` 可以正常对话
3. ReAct 循环正常工作（工具调用）
4. Channel 插件（Telegram）正常工作
5. 测试覆盖率不低于当前

## 参考资料

- [BAML 官方文档](https://docs.boundaryml.com/)
- [BAML Go 安装指南](https://docs.boundaryml.com/guide/installation-language/go)
- [BAML Function 参考](https://docs.boundaryml.com/ref/baml/function)
