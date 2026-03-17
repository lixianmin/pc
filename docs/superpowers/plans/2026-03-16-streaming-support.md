# Streaming Support in ReAct Loop - Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add streaming support to ReAct loop for real-time token output, improving user experience with faster perceived response time.

**Architecture:** 
- Add streaming method to Engine
- Update RPC protocol to support streaming
- Update TUI to handle streaming responses
- Keep backward compatibility with non-streaming mode

**Tech Stack:** Go 1.25+, channels, context, github.com/lixianmin/logo

---

## Chunk 1: Define Streaming Types

### Task 1: Create Streaming Types

**Files:**
- Create: `internal/engine/stream.go`
- Modify: `pkg/protocol/rpc.go`

- [ ] **Step 1: Define StreamChunk type**

```go
// StreamChunk represents a chunk of streaming response.
type StreamChunk struct {
    Content string `json:"content"`
    Done    bool   `json:"done"`
    Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 2: Add streaming RPC method**

In `pkg/protocol/rpc.go`:
```go
const (
    // ... existing methods ...
    RPCMethodProcessMessageStream RPCMethod = "ProcessMessageStream"
)

type ProcessMessageStreamChunk struct {
    Content string `json:"content"`
    Done    bool   `json:"done"`
    Error   string `json:"error,omitempty"`
}
```

- [ ] **Step 3: Run tests**

Run: `go build ./pkg/protocol/...`

Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(protocol): add streaming types for ProcessMessageStream"
```

---

## Chunk 2: Add Engine Streaming Support

### Task 2: Implement Engine Streaming Method

**Files:**
- Create: `internal/engine/stream.go`

- [ ] **Step 1: Implement ProcessMessageStream**

```go
func (my *Engine) ProcessMessageStream(ctx context.Context, sessionId, message string) <-chan StreamChunk {
    ch := make(chan StreamChunk, 100)
    
    go func() {
        defer close(ch)
        
        // Create session if needed
        if _, exists := my.sessions[sessionId]; !exists {
            my.CreateSession(sessionId)
        }
        
        // Get or create LLM plugin
        if my.llmPlugin == nil {
            ch <- StreamChunk{Error: "no LLM plugin configured", Done: true}
            return
        }
        
        // Call streaming method
        streamCh, err := my.callLLMStream(ctx, sessionId, message)
        if err != nil {
            ch <- StreamChunk{Error: err.Error(), Done: true}
            return
        }
        
        var fullContent strings.Builder
        for chunk := range streamCh {
            fullContent.WriteString(chunk)
            ch <- StreamChunk{Content: chunk, Done: false}
        }
        
        // Save to session
        session := my.sessions[sessionId]
        session.AddMessage("user", message)
        session.AddMessage("assistant", fullContent.String())
        
        ch <- StreamChunk{Done: true}
    }()
    
    return ch
}
```

- [ ] **Step 2: Implement callLLMStream**

```go
func (my *Engine) callLLMStream(ctx context.Context, sessionId, message string) (<-chan string, error) {
    session := my.sessions[sessionId]
    history := session.GetMessages()
    
    messages := my.buildMessages(history, message)
    
    params := map[string]any{
        "messages": messages,
    }
    
    // Call plugin stream method
    result, err := my.pluginManager.CallPlugin(my.llmPlugin, "stream", params)
    if err != nil {
        return nil, err
    }
    
    streamCh, ok := result.(<-chan string)
    if !ok {
        return nil, fmt.Errorf("unexpected stream response format")
    }
    
    return streamCh, nil
}
```

- [ ] **Step 3: Add streaming tests**

Create `internal/engine/stream_test.go` with table-driven tests

- [ ] **Step 4: Run tests**

