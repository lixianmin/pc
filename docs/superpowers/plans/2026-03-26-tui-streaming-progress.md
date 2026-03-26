# TUI Streaming Progress Display Implementation Plan

> **For agentic workers:** REQUIRED to use superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Display ReAct loop progress in TUI message area with human-readable format

**Architecture:** TUI polls streaming chunks from RPC client, formats them as progress messages, and displays in message area with icons and summaries

**Tech Stack:** Go, Bubbletea TUI framework, RPC streaming

---

## Task 1: Add streaming fields to Model struct

**Files:**
- Modify: `internal/tui/model.go:43-45`

- [ ] **Step 1: Add streamChannel field**

Add import and field:
```go
import (
    // ... existing imports ...
    "github.com/lixianmin/pc/pkg/protocol"
)

// In Model struct, after streamBuffer:
streamChannel <-chan protocol.ProcessMessageStreamChunk
```

---

## Task 2: Update streamChunkMsg to include chunk data

**Files:**
- Modify: `internal/tui/model.go:789-792`

- [ ] **Step 1: Replace streamChunkMsg struct**

```go
type streamChunkMsg struct {
    chunk protocol.ProcessMessageStreamChunk
}
```

---

## Task 3: Implement sendToAgent with streaming

**Files:**
- Modify: `internal/tui/model.go:712-726`

- [ ] **Step 1: Replace sendToAgent method**

```go
func (my *Model) sendToAgent(message string) tea.Cmd {
    return func() tea.Msg {
        if my.rpcClient == nil {
            return errorMsg("not connected to gateway")
        }

        ch, err := my.rpcClient.ProcessMessageStreamRaw(my.sessionID, message)
        if err != nil {
            return errorMsg(err.Error())
        }

        my.streamChannel = ch
        return streamStartMsg{}
    }
}
```

---

## Task 4: Implement pollStreamChunk method

**Files:**
- Modify: `internal/tui/model.go` (after sendToAgent)

- [ ] **Step 1: Add pollStreamChunk method**

```go
func (my *Model) pollStreamChunk() tea.Cmd {
    return func() tea.Msg {
        if my.streamChannel == nil {
            return streamDoneMsg{}
        }

        chunk, ok := <-my.streamChannel
        if !ok {
            my.streamChannel = nil
            return streamDoneMsg{fullContent: my.streamBuffer.String()}
        }

        if chunk.Error != "" {
            my.streamChannel = nil
            return errorMsg(chunk.Error)
        }

        if chunk.Done {
            my.streamChannel = nil
            return streamDoneMsg{fullContent: my.streamBuffer.String()}
        }

        return streamChunkMsg{chunk: chunk}
    }
}
```

---

## Task 5: Update streamStartMsg handler

**Files:**
- Modify: `internal/tui/model.go:287-290`

- [ ] **Step 1: Update handler to start polling**

```go
case streamStartMsg:
    my.isStreaming = true
    my.streamBuffer.Reset()
    my.status = "Streaming..."
    return my, my.pollStreamChunk()
```

---

## Task 6: Implement chunk type formatting

**Files:**
- Modify: `internal/tui/model.go` (add new helper function)

- [ ] **Step 1: Add formatChunkContent helper**

```go
func formatChunkContent(chunk protocol.ProcessMessageStreamChunk) string {
    switch chunk.Type {
    case protocol.ChunkTypeThinking:
        return fmt.Sprintf("⚙ %s", chunk.Content)
    case protocol.ChunkTypeToolCall:
        desc := tools.StrTake(chunk.Content, 60)
        return fmt.Sprintf("▶ %s: %s", chunk.Tool, desc)
    case protocol.ChunkTypeToolResult:
        return fmt.Sprintf("  ✓ %s", summarizeToolResult(chunk.Tool, chunk.Content))
    case protocol.ChunkTypeResponse:
        return chunk.Content
    default:
        return chunk.Content
    }
}

func summarizeToolResult(tool, output string) string {
    switch tool {
    case "bash":
        lines := strings.Count(output, "\n") + 1
        return fmt.Sprintf("完成 (%d 行输出)", lines)
    case "read":
        return fmt.Sprintf("已读取 %d 字符", len(output))
    case "write":
        return "文件已写入"
    case "edit":
        return "文件已修改"
    default:
        return "完成"
    }
}
```

---

## Task 7: Update streamChunkMsg handler

**Files:**
- Modify: `internal/tui/model.go:292-312`

- [ ] **Step 1: Update handler to format and display chunks**

```go
case streamChunkMsg:
    formatted := formatChunkContent(msg.chunk)
    my.streamBuffer.WriteString(formatted)
    my.streamBuffer.WriteString("\n")
    
    if my.ready {
        lastIdx := len(my.messages) - 1
        if lastIdx >= 0 && my.messages[lastIdx].Role == "agent-streaming" {
            my.messages[lastIdx].Content = my.streamBuffer.String()
        } else {
            my.messages = append(my.messages, Message{
                Role:    "agent-streaming",
                Content: my.streamBuffer.String(),
            })
        }
        my.viewport.SetContent(my.renderMessages())
        if !my.userScrolled {
            my.viewport.GotoBottom()
        }
    }

    return my, my.pollStreamChunk()
```

---

## Task 8: Update streamDoneMsg handler

**Files:**
- Modify: `internal/tui/model.go:314-327`

- [ ] **Step 1: Update handler to finalize message**

```go
case streamDoneMsg:
    my.isStreaming = false
    my.status = "Connected"
    if my.ready {
        lastIdx := len(my.messages) - 1
        if lastIdx >= 0 && my.messages[lastIdx].Role == "agent-streaming" {
            my.messages[lastIdx].Role = "agent"
            my.messages[lastIdx].Content = msg.fullContent
        }
        my.viewport.SetContent(my.renderMessages())
        if !my.userScrolled {
            my.viewport.GotoBottom()
        }
    }
```

---

## Task 9: Update renderMessages to handle agent-streaming

**Files:**
- Modify: `internal/tui/model.go:700-707`

- [ ] **Step 1: Add agent-streaming case**

```go
for _, msg := range my.messages {
    switch msg.Role {
    case "user":
        b.WriteString(userStyle.Render("You: " + msg.Content))
    case "agent", "agent-streaming":
        b.WriteString(agentStyle.Render("Agent: " + msg.Content))
    }
    b.WriteString("\n\n")
}
```

---

## Task 10: Build and test

- [ ] **Step 1: Build project**

Run: `make build`
Expected: SUCCESS

- [ ] **Step 2: Run tests**

Run: `make test`
Expected: All tests pass

- [ ] **Step 3: Manual test**

1. Start gateway: `./pc gateway run`
2. Start TUI: `./pc tui`
3. Send: "统计一下downloads下面每张图片的大小"
4. Expected: See progress messages like "⚙ Thinking...", "▶ Bash: find...", "✓ 完成"

---

## Task 11: Commit changes

- [ ] **Step 1: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat(tui): display streaming progress with human-readable format"
```
