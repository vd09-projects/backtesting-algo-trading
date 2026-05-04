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
  "workflow": "build | evaluate | evaluation-run | design | code-review | bugfix | review",
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
  "hard_stop_active": null,
  "preflight_passed": false,
  "is_bugfix": false,
  "quality_review_round": 0,
  "prior_rounds_findings": []
}
```

Field notes:
- `step_completed` — 0 = session started, no steps done yet; N = last fully completed step. Sub-step preflight gates (build-session Step 1.5) do NOT advance this — they set `preflight_passed` instead.
- `verdicts.*` — null until the step runs; set to the sub-agent's verdict payload on completion
- `decision_marks_pending` — `**Decision (...)**` marks collected across all steps; consumed by session-end
- `hard_stop_active` — null during normal operation; set to a string if a Hard STOP fired
- `preflight_passed` — build-session Step 1.5 gates (wrong-agent redirect, bugfix detection, strategy registration, sentinel freshness) all passed
- `is_bugfix` — task source is bug; build-session collapses the plan step
- `quality_review_round` — incremented each time `go-quality-review-runner` is spawned
- `prior_rounds_findings` — accumulated `[{file, line, description}]` from every prior quality-review verdict in this session; passed to priya-iterate so it can detect recurring findings

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
`quality_gate` here = per-package `go build ./<pkg>/...` + `go test -race ./<pkg>/...` only. Repo-wide
`golangci-lint run ./...` runs in Step 5b via `go-quality-review-runner`, not here. Keep field name
for backward compatibility; semantics narrowed.

### go-quality-review-runner
```json
{
  "gate_status": "clean | warnings_cosmetic | warnings_blocking | failed",
  "quality_gate": "PASS | FAIL",
  "blocker_count": 0,
  "warning_count": 0,
  "warning_cosmetic_count": 0,
  "warning_code_change_count": 0,
  "suggestion_count": 0,
  "benchmark_ran": false,
  "benchmark_ns_per_op": null,
  "findings": [
    {
      "severity": "blocker | warning | suggestion",
      "warning_class": "cosmetic | code_change_required | null",
      "file": "...",
      "line": 0,
      "description": "...",
      "why": "...",
      "fix": "..."
    }
  ],
  "sentinel_written": true
}
```
Orchestrator branches on `gate_status` (see go-quality-review-runner.md). `quality_gate` retained
as legacy field: PASS ⟺ `gate_status ∈ {clean, warnings_cosmetic}`. `benchmark_ran` is true for
changes touching `internal/engine/`; `benchmark_ns_per_op > 1_000_000` is recorded as a blocker.

### priya-iterate
```json
{
  "status": "RESOLVED | PARTIAL | BLOCKED",
  "files_modified": ["..."],
  "resolved_findings": ["..."],
  "unresolved_findings": ["..."],
  "recurring_findings": ["..."],
  "blocker_count_remaining": 0,
  "follow_up_suggestions": ["..."],
  "decision_marks": ["..."]
}
```
`recurring_findings` populated when `iterate_round >= 2` and a finding's `file:line` matches a
prior round. Non-empty `recurring_findings` forces `status = "BLOCKED"` regardless of
`blocker_count_remaining` — prevents ping-pong.

### session-end
```json
{
  "tasks_created": [{"id": "TASK-XXXX", "title": "...", "priority": "high"}],
  "decisions_written": ["decisions/algorithm/2026-04-22-...md"],
  "suggested_commit": "verb what changed (TASK-NNNN)\n\n- bullet 1\n- bullet 2\n\nCo-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
}
```
