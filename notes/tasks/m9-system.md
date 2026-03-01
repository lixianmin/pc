---
milestone: M9
title: System Prompt 优化增强
status: ✅ 已完成
completed: 2026-03-01
---

# M9: System Prompt 优化增强

> Gateway Restart 命令、SystemPromptBuilder 实际集成、agents.md 配置简化

---

## M9-001: Gateway Restart 命令

**优先级**: P1 | **需求**: CLI-014
**架构映射**: cmd/pc/gateway/, internal/gateway/

### 实现步骤

1. 实现 `pc gateway restart` 命令
2. 复用 stop 逻辑：读取 pidfile、发送 SIGTERM、等待退出、清理 pidfile
3. 复用 start 逻辑：初始化 Gateway、启动 RPC Server、写入 pidfile
4. 确保新编译的版本能被加载
5. 编写表格驱动测试

### 验收标准

- [x] `pc gateway restart` 正确停止并重新启动 Gateway
- [x] 新编译的版本生效
- [x] 单元测试通过

---

## M9-002: SystemPromptBuilder 实际集成

**优先级**: P1 | **需求**: CLI-012, CLI-013
**架构映射**: internal/agent/

### 实现步骤

1. 在 Agent 初始化时创建 SystemPromptBuilder
2. 实现从 config.yml 读取 system_prompt.file 配置
3. 读取 agents.md 文件内容作为 system prompt 前缀
4. 集成 skills 描述收集
5. 集成工具列表说明生成
6. 在 LLM 调用时注入完整 system prompt
7. 编写表格驱动测试

### 验收标准

- [x] SystemPromptBuilder 在 Engine/Agent 中实际使用
- [x] 完整 system prompt 包含 agents.md 内容
- [x] 完整 system prompt 包含 skills 描述
- [x] 完整 system prompt 包含工具列表
- [x] LLM 调用时正确注入 system prompt
- [x] 单元测试通过

---

## M9-003: agents.md 配置简化

**优先级**: P1 | **需求**: CLI-015, CF-009
**架构映射**: internal/config/, internal/agent/

### 实现步骤

1. 在 config.yml 中添加 system_prompt.file 配置项
2. 修改 Config 结构体添加 SystemPromptFile 字段
3. 简化 agents.md 加载逻辑：直接读取文件内容，无需复杂解析
4. 支持默认路径（~/.pc/agents.md）
5. 支持自定义路径（相对或绝对路径）
6. 编写表格驱动测试

### 验收标准

- [x] config.yml 支持配置 agents.md 路径
- [x] 文件路径支持相对和绝对路径
- [x] 简化后的加载逻辑直接读取文件内容
- [x] 默认路径为 ~/.pc/agents.md
- [x] 单元测试通过
