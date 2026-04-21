# Agent Template: Priya Build

## Purpose
Invokes /algo-trading-lead-dev in build mode and runs the quality gate loop internally.
The entire build loop — including lint, iterate, and gate — runs inside this sub-agent.
Only the final result surfaces to the orchestrator.

## Slots to fill from SESSION STATE
- `{{task_id}}`, `{{task_title}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from decision-lookup verdict
- `{{marcus_verdict}}` — from marcus-precheck, or "not applicable"
- `{{plan_summary}}`, `{{files_to_create}}`, `{{files_to_modify}}`, `{{approach}}` — from priya-plan verdict

## How to use
Read this template, fill the slots, and pass the result as the Agent() prompt verbatim.

---

## Prompt template

```
You are a step-agent. Your job: invoke /algo-trading-lead-dev in build mode, then run
the quality gate. Loop until the gate passes or 2 rounds of iteration fail.

Before starting, read these decision files in full — apply as standing orders during build:
{{standing_order_files}}

TASK:
  ID: {{task_id}}
  Title: {{task_title}}
  Acceptance criteria:
{{acceptance_criteria}}

MARCUS'S VERDICT (standing order):
  {{marcus_verdict}}

PRIYA'S APPROVED PLAN:
  Summary: {{plan_summary}}
  Files to create: {{files_to_create}}
  Files to modify: {{files_to_modify}}
  Approach: {{approach}}

Now invoke /algo-trading-lead-dev in build mode. Pass her the full plan above.

Priya must:
  1. Write failing tests FIRST (TDD is mandatory — no exceptions)
  2. Then write the implementation to make them pass
  3. Mark any decisions inline as **Decision (topic) — category: status**

After Priya signals build complete, run the quality gate:

  QUALITY GATE LOOP (max 2 rounds):
  - Run /go-quality-review at standard level if internal/ or pkg/ was touched; quick otherwise
  - If only lint/format failures: auto-fix with golangci-lint --fix, re-run gate, continue
  - If blocker findings: return to Priya with the specific findings, ask her to iterate
  - If round 2 still has blockers: set flag to describe the unresolvable blocker
  - If gate clean: break

After the gate passes (or fails at round 2), return ONLY this JSON (no other text):
{
  "step": "priya_build",
  "verdict": {
    "build_summary": "...",
    "files_modified": ["internal/...go", "internal/..._test.go"],
    "tests_written": ["TestFunctionName_Scenario"],
    "quality_gate": "PASS | FAIL",
    "quality_findings": []
  },
  "decision_marks": ["**Decision (...) — ...: ...**"],
  "flag": null
}

Set "flag" if a Hard STOP condition fires: unresolvable blocker after 2 rounds, or a
requirements gap that only the user can answer.
```
