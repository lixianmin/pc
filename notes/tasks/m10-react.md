---
milestone: M10
title: ReAct 工具调用循环
status: 🔄 进行中
created: 2026-03-02
---

# M10: ReAct 工具调用循环

> 实现 Agent 的 ReAct (Reasoning + Acting) 循环，使 LLM 能够主动调用工具完成任务

---

## 需求映射

| 需求编号 | 需求描述 | 优先级 |
|---------|---------|--------|
| ACD-007 | Agent 支持 ReAct 循环：分析需求→调用工具→整合结果→生成回复 | P0 |
| ACD-008 | Agent 能够识别何时需要使用工具完成任务 | P0 |
| ACD-009 | Agent 能够解析工具执行结果并继续下一步决策 | P0 |
| TLP-007 | Tool 插件调用结果自动返回给 LLM 进行后续处理 | P0 |
| TLP-008 | 支持多轮工具调用直到任务完成 | P1 |

---

## M10-001: System Prompt 工具使用指南 ✅ 已完成

**优先级**: P0 | **需求**: ACD-008
**架构映射**: `internal/agent/system_prompt.go`
**完成时间**: 2026-03-02

### 实现步骤

1. ✅ 在 `SystemPromptBuilder` 中新增工具使用指南模板
2. ✅ 定义工具描述格式（名称、描述、参数 schema）
3. ✅ 添加 XML 格式工具调用示例
4. ✅ 添加 ReAct 工作流程说明
5. ✅ 编写单元测试验证生成的 prompt 格式

### 实现细节

**新增 `ParamsSchema` 字段到 `ToolInfo`：**
```go
type ToolInfo struct {
    Name         string
    Description  string
    Type         string
    ParamsSchema map[string]string  // 新增：参数名 -> 描述
}
```

**新增 `buildToolUsageGuide()` 方法：**
- 生成完整的工具使用指南
- 包含可用工具列表（带参数说明）
- 包含 XML 格式工具调用示例
- 包含 ReAct 工作流程说明

**生成的 System Prompt 结构：**
```
{basePrompt}

## 可用技能
- skill_name: description

## 工具使用指南

当你需要获取外部信息或执行操作时...

### 可用工具
- **tool_name**: description
    - `param`: description

### 工具调用格式
<tool_call>
<name>工具名</name>
<params>{"key": "value"}</params>
</tool_call>

### 示例
...

### 工作流程 (ReAct)
1. **思考 (Think)**: ...
2. **行动 (Act)**: ...
3. **观察 (Observe)**: ...
4. **回复 (Respond)**: ...
```

### 验收标准

- [x] System prompt 包含完整的工具使用指南
- [x] 工具列表动态从 plugin manager 获取
- [x] 生成的 prompt 包含 XML 格式示例
- [x] 单元测试验证 prompt 格式正确（新增 2 个测试函数）

---

## M10-002: 工具调用解析器 ✅ 已完成

**优先级**: P0 | **需求**: ACD-009
**架构映射**: `internal/engine/tool_parser.go` (新建)
**完成时间**: 2026-03-02

### 实现步骤

1. ✅ 创建 `tool_parser.go` 文件
2. ✅ 定义 `ToolCall` 和 `ToolResult` 结构体
3. ✅ 实现 `ParseToolCalls(content string) ([]ToolCall, error)` 函数
4. ✅ 使用正则表达式提取 `<tool_call>` 标签
5. ✅ 解析 `<name>` 和 `<params>` 内容
6. ✅ 处理 JSON 参数解析
7. ✅ 编写表格驱动测试

### 实现的功能

**`tool_parser.go` 包含：**

```go
// ToolCall 表示解析后的工具调用
type ToolCall struct {
    Name   string
    Params map[string]interface{}
}

// ToolResult 表示工具执行结果
type ToolResult struct {
    Name   string
    Output string
    Error  error
}
```

**主要函数：**
- `ParseToolCalls(content string) ([]ToolCall, error)` - 解析 LLM 输出中的工具调用
- `HasToolCalls(content string) bool` - 检查内容是否包含工具调用
- `FormatToolResult(result ToolResult) string` - 格式化工具结果为 XML
- `FormatToolResults(results []ToolResult) string` - 格式化多个工具结果

### 验收标准

- [x] 正确解析单个 `<tool_call>` 标签
- [x] 正确解析多个 `<tool_call>` 标签
- [x] 处理嵌套的 XML 内容
- [x] 处理格式错误的输入（返回原始字符串作为 `_raw` 参数）
- [x] 单元测试覆盖各种边界情况（10 个测试用例）

---

## M10-003: 工具执行器 ✅ 已完成

**优先级**: P0 | **需求**: TLP-007
**架构映射**: `internal/engine/tool_executor.go` (新建)
**完成时间**: 2026-03-02

### 实现步骤

1. ✅ 创建 `tool_executor.go` 文件
2. ✅ 定义 `PluginManager` 接口（支持 mock 测试）
3. ✅ 定义 `ToolExecutor` 结构体
4. ✅ 集成 `PluginManager` 调用 tool 插件
5. ✅ 实现超时控制（默认 30s）
6. ✅ 编写表格驱动测试

