# Build Session

The most common session type (~40% of all sessions). A task is picked and implemented
autonomously from planning through close. No confirmation steps between stages.

## Trigger

User says "what's next," "start next task," picks a task by ID, or names work to implement.
Task type is a feature, refactor, or implementation — not a bug (see `bugfix.md`).

---

## Execution

### Step 1 — Pick the task

Read `tasks/BACKLOG.md`. Take the top item from **In Progress** (resume if one exists), or
the top item from **Up Next**. Log:

```
[AUTO] Step 1 — Task: TASK-NNNN "<title>" picked from <section>.
```

If the top task is blocked, take the next unblocked item and log the skip reason.
Extract: task ID, title, acceptance criteria, context, notes, any related decision references.

### Step 2 — Prior decision check

Before planning, scan the decision journal for relevant prior calls:

- Read `decisions/INDEX.md`. Filter entries whose tags overlap with the task's domain
  (e.g., a task touching analytics → look for `analytics`, `sharpe`, `drawdown` tags).
- Read the 1–3 most relevant decision files in full.
- Note which decisions apply and will be used as standing orders during the build.

Log:
```
[AUTO] Step 2 — Prior decisions: found N relevant entries. Applying: <titles>.
```

If zero relevant decisions exist, log that and proceed. It means the build has more
latitude — Priya and Marcus will establish new conventions as needed.

### Step 3 — Marcus pre-check (conditional)

**Only run this step if the task touches:** fill model, position sizing, performance metrics,
kill-switch logic, test plan methodology, or any backtest evaluation claim.

Check whether the methodology question is already answered by a prior `algorithm`-category
decision (Step 2 should have surfaced this). If yes, skip this step entirely.

If genuinely new: auto-invoke `/algo-trading-veteran`. Share the task context and the specific
methodology question. Marcus reads prior decisions and either:
- Applies an existing call → confirms, Step 3 ends
- Makes a new call → marks it inline as `[algorithm/experimental]`, Step 3 ends
- Hard STOP → "genuinely new methodology call" condition (see INDEX.md)

Log:
```
[AUTO] Step 3 — Marcus pre-check: <skipped (prior decision applies) | new call made>.
[DECISION] Marcus [algorithm]: <one-line summary if a new call was made>
```

### Step 4 — Plan

Auto-invoke `/algo-trading-lead-dev`. Pass: task ID, acceptance criteria, context, relevant
prior decisions from Step 2, Marcus's call from Step 3 (if any).

Priya reads the codebase, checks `decisions/` for prior structural calls, and produces a plan.

**If `Plan ready.`** → log and proceed to Step 5.

**If `Blocked — need input.`** → check what is needed:
- Methodology question → run Step 3 (Marcus) now, return to Priya with answer, resume plan
- Requirements question → Hard STOP: state the gap and the two most likely interpretations
- Data question → Hard STOP: state what data detail is missing

Log:
```
[AUTO] Step 4 — Plan: complete. Approach: <one-sentence summary of Priya's approach>.
```

### Step 5 — Build loop

Auto-invoke `/algo-trading-lead-dev` in build mode. Pass: the approved plan, all context.
Priya writes tests first, then implementation, then marks any decisions inline.

This loop runs until quality gate is clean or a Hard STOP fires:

```
while true:
    Priya builds (may be multiple turns)

    if Priya says "Ready for review — flagging for Marcus.":
        invoke Marcus with the flagged item
        if Marcus overrides: return to Priya with override, she iterates
        continue loop

    if Priya says "Blocked — need input.":
        evaluate: requirements gap? → Hard STOP
        otherwise route as in Step 4 and resume

    run quality gate (standard for internal/ changes, quick otherwise)

    if gate has lint/format failures only:
        auto-fix (golangci-lint --fix), re-run gate, continue

    if gate has blocker findings:
        return to Priya in iterate mode with specific findings (round N of 2)
        if round 2 still has blockers: Hard STOP — unresolvable blocker

    if gate clean:
        break
```

Log for each iteration:
```
[AUTO] Step 5 — Build: complete. Tests written first (TDD). Quality gate: PASS.
[DECISION] Priya [<category>]: <any inline decision marks Priya made>
[WARN] <any quality gate warnings, with follow-up task IDs>
```

### Step 6 — Verify and close

Check every acceptance criterion in the task block against what was built. For each criterion:
- If met: mark `[x]`
- If not met: log `[FLAGGED]` and create a follow-up task for the gap

If all criteria met: invoke `/task-manager` — mark task done, archive it.
Log:
```
[CLOSED] TASK-NNNN done. All criteria met. Archived.
```

If any criteria unmet: close the task for what it achieved, log the gaps as `[FLAGGED]`,
create follow-up tasks. Do not leave the task lingering in "In Progress."

### Step 7 — Session end

Go to `session-end.md`.
