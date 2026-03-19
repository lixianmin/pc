# AGENTS.md - AI Agent Development Guide

This file provides essential information for AI coding agents working in this repository.

## Core Principles (Non-Negotiable)

Aligned with [00.constitution.md](./notes/00.constitution.md):

1. **Spec is Source of Truth** - All architecture and code serve spec.md (spec.md → arch.md → tasks.md)
2. **Test-First Imperative** - All features start with failing unit tests (Red-Green-Refactor)
3. **Less is More** - No unnecessary abstractions, no unnecessary dependencies (Library-First)
4. **Code Readability First** - Code's primary purpose is to be understood by humans (no global mutable state)

## Documentation Index

| File | Responsibility | When to Read |
|------|----------------|--------------|
| [00.constitution.md](./notes/00.constitution.md) | Project constitution, non-negotiable principles | **First step of every task** |
| [01.spec.md](./notes/01.spec.md) | Requirements, single source of truth | Before feature development |
| [02.arch.md](./notes/02.arch.md) | Architecture, tech stack decisions | Before architecture changes |
| [03.tasks.md](./notes/03.tasks.md) | Task list **index** (milestone overview) | During development tracking |
| [04.lesson.md](./notes/04.lesson.md) | Lessons learned, historical mistakes | **Second step of every task** |
| [05.todo.md](./notes/05.todo.md) | Pending tasks and decisions | For tracking todos |

### Task Execution Order

1. Read `00.constitution.md` relevant sections
2. Read `01.spec.md` (requirements first)
3. Read `02.arch.md` as needed (architecture)
4. Read `04.lesson.md` relevant lessons
5. Verify task aligns with principles
6. Execute

## Build, Test, and Lint Commands

### Building
```bash
make build              # Build the main binary
make build-plugins      # Build all plugins
make all                # Full build: fmt, vet, test, build, build-plugins
```

### Testing
```bash
make test                                    # Run all tests with coverage
go test -v ./...                             # Run all tests (verbose)
go test -v -run TestName ./path/to/package   # Run single test
go test -v -run TestFunction ./path/to/pkg   # Run specific test function
go test -v -coverprofile=coverage.out ./...  # Generate coverage report
make test-coverage                           # Generate HTML coverage report
```

### Code Quality
```bash
make fmt                # Format code (go fmt)
make vet                # Vet code (go vet)
make lint               # Run golangci-lint
make deps               # Download and tidy dependencies
```

## Code Style Guidelines

### Language and Formatting
- Go version: 1.25+
- Use `go fmt` for formatting (run via `make fmt`)
- No code comments unless explicitly requested
- Tabs for indentation (Go standard)
- **Variable Declaration**: Prefer `var` over `:=` for explicit type clarity

Example:
```go
// Preferred
var name string = "example"
var count int = 10
var items []string = make([]string, 0)

// Avoid when type should be explicit
name := "example"
count := 10
items := []string{}
```

### Imports Organization
Organize imports in two groups, separated by blank lines:
1. Standard library imports
2. External packages (third-party)

Example:
```go
import (
    "context"
    "fmt"
    
    "github.com/lixianmin/logo"
    "github.com/lixianmin/pc/pkg/types"
)
```

### Receiver Naming
- Always use `my` as the receiver name for methods
- Consistent across all types

Example:
```go
func (my *Engine) Process() error { ... }
func (my *Config) Validate() error { ... }
```

### Naming Conventions

| Category | Convention | Example |
|----------|-----------|---------|
| Abbreviations | CamelCase (not all caps) | `sessionId` not `sessionID` |
| File names | Reflect main type | `skill_manager.go` → `SkillManager` |
| Timestamps | PascalCase | `CreateAt`, `UpdateAt`, `Ts` |
| ID generation | ULID | Use `github.com/oklog/ulid/v2` |
| Constants | PascalCase or camelCase | `MessageTypeCall`, `Version` |

### Type Definitions
- Use `any` instead of `interface{}`
- Prefer concrete types over interfaces when there's only one implementation
- Define types at package level, not inside functions

### Error Handling
- Always handle errors explicitly (non-negotiable per constitution)
- Use `fmt.Errorf` with `%w` for error wrapping
- Never ignore errors with `_`

Example:
```go
if err != nil {
    return fmt.Errorf("failed to process: %w", err)
}
```

### Functions
- Keep functions focused on a single responsibility
- Pre-allocate slice capacity when size is known
- Return early to reduce nesting

Example:
```go
func (my *Engine) ProcessMessage(ctx context.Context, sessionId, message string) (string, error) {
    if sessionId == "" {
        return "", fmt.Errorf("session ID cannot be empty")
    }
    if message == "" {
        return "", fmt.Errorf("message cannot be empty")
    }
    // ... rest of function
}
```

