---
name: subagent-driven-development
description: "Execute implementation plans by dispatching a fresh agent per task with two-stage review (spec compliance + code quality)."
triggers:
  - "execute plan"
  - "run plan"
  - "start implementation"
  - "subagent"
requires:
  - "implementation plan document"
priority: 85
---

# Subagent-Driven Development

## Overview

Execute plans by spawning a **fresh agent for each task**, followed by **two-stage review**:
1. **Spec Review** — Does the code match the plan?
2. **Quality Review** — Is the code well-written?

**Core principle:** Fresh context per task = no pollution, no drift, higher quality.

## Why Fresh Agents?

| Approach | Problem |
|----------|---------|
| Single long session | Context pollution, forgets instructions, drifts from plan |
| Manual task switching | Slow, error-prone, human bottleneck |
| Fresh agent per task | Clean slate, follows instructions exactly, parallel-safe |

## The Process

```
┌─────────────────────────────────────────────────────────────────┐
│                    Controller (You)                             │
├─────────────────────────────────────────────────────────────────┤
│  1. Load plan                                                   │
│  2. Extract all tasks upfront                                   │
│  3. For each task:                                              │
│     ├─ Spawn Implementer Agent                                  │
│     │   └─ Implements, tests, commits                           │
│     ├─ Spawn Spec Reviewer Agent                                │
│     │   └─ Verifies code matches spec                           │
│     ├─ Spawn Quality Reviewer Agent                             │
│     │   └─ Checks code quality                                  │
│     └─ Mark task complete                                       │
│  4. Final review of entire implementation                       │
│  5. Finish branch                                               │
└─────────────────────────────────────────────────────────────────┘
```

## Setup

### 1. Load the Plan

```bash
cat docs/plans/YYYY-MM-DD-*-plan.md
```

### 2. Extract All Tasks

Create a task list with full context:

```markdown
## Task List

### Task 1: [Title]
**From plan section:** Stream 1
**Full text:** [Copy exact task description]
**Dependencies:** None
**Status:** Pending

### Task 2: [Title]
**From plan section:** Stream 1
**Full text:** [Copy exact task description]
**Dependencies:** Task 1
**Status:** Pending

...
```

## Per-Task Workflow

### Step 1: Dispatch Implementer Agent

Create a new agent session with:

```markdown
# Task Context

You are implementing a specific task from a larger plan.

## Your Task
[Full task text from plan]

## Project Context
- Working directory: [path]
- Language: [Go/TypeScript/etc]
- Relevant files: [list files to read first]

## Instructions
1. Read relevant existing code first
2. Implement exactly what the task describes
3. Write tests (TDD: test first, then implementation)
4. Run tests and verify they pass
5. Commit with message: "feat: [task title]"
6. Self-review your changes before finishing

## Constraints
- Do NOT implement beyond what the task describes
- Do NOT refactor unrelated code
- Do NOT add features "while you're there"
- If task is unclear, ask BEFORE implementing

## When Done
Report:
- Files created/modified
- Tests added and their status
- Commit SHA
- Any issues or concerns
```

### Step 2: Handle Implementer Questions

If the implementer asks questions:
1. Answer clearly and completely
2. Provide additional context if needed
3. Let them continue after getting answers

Do NOT rush them into implementation.

### Step 3: Dispatch Spec Reviewer Agent

After implementer commits, spawn a new agent:

```markdown
# Spec Review

You are reviewing code for **specification compliance only**.

## The Task Specification
[Full task text from plan]

## Changes to Review
```bash
git diff [before-sha]..[after-sha]
```

## Your Job
Answer these questions:

1. **Complete:** Does the code implement ALL requirements in the spec?
2. **Accurate:** Does the implementation match what was requested?
3. **Minimal:** Did they add anything NOT in the spec?

## Output Format

### Verdict: [PASS / FAIL]

### Findings:

**Missing (must fix):**
- [What's in spec but not in code]

**Extra (should remove):**
- [What's in code but not in spec]

**Correct:**
- [What matches spec perfectly]
```

### Step 4: Fix Spec Issues

If spec review fails:
1. The **same implementer** fixes the issues
2. Spec reviewer reviews again
3. Repeat until PASS

### Step 5: Dispatch Quality Reviewer Agent

Only after spec review passes, spawn quality reviewer:

