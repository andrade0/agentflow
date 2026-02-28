# Understanding Skills

Skills are the heart of AgentFlow. They define structured workflows that guide the AI through complex tasks.

## What is a Skill?

A skill is a markdown file that describes:
- **When** to use it (triggers, conditions)
- **What** to do (process, steps)
- **How** to verify success (checklists, verification)

Skills transform vague requests into systematic processes.

## How Skills Work

### 1. Task Matching

When you run `agentflow run "build a REST API"`, AgentFlow:

1. Scans all available skills
2. Matches based on:
   - **Triggers**: Keywords in your request
   - **Description**: Semantic similarity
   - **Priority**: Higher priority skills match first

3. Selects the best match (or asks if ambiguous)

### 2. Skill Execution

Once matched, the skill guides the agent:

```
User: "build a REST API"
       ↓
[Matches: brainstorming skill]
       ↓
Agent follows brainstorming process:
  1. Explore project context
  2. Ask clarifying questions
  3. Propose approaches
  4. Present design
  5. Get approval
       ↓
[Matches: writing-plans skill]
       ↓
Agent creates implementation plan
       ↓
[Matches: subagent-driven-development skill]
       ↓
Subagents execute each task
```

## Built-in Skills

AgentFlow ships with these production-ready skills:

### Workflow Skills

| Skill | Triggers | Purpose |
|-------|----------|---------|
| `brainstorming` | build, create, feature, new | Mandatory design before coding |
| `writing-plans` | plan, break down, tasks | 2-5 min task breakdown |
| `executing-plans` | execute, implement | Human-in-loop execution |
| `subagent-driven-development` | run, start | Automated execution |

### Quality Skills

| Skill | Triggers | Purpose |
|-------|----------|---------|
| `test-driven-development` | test, TDD | RED-GREEN-REFACTOR |
| `systematic-debugging` | bug, error, fix | 4-phase root cause |
| `verification-before-completion` | done, complete | Evidence checks |
| `requesting-code-review` | review | Two-stage review |

### Git Skills

| Skill | Triggers | Purpose |
|-------|----------|---------|
| `using-git-worktrees` | branch, worktree | Isolated development |
| `finishing-a-development-branch` | merge, PR | Clean branch completion |

## Skill Anatomy

A skill file (`SKILL.md`) has two parts:

### 1. Front Matter (YAML)

```yaml
---
name: brainstorming
description: "Mandatory before any creative work - design before code"
triggers:
  - "build"
  - "create"
  - "feature"
priority: 100
requires:
  - "project context"
---
```

| Field | Purpose |
|-------|---------|
| `name` | Unique identifier |
| `description` | When to use (for matching) |
| `triggers` | Keywords that activate this skill |
| `priority` | Higher = more likely to match (0-100) |
| `requires` | Prerequisites before running |

### 2. Content (Markdown)

```markdown
# Brainstorming

## Overview
What this skill does and why...

## Process
1. Step one
2. Step two
...

## Checklist
- [ ] Did X
- [ ] Did Y

## Red Flags
Don't do these things...
```

## Skill Chaining

Skills often invoke other skills:

```
brainstorming
    ↓ (after design approved)
writing-plans
    ↓ (after plan created)
subagent-driven-development
    ↓ (after all tasks)
finishing-a-development-branch
```

This creates a complete workflow from idea to merged code.

## Skill Priority

When multiple skills match, priority determines which runs:

| Priority | When to Use |
|----------|-------------|
| 100 | Must always run (brainstorming, verification) |
| 90 | Core workflow (planning, execution) |
| 80 | Quality enforcement (TDD, debugging) |
| 70 | Git operations (worktrees, branches) |
| 50 | Default for custom skills |
| 0-49 | Optional, rarely matched |

## Skill Matching Examples

### Request: "add user authentication"

1. Matches `brainstorming` (triggers: "add", priority: 100)
2. Brainstorming completes → invokes `writing-plans`
3. Plan ready → invokes `subagent-driven-development`

### Request: "fix the login bug"

1. Matches `systematic-debugging` (triggers: "fix", "bug", priority: 95)
2. Debugging finds root cause → applies fix
3. Invokes `verification-before-completion`

### Request: "run tests"

1. Matches `test-driven-development` (triggers: "test", priority: 80)
2. Runs test suite with verification

## Skill Directories

Skills are loaded from (in order):

1. **Project skills**: `.agentflow/skills/`
2. **User skills**: `~/.agentflow/skills/`
3. **Built-in skills**: (bundled with AgentFlow)

Later directories override earlier ones (project overrides user, etc.).

## Viewing Skills

```bash
# List all available skills
agentflow skill list

# Show skill details
agentflow skill show brainstorming

# See which skill would match
agentflow skill match "build a REST API"
```

## What's Next?

- [Creating Custom Skills](./custom-skills.md)
- [Skill Best Practices](./skill-best-practices.md)
- [Subagent Development](./subagent-development.md)
