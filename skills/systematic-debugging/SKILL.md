---
name: systematic-debugging
description: "Use when encountering any bug, test failure, or unexpected behavior. BEFORE proposing fixes."
triggers:
  - "bug"
  - "error"
  - "failing"
  - "broken"
  - "doesn't work"
  - "unexpected"
priority: 95
---

# Systematic Debugging

## Overview

**Stop guessing. Start investigating.**

Random fixes waste time and create new bugs. Quick patches mask underlying issues.

## The Iron Law

```
NO FIXES WITHOUT ROOT CAUSE INVESTIGATION FIRST
```

If you haven't found the root cause, you cannot propose fixes.

## When to Use

Use for **ANY** technical issue:
- Test failures
- Runtime errors
- Unexpected behavior
- Performance problems
- Build failures
- Integration issues

**Especially use when:**
- Under time pressure (rushing = thrashing)
- "Quick fix" seems obvious (obvious is often wrong)
- You've already tried multiple fixes
- Previous fix didn't work
- You don't understand why it's broken

## The Four Phases

### Phase 1: Root Cause Investigation

**BEFORE attempting ANY fix:**

#### 1.1 Read Error Messages Carefully

```bash
# Don't skim — READ
cat error.log | head -100

# Stack traces contain the answer
grep -A 20 "panic:" output.txt
grep -A 20 "Error:" output.txt
```

What to note:
- Exact error text
- File and line number
- Stack trace order (most recent call last)
- Timestamps (when did it start?)

#### 1.2 Reproduce Consistently

```bash
# Can you trigger it reliably?
./run-command
# Same error?

# Try 3 times
for i in 1 2 3; do ./run-command; done
```

Ask:
- Same error every time?
- Or intermittent?
- What are the exact steps?
- What inputs trigger it?

**If not reproducible → gather more data, don't guess.**

#### 1.3 Check Recent Changes

```bash
# What changed?
git log --oneline -20
git diff HEAD~5

# Who changed relevant files?
git log --oneline -- path/to/broken/file.go

# When did it last work?
git bisect start
git bisect bad HEAD
git bisect good v1.2.3
```

#### 1.4 Gather Evidence

For multi-component systems, **instrument before fixing**:

```bash
# Add debug logging at each boundary
echo "=== Input to component A ===" 
echo "$INPUT"

echo "=== Output from component A ==="
component_a_result

echo "=== Input to component B ==="
echo "$component_a_result"
```

Run once to see **WHERE** it breaks, then investigate that component.

#### 1.5 Trace Data Flow

Follow the bad value backward:

```
Error: "invalid user ID: abc"
  └─ Where does "abc" come from?
       └─ parseUserID() returns it
            └─ request.Query("id") returns it
                 └─ URL has ?id=abc
                      └─ Frontend sent wrong ID
                           └─ ROOT CAUSE: Frontend bug
```

### Phase 2: Pattern Analysis

#### 2.1 Find Working Examples

```bash
# Similar working code?
grep -r "similar_function" --include="*.go"

# Same pattern used elsewhere?
grep -rn "database.Query" --include="*.go" | head -20
```

#### 2.2 Compare Working vs Broken

```
Working code:
  db.Query(ctx, "SELECT * FROM users WHERE id = $1", id)

Broken code:
  db.Query("SELECT * FROM users WHERE id = $1", id)
  # Missing ctx!
```

List every difference, no matter how small.

#### 2.3 Read References Completely

If implementing a pattern, read the **entire** reference:

```bash
# Don't skim — read every line
cat examples/reference_implementation.go

# Check documentation fully
open https://pkg.go.dev/database/sql
```

### Phase 3: Hypothesis and Testing

#### 3.1 Form Single Hypothesis

Write it down:

```markdown
**Hypothesis:** The error occurs because `ctx` is nil when 
passed to `db.Query()` due to the context being canceled 
before the query runs.

**Evidence:**
- Error says "context canceled"
- Query is at end of request handler
- Works when request is slow

**Test:** Add logging before query to print ctx.Err()
```

