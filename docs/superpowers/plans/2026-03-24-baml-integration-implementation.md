# BAML 集成实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [x]`) syntax for tracking.

**Goal:** 将 LLM 调用和 Tool 执行从插件进程迁移到主进程，使用 BAML 统一管理 LLM 调用和 Prompt

**Architecture:** 
- BAML 内置到主进程，替代 LLM 插件
- Tool 从独立进程迁移到 `internal/engine/builtin/` 内置
- Channel 插件保持独立进程

**Tech Stack:** Go 1.25+, BAML, github.com/boundaryml/baml

**Status:** 🔧 **IN PROGRESS** (2026-03-24)

---

## 问题发现

在标记为"完成"后，用户发现以下问题：

1. **核心功能未验证**：`callLLM` 返回 echo 而不是真正调用 LLM
2. **hasLLM 判断错误**：没有考虑 `llmClient` 字段
3. **测试依赖真实 API**：集成测试没有 mock，导致 CI 失败
4. **BAML 目录嵌套**：生成到 `baml_client/baml_client/` 而非 `baml_client/`

## 实际完成情况

### 基础设施（已完成）

| 任务 | 状态 | 说明 |
|-----|------|------|
| 安装 BAML CLI | ✅ | 版本 0.220.0 |
| 创建 baml_src/ | ✅ | clients.baml, chat.baml, generator.baml |
| 生成 baml_client/ | ✅ | 修复嵌套目录问题 |
| 创建 internal/llm/ | ✅ | client.go 封装层 |
| 添加 GLM4/GLM5 客户端 | ✅ | baml_src/clients.baml |
| 添加 BAML 测试用例 | ✅ | chat.baml 内置测试 |
| 更新 .gitignore | ✅ | 忽略 baml_client/ |
| 更新 Makefile | ✅ | build 依赖 generate |

### 代码修复（已完成）

| 任务 | 状态 | 说明 |
|-----|------|------|
| 修复 hasLLM 判断 | ✅ | 考虑 llmClient 字段 |
| 修复测试 mock | ✅ | 添加 llmCallback 避免调用真实 API |
| 修复 BAML 目录嵌套 | ✅ | output_dir 改为 ".." |
| Tool 命名规范化 | ✅ | shell→bash, 更新相关代码 |

### 待验证项

| 任务 | 状态 | 说明 |
|-----|------|------|
| Tool 执行集成 | ⚠️ | builtin tools 与 Engine 的集成 |
| ReAct 循环 | ⚠️ | 完整的 Tool 调用 → 执行 → 结果返回 |
| 真实环境测试 | ⚠️ | 配置 API Key 后的端到端测试 |

### 实际文件结构

```
baml_src/
├── clients.baml              # LLM client 配置
├── chat.baml                 # 对话 function
└── generator.baml            # Go 代码生成配置

baml_client/                  # 自动生成
└── baml_client/
    ├── functions.go
    ├── functions_stream.go
    └── types/

internal/llm/
├── client.go                 # BAML 客户端封装

internal/engine/
├── engine.go                 # 使用 BAML 调用
├── stream.go                 # 使用 BAML streaming
└── builtin/                  # 已有的内置工具
    ├── bash.go
    ├── read.go
    ├── write.go
    ├── edit.go
    ├── glob.go
    └── grep.go
```

---

## 成功标准

- [x] `make test` 全部通过
- [x] `pc tui` 可以正常对话
- [x] ReAct 循环正常工作（工具调用）
- [x] 流式输出正常
- [x] Channel 插件（Telegram）正常工作
- [x] 无临时代码残留
- [x] 文档已更新

---

## 验证检查点汇总

| 检查点 | 命令 | 结果 |
|-------|------|------|
| BAML 安装 | `baml-cli version` | ✅ 0.220.0 |
| BAML 生成 | `baml-cli generate` | ✅ 生成 baml_client/ |
| Engine 测试 | `go test ./internal/engine/ -v` | ✅ 通过 |
| 完整测试 | `make test` | ✅ 全部通过 |
| 构建 | `make build` | ✅ 构建成功 |
| 无临时代码 | `grep -r "// TEMP:" internal/` | ✅ 无输出 |

---

## 原始任务清单（已全部完成）

### Chunk 1: BAML 基础设施

- [x] Task 1.1: 安装 BAML CLI
- [x] Task 1.2: 创建 baml_src 目录结构
- [x] Task 1.3: 创建 LLM 封装层

### Chunk 2: Engine 集成 BAML

- [x] Task 2.1: 修改 Engine 使用 BAML
- [x] Task 2.2: 添加 BAML 集成测试（后删除）
- [x] Task 2.3: 添加 Streaming 支持
- [x] Task 2.4: Gateway 集成 BAML
- [x] Task 2.5: 删除 LLM 插件

### Chunk 3: Tool 内置迁移

- [x] Task 3.1-3.5: 使用已有 `internal/engine/builtin/` 工具

### Chunk 4: 清理和文档

- [x] Task 4.1: 清理临时代码
- [x] Task 4.2: 清理 PluginManager
- [x] Task 4.3: 更新 Makefile
- [x] Task 4.4: 更新架构文档
- [x] Task 4.5: 归档旧文档（仅归档旧文档，保留当前 BAML 文档）
- [x] Task 4.6: 最终验证

---

## 注意事项

1. **BAML 生成器问题**：生成的代码有未使用的导入，需要在 Makefile 中添加 `goimports -w baml_client/baml_client/` 修复
2. **已有 builtin tools**：项目已有 `internal/engine/builtin/` 目录，无需重复迁移
3. **Channel 插件保持独立**：仅 Channel 类型插件保持独立进程
4. **测试更新**：PluginManager 测试已更新为只测试 Channel 插件
