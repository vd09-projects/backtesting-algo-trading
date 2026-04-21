# Review / Planning Session

User has results and wants expert review. Runs autonomously through Marcus, Priya (if needed),
and task creation. Ends with a Session Summary. (~20% of sessions.)

## Trigger

User has run outputs (in `runs/`), backtest results, performance metrics, or test results.
Usually: "I ran both strategies, here are the results," "what do you make of this output,"
"review these numbers."

---

## Execution

### Step 1 — Context check (sub-agent)

Read `tasks/BACKLOG.md` directly (orchestrator reads this — it's a single small file).
Identify tasks the results might resolve, unblock, or inform.

Then spawn a decision-lookup sub-agent for the algorithm domain:
Read `workflows/agents/decision-lookup.md`. Fill slots:
- `{{task_id}}` — "review" (no task ID)
- `{{task_title}}` — "strategy results review"
- `{{task_context}}` — what the user shared (strategy names, timeframe, key metrics seen)

Spawn sub-agent. These file paths are the ground truth Marcus will compare results against.
Update SESSION STATE: `verdicts.decision_lookup`. Write `.session-state.json`.

Log:
```
[AUTO] Step 1 — Context: N in-progress tasks, M relevant prior decisions found.
```

### Step 2 — Methodology review (sub-agent)

Read `workflows/agents/marcus-precheck.md`. Fill slots:
- `{{task_id}}` — "review", `{{task_title}}` — "strategy results review"
- `{{task_context}}` — the results the user shared (numbers, equity curves, output)
- `{{acceptance_criteria}}` — "go/iterate/kill verdict per strategy reviewed"
- `{{standing_order_files}}` and `{{context_files}}` — from Step 1 verdict
- `{{methodology_question}}` — "Review these results. Are the numbers honest? Does the
  edge thesis hold? Apply the proliferation gate and any other standing order decisions."

Also set `go_iterate_kill` per strategy in the verdict.

Spawn sub-agent. Parse JSON. Update SESSION STATE: `verdicts.marcus`.
Append any `decision_marks` to `decision_marks_pending`. Write `.session-state.json`.

Log:
```
[AUTO] Step 2 — Marcus review: complete. Verdicts: <strategy → verdict, ...>
[DECISION] Marcus [algorithm]: <any new calls>
```

### Step 3 — Dev review (sub-agent, conditional)

Only run this step if Marcus's review surfaces implementation questions:
- "How is this metric computed in the code exactly?"
- "The equity curve shape suggests a data pipeline issue"
- "These numbers look like there's a lookahead — Priya should check"

Spawn a sub-agent with this prompt (fill in Marcus's specific question):
```
You are a step-agent. Invoke /algo-trading-lead-dev with a specific implementation question.

MARCUS'S QUESTION: {{marcus_implementation_question}}

Ask Priya to read the relevant code and answer: is the metric correct, is there a code issue,
what is the engineering cost of Marcus's recommendation.

Return ONLY:
{
  "step": "priya_dev_review",
  "verdict": {"findings": "...", "code_bug_found": true|false, "bug_description": "..."},
  "decision_marks": [],
  "flag": null
}
```

If Priya finds a code bug: note it for task creation in Step 4.
If methodology tension: include in SESSION STATE execution_log as [FLAGGED].

Log:
```
[AUTO] Step 3 — Priya dev review: <skipped (not needed) | complete — findings: ...>
[DECISION] Priya [architecture|convention]: <any decisions marked>
```

### Step 4 — Task creation

Auto-create tasks from Marcus's and Priya's outputs. Use this mapping:

| Source | Task priority | Source field |
|---|---|---|
| Marcus verdict: "go" → implement strategy | high | decision |
| Marcus verdict: "iterate" → specific change needed | high | decision |
| Marcus verdict: "kill" → record rejection in journal | n/a | auto-harvest only, no task |
| Priya found a code bug | high | discovery |
| Priya found a metric computation mismatch | high | discovery |
| Marcus flagged a sizing or risk concern | medium | decision |
| Marcus flagged a regime or stress test to run | medium | decision |
| Priya flagged technical debt or a structural concern | low–medium | discovery |

For each task, auto-invoke `/task-manager` to create it. Do not ask for confirmation — create
them all and list them in the Summary. The user can cancel or reprioritize from the Summary.

For "kill" verdicts: do not create implementation tasks. Do auto-harvest the rejection as a
`algorithm: rejected` decision in the journal so the reasoning is preserved.

Log:
```
[AUTO] Step 4 — Tasks created: N new tasks. (list them)
```

### Step 5 — Session end

Go to `session-end.md`.
