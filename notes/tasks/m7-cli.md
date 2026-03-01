---
milestone: M7
title: CLI & Gateway
status: ✅ 已完成
completed: 2026-02-28
---

# M7: CLI & Gateway

> CLI 工具、Gateway 守护进程、TUI 交互界面

---

## M7-001: CLI 框架与子命令体系

**优先级**: P0 | **需求**: CLI-008, CLI-001 ~ CLI-003
**架构映射**: cmd/pc (新目录结构)

### 实现步骤

1. 引入 `github.com/spf13/cobra` CLI 框架
2. 重构 cmd/pc 目录结构：
   ```
   cmd/pc/
   ├── main.go              # 入口，初始化日志、配置
   ├── root.go              # 根命令定义
   ├── version.go           # version 子命令
   ├── gateway/             # gateway 子命令组
   │   ├── gateway.go
   │   ├── start.go
   │   ├── stop.go
   │   └── status.go
   ├── tui.go               # tui 子命令
   └── task/                # task 子命令组
       ├── task.go
       ├── list.go
       └── add.go
   ```
3. 实现 `pc version` 命令（--version 标志支持）
4. 实现 `pc gateway` 子命令组框架
5. 编写表格驱动测试

### 验收标准

- [x] `pc --help` 显示完整命令树
- [x] `pc version` 和 `pc --version` 正常工作
- [x] 子命令结构符合架构设计
- [x] 单元测试覆盖所有命令

---

## M7-002: Gateway Daemon 核心

**优先级**: P0 | **需求**: CLI-001, CLI-002, CF-007, CF-008
**架构映射**: internal/gateway (新增模块)

### 实现步骤

1. 创建 `internal/gateway/` 模块：
   - `gateway.go`: Gateway 结构体定义
   - `daemon.go`: Daemon 生命周期管理（start/stop/restart）
   - `pidfile.go`: pidfile 读写和进程检查
   - `server.go`: RPC server 实现
2. 实现 Gateway 作为核心协调器
3. 实现 `pc gateway start`：
   - 检查 pidfile，防止重复启动
   - 初始化 Gateway（加载 Engine、PluginManager、Agent）
   - 启动 RPC server（Unix Socket: `~/.pc/pc.sock`）
   - 写入 pidfile (`~/.pc/pc.pid`)
   - 信号处理（SIGTERM/SIGINT 优雅关闭）
4. 实现 `pc gateway stop`：
   - 读取 pidfile 获取 PID
   - 发送 SIGTERM 信号
   - 等待进程退出（带超时）
   - 清理 pidfile
5. 实现 `pc gateway status`：
   - 检查 pidfile 和进程存活
   - 尝试连接 RPC server 获取详细状态
6. 实现 daemon 日志文件输出 (`~/.pc/logs/pc.log`)
7. 编写表格驱动测试

### 验收标准

- [x] `pc gateway start` 后台启动，pidfile 正确写入
- [x] `pc gateway stop` 正确停止并清理
- [x] `pc gateway status` 准确显示运行状态
- [x] 信号处理触发优雅关闭
- [x] 崩溃后 pidfile stale 检测和清理
- [x] 单元测试通过

---

## M7-003: Gateway RPC 协议

**优先级**: P0 | **需求**: CLI-001, CLI-004
**架构映射**: internal/gateway/rpc.go, pkg/protocol/rpc.go

### 实现步骤

1. 定义 RPC 消息协议（JSON over Unix Socket）
2. 定义 RPC 接口方法：
   | Method | 参数 | 返回 | 说明 |
   |--------|------|------|------|
   | `ProcessMessage` | `{sessionId, message}` | `{response}` | 处理用户消息 |
   | `GetStatus` | `{}` | `{status, uptime, plugins}` | 获取状态 |
   | `ListSkills` | `{}` | `{skills: [...]}` | 列出技能 |
   | `ExecuteSkill` | `{name, params}` | `{result}` | 执行技能 |
   | `ListTasks` | `{status?}` | `{tasks: [...]}` | 列出任务 |
3. 实现 RPC server（daemon 端）
4. 实现 RPC client（CLI 端）
5. 错误处理定义（标准 RPC 错误码）
6. 编写表格驱动测试

### 验收标准

- [x] RPC 协议定义完整
- [x] 所有方法正常工作
- [x] 错误处理规范统一
- [x] 超时机制有效
- [x] 单元测试通过

---

## M7-004: TUI 交互界面

**优先级**: P0 | **需求**: CLI-004
**架构映射**: internal/tui/ (新增模块)

### 实现步骤

