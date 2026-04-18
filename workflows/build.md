# Build Session

The most common session type (~40% of all sessions). User picks a task and implements it.

## Trigger

User says "what's next," picks a task from the backlog, or names a specific task to work on.
Task type is a feature, refactor, or implementation — not a bug (see `bugfix.md`), not a review.

## Flow

### Step 1 — Pick the task

Invoke `/task-manager` — ask for the next task or the specific task the user named.
The task-manager shows: task ID, title, acceptance criteria, context, any blockers.

**Carry forward:** task ID, acceptance criteria, any notes or related decision references.

### Step 2 — Plan

Invoke `/algo-trading-lead-dev` — share the task ID, acceptance criteria, and any relevant
context. Priya plans the approach.

Priya will read the codebase, check `decisions/` for prior calls, and produce a plan.
She ends with `Plan ready.` or `Blocked — need input.`

**If `Blocked — need input`:** Read what Priya needs. Usually one of:
- A methodology question → suggest invoking `/algo-trading-veteran` (Marcus) to answer it.
  After Marcus answers, return to Priya with his answer and resume planning.
- A requirements question → ask the user directly. After they answer, resume planning.
- A data question → the user needs to provide data details. After they do, resume planning.

**If `Plan ready`:** Ask the user: "Approve the plan and start building?"

### Step 3 — Build

Continue with `/algo-trading-lead-dev` — Priya switches to build mode. She writes code,
tests, and marks any decisions inline.

This is usually the longest step. Priya may need multiple turns. She may:
- Discover something mid-build that changes the plan. She'll surface it and ask to re-plan.
- Hit a methodology question she can't answer. She'll say she's blocked. Route to Marcus
  (same as Step 2's blocked handling), then return to Priya.
- Find that the task is bigger than expected and should be split. Suggest invoking
  `/task-manager` to decompose the task, then continue building the first subtask.

She ends with one of:
- `Ready for review.` → go to Step 4 or Step 5.
- `Ready for review — flagging for Marcus.` → go to Step 4a before Step 4/5.
- `Blocked — need input.` → handle the block (same pattern as Step 2).

### Step 4a — Marcus review (only if flagged)

Priya flagged something for Marcus — usually a methodology-adjacent implementation choice
where his sign-off matters (fill model, embargo size, sizing logic, test plan fidelity).

Invoke `/algo-trading-veteran` — share what Priya flagged and why. Marcus reviews and either
confirms Priya's approach or overrides it with a different recommendation.

**If Marcus confirms:** proceed to Step 4 or 5.
**If Marcus overrides:** return to Priya with Marcus's decision. She iterates on the override
(usually a small change), then reaches `Ready for review.` again. Proceed to Step 4 or 5.

### Step 4 — Quality gate (optional)

Not every build needs a formal review. Use the quality gate when:
- The build touches structural code (new packages, interface changes, engine internals)
- The build modifies invariant-sensitive code (accounting, fills, metrics, event loop)
- The user explicitly asks for a review
- The task's acceptance criteria include "passes pre-merge review"

Skip it when:
- Small config changes, documentation, test-only additions
- The user says "skip the reviewer"

Invoke `/go-quality-review` — at the appropriate level (standard for most, deep or pre-merge
for structural/invariant changes).

**If blockers found:** return to `/algo-trading-lead-dev` in iterate mode. Share the specific
findings. Priya fixes them. She may push back on a finding with reasoning — if she does
and marks a tradeoff decision, that's fine. After iteration, re-run the reviewer to verify.

**If clean:** proceed to Step 5.

### Step 5 — Verify and close

Invoke `/task-manager` — ask it to verify the acceptance criteria against what was built.
The task-manager checks each criterion, marks done or surfaces gaps.

**If all criteria met:** task-manager marks the task done and archives it.
**If gaps remain:** user decides — continue building (return to Step 3), or create a follow-up
task for the gap and close this task as done for what it achieved.

### Step 6 — Session end

Go to `session-end.md`.
