---
name: verification-before-completion
description: "Require running verification commands and confirming real output before claiming task completion."
triggers:
  - "done"
  - "complete"
  - "finished"
  - "fixed"
  - "working"
  - "passing"
priority: 100
---

# Verification Before Completion

## Overview

**Evidence before assertions.**

You cannot claim something is "done", "fixed", or "passing" without showing the actual output that proves it.

## The Iron Law

```
NO COMPLETION CLAIMS WITHOUT VERIFICATION OUTPUT
```

| Claim | Required Proof |
|-------|----------------|
| "Tests pass" | Actual test output showing PASS |
| "It works" | Actual execution output |
| "Bug is fixed" | Before/after comparison |
| "Build succeeds" | Actual build output |
| "Deploy complete" | Deployment logs/status |

## Why This Matters

Agents have a tendency to:
- Claim success without running commands
- Assume code works because it "looks right"
- Skip verification to move faster
- Conflate "I wrote it" with "it works"

**None of these are verification.**

## Verification Patterns

### Pattern 1: Test Suite

```bash
# Run tests
go test -v ./...

# Expected output:
=== RUN   TestUserValidate
--- PASS: TestUserValidate (0.00s)
=== RUN   TestUserCreate
--- PASS: TestUserCreate (0.01s)
PASS
ok      myproject/internal/models   0.123s
```

**Verification:** See "PASS" for all tests.

### Pattern 2: Linting

```bash
# Run linter
golangci-lint run

# Expected output (nothing = success):
# (no output)

# Or explicit success:
✓ No issues found
```

**Verification:** No errors or warnings.

### Pattern 3: Build

```bash
# Build project
go build -o bin/app ./cmd/app

# Verify binary exists
ls -la bin/app
# -rwxr-xr-x  1 user  staff  12345678 Jan 15 10:00 bin/app

# Verify it runs
./bin/app --version
# app version 1.0.0
```

**Verification:** Binary exists and executes.

### Pattern 4: API Endpoint

```bash
# Start server (background)
./bin/app serve &
sleep 2

# Test endpoint
curl -s http://localhost:8080/health
# {"status":"healthy"}

# Test actual feature
curl -s -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
# {"id":"123","email":"test@example.com"}
```

**Verification:** Expected response received.

### Pattern 5: Bug Fix

```bash
# BEFORE: Show the bug
./reproduce-bug.sh
# Error: invalid user ID: abc

# AFTER: Show it's fixed
./reproduce-bug.sh
# Success: user found
```

**Verification:** Error no longer occurs.

### Pattern 6: Database Migration

```bash
# Run migration
./bin/app migrate up

# Verify tables exist
psql -c "\d users"
#                      Table "public.users"
#    Column   |           Type           | Nullable |
# ------------+--------------------------+----------+
#  id         | uuid                     | not null |
#  email      | character varying(255)   | not null |
#  created_at | timestamp with time zone | not null |
```

**Verification:** Schema matches expectation.

## What Doesn't Count

| Not Verification | Why |
|------------------|-----|
| "I wrote the code" | Code exists ≠ code works |
| "It should work" | Should ≠ does |
| "Tests exist" | Existing ≠ passing |
| "I checked manually" | Show the output |
| "No errors during writing" | Compile ≠ correct |
| "Similar code works" | Similar ≠ same |

## Completion Checklist

Before saying ANYTHING is done:

- [ ] Wrote the code/fix
- [ ] Ran verification command
- [ ] **Showed actual output**
- [ ] Output matches expected result
- [ ] Edge cases verified (if applicable)
- [ ] Committed the changes

## Output Format

When completing a task, always include:

```markdown
## Task Complete: [Name]

### Changes Made
- Modified `internal/models/user.go`: added Validate() method
- Created `internal/models/user_test.go`: 3 test cases

### Verification

**Tests:**
```
$ go test -v ./internal/models/...
=== RUN   TestUserValidate_EmptyEmail
--- PASS: TestUserValidate_EmptyEmail (0.00s)
=== RUN   TestUserValidate_InvalidEmail
--- PASS: TestUserValidate_InvalidEmail (0.00s)
=== RUN   TestUserValidate_ValidEmail
--- PASS: TestUserValidate_ValidEmail (0.00s)
PASS
```

**Lint:**
```
$ golangci-lint run ./internal/models/...
(no output - clean)
```

### Commit
```
abc1234 feat: add email validation to User model
```
```

## Model-Specific Notes

### For smaller models (7B-13B)

Be explicit about requiring output:
```
Now run `go test ./...` and paste the COMPLETE output.
Do not summarize. Paste the actual terminal output.
```

### For larger models (70B+)

Can summarize but must include:
- The exact command run
- Key output lines (PASS/FAIL, error messages)
- Final status

## Red Flags

If the agent says:
- "Tests should pass now" → Ask: "Show me the output"
- "I fixed the bug" → Ask: "Show before/after"
- "Build succeeds" → Ask: "Show the build output"
- "It works" → Ask: "Show it working"

**Never accept claims without evidence.**

## Integration

- **test-driven-development** — Tests ARE the verification
- **subagent-driven-development** — Each task needs verification
- **systematic-debugging** — Verify the fix worked
- **finishing-a-development-branch** — Final verification before merge
