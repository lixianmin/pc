---
milestone: M1
title: 核心框架
status: ✅ 已完成
completed: 2026-02-25
---

# M1: 核心框架

> 项目基础结构、日志系统、配置管理、stdio 协议、公共类型

---

## M1-001: 项目基础结构

**优先级**: P0 | **需求**: CF-006, CF-001, CF-002

### 实现步骤

1. 创建项目目录结构（cmd/, internal/, pkg/, examples/）
2. 创建 go.mod，设置 Go >= 1.25
3. 添加 github.com/lixianmin/logo 依赖
4. 创建 Makefile（test、build、clean）
5. 编写单元测试验证项目结构

### 验收标准

- [x] 目录结构完整
- [x] go build 编译通过
- [x] make test 执行成功

---

## M1-002: 日志系统

**优先级**: P0 | **需求**: CF-005

### 实现步骤

1. 在 internal/config/ 定义日志配置结构
2. 初始化 logo 实例，支持控制台和文件输出
3. 实现日志级别控制（debug、info、warn、error）
4. 创建日志中间件支持自定义 hook
5. 编写表格驱动测试

### 验收标准

- [x] 支持日志级别配置
- [x] 支持控制台和文件双输出
- [x] 单元测试通过

---

## M1-003: 配置管理系统

**优先级**: P0 | **需求**: CF-001, CF-002, CF-003

### 实现步骤

1. 定义 Config 结构体（Agent、Workspace、LogLevel）
2. 实现 YAML 配置文件加载
3. 实现环境变量覆盖机制
4. 创建配置验证函数
5. 编写表格驱动测试

### 验收标准

- [x] 支持 YAML 格式配置文件
- [x] 支持环境变量覆盖
- [x] 配置验证正确
- [x] 单元测试通过

---

## M1-004: stdio 协议实现

**优先级**: P0 | **需求**: LLP-001, WSP-001, CHP-001, TLP-003

### 实现步骤

1. 在 pkg/protocol/ 定义消息结构体（Request、Response）
2. 定义 Protocol 接口（Connect/Call/Close）
3. 实现 StdioProtocol（JSON 编解码）
4. 实现消息 ID 生成和匹配
5. 实现超时处理和错误传播
6. 编写表格驱动测试

### 验收标准

- [x] Protocol 接口定义正确
- [x] StdioProtocol 实现
- [x] 支持 JSON 消息格式
- [x] 单元测试通过

---

## M1-005: 公共类型定义

**优先级**: P0 | **需求**: ALC-002, ALC-003, SKL-001

### 实现步骤

1. 在 pkg/types/ 定义插件类型常量
2. 定义 Agent 结构体
3. 定义 Plugin 元信息结构体
4. 定义 Skill 结构体
5. 定义任务状态常量
6. 编写单元测试

### 验收标准

- [x] 核心类型定义完整
- [x] JSON/YAML 序列化支持
- [x] 单元测试通过
