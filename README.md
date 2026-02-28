# PersonalClaw (pc)

PersonalClaw (PC) is an open-source, self-hosted AI agent platform built in Go. It serves as an AI Agent Operating System Kernel, providing infrastructure for running autonomous digital entities.

## Features

- **Plugin Architecture**: Extensible plugin system supporting LLM providers, channels, tools, and skills
- **Multi-Channel Support**: Interact via TUI (Terminal User Interface) or messaging platforms (Telegram, Discord, etc.)
- **Task Management**: Built-in task scheduling and management via `task_list.md`
- **Autonomous Operation**: Agent can decompose goals and execute tasks autonomously
- **Self-Evolution**: Capable of code analysis, modification, and self-improvement

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/lixianmin/pc.git
cd pc

# Build the binary
make build

# Or install directly
go install ./cmd/pc
```

### Initial Setup

Run `pc` for the first time to start the initialization wizard:

```bash
pc
```

The wizard will guide you through:
- Setting up Agent name (default: PersonalClaw)
- Configuring workspace path (default: ~/workspace)
- Creating example plugins

### Starting the Gateway

PersonalClaw runs as a daemon (gateway) that maintains the core engine:

```bash
# Start the gateway daemon
pc gateway start

# Check status
pc gateway status

# Stop the gateway
pc gateway stop
```

### Using the TUI

Once the gateway is running, launch the interactive Terminal User Interface:

```bash
pc tui
```

**Available Commands in TUI:**
- `/quit` or `/q` - Exit the TUI
- `/clear` or `/c` - Clear the screen
- `/skills` or `/s` - List available skills
- `/status` - Show gateway status
- `/task list` - List tasks
- `/task add <title>` - Add a new task
- `/task complete <id>` - Complete a task
- `/help` or `/h` - Show help

**Special Features:**
- `@skillname` - Reference a skill (injects skill description)
- `@filepath` - Reference a file (reads content into context)
- Tab completion for commands and skills
- Command history (Up/Down arrows)

### Task Management CLI

Manage tasks directly from command line:

```bash
# List all tasks
pc task list

# List pending tasks
pc task list --status pending

# Add a new task
pc task add "Implement feature X"

# Complete a task
pc task complete task-001

# Delete a task
pc task delete task-001
```

## Architecture

PersonalClaw follows a kernel-like architecture:

```
┌─────────────────────────────────────────────────────────┐
│                      CLI Layer (cmd/pc)                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   gateway    │  │     tui      │  │     task     │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────┐
│                   Gateway (internal/gateway)             │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │   Daemon     │  │  RPC Server  │  │ Input Sources│  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────┐
│                   Core Engine (internal/)                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │    Engine    │  │PluginManager │  │AgentManager  │  │
│  └──────────────┘  └──────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────┘
                           │
┌─────────────────────────────────────────────────────────┐
│                      Plugins                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────┐ │
│  │  LLM     │  │ Channel  │  │   Tool   │  │ Skill  │ │
│  │ (OpenAI) │  │(Telegram)│  │ (Linux)  │  │(Custom)│ │
│  └──────────┘  └──────────┘  └──────────┘  └────────┘ │
└─────────────────────────────────────────────────────────┘
```

## Configuration

Configuration is stored in `~/.pc/config.yml`:

```yaml
agent:
  name: PersonalClaw
  profession: 通用助手
  personality:
    - 友好
    - 专业

workspace: ~/workspace
log_level: info
```

## Plugin Development

Plugins are stored in `~/.pc/plugins/` with the following structure:

```
~/.pc/plugins/
├── llm/
│   └── openai/
│       ├── plugin.yml    # Plugin metadata
│       ├── config.yml    # Plugin configuration
│       └── bin/openai    # Plugin executable
└── channel/
    └── telegram/
        ├── plugin.yml
        ├── config.yml
        └── bin/telegram
```

### Plugin Metadata (plugin.yml)

```yaml
name: openai-llm
type: llm
enabled: true
version: 1.0.0
entry: ./bin/openai
```

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests for specific package
go test ./internal/gateway/... -v
```

### Project Structure

```
pc/
├── cmd/pc/              # CLI commands
├── internal/            # Internal packages
│   ├── agent/          # Agent management
│   ├── config/         # Configuration
│   ├── engine/         # Core engine
│   ├── gateway/        # Gateway daemon
│   ├── plugin/         # Plugin management
│   ├── skill/          # Skill system
│   ├── task/           # Task management
│   ├── tui/            # Terminal UI
│   └── wizard/         # Init wizard
├── pkg/                # Public packages
│   ├── protocol/       # Communication protocols
│   └── types/          # Shared types
├── examples/           # Example plugins
└── notes/              # Documentation
```

## License

MIT License - see LICENSE file for details.

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) for details.

## Acknowledgments

PersonalClaw is inspired by:
- The Linux kernel architecture
- Claude Code by Anthropic
- Open-source AI agent frameworks