Run: `go test ./internal/engine/... -v`

Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat(engine): add ProcessMessageStream for real-time output"
```

---

## Chunk 3: Update Plugin Protocol

### Task 3: Add Streaming to Plugin Protocol

**Files:**
- Modify: `pkg/protocol/stdio.go`
- Modify: `examples/plugins/llm/openai/openai.go`

- [ ] **Step 1: Add stream method handling to protocol**

In `StdioProtocol`, handle streaming responses:
```go
func (my *StdioProtocol) CallStream(method string, params any) (<-chan string, error) {
    // Similar to Call but returns a channel
}
```

- [ ] **Step 2: Implement real streaming in OpenAI plugin**

Update `examples/plugins/llm/openai/openai.go`:
```go
func (my *GLMClient) Stream(messages []map[string]string, options map[string]any) (<-chan string, error) {
    ch := make(chan string)
    
    go func() {
        defer close(ch)
        // Real implementation with SSE/streaming HTTP
        // For now, mock streaming
        words := strings.Split("This is a streaming response", " ")
        for _, word := range words {
            ch <- word + " "
            time.Sleep(100 * time.Millisecond)
        }
    }()
    
    return ch, nil
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./examples/plugins/llm/openai/... -v`

Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(plugin): add streaming support to LLM protocol"
```

---

## Chunk 4: Add RPC Streaming Support

### Task 4: Implement RPC Streaming Handler

**Files:**
- Modify: `internal/gateway/rpc_server.go`
- Modify: `internal/gateway/rpc_client.go`

- [ ] **Step 1: Add streaming handler to RPC server**

```go
func (my *RPCServer) handleProcessMessageStream(ctx context.Context, params json.RawMessage) (interface{}, error) {
    var req protocol.ProcessMessageParams
    if err := json.Unmarshal(params, &req); err != nil {
        return nil, fmt.Errorf("invalid params: %w", err)
    }
    
    // Return a channel that will be serialized as SSE
    streamCh := my.engine.ProcessMessageStream(ctx, req.SessionID, req.Message)
    
    // Convert to array for JSON response
    // Client will poll for chunks
    return streamCh, nil
}
```

Note: For simplicity, we'll use a polling approach first. SSE can be added later.

- [ ] **Step 2: Add client streaming method**

```go
func (my *RPCClient) ProcessMessageStream(sessionID, message string) (<-chan protocol.ProcessMessageStreamChunk, error) {
    // Polling-based streaming for simplicity
    ch := make(chan protocol.ProcessMessageStreamChunk)
    
    go func() {
        defer close(ch)
        // Implementation
    }()
    
    return ch, nil
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/gateway/... -v`

Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat(gateway): add ProcessMessageStream RPC method"
```

---

## Chunk 5: Update TUI for Streaming

### Task 5: Update TUI to Handle Streaming

**Files:**
- Modify: `internal/tui/model.go`

- [ ] **Step 1: Add streaming state to model**

```go
type model struct {
    // ... existing fields ...
    isStreaming bool
    streamBuffer strings.Builder
}
```

- [ ] **Step 2: Update ProcessMessage to use streaming**

```go
func (my *model) processMessageWithStream(message string) tea.Cmd {
    return func() tea.Msg {
        streamCh, err := my.rpcClient.ProcessMessageStream(my.sessionID, message)
        if err != nil {
            return errMsg{err}
        }
        
        my.isStreaming = true
        my.streamBuffer.Reset()
        
        for chunk := range streamCh {
            if chunk.Error != "" {
                return errMsg{fmt.Errorf(chunk.Error)}
            }
            my.streamBuffer.WriteString(chunk.Content)
            // Trigger UI update
        }
        
        my.isStreaming = false
        return streamCompleteMsg{my.streamBuffer.String()}
    }
}
```

- [ ] **Step 3: Add streaming message types**

```go
type streamChunkMsg struct {
    content string
}
type streamCompleteMsg struct {
    fullContent string
}
```

- [ ] **Step 4: Update Update() to handle streaming**

```go
case streamChunkMsg:
    my.streamBuffer.WriteString(msg.content)
    // Update last message content
case streamCompleteMsg:
    my.isStreaming = false
    // Finalize message
```

- [ ] **Step 5: Run TUI tests**

Run: `go test ./internal/tui/... -v`

Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat(tui): add streaming response support"
```

---

## Chunk 6: Integration Testing

### Task 6: Add Integration Tests

**Files:**
- Create: `tests/integration/stream_test.go`

- [ ] **Step 1: Create streaming integration test**

```go
func TestStreamingEndToEnd(t *testing.T) {
    // Start gateway
    // Send streaming message
    // Verify chunks received
    // Verify final response saved
}
```

- [ ] **Step 2: Run integration tests**

Run: `go test ./tests/integration/... -v`

Expected: All tests pass

- [ ] **Step 3: Manual testing**

Test:
1. Start gateway
2. Run TUI
3. Send message
4. Verify streaming output

- [ ] **Step 4: Final commit**

```bash
git add -A && git commit -m "test(integration): add streaming e2e tests"
git push origin dev
```

---

## Summary

| Task | Description | Effort | Status |
|------|-------------|--------|--------|
| Task 1 | Define streaming types | 30 min | ✅ Complete |
| Task 2 | Engine streaming support | 1-2 hours | ✅ Complete |
| Task 3 | Plugin protocol streaming | 1-2 hours | ✅ Complete |
| Task 4 | RPC streaming support | 1-2 hours | ✅ Complete |
| Task 5 | TUI streaming support | 1-2 hours | ✅ Complete |
| Task 6 | Integration testing | 1 hour | ⏳ Pending |

**Completed:** Tasks 1-5 (core streaming infrastructure + TUI)
**Remaining:** Task 6 (full e2e tests)

**Actual Time:** 2-3 hours for core infrastructure

**Backward Compatibility:** Non-streaming `ProcessMessage` remains unchanged

---

## Implementation Notes

### Phase 1: Core Infrastructure (Tasks 1-3)
- Define types
- Add Engine streaming
- Update plugin protocol

### Phase 2: Gateway & TUI (Tasks 4-5)
- RPC streaming
- TUI streaming display

### Phase 3: Testing (Task 6)
- Integration tests
- Manual verification

### Future Enhancements
- SSE (Server-Sent Events) for true streaming over HTTP
- WebSocket support for bidirectional streaming
- Streaming tool execution results
