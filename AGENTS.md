# AGENTS.md

## 何时读什么

| 触发条件 | 读什么 |
|----------|--------|
| 开始任何任务 | [constitution.md](./docs/00.constitution.md) |
| 写代码前 | [architecture.md 第六章](./docs/01.architecture.md#六编码规范) |
| 遇到 bug 或失败 | [lesson.md](./docs/02.lesson.md) |

## 项目约定

| 约定 | 值 |
|------|-----|
| 语言 | Go 1.25+ |
| 配置 | YAML |
| 接收者 | `my` |
| Goroutines | `loom.Go()` |

## Build

```bash
make build              # 构建主程序
make test               # 运行所有测试
make lint               # golangci-lint
make generate           # 生成 BAML 客户端
baml test <func>        # 运行 BAML 测试
```
