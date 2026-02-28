---
name: brainstorming
description: Mandatory design thinking phase before any implementation
tags:
  - design
  - planning
  - architecture
---

# Brainstorming Skill

## Purpose

Before writing ANY code, you MUST complete a thorough brainstorming phase. This ensures the solution is well-designed and considers edge cases.

## Process

### 1. Understand the Problem
- What is the user actually asking for?
- What are the explicit requirements?
- What are the implicit requirements?

### 2. Explore the Context
- What existing code/architecture is relevant?
- What constraints exist (performance, compatibility, etc.)?
- What similar problems have been solved before?

### 3. Ask Clarifying Questions
Before proceeding, ask 2-3 critical questions:
- Ambiguities that need resolution
- Trade-offs that need user input
- Assumptions that should be validated

### 4. Design Alternatives
Present at least 2-3 approaches:
- Describe each approach briefly
- List pros and cons
- Recommend one with justification

### 5. Get Approval
Present your recommended design and get explicit approval before implementing.

## Rules

1. NEVER skip brainstorming, even for "simple" tasks
2. ALWAYS present your design before coding
3. ASK questions when uncertain
4. DOCUMENT your reasoning

## Example

User: "Add user authentication"

BAD: Immediately start writing auth code

GOOD:
1. "Before I implement authentication, I have a few questions..."
2. "Here are 3 approaches: JWT, sessions, OAuth..."
3. "I recommend JWT because... Do you approve?"
4. [User approves]
5. Now implement
