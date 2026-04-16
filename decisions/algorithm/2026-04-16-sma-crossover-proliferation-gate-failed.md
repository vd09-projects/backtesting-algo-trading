# SMA Crossover fails proliferation gate — NSE:RELIANCE 2018–2025

| Field    | Value     |
|----------|-----------|
| Date     | 2026-04-16 |
| Status   | accepted   |
| Category | algorithm  |
| Tags     | sma-crossover, proliferation-gate, sharpe, NSE:RELIANCE, TASK-0028, TASK-0019 |

## Context

TASK-0028 required running both baseline strategies against NSE:RELIANCE over 2018-01-01 to
2025-01-01 (daily bars, vol-targeting 10%, Zerodha commission defaults) and checking each result
against the proliferation gate. The gate threshold — Sharpe ≥ 0.5 vs buy-and-hold after costs —
was pre-committed before seeing any results (per `decisions/algorithm/2026-04-10-strategy-proliferation-gate.md`
and the instrument declaration rule).

The instrument (NSE:RELIANCE) was declared in TASK-0028's acceptance criteria before the first
run. Results are stored in `runs/sma-crossover-2018-2024.json`.

## Decision

SMA crossover (fast=10, slow=50) on NSE:RELIANCE over the 2018–2025 window **fails the
proliferation gate** with Sharpe 0.447 (threshold: 0.500).

**Consequence: TASK-0019 (MACD trend-following) is cancelled.** The gate rule is: if SMA
crossover doesn't pass, MACD is not built, because both sit in the trend-following thesis bucket.
A variation strategy doesn't rescue a thesis category that the baseline can't validate.

## Experiments

Full results from `runs/sma-crossover-2018-2024.json` (re-run post MaxDrawdown bug fix,
2026-04-16):

| Metric | Value |
|---|---|
| TotalPnL | ₹25,624 |
| TradeCount | 22 |
| WinCount / LossCount | 8 / 14 |
| WinRate | 36.4% |
| AvgWin / AvgLoss | ₹5,957 / ₹1,574 |
| ProfitFactor | 2.16 |
| SharpeRatio | **0.447** |
| SortinoRatio | 0.642 |
| CalmarRatio | 0.223 |
| MaxDrawdown | 16.4% |
| MaxDrawdownDuration | ~3.1 years |
| TailRatio | 1.08 |

Parameters: `--fast-period 10 --slow-period 50 --sizing-model vol-target --vol-target 0.10
--cash 100000 --from 2018-01-01 --to 2025-01-01 --timeframe daily`

Notable: ProfitFactor of 2.16 means winners are substantially larger than losers, but a 36%
win rate with only 22 trades over 7 years is a thin sample — most of the P&L comes from a small
number of large-win trades, which is characteristic of momentum strategies in trending regimes.
The 3.1-year MaxDrawdownDuration is long relative to the 7-year window.

## Consequences

- TASK-0019 (MACD) cancelled. Do not build unless the SMA crossover baseline is re-evaluated
  on a different instrument or regime, passes the gate, and a new decision supersedes this one.
- The trend-following thesis on RELIANCE over this period is inconclusive at this parameter
  set. This does not mean trend-following doesn't work; it means this specific baseline didn't
  clear the bar on this instrument in this window.
- Regime sub-window analysis (per TASK-0028's remaining criterion) may clarify whether the
  strategy has edge in specific periods that averages out over the full window.

## Related decisions

- [Strategy proliferation gate](../algorithm/2026-04-10-strategy-proliferation-gate.md) — the rule that triggered this verdict
- [Baseline backtest period 2018–2024](../algorithm/2026-04-15-baseline-backtest-period-2018-2024.md) — period commitment that governs this evaluation
- [Target instrument declared before first run](../algorithm/2026-04-15-instrument-declared-before-first-run.md) — instrument was NSE:RELIANCE, declared before any run

## Revisit trigger

If the strategy is retested on a different Nifty 50 constituent, a different parameter set
is formally evaluated, or the evaluation window is extended or changed — a new gate check
supersedes this decision. Do not rerun and cherry-pick; pre-commit the new parameters and
instrument before running.
