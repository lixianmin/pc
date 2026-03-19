# 内置工具系统实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 PersonalClaw 添加内置工具支持（bash, read, write, edit, glob, grep），无需安装插件

**Architecture:** 在 Engine 中添加 builtinRegistry， 工具执行时优先检查内置工具，然后回退到插件工具

**Tech Stack:** Go 1.25+, 现有 Engine、 ToolExecutor

 SecurityChecker

---

## Task Structure

````markdown
### Task 1: 创建接口和注册表

**Files:**
- Create: `internal/engine/builtin_tool.go`
- Create: `internal/engine/builtin_registry.go`

- [ ] **Step 1: Write BuiltinTool interface**
```go
package engine

type ParamSchema struct {
	Type        string `json:"type"`
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
```

- [ ] **Step 2: Write Registry implementation**
```go
package engine

type Registry struct {
    mu    sync.RWMutex
    tools map[string]BuiltinTool
}

func NewRegistry() *Registry {
    return &Registry{
        tools: make(map[string]BuiltinTool),
    }
}

func (my *Registry) Register(tool BuiltinTool) {
    my.mu.Lock()
    defer my.mu.Unlock()
    if _, exists := my.tools[tool.Name()] {
        return
    }
    my.tools[tool.Name()] = tool
}

func (my *Registry) Get(name string) (BuiltinTool, bool) {
    my.mu.RLock()
    defer my.mu.Unlock()
    tool, exists := my.tools[name]
    return tool, exists
}

func (my *Registry) List() map[string]BuiltinTool {
    my.mu.RLock()
    defer my.mu.Unlock()
    result := make(map[string]BuiltinTool)
    for _, tool := range my.tools {
        result[tool.Name()] = tool
    }
    return result
}
```

- [ ] **Step 3: Commit**
```bash
git add internal/engine/builtin_tool.go internal/engine/builtin_registry.go
git commit -m "feat(engine): add builtin tool interface and registry"
```
---

### Task 2: 实现 Bash 工具
**Files:**
- Create: `internal/engine/builtin/bash.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "testing"
    "time"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestBashTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]any
        want      string
        wantErr    bool
        errContains string
    }{
        {
            name:      "valid echo command",
            params:    map[string]any{"command": "echo hello"},
            want:      "hello\n",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "command with workdir",
            params:    map[string]any{"command": "pwd", "workdir": "/tmp"},
            want:      "/tmp\n",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "timeout",
            params:    map[string]any{"command": "sleep 1", "timeout": 0.1},
            want:      "",
            wantErr:    true,
            errContains: "context deadline exceeded",
        },
        {
            name:      "dangerous command blocked",
            params:    map[string]any{"command": "rm -rf /"},
            want:      "",
            wantErr:    true,
            errContains: "dangerous",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            tool := builtin.NewBashTool()
            result, err := tool.Execute(ctx, tt.params)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if result != tt.want {
                    t.Errorf("expected %q, got %q", tt.want)
 result)
                }
            }
        })
    }
}
```