```markdown
# Code Quality Review

You are reviewing code for **quality only** (spec compliance already verified).

## Changes to Review
```bash
git diff [before-sha]..[after-sha]
```

## Review Criteria

### Critical (blocks merge)
- Security vulnerabilities
- Data loss risks
- Race conditions
- Obvious bugs

### Important (should fix)
- Missing error handling
- Poor test coverage
- Hardcoded values
- Code duplication

### Minor (nice to have)
- Naming improvements
- Additional documentation
- Style consistency

## Output Format

### Verdict: [APPROVED / CHANGES REQUESTED]

### Critical Issues:
[List or "None"]

### Important Issues:
[List or "None"]

### Minor Suggestions:
[List or "None"]

### Strengths:
[What's done well]
```

### Step 6: Fix Quality Issues

If quality review requests changes:
1. Implementer fixes issues
2. Quality reviewer reviews again
3. Repeat until APPROVED

### Step 7: Mark Task Complete

Update task status:
```markdown
### Task 1: [Title]
**Status:** ✅ Complete
**Commit:** abc1234
**Spec Review:** PASS
**Quality Review:** APPROVED
```

## Order of Operations

```
WRONG ORDER:
  Code → Quality Review → Spec Review
  (Quality review wastes time if spec is wrong)

RIGHT ORDER:
  Code → Spec Review → Quality Review
  (Only review quality of spec-compliant code)
```

## Red Flags

### Never Do These

- ❌ Skip spec review ("code looks fine")
- ❌ Skip quality review ("we're in a hurry")
- ❌ Continue with unfixed issues
- ❌ Let implementer self-review replace actual review
- ❌ Start quality review before spec is PASS
- ❌ Move to next task with open issues
- ❌ Run multiple implementers in parallel on same files

### If Things Go Wrong

**Implementer fails completely:**
- Don't try to fix manually (context pollution)
- Spawn a new implementer with more context
- Include what went wrong and what to avoid

**3+ review cycles:**
- Stop and assess
- Is the task too big? Break it down.
- Is the spec unclear? Clarify it.
- Is the model struggling? Try a larger model.

## Model Configuration

Recommended model assignments:

```yaml
roles:
  implementer:
    model: ollama/codellama:34b  # Good at coding
    context: 8192
    
  spec_reviewer:
    model: groq/llama-3.3-70b-versatile  # Good at instruction following
    context: 4096
    
  quality_reviewer:
    model: together/Qwen2.5-Coder-32B  # Good at code analysis
    context: 8192
```

Use larger models for complex tasks, smaller for simple ones.

## Example Session

```
Controller: Starting subagent-driven-development

[Load plan: docs/plans/2024-01-15-auth-plan.md]
[Extract 5 tasks]

--- Task 1: Create User model ---

[Spawn implementer with task context]

Implementer: Before I start - should ID be int or uuid?

Controller: UUID, we need distributed-friendly IDs.

Implementer: Got it. Implementing now...
  - Created internal/models/user.go
  - Added User struct with uuid.UUID
  - Created user_test.go with 3 tests
  - All tests passing
  - Committed: abc1234

[Spawn spec reviewer]

Spec Reviewer: 
  Verdict: PASS
  - ✅ User struct with ID, Email, CreatedAt
  - ✅ Validate() method checking email
  - ✅ TableName() returning "users"
  - No missing requirements
  - No extra additions

[Spawn quality reviewer]

Quality Reviewer:
  Verdict: APPROVED
  - Critical: None
  - Important: None  
  - Minor: Consider adding String() method
  - Strengths: Clean code, good test coverage

[Mark Task 1 complete]

--- Task 2: Create User repository ---
...
```

## Checklist

Before starting:
- [ ] Plan document exists and is committed
- [ ] Git worktree/branch is set up
- [ ] All tasks extracted with full context
- [ ] Model configuration ready

Per task:
- [ ] Implementer dispatched with full context
- [ ] Implementer questions answered
- [ ] Code committed
- [ ] Spec review PASS
- [ ] Quality review APPROVED
- [ ] Task marked complete

After all tasks:
- [ ] Final review of entire implementation
- [ ] All tests passing
- [ ] Ready for `finishing-a-development-branch`

## Related Skills

- **brainstorming** — Creates the design
- **writing-plans** — Creates the plan this executes
- **test-driven-development** — Implementers should follow TDD
- **finishing-a-development-branch** — After all tasks complete
