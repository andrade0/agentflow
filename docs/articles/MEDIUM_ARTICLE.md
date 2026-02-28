# AgentFlow: I Built a Free, Open-Source Alternative to Claude Code

*No API keys. No subscriptions. No cloud. Just you and your local LLM.*

---

I love Claude Code. The structured workflows, the brainstorming before coding, the subagent architecture — it's genuinely changed how I work.

But I have a problem: **I don't want to pay $20/month** to code with an AI assistant. And I don't want my code going to a cloud API.

So I built AgentFlow.

## What is AgentFlow?

AgentFlow is a **terminal-based AI coding assistant** that works exactly like Claude Code, but runs entirely on open-source models. No API keys required.

```
🚀 AgentFlow v0.1.0
ollama/llama3.3:70b • /help for commands

You > build a REST API for user authentication

⚡ Skill: brainstorming

Agent > Before I start coding, I have some questions:

1. What authentication method do you prefer? (JWT, sessions, OAuth)
2. Do you need refresh tokens?
3. Where should user data be stored?
```

It's Claude Code for the rest of us.

## Why I Built This

Three reasons:

### 1. Cost

Claude Code requires an Anthropic API subscription. That's $20/month minimum, and it adds up fast if you're coding all day.

AgentFlow costs **$0**. Run Llama 3.3 locally with Ollama, and you're set.

### 2. Privacy

With Claude Code, every line of your code goes to Anthropic's servers. For personal projects, that's fine. For work? Maybe not.

AgentFlow runs **100% locally**. Your code never leaves your machine.

### 3. Open Source

I believe the future of AI coding assistants should be open. AgentFlow is MIT licensed — fork it, modify it, sell it, I don't care.

## Features

### Full Terminal UI

AgentFlow isn't a simple CLI. It's a complete terminal UI built with [Bubbletea](https://github.com/charmbracelet/bubbletea):

- Streaming responses with real-time output
- Status bar showing model, tokens, session time
- Vim mode for power users
- Syntax highlighting for code blocks
- Command history with Ctrl+R reverse search

### Session Persistence

Close your terminal, come back tomorrow — your conversation is still there.

```bash
agentflow -c              # Continue last session
agentflow -r my-feature   # Resume session by name
agentflow sessions        # List all sessions
```

### Composable Skills

Skills are markdown files that define structured workflows. AgentFlow ships with:

- **brainstorming** — Mandatory design before coding
- **writing-plans** — Break work into 2-5 minute tasks
- **subagent-driven-development** — Fresh agents per task
- **test-driven-development** — RED-GREEN-REFACTOR enforcement
- **systematic-debugging** — 4-phase root cause analysis

### Bash Mode

Prefix with `!` to run commands directly:

```
You > !git status
You > !npm test
You > !docker ps
```

Output gets added to the conversation context automatically.

### Multiline Input

No more awkward one-liners:

- Alt+Enter for new line
- `\` at end of line to continue
- Ctrl+Enter to submit

## Installation

### Quick Start (2 minutes)

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull a model
ollama pull llama3.3:70b

# Install AgentFlow
go install github.com/andrade0/agentflow/cmd/agentflow@latest

# Start coding
agentflow
```

That's it. No API keys, no configuration files, no signup.

### Remote GPU Support

Don't have a GPU? Run the model on a remote server:

```yaml
# ~/.agentflow/config.yaml
providers:
  vllm:
    base_url: http://my-gpu-server:8000/v1
    models: [meta-llama/Llama-3.3-70B-Instruct]
```

AgentFlow works with any OpenAI-compatible API: vLLM, llama.cpp, TGI, LocalAI, LM Studio.

## How It Compares

| Feature | Claude Code | AgentFlow |
|---------|-------------|-----------|
| Models | Claude only | Any open-source model |
| Cost | $20+/month | Free |
| Privacy | Cloud API | 100% local |
| API Keys | Required | None |
| Open Source | No | Yes (MIT) |
| Vim Mode | Yes | Yes |
| Session Save | Yes | Yes |
| Skills | Yes | Yes |
| Subagents | Yes | Yes |

## The Skill System

Skills are what make AgentFlow powerful. Here's the brainstorming skill in action:

```
You > add user authentication

⚡ Skill: brainstorming

Agent > Before I implement authentication, I have questions:

1. Should tokens be short-lived (15min) or long-lived (7 days)?

You > Short with refresh tokens

Agent > Got it. Here are three approaches:

**Approach A: Standard JWT + Refresh**
- Pros: Industry standard, stateless
- Cons: Can't revoke access tokens

**Approach B: JWT + Redis Blacklist**  
- Pros: Can revoke, still fast
- Cons: Adds Redis dependency

**Approach C: Session-based**
- Pros: Simple, revocable
- Cons: Not stateless

I recommend Approach B. Do you approve?

You > yes

[Skill: writing-plans activated]
...
```

The AI can't just start coding. It must:
1. Ask clarifying questions
2. Present multiple approaches
3. Get explicit approval

This eliminates the "guess and fix" loop that wastes so much time.

## Recommended Models

| Use Case | Model | VRAM |
|----------|-------|------|
| Best quality | llama3.3:70b | 48GB |
| Balanced | qwen2.5-coder:14b | 12GB |
| Code-focused | codellama:34b | 24GB |
| Fast/Low VRAM | llama3.2:3b | 4GB |
| CPU only | phi-3:3.8b-q4_0 | 8GB RAM |

## What's Next?

AgentFlow is actively developed. Coming soon:

- [ ] MCP (Model Context Protocol) integration
- [ ] VS Code extension
- [ ] Background tasks
- [ ] Agent teams
- [ ] Web UI

## Try It

```bash
# Install
go install github.com/andrade0/agentflow/cmd/agentflow@latest

# Or build from source
git clone https://github.com/andrade0/agentflow
cd agentflow && go build -o agentflow ./cmd/agentflow
```

Star the repo if you find it useful: [github.com/andrade0/agentflow](https://github.com/andrade0/agentflow)

---

**AgentFlow is proof that you don't need to pay for AI coding assistance.** The open-source models are good enough. The tooling just needed to catch up.

Now it has.

*No API keys. No cloud. No costs. Just code.*

---

## About

I'm building tools for developers who want AI assistance without vendor lock-in. Follow me for more open-source AI projects.

**Tags:** #AI #OpenSource #Coding #LLM #Ollama #Go #CLI #DeveloperTools