## Testing Requirements

### Test-First Development (TDD)
- **Mandatory**: All features start with a failing test (Red-Green-Refactor)
- No production code without a test
- Tests written before implementation

### Table-Driven Tests
- **Required**: Use table-driven test pattern for all unit tests
- Define test cases as a slice of structs
- Use `t.Run()` for subtests

Example:
```go
func TestProcessMessage(t *testing.T) {
    tests := []struct {
        name      string
        sessionId string
        message   string
        wantErr   bool
    }{
        {
            name:      "valid message",
            sessionId: "test-1",
            message:   "Hello",
            wantErr:   false,
        },
        {
            name:      "empty message",
            sessionId: "test-2",
            message:   "",
            wantErr:   true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test implementation
        })
    }
}
```

### Integration Tests
- Prefer integration tests over mocks
- Use real dependencies when possible
- Integration test files: `*_integration_test.go` or in `tests/` directory

## Architecture Principles

### Directory Structure
```
pc/
├── cmd/pc/              # Main application entry point
├── internal/            # Private application code
│   ├── agent/          # Agent management
│   ├── engine/         # Core engine
│   ├── gateway/        # Gateway daemon and RPC
│   ├── plugin/         # Plugin system
│   ├── config/         # Configuration management
│   ├── memory/         # Memory system
│   ├── tui/            # Terminal UI
│   └── logger/         # Logging wrapper
├── pkg/                 # Public packages (importable by plugins)
│   ├── protocol/       # Plugin protocol definitions
│   └── types/          # Shared type definitions
├── examples/            # Example plugins
├── notes/               # Project documentation
│   ├── 00.constitution.md
│   ├── 01.spec.md
│   ├── 02.arch.md
│   └── 03.tasks.md
└── Makefile
```

### Library-First Principle
- Every feature must first be implemented as an independent library
- Use standard library when possible (e.g., `net/http` over frameworks)
- Avoid unnecessary abstractions
- Keep packages cohesive and focused

### No Global Mutable State
- **Forbidden**: Global variables for runtime state (counters, caches, connection pools)
- **Allowed**: Package-level read-only constants
- **Required**: All dependencies injected via function parameters or struct fields

### Goroutine Management
- **Mandatory**: Use `loom.Go()` instead of `go` keyword
- Package: `github.com/lixianmin/got/loom`

Example:
```go
import "github.com/lixianmin/got/loom"

// Wrong
go my.backgroundTask()

// Correct
loom.Go(func(later loom.Later) {
    my.backgroundTask()
})
```

## Project-Specific Conventions

### Logging
- Library: `github.com/lixianmin/logo`
- Methods: `logo.Info()`, `logo.Error()`, etc.
- **Forbidden**: `fmt.Println()` for logging (except CLI user interaction)

Example:
```go
logo.Info("[Session:", sessionId, "] User:", message)
logo.Error("[Session:", sessionId, "] Error:", err)
```

### Configuration
- Format: YAML (`.yml`), never JSON
- Config file: `~/.pc/config.yml`
- Plugin configs: `~/.pc/plugins/<type>/<name>/config.yml`
- Use `gopkg.in/yaml.v3` for parsing

### Plugin Protocol
- Communication: stdio (JSON) or Unix Socket (RPC)
- Message format defined in `pkg/protocol/`
- Plugins are separate processes with their own `go.mod`

### ID Generation
- Use ULID: `github.com/oklog/ulid/v2`
- Example: `ulid.Make().String()`

## Common Mistakes to Avoid

Based on `notes/04.lesson.md`:

1. **Writing code before tests** - Violates constitution Article 2
2. **Using JSON instead of YAML** - Project uses YAML for all configs
3. **Unused imports** - Always clean up after refactoring
4. **Direct slice comparison** - Slices are not comparable in Go
5. **Over-engineering abstractions** - Follow "less is more" principle
6. **Mock-heavy tests** - Prefer integration tests with real dependencies
7. **Nil pointer dereference** - Check logger initialization before use
8. **Code not integrated to main flow** - Ensure features work in production

## Commit Message Format

```
<type>(<scope>): <subject>

# Examples:
feat(engine): add ReAct loop support
fix(gateway): resolve pidfile race condition
test(config): add validation tests
docs(arch): update plugin protocol specification
```

## Quick Reference

- **Language**: Go 1.25+
- **Build**: `make test`, `make build`
- **Single test**: `go test -v -run TestName ./path/to/package`
- **Commit**: `<type>(<scope>): <subject>`
- **Config**: YAML format only
- **Receiver**: Always `my`
- **Logging**: `github.com/lixianmin/logo`
- **Goroutines**: Always use `loom.Go()`
- **Tests**: Table-driven, test-first (TDD)
