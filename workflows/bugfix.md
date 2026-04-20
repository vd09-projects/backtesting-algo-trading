# Bug Fix Session

User picks a bug and it is fixed autonomously. Shorter than a build session — fewer decisions,
faster iteration. (~10% of sessions.)

## Trigger

Task type is a bug, or user describes a specific issue: "this is broken," "the output is wrong,"
"there's a race condition in X."

---

## Execution

### Step 1 — Pick the bug

Read `tasks/BACKLOG.md`. Take the named bug task, or if none specified, the top bug in the
backlog. Extract: task ID, what's broken, acceptance criteria, reproduction steps if available.

Log:
```
[AUTO] Step 1 — Bug: TASK-NNNN "<title>" picked.
```

### Step 2 — Prior decision check

Same as build.md Step 2. Scan the decision journal for prior calls relevant to the broken area.
A bug near accounting / fills / metrics often has prior decisions about correct behavior — those
are the ground truth for what "fixed" means.

Log:
```
[AUTO] Step 2 — Prior decisions: N relevant entries. Applying: <titles>.
```

### Step 3 — Investigate and fix

Auto-invoke `/algo-trading-lead-dev` in build mode (bugs don't need a separate plan step unless
the fix is architecturally significant). Pass: task ID, what's broken, prior decisions.

Priya investigates: reads relevant code, diagnoses root cause, writes a regression test first,
then the fix.

**The methodology pivot:** If Priya finds this isn't a code bug but a methodology issue
(fill model producing wrong results, metric computation differs from spec, behavior is correct
code but wrong trading logic), she will flag it. When this happens:

Auto-invoke `/algo-trading-veteran` with what Priya found. Marcus evaluates whether the current
methodology is correct or needs changing. If he recommends a change, it is an `algorithm`
decision — log it. Return to Priya with Marcus's ruling.

Priya ends with `Ready for review.` or `Blocked — need input.`

- If `Blocked — need input.` and it is a requirements gap → Hard STOP
- Otherwise route the blocker (methodology → Marcus, data → Hard STOP) and resume

Log:
```
[AUTO] Step 3 — Fix: root cause diagnosed, regression test written, fix applied.
[DECISION] Marcus [algorithm]: <if methodology pivot occurred>
```

### Step 4 — Quality check

Auto-run `/go-quality-review`:
- Default level: **quick** (lint + race detection)
- Bump to **standard** if the fix touched: accounting, fills, metrics computation, event loop,
  position sizing, or any file in `internal/engine/` or `internal/analytics/`

If lint/format failures: auto-fix (`golangci-lint --fix`), re-run.
If blocker findings: return to Priya in iterate mode (max 2 rounds). Hard STOP if unresolved.
If clean: proceed.

Log:
```
[AUTO] Step 4 — Quality gate (<level>): PASS.
[WARN] <any warnings with follow-up task IDs>
```

### Step 5 — Verify and close

Check acceptance criteria. The primary criterion for a bug is "the bug no longer reproduces"
plus any specific verification the task states.

If all criteria met: auto-close, auto-archive.
If any unmet: log `[FLAGGED]`, create follow-up task, close this task for what it achieved.

Log:
```
[CLOSED] TASK-NNNN done. Bug verified fixed. Archived.
```

### Step 6 — Session end

Go to `session-end.md`.

Bug fixes rarely produce decision marks unless a methodology pivot occurred. The decision
harvest may come back empty — that is fine.
