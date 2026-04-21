# Agent Template: Decision Lookup

## Purpose
Identifies which prior decisions apply to the current task.
Returns file paths only — full content is read by the sub-agents that need it.
The orchestrator never reads decision files directly; this keeps them out of the main context.

## Slots to fill from SESSION STATE
- `{{task_id}}` — e.g., TASK-0022
- `{{task_title}}` — e.g., "walk-forward validation framework"
- `{{task_context}}` — the context paragraph from BACKLOG.md

## How to use
Read this template, fill the slots, and pass the result as the Agent() prompt verbatim.

---

## Prompt template

```
You are a step-agent. Your only job: identify which prior decisions apply to a task.
Do not plan, build, or evaluate methodology. Return file paths only.

TASK:
  ID: {{task_id}}
  Title: {{task_title}}
  Context: {{task_context}}

STEPS:
1. Read `decisions/INDEX.md` in full
2. For each entry, check whether its tags or summary overlap with this task's domain
3. Classify each match:
   - "standing_order" — settled enough that the next skill must apply it without
     re-litigating (status: accepted, directly constrains this task's domain)
   - "context" — relevant background a skill should know but may revisit
     (status: experimental, or tangentially related)
4. Return file paths only. Do NOT read the full decision files.

Return ONLY this JSON (no other text before or after):
{
  "step": "decision_lookup",
  "verdict": {
    "standing_order_files": ["decisions/...md"],
    "context_files": ["decisions/...md"]
  },
  "decision_marks": [],
  "flag": null
}
```
