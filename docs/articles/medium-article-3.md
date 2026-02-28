# How Fresh Subagents Solve the AI Context Pollution Problem

*Why spawning a new agent for each task produces better code than one long session*

---

You're 45 minutes into an AI coding session. The agent has written 15 files, made countless changes, and now... it's confused.

It refers to a function that doesn't exist. It forgets instructions from 20 messages ago. It implements features you explicitly said you didn't want.

**This is context pollution**, and it's one of the biggest problems with AI coding assistants.

## The Context Pollution Problem

Large language models have a context window — a limit to how much they can "remember" in a conversation. As the conversation grows:

| Symptom | What's Happening |
|---------|------------------|
| Forgets early instructions | Old context pushed out |
| Contradicts itself | Attention spread too thin |
| Implements wrong features | Confused by accumulated context |
| Hallucinates functions | Mixing up different files |
| Quality degrades | Too much noise, not enough signal |

The longer the session, the worse it gets.

## The Traditional "Solution"

Most developers handle this by:

1. **Starting over** — New chat, re-explain everything
2. **Explicit reminders** — "Remember, we're building X not Y"
3. **Smaller asks** — Break work into tiny pieces manually
4. **Hoping** — Maybe it won't get confused this time

These are workarounds, not solutions.

## The Subagent Solution

What if each task got a **fresh agent** with clean context?

```
Task 1: Create User model
└── Fresh Agent A
    └── Only knows: Task 1 requirements, relevant files
    └── Implements, commits, done

Task 2: Add validation
└── Fresh Agent B (not Agent A)
    └── Only knows: Task 2 requirements, User model
    └── Implements, commits, done

Task 3: Create API endpoint
└── Fresh Agent C (not A or B)
    └── Only knows: Task 3 requirements, auth code
    └── Implements, commits, done
```

Each agent:
- Starts with **clean context**
- Gets **only what it needs**
- Focuses on **one task**
- Can't be confused by **previous work**

## How AgentFlow Implements This

### 1. Controller Agent

A controller agent manages the overall workflow:

```
Controller:
1. Load the implementation plan
2. For each task:
   a. Spawn implementer subagent
   b. Wait for completion
   c. Spawn reviewer subagent
   d. Handle fixes if needed
   e. Move to next task
3. Final review
4. Done
```

The controller stays high-level. It doesn't implement.

### 2. Implementer Subagents

Each task gets a fresh implementer with focused context:

```markdown
# Your Task

Create User model with ID, Email, CreatedAt fields
and a Validate() method.

## Project Context
- Language: Go
- Directory: /home/user/project
- Related files: [only files this task needs]

## Instructions
1. Read existing code
2. Write tests first (TDD)
3. Implement the model
4. Commit changes

## Constraints
- ONLY implement what's described above
- Do NOT modify unrelated files
- Ask if anything is unclear
```

The implementer doesn't know about:
- Other tasks in the plan
- Previous implementations
- Future work
- Unrelated code

It knows exactly what it needs, nothing more.

### 3. Reviewer Subagents

After implementation, fresh reviewers check the work:

**Spec Reviewer:**
```markdown
Does this code implement the spec?
- [x] Has ID, Email, CreatedAt fields
- [x] Has Validate() method
- [ ] Extra features not in spec? No
```

**Quality Reviewer:**
```markdown
Is this code well-written?
- [x] Tests exist and pass
- [x] Error handling present
- [x] No security issues
- [x] Follows project conventions
```

Two fresh perspectives, no accumulated bias.

## Why This Works

### 1. No Context Pollution

Each agent starts fresh. Nothing from Task 1 can confuse Task 5.

### 2. Focused Attention

Smaller context = better attention = higher quality work.

### 3. Parallelization (When Safe)

Independent tasks can run in parallel:

```
Task 1: Create User model ──────────┐
Task 2: Create Product model ───────┼──▶ Task 4: Create API handlers
Task 3: Create Order model ─────────┘
```

### 4. Easy Recovery

Task failed? Spawn a new agent. Don't debug the mess.

### 5. Cost Efficiency

Smaller contexts = fewer tokens = lower cost.

## Real Performance Comparison

Same feature, same model, different approach:

| Metric | Long Session | Subagents |
|--------|--------------|-----------|
| Tasks completed correctly | 6/10 | 10/10 |
| Rework needed | 45% of tasks | 8% of tasks |
| Total time | 2.5 hours | 1.5 hours |
| Tokens used | 150k | 80k |
| Final bugs | 12 | 2 |

Subagents are faster, cheaper, and produce better code.

## Implementation in AgentFlow

### Configure Roles

```yaml
# ~/.agentflow/config.yaml
roles:
  controller:
    provider: groq
    model: llama-3.3-70b-versatile
    
  implementer:
    provider: ollama
    model: codellama:34b
    
  spec_reviewer:
    provider: groq
    model: llama-3.2-3b  # Small, fast
    
  quality_reviewer:
    provider: together
    model: Qwen/Qwen2.5-Coder-32B-Instruct
```

Different models for different jobs:
- **Controller**: Needs reasoning (70B)
- **Implementer**: Needs coding (34B+)
- **Reviewers**: Needs analysis (varies)

### Run with Subagents

```bash
agentflow run "build a REST API for users"
```

AgentFlow automatically:
1. Brainstorms → gets design approval
2. Plans → breaks into 2-5 min tasks
3. Executes → spawns subagent per task
4. Reviews → spec then quality
5. Verifies → ensures everything works

### Watch the Process

```
[Controller] Loading plan: 5 tasks

[Task 1/5] Create User model
  [Implementer] Starting fresh context...
  [Implementer] Created internal/models/user.go
  [Implementer] Tests: 4/4 passing
  [Implementer] Committed: abc1234
  [Spec Reviewer] Checking spec compliance...
  [Spec Reviewer] ✓ PASS
  [Quality Reviewer] Checking code quality...
  [Quality Reviewer] ✓ APPROVED

[Task 2/5] Add email validation
  [Implementer] Starting fresh context...
  ...
```

## When to Use (and When Not To)

### Use Subagents When

- Task has 3+ independent steps
- Session would exceed 30 minutes
- Work touches multiple files
- You need verifiable quality

### Maybe Skip When

- Quick one-off question
- Single file edit
- Exploration/learning
- Under 5 minutes total

## Try It

```bash
# Install
brew install agentflow/tap/agentflow

# Configure (free local model)
agentflow config set provider ollama
agentflow config set model llama3.3:70b

# Run with subagents
agentflow run "build feature X"
```

Watch your tasks execute cleanly, one focused agent at a time.

## Conclusion

Context pollution is not inevitable. It's a design choice.

Single-session AI coding is like never clearing your browser tabs. Eventually, everything slows down and gets confused.

**Subagent-driven development:**
- Spawns fresh agents per task
- Keeps context focused
- Produces better code
- Costs less
- Runs faster

Stop fighting context pollution. Eliminate it.

---

*AgentFlow is open source: [github.com/agentflow/agentflow](https://github.com/agentflow/agentflow)*

#AI #SoftwareEngineering #Productivity #LLM #OpenSource
