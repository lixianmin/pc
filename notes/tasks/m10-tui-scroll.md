---
milestone: M10-TUI
title: TUI 滚动功能增强
status: 🔄 进行中
created: 2026-03-02
---

# M10-TUI: TUI 滚动功能增强

> 增强 TUI 的滚动体验，支持 scrollbar 视觉指示器和智能滚动行为

---

## 需求映射

| 需求编号 | 需求描述 | 优先级 |
|---------|---------|--------|
| CLI-016 | TUI 支持滚动查看超出屏幕的历史输出内容（带 scrollbar 指示器） | P1 |

---

## M10-TUI-001: Scrollbar 视觉指示器 ✅ 已完成

**优先级**: P1 | **需求**: CLI-016
**架构映射**: `internal/tui/model.go`
**完成时间**: 2026-03-02

### 实现步骤

1. ✅ 在 `Model` 结构体添加 `userScrolled` 字段
2. ✅ 在 `Styles` 结构体添加 `ScrollbarTrack` 和 `ScrollbarThumb` 样式
3. ✅ 实现 `renderScrollbar()` 函数
4. ✅ 在 `View()` 方法中集成 scrollbar 渲染
5. ✅ 计算 scrollbar 位置和高度比例

### 实现代码

```go
// renderScrollbar 渲染 scrollbar
func (m *Model) renderScrollbar() string {
    if !m.ready || m.viewport.TotalLineCount() <= m.viewport.Height {
        return ""
    }

    // 计算 scrollbar 参数
    totalLines := m.viewport.TotalLineCount()
    visibleLines := m.viewport.Height
    scrollPercent := float64(m.viewport.YOffset) / float64(totalLines-visibleLines)

    // 计算 thumb 位置和高度
    trackHeight := visibleLines
    thumbHeight := max(1, visibleLines*visibleLines/totalLines)
    thumbPos := int(scrollPercent * float64(trackHeight-thumbHeight))

    // 构建 scrollbar
    var sb strings.Builder
    for i := 0; i < trackHeight; i++ {
        if i >= thumbPos && i < thumbPos+thumbHeight {
            sb.WriteString(m.styles.ScrollbarThumb.Render("█"))
        } else {
            sb.WriteString(m.styles.ScrollbarTrack.Render("░"))
        }
        if i < trackHeight-1 {
            sb.WriteString("\n")
        }
    }

    return sb.String()
}
```

### 验收标准

- [x] Scrollbar 在内容超出一屏时显示
- [x] Scrollbar thumb 位置随滚动同步更新
- [x] Scrollbar thumb 高度反映可见内容比例
- [x] Scrollbar 样式符合整体设计（紫色 thumb，灰色 track）

---

## M10-TUI-002: 增强键盘快捷键 ✅ 已完成

**优先级**: P1 | **需求**: CLI-016
**架构映射**: `internal/tui/model.go`

### 实现步骤

1. 在 `Update()` 方法中添加新的按键处理
2. 实现 `PgUp/PgDn` 翻页
3. 实现 `Home/End` 跳到开始/结束
4. 实现 `Shift+↑/↓` 临时滚动
5. 更新帮助信息

### 快捷键映射

| 按键 | 动作 | 状态 |
|------|------|------|
| `↑/↓` | 逐行滚动 | ✅ 已有 |
| `PgUp/PgDn` | 翻页 | ⏳ 待实现 |
| `Home` | 跳到第一条消息 | ⏳ 待实现 |
| `End` | 跳到最后一条消息 | ⏳ 待实现 |
| `Shift+↑/↓` | 临时滚动（不锁定） | ⏳ 待实现 |
| `Ctrl+C/Esc` | 退出 | ✅ 已有 |

### 代码示例

```go
case tea.KeyMsg:
    switch msg.Type {
    // ... 已有按键处理 ...

    case tea.KeyPgUp:
        m.viewport.LineUp(10) // 向上翻 10 行
        m.userScrolled = true

    case tea.KeyPgDown:
        m.viewport.LineDown(10) // 向下翻 10 行
        // 如果滚动到底部，重置 userScrolled
        if m.viewport.AtBottom() {
            m.userScrolled = false
        }

    case tea.KeyHome:
        m.viewport.GotoTop()
        m.userScrolled = true

    case tea.KeyEnd:
        m.viewport.GotoBottom()
        m.userScrolled = false
    }
```

### 验收标准

- [ ] `PgUp/PgDn` 实现翻页
- [ ] `Home` 跳到第一条消息
- [ ] `End` 跳到最后一条消息
- [ ] `Shift+↑/↓` 实现临时滚动
- [ ] 帮助信息更新

---

## M10-TUI-003: 智能滚动行为 ✅ 已完成

**优先级**: P1 | **需求**: CLI-016
**架构映射**: `internal/tui/model.go`

### 实现步骤

