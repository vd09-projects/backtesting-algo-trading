# Review / Planning Session

User has results (backtest runs, metrics, test output) and wants expert review to plan next
steps. Often produces a batch of new tasks. (~20% of sessions.)

## Trigger

User has run outputs (in `runs/`), backtest results, performance metrics, or test results
and wants them reviewed. Usually sounds like: "I ran both strategies, here are the results,"
"can you review these numbers," "what do you make of this output."

## Flow

### Step 1 — Check what's pending

Invoke `/task-manager` — ask what's in progress, what's pending, any blocked items that might
relate to the results being reviewed. This grounds the review in the project's current state.

**Carry forward:** in-progress task IDs, any pending items the results might resolve or inform.

### Step 2 — Methodology review

Invoke `/algo-trading-veteran` — share the results. Marcus reviews from the methodology side:
are the numbers honest, does the edge thesis hold, are the metrics computed correctly, what
does the data say about the strategy's viability.

Marcus will read the run files, check prior `algorithm`-category decisions for consistency,
and give his assessment. He may mark new algorithm decisions inline (sizing adjustments,
kill-switch refinements, feature verdicts).

He doesn't hit a terminal state the same way Priya does — he finishes when his assessment
is complete. Look for his verdict (go / iterate / kill) or his recommendations.

**Carry forward:** Marcus's key findings, any decisions he marked, specific questions he raised
about the implementation.

### Step 3 — Dev perspective (when valuable)

Not always needed. Invoke `/algo-trading-lead-dev` only when:
- Marcus's review raises implementation questions (data pipeline shape, code structure concerns,
  performance of the engine, how a metric is computed)
- The results suggest a code issue (unexpected numbers that might be a bug vs a real finding)
- The next step involves implementation planning that benefits from Priya's perspective
- The user explicitly asks for Priya's take

Share Marcus's review with Priya. She adds the dev angle: is the metric computation correct
in the code, are there implementation constraints Marcus should know about, what's the
engineering cost of his recommendations.

Priya may mark architecture or convention decisions based on what she sees. She may also
flag things back to Marcus — "the metric is computed this way because of X constraint,
Marcus should know this before making his sizing call." If this happens, briefly return to
Marcus with Priya's clarification, then move on.

### Step 4 — Create tasks from recommendations

Invoke `/task-manager` — ask it to create tasks from Marcus's and Priya's recommendations.
The user should review the task list before confirming.

Common task patterns from review sessions:
- "Implement Marcus's sizing adjustment" (source: decision)
- "Fix metric computation Priya flagged" (source: discovery)
- "Run the backtest again with updated parameters" (source: session)
- "Investigate anomaly in regime 3 results" (source: discovery)

### Step 5 — Session end

Go to `session-end.md`.