### 实现的功能

**`tool_executor.go` 包含：**

```go
// PluginManager 接口
type PluginManager interface {
    ListPlugins() []*types.Plugin
    CallPlugin(plugin *types.Plugin, method string, params any) (any, error)
}

// ToolExecutor 结构体
type ToolExecutor struct {
    pluginManager PluginManager
    timeout       time.Duration
}
```

**主要方法：**
- `Execute(ctx context.Context, call ToolCall) ToolResult` - 执行单个工具调用
- `ExecuteMultiple(ctx context.Context, calls []ToolCall) []ToolResult` - 批量执行工具调用
- `IsToolAvailable(name string) bool` - 检查工具是否可用
- `ListAvailableTools() []string` - 列出可用工具

### 工具执行流程

```
ToolCall
    │
    ▼
查找对应的 Tool 插件 (findToolPlugin)
    │
    ▼
调用 pluginManager.CallPlugin(plugin, "call", params)
    │
    ▼
格式化结果 (formatResult)
    │
    ▼
返回 ToolResult
```

### 验收标准

- [x] 能够执行已加载的 tool 插件
- [x] 正确处理工具执行超时（使用 context）
- [x] 正确处理工具执行错误
- [x] 返回格式化的执行结果（支持 output/stdout/content 字段）
- [x] 单元测试覆盖成功和失败场景（7 个测试函数）

---

## M10-004: Engine ReAct 循环改造

**优先级**: P0 | **需求**: ACD-007
**架构映射**: `internal/engine/engine.go`

### 实现步骤

1. 修改 `Engine` 结构体，添加依赖：
   ```go
   type Engine struct {
       pluginManager *plugin.PluginManager
       sessions      map[string]*Session
       llmPlugin     *types.Plugin
       systemPrompt  string
       maxIterations int  // 默认 10
       toolTimeout   time.Duration  // 默认 30s
   }
   ```

2. 重构 `ProcessMessage` 方法，实现 ReAct 循环：
   ```go
   func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
       // ... 验证和初始化 ...

       for i := 0; i < my.maxIterations; i++ {
           // 构建包含工具说明的 system prompt
           prompt := my.buildSystemPrompt()

           // 调用 LLM
           response, err := my.callLLMWithPrompt(ctx, session, prompt)
           if err != nil {
               return "", err
           }

           // 解析工具调用
           toolCalls, err := ParseToolCalls(response)
           if err != nil || len(toolCalls) == 0 {
               // 没有工具调用，返回最终回复
               return response, nil
           }

           // 执行工具
           var toolResults []ToolResult
           for _, call := range toolCalls {
               result, err := my.executeTool(call)
               toolResults = append(toolResults, ToolResult{
                   Name:   call.Name,
                   Output: result,
                   Error:  err,
               })
           }

           // 将工具调用和结果添加到会话历史
           session.AddMessage("assistant", response)
           session.AddMessage("system", formatToolResults(toolResults))
       }

       return "", fmt.Errorf("exceeded maximum iterations (%d)", my.maxIterations)
   }
   ```

3. 实现 `buildSystemPrompt()` 方法
4. 实现 `executeTool()` 方法
5. 实现 `formatToolResults()` 方法
6. 编写集成测试

### ReAct 循环流程

```
用户消息 → ProcessMessage
    │
    ▼
┌──────────────────────────────────┐
│         ReAct Loop               │
│  (max 10 iterations)             │
│                                  │
│  1. Build system prompt          │
│     with tool descriptions       │
│     │                            │
│     ▼                            │
│  2. Call LLM                     │
│     │                            │
│     ▼                            │
│  3. Parse response               │
│     ├─ Contains <tool_call>?     │
│     │   ├─ Yes → Execute tool    │
│     │   │      → Add result to   │
│     │   │        history         │
│     │   │      → Continue loop   │
│     │   │                        │
│     │   └─ No → Return response  │
│     │            to user         │
│     │                            │
│  4. (Loop back to step 1)        │
│                                  │
└──────────────────────────────────┘
```

### 验收标准

- [ ] ProcessMessage 实现 ReAct 循环
- [ ] 支持多轮工具调用
- [ ] 最大迭代次数保护
- [ ] 工具结果正确格式化为 system message
- [ ] 集成测试覆盖完整 ReAct 流程

---

## M10-005: Shell Tool 插件示例

**优先级**: P1 | **需求**: 示例实现
**架构映射**: `examples/plugins/tool/shell/`

### 实现步骤

1. 创建目录 `examples/plugins/tool/shell/`
2. 实现 shell 工具插件：
   - 接收 `command` 参数
   - 在本地执行 shell 命令
   - 返回 stdout/stderr
3. 添加 `plugin.yml` 元信息
4. 添加 `config.yml` 配置（如 allowed_commands）
5. 编写 README 说明

### 插件接口

```yaml
# plugin.yml
name: shell
type: tool
enabled: true
version: 1.0.0
entry: ./bin/shell-tool
```

### call 方法参数

