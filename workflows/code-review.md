# Code Review Session

User wants existing code reviewed without a specific task driving it. Could be a periodic
quality pass, a pre-merge gate on a branch, or "I feel like this package is getting messy."
(~5% of sessions.)

## Trigger

User says "review this code," "run a quality check on X," "pre-merge review," "what's wrong
with this package," or names a specific package or file to review.

## Flow

### Step 1 — Run the review

Invoke `/go-quality-review` — at the level the user requested or implied.

- "quick check" / "just lint" → quick
- "review this" / "PR review" → standard
- "deep review" / "this package is critical" → deep
- "pre-merge" / "ready to ship" → pre-merge
- Ambiguous → default to standard, mention deeper levels are available

The reviewer produces a findings report: blockers, warnings, suggestions.

### Step 2 — Handle findings

Three paths based on what the user wants:

**Fix now:** Invoke `/algo-trading-lead-dev` in iterate mode — share the specific findings.
Priya addresses them. She may:
- Fix everything the reviewer flagged.
- Push back on specific findings with reasoning. If she does, she marks a `tradeoff` decision
  explaining why the override is intentional (e.g., "this function is long because splitting
  it harms readability").
- Discover deeper issues while fixing surface findings. If structural, she may need to re-plan.

After Priya finishes iterating, re-run `/go-quality-review` at the same level to verify the
fixes. If new issues surface from the fixes (happens occasionally — fixing one thing reveals
another), iterate again. Usually converges in 1-2 rounds.

**Track for later:** Invoke `/task-manager` — create tasks for the findings the user wants to
address but not right now. Each finding becomes a task with the reviewer's severity as priority
guidance (blocker → high, warning → medium, suggestion → low).

**Accept as-is:** User acknowledges the findings and chooses not to act. No further steps.
If the user explicitly overrides a finding, suggest Priya mark a `tradeoff` decision so the
reasoning is preserved (otherwise the same finding will fire again on the next review with
no record of why it was previously accepted).

### Step 3 — Session end

Go to `session-end.md` — but only if decisions were marked during iteration. If the review
was clean or the user just tracked tasks, the harvest will come back empty. That's fine.
