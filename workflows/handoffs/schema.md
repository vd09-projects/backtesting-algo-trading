# Handoff Schemas

Canonical schemas for SESSION STATE and all sub-agent responses.
The orchestrator writes SESSION STATE to `workflows/sessions/{today}-{TASK-ID}.json` after every step.
Sub-agents return structured JSON — the orchestrator reads only that JSON, nothing else.

---

## SESSION STATE

Written to `workflows/sessions/{today}-{TASK-ID}.json` after every completed step.

```json
{
  "session_date": "YYYY-MM-DD",
  "workflow": "build | evaluate | code-review | bugfix | review",
  "task_id": "TASK-NNNN",
  "task_title": "...",
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

Field notes:
- `step_completed` — 0 = session started, no steps done yet; N = last fully completed step
- `verdicts.*` — null until the step runs; set to the sub-agent's verdict payload on completion
- `decision_marks_pending` — `**Decision (...)**` marks collected across all steps; consumed by session-end
- `hard_stop_active` — null during normal operation; set to a string if a Hard STOP fired

---

## Sub-agent return schema (all agents)

Every sub-agent returns ONLY this JSON block, with no surrounding text:

```json
{
  "step": "<step name>",
  "verdict": {},
  "decision_marks": [],
  "flag": null
}
```

- `step` — identifies which step this response belongs to
- `verdict` — step-specific payload (see per-agent schemas below)
- `decision_marks` — any `**Decision (...)** ` marks produced by skills in this step
- `flag` — null, or a string describing uncertainty for the orchestrator to evaluate

The orchestrator evaluates `flag` autonomously. It hard-stops only on the three conditions
defined in `INDEX.md`. All other flags are resolved by applying prior decisions or making
the call and logging it.

---

## Per-agent verdict schemas

### decision-lookup
```json
{
  "standing_order_files": ["decisions/algorithm/2026-04-10-strategy-proliferation-gate.md"],
  "context_files": ["decisions/convention/2026-04-13-function-parameter-injection.md"]
}
```
`standing_order_files` — decisions to apply without re-litigating (status: accepted)  
`context_files` — relevant background; skills may revisit (status: experimental or tangential)

### marcus-precheck
```json
{
  "summary": "2-4 sentence verdict on the methodology question",
  "go_iterate_kill": "go | iterate | kill | n/a"
}
```
`go_iterate_kill` — set only for evaluate sessions; use "n/a" for build sessions

### priya-plan
```json
{
  "plan_summary": "...",
  "files_to_create": ["internal/walkforward/run.go"],
  "files_to_modify": ["pkg/strategy/strategy.go"],
  "approach": "one paragraph describing the implementation approach"
}
```

### priya-build
```json
{
  "build_summary": "...",
  "files_modified": ["internal/walkforward/run.go", "internal/walkforward/run_test.go"],
  "tests_written": ["TestRun_SyntheticKnownSignal"],
  "quality_gate": "PASS | FAIL",
  "quality_findings": []
}
```

### session-end
```json
{
  "tasks_created": [{"id": "TASK-XXXX", "title": "...", "priority": "high"}],
  "decisions_written": ["decisions/algorithm/2026-04-22-...md"],
  "suggested_commit": "verb what changed (TASK-NNNN)\n\n- bullet 1\n- bullet 2\n\nCo-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
}
```
