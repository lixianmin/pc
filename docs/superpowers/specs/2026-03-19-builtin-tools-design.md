# 内置工具系统设计

## 概述

为 PersonalClaw 添加内置工具系统，使基础工具（bash, read, write, edit, glob, grep）无需安装插件即可使用。

## 问题背景

当前架构下，工具完全依赖插件实现。用户未安装 tool 插件时，LLM 无法执行任何工具操作，导致用户请求如"列出目录文件"无法完成。

## 设计目标

1. 基础工具开箱即用，无需安装插件
2. 与现有插件系统共存，优先级：内置 > 插件
3. 复用现有安全检查机制

## 架构设计

### 组件结构

```
internal/engine/
├── builtin_tool.go          # BuiltinTool 接口定义
├── engine.go                # Engine 集成（修改）
└── builtin/
    ├── registry.go          # 内置工具注册表
    ├── bash.go              # Shell 命令执行
    ├── read.go              # 文件读取
    ├── write.go             # 文件写入
    ├── edit.go              # 文件编辑
    ├── glob.go              # 文件名匹配
    └── grep.go              # 内容搜索
```

### 接口定义

```go
type ParamSchema struct {
    Type        string `json:"type"`         // "string", "number", "boolean"
    Required    bool   `json:"required"`
    Description string `json:"description"`
    Default     any    `json:"default,omitempty"`
}

type BuiltinTool interface {
    Name() string
    Description() string
    Parameters() map[string]ParamSchema
    Execute(ctx context.Context, params map[string]any) (string, error)
}

type ToolInfo struct {
    Name        string
    Description string
    Parameters  map[string]ParamSchema
}
```

### ToolExecutor 集成

修改 `tool_executor.go`:

```go
type ToolExecutor struct {
    // ... existing fields
    builtinRegistry *builtin.Registry
}

func (e *ToolExecutor) Execute(ctx context.Context, call ToolCall) ToolResult {
    // 1. 优先查找内置工具
    if builtin, ok := e.builtinRegistry.Get(call.Name); ok {
        return e.executeBuiltin(ctx, call, builtin)
    }
    
    // 2. 查找插件工具
    plugin, err := e.findToolPlugin(call.Name)
    // ... existing code
}

func (e *ToolExecutor) executeBuiltin(ctx context.Context, call ToolCall, tool builtin.BuiltinTool) ToolResult {
    params, err := e.validateParams(call.Name, call.Arguments, tool.Parameters())
    if err != nil {
        return ToolResult{CallId: call.Id, Error: err.Error()}
    }
    
    result, err := tool.Execute(ctx, params)
    if err != nil {
        return ToolResult{CallId: call.Id, Error: err.Error()}
    }
    return ToolResult{CallId: call.Id, Output: result}
}

func (e *ToolExecutor) validateParams(toolName string, args map[string]any, schema map[string]ParamSchema) (map[string]any, error) {
    result := make(map[string]any)
    
    for name, param := range schema {
        val, exists := args[name]
        if !exists {
            if param.Required {
                return nil, fmt.Errorf("missing required parameter: %s", name)
            }
            if param.Default != nil {
                val = param.Default
            } else {
                continue
            }
        }
        
        switch param.Type {
        case "string":
            if s, ok := val.(string); ok {
                result[name] = s
            } else {
                return nil, fmt.Errorf("parameter %s must be string", name)
            }
        case "number":
            switch v := val.(type) {
            case int:
                result[name] = v
            case float64:
                result[name] = int(v)
            default:
                return nil, fmt.Errorf("parameter %s must be number", name)
            }
        case "boolean":
            if b, ok := val.(bool); ok {
                result[name] = b
            } else {
                return nil, fmt.Errorf("parameter %s must be boolean", name)
            }
        }
    }
    
    return result, nil
}
```

### Engine 初始化

```go
type Engine struct {
    // ... existing fields
    builtinRegistry *builtin.Registry
    workDir         string
}

func NewEngine(pluginManager PluginManager) *Engine {
    e := &Engine{
        // ... existing fields
        builtinRegistry: builtin.NewRegistry(),
        workDir:         "",  // 默认空，由 SetWorkDir 设置
    }
    return e
}

func (my *Engine) SetWorkDir(dir string) {
    my.workDir = dir
    my.builtinRegistry.SetWorkDir(dir)
}

func (my *Engine) GetWorkDir() string {
    return my.workDir
}
```

### 工作目录解析

- 工作目录通过 `Engine.SetWorkDir()` 设置
- 默认值：启动 gateway 时的当前目录
- 来源：`run.go` 中调用 `engine.SetWorkDir(cfg.GetWorkDir())`
- 配置项：`config.yml` 新增 `work_dir` 字段

## 工具规范

### bash 工具

**名称**: `bash` (注意：SecurityChecker 中需新增 `bash` 支持)

