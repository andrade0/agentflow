# 🚀 AgentFlow

**Superpowers for everyone.** An open-source agentic coding tool that lives in your terminal — like Claude Code, but 100% free and open source. No API keys, no subscriptions, no cloud dependencies.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)

---

## Philosophy

**No API keys. No cloud. No costs. Just open source models.**

AgentFlow is designed for developers who want the power of agentic coding assistants without:
- Paying for API subscriptions
- Sending code to cloud services
- Depending on external providers

Run everything locally with Ollama, or connect to your own GPU server running vLLM, TGI, or any OpenAI-compatible endpoint.

## Why AgentFlow?

| Feature | Claude Code | AgentFlow |
|---------|-------------|-----------|
| Models | Claude (proprietary) | Llama, Qwen, DeepSeek, Mistral... |
| Cost | $20+/month | **Free forever** |
| Privacy | Cloud API | **100% local** |
| API Keys | Required | **None needed** |
| Open Source | No | **Yes, MIT licensed** |

## Supported Backends

| Backend | Local | Remote GPU | Setup |
|---------|-------|------------|-------|
| **Ollama** | ✅ | ✅ | `ollama serve` |
| **vLLM** | ✅ | ✅ | `vllm serve model` |
| **llama.cpp** | ✅ | ✅ | `llama-server` |
| **TGI** | ✅ | ✅ | HuggingFace TGI |
| **LocalAI** | ✅ | ✅ | LocalAI server |
| **LM Studio** | ✅ | ❌ | GUI app |
| **Aphrodite** | ✅ | ✅ | Aphrodite Engine |

All backends expose OpenAI-compatible APIs — AgentFlow works with any of them.

## Features

### 🖥️ Full Terminal UI

```
🚀 AgentFlow v0.1.0
ollama/llama3.3:70b • Ctrl+Enter to send • /help for commands

You 14:32
build a REST API for users

⚡ Skill: brainstorming

Agent 14:32 ●
Before I start coding, I have some questions...

╭──────────────────────────────────────────────────────────╮
│ Type a message...                                        │
╰──────────────────────────────────────────────────────────╯
┌──────────────────────────────────────────────────────────┐
│ llama3.3:70b │ 1.2k tokens │ ↑5 msgs • 3m 15s           │
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

### 1. Install Ollama

```bash
# macOS / Linux
curl -fsSL https://ollama.com/install.sh | sh

# Start Ollama
ollama serve

# Pull a model
ollama pull llama3.3:70b
```

### 2. Start AgentFlow

```bash
# Start interactive session
agentflow

# Or with initial prompt
agentflow "explain this project"
```

That's it! No API keys, no configuration needed.

## Configuration (Optional)

### For Local Ollama

Works out of the box — no config needed!

### For Remote GPU Server

Create `~/.agentflow/config.yaml`:

```yaml
providers:
  # Remote Ollama instance
  ollama:
    base_url: http://gpu-server.local:11434
    models: [llama3.3:70b, codellama:34b, deepseek-coder:33b]
  
  # vLLM server
  vllm:
    base_url: http://gpu-server.local:8000/v1
    models: [meta-llama/Llama-3.3-70B-Instruct]
  
  # llama.cpp server
  llamacpp:
    base_url: http://gpu-server.local:8080/v1
    models: [default]
  
  # Text Generation Inference
  tgi:
    base_url: http://gpu-server.local:8080/v1
    models: [default]

defaults:
  main: ollama/llama3.3:70b
  subagent: ollama/codellama:34b
  reviewer: ollama/deepseek-coder:33b

skills:
  paths:
    - ./skills
    - ~/.agentflow/skills
```

## CLI Commands

```bash
# Interactive mode (default)
agentflow                      # Start TUI
agentflow "task"               # Start with prompt

# Session management
agentflow -c                   # Continue last session
agentflow -r <id|name>         # Resume specific session

# Non-interactive
agentflow run "task"           # Execute and exit

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
| `/context` | Visualize context |
| `/sessions` | List saved sessions |
| `/resume [id]` | Resume session |
| `/export [file]` | Export conversation |
| `/skills` | List skills |
| `/vim` | Toggle vim mode |

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+C` | Cancel / Exit |
| `Ctrl+L` | Clear screen |
| `Ctrl+R` | Reverse search history |
| `Ctrl+B` | Background running task |
| `Up/Down` | Navigate history |
| `PgUp/PgDown` | Scroll viewport |
| `Option+Enter` | Multiline input |
| `Tab` | Autocomplete |
| `!command` | Run bash directly |

## Recommended Models

| Use Case | Model | Size | VRAM |
|----------|-------|------|------|
| **General coding** | llama3.3:70b | 40GB | 48GB |
| **Code generation** | codellama:34b | 19GB | 24GB |
| **Code review** | deepseek-coder:33b | 18GB | 24GB |
| **Fast responses** | llama3.2:3b | 2GB | 4GB |
| **Balanced** | qwen2.5-coder:14b | 8GB | 12GB |
| **Low VRAM** | phi-3:3.8b | 2GB | 4GB |

### Running on CPU (Slow but Works)

```bash
# Use smaller quantized models
ollama pull llama3.2:3b-q4_0
ollama pull phi-3:3.8b-q4_0
```

### Running on Remote GPU

Set up vLLM on your GPU server:

```bash
# On GPU server
pip install vllm
vllm serve meta-llama/Llama-3.3-70B-Instruct --port 8000

# In AgentFlow config
providers:
  vllm:
    base_url: http://gpu-server:8000/v1
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

## Roadmap

- [x] Interactive TUI
- [x] Streaming responses
- [x] Skill system
- [x] Subagent support
- [ ] Session persistence
- [ ] Vim mode
- [ ] Background tasks
- [ ] IDE plugins

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Credits

- [Jesse Vincent](https://github.com/obra) for [Superpowers](https://github.com/obra/superpowers)
- [Anthropic](https://anthropic.com) for Claude Code inspiration
- [Ollama](https://ollama.com) for making local LLMs easy
- The open-source LLM community

## License

MIT License — Use it, fork it, build cool stuff.

---

**Star ⭐ if you find this useful!**

**No API keys. No cloud. No costs. Just code.**

[Documentation](docs/) · [Issues](https://github.com/andrade0/agentflow/issues)
