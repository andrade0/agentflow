---
name: test-driven-development
description: "Enforce RED-GREEN-REFACTOR cycle. Tests must be written first and must fail before implementation."
triggers:
  - "implement"
  - "code"
  - "feature"
  - "fix"
  - "TDD"
priority: 80
---

# Test-Driven Development

## Overview

Write tests **before** code. Watch them **fail**. Then make them **pass**.

```
RED → GREEN → REFACTOR
 │       │        │
 │       │        └─ Clean up, no new behavior
 │       └─ Minimal code to pass
 └─ Write failing test first
```

## The Iron Law

```
NO IMPLEMENTATION CODE WITHOUT A FAILING TEST FIRST
```

If you wrote code before tests, **delete it** and start over.

## Why TDD?

| Without TDD | With TDD |
|-------------|----------|
| "I think it works" | "I know it works" |
| Tests written to pass existing code | Code written to pass tests |
| Edge cases discovered in production | Edge cases caught upfront |
| Refactoring is scary | Refactoring is safe |

## The Cycle

### 1. RED: Write a Failing Test

```go
func TestUserValidate_EmptyEmail(t *testing.T) {
    user := User{Email: ""}
    err := user.Validate()
    
    if err == nil {
        t.Error("expected error for empty email, got nil")
    }
}
```

**Run it:**
```bash
go test ./internal/models/...
# FAIL: TestUserValidate_EmptyEmail
```

✅ Test exists and **fails** — you're in RED.

### 2. GREEN: Write Minimal Code to Pass

```go
func (u *User) Validate() error {
    if u.Email == "" {
        return errors.New("email required")
    }
    return nil
}
```

**Run it:**
```bash
go test ./internal/models/...
# PASS
```

✅ Test **passes** — you're in GREEN.

### 3. REFACTOR: Clean Up

```go
var ErrEmailRequired = errors.New("email required")

func (u *User) Validate() error {
    if u.Email == "" {
        return ErrEmailRequired
    }
    return nil
}
```

**Run again:**
```bash
go test ./internal/models/...
# PASS
```

✅ Tests still pass, code is cleaner — REFACTOR complete.

### 4. COMMIT

```bash
git add .
git commit -m "feat: add email validation to User"
```

### 5. REPEAT

Next test case: invalid email format.

## What Makes a Good Test

### Test One Thing

```go
// BAD: Tests multiple behaviors
func TestUser(t *testing.T) {
    // tests creation, validation, persistence...
}

// GOOD: Tests one behavior
func TestUserValidate_EmptyEmail(t *testing.T) {
    // only tests empty email case
}
```

### Clear Name

```go
// BAD: Unclear what it tests
func TestValidation(t *testing.T)

// GOOD: Name describes scenario and expectation
func TestUserValidate_MalformedEmail_ReturnsError(t *testing.T)
```

### Arrange-Act-Assert

```go
func TestUserValidate_ValidEmail(t *testing.T) {
    // Arrange
    user := User{Email: "test@example.com"}
    
    // Act
    err := user.Validate()
    
    // Assert
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### No Test Interdependencies

Each test should:
- Set up its own data
- Clean up after itself
- Pass regardless of other test results
- Run in any order

## Test Coverage Targets

| Component Type | Coverage |
|---------------|----------|
| Core business logic | 90%+ |
| Data validation | 100% |
| API handlers | 80%+ |
| Utilities | 70%+ |
| Generated code | 0% (skip) |

Coverage is a floor, not a ceiling. 100% coverage doesn't mean bug-free.

## Common Anti-Patterns

### 1. Test After

```
❌ Write code → Write tests that pass
✅ Write test → Watch it fail → Write code
```

Tests written after tend to verify implementation, not behavior.

### 2. Testing Implementation Details

```go
// BAD: Tests internal structure
func TestUserHasPrivateFields(t *testing.T) {
    // checks internal implementation
}

// GOOD: Tests behavior
func TestUserPersistsAndReloads(t *testing.T) {
    // checks observable behavior
}
```

### 3. Excessive Mocking

```go
// BAD: Mocks everything
func TestHandler(t *testing.T) {
    mockDB := new(MockDB)
    mockCache := new(MockCache)
    mockLogger := new(MockLogger)
    mockMetrics := new(MockMetrics)
    // ... 20 lines of mock setup
}

// GOOD: Minimal mocks, real behavior
func TestHandler(t *testing.T) {
    db := setupTestDB(t)
    handler := NewHandler(db)
    // test actual behavior
}
```

### 4. Brittle Tests

```go
// BAD: Breaks when unrelated things change
func TestUserJSON(t *testing.T) {
    expected := `{"id":"123","email":"test@example.com","created_at":"2024-01-01T00:00:00Z"}`
    // fails if field order changes
}

// GOOD: Tests semantics
func TestUserJSON(t *testing.T) {
    var result map[string]interface{}
    json.Unmarshal(data, &result)
    if result["email"] != "test@example.com" {
        t.Error("email mismatch")
    }
}
```

### 5. No Assertions

```go
// BAD: Test with no assertions
func TestUser(t *testing.T) {
    user := NewUser("test@example.com")
    _ = user.Validate()
    // "it didn't panic, so it works!"
}

// GOOD: Explicit assertions
func TestUser(t *testing.T) {
    user := NewUser("test@example.com")
    err := user.Validate()
    if err != nil {
        t.Fatalf("validation failed: %v", err)
    }
    if user.Email != "test@example.com" {
        t.Errorf("email mismatch: got %s", user.Email)
    }
}
```

## Edge Cases to Always Test

### Strings
- Empty string
- Very long string
- Unicode characters
- Whitespace only
- SQL/HTML injection attempts

### Numbers
- Zero
- Negative
- Very large (overflow)
- Boundaries (e.g., max int)

### Collections
- Empty
- One element
- Many elements
- Duplicates
- nil/null

### Time
- Now
- Past
- Future
- Timezone boundaries
- Daylight saving transitions

## Language-Specific Commands

### Go
```bash
go test ./...
go test -v ./internal/models/...
go test -cover ./...
go test -race ./...
```

### TypeScript/Bun
```bash
bun test
bun test --watch
bun test --coverage
```

### TypeScript/Node
```bash
npm test
npx jest --watch
npx jest --coverage
```

### Python
```bash
pytest
pytest -v
pytest --cov=src
```

## Model-Specific Notes

### For smaller models (7B-13B)

Be extra explicit about the cycle:
```
1. First, write test: [exact code]
2. Run it: [exact command]
3. Verify it fails (paste the failure)
4. Now write implementation: [exact code]
5. Run test again: [exact command]
6. Verify it passes (paste the success)
```

### For larger models (70B+)

Can handle the cycle implicitly, but enforce checkpoints:
- Must show test failure output
- Must show test pass output
- Must commit after each RED-GREEN-REFACTOR cycle

## Checklist

For each piece of functionality:

- [ ] Write test first
- [ ] Run test and verify it **fails**
- [ ] Write minimal implementation
- [ ] Run test and verify it **passes**
- [ ] Refactor if needed
- [ ] Run tests again (still pass)
- [ ] Commit with meaningful message
- [ ] Move to next functionality

## Integration with Other Skills

- **subagent-driven-development** — Implementers must follow TDD
- **verification-before-completion** — Tests are the verification
- **requesting-code-review** — Reviewers check test quality
