# Architecture Review Design

> **Status**: Draft  
> **Created**: 2026-03-16  
> **Author**: Claude (with superpowers)

---

## Goal

Review PersonalClaw architecture and identify improvements to enhance maintainability, reliability, and developer experience.

---

## Background

PersonalClaw has completed M1-M10 milestones and has accumulated some technical debt. This review identifies critical issues and proposes improvements to address them.

---

## Current Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      CLI Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   gateway    │  │     tui      │  │     task    │       │
│  │   命令       │  │   交互界面   │  │   命令       │       │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘       │
└─────────┼─────────────────┼─────────────────┼──────────┘
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │
                    ┌───────▼───────┐
                    │   RPC Client  │
                    │ (Unix Socket) │
                    └───────┬───────┘
                            │
┌───────────────────────────▼─────────────────────────────┐
│                   Gateway Layer                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐       │
│  │   Daemon     │  │  RPC Server  │  │   Gateway    │       │
│  │  生命周期    │  │   请求处理   │  │   协调器     │       │
│  └──────────────┘  └──────────────┘  └──────┬───────┘       │
└─────────────────────────────────────────────┼───────────┘
                                               │
┌─────────────────────────────────────────────┼───────────┐
│                     PC 核心                  │          │
│                                              │          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────▼───────┐  │
│  │  Agent       │  │  Plugin      │  │  Memory      │  │
│  │  Manager     │  │  Manager     │  │  Service     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  │
│         │                 │                 │          │
│  ┌──────▼─────────────────▼─────────────────▼───────┐  │
│  │            Core Engine                           │  │
│  │  (ReAct loop, session management, tool execution) │  │
│  └─────────────────────────────────────────────────┘  │
│         │                                              │
│  ┌──────▼──────────────────────────────────────────┐  │
│  │            Protocol Layer                        │  │
│  │  (stdio/gRPC 协议实现)                            │  │
│  └─────────────────────────────────────────────────┘  │
└───────────────────────────────────────────────────────┘
```

---

## Issues Identified

### Issue 1: Session State Duplication (P0)

**Current State**: Session state is managed in 3 places:
- `internal/memory/memory.go` - `MemoryService` with sessions map
- `internal/engine/session.go` - `Session` struct  
- `internal/engine/engine.go` - another sessions map

**Problem**: 
- Confusion about where state lives
- Risk of inconsistency
- Testing complexity

**Proposed Solution**: Consolidate to Engine only
1. Remove `internal/memory/memory.go` (or mark deprecated)
2. Use `Session` struct in `internal/engine/session.go` exclusively
3. Update all callers to use Engine sessions

**Impact**: 
- Cleaner architecture
- Single source of truth
- Easier testing

---

### Issue 2: Gateway Daemon Not Fully Implemented (P0) - ✅ FIXED

**Current State**: `daemon.Run()` has TODO comments

**Solution**: ✅ Implemented full initialization:
- Initialize PluginManager
- Scan plugins
- Initialize Engine
- Start RPC server
- Handle shutdown gracefully

---

### Issue 3: Debug Logging in Production Code (P1) - ✅ FIXED

**Current State**: `pkg/protocol/stdio.go` has `fmt.Fprintf(os.Stderr, ...)` debug prints

**Solution**: ✅ Replaced with `logo` package

---

### Issue 4: File Naming Typo (P1) - ✅ FIXED

**Current State**: `internal/plugin/plguin_manager.go` has typo

**Solution**: ✅ Renamed to `plugin_manager.go`

---

## Future Improvements (P2)

### Improvement 1: Streaming Support in ReAct Loop

**Current**: LLM response is buffered completely before processing

**Opportunity**: Stream tokens as they arrive for faster perceived response

**Proposed**: Add streaming support to Engine

---

### Improvement 2: Plugin Health Monitoring

**Current**: No way to detect unhealthy plugins

**Opportunity**: Add health check mechanism

**Proposed**: 
- Add `HealthCheck()` method to protocol
- Periodic health monitoring in Gateway
- Auto-restart unhealthy plugins

---

## Implementation Priority

| Priority | Task | Effort | Status |
|----------|------|--------|--------|
| P0 | Session consolidation | 4-6 hours | ⏳ Pending |
| P0 | Gateway daemon implementation | 2-3 hours | ✅ Done |
| P1 | Debug logging cleanup | 30 min | ✅ Done |
| P1 | File naming fix | 5 min | ✅ Done |
| P2 | Streaming support | 4-6 hours | ⏳ Future |
| P2 | Plugin health monitoring | 2-3 hours | ⏳ Future |

---

## Recommendations

1. **Immediate**: Session consolidation should be done next sprint
2. **Medium-term**: Consider streaming support for better UX
3. **Long-term**: Plugin health monitoring improves reliability

---

## Approval Checklist

- [ ] User has reviewed this design
- [ ] Implementation plan has been created
- [ ] All stakeholders agree on priority

