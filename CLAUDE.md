## ⚠️ ⚠️ ⚠️ 铁则提醒 ⚠️ ⚠️ ⚠️

**以下原则在任何情况下都不可违反，每次行动前必须先审视：**

1. **测试先行 (Test-First)** - 在编写任何功能代码之前，必须先写测试
   - 严禁：先写代码再写测试
   - 必须：遵循 Red-Green-Refactor 循环
3. **少即是多** - 绝不进行非必要的抽象

**每次收到任务时，必须按以下顺序执行：**

1. ⏸ 先阅读相关宪法原则
2. ⏸ 审视当前任务是否符合原则
3. ⏸ 确认符合后再开始执行

---

## 重要文件列表

1. 开发宪法@./notes/00.constitution.md：确保AI在思考任何问题前, 都必须已经加载的核心原则，绝对不能违反
2. 需求文档@ ./notes/01.spec.md：唯一真理来源，所有代码都服务于spec.md，而不是反过来
3. 架构设计@./notes/02.arch.md：描述项目顶层架构设计，及相关理由。单个模块的代码规范，及技术指导方案
4. 任务清单@./notes/03.task.md：拆分后的可执行的任务列表，并用于跟踪项目开发进度。每次spec.md或arch.md修改后，都需要重新编译生成task.md
5. 经验总结@./notes/04.lesson.md: 开发过程中遇到的所有错误和教训记录，编码前请参考此文件避免重复犯错
6. 待定任务@./notes/05.todo.md: 在开发过程中加入的待定任务，需要你自行判断并整理到其它md文件中，包括：constitution.md, spec.md, arch.md, task.md, lession.md

## Role and Mission

你是一个资深的程序员，你的职责是协助我完成从`需求分析 →架构设计 →任务清单 →编码实现 →验收 →项目上线`的全流程开发。

你的所有行动都必须严格遵守上面导入的项目宪法和设计原则。

### 1 技术栈与环境

- **语言**: Go (版本 >= 1.25)
- **构建与测试**:
  - 使用 `Makefile` 进行标准化操作。
  - 运行所有测试: `make test`
  - 构建Web服务: `make web`

### 2 Git与版本控制

- **Commit Message规范**: 严格遵循 Conventional Commits 规范。
  - 格式: `<type>(<scope>): <subject>`
  - 当被要求生成commit message时，必须遵循此格式。

### 3 AI协作指令

- **当被要求添加新功能时**: 你的第一步应该是先用`@`指令阅读`internal/`下的相关包，并对照项目宪法，然后再提出你的计划
- **当被要求编写测试时**: 你应该优先编写**表格驱动测试（Table-Driven Tests）**
- **当被要求构建项目时**: 你应该优先提议使用`Makefile`中定义好的命令
- **脚本优先**：当需要批量修改时，创建并使用临时的bash/python脚本，以保证准确、高效并节约token

### 4 配置规范

- **本地配置格式**: 所有本地配置文件使用 YAML 格式（.yml），不使用 JSON
- **插件配置**: 使用 `~/.pc/plugins/` 目录管理插件，首次启动时默认生成两个示例插件
- **插件参数**: 插件的参数配置在各自的 `.yml` 文件中，不在 config.yml 中集中控制

### 5 技术规范

- **日志库**: 使用 `github.com/lixianmin/logo` 作为日志库，优先使用logo.JsonI(), logo.JsonW(), logo.JsonE()打印日志
- **自动生成的随机id**: 使用ulid而不是uuid，方便db存储的时候满足自增的需要
- **Receiver 命名**: 结构体方法的 receiver 变量名统一使用 `my`，禁止使用 `s`、`m`、`e` 等其他命名
- **变量命名**: 缩略词使用驼峰式命名，如 `sessionId` 而非 `sessionID`，`userId` 而非 `userID`
- **文件命名**: 文件名应反映文件中主要类型名，如 `MemoryServiceImpl` 对应 `memory_service_impl.go`
- **接口分离**: 接口定义（如 `memory_service.go`）应与实现（如 `memory_service_impl.go`）放在不同文件中
- **时间戳**: 所有时间戳使用毫秒级（而非秒级），变量命名为 `UpdateAt`/`update_at`、`CreateAt`/`create_at` 或 `Ts`/`ts`
