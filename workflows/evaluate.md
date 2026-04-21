# Strategy Evaluation

User has a new strategy idea. Marcus evaluates it autonomously. Verdict auto-routes to the
next action. Pure methodology — no code yet. (~10% of sessions.)

## Trigger

User describes a strategy idea, asks "is this worth building," shares an edge thesis, or asks
about a new market/instrument.

---

## Execution

### Step 1 — Prior decision check (sub-agent)

Read `workflows/agents/decision-lookup.md`. Fill slots with the user's strategy idea as
task context, and tags inferred from the idea (instrument, edge category, thesis type).

Spawn sub-agent. Parse JSON. If `standing_order_files` contains a prior rejection of this
exact thesis, log it:
```
[AUTO] Step 1 — Prior decisions: found rejection of <similar idea> on <date>.
       Proceeding with Marcus's re-evaluation in case context has changed.
```

Always proceed to Marcus even if a prior rejection exists.

Log:
```
[AUTO] Step 1 — Decision lookup: N relevant decisions found.
```

### Step 2 — Marcus's interrogation (sub-agent)

Read `workflows/agents/marcus-precheck.md`. Fill slots:
- `{{task_id}}` — "evaluate" (no task ID for evaluate sessions)
- `{{task_title}}` — the strategy idea in one line
- `{{task_context}}` — the user's full description of the idea
- `{{acceptance_criteria}}` — "go/iterate/kill verdict + test plan + sizing + kill-switch"
- `{{standing_order_files}}` and `{{context_files}}` — from Step 1 verdict
- `{{methodology_question}}` — "Evaluate this edge thesis: <one-line summary>"

Also ask Marcus for `go_iterate_kill` (set this field in the verdict, not "n/a").

**If information is missing that the user must supply** (which instrument, what capital, what
data source): the sub-agent will return a `flag`. If that flag describes a requirements gap
only the user can fill — this is the one Hard STOP in evaluate.md. Ask those specific questions
and wait. Without these answers Marcus cannot give an honest assessment.

Spawn sub-agent. Parse JSON. Set `verdict.marcus.go_iterate_kill`.

Log:
```
[AUTO] Step 2 — Marcus evaluation: complete. Verdict: <go | iterate | kill>.
[DECISION] Marcus [algorithm]: <each specific call he made>
```

### Step 3 — Verdict routing

**Verdict: go**

Marcus has signed off with a sizing recommendation and kill-switch line. Auto-create
implementation tasks via `/task-manager`:

- Task: "Implement [strategy name] backtest" — high priority, source: decision
- Task: "Run walk-forward validation — [strategy name]" — high priority, source: decision
- Task: "Document kill-switch line in decisions journal" — medium priority, source: decision

Log each task created. Then go to session-end.md.

**Verdict: iterate**

The idea needs refinement — different instrument, different timeframe, different edge framing.
Stay with Marcus. The user refines the idea and Marcus re-evaluates. This loop runs until
Marcus gives "go" or "kill."

If the iterate loop surfaces a feasibility question only Priya can answer ("is it feasible to
build 1-minute-bar support?" / "how hard is funding-rate handling?"): auto-invoke
`/algo-trading-lead-dev` for a feasibility check, return to Marcus with the answer, resume.

Log each iteration:
```
[AUTO] Step 3 — Iterate round N: user refined thesis. Marcus re-evaluating.
```

**Verdict: kill**

The idea did not pass. Marcus explains why. Do not create implementation tasks.

Auto-harvest the rejection as a decision record: `/decision-journal` in record mode, category
`algorithm`, status `rejected`. The file should capture: the idea, why it was rejected, and
what would need to change for it to be worth re-evaluating.

Log:
```
[AUTO] Step 3 — Kill verdict. Rejection harvested to decisions/algorithm/.
```

Then go to session-end.md.

### Step 4 — Session end

Go to `session-end.md`.

For "go" sessions: the decisions and implementation tasks are in the Summary.
For "kill" sessions: the Summary shows the rejection record — this is the primary output.
