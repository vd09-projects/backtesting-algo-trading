# Strategy Evaluation

User has a new strategy idea. Marcus evaluates it autonomously. Verdict auto-routes to the
next action. Pure methodology — no code yet. (~10% of sessions.)

## Trigger

User describes a strategy idea, asks "is this worth building," shares an edge thesis, or asks
about a new market/instrument.

---

## Execution

### Step 1 — Prior decision check

Before invoking Marcus, check `decisions/algorithm/` for any prior evaluations of the same
idea, instrument, or edge category. If a prior decision rejected this exact thesis, surface it
immediately:

```
[AUTO] Step 1 — Prior decisions: found rejection of <similar idea> on <date>.
       Reason: <one-line from the decision summary>.
       Proceeding with Marcus's re-evaluation in case context has changed.
```

Always proceed to Marcus even if a prior rejection exists — context changes (new data, different
timeframe, refined thesis). The prior decision informs Marcus; it does not replace his evaluation.

### Step 2 — Marcus's interrogation

Auto-invoke `/algo-trading-veteran`. Pass: the user's idea and any prior decisions surfaced
in Step 1.

Marcus runs his 5-minute interrogation: edge thesis in one sentence, instrument/timeframe/
capital/capacity, stage, data. This is the most important part — do not rush it. Marcus asks
follow-up questions if the user's initial description is incomplete.

**If information is missing that the user must supply** (which instrument, what capital, what
data source): this is a Hard STOP — ask those specific questions. This is the one point in
`evaluate.md` where the workflow pauses. Without these answers Marcus cannot give an honest
assessment.

Marcus delivers:
- Edge thesis verdict: go / iterate / kill
- Test plan (which data, methodology, stress periods)
- Sizing recommendation (if the idea has legs)
- Kill-switch line (pre-committed halt condition)
- Inline `algorithm` decision marks for all specific calls

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
