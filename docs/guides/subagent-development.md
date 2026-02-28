# Subagent-Driven Development

Subagent-driven development is the secret sauce that makes AgentFlow effective. Instead of one long AI session that loses context, AgentFlow spawns fresh agents for each task.

## The Problem with Long Sessions

Traditional AI coding sessions suffer from:

| Issue | Symptom |
|-------|---------|
| Context pollution | Agent forgets earlier instructions |
| Drift | Deviates from the plan |
| Assumptions | Stops asking questions, assumes it knows |
| Quality degradation | Later code is worse than earlier |

## The Subagent Solution

AgentFlow solves this with **fresh agents per task**:

```
Controller Agent
    │
    ├── Task 1: Create User model
    │   └── Implementer Subagent (fresh)
    │       └── Spec Reviewer (fresh)
    │           └── Quality Reviewer (fresh)
    │
    ├── Task 2: Add validation
    │   └── Implementer Subagent (fresh)
    │       └── Spec Reviewer (fresh)
    │           └── Quality Reviewer (fresh)
    │
    └── Task 3: Write tests
        └── Implementer Subagent (fresh)
            └── Spec Reviewer (fresh)
                └── Quality Reviewer (fresh)
```

Each subagent:
- Starts with a clean context
- Gets only the information it needs
- Focuses on one task
- Can't be confused by previous work

## How It Works

### 1. Controller Loads Plan

```markdown
## Task 1: Create User model
File: internal/models/user.go
- ID (uuid), Email, CreatedAt
- Validate() method

## Task 2: Add email validation
File: internal/models/user.go
- Check email format in Validate()
```

### 2. Controller Spawns Implementer

The controller sends to a fresh subagent:

```markdown
# Your Task

Create User model with ID, Email, CreatedAt fields
and a Validate() method.

## Context
- Working directory: /home/user/myproject
- Language: Go
- Existing files: [list]

## Instructions
1. Read existing code
2. Implement the model
3. Write tests (TDD)
4. Commit changes

## Constraints
- Only implement what's described
- Don't modify other files
- Ask if unclear
```

### 3. Implementer Executes

The implementer:
1. Reads relevant existing code
2. Writes tests first (TDD)
3. Implements the feature
4. Self-reviews before finishing
5. Commits with descriptive message

### 4. Spec Reviewer Checks

A fresh spec reviewer receives:

```markdown
# Spec Review

## Original Spec
Create User model with ID, Email, CreatedAt fields
and a Validate() method.

## Changes
[git diff output]

## Your Job
Answer:
1. Does code implement ALL requirements?
2. Does it match the spec exactly?
3. Any extra features not in spec?
```

### 5. Quality Reviewer Checks

After spec passes, a fresh quality reviewer:

```markdown
# Quality Review

## Changes
[git diff output]

## Check For
- Security issues
- Error handling
- Test coverage
- Code style
- Performance
```

### 6. Loop Until Approved

If reviewers find issues:
1. Implementer (same session) fixes
2. Reviewer checks again
3. Repeat until approved

### 7. Controller Moves On

Task complete → next task → new subagents

## Two-Stage Review

Why two separate reviews?

| Review | Purpose | Catches |
|--------|---------|---------|
| **Spec Review** | "Did they build the right thing?" | Missing features, extra features, wrong behavior |
| **Quality Review** | "Did they build it right?" | Bugs, bad patterns, missing tests, security issues |

**Order matters**: Spec first, then quality. No point reviewing quality of wrong code.

## Subagent Benefits

### Clean Context
Each subagent starts fresh. No accumulated confusion.

### Focused Work
One task, full attention. No juggling multiple concerns.

### Parallel-Safe
Subagents don't know about each other. No conflicts.

### Easy Retry
Task failed? Spawn a new subagent. Don't debug the mess.

### Cost Control
Smaller contexts = fewer tokens = lower cost.

## Model Assignment

Different roles can use different models:

```yaml
roles:
  # Main controller - needs reasoning
  controller:
    provider: groq
    model: llama-3.3-70b-versatile
    
  # Implementers - need coding ability
  implementer:
    provider: ollama
    model: codellama:34b
    
  # Reviewers - need analysis
  reviewer:
    provider: together
    model: Qwen/Qwen2.5-Coder-32B-Instruct
```

### Role Recommendations

| Role | Model Size | Why |
|------|-----------|-----|
| Controller | 70B | Needs planning, reasoning |
| Implementer | 34B+ | Needs coding ability |
| Spec Reviewer | 13B+ | Instruction following |
| Quality Reviewer | 32B+ | Code analysis |

## Configuration

### Enable Subagent Development

It's the default for plan execution:

```bash
agentflow run "build feature X"
# Uses subagent-driven-development automatically
```

### Configure Pool Size

```yaml
subagents:
  max_parallel: 1  # Sequential (safest)
  # max_parallel: 4  # Parallel (faster, careful with file conflicts)
```

### Configure Review Strictness

```yaml
review:
  spec:
    required: true  # Always run spec review
    retry_limit: 3  # Max fix attempts
  quality:
    required: true
    fail_on: critical  # critical | important | minor
```

## Best Practices

### 1. Keep Tasks Small
2-5 minutes per task. Smaller = more focused subagents.

### 2. Provide Full Context
Don't make subagents guess. Give them exactly what they need.

### 3. Trust the Reviews
If a reviewer rejects, there's a reason. Don't skip.

### 4. Use Different Models
Expensive models for important roles, cheap for simple tasks.

### 5. Don't Parallelize File Access
If two tasks touch the same file, run them sequentially.

## Troubleshooting

### Subagent Keeps Failing

1. Is the task too big? Break it down.
2. Is the spec unclear? Clarify it.
3. Is the model too small? Try larger.

### Review Loop Never Ends

After 3 cycles:
1. Stop and assess
2. Is the task well-defined?
3. Is there an architectural problem?
4. Escalate to human review

### Too Slow

1. Use faster models (Groq for implementers)
2. Reduce review strictness
3. Increase parallel subagents (carefully)

## What's Next?

- [Custom Skills](./custom-skills.md)
- [Model Selection Guide](./model-selection.md)
- [Cost Optimization](./cost-optimization.md)
