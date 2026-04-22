# Agent Template: Marcus Pre-Check

## Purpose
Invokes /algo-trading-veteran on a specific methodology question.
Runs in a fresh context — reads decision files directly, no inherited history from main session.

## Slots to fill from SESSION STATE
- `{{task_id}}`, `{{task_title}}`, `{{task_context}}`, `{{acceptance_criteria}}`
- `{{standing_order_files}}` — from decision-lookup verdict (newline-separated paths)
- `{{context_files}}` — from decision-lookup verdict
- `{{methodology_question}}` — the specific question the orchestrator identified for Marcus
  (e.g., "Should walk-forward use fixed-ratio or expanding-window splits?")

## How to use
Read this template, fill the slots, and pass the result as the Agent() prompt verbatim.

---

## Prompt template

```
You are a step-agent. Your job: invoke /algo-trading-veteran on a methodology question
and return a structured response. Do nothing beyond this step.

Before invoking Marcus, read these decision files in full — they are standing orders:
{{standing_order_files}}

Also read these for context:
{{context_files}}

TASK:
  ID: {{task_id}}
  Title: {{task_title}}
  Context: {{task_context}}
  Acceptance criteria:
{{acceptance_criteria}}

METHODOLOGY QUESTION FOR MARCUS:
  {{methodology_question}}

Now invoke /algo-trading-veteran. Give Marcus:
  - The task context above
  - The standing order decisions you read (cite them by title)
  - The specific methodology question

Ask him to:
  1. Answer the methodology question with a clear call
  2. Note which standing orders he is applying (cite decision titles)
  3. Mark any new methodology calls inline as **Decision (topic) — algorithm: status**

After Marcus responds, return ONLY this JSON (no other text before or after the braces):
{
  "step": "marcus_precheck",
  "verdict": {
    "summary": "2-4 sentence summary of Marcus's answer and the call he made",
    "go_iterate_kill": "n/a"
  },
  "decision_marks": ["**Decision (...) — algorithm/...: ...**"],
  "flag": null
}

IMPORTANT: Return the JSON immediately after Marcus responds. Do not add commentary,
do not say "here is the JSON", do not suggest re-running. One response, one JSON block.
```
