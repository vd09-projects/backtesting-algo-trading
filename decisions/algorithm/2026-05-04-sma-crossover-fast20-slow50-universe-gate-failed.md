# SMA crossover (fast=20, slow=50) fails universe gate — zero sufficient instruments

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-04       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | sma-crossover, universe-gate, insufficient-trades, zero-sufficient, fast-20, slow-50, TASK-0068, kill |

## Context

After sma-crossover was killed at the instrument-count gate in TASK-0053 (fast=10, slow=20 produced 4 of 12 WF passes), a pre-committed revisit trigger was set: if the strategy failed the instrument-count gate with fewer than 6 qualifying instruments, re-test with structurally motivated slower parameters before treating it as a definitive kill. The rationale was that fast=10/slow=20 on daily bars may introduce too much noise, causing the strategy to trade on short-term fluctuations rather than regime changes.

The parameter change to fast=20/slow=50 was committed before WF results were analyzed (see TASK-0067 decision), making the re-test methodologically clean.

TASK-0068 ran the universe sweep on all 15 nifty50-large-cap instruments with the new parameters over the 2018-01-01 to 2024-01-01 window with commission=zerodha_full.

## Decision

SMA crossover at fast=20/slow=50 is **killed at the universe gate**. All 15 instruments produced InsufficientData=true (trade_count 12-20, well below the MinTradesForMetrics threshold of 30).

The `ApplyUniverseGate` function skips all InsufficientData=true instruments. With 0 sufficient instruments, DSRAverageSharpe is undefined and GatePass cannot be true. The gate fails before any DSR or PassFraction calculation is possible.

The strategy is definitively killed. No walk-forward is run.

## Consequences

- fast=20/slow=50 on daily bars generates approximately 2-3 trades per year per instrument over the 6-year window. This is too sparse for reliable Sharpe estimation, DSR correction, or walk-forward fold validation.
- The parameter space for daily-bar SMA crossover that produces ≥30 trades appears to be bounded by fast=10/slow=20 (or similar). Slower parameters below the 30-trade floor are statistically infeasible for this universe and window.
- SMA crossover is now killed at two parameter settings. The fast=10/slow=20 kill was at the instrument-count gate (too few WF passes); the fast=20/slow=50 kill is at the universe gate (too few trades). These are different failure modes.
- Both kills are clean: the first was a regime-stability failure, the second is a statistical infeasibility failure.

## Related decisions

- [SMA crossover fails walk-forward instrument-count gate — 4 of 12 eligible instruments pass](2026-05-04-sma-crossover-walk-forward-instrument-count-gate.md) — prior kill at WF gate that triggered this revisit
- [SMA re-test after walk-forward kill: pre-committed revisit trigger is methodologically clean](2026-05-04-sma-retune-methodology-pre-committed-trigger.md) — the methodology decision that authorized this re-test
- [Momentum fails universe gate — zero sufficient instruments](2026-05-03-momentum-universe-gate-failed.md) — same failure mode (zero sufficient instruments due to parameter-induced sparsity)

## Revisit trigger

Do not revisit sma-crossover on daily bars unless a parameter set can be identified a priori (before running) that produces ≥30 trades/instrument across the nifty50-large-cap universe. At daily timeframe with a 6-year window, the valid parameter space is approximately fast ≤ 15 / slow ≤ 30 — but this range was already killed at the instrument-count gate.

Consider weekly bars or intraday if SMA crossover is to be revisited.
