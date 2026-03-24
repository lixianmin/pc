---
name: todo-dispatch
description: 处理 `docs/03.todo.md` 中的待办项。简单任务直接执行，复杂任务用 `writing-plans` skill 规划。
---

# Todo Dispatch

## 触发条件

用户说"整理 todo.md"、"处理 todo"或类似指令。

## 工作流

```
读取 docs/03.todo.md
       │
       ▼
┌─────────────────┐
│  对每一项分类    │
└─────────────────┘
       │
       ├─ 简单任务 ──→ 直接执行 ──┐
       │                        │
       └─ 复杂任务 ──→ writing-plans ─┘
                                │
                                ▼
                         清空 todo.md
```

## 任务分类

| 类型 | 示例 | 处理 |
|------|------|------|
| 简单 | 代码清理、文档修改、小 bugfix | 直接处理 |
| 复杂 | 新功能、架构变更、重构 | `writing-plans` skill |

## 清空

所有项处理完成后：

```markdown
# 临时想法收集器

```
