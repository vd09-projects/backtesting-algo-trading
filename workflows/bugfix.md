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

### Step 2 — Prior decision check (sub-agent)

Same as build.md Step 2. Read `workflows/agents/decision-lookup.md`. Fill slots with the
task context and the area of the codebase the bug is in. Spawn sub-agent.

A bug near accounting / fills / metrics often has prior decisions about correct behavior —
those are the ground truth for what "fixed" means.

Parse JSON. Update SESSION STATE: `verdicts.decision_lookup`. Write session file.

Log:
```
[AUTO] Step 2 — Decision lookup: N standing orders, M context files.
```

### Step 3 — Investigate and fix (sub-agent)

Read `workflows/agents/priya-build.md`. Bugs skip the separate plan step — pass the task
block directly as the "plan." Fill slots:
- `{{task_id}}`, `{{task_title}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from Step 2
- `{{marcus_verdict}}` — "not applicable — bug fix session"
- `{{plan_summary}}` — the bug description and what "fixed" means per acceptance criteria
- `{{files_to_create}}` and `{{files_to_modify}}` — leave blank; Priya identifies these
- `{{approach}}` — "Diagnose root cause, write regression test first, then fix"

Also add to the prompt: the methodology pivot instruction — if Priya finds this is a
methodology issue (not a code bug), she should flag it. The sub-agent will include this
in `flag`. When `flag` describes a methodology pivot, spawn a marcus-precheck sub-agent
(Step 3a), then re-spawn priya-build with Marcus's ruling in `{{marcus_verdict}}`.

The sub-agent handles the quality gate loop internally (quick by default; standard if the
fix touches `internal/engine/` or `internal/analytics/`).

Parse JSON. Update SESSION STATE. Write session file. Update `step_completed` = 3.

Log:
```
[AUTO] Step 3 — Fix: root cause diagnosed, regression test written, fix applied.
[AUTO] Step 4 — Quality gate (<level>): PASS.
[DECISION] Marcus [algorithm]: <if methodology pivot occurred>
[WARN] <any warnings with follow-up task IDs>
```

### Step 4 — Verify and close (orchestrator)

Check acceptance criteria against `verdicts.build`. The primary criterion for a bug is
"the bug no longer reproduces" plus any specific verification the task states.

If all criteria met: log `[CLOSED]`.
If any unmet: log `[FLAGGED]`, note for a follow-up task (session-end will create it).

Log:
```
[CLOSED] TASK-NNNN done. Bug verified fixed. Archived.
```

### Step 5 — Session end

Go to `session-end.md`.

Bug fixes rarely produce decision marks unless a methodology pivot occurred. The decision
harvest may come back empty — that is fine.
