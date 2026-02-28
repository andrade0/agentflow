# AgentFlow

**Superpowers for Everyone** — An open-source agentic workflow framework that brings structured AI development to developers using free and local models.

## Features

- 🤖 **Universal Model Support** — Ollama, Groq, Together AI, any OpenAI-compatible API
- 🎯 **Composable Skills** — Brainstorming, planning, TDD, debugging, code review
- 👥 **Subagent Workflows** — Fresh agents per task, parallel execution
- ⚡ **Fast & Lightweight** — Single Go binary, <50ms startup, ~20MB memory

## Installation

```bash
# From source
go install github.com/agentflow/agentflow/cmd/agentflow@latest

# Or build locally
git clone https://github.com/agentflow/agentflow
cd agentflow
go build -o agentflow ./cmd/agentflow
```

## Quick Start

```bash
# Initialize configuration
agentflow config init

# Run a simple query
agentflow run "Explain the factory pattern"

# Use a specific model
agentflow run -m ollama/llama3.3:70b "Write a hello world in Go"

# Stream the response
agentflow run -s "Tell me a story"

# List available skills
agentflow skill list

# Run with a skill
agentflow skill run brainstorming "Design a user auth system"

# Spawn a subagent for a task
agentflow subagent "Implement the login endpoint"
```

## Configuration

Create `~/.agentflow/config.yaml`:

```yaml
providers:
  ollama:
    base_url: http://localhost:11434
    models:
      - llama3.3:70b
      - codellama:34b
  
  groq:
    api_key: ${GROQ_API_KEY}
    models:
      - llama-3.3-70b-versatile
      - mixtral-8x7b-32768
  
  together:
    api_key: ${TOGETHER_API_KEY}
    models:
      - meta-llama/Llama-3.3-70B-Instruct-Turbo

defaults:
  main: groq/llama-3.3-70b-versatile
  subagent: ollama/llama3.3:70b
  reviewer: together/meta-llama/Llama-3.3-70B-Instruct-Turbo

skills:
  paths:
    - skills
    - ~/.agentflow/skills
```

## Skills

Skills are markdown files with YAML front-matter:

```markdown
---
name: my-skill
description: What this skill does
tags:
  - coding
  - review
---

# Skill Content

Instructions for the AI...
```

Place skills in `skills/` or `~/.agentflow/skills/`.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        AgentFlow CLI                         │
├──────────────────────────────────────────────────────────────┤
│  Commands: run | skill | config | subagent | providers      │
└─────────────────────────┬────────────────────────────────────┘
                          │
┌─────────────────────────┴────────────────────────────────────┐
│                      Core Engine                             │
├──────────────┬───────────────┬───────────────┬───────────────┤
│ Skill Runner │ Subagent Pool │ Model Router  │ Context Mgr   │
└──────────────┴───────────────┴───────────────┴───────────────┘
                          │
┌─────────────────────────┴────────────────────────────────────┐
│                    Provider Adapters                         │
├──────────┬──────────┬──────────┬─────────────────────────────┤
│  Ollama  │   Groq   │ Together │     OpenAI-compatible       │
└──────────┴──────────┴──────────┴─────────────────────────────┘
```

## Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build
go build -o agentflow ./cmd/agentflow

# Run
./agentflow --help
```

## License

MIT

## Credits

Inspired by [Superpowers](https://github.com/obra/superpowers) by Jesse Vincent.
