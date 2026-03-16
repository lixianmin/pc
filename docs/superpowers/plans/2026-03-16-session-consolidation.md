# Session Consolidation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Consolidate session state management into Engine, eliminating duplication between memory.Service and engine.Session.

**Architecture:** 
- Remove `internal/memory/memory.go` dependency from Engine
- Keep `Session` struct in `internal/engine/session.go` as single source of truth
- Update all callers to use Engine sessions directly

**Tech Stack:** Go 1.25+, github.com/lixianmin/logo

---

## Chunk 1: Analysis and Preparation

### Task 1: Analyze Current Usage

**Files:**
- Read: `internal/memory/memory.go`
- Read: `internal/engine/session.go`
- Read: `internal/engine/engine.go`

- [ ] **Step 1: Identify MemoryService callers**

Run: `grep -r "memory.Service\|MemoryService\|memory.NewService" --include="*.go" .`

Expected: List of files that use MemoryService

- [ ] **Step 2: Identify Session usage in Engine**

Run: `grep -n "sessions\|Session" internal/engine/engine.go`

Expected: Current session management code

- [ ] **Step 3: Document dependencies**

Create list of all files that need modification

---

## Chunk 2: Remove MemoryService from Engine

### Task 2: Update Engine to Not Use MemoryService

**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: Remove memory.Service import and field**

Remove any reference to `*memory.Service` in Engine struct

- [ ] **Step 2: Verify Session is self-contained**

Ensure `Session` struct has all needed methods:
- `AddMessage(role, content string) error`
- `GetMessages() []Message`
- `SetContextLimit(limit int) error`
- `Clear()`

- [ ] **Step 3: Run tests**

Run: `go test ./internal/engine/... -v`

Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "refactor(engine): remove memory.Service dependency"
```

---

## Chunk 3: Update Gateway RPC Server

### Task 3: Update RPC Server Session Handling

**Files:**
- Modify: `internal/gateway/rpc_server.go`

- [ ] **Step 1: Check for MemoryService usage**

Run: `grep -n "memory\|Memory" internal/gateway/rpc_server.go`

Expected: No direct memory.Service usage (uses Engine)

- [ ] **Step 2: Verify session creation through Engine**

Ensure `handleProcessMessage` uses `engine.CreateSession()`

- [ ] **Step 3: Run tests**

Run: `go test ./internal/gateway/... -v`

Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "refactor(gateway): use Engine for session management"
```

---

## Chunk 4: Deprecate or Remove Memory Package

### Task 4: Handle Memory Package

**Files:**
- Modify: `internal/memory/memory.go`
- Modify: `internal/memory/memory_test.go`

- [ ] **Step 1: Add deprecation notice**

Add comment to `memory.go`:
```go
// Deprecated: Use engine.Session instead. This package will be removed in a future version.
```

- [ ] **Step 2: Check for external dependencies**

Run: `grep -r "internal/memory" --include="*.go" . | grep -v "memory/" | grep -v "_test.go"`

Expected: No external usage (only tests)

- [ ] **Step 3: Keep for backward compatibility (optional)**

If there are external usages, keep package with deprecation notice

- [ ] **Step 4: Run all tests**

Run: `go test ./...`

Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "refactor(memory): deprecate in favor of engine.Session"
```

---

## Chunk 5: Update Tests

### Task 5: Update All Related Tests

**Files:**
- Modify: `internal/engine/*_test.go`
- Modify: `internal/gateway/*_test.go`

- [ ] **Step 1: Update engine tests**

Remove any memory.Service mock usage in engine tests

- [ ] **Step 2: Update gateway tests**

Ensure gateway tests use Engine sessions

- [ ] **Step 3: Run full test suite**

Run: `go test ./... -v`

Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "test: update tests for session consolidation"
```

---

## Chunk 6: Final Verification

### Task 6: Verify Complete Consolidation

**Files:**
- Verify: All `*.go` files

- [ ] **Step 1: Search for remaining MemoryService usage**

Run: `grep -r "MemoryService\|memory\.Service\|memory\.NewService" --include="*.go" . | grep -v "_test.go" | grep -v "Deprecated"`

Expected: No production code usage

- [ ] **Step 2: Verify single source of truth**

Run: `grep -rn "sessions.*map\|Session{" internal/engine/`

Expected: Only in engine.go and session.go

- [ ] **Step 3: Run full build and test**

Run: `make all`

Expected: Build and all tests pass

- [ ] **Step 4: Final commit**

```bash
git add -A && git commit -m "refactor: complete session consolidation into Engine"
git push origin dev
```

---

## Summary

| Task | Description | Status |
|------|-------------|--------|
| Task 1 | Analyze current usage | ⏳ |
| Task 2 | Remove MemoryService from Engine | ⏳ |
| Task 3 | Update Gateway RPC Server | ⏳ |
| Task 4 | Deprecate memory package | ⏳ |
| Task 5 | Update tests | ⏳ |
| Task 6 | Final verification | ⏳ |

**Estimated Time:** 2-3 hours
