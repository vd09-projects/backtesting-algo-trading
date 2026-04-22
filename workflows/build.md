# Build Session

The most common session type (~40% of all sessions). A task is picked and implemented
autonomously from planning through close. Every heavy step runs in a sub-agent.

## Trigger

User says "what's next," "start next task," picks a task by ID, or names work to implement.
Task type is a feature, refactor, or implementation — not a bug (see `bugfix.md`).

---

## SESSION STATE initialization

Initialize SESSION STATE at the start:

```json
{
  "session_date": "YYYY-MM-DD",
  "workflow": "build",
  "task_id": null,
  "task_title": null,
  "step_completed": 0,
  "verdicts": {
    "decision_lookup": null,
    "marcus": null,
    "priya_plan": null,
    "build": null,
    "session_end": null
  },
  "execution_log": [],
  "decision_marks_pending": [],
  "hard_stop_active": null
}
```

Check `workflows/.session-state.json` first — if it exists, load it and resume from
`step_completed + 1` per the resume protocol in `INDEX.md`.

---

## Execution

### Step 1 — Pick the task (orchestrator reads directly)

Read `tasks/BACKLOG.md`. Take the top item from **In Progress** (resume if one exists), or
the top item from **Up Next**. Extract: task ID, title, acceptance criteria, context.

Update SESSION STATE: `task_id`, `task_title`. Write `.session-state.json`.

If the top task is blocked, take the next unblocked item and log the skip reason.

Log:
```
[AUTO] Step 1 — Task: TASK-NNNN "<title>" picked from <section>.
```

### Step 2 — Decision lookup (sub-agent)

Read `workflows/agents/decision-lookup.md`. Fill slots:
- `{{task_id}}` — from SESSION STATE
- `{{task_title}}` — from SESSION STATE
- `{{task_context}}` — task context paragraph from BACKLOG.md

Spawn sub-agent via Agent() tool. Parse returned JSON. Update SESSION STATE:
`verdicts.decision_lookup` = the verdict payload. Append to `decision_marks_pending`
if any marks returned. Write `.session-state.json`. Update `step_completed` = 2.

Log:
```
[AUTO] Step 2 — Decision lookup: N standing orders, M context files.
```

### Step 3 — Marcus pre-check (sub-agent, conditional)

**Only run this step if the task touches:** fill model, position sizing, performance metrics,
kill-switch logic, test plan methodology, walk-forward, or any backtest evaluation claim.

If the methodology question is already answered by a standing order from Step 2, skip this
step and log `[AUTO] Step 3 — Marcus: skipped (prior decision applies: <title>)`.

If genuinely new: read `workflows/agents/marcus-precheck.md`. Fill slots:
- `{{task_id}}`, `{{task_title}}`, `{{task_context}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from `verdicts.decision_lookup.standing_order_files`
- `{{context_files}}` — from `verdicts.decision_lookup.context_files`
- `{{methodology_question}}` — the specific question inferred from the task

Spawn sub-agent exactly once. If the returned JSON is structurally valid, accept it — do not
re-spawn. If `flag` is non-null: evaluate it. If it meets Hard STOP condition 2 (genuinely new
methodology with no basis to decide) → Hard STOP. Otherwise resolve autonomously and continue.

Update SESSION STATE: `verdicts.marcus`, append any `decision_marks` to `decision_marks_pending`.
Write `.session-state.json`. Update `step_completed` = 3.
Immediately proceed to Step 4 — do not output anything to the user.

Log:
```
[AUTO] Step 3 — Marcus: <skipped (prior decision applies) | new call made>.
[DECISION] Marcus [algorithm]: <one-line summary if a new call was made>
```

### Step 4 — Plan (sub-agent)

Read `workflows/agents/priya-plan.md`. Fill slots:
- `{{task_id}}`, `{{task_title}}`, `{{task_context}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from `verdicts.decision_lookup.standing_order_files`
- `{{context_files}}` — from `verdicts.decision_lookup.context_files`
- `{{marcus_verdict}}` — from `verdicts.marcus.summary`, or "not applicable — step skipped"

Spawn sub-agent. Parse JSON. If `flag` is non-null: evaluate it.
- Methodology question → spawn Marcus sub-agent (Step 3 pattern), return to Priya sub-agent
  with his answer by re-running Step 4 with the Marcus answer filled in
- Requirements gap → Hard STOP: state the gap and two most likely interpretations
- Data question → Hard STOP: state what data detail is missing

Update SESSION STATE: `verdicts.priya_plan`, append any `decision_marks`.
Write `.session-state.json`. Update `step_completed` = 4.

Log:
```
[AUTO] Step 4 — Plan: complete. Approach: <one-sentence from verdict.approach>.
```

### Step 5 — Build loop (sub-agent)

Read `workflows/agents/priya-build.md`. Fill slots:
- `{{task_id}}`, `{{task_title}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from `verdicts.decision_lookup.standing_order_files`
- `{{marcus_verdict}}` — from `verdicts.marcus.summary`, or "not applicable"
- `{{plan_summary}}`, `{{files_to_create}}`, `{{files_to_modify}}`, `{{approach}}`
  — from `verdicts.priya_plan`

Spawn sub-agent exactly once. The sub-agent owns the full build+gate loop — it runs tests
and lint itself and does not return until both pass. Do NOT run any Bash commands, file
reads, or verification steps after the sub-agent returns. The returned JSON is ground truth.
Parse it. If `flag` is non-null: evaluate it.
- Unresolvable blocker after 2 rounds → Hard STOP

Update SESSION STATE: `verdicts.build`, append any `decision_marks` to `decision_marks_pending`.
Write `.session-state.json`. Update `step_completed` = 5.

Log:
```
[AUTO] Step 5 — Build: complete. Quality gate: PASS. Files: <list from verdict.files_modified>.
[DECISION] Priya [<category>]: <any decision marks Priya made>
[WARN] <any quality findings, with follow-up task IDs>
```

### Step 6 — Verify and close (orchestrator)

Check every acceptance criterion in the task block against `verdicts.build`.
For each criterion:
- If met: mark `[x]`
- If not met: log `[FLAGGED]` and note it for a follow-up task (session-end will create it)

Go to session-end.md.