1. 添加 `userScrolled` 状态标记
2. 新消息到达时，根据 `userScrolled` 决定是否自动滚动
3. 用户手动滚动时设置 `userScrolled = true`
4. 用户输入新消息时重置 `userScrolled = false`
5. 滚动到底部时自动重置 `userScrolled`

### 滚动行为状态机

```
                    新消息到达
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
   userScrolled    userScrolled      滚动到底部
      == true        == false            │
         │               │               │
         ▼               ▼               ▼
    保持当前      自动滚动到底部     userScrolled
    位置不变                              = false
         │                               │
         └───────────────┬───────────────┘
                         │
                    用户手动滚动
                         │
                         ▼
                   userScrolled = true
```

### 代码示例

```go
type Model struct {
    // ... 已有字段 ...
    userScrolled bool  // 用户是否手动滚动过
}

// Update 中处理响应消息
case responseMsg:
    m.status = "Connected"
    m.addMessage("agent", string(msg))
    if m.ready {
        m.viewport.SetContent(m.renderMessages())
        // 只有用户没有手动滚动时才自动滚动到底部
        if !m.userScrolled {
            m.viewport.GotoBottom()
        }
    }

// 处理键盘滚动
case tea.KeyMsg:
    switch msg.Type {
    case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
        m.userScrolled = true
        // ... 滚动处理 ...
    }
```

### 验收标准

- [ ] 新消息到达时自动滚动到底部（如果用户未手动滚动）
- [ ] 用户手动滚动后保持当前位置
- [ ] 用户输入新消息时恢复自动滚动
- [ ] 滚动到底部时重置 userScrolled 状态
- [ ] 单元测试验证滚动行为

---

## M10-TUI-004: 视图布局重构 ✅ 已完成

**优先级**: P1 | **需求**: CLI-016
**架构映射**: `internal/tui/view.go`

### 实现步骤

1. 创建 `view.go` 文件
2. 将 `View()` 方法中的布局逻辑移到 `view.go`
3. 实现 `layout()` 函数，使用 lipgloss 进行布局
4. 集成 scrollbar 到主布局
5. 保持现有样式

### 布局结构

```
┌─────────────────────────────────────────────────┐
│ Header (PersonalClaw v0.1.0)                    │
├────────────────────────────────────────────┬────┤
│                                            │    │
│  Viewport (消息显示区域)                      │ SB │  <- Scrollbar
│                                            │    │
│                                            │    │
├────────────────────────────────────────────┴────┤
│ Status: Connected | Session: abc12345          │
├─────────────────────────────────────────────────┤
│ > [Input area...                                ]│
└─────────────────────────────────────────────────┘
```

### 代码示例

```go
// View 渲染主界面
func (m *Model) View() string {
    if !m.ready {
        return "Loading..."
    }

    if m.quitting {
        return "Goodbye!\n"
    }

    // 各个组件
    header := m.styles.Header.Render("PersonalClaw v0.1.0")
    status := m.styles.StatusBar.Render(
        fmt.Sprintf("Status: %s | Session: %s", m.status, m.sessionID[:8]),
    )
    prompt := m.styles.InputPrompt.Render("> ")
    input := m.textarea.View()

    // 主内容区域（viewport + scrollbar）
    content := lipgloss.JoinHorizontal(
        lipgloss.Top,
        m.viewport.View(),
        m.renderScrollbar(),
    )

    // 组装最终布局
    return lipgloss.JoinVertical(
        lipgloss.Left,
        header,
        content,
        status,
        prompt+input,
    )
}
```

### 验收标准

- [ ] 布局代码分离到 view.go
- [ ] Scrollbar 正确集成到布局
- [ ] 界面渲染正常
- [ ] 窗口大小变化时正确调整
- [ ] 代码结构清晰

---

## 任务清单汇总

| ID | 任务 | 优先级 | 状态 |
|----|-----|--------|------|
| M10-TUI-001 | Scrollbar 视觉指示器 | P1 | ✅ 已完成 |
| M10-TUI-002 | 增强键盘快捷键 | P1 | ✅ 已完成 |
| M10-TUI-003 | 智能滚动行为 | P1 | ✅ 已完成 |
| M10-TUI-004 | 视图布局重构 | P1 | ✅ 已完成 |

---

## 技术决策记录

### 决策 1: Scrollbar 实现方式

**选择**: 使用 lipgloss 手动渲染，而非 bubbles/viewport 内置

**理由**:
- bubbles/viewport 没有内置 scrollbar 组件
- 手动渲染可以完全控制样式和行为
- 代码量不大，易于维护

### 决策 2: 智能滚动触发条件

**选择**: 基于 `userScrolled` 状态标记

**理由**:
- 简单直观
- 与 Claude Code 行为一致
- 用户可以随时通过按 End 键恢复自动滚动

### 决策 3: Scrollbar 位置

**选择**: 放在 viewport 右侧

**理由**:
- 符合大多数 GUI 应用的惯例
- 不占用消息显示空间
- 视觉上清晰