#### 3.2 Test Minimally

Make the **smallest possible change** to test:

```go
// Add ONE line of logging
log.Printf("ctx before query: err=%v", ctx.Err())
result, err := db.Query(ctx, query)
```

**Not:**
```go
// DON'T change multiple things
log.Printf("ctx: %v", ctx)
newCtx := context.Background()  // changed context
result, err := db.Query(newCtx, query)  // different behavior
```

#### 3.3 Interpret Results

- Hypothesis confirmed → proceed to Phase 4
- Hypothesis wrong → form NEW hypothesis
- Unclear → add more instrumentation

**DON'T pile fixes on top of each other.**

### Phase 4: Implementation

#### 4.1 Create Failing Test

```go
func TestQueryWithCanceledContext(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately
    
    _, err := db.Query(ctx, "SELECT 1")
    
    if err == nil {
        t.Error("expected error for canceled context")
    }
}
```

Run it, verify it **fails** (TDD principle).

#### 4.2 Implement Single Fix

Address the root cause:

```go
func handleRequest(w http.ResponseWriter, r *http.Request) {
    // Create query-specific context with timeout
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()
    
    result, err := db.Query(ctx, query)
    // ...
}
```

**One change.** No "while I'm here" improvements.

#### 4.3 Verify Fix

```bash
go test ./...
# All pass?

./run-original-failing-scenario
# Works now?
```

#### 4.4 If Fix Doesn't Work

**Count your attempts.**

| Attempts | Action |
|----------|--------|
| 1-2 | Return to Phase 1, re-analyze |
| 3+ | **STOP** — this is an architectural problem |

After 3+ failed fixes:
- Don't try fix #4
- Question the architecture
- Is this pattern fundamentally sound?
- Should we refactor instead of patch?
- Discuss with team before continuing

## Red Flags — STOP and Go Back

If you catch yourself thinking:

| Thought | Problem |
|---------|---------|
| "Quick fix, investigate later" | You'll never investigate later |
| "Just try X and see" | Guessing, not investigating |
| "Add multiple changes, test" | Can't isolate what works |
| "Skip test, manual verify" | Untested fixes don't stick |
| "Probably X, let me fix" | "Probably" = don't understand |
| "I'll adapt the pattern" | Partial understanding = bugs |
| "One more fix attempt" | 3+ = wrong architecture |

**All of these mean: STOP. Return to Phase 1.**

## Quick Reference

| Phase | Activities | Exit Criteria |
|-------|------------|---------------|
| 1. Root Cause | Read errors, reproduce, trace | Understand WHAT and WHY |
| 2. Pattern | Find working code, compare | Identify key differences |
| 3. Hypothesis | Form theory, test minimally | Theory confirmed |
| 4. Implement | Create test, fix, verify | Bug resolved, tests pass |

## Anti-Debugging Checklist

Before proposing ANY fix:

- [ ] I read the full error message
- [ ] I can reproduce the issue
- [ ] I know what changed recently
- [ ] I traced the data flow
- [ ] I found a working example to compare
- [ ] I formed a specific hypothesis
- [ ] I tested the hypothesis minimally
- [ ] I understand the root cause

## Model-Specific Notes

### For smaller models (7B-13B)

Enforce each phase explicitly:
```
Phase 1: What is the exact error message?
Phase 1: Can you reproduce it? Show output.
Phase 2: Find similar working code.
Phase 3: What is your hypothesis?
...
```

### For larger models (70B+)

Can combine phases but must show:
- Evidence gathered
- Hypothesis formed
- Minimal test performed
- Root cause identified

## Integration

- **test-driven-development** — Create failing test for bug
- **verification-before-completion** — Verify fix actually works
- **requesting-code-review** — Review the fix for quality
