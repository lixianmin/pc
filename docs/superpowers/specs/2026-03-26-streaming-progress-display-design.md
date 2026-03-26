# Streaming Progress Display Design

## Problem

When users ask questions that require multiple ReAct loop iterations (e.g., "list image sizes in downloads folder"), the TUI shows no progress until the final response. This makes users feel the system is stuck or timed out.

## Solution

Display intermediate progress in the message area with a clean, human-readable format.

## Design

### Message Types

Extend the `Message` struct to support progress updates:

```go
type Message struct {
    Role      string    // "user", "agent", "progress"
    Type      string    // "thinking", "tool_call", "tool_result", "response"
    Content   string
    Tool      string    // "bash", "read", "write", "edit"
    Timestamp time.Time
}
```

### Chunk Handling

| Chunk Type | Message Display | Status Bar |
|------------|-----------------|------------|
| thinking | `⚙ Thinking...` | "Thinking..." |
| tool_call | `▶ {Tool}: {description}` | "Running {tool}..." |
| tool_result | `  ✓ {summary}` | "Connected" |
| response | Agent response | "Connected" |

### Display Format

Example conversation flow:

```
User: 统计一下downloads下面每张图片的大小

⚙ Thinking...

▶ Bash: find ~/Downloads -type f -name "*.jpg"...
  ✓ 完成 (4 个文件)

▶ Bash: ls -lh ...
  ✓ 完成 (6 行输出)

Agent: 以下是您的图片大小统计表格：
| 文件名 | 大小 |
...
```

### Summary Generation

```go
func summarizeToolResult(tool, output string, meta any) string {
    switch tool {
    case "bash":
        lines := strings.Count(output, "\n") + 1
        if lines == 1 {
            return "完成"
        }
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

### TUI Changes

1. **Restore streaming support** - Add back `pollStreamChunk()` and `streamChannel`
2. **Handle chunk types** - Process `Type`, `Tool`, `Meta` fields from chunks
3. **Render progress messages** - Style progress messages differently from user/agent messages
4. **Update status bar** - Show current activity in status bar

### Protocol

Already implemented in `pkg/protocol/rpc.go`:

```go
type StreamChunkType string

const (
    ChunkTypeThinking   StreamChunkType = "thinking"
    ChunkTypeToolCall   StreamChunkType = "tool_call"
    ChunkTypeToolResult StreamChunkType = "tool_result"
    ChunkTypeResponse   StreamChunkType = "response"
    ChunkTypeDone       StreamChunkType = "done"
    ChunkTypeError      StreamChunkType = "error"
)

type ProcessMessageStreamChunk struct {
    Type    StreamChunkType `json:"type"`
    Content string          `json:"content,omitempty"`
    Tool    string          `json:"tool,omitempty"`
    Meta    any             `json:"meta,omitempty"`
    Done    bool            `json:"done"`
    Error   string          `json:"error,omitempty"`
}
```

## Implementation Steps

1. Add `pollStreamChunk()` method to TUI Model
2. Add `streamChannel` field to TUI Model
3. Update `sendToAgent()` to use `ProcessMessageStreamRaw()`
4. Handle different chunk types in `Update()`
5. Update `renderMessages()` to style progress messages
6. Update status bar based on chunk type

## Files to Modify

- `internal/tui/model.go` - Add streaming support and progress rendering
- `internal/tui/styles.go` - Add progress message styles (if needed)