- [ ] **Step 2: Write Bash tool implementation**
```go
package builtin

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "time"
)

type BashTool struct {
    workdir string
    timeout time.Duration
}

func NewBashTool() *BashTool {
    return &BashTool{
        workdir: "",
        timeout: 120 * time.Second,
    }
}

func (my *BashTool) Name() string {
    return "bash"
}

func (my *BashTool) Description() string {
    return "Execute shell commands"
}

func (my *BashTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "command": {
            Type:        "string",
            Required:    true,
            Description: "The shell command to execute",
        },
        "workdir": {
            Type:        "string",
            Required:    false,
            Description: "Working directory (default: current directory)",
            Default:     "",
        },
        "timeout": {
            Type:        "number",
            Required:    false,
            Description: "Timeout in seconds (default: 120)",
            Default:     float64(120),
        },
    }
}

func (my *BashTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    command, ok := params["command"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: command")
    }
    
    workdir, _ := params["workdir"].(string)
    if workdir == "" {
        workdir = my.workdir
    }
    
    timeoutSec, _ := params["timeout"].(float64)
    if timeoutSec <= 0 {
        timeoutSec = 30
    }
    timeout := time.Duration(timeoutSec) * time.Second
    }
    timeout = t.timeout
    }
    
    var cmd *exec.Cmd
    if workdir != "" {
        cmd = exec.Command(command)
    } else {
        cmd = exec.CommandContext(ctx, command)
        cmd.Dir = workdir
    }
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    
    done := make(chan struct{})
        select {
        case <-ctx.Done():
            cmd.Process.Kill()
            return "", ctx.Err()
        case <-time.After(timeout):
            return "", fmt.Errorf("command timed out after %s", timeout)
        case out := <-struct {
            output string
            err    error
        }:
            output: stdout.String(),
            err:    nil,
        }
    }
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestBashTool ./internal/engine/builtin/
```

- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/bash.go internal/engine/builtin/bash_test.go
git commit -m "feat(engine): add bash builtin tool"
```
---

### Task 3: 实现 Read 工具
**Files:**
- Create: `internal/engine/builtin/read.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestReadTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]any
        want      string
        wantErr    bool
        errContains string
    }{
        {
            name:      "read file",
            params:    map[string]any{"path": "/etc/passwd"},
            want:      "",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "read with offset",
            params:    map[string]any{"path": "/etc/passwd", "offset": 1, "limit": 5},
            want:      "root:",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "file not found",
            params:    map[string]any{"path": "/nonexistent/file.txt"},
            want:      "",
            wantErr:    true,
            errContains: "file not found",
        },
        {
            name:      "is a directory",
            params:    map[string]any{"path": "/tmp"},
            want:      "",
            wantErr:    true,
            errContains: "is a directory",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.name == "is a directory" {
                os.MkdirTemp(tt.params["path"].(string), 10)
            }
            
            ctx := context.Background()
            tool := builtin.NewReadTool()
            result, err := tool.Execute(ctx, tt.params)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if !strings.Contains(result, tt.want) {
                    t.Errorf("expected %q, got %q", tt.want, result)
                }
            }
        })
    }
}
```

- [ ] **Step 2: Write Read tool implementation**
```go
package builtin

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "strconv"
    "strings"
)

type ReadTool struct{}

func NewReadTool() *ReadTool {
    return &ReadTool{}
}

func (my *ReadTool) Name() string {
    return "read"
}

func (my *ReadTool) Description() string {
    return "Read file contents"
}

func (my *ReadTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "path": {
            Type:        "string",
            Required:    true,
            Description: "The absolute path to the file",
        },
        "offset": {
            Type:        "number",
            Required:    false,
            Description: "Starting line number (1-indexed)",
            Default:     float64(1),
        },
        "limit": {
            Type:        "number",
            Required:    false,
            Description: "Maximum number of lines to return",
            Default:     float64(2000),
        },
    }
}

func (my *ReadTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    path, ok := params["path"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: path")
    }
    
    offset, _ := params["offset"].(float64)
    if !ok {
        offset = 1
    }
    
    limit, _ := params["limit"].(float64)
    if !ok {
        limit = 2000
    }
    
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", fmt.Errorf("failed to resolve path: %w", err)
    }
    
    info, err := os.Stat(absPath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("file not found: %s", path)
        }
        return "", fmt.Errorf("failed to stat file: %w", err)
    }
    
    if info.IsDir() {
        return "", fmt.Errorf("is a directory: %s", path)
    }
    
    file, err := os.Open(absPath)
    if err != nil {
        return "", fmt.Errorf("permission denied: %w", err)
    }
    defer file.Close()
    
    var lines []string
    scanner := bufio.NewScanner(file)
    lineNum := 0
    for scanner.Scan() {
        lineNum++
        if lineNum >= int(offset) && (limit <= 0 || lineNum < int(offset)+int(limit)) {
            lines = append(lines, scanner.Text())
        }
    }
    
    if scanner.Err() != nil {
        return "", fmt.Errorf("error reading file: %w", scanner.Err())
    }
    
    return strings.Join(lines, "\n"), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestReadTool ./internal/engine/builtin/