1. 创建 `internal/tui/` 模块
2. 引入依赖：
   - `github.com/charmbracelet/bubbletea` - TUI 框架
   - `github.com/charmbracelet/bubbles` - 预置组件
   - `github.com/charmbracelet/lipgloss` - 样式
3. 实现 `pc tui` 命令
4. 实现核心功能：
   - REPL 消息循环
   - 命令历史记录（↑↓ 键，保存到 `~/.pc/history`）
   - Tab 补全（skills、/commands）
   - `/quit` 或 Ctrl+C 退出
   - `/clear` 清屏
   - `/skills` 列出可用技能
   - `/status` 显示 gateway 状态
5. 编写表格驱动测试

### 验收标准

- [x] TUI 启动成功，界面渲染正常
- [x] 能与 Agent 进行多轮对话
- [x] 命令历史记录可用（跨会话保留）
- [x] Tab 补全提示正确
- [x] 退出命令正常工作
- [x] 单元测试通过

---

## M7-005: TUI 富功能支持

**优先级**: P1 | **需求**: CLI-010
**架构映射**: internal/tui/

### 实现步骤

1. 实现 @命令解析器：
   - `@skillname` → 引用 Skill，注入 Skill 描述到上下文
   - `@filepath` → 引用文件，读取文件内容注入上下文
2. 实现 Slash Commands 处理：
   - `/quit` 或 `/exit` → 退出 TUI
   - `/clear` → 清屏
   - `/skills` → 列出可用技能
   - `/status` → 显示 gateway 状态
   - 所有 slash commands 本地处理，不发送到 LLM
3. 集成代码高亮：
   - 引入 `github.com/charmbracelet/lipgloss` 或 `chroma`
   - 检测 markdown code block 并应用语法高亮
4. 优化 Tab 补全：
   - 补全技能名（`@skill` 时）
   - 补全文件路径（`@path` 时）
   - 补全 slash commands（输入 `/` 时）
5. 编写表格驱动测试

### 验收标准

- [x] `@skillname` 正确引用技能并注入上下文
- [x] `@filepath` 正确读取文件并注入上下文
- [x] 所有 slash commands 本地处理，不发送到 LLM
- [x] 基础样式通过 lipgloss 支持
- [x] Tab 补全覆盖 skills、files、commands
- [x] 单元测试通过

---

## M7-006: TUI 与 Channel 共存

**优先级**: P1 | **需求**: CLI-011
**架构映射**: internal/gateway/, internal/channel/

### 实现步骤

1. 架构设计：
   - TUI 和 Channel 插件作为独立的输入源
   - 两者消息都路由到同一个 Agent/Engine
   - TUI 走 Unix Socket RPC，Channel 插件走 stdio 协议
2. 修改 Gateway 架构：
   - Gateway 支持多输入源并发监听
   - 每个输入源标识来源（tui/telegram/...）
   - 消息统一路由到 Engine 处理
3. 实现输入源管理：
   - `InputSource` 接口定义
   - `TUISource` 实现
   - `ChannelSource` 适配已有 Channel 插件
4. 响应路由：
   - 根据消息来源路由响应
   - TUI 消息 → TUI 客户端
   - Channel 消息 → Channel 插件
5. 编写集成测试

### 验收标准

- [x] `InputSource` 接口定义完成
- [x] `InputSourceRegistry` 多输入源管理实现
- [x] `MessageRouter` 消息路由实现
- [x] `RoutedMessage` 支持来源标识和上下文
- [x] 单元测试通过
- [x] TUI Source 具体实现集成
- [x] Channel Source 具体实现集成
- [x] 集成测试通过

---

## M7-007: Task CLI 命令

**优先级**: P1 | **需求**: CLI-006, CLI-007
**架构映射**: cmd/pc/task/, internal/task/client.go

### 实现步骤

1. 实现 `pc task list`
2. 实现 `pc task add <title>`
3. 实现 `pc task complete <id>`
4. 实现 `pc task delete <id>`
5. 统一输出格式（表格或 JSON）
6. 编写表格驱动测试

### 验收标准

- [x] `pc task list` 显示任务列表
- [x] `pc task add` 创建新任务
- [x] `pc task complete` 标记完成
- [x] `pc task delete` 删除任务 (stub)
- [x] 所有命令在 TUI 模式可用
- [x] 单元测试通过

---

## M7-008: 集成测试与文档

**优先级**: P1 | **需求**: 质量保障

### 实现步骤

1. 编写集成测试
2. 性能测试
3. 文档更新

### 验收标准

- [x] 集成测试覆盖主要场景
- [x] 文档完整可执行
- [x] 示例配置可运行
