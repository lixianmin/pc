---
milestone: M3
title: LLM 插件 + Channel 插件
status: ✅ 已完成
completed: 2026-02-26
---

# M3: LLM 插件 + Channel 插件

> OpenAI LLM 插件、Telegram Channel 插件、核心引擎消息处理

---

## M3-001: OpenAI LLM 插件

**优先级**: P0 | **需求**: LLP-006, LLP-004, LLP-005

### 实现步骤

1. 创建插件目录结构：~/.pc/plugins/llm/openai/
2. 编写 plugin.yml 元信息
3. 编写 config.yml 配置模板
4. 实现 openai-llm 可执行文件
5. 实现 complete 方法（文本生成）
6. 实现 stream 方法（流式输出）
7. 实现 models 方法（获取模型列表）
8. 实现 stdio 协议通信
9. 编写插件测试

### 验收标准

- [x] 插件正确加载和调用
- [x] complete 方法正常工作
- [ ] stream 方法正常工作（可选）
- [ ] 集成测试通过

---

## M3-002: Telegram Channel 插件

**优先级**: P0 | **需求**: CHP-004, CHP-005, CHP-006, CHP-007

### 实现步骤

1. 创建插件目录结构：~/.pc/plugins/channel/telegram/
2. 编写 plugin.yml 元信息
3. 编写 config.yml 配置模板（bot_token）
4. 实现 telegram-bot 可执行文件
5. 实现 start 方法（启动 Bot）
6. 实现 stop 方法（停止 Bot）
7. 实现 send 方法（发送消息）
8. 实现 listen 方法（监听消息并转发给核心）
9. 实现消息确认和重试机制
10. 编写插件测试

### 验收标准

- [x] 插件正确加载和调用
- [x] Bot 正常启动和停止
- [x] 消息收发正常
- [ ] 集成测试通过

---

## M3-003: 核心引擎消息处理

**优先级**: P0 | **需求**: ACD-001, ACD-002

### 实现步骤

1. 定义 CoreEngine 接口
2. 实现消息接收和分发逻辑
3. 实现会话上下文管理
4. 实现 Agent 响应生成（调用 LLM 插件）
5. 实现响应回传给 Channel
6. 编写表格驱动测试

### 验收标准

- [x] 消息正确处理
- [x] 会话上下文正确维护
- [x] LLM 插件集成
- [x] Channel 插件集成（架构支持）
- [x] 单元测试通过
