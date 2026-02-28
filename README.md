# 🚀 AgentFlow

**Superpowers for everyone.** An open-source agentic coding tool that lives in your terminal — like Claude Code, but for free and local models.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)

---

## Why AgentFlow?

Claude Code is amazing, but requires an Anthropic subscription. AgentFlow brings the same powerful experience to everyone:

| Feature | Claude Code | AgentFlow |
|---------|-------------|-----------|
| Models | Claude only | Ollama, Groq, Together, any model |
| Cost | $20+/month | Free (local) or cheap |
| Privacy | Cloud API | Run fully offline |
| Open Source | No | Yes, MIT licensed |

## Features

### 🖥️ Full Terminal UI

```
🚀 AgentFlow v0.1.0
ollama/llama3.3 • Ctrl+Enter to send • /help for commands

You 14:32
build a REST API for users

⚡ Skill: brainstorming

Agent 14:32 ●
Before I start coding, I have some questions...

╭──────────────────────────────────────────────────────────╮
│ Type a message...                                        │
╰──────────────────────────────────────────────────────────╯
┌──────────────────────────────────────────────────────────┐
│ ollama/llama3.3 │ 1.2k tokens │ $0.00 │ ↑5 msgs • 3m    │
└──────────────────────────────────────────────────────────┘
```

### 📚 Claude Code-Compatible Features

- **Session persistence** — Save and resume conversations
- **Slash commands** — /help, /model, /compact, /export, /vim...
- **Keyboard shortcuts** — Ctrl+R search, Ctrl+B background, vim mode
- **Multiline input** — Option+Enter, Shift+Enter
- **Autocomplete** — Tab for commands and files
- **Background tasks** — Run long commands async
- **Token tracking** — Know your context usage
- **Cost estimation** — Track spending
- **Themes** — Customize your experience

### 🧠 Composable Skills

Built-in skills for structured workflows:

- **brainstorming** — Mandatory design before coding
- **writing-plans** — 2-5 minute task breakdown
- **subagent-driven-development** — Fresh agents per task
- **test-driven-development** — RED-GREEN-REFACTOR
- **systematic-debugging** — 4-phase root cause analysis
- **verification-before-completion** — Evidence before claims

## Installation

### From Source

```bash
git clone https://github.com/andrade0/agentflow.git
cd agentflow
go build -o agentflow ./cmd/agentflow
sudo mv agentflow /usr/local/bin/
```

### From Go

```bash
go install github.com/andrade0/agentflow/cmd/agentflow@latest
```

## Quick Start

### 1. Configure a Provider

```bash
# Create config
mkdir -p ~/.agentflow
cat > ~/.agentflow/config.yaml << 'EOF'
providers:
  ollama:
    base_url: http://localhost:11434
    models: [llama3.3:70b, codellama:34b]
  groq:
    api_key: ${GROQ_API_KEY}
    models: [llama-3.3-70b-versatile]

defaults:
  main: ollama/llama3.3:70b
EOF
```

### 2. Start AgentFlow

```bash
# Start interactive session
agentflow

# Or with initial prompt
agentflow "explain this project"
```

## CLI Commands

```bash
# Interactive mode (default)
agentflow                      # Start TUI
agentflow "task"               # Start with prompt

# Session management
agentflow -c                   # Continue last session
agentflow -r <id|name>         # Resume specific session
agentflow --fork-session       # Fork when resuming

# Non-interactive
agentflow run "task"           # Execute and exit
agentflow -p "task"            # Print mode (for scripts)
cat file | agentflow -p "explain"  # Pipe content

# Configuration
agentflow config init          # Create .agentflow/
agentflow config show          # Show config

# Skills & Subagents
agentflow skill list           # List skills
agentflow agents               # List subagents
```

## Slash Commands

| Command | Description |
|---------|-------------|
| `/help` | Show all commands |
| `/quit`, `/exit` | Exit session |
| `/clear` | Clear conversation |
| `/compact [focus]` | Compact context |
| `/model [name]` | Show/change model |
| `/status` | Session statistics |
| `/cost` | Token usage & costs |
| `/context` | Visualize context |
| `/sessions` | List saved sessions |
| `/resume [id]` | Resume session |
| `/rename [name]` | Rename session |
| `/export [file]` | Export conversation |
| `/copy` | Copy last response |
| `/skills` | List skills |
| `/vim` | Toggle vim mode |
| `/theme` | Change theme |

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel / Exit |
| `Ctrl+L` | Clear screen |
| `Ctrl+R` | Reverse search history |
| `Ctrl+B` | Background running task |
| `Ctrl+T` | Toggle task list |
| `Ctrl+O` | Toggle verbose mode |
| `Up/Down` | Navigate history |
| `PgUp/PgDown` | Scroll viewport |
| `Option+Enter` | Multiline input |
| `Tab` | Autocomplete |
| `Esc+Esc` | Rewind conversation |

### Vim Mode

Enable with `/vim`. Full vim keybindings:
- `hjkl` navigation
- `w`, `b`, `e` word movement
- `dd`, `yy`, `p` edit operations
- `iw`, `aw`, `i"` text objects
- `i`, `a`, `o` insert modes

### Bash Mode

Prefix with `!` to run commands directly:
```
!git status
!npm test
!docker ps
```

## Configuration

### Global Config (~/.agentflow/config.yaml)

```yaml
providers:
  ollama:
    base_url: http://localhost:11434
    models: [llama3.3:70b, codellama:34b, deepseek-coder:33b]
  
  groq:
    api_key: ${GROQ_API_KEY}
    models: [llama-3.3-70b-versatile, mixtral-8x7b-32768]
  
  together:
    api_key: ${TOGETHER_API_KEY}
    models: [meta-llama/Llama-3.3-70B-Instruct-Turbo]

defaults:
  main: groq/llama-3.3-70b-versatile
  subagent: ollama/llama3.3:70b
  reviewer: together/Qwen/Qwen2.5-Coder-32B-Instruct

skills:
  paths:
    - ./skills
    - ~/.agentflow/skills

session:
  auto_save: true
  max_sessions: 50
```

### Project Config (.agentflow/config.yaml)

```yaml
project:
  name: my-api
  language: go
  test_command: go test ./...
  lint_command: golangci-lint run
```

## Skills

Skills are markdown files that define workflows:

```markdown
---
name: my-skill
description: "When to use this skill"
triggers: ["keyword1", "keyword2"]
priority: 50
---

# My Skill

## Process
1. Step one
2. Step two
...
```

Place in `./skills/` or `~/.agentflow/skills/`.

## Model Recommendations

| Use Case | Model | Provider |
|----------|-------|----------|
| General coding | llama-3.3-70b | Groq (free) |
| Code generation | codellama:34b | Ollama |
| Code review | Qwen2.5-Coder-32B | Together |
| Fast responses | llama-3.2-3b | Ollama |

## Roadmap

- [x] Interactive TUI
- [x] Streaming responses
- [x] Skill system
- [x] Subagent support
- [ ] Session persistence
- [ ] Vim mode
- [ ] Background tasks
- [ ] MCP integration
- [ ] IDE plugins

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Credits

- [Jesse Vincent](https://github.com/obra) for [Superpowers](https://github.com/obra/superpowers)
- [Anthropic](https://anthropic.com) for Claude Code inspiration
- The open-source LLM community

## License

MIT License — Use it, fork it, build cool stuff.

---

**Star ⭐ if you find this useful!**

[Documentation](docs/) · [Issues](https://github.com/andrade0/agentflow/issues)
