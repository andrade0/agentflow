# AgentFlow - Superpowers for Everyone

## Vision

**AgentFlow** is an open-source agentic workflow framework that brings the power of structured AI development to developers using free and local models.

Inspired by Jesse Vincent's [Superpowers](https://github.com/obra/superpowers), AgentFlow democratizes sophisticated AI coding workflows by:

1. **Supporting free/local models** — Ollama, Groq (free tier), Together AI, Anthropic, OpenAI, any OpenAI-compatible API
2. **Providing composable skills** — Brainstorming, planning, TDD, debugging, code review
3. **Enabling subagent workflows** — Fresh agents per task, two-stage review (spec + quality)
4. **Including smart automation** — Autonomous continue hooks, verification gates

## Two Implementations

| Feature | agentflow (Go) | agentflow-ts (Bun) |
|---------|---------------|-------------------|
| Runtime | Single binary | Bun/Node |
| Startup | Instant (<50ms) | Fast (<200ms) |
| Memory | ~20MB | ~80MB |
| Plugins | Go plugins or exec | JS/TS native |
| Target | Production, CI/CD | Dev, prototyping |
| Install | `brew install agentflow` | `bunx agentflow-ts` |

## Core Features

### 1. Universal Model Support
```yaml
# ~/.agentflow/config.yaml
providers:
  ollama:
    base_url: http://localhost:11434
    models:
      - llama3.3:70b
      - codellama:34b
      - deepseek-coder:33b
  
  groq:
    api_key: ${GROQ_API_KEY}
    models:
      - llama-3.3-70b-versatile
      - mixtral-8x7b-32768
  
  together:
    api_key: ${TOGETHER_API_KEY}
    models:
      - meta-llama/Llama-3.3-70B-Instruct-Turbo
      - Qwen/Qwen2.5-Coder-32B-Instruct

  anthropic:
    api_key: ${ANTHROPIC_API_KEY}
    models:
      - claude-sonnet-4-20250514

defaults:
  main: groq/llama-3.3-70b-versatile
  subagent: ollama/llama3.3:70b
  reviewer: together/Qwen/Qwen2.5-Coder-32B-Instruct
```

### 2. Superpowers-Compatible Skills

All original Superpowers skills, adapted for multi-model:

- **brainstorming** — Mandatory design phase before coding
- **writing-plans** — 2-5 min task breakdown
- **subagent-driven-development** — Fresh agent per task
- **test-driven-development** — RED-GREEN-REFACTOR enforcement
- **systematic-debugging** — 4-phase root cause analysis
- **verification-before-completion** — Evidence before claims
- **requesting-code-review** — Two-stage: spec + quality
- **using-git-worktrees** — Isolated development branches

### 3. Autonomous Continue (DSL Hook)

```go
// When agent wants to stop, delegate decision to another agent
type ContinueDecision struct {
    Continue bool
    Reason   string
}

func (h *DSLHook) ShouldContinue(ctx AgentContext) ContinueDecision {
    // Spawn lightweight decision agent
    decision := h.decisionAgent.Evaluate(ctx.LastMessage, ctx.TaskProgress)
    
    // Anti-loop: 3 stops in 5 min = bail out
    if h.stopCount.InWindow(5*time.Minute) >= 3 {
        return ContinueDecision{Continue: false, Reason: "circuit breaker"}
    }
    
    return decision
}
```

### 4. Project-Aware Context

```bash
# Auto-detect project type, load relevant context
agentflow init

# Creates:
# .agentflow/
#   config.yaml      # Project config
#   skills/          # Custom skills
#   context/         # Auto-generated context
#     project.md     # Tech stack, structure
#     conventions.md # Code style, patterns
```

### 5. Interactive TUI Mode

```bash
agentflow tui

┌─────────────────────────────────────────────────────────────┐
│ AgentFlow v1.0.0                    Model: groq/llama3.3   │
├─────────────────────────────────────────────────────────────┤
│ [Brainstorming] Feature: User authentication               │
│                                                             │
│ ✓ Project context explored                                  │
│ ✓ Questions asked (3/3)                                     │
│ → Presenting design...                                      │
│                                                             │
│ Design Section 1/3: Architecture                            │
│ ─────────────────────────────────────────────────────────── │
│ JWT-based auth with refresh tokens. Redis for session       │
│ invalidation. Middleware pattern for route protection.      │
│                                                             │
│ [a]pprove  [r]evise  [q]uestions                           │
└─────────────────────────────────────────────────────────────┘
```

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                        AgentFlow CLI                         │
├──────────────────────────────────────────────────────────────┤
│  Commands: init | run | skill | tui | config | serve        │
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
├──────────┬──────────┬──────────┬──────────┬──────────────────┤
│  Ollama  │   Groq   │ Together │ Anthropic │ OpenAI-compat  │
└──────────┴──────────┴──────────┴──────────┴──────────────────┘
```

## Roadmap

### Phase 1: Core (Week 1-2)
- [ ] Go: CLI skeleton, config, model providers
- [ ] TS: Bun CLI skeleton, config, model providers
- [ ] Both: Ollama, Groq, Together support
- [ ] Both: Skill loader (SKILL.md parser)
- [ ] Both: Basic brainstorming skill

### Phase 2: Skills (Week 3-4)
- [ ] Port all Superpowers skills
- [ ] Subagent spawning
- [ ] Two-stage code review
- [ ] Git worktree integration
- [ ] TDD enforcement

### Phase 3: Polish (Week 5-6)
- [ ] TUI mode (bubbletea/ink)
- [ ] DSL hook (autonomous continue)
- [ ] Project context auto-detection
- [ ] Verification gates
- [ ] Documentation

### Phase 4: Launch
- [ ] GitHub repos
- [ ] Homebrew formula
- [ ] npm/bun package
- [ ] Medium articles (3)
- [ ] Reddit/HN launch

## Why AgentFlow?

| Superpowers | AgentFlow |
|-------------|-----------|
| Claude Code only | Any model, any provider |
| Plugin marketplace | Skills are just markdown |
| Anthropic API required | Ollama works offline |
| $$$$ | Free/cheap |

**Target users:**
- Developers who can't/won't pay for Claude
- Teams with Ollama/local LLM setups
- Cost-conscious startups
- Privacy-focused orgs (air-gapped)
- Open source contributors

## License

MIT — Use it, fork it, sell it, we don't care. Just build cool stuff.

## Credits

- Jesse Vincent for [Superpowers](https://github.com/obra/superpowers) — the original vision
- The open source LLM community