```

- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/read.go internal/engine/builtin/read_test.go
git commit -m "feat(engine): add read builtin tool"
```
---

### Task 4: 实现 Write 工具
**Files:**
- Create: `internal/engine/builtin/write.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestWriteTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]any
        wantErr    bool
        errContains string
    }{
        {
            name:      "write new file",
            params:    map[string]any{"path": "/tmp/test_write.txt", "content": "hello world"},
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "overwrite existing file",
            params:    map[string]any{"path": "/tmp/test_overwrite.txt", "content": "new content"},
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "missing path parameter",
            params:    map[string]any{"content": "test"},
            wantErr:    true,
            errContains: "missing required parameter: path",
        },
        {
            name:      "missing content parameter",
            params:    map[string]any{"path": "/tmp/test.txt"},
            wantErr:    true,
            errContains: "missing required parameter: content",
        },
    }
    
    tmpDir := os.TempDir("", "test_write_*")
    defer os.RemoveAll(tmpDir)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
                if tt.name == "write new file" || tt.name == "overwrite existing file" {
                    tmpFile := filepath.Join(tmpDir, tt.name+"_write.txt")
                    if tt.name == "overwrite existing file" {
                        os.WriteFile(tmpFile, []byte("old content"), 0644)
                    }
                }
                
                ctx := context.Background()
                tool := builtin.NewWriteTool()
                result, err := tool.Execute(ctx, tt.params)
                
                if tt.wantErr {
                    if err == nil {
                        t.Errorf("expected error, got nil")
                    }
                    if !strings.Contains(err.Error(), tt.errContains) {
                        t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                    }
                } else {
                    if err != nil {
                        t.Errorf("unexpected error: %v", err)
                    }
                    content, _ := os.ReadFile(tt.params["path"].(string))
                    if string(content) != tt.params["content"].(string) {
                        t.Errorf("content mismatch")
                    }
                }
            }
        })
    }
}
```

- [ ] **Step 2: Write Write tool implementation**
```go
package builtin

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
)

type WriteTool struct{}

func NewWriteTool() *WriteTool {
    return &WriteTool{}
}

func (my *WriteTool) Name() string {
    return "write"
}

func (my *WriteTool) Description() string {
    return "Write file contents"
}

func (my *WriteTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "path": {
            Type:        "string",
            Required:    true,
            Description: "The absolute path to write to",
        },
        "content": {
            Type:        "string",
            Required:    true,
            Description: "The content to write to the file",
        },
    }
}

func (my *WriteTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    path, ok := params["path"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: path")
    }
    
    content, ok := params["content"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: content")
    }
    
    absPath, err := filepath.Abs(path)
    if err != nil {
        return "", fmt.Errorf("failed to resolve path: %w", err)
    }
    
    dir := filepath.Dir(absPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", fmt.Errorf("failed to create directory: %w", err)
    }
    
    if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
        return "", fmt.Errorf("failed to write file: %w", err)
    }
    
    return fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), path), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestWriteTool ./internal/engine/builtin/
```

- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/write.go internal/engine/builtin/write_test.go
git commit -m "feat(engine): add write builtin tool"
```
---

### Task 5: 实现 Edit 工具
**Files:**
- Create: `internal/engine/builtin/edit.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "os"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestEditTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        setup         func()
        params    map[string]any
        want      string
        wantErr    bool
        errContains string
    }{
        {
            name:      "replace single occurrence",
            setup: func() {
                os.WriteFile("/tmp/test_edit.txt", []byte("hello world"), 0644)
            },
            params:    map[string]any{
                "path":       "/tmp/test_edit.txt",
                "old_string": "hello",
                "new_string": "goodbye",
            },
            want:      "Successfully edited /tmp/test_edit.txt",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "replace all occurrences",
            setup: func() {
                os.WriteFile("/tmp/test_edit_all.txt", []byte("hello hello hello"), 0644)
            },
            params:    map[string]any{
                "path":        "/tmp/test_edit_all.txt",
                "old_string": "hello",
                "new_string": "hi",
                "replace_all": true,
            },
            want:      "Successfully edited /tmp/test_edit_all.txt",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "old_string not found",
            setup: func() {
                os.WriteFile("/tmp/test_notfound.txt", []byte("different content"), 0644)
            },
            params:    map[string]any{
                "path":       "/tmp/test_notfound.txt",
                "old_string": "nonexistent",
                "new_string": "replacement",
            },
            want:      "",
            wantErr:    true,
            errContains: "old_string not found in file",
        },
        {
            name:      "multiple matches without replace_all",
            setup: func() {
                os.WriteFile("/tmp/test_multi.txt", []byte("hello hello"), 0644)
            },
            params:    map[string]any{
                "path":        "/tmp/test_multi.txt",
                "old_string": "hello",
                "new_string": "hi",
            },
            want:      "",
            wantErr:    true,
            errContains: "found multiple matches",
        },
    }
    
    tmpDir := os.TempDir("", "test_edit_*")
    defer os.RemoveAll(tmpDir)
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.setup != nil {
                tt.setup()
            }
            
            ctx := context.Background()
            tool := builtin.NewEditTool()
            result, err := tool.Execute(ctx, tt.params)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if !strings.Contains(result, tt.want) {
                    t.Errorf("expected %q, got %q", tt.want, result)
                }
            }
            
            if tt.setup != nil {
                os.Remove(tt.params["path"].(string))
            }
        })
    }
}
```

- [ ] **Step 2: Write Edit tool implementation**
```go
package builtin

import (
    "context"
    "fmt"
    "os"
    "strings"
)

type EditTool struct{}

func NewEditTool() *EditTool {
    return &EditTool{}
}

func (my *EditTool) Name() string {
    return "edit"
}

func (my *EditTool) Description() string {
    return "Edit file by string replacement"
}

func (my *EditTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "path": {
            Type:        "string",
            Required:    true,
            Description: "The absolute path to edit",
        },
        "old_string": {
            Type:        "string",
            Required:    true,
            Description: "The text to replace",
        },
        "new_string": {
            Type:        "string",
            Required:    true,
            Description: "The replacement text",
        },
        "replace_all": {
            Type:        "boolean",
            Required:    false,
            Description: "Replace all occurrences (default false)",
            Default:     false,
        },
    }
}

func (my *EditTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    path, ok := params["path"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: path")
    }
    
    oldString, ok := params["old_string"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: old_string")
    }
    
    newString, ok := params["new_string"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: new_string")
    }
    
    replaceAll := params["replace_all"].(bool)
    
    content, err := os.ReadFile(path)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }
    
    contentStr := string(content)
    
    if !strings.Contains(contentStr, oldString) {
        return "", fmt.Errorf("old_string not found in file")
    }
    
    count := strings.Count(contentStr, oldString)
    if count > 1 && !replaceAll {
        return "", fmt.Errorf("found multiple matches (%d), please use replace_all=true or provide more context", count)
    }
    
    newContent := strings.Replace(contentStr, oldString, newString)
    if replaceAll {
        newContent = strings.ReplaceAll(newContent, oldString, newString)
    }
    
    if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
        return "", fmt.Errorf("failed to write file: %w", err)
    }
    
    return fmt.Sprintf("Successfully edited %s", path), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestEditTool ./internal/engine/builtin/
```
- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/edit.go internal/engine/builtin/edit_test.go
git commit -m "feat(engine): add edit builtin tool"
```
---