```json
{
  "command": "ls -la ~",
  "timeout": 30,
  "working_dir": "/home/user"
}
```

### call 方法返回

```json
{
  "stdout": "total 128\ndrwxr-xr-x  20 user  staff   640 Mar  1 10:00 .\n...",
  "stderr": "",
  "exit_code": 0,
  "duration_ms": 150
}
```

### 验收标准

- [ ] Shell 插件可以独立运行
- [ ] 支持基本的 shell 命令执行
- [ ] 返回 stdout/stderr/exit_code
- [ ] 支持超时控制
- [ ] 示例配置可运行

---

## M10-006: 工具调用安全控制

**优先级**: P1 | **需求**: NFP-SEC-001
**架构映射**: `internal/engine/tool_executor.go`

### 实现步骤

1. 实现危险命令检测（如 `rm -rf /`）
2. 实现命令白名单/黑名单机制
3. 实现用户确认机制（可选）
4. 添加配置项 `tool.confirm_dangerous`
5. 编写安全测试

### 危险命令检测规则

```go
var dangerousPatterns = []string{
    `rm\s+-rf\s+/`,
    `>\s*/dev/sda`,
    `mkfs\.`,
    `dd\s+if=.*\s+of=/dev/`,
}
```

### 配置示例

```yaml
# ~/.pc/config.yml
tools:
  shell:
    confirm_dangerous: true
    allowed_commands: ["ls", "cat", "grep", "find"]
    timeout: 30s
```

### 验收标准

- [ ] 识别危险命令模式
- [ ] 可配置是否启用确认
- [ ] 可配置命令白名单
- [ ] 超时控制有效
- [ ] 安全测试覆盖

---

## M10-007: 集成测试

**优先级**: P1 | **需求**: NFP-MAINT-004
**架构映射**: `internal/engine/engine_test.go`

### 实现步骤

1. 编写完整的 ReAct 流程测试：
   - 用户请求需要工具
   - LLM 输出 tool_call
   - 执行工具
   - LLM 生成最终回复
2. 编写多轮工具调用测试
3. 编写工具执行失败测试
4. 编写最大迭代次数测试

### 测试场景

```go
// 场景 1: 单轮工具调用
{
    name: "list home directory",
    userMessage: "list files in my home directory",
    mockLLMResponses: []string{
        `<tool_call><name>shell</name><params>{"command":"ls ~"}</params></tool_call>`,
        `Your home directory contains: Desktop, Documents, Downloads...`,
    },
    expectedToolCalls: []string{"shell"},
    expectedFinalResponse: contains "Desktop",
}

// 场景 2: 多轮工具调用
{
    name: "multi-step task",
    userMessage: "find all Go files and count lines",
    mockLLMResponses: []string{
        `<tool_call><name>shell</name><params>{"command":"find . -name '*.go'"}</params></tool_call>`,
        `<tool_call><name>shell</name><params>{"command":"find . -name '*.go' | xargs wc -l"}</params></tool_call>`,
        `Found 42 Go files with total 1234 lines of code.`,
    },
    expectedToolCalls: []string{"shell", "shell"},
}
```

### 验收标准

- [ ] 单轮工具调用测试通过
- [ ] 多轮工具调用测试通过
- [ ] 工具执行失败测试通过
- [ ] 最大迭代次数保护测试通过
- [ ] 集成测试覆盖率 > 70%

---

## 任务清单汇总

| ID | 任务 | 优先级 | 状态 |
|----|-----|--------|------|
| M10-001 | System Prompt 工具使用指南 | P0 | ✅ 已完成 |
| M10-002 | 工具调用解析器 | P0 | ✅ 已完成 |
| M10-003 | 工具执行器 | P0 | ✅ 已完成 |
| M10-004 | Engine ReAct 循环改造 | P0 | ⏳ 待开始 |
| M10-005 | Shell Tool 插件示例 | P1 | ⏳ 待开始 |
| M10-006 | 工具调用安全控制 | P1 | ⏳ 待开始 |
| M10-007 | 集成测试 | P1 | ⏳ 待开始 |

---

## 技术决策记录

### 决策 1: 工具调用格式选择

**选择**: XML 标签格式 `<tool_call>...</tool_call>`

**理由**:
- 易于解析（正则或 XML 解析器）
- 与 LLM 自然语言输出清晰区分
- 被 Manus、Claude Code 等验证有效
- 比 JSON 格式更不容易与代码混淆

**替代方案**:
- JSON 格式: `{"tool": "name", "params": {}}`
- Function Calling 格式（OpenAI 风格）

### 决策 2: 工具调用结果格式

**选择**: 以 system message 形式返回给 LLM

**理由**:
- 符合对话式模型的消息格式
- 工具结果成为上下文的一部分
- 便于 LLM 理解 "Observation"

**格式**:
```
<tool_result>
<name>shell</name>
<output>command output here</output>
</tool_result>
```

### 决策 3: 最大迭代次数

**选择**: 默认 10 次

**理由**:
- 防止无限循环
- 足够处理大多数多步骤任务
- 可配置以适应不同场景
