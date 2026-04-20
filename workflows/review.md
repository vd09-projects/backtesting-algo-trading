# Review / Planning Session

User has results and wants expert review. Runs autonomously through Marcus, Priya (if needed),
and task creation. Ends with a Session Summary. (~20% of sessions.)

## Trigger

User has run outputs (in `runs/`), backtest results, performance metrics, or test results.
Usually: "I ran both strategies, here are the results," "what do you make of this output,"
"review these numbers."

---

## Execution

### Step 1 — Context check

Read `tasks/BACKLOG.md` — check what is in progress and what is pending. Identify any tasks
the results might resolve, unblock, or inform. Note them.

Read `decisions/algorithm/` index — surface any prior methodology decisions relevant to the
strategies or metrics being reviewed (e.g., proliferation gate threshold, annualization factors,
kill-switch line). These are the ground truth Marcus will compare against.

Log:
```
[AUTO] Step 1 — Context: N in-progress tasks, M relevant prior decisions loaded.
```

### Step 2 — Methodology review

Auto-invoke `/algo-trading-veteran`. Pass: the results, any run files or output the user
shared, in-progress task context, relevant prior decisions.

Marcus reviews from the methodology side: are the numbers honest, does the edge thesis hold,
are the metrics computed correctly, what does the data say about strategy viability. He reads
the results against prior decisions — if a prior proliferation gate decision exists, he applies
it rather than re-evaluating the threshold.

Marcus delivers a verdict for each strategy reviewed:
- **go** → edge thesis holds, size it and proceed
- **iterate** → something needs changing before a verdict is possible
- **kill** → edge not present; record the rejection

He marks new `algorithm` decisions inline for any specific calls (sizing adjustments,
updated kill-switch line, feature verdict on a new indicator).

Log:
```
[AUTO] Step 2 — Marcus review: complete. Verdicts: <strategy → verdict, ...>
[DECISION] Marcus [algorithm]: <any new calls>
```

### Step 3 — Dev review (conditional)

Only run this step if Marcus's review surfaces implementation questions:
- "How is this metric computed in the code exactly?"
- "The equity curve shape suggests a data pipeline issue"
- "I want to verify the fill model is doing what I think"
- "These numbers look like there's a lookahead — Priya should check"

Auto-invoke `/algo-trading-lead-dev`. Pass Marcus's review and the specific implementation
question. Priya reads the relevant code and answers: is the metric correct, is there a code
issue, what is the engineering cost of Marcus's recommendation.

If Priya finds a code issue, it becomes a bug task (Step 4). If she finds a methodology
tension ("the code does X but Marcus wants Y — there's a constraint"), briefly return to Marcus
with the clarification, then proceed.

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