### Task 6: 实现 Glob 工具
**Files:**
- Create: `internal/engine/builtin/glob.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestGlobTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        params    map[string]any
        want      string
        wantErr    bool
        errContains string
    }{
        {
            name:      "match go files",
            params:    map[string]any{"pattern": "**/*.go", "path": "."},
            want:      "",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "no matches",
            params:    map[string]any{"pattern": "*.nonexistent", "path": "."},
            want:      "",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "missing pattern",
            params:    map[string]any{"path": "."},
            want:      "",
            wantErr:    true,
            errContains: "missing required parameter: pattern",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx := context.Background()
            tool := builtin.NewGlobTool()
            result, err := tool.Execute(ctx, tt.params)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if tt.name == "match go files" {
                    if !strings.Contains(result, ".go") {
                        t.Errorf("expected .go files in result")
                    }
                }
                if tt.name == "no matches" {
                    if result != "" {
                        t.Errorf("expected empty result for no matches")
                    }
                }
            }
        })
    }
}
```

- [ ] **Step 2: Write Glob tool implementation**
```go
package builtin

import (
    "context"
    "fmt"
    "path/filepath"
)

type GlobTool struct{}

func NewGlobTool() *GlobTool {
    return &GlobTool{}
}

func (my *GlobTool) Name() string {
    return "glob"
}

func (my *GlobTool) Description() string {
    return "File name pattern matching"
}

func (my *GlobTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "pattern": {
            Type:        "string",
            Required:    true,
            Description: "The glob pattern (e.g., **/*.go)",
        },
        "path": {
            Type:        "string",
            Required:    false,
            Description: "The directory to search (default: current directory)",
            Default:     "",
        },
    }
}

func (my *GlobTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    pattern, ok := params["pattern"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: pattern")
    }
    
    searchPath, _ := params["path"].(string)
    if searchPath == "" {
        searchPath = "."
    }
    
    matches, err := filepath.Glob(filepath.Join(searchPath, pattern))
    if err != nil {
        return "", fmt.Errorf("failed to match pattern: %w", err)
    }
    
    if len(matches) == 0 {
        return "", nil
    }
    
    return strings.Join(matches, "\n"), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestGlobTool ./internal/engine/builtin/
```
- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/glob.go internal/engine/builtin/glob_test.go
git commit -m "feat(engine): add glob builtin tool"
```
---

### Task 7: 实现 Grep 工具
**Files:**
- Create: `internal/engine/builtin/grep.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "os"
    "path/filepath"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestGrepTool_Execute(t *testing.T) {
    tests := []struct {
        name      string
        setup         func()
        params    map[string]any
        want      string
        wantErr    bool
        errContains string
    }{
        {
            name:      "search in file",
            setup: func() {
                tmpFile := filepath.Join(os.TempDir(), "test_grep.txt")
                os.WriteFile(tmpFile, []byte("hello world\nfoo bar\nhello again"), 0644)
            },
            params:    map[string]any{"pattern": "hello", "path": os.TempDir()},
            want:      "",
            wantErr:    false,
            errContains: "",
        },
        {
            name:      "invalid regex",
            params:    map[string]any{"pattern": "[invalid", "path": "."},
            want:      "",
            wantErr:    true,
            errContains: "invalid regex pattern",
        },
        {
            name:      "missing pattern",
            params:    map[string]any{"path": "."},
            want:      "",
            wantErr:    true,
            errContains: "missing required parameter: pattern",
        },
    }
    
    tmpDir := ""
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.name == "search in file" {
                tmpDir = os.TempDir()
                defer os.RemoveAll(tmpDir)
                tt.setup()
                tt.params["path"] = tmpDir
            }
            
            ctx := context.Background()
            tool := builtin.NewGrepTool()
            result, err := tool.Execute(ctx, tt.params)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error, got nil")
                }
                if !strings.Contains(err.Error(), tt.errContains) {
                    t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
                if tt.name == "search in file" {
                    if !strings.Contains(result, "hello") {
                        t.Errorf("expected 'hello' in result")
                    }
                    if strings.Contains(result, "test_grep.txt") {
                        t.Errorf("expected file path in result")
                    }
                }
            }
        })
    }
}
```

- [ ] **Step 2: Write Grep tool implementation**
```go
package builtin

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

