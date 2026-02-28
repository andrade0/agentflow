---
name: writing-plans
description: "Convert approved designs into step-by-step implementation plans with 2-5 minute tasks."
triggers:
  - "plan"
  - "break down"
  - "tasks"
  - "implementation plan"
requires:
  - "approved design document"
priority: 90
---

# Writing Implementation Plans

## Overview

Transform approved designs into detailed, executable plans. Each task should be completable in **2-5 minutes** by a focused agent.

**The goal:** Create a plan so detailed that an enthusiastic junior developer with no project context could follow it.

## Why Small Tasks?

| Task Size | Agent Performance |
|-----------|-------------------|
| 30+ min   | Loses context, makes assumptions, drifts |
| 10-15 min | Occasionally loses track |
| 2-5 min   | Stays focused, follows instructions exactly |

Small tasks = higher quality + easier review + faster iteration.

## The Process

### 1. Read the Design Document

```bash
cat docs/plans/YYYY-MM-DD-*-design.md
```

Understand:
- What we're building
- Key decisions already made
- Constraints and requirements

### 2. Identify Work Streams

Break the design into logical streams:

```markdown
## Work Streams

1. **Data Layer** — Models, database, migrations
2. **Business Logic** — Core functions, validation
3. **API Layer** — Endpoints, middleware
4. **UI Layer** — Components, pages
5. **Testing** — Unit tests, integration tests
```

### 3. Break Down Each Stream

For each stream, create tasks of 2-5 minutes each.

**Good task:**
```markdown
### Task 1.1: Create User model
- File: `internal/models/user.go`
- Fields: ID (uuid), Email (string), CreatedAt (time.Time)
- Add `Validate()` method that checks email format
- Add `TableName()` returning "users"
```

**Bad task (too vague):**
```markdown
### Task: Set up the data layer
- Create models and stuff
```

**Bad task (too big):**
```markdown
### Task: Implement user authentication
- Handle signup, login, logout, password reset, email verification
```

### 4. Task Template

Every task must have:

```markdown
### Task X.Y: [Specific action]

**File(s):** [Exact paths]

**What to do:**
1. [Step 1]
2. [Step 2]
3. [Step 3]

**Verification:**
- [ ] File exists at correct path
- [ ] Tests pass: `go test ./internal/models/...`
- [ ] No linting errors: `golangci-lint run`

**Dependencies:** [Tasks that must complete first, if any]
```

### 5. Order Tasks

Consider dependencies:
- Models before services
- Services before handlers
- Interfaces before implementations
- Tests alongside implementation (TDD)

### 6. Write the Plan Document

Save to: `docs/plans/YYYY-MM-DD-<topic>-plan.md`

```markdown
# Implementation Plan: [Feature]

## Overview
[Brief summary of what we're implementing]

## Design Reference
See: `docs/plans/YYYY-MM-DD-<topic>-design.md`

## Work Streams

### Stream 1: [Name]

#### Task 1.1: [Title]
[Task details]

#### Task 1.2: [Title]
[Task details]

### Stream 2: [Name]
...

## Execution Order

1. Task 1.1 (no dependencies)
2. Task 1.2 (depends on 1.1)
3. Task 2.1 (no dependencies, can parallel with 1.x)
...

## Success Criteria

- [ ] All tasks completed
- [ ] All tests passing
- [ ] Code reviewed
- [ ] Documentation updated
```

### 7. Commit the Plan

```bash
git add docs/plans/
git commit -m "docs: add implementation plan for [feature]"
```

## Task Granularity Examples

### Too Big → Just Right

**Too big:** "Implement API endpoints"

**Just right:**
1. Create router setup in `cmd/server/routes.go`
2. Add GET /users endpoint returning empty list
3. Add POST /users endpoint with request validation
4. Add GET /users/:id endpoint with 404 handling
5. Add middleware for request logging
6. Add middleware for authentication

### Too Vague → Specific

**Too vague:** "Add tests"

**Specific:**
1. Create `internal/models/user_test.go` with TestUserValidate
2. Add test case: valid email passes
3. Add test case: empty email fails
4. Add test case: malformed email fails
5. Create `internal/handlers/user_test.go` with TestCreateUser
6. Add test case: valid request returns 201
7. Add test case: missing email returns 400

## Principles

### YAGNI (You Aren't Gonna Need It)

Remove tasks for features that aren't in the design:
- ❌ "Add pagination (might need it later)"
- ✅ Only implement what's in the approved design

### DRY (Don't Repeat Yourself)

Identify shared code early:
- If 3+ tasks need the same helper, add a task to create it first
- Reference the helper task as a dependency

### TDD Integration

Pair implementation tasks with test tasks:
```markdown
#### Task 3.1: Write test for CreateUser handler
**Test cases:** valid input, missing email, duplicate email

#### Task 3.2: Implement CreateUser handler
**Verification:** Tests from 3.1 pass
```

## Model-Specific Notes

### For smaller models (7B-13B)

Be extra explicit:
- Include exact function signatures
- Show expected imports
- Provide code snippets for complex tasks

### For larger models (70B+)

Can handle more abstract tasks:
- Reference patterns instead of spelling out every line
- Trust the model to fill in obvious details
- Focus on edge cases and tricky bits

## Checklist

Before executing the plan:

- [ ] Design document exists and is committed
- [ ] Every task is 2-5 minutes of work
- [ ] Every task has verification criteria
- [ ] Dependencies are clearly marked
- [ ] Execution order accounts for dependencies
- [ ] Plan document committed to git

## Next Skill

After plan is written → invoke `subagent-driven-development` or `executing-plans` to begin implementation.
