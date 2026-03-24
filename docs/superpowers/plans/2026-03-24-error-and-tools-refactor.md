# Error and Tools Refactor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 引入 AppError 类型统一错误处理，将工具列表移至 BAML 定义

**Architecture:** 
1. 定义 AppError 结构体实现 error 接口，替代现有 fmt.Errorf
2. 在 BAML 中定义工具 class，LLM 返回具体工具实例
3. Skill 作为 UseSkill 特殊工具处理

**Tech Stack:** Go 1.25+, BAML

---

## Task 1: AppError 类型定义

**Files:**
- Create: `pkg/error/error.go`
- Create: `pkg/error/error_test.go`

- [ ] **Step 1: Write the failing test**

```go
package error

import (
	"testing"
)

func TestAppError_Error(t *testing.T) {
	err := &AppError{
		Code:    "EngineProcess",
		Message: "failed to process message: invalid input",
	}

	expected := "[EngineProcess] failed to process message: invalid input"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}

func TestAppError_IsAppError(t *testing.T) {
	err := &AppError{
		Code:    "TestCode",
		Message: "test message",
	}

	var appErr *AppError
	if !AsAppError(err, &appErr) {
		t.Error("AsAppError should return true for AppError")
	}
	if appErr.Code != "TestCode" {
		t.Errorf("appErr.Code = %q, want %q", appErr.Code, "TestCode")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./pkg/error/...`
Expected: FAIL - package not found

- [ ] **Step 3: Write minimal implementation**

```go
package error

import (
	"errors"
	"fmt"
)

type AppError struct {
	Code    string
	Message string
}

func (my *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", my.Code, my.Message)
}

func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func AsAppError(err error, target **AppError) bool {
	return errors.As(err, target)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./pkg/error/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/error/
git commit -m "feat: add AppError type with Code and Message"
```

---

## Task 2: 定义 BAML 工具 Class

**Files:**
- Modify: `baml_src/tools.baml`

- [ ] **Step 1: 添加工具 Class 定义到 BAML**

```baml
// baml_src/tools.baml

class BashTool {
  tool_name "bash" @description("Execute bash commands")
  command string @description("The bash command to execute")
}

class ReadTool {
  tool_name "read" @description("Read file content")
  file_path string @description("Path to the file to read")
}

class WriteTool {
  tool_name "write" @description("Write content to file")
  file_path string @description("Path to the file to write")
  content string @description("Content to write")
}

class EditTool {
  tool_name "edit" @description("Edit file by replacing old string with new string")
  file_path string @description("Path to the file to edit")
  old_string string @description("String to replace")
  new_string string @description("New string")
}

class WebSearchTool {
  tool_name "web_search" @description("Search the web for information")
  query string @description("Search query")
}

class WebFetchTool {
  tool_name "web_fetch" @description("Fetch content from a URL")
  url string @description("URL to fetch")
}

class UseSkill {
  tool_name "use_skill" @description("Use a skill by name with input")
  skill_name string @description("Name of the skill to use")
  input string @description("Input for the skill")
}

function ChooseTool(query: string, context: string) -> BashTool | ReadTool | WriteTool | EditTool | WebSearchTool | WebFetchTool | UseSkill {
  client "openai/gpt-5"
  prompt #"
    Given the user query and context, determine which tool to use.

    {{ ctx.output_format }} 

    {{ _.role('user') }}
    Query: {{ query }}
    Context: {{ context }}
  "#
}
```

- [ ] **Step 2: 生成 BAML 客户端**

Run: `make generate`
Expected: BAML client generated successfully

- [ ] **Step 3: Commit**

```bash
git add baml_src/tools.baml baml_client/
git commit -m "feat: add tool classes to BAML with ChooseTool function"
```

---

## Task 3: 重构 GetSkill 移除不必要的 error

**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: 简化 GetSkill 方法**

将：
```go
func (my *Engine) GetSkill(name string) (*skill.Skill, error) {
    if my.skillManager == nil {
        return nil, fmt.Errorf("skill manager not initialized")
    }
    return my.skillManager.GetSkill(name)
}
```

改为：
```go
func (my *Engine) GetSkill(name string) *skill.Skill {
    return my.skillManager.GetSkill(name)
}
```

- [ ] **Step 2: 更新调用方**

搜索所有调用 `GetSkill` 的地方，移除 error 处理。

- [ ] **Step 3: Commit**

```bash
git add internal/engine/engine.go
git commit -m "refactor: simplify GetSkill, remove unnecessary error return"
```

---

## Task 4: 逐步替换现有 error 为 AppError

**Files:**
- Multiple files in `internal/`

- [ ] **Step 1: 替换 internal/engine/ 中的 error**

将 `fmt.Errorf(...)` 替换为 `error.NewAppError("Code", "message")`

优先替换：
- engine.go
- stream.go
- session.go

- [ ] **Step 2: 替换 internal/gateway/ 中的 error**

- [ ] **Step 3: 替换其他模块**

- [ ] **Step 4: Commit**

```bash
git add internal/
git commit -m "refactor: replace fmt.Errorf with AppError across codebase"
```

---

## Summary

| Task | 描述 | 状态 |
|------|------|------|
| 1 | AppError 类型定义 | 待执行 |
| 2 | BAML 工具 Class 定义 | 待执行 |
| 3 | GetSkill 简化 | 待执行 |
| 4 | 替换现有 error 为 AppError | 待执行 |