type GrepTool struct{}

func NewGrepTool() *GrepTool {
    return &GrepTool{}
}

func (my *GrepTool) Name() string {
    return "grep"
}

func (my *GrepTool) Description() string {
    return "Search file contents using regex"
}

func (my *GrepTool) Parameters() map[string]engine.ParamSchema {
    return map[string]engine.ParamSchema{
        "pattern": {
            Type:        "string",
            Required:    true,
            Description: "The regex pattern to search for",
        },
        "path": {
            Type:        "string",
            Required:    false,
            Description: "The directory to search (default: current directory)",
            Default:     "",
        },
        "include": {
            Type:        "string",
            Required:    false,
            Description: "File pattern filter (e.g., *.go)",
            Default:     "",
        },
    }
}

func (my *GrepTool) Execute(ctx context.Context, params map[string]any) (string, error) {
    pattern, ok := params["pattern"].(string)
    if !ok {
        return "", fmt.Errorf("missing required parameter: pattern")
    }
    
    searchPath, _ := params["path"].(string)
    if searchPath == "" {
        searchPath = "."
    }
    
    includePattern, _ := params["include"].(string)
    
    re, err := regexp.Compile(pattern)
    if err != nil {
        return "", fmt.Errorf("invalid regex pattern: %w", err)
    }
    
    var results []string
    err := filepath.WalkDir(searchPath, func(path string, d os.DirEntry, err error) {
        if err != nil {
                return err
            }
            if d.IsDir() {
                return nil
            }
            
            if includePattern != "" {
                matched, err := filepath.Match(includePattern, filepath.Base(path))
                if err != nil || !matched {
                    return nil
                }
            }
            
            file, err := os.Open(path)
            if err != nil {
                return nil
            }
            defer file.Close()
            
            scanner := bufio.NewScanner(file)
            lineNum := 0
            for scanner.Scan() {
                lineNum++
                if re.MatchString(scanner.Text()) {
                    results = append(results, fmt.Sprintf("%s:%d: %s", path, lineNum, scanner.Text()))
                }
            }
            
            return nil
        })
    )
    
    if err != nil {
        return "", fmt.Errorf("error walking directory: %w", err)
    }
    
    return strings.Join(results, "\n"), nil
}
```

- [ ] **Step 3: Run tests to verify they pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestGrepTool ./internal/engine/builtin/
```
- [ ] **Step 4: Commit**
```bash
git add internal/engine/builtin/grep.go internal/engine/builtin/grep_test.go
git commit -m "feat(engine): add grep builtin tool"
```
---

### Task 8: Modify ToolExecutor to support built-in tools
**Files:**
- Modify: `internal/engine/tool_executor.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
    "github.com/lixianmin/pc/internal/engine/builtin"
)

func TestToolExecutor_BuiltinTools(t *testing.T) {
    ctx := context.Background()
    
    executor := engine.NewToolExecutor(nil)
    
    if !executor.IsToolAvailable("bash") {
        t.Error("bash tool should be available")
    }
    if !executor.IsToolAvailable("read") {
        t.Error("read tool should be available")
    }
    if !executor.IsToolAvailable("write") {
        t.Error("write tool should be available")
    }
    if !executor.IsToolAvailable("edit") {
        t.Error("edit tool should be available")
    }
    if !executor.IsToolAvailable("glob") {
        t.Error("glob tool should be available")
    }
    if !executor.IsToolAvailable("grep") {
        t.Error("grep tool should be available")
    }
}
```

