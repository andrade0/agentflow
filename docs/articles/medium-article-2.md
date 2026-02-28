# Stop Letting Your AI Coding Agent Guess: The Case for Mandatory Brainstorming

*Why the first 5 minutes of any AI coding task should be questions, not code*

---

You open your AI coding assistant. You type:

> "Add user authentication"

What happens next usually looks like this:

```python
# The AI immediately writes 200 lines of auth code
class AuthService:
    def __init__(self):
        self.secret_key = "your-secret-key"  # Security issue
        ...
```

Three problems:
1. It assumed JWT (maybe you wanted OAuth)
2. It assumed your tech stack (maybe you're not using Python)
3. It made security decisions without asking

**This is guess-and-fix programming, and it's wasting your time.**

## The Cost of Assumptions

Every assumption the AI makes has a cost:

| Assumption | Cost if Wrong |
|------------|---------------|
| Tech stack | Rewrite everything |
| Architecture pattern | Refactor later |
| Security model | Vulnerability |
| Scope | Over-engineering or missing features |

When the AI guesses wrong, you either:
- Spend time correcting it (context thrashing)
- Accept the wrong choice (technical debt)
- Start over (wasted work)

## The Brainstorming Solution

What if the AI **couldn't write code** until it understood the problem?

That's the idea behind mandatory brainstorming. Before any implementation, the AI must:

1. **Explore context** — What does the project look like?
2. **Ask questions** — What do you actually need?
3. **Propose options** — Here are 2-3 approaches
4. **Get approval** — Do you approve this design?

Only then can it write code.

## How It Works in AgentFlow

AgentFlow has a brainstorming skill with a **hard gate**:

```markdown
<HARD-GATE>
Do NOT write any code until:
1. You have presented a design
2. The user has explicitly approved it

This applies to EVERY project regardless of perceived simplicity.
</HARD-GATE>
```

When you run:

```bash
agentflow run "add user authentication"
```

The AI doesn't immediately start coding. Instead:

### Step 1: Explore Context

```
I see you're working on a Go project with:
- Gin web framework
- PostgreSQL database
- Existing User model
- No current auth implementation

Is this accurate?
```

### Step 2: Ask One Question at a Time

```
Before I design the authentication system, I have some questions.

First: Should sessions be stateless (JWT) or stateful (server-side)?
```

*One question per message.* Not a wall of 10 questions.

After you answer, it asks the next:

```
Got it, JWT-based. 

Should tokens be short-lived (15 minutes) with refresh tokens,
or long-lived (7 days)?
```

### Step 3: Propose Approaches

After understanding requirements:

```
Based on your answers, here are three approaches:

## Approach A: Standard JWT
- Access token only, 1-hour expiry
- Simple, stateless
- Con: Can't revoke tokens

## Approach B: JWT + Refresh
- Short access (15min) + long refresh (7 days)
- Can revoke refresh tokens
- Industry standard

## Approach C: JWT + Redis Blacklist
- Like B, but can also revoke access tokens
- Most secure
- Adds Redis dependency

I recommend **Approach B** for your use case because:
- You have a database for refresh tokens
- You don't need instant revocation
- Simpler than C, more secure than A

Do you approve?
```

### Step 4: Get Explicit Approval

The AI waits for explicit approval:

- ✅ "Yes, looks good"
- ✅ "Approved"
- ✅ "Let's do it"
- ❌ "Sure" (ambiguous — will ask for clarification)
- ❌ "I guess" (ambiguous)

Only after approval does implementation begin.

## Why One Question at a Time?

Traditional approach:
```
Here are my questions:
1. What auth method?
2. Token duration?
3. Where to store users?
4. Need password reset?
5. Need email verification?
6. OAuth providers?
7. ...
```

This overwhelms. Users skim, miss questions, give incomplete answers.

AgentFlow's approach:
```
Q: What auth method?
A: JWT

Q: Token duration?
A: Short with refresh

Q: Need password reset?
A: Yes
...
```

One at a time. Full attention. Better answers.

## "But This Is Slow!"

Is it though?

| Approach | Time to First Code | Rework Time | Total |
|----------|-------------------|-------------|-------|
| Guess immediately | 30 seconds | 45 minutes | 45.5 min |
| Brainstorm first | 5 minutes | 5 minutes | 10 min |

Brainstorming is **faster** because:
- Catches wrong assumptions early
- Aligns expectations before work starts
- Reduces context-switching
- Eliminates "wait, I wanted X not Y"

## "Simple Tasks Don't Need This"

The brainstorming skill addresses this:

```markdown
## Anti-Pattern: "This Is Too Simple To Need A Design"

Every project goes through this process. A todo list, a single-function 
utility, a config change — all of them.

"Simple" projects are where unexamined assumptions cause the most 
wasted work.

The design can be short (a few sentences for truly simple projects), 
but you MUST present it and get approval.
```

Simple request:
```
> Add a health check endpoint

I'll add a GET /health endpoint that returns {"status":"healthy"}.
It will be in internal/handlers/health.go and use the standard response format.

Approve?

> Yes
```

30 seconds of brainstorming. Zero rework.

## Try It

```bash
# Install AgentFlow
brew install agentflow/tap/agentflow

# Use free model
agentflow config set provider ollama
agentflow config set model llama3.3:70b

# Run with mandatory brainstorming
agentflow run "add user authentication"
```

Watch the AI ask questions instead of guessing. Feel the difference.

## Conclusion

Your AI coding assistant is not a mind reader. Stop treating it like one.

**Mandatory brainstorming:**
- Catches assumptions before they become bugs
- Saves time by avoiding rework
- Produces better designs through dialogue
- Works even with free, local models

The first 5 minutes shouldn't be code. They should be questions.

---

*AgentFlow is open source: [github.com/agentflow/agentflow](https://github.com/agentflow/agentflow)*

#AI #CodingAssistants #DeveloperProductivity #SoftwareDesign #OpenSource