**描述**: 执行 shell 命令

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `command` | string | 是 | - | 要执行的命令 |
| `workdir` | string | 否 | Engine.workDir | 工作目录 |
| `timeout` | number | 否 | 120 | 超时秒数 |

**安全检查**:
- 阻止危险命令: `rm -rf /`, `mkfs`, `dd if=`, `> /dev/`, `:(){ :|:& };:`
- 需要确认的命令: `rm -rf`, `chmod 777`, `chown`
- 更新 `security_checker.go`:
  ```go
  // 在 dangerousCommands 检查中支持 "bash" 工具名
  // 原有: toolName == "shell"
  // 新增: toolName == "shell" || toolName == "bash"
  ```

**错误处理**:
- 超时: 返回 "command timed out after {timeout}s"
- 命令不存在: 返回 "command not found: {cmd}"
- 非零退出码: 返回 "exit code {code}: {stderr}"

### read 工具

**描述**: 读取文件内容

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `path` | string | 是 | - | 文件绝对路径 |
| `offset` | number | 否 | 1 | 起始行号 |
| `limit` | number | 否 | 2000 | 最大行数 |

**安全检查**:
- 检查路径遍历: 拒绝 `../` 序列
- 检查符号链接: 不跟随指向目录外的符号链接

**错误处理**:
- 文件不存在: 返回 "file not found: {path}"
- 权限拒绝: 返回 "permission denied: {path}"
- 目录: 返回 "is a directory: {path}"
- 二进制文件: 返回 "binary file, cannot display"

### write 工具

**描述**: 写入文件（覆盖或创建）

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `path` | string | 是 | - | 文件绝对路径 |
| `content` | string | 是 | - | 文件内容 |

**安全检查**:
- 路径遍历检查
- 敏感路径保护: `/etc/passwd`, `/etc/shadow`, `~/.ssh/`, `~/.gnupg/`
- 只允许在工作目录及子目录下写入（可配置）

**错误处理**:
- 权限拒绝: 返回 "permission denied: {path}"
- 目录已存在: 返回 "is a directory: {path}"
- 磁盘满: 返回 "disk full or quota exceeded"

### edit 工具

**描述**: 编辑文件（字符串替换）

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `path` | string | 是 | - | 文件绝对路径 |
| `old_string` | string | 是 | - | 要替换的字符串 |
| `new_string` | string | 是 | - | 替换后的字符串 |
| `replace_all` | boolean | 否 | false | 替换所有匹配 |

**安全检查**: 同 write 工具

**错误处理**:
- 文件不存在: 返回 "file not found: {path}"
- 未找到匹配: 返回 "old_string not found in file"
- 多次匹配: 返回 "found multiple matches, use replace_all=true or provide more context"

### glob 工具

**描述**: 文件名模式匹配

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `pattern` | string | 是 | - | glob 模式，如 `**/*.go` |
| `path` | string | 否 | Engine.workDir | 搜索目录 |

**安全检查**:
- 路径遍历检查
- 结果集大小限制: 最多 10000 条结果

**错误处理**:
- 无效模式: 返回 "invalid glob pattern: {pattern}"
- 目录不存在: 返回 "directory not found: {path}"
- 结果过多: 返回前 10000 条，附加 "truncated, 10000+ matches"

### grep 工具

**描述**: 文件内容搜索

**参数**:
| 参数 | 类型 | 必需 | 默认值 | 描述 |
|------|------|------|--------|------|
| `pattern` | string | 是 | - | 正则表达式 |
| `path` | string | 否 | Engine.workDir | 搜索目录 |
| `include` | string | 否 | `*` | 文件名过滤 |

**安全检查**:
- 路径遍历检查
- ReDoS 保护: 正则编译超时 5 秒
- 结果集大小限制: 最多 2000 行

**错误处理**:
- 无效正则: 返回 "invalid regex: {pattern}"
- 目录不存在: 返回 "directory not found: {path}"
- 结果过多: 返回前 2000 行，附加 "truncated"

## Context 取消支持

所有工具必须尊重 context 取消：

```go
func (t *BashTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    cmd := exec.CommandContext(ctx, "sh", "-c", command)
    // ...
}
```

- bash: 使用 `exec.CommandContext`
- read/write/edit: 检查 `ctx.Done()`
- glob/grep: 使用 `filepath.Walk` 时检查 context

## Registry 实现