- [ ] **Step 2: Modify ToolExecutor to add builtinRegistry field**
```go
package engine

import (
    "context"
    "time"

    "github.com/lixianmin/pc/internal/engine/builtin"
)

type ToolExecutor struct {
    pluginManager    PluginManager
    builtinRegistry  *builtin.Registry
    timeout           time.Duration
}

func NewToolExecutor(pm PluginManager) *ToolExecutor {
    return &ToolExecutor{
        pluginManager:    pm,
        builtinRegistry: builtin.NewRegistry(),
        timeout:           30 * time.Second,
    }
}

func (my *ToolExecutor) SetTimeout(timeout time.Duration) {
    my.timeout = timeout
}
```

- [ ] **Step 3: Modify Execute method to check built-in tools first**
```go
func (my *ToolExecutor) Execute(ctx context.Context, call ToolCall) ToolResult {
    if my.builtinRegistry == nil {
        return ToolResult{
            Name:  call.Name,
            Error: fmt.Errorf("builtin registry not initialized"),
        }
    }
    
    if tool, ok := my.builtinRegistry.Get(call.Name); ok {
        return my.executeBuiltin(ctx, call, tool)
    }
    
    if my.pluginManager == nil {
        return ToolResult{
            Name:  call.Name,
            Error: fmt.Errorf("plugin manager not initialized"),
        }
    }
    
    plugin, err := my.findToolPlugin(call.Name)
    if err != nil {
        return ToolResult{
            Name:  call.Name,
            Error: err,
        }
    }
    
    return my.executePlugin(ctx, call, plugin)
}

```

- [ ] **Step 4: Add executeBuiltin method**
```go
func (my *ToolExecutor) executeBuiltin(ctx context.Context, call ToolCall, tool builtin.BuiltinTool) ToolResult {
    paramsJSON, err := json.Marshal(call.Params)
    if err != nil {
        return ToolResult{
            Name:  call.Name,
            Error: fmt.Errorf("failed to marshal params: %w", err),
        }
    }
    
    var params map[string]any
    if err := json.Unmarshal(paramsJSON, &params); err != nil {
        params = make(map[string]any)
    }
    
    result, err := tool.Execute(ctx, params)
    if err != nil {
        return ToolResult{
            Name:  call.Name,
            Error: err,
        }
    }
    
    output, ok := result.(string)
    if !ok {
        output = result
    }
    
    return ToolResult{
        Name:   call.Name,
        Output: output,
    }
}
```

- [ ] **Step 5: Run tests to verify the pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestToolExecutor ./internal/engine/
```

- [ ] **Step 6: Commit**
```bash
git add internal/engine/tool_executor.go internal/engine/tool_executor_test.go
git commit -m "feat(engine): integrate builtin tools into ToolExecutor"
```
---

### Task 9: Modify Engine to support built-in tools
**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: Write the failing test**
```go
package engine_test

import (
    "context"
    "testing"

    "github.com/lixianmin/pc/internal/engine"
)

