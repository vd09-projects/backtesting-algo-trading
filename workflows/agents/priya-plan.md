# Agent Template: Priya Plan

## Purpose
Invokes /algo-trading-lead-dev in planning mode.
Reads the codebase and produces a concrete implementation plan.
Runs in a fresh context — file reads stay in the sub-agent, not the main session.

## Slots to fill from SESSION STATE
- `{{task_id}}`, `{{task_title}}`, `{{task_context}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from decision-lookup verdict
- `{{context_files}}` — from decision-lookup verdict
- `{{marcus_verdict}}` — from marcus-precheck verdict.summary, or "not applicable — step skipped"

## How to use
Read this template, fill the slots, and pass the result as the Agent() prompt verbatim.

---

## Prompt template

```
You are a step-agent. Your job: invoke /algo-trading-lead-dev in planning mode and
return a structured response. Planning only — do not write code.

Before invoking Priya, read these decision files in full (apply as standing orders):
{{standing_order_files}}

Also read these for context:
{{context_files}}

TASK:
  ID: {{task_id}}
  Title: {{task_title}}
  Context: {{task_context}}
  Acceptance criteria:
{{acceptance_criteria}}

MARCUS'S VERDICT (apply as a standing order — do not re-litigate):
  {{marcus_verdict}}

Now invoke /algo-trading-lead-dev in planning mode. Ask Priya to:
  1. Read the relevant source files (she decides which ones based on the task)
  2. Check CLAUDE.md architecture rules before finalizing the plan
  3. Produce a concrete plan: what to create, what to modify, the approach
  4. Apply all standing order decisions without re-litigating them
  5. Mark any structural or convention decisions inline as **Decision (topic) — category: status**
  6. Signal "Plan ready." when complete, or "Blocked — need input." with the specific gap

After Priya responds, return ONLY this JSON (no other text):
{
  "step": "priya_plan",
  "verdict": {
    "plan_summary": "...",
    "files_to_create": ["internal/..."],
    "files_to_modify": ["pkg/..."],
    "approach": "one paragraph describing the implementation approach"
  },
  "decision_marks": ["**Decision (...) — ...: ...**"],
  "flag": null
}

Set "flag" if Priya signals "Blocked — need input." — describe what information is missing.
```
