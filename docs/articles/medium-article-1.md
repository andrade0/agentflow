# AgentFlow: Superpowers for Everyone — Why I Built an Open-Source Alternative to Paid AI Coding Agents

*How to get the structured workflow benefits of Superpowers without paying for Claude*

---

Jesse Vincent's [Superpowers](https://github.com/obra/superpowers) changed how I think about AI coding. Instead of treating AI as a code generator, it treats it as a disciplined developer who follows processes.

But there's a catch: **it requires Claude Code and an Anthropic API key.** That's $20/month minimum, and the costs add up fast.

I built **AgentFlow** to bring these workflows to everyone — including developers running free local models with Ollama.

## The Problem with AI Coding Today

Most developers use AI coding assistants like this:

```
User: "Add authentication"
AI: [writes 500 lines of code]
User: "Wait, I wanted OAuth, not JWT"
AI: [rewrites everything]
User: "Actually, the tests are failing"
AI: [adds random fixes]
```

This is **guess-and-fix** programming. The AI guesses what you want, you correct it, repeat forever.

## The Superpowers Insight

Superpowers introduced a different approach:

1. **Brainstorm first** — Ask questions, understand requirements, propose options
2. **Design before code** — Present a design, get explicit approval
3. **Plan in small tasks** — Break work into 2-5 minute chunks
4. **Execute with fresh agents** — New context per task, no drift
5. **Review twice** — Spec compliance, then code quality
6. **Verify before done** — Show actual output, not "it should work"

This turns AI from a code generator into a **process follower**. And process followers are predictable.

## Why AgentFlow?

Superpowers is amazing, but:

| Limitation | AgentFlow Solution |
|------------|-------------------|
| Claude only | Ollama, Groq, Together, any model |
| Paid API | Free with local models |
| Cloud dependency | Runs fully offline |
| Plugin marketplace | Skills are just markdown |

## How AgentFlow Works

### 1. Install and Configure

```bash
# Install
brew install agentflow/tap/agentflow

# Use free local model
ollama pull llama3.3:70b
agentflow config set provider ollama
agentflow config set model llama3.3:70b
```

### 2. Run a Task

```bash
agentflow run "add user authentication with JWT"
```

### 3. Watch the Magic

```
[Brainstorming Skill Activated]

Before I implement authentication, I have a few questions:

1. Should tokens be short-lived (15min) or long-lived (7 days)?
2. Do you need refresh tokens?
3. Where should user data be stored (database, file, in-memory)?

> Short-lived with refresh tokens, PostgreSQL

Got it. Here are three approaches:

**Approach A: Standard JWT + Refresh**
- Access token (15min) + refresh token (7 days)
- Pros: Industry standard, stateless
- Cons: Can't revoke access tokens

**Approach B: JWT + Redis Blacklist**
- Same as A, but track revoked tokens
- Pros: Can revoke, still fast
- Cons: Adds Redis dependency

**Approach C: Session-based**
- Server-side sessions, cookie auth
- Pros: Simple, revocable
- Cons: Not stateless, scaling issues

I recommend **Approach B** because it balances security with scalability.

Do you approve this design?

> Yes

[Writing Plan...]

## Implementation Plan

### Task 1: Create JWT utilities (3 min)
- File: internal/auth/jwt.go
- Functions: GenerateToken, ValidateToken, RefreshToken

### Task 2: Create auth middleware (4 min)
- File: internal/middleware/auth.go
- Extract token, validate, set user context

### Task 3: Create auth handlers (5 min)
- File: internal/handlers/auth.go
- POST /login, POST /refresh, POST /logout

...

[Executing with Subagents...]

Task 1: Spawning implementer...
- Created internal/auth/jwt.go ✓
- Added tests (4/4 passing) ✓
- Spec review: PASS ✓
- Quality review: APPROVED ✓

Task 2: Spawning implementer...
...
```

## The Key Innovations

### 1. Mandatory Brainstorming

The brainstorming skill has a **hard gate**:

```markdown
<HARD-GATE>
Do NOT write any code until:
1. You have presented a design
2. The user has explicitly approved it
</HARD-GATE>
```

Even "simple" tasks go through brainstorming. Simple tasks are where assumptions cause the most wasted work.

### 2. Small Tasks

Plans break work into 2-5 minute tasks. Why?

| Task Duration | Agent Performance |
|--------------|-------------------|
| 30+ min | Loses context, drifts |
| 10-15 min | Sometimes loses track |
| 2-5 min | Stays focused |

### 3. Fresh Subagents

Each task gets a fresh agent with clean context:

```
Controller → Implementer (fresh) → Spec Reviewer (fresh) → Quality Reviewer (fresh)
```

No context pollution. No accumulated confusion.

### 4. Two-Stage Review

After each task:
1. **Spec Review**: Did they build the right thing?
2. **Quality Review**: Did they build it right?

Order matters. No point reviewing quality of wrong code.

### 5. Evidence Over Claims

The verification skill requires **actual output**:

```markdown
# Not Accepted
"Tests should pass now"

# Accepted
$ go test ./...
PASS
ok   myproject/internal/auth  0.123s
```

## Free Model Recommendations

You don't need expensive APIs:

| Provider | Model | Cost | Best For |
|----------|-------|------|----------|
| Ollama | llama3.3:70b | Free | Everything, offline |
| Ollama | codellama:34b | Free | Code generation |
| Groq | llama-3.3-70b | Free (30/min) | Fast inference |
| Together | Qwen2.5-Coder-32B | $0.20/M | Code review |

My setup:
- **Controller**: Groq (fast, free tier)
- **Implementers**: Ollama (unlimited, good quality)
- **Reviewers**: Together (code-specialized)

## Results

I've been using AgentFlow for 3 weeks:

- **Context drift**: Zero (fresh subagents)
- **Rework**: Down 70% (brainstorming catches issues early)
- **Test coverage**: Up (TDD skill enforces it)
- **Time to feature**: Down 40% (no guess-and-fix loops)
- **Cost**: $0 (all local models)

## Getting Started

```bash
# Install
curl -fsSL https://agentflow.dev/install.sh | sh

# Or with Homebrew
brew install agentflow/tap/agentflow

# Set up Ollama (free)
ollama pull llama3.3:70b
agentflow config set provider ollama

# Initialize your project
cd your-project
agentflow init

# Start building
agentflow run "add feature X"
```

## The Future

AgentFlow is just the beginning. We're building:

- **More skills**: Security auditing, documentation, refactoring
- **IDE integrations**: VS Code, Cursor, Zed
- **Team features**: Shared skills, review workflows
- **Model fine-tuning**: Custom models for your codebase

## Conclusion

AI coding doesn't have to be expensive. It doesn't have to be unpredictable. With the right workflows, free models can perform like paid ones.

AgentFlow brings Superpowers-style workflows to everyone. No subscriptions, no cloud lock-in, no excuses.

**Try it**: [github.com/agentflow/agentflow](https://github.com/agentflow/agentflow)

---

*Star the repo if you find it useful. Contributions welcome!*

#AI #Coding #OpenSource #LLM #DeveloperTools
