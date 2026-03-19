# M5 Task Management Integration Plan

> **For agentic workers:** REQUIRED: Use superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate M5 Task + Decomposer functionality into Engine for goal decomposition and task management.

**Architecture:** Add TaskManager and Decomposer to Engine. When user sends goal-triggering messages (`/task` or prefixes like "我想"), decompose into steps, create task, and return formatted response.

**Tech Stack:** Go 1.25+, internal/task package

---

## File Structure

| File | Responsibility |
|------|----------------|
| `internal/engine/engine.go` | Core engine with task management |
| `internal/engine/engine_test.go` | Tests for task decomposition |
| `internal/task/task.go` | Task model and Manager |
| `internal/task/decomposer.go` | Goal decomposition logic |

---

## Task 1: Fix Test Compilation Errors

**Files:**
- Modify: `internal/engine/engine_test.go:1-100`

- [x] **Step 1: Add missing import**

```go
import (
	"context"
	"testing"

	"github.com/lixianmin/pc/internal/task"
)
```

- [x] **Step 2: Fix variable naming conflict**

Change `for _, task := range tasks` to `for _, tk := range tasks` to avoid shadowing.

- [x] **Step 3: Fix undefined variable**

Change `if !found` to `if foundTask == nil`.

- [x] **Step 4: Run tests to verify**

Run: `go test -v ./internal/engine/ -run TestEngine_TaskDecomposition`
Expected: PASS

---

## Task 2: Verify Integration

**Files:**
- Test: `internal/engine/engine_test.go`

- [ ] **Step 1: Run all engine tests**

Run: `go test -v ./internal/engine/`
Expected: All tests pass

- [ ] **Step 2: Run full test suite**

Run: `make test`
Expected: All tests pass

---

## Task 3: Update Documentation

**Files:**
- Modify: `notes/03.tasks.md`
- Modify: `notes/tasks/completed/m5-task.md`

- [ ] **Step 1: Update M5 status**

In `notes/tasks/completed/m5-task.md`, change status from `🔄 代码完成，待集成` to `✅ 已集成`.

- [ ] **Step 2: Update task index**

In `notes/03.tasks.md`, update M5 row status to `✅ 已完成`.

- [ ] **Step 3: Commit documentation**

```bash
git add notes/03.tasks.md notes/tasks/completed/m5-task.md
git commit -m "docs: update M5 status to completed"
```

---

## Verification Checklist

- [ ] All tests pass (`make test`)
- [ ] Build succeeds (`make build`)
- [ ] No lint errors (`make lint`)
- [ ] Documentation updated
- [ ] Commits made with verification

---

## Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-03-19 | Initial plan created |
