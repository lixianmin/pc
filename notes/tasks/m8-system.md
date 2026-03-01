---
milestone: M8
title: System Prompt 优化
status: ✅ 已完成
completed: 2026-02-28
---

# M8: System Prompt 优化

> 加载 agents.md 作为 system prompt，构建完整 system prompt

---

## M8-001: 加载 agents.md 作为 system prompt

**优先级**: P1 | **需求**: CLI-012
**架构映射**: internal/agent/

### 实现步骤

1. 定义 agents.md 文件格式规范（Markdown 格式）
2. 实现 agents.md 文件解析器
3. 在 Agent 初始化时加载 agents.md
4. 将 agents.md 内容作为基础 system prompt
5. 编写表格驱动测试

### 验收标准

- [x] agents.md 文件正确解析
- [x] 内容正确加载为 system prompt
- [x] 单元测试通过

---

## M8-002: 构建完整 system prompt

**优先级**: P1 | **需求**: CLI-013
**架构映射**: internal/agent/, internal/skill/, internal/plugin/

### 实现步骤

1. 设计 system prompt 构建流程
2. 实现 skills 描述收集（从 skill 目录）
3. 实现工具列表说明生成（从 plugin manager）
4. 整合：agents.md + skills + 工具列表 → 完整 system prompt
5. 实现 LLM 调用时自动注入完整 system prompt
6. 编写表格驱动测试

### 验收标准

- [x] system prompt 包含 agents.md 内容
- [x] system prompt 包含 skills 说明
- [x] system prompt 包含可用工具列表
- [x] LLM 交互时正确注入完整 system prompt
- [x] 单元测试通过
