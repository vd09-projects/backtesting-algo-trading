# SMA re-test after walk-forward kill: pre-committed revisit trigger is methodologically clean

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-04       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | sma-crossover, parameter-retune, walk-forward-kill, revisit-trigger, methodology-hygiene, TASK-0067 |

## Context

TASK-0053 killed sma-crossover (fast=10, slow=20) at the instrument-count walk-forward gate: 4/12 instruments passed (33%). The TASK-0053 kill decision included a pre-written revisit trigger: "If the parameter space is explored (e.g., slower SMA periods like fast=20/slow=50 that generate fewer but higher-quality signals), re-test with a fresh universe sweep before re-entering the walk-forward gate."

The question: is re-testing with fast=20/slow=50 after the walk-forward kill methodologically clean, given that RELIANCE, WIPRO, TCS, ITC, KOTAKBANK, SBIN, HINDUNILVR, and ICICIBANK all failed the WF gate at fast=10/slow=20?

## Decision

Re-testing sma-crossover at fast=20/slow=50 is methodologically clean under the following conditions:

1. **The revisit trigger was pre-committed.** The parameter set fast=20/slow=50 was named in the kill decision as a future candidate **before** anyone analyzed which specific instruments failed or why. The trigger was written at the moment of kill, not after studying the per-instrument WF breakdown.

2. **The motivation is structural, not result-derived.** The diagnosis (fast=10/slow=20 generates noise-driven signals on daily NSE bars) was the a priori expectation for a 10-day SMA on liquid large-cap daily data. The diagnosis did not require the WF failure data to reach. Moving to slower periods reduces signal frequency and targets trend structure rather than noise.

3. **The prohibited alternative is cherry-picking the WF results.** A contaminated re-test would be: "ICICIBANK failed because NegFoldFlag=True, SBIN failed because OOSISRatio=0.28, so let me tune specifically to fix those." That fits to the WF fold structure. The current re-test does not use the per-instrument failure patterns to select parameters.

**Conditions that must be satisfied for this re-test to remain clean:**
- The universe sweep at fast=20/slow=50 is treated as a clean run with no special expectations
- If the new variant fails the universe gate, it is killed there — not given a second chance
- The WF instrument-count threshold resets to however many instruments pass the new universe gate (not the prior 12)

## Consequences

sma-crossover (fast=20, slow=50) re-enters the evaluation pipeline at TASK-0068. This is a new variant, not a continuation of the killed variant. Its results must be recorded independently.

## Related decisions

- [SMA crossover fails walk-forward instrument-count gate (fast=10, slow=20)](./2026-05-04-sma-crossover-walk-forward-instrument-count-gate.md) — the kill this re-test responds to
- [SMA crossover passes universe gate (fast=10, slow=20)](./2026-05-03-sma-crossover-universe-gate-passed.md) — the universe gate result for the prior variant

## Revisit trigger

If sma-crossover at fast=20/slow=50 also fails the WF instrument-count gate with a similar failure pattern (8+ instruments, dominated by OverfitFlag+NegFoldFlag), the strategy should be killed permanently — the edge thesis for SMA crossover on daily Nifty50 large-caps is insufficiently robust regardless of parameter choice.
