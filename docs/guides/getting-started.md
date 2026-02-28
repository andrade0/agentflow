# Getting Started with AgentFlow

This guide will get you up and running with AgentFlow in under 5 minutes.

## Prerequisites

- **Go 1.22+** (for building from source)
- **A model provider**: Ollama (local), Groq (free), or any OpenAI-compatible API

## Installation

### Option 1: Binary Download (Recommended)

```bash
# macOS/Linux
curl -fsSL https://agentflow.dev/install.sh | sh

# Or with Homebrew
brew install agentflow/tap/agentflow

# Verify installation
agentflow --version
```

### Option 2: From Source

```bash
git clone https://github.com/agentflow/agentflow.git
cd agentflow
go build -o agentflow ./cmd/agentflow
sudo mv agentflow /usr/local/bin/
```

## Set Up a Model Provider

AgentFlow needs an LLM to run. Choose one:

### Ollama (Local, Free, Private)

Best for: Privacy, no rate limits, offline use.

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model (llama3.3 70B recommended)
ollama pull llama3.3:70b

# Or use a smaller model for faster responses
ollama pull llama3.2:3b
```

Configure AgentFlow:
```bash
agentflow config set provider ollama
agentflow config set model llama3.3:70b
```

### Groq (Cloud, Free Tier)

Best for: Fast inference, no GPU needed.

1. Get a free API key at [console.groq.com](https://console.groq.com)
2. Set your key:

```bash
export GROQ_API_KEY=gsk_your_key_here

# Or add to ~/.zshrc or ~/.bashrc
echo 'export GROQ_API_KEY=gsk_your_key_here' >> ~/.zshrc

agentflow config set provider groq
agentflow config set model llama-3.3-70b-versatile
```

### Together AI (Cloud, Pay-per-use)

Best for: Access to many models, good quality.

```bash
export TOGETHER_API_KEY=your_key_here
agentflow config set provider together
agentflow config set model meta-llama/Llama-3.3-70B-Instruct-Turbo
```

## Initialize Your Project

Navigate to your project and initialize:

```bash
cd your-project
agentflow init
```

This creates:
```
.agentflow/
├── config.yaml     # Project-specific config
├── context/        # Auto-generated project context
└── skills/         # Custom skills (optional)
```

## Your First Task

Ask AgentFlow to do something:

```bash
agentflow run "add a health check endpoint to the API"
```

AgentFlow will:
1. **Brainstorm** — Ask clarifying questions, understand context
2. **Design** — Propose approaches, get your approval
3. **Plan** — Break work into small tasks
4. **Execute** — Run each task with fresh subagents
5. **Verify** — Ensure everything works

## Interactive Mode

For ongoing conversations:

```bash
agentflow tui
```

This opens an interactive terminal UI where you can:
- Chat naturally with AgentFlow
- See task progress in real-time
- Approve/reject designs
- View subagent activity

## CLI Commands

```bash
# Run a task
agentflow run "build X"

# Interactive mode
agentflow tui

# List available skills
agentflow skill list

# Run a specific skill
agentflow skill run brainstorming

# Show current config
agentflow config show

# Set a config value
agentflow config set model llama3.3:70b
```

## Configuration

### Global Config (~/.agentflow/config.yaml)

```yaml
defaults:
  provider: ollama
  model: llama3.3:70b

providers:
  ollama:
    base_url: http://localhost:11434
  groq:
    api_key: ${GROQ_API_KEY}
  together:
    api_key: ${TOGETHER_API_KEY}
```

### Project Config (.agentflow/config.yaml)

```yaml
project:
  name: my-api
  language: go
  test_command: go test ./...
  lint_command: golangci-lint run

# Override global defaults
defaults:
  model: codellama:34b
```

## What's Next?

- [Understanding Skills](./skills.md) — How AgentFlow decides what to do
- [Subagent Development](./subagent-development.md) — How tasks get executed
- [Custom Skills](./custom-skills.md) — Create your own workflows
- [Model Selection](./model-selection.md) — Choosing the right model

## Troubleshooting

### "Cannot connect to Ollama"

```bash
# Check if Ollama is running
curl http://localhost:11434/api/tags

# If not, start it
ollama serve
```

### "Rate limit exceeded" (Groq)

Groq's free tier allows 30 requests/minute. Either:
- Wait a minute
- Use Ollama for subagents (unlimited)
- Upgrade to Groq paid tier

### "Model not found"

```bash
# List available models
ollama list  # For Ollama
agentflow models  # For all providers

# Pull the model you need
ollama pull llama3.3:70b
```

## Getting Help

- [GitHub Issues](https://github.com/agentflow/agentflow/issues)
- [Discord](https://discord.gg/agentflow)
- [Documentation](https://agentflow.dev/docs)