func TestEngine_BuiltinToolsIntegration(t *testing.T) {
    pm := &MockPluginManager{}
    e := engine.NewEngine(pm)
    
    tools := e.GetAvailableTools()
    toolNames := make(map[string]bool)
    for _, tool := range tools {
        toolNames[tool.Name] = true
    }
    
    expectedTools := []string{"bash", "read", "write", "edit", "glob", "grep"}
    for _, name := range expectedTools {
        if !toolNames[name] {
            t.Errorf("expected tool %s to be available", name)
        }
    }
}
```

- [ ] **Step 2: Add builtinRegistry field to Engine struct**
```go
type Engine struct {
    pluginManager    PluginManager
    builtinRegistry  *builtin.Registry
    workDir          string
    llmPlugin        *types.Plugin
    llmCallback      LLMCallback
    mockCaller       *MockPluginCaller
    skillManager     *skill.Manager
    systemPrompt     string
    sessions         map[string]*Session
    mu               sync.RWMutex
    maxIterations   int
    toolTimeout      time.Duration
}
```

- [ ] **Step 3: Initialize builtin registry in NewEngine**
```go
func NewEngine(pluginManager PluginManager) *Engine {
    e := &Engine{
        pluginManager:    pluginManager,
        builtinRegistry:  builtin.NewRegistry(),
        workDir:          "",
        llmPlugin:        nil,
        skillManager:     skill.NewManager(),
        systemPrompt:     "",
        sessions:         make(map[string]*Session),
        mu:               sync.RWMutex{},
        maxIterations:   10,
        toolTimeout:      30 * time.Second,
    }
    
    e.builtinRegistry.Register(builtin.NewBashTool())
    e.builtinRegistry.Register(builtin.NewReadTool())
    e.builtinRegistry.Register(builtin.NewWriteTool())
    e.builtinRegistry.Register(builtin.NewEditTool())
    e.builtinRegistry.Register(builtin.NewGlobTool())
    e.builtinRegistry.Register(builtin.NewGrepTool())
    
    return e
}
```

- [ ] **Step 4: Modify getAvailableTools to include built-in tools**
```go
func (my *Engine) getAvailableTools() []ToolInfo {
    var tools []ToolInfo
    
    my.mu.RLock()
    defer my.mu.RUnlock()
    
    for name, tool := range my.builtinRegistry.List() {
        tools = append(tools, ToolInfo{
            Name:        name,
            Description: tool.Description(),
            Type:        "builtin",
        })
    }
    
    if my.pluginManager != nil {
        plugins := my.pluginManager.ListPlugins()
        for _, p := range plugins {
            if p.Type == types.PluginTypeTool && p.Enabled {
                tools = append(tools, ToolInfo{
                    Name:        p.Name,
                    Description: fmt.Sprintf("%s tool", p.Name),
                    Type:        string(p.Type),
                })
            }
        }
    }
    
    return tools
}
```

- [ ] **Step 5: Update buildToolGuide to include parameter descriptions**
```go
func (my *Engine) buildToolGuide(tools []ToolInfo) string {
    var parts []string

    parts = append(parts, "## 工具使用指南")
    parts = append(parts, "")
    parts = append(parts, "当你需要获取外部信息或执行操作时，可以使用以下工具。")
    parts = append(parts, "")

    parts = append(parts, "### 可用工具")
    for _, t := range tools {
        parts = append(parts, fmt.Sprintf("- **%s**: %s", t.Name, t.Description))
    }
    parts = append(parts, "")
    parts = append(parts, "### 工具调用格式")
    parts = append(parts, "")
    parts = append(parts, "使用以下 XML 格式调用工具：")
    parts = append(parts, "")
    parts = append(parts, "<invoke>")
    parts = append(parts, "<name>工具名</name>")
    parts = append(parts, "<params>{\"参数\": \"值\"}</params>")
    parts = append(parts, "</invoke>")
    parts = append(parts, "")
    parts = append(parts, "### ReAct 工作流程")
    parts = append(parts, "1. 思考：分析用户需求")
    parts = append(parts, "2. 行动:调用工具")
    parts = append(parts, "3. 观察:接收工具结果")
    parts = append(parts, "4. 回复:基于结果回答用户")

    parts = append(parts, "")

    return strings.Join(parts, "\n")
}
```

- [ ] **Step 6: Run tests to verify the pass**
```bash
cd /Users/xmli/me/code/pc
 go test -v -run TestEngine_BuiltinToolsIntegration ./internal/engine/
```
- [ ] **Step 7: Commit**
```bash
git add internal/engine/engine.go internal/engine/engine_test.go
git commit -m "feat(engine): integrate builtin tools into Engine"
```
---

### Task 10: Run all tests and verify integration
**Files:**
- No new files

- [ ] **Step 1: Run all tests**
```bash
cd /Users/xmli/me/code/pc
 go test ./internal/engine/... -v
```

- [ ] **Step 2: Verify no regressions**
Check output for any test failures or errors

- [ ] **Step 3: Manual integration test (optional)**
Start the gateway and test with TUI:
Send a message that requires tool usage
 Verify tool is called and executed correctly
