# Bug Fix Session

User picks a bug from the backlog and fixes it. Usually shorter and more focused than a build
session — fewer decisions, faster iteration. (~10% of sessions.)

## Trigger

Task type is a bug. Or user describes a specific issue: "this is broken," "the output is wrong,"
"there's a race condition in X."

## Flow

### Step 1 — Pick the bug

Invoke `/task-manager` — pick the bug task. The task-manager shows: task ID, what's broken,
acceptance criteria (usually "the bug no longer reproduces" + a specific verification).

**Carry forward:** task ID, what's broken, reproduction steps if available.

### Step 2 — Investigate and fix

Invoke `/algo-trading-lead-dev` — share the bug details. Priya investigates: reads the relevant
code, reproduces the issue, diagnoses the root cause, writes the fix, adds a regression test.

This is usually 1-3 turns. Priya is in build mode (bugs don't need a separate plan step unless
the fix is architecturally significant).

**The methodology pivot:** Sometimes what looks like a code bug is actually a methodology issue.
Examples:
- "The backtest results look wrong" → the fill model is producing incorrect results (not a code
  bug — the code does what it says; the model is wrong)
- "The Sharpe number doesn't match my Python notebook" → different computation methodology
  between Go and Python (not a bug — a methodology question about which computation is correct)
- "The equity curve has a weird jump" → the strategy is behaving as coded, but the behavior
  is wrong from a trading perspective

If Priya identifies this: she'll say something like "this isn't a code bug — the [fill model /
computation / behavior] is working as implemented, but the implementation may be wrong from a
methodology standpoint. Flagging for Marcus."

When this happens, invoke `/algo-trading-veteran` — share what Priya found. Marcus evaluates
whether the current methodology is correct or needs changing. If he recommends a change, that's
an `algorithm` decision. Return to Priya with Marcus's recommendation and she implements the
corrected version.

### Step 3 — Quick quality check

Invoke `/go-quality-review` — at the `quick` level (lint + race detection). Bug fixes don't
usually need a deep review unless the fix is structurally significant.

**If the fix touched invariant-sensitive code** (accounting, fills, metrics, event loop), bump
to `standard` or `deep` level.

**If issues found:** return to Priya, iterate, re-check.
**If clean:** proceed.

### Step 4 — Verify and close

Invoke `/task-manager` — verify the bug no longer reproduces and acceptance criteria are met.
Mark done.

### Step 5 — Session end

Go to `session-end.md`.

Bug fixes rarely produce decision marks unless the fix involved a tradeoff or the methodology
pivot happened. The journal harvest may come back empty — that's fine.
