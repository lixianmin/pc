# Skill System Integration Plan

> **For agentic workers:** REQUIRED: Use superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Integrate SkillManager into Engine to allow LLM to use skills in ReAct loop.

**Architecture:** Add SkillManager to Engine. Load skills from `~/.pc/skills/` on startup. Make skills available in system prompt.

**Tech Stack:** Go 1.25+, internal/skill package

---

## File Structure

| File | Responsibility |
|------|----------------|
| `internal/engine/engine.go` | Add SkillManager field |
| `internal/engine/engine_test.go` | Test skill integration |
| `~/.pc/skills/example.md` | Example skill file |

---

## Task 1: Write Failing Test

**Files:**
- Modify: `internal/engine/engine_test.go`

- [ ] **Step 1: Write test for skill loading**

```go
func TestEngine_LoadSkills(t *testing.T) {
    e := NewEngine(nil)
    e.SetSkillDir("/tmp/test-skills")
    
    skills := e.ListSkills()
    if len(skills) == 0 {
        t.Error("Expected skills to be loaded")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/engine/ -run TestEngine_LoadSkills`
Expected: FAIL with "method not defined"

---

## Task 2: Add SkillManager to Engine
**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: Add SkillManager field to Engine struct**

Add `skillManager *skill.SkillManager` field.

- [ ] **Step 2: Add SetSkillDir method**

```go
func (my *Engine) SetSkillDir(dir string) error {
    my.skillManager = skill.NewSkillManager()
    return my.skillManager.LoadSkills(dir)
}
```

- [ ] **Step 3: Add ListSkills method**

```go
func (my *Engine) ListSkills() []skill.Skill {
    if my.skillManager == nil {
        return nil
    }
    return my.skillManager.ListSkills()
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/engine/ -run TestEngine_LoadSkills`

---

## Task 3: Include Skills in System Prompt
**Files:**
- Modify: `internal/engine/engine.go`

- [ ] **Step 1: Update buildDynamicSystemPrompt**

Add skill descriptions to system prompt when skills are available.

- [ ] **Step 2: Write test**

```go
func TestEngine_SkillsInSystemPrompt(t *testing.T) {
    e := NewEngine(nil)
    e.SetSkillDir("/tmp/test-skills")
    e.SetSystemPrompt("You are an assistant.")
    
    prompt := e.buildDynamicSystemPrompt()
    if !strings.Contains(prompt, "Skills") {
        t.Error("Expected skills in system prompt")
    }
}
```

- [ ] **Step 3: Run test**

---

## Task 4: Create Example Skill
**Files:**
- Create: `~/.pc/skills/code-review.md`

- [ ] **Step 1: Create skill file**

```markdown
# Code Review Skill

## Description
Perform code review on given code.

## Steps
1. Get the code to review from user
2. Analyze code quality and best practices
3. Return review comments
```

- [ ] **Step 2: Test skill loading**

---

## Verification Checklist
- [ ] All tests pass (`make test`)
- [ ] Build succeeds (`make build`)
- [ ] No lint errors (`make lint`)
- [ ] Skills load from `~/.pc/skills/`
- [ ] Skills appear in system prompt

---

## Change Log

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-03-19 | Initial plan created |