```go
type Registry struct {
    mu      sync.RWMutex
    tools   map[string]BuiltinTool
    workDir string
}

func NewRegistry() *Registry {
    r := &Registry{
        tools: make(map[string]BuiltinTool),
    }
    r.registerDefaults()
    return r
}

func (r *Registry) registerDefaults() {
    r.Register(&BashTool{})
    r.Register(&ReadTool{})
    r.Register(&WriteTool{})
    r.Register(&EditTool{})
    r.Register(&GlobTool{})
    r.Register(&GrepTool{})
}

func (r *Registry) Register(tool BuiltinTool) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.tools[tool.Name()] = tool
}

func (r *Registry) Get(name string) (BuiltinTool, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    tool, ok := r.tools[name]
    return tool, ok
}

func (r *Registry) Has(name string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    _, ok := r.tools[name]
    return ok
}

func (r *Registry) List() []ToolInfo {
    r.mu.RLock()
    defer r.mu.RUnlock()
    var result []ToolInfo
    for _, tool := range r.tools {
        result = append(result, ToolInfo{
            Name:        tool.Name(),
            Description: tool.Description(),
            Parameters:  tool.Parameters(),
        })
    }
    return result
}

func (r *Registry) SetWorkDir(dir string) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.workDir = dir
}
```

## 工具冲突处理

在 `ToolExecutor.findToolPlugin()` 中：

```go
func (e *ToolExecutor) findToolPlugin(name string) (*types.Plugin, error) {
    // 检查是否与内置工具冲突
    if e.builtinRegistry.Has(name) {
        logo.Warn("[ToolExecutor] Plugin tool '", name, "' conflicts with built-in tool, ignoring plugin")
        return nil, fmt.Errorf("tool '%s' exists as built-in", name)
    }
    // ... existing code
}
```

## buildToolGuide 集成

在 `engine.go` 中更新 `buildToolGuide()`:

```go
func (my *Engine) buildToolGuide() string {
    var sb strings.Builder
    
    // 1. 添加内置工具说明
    for _, tool := range my.builtinRegistry.List() {
        sb.WriteString(fmt.Sprintf("## %s\n%s\n\n", tool.Name, tool.Description))
        sb.WriteString("Parameters:\n")
        for paramName, param := range tool.Parameters {
            required := ""
            if param.Required {
                required = " (required)"
            }
            sb.WriteString(fmt.Sprintf("- %s: %s%s - %s\n", paramName, param.Type, required, param.Description))
        }
        sb.WriteString("\n")
    }
    
    // 2. 添加插件工具说明（原有逻辑）
    // ...
    
    return sb.String()
}
```

## 执行流程

```
用户请求 → Engine.ProcessMessage
    │
    ├── buildDynamicSystemPrompt (包含工具说明)
    │
    ├── reactLoop
    │   ├── callLLM → LLM 响应
    │   ├── ParseToolCalls → []ToolCall
    │   ├── ToolExecutor.ExecuteMultiple
    │   │   ├── 检查 builtinRegistry.Get(name)
    │   │   ├── 找到 → executeBuiltin(ctx, call)
    │   │   └── 未找到 → findToolPlugin(name) → executePlugin
    │   └── FormatToolResults → 下一轮
```

## 测试策略

### 单元测试（Table-Driven）

每个工具一个测试文件，使用 table-driven 测试：

```go
func TestBashTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]any
        want      string
        wantErr    bool
        errSubstr string
    }{
        {"valid command", map[string]any{"command": "echo hello"}, "hello\n", false, ""},
        {"timeout", map[string]any{"command": "sleep 10", "timeout": 0.1}, "", true, "timeout"},
        {"dangerous", map[string]any{"command": "rm -rf /"}, "", true, "dangerous"},
    }
    // ...
}
```

### 集成测试

```go
func TestEngine_BuiltinToolIntegration(t *testing.T) {
    // 测试 LLM 调用内置工具
    // 测试工具结果正确返回
    // 测试多轮 ReAct 循环
}
```

### 安全测试

```go
func TestSecurityChecker_BuiltinTools(t *testing.T) {
    // 测试 bash 危险命令检测
    // 测试路径遍历检测
    // 测试敏感路径保护
}
```

## 实现步骤

1. 创建 `internal/engine/builtin/` 目录和 `registry.go` 接口
2. 实现 6 个内置工具（含单元测试）: `bash.go`, `read.go`, `write.go`, `edit.go`, `glob.go`, `grep.go`
3. 更新 `SecurityChecker` 支持 `bash` 工具名（与 `shell` 并行）
4. 修改 `ToolExecutor` 添加 `builtinRegistry` 字段和 `Get()` 优先查找
5. 修改 `Engine` 添加 `builtinRegistry` 和 `workDir` 字段
6. 更新 `buildToolGuide()` 调用 `builtinRegistry.List()` 获取 `ToolInfo`（含 `Parameters`）
7. 添加配置支持 `work_dir`
8. 添加集成测试

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| 工具名冲突 | ToolExecutor 检测冲突，日志警告，忽略插件 |
| 安全问题 | SecurityChecker + 路径遍历检查 + 结果大小限制 |
| 性能影响 | Context 取消 + 超时限制 + 结果截断 |
| 配置复杂度 | work_dir 默认当前目录，无需配置即可使用 |
