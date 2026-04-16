# RSI mean-reversion fails proliferation gate — NSE:RELIANCE 2018–2025

| Field    | Value     |
|----------|-----------|
| Date     | 2026-04-16 |
| Status   | accepted   |
| Category | algorithm  |
| Tags     | rsi-mean-reversion, proliferation-gate, sharpe, trade-count, NSE:RELIANCE, TASK-0028, TASK-0020 |

## Context

TASK-0028 required running both baseline strategies against NSE:RELIANCE over 2018-01-01 to
2025-01-01 (daily bars, vol-targeting 10%, Zerodha commission defaults) and checking each result
against the proliferation gate. The gate threshold — Sharpe ≥ 0.5 vs buy-and-hold after costs —
was pre-committed before seeing any results.

Results are stored in `runs/rsi-mean-rev-2018-2024.json`. The MaxDrawdown figure in the original
run was artifactual (0% due to a bug); corrected results were produced after the
`computeMaxDrawdownDepth` fix on 2026-04-16.

## Decision

RSI mean-reversion (period=14, oversold=30, overbought=70) on NSE:RELIANCE over the 2018–2025
window **fails the proliferation gate** on two counts:

1. **Sharpe 0.469 — below the 0.500 threshold.** Gate failed on the primary criterion.
2. **Only 7 trades in 7 years — sample is statistically meaningless.** A 100% win rate and
   positive P&L over 7 trades carries no evidential weight. The confidence interval on a 7-trade
   Sharpe spans roughly the entire real line.

**Consequence: TASK-0020 (Bollinger Band mean-reversion) is cancelled.** The gate rule is: if
RSI mean-reversion doesn't pass, the Bollinger Band variant is not built, because both sit in
the mean-reversion thesis bucket.

## Experiments

Full results from `runs/rsi-mean-rev-2018-2024.json` (re-run post MaxDrawdown bug fix,
2026-04-16):

| Metric | Value |
|---|---|
| TotalPnL | ₹32,947 |
| TradeCount | 7 |
| WinCount / LossCount | 7 / 0 |
| WinRate | 100% |
| AvgWin / AvgLoss | ₹4,707 / 0 |
| ProfitFactor | 0 (no losing trades — not meaningful) |
| SharpeRatio | **0.469** |
| SortinoRatio | 0.717 |
| CalmarRatio | 0.210 |
| MaxDrawdown | 17.4% |
| MaxDrawdownDuration | ~1.0 year |
| TailRatio | 1.04 |

Parameters: `--rsi-period 14 --oversold 30 --overbought 70 --sizing-model vol-target
--vol-target 0.10 --cash 100000 --from 2018-01-01 --to 2025-01-01 --timeframe daily`

Notable: 100% win rate with 7 trades is a red flag, not a good sign. RSI at 30/70 thresholds
barely fired on RELIANCE over 7 years. Possible causes: RELIANCE's volatility profile rarely
reaches fixed 30/70 RSI thresholds on daily bars; vol-targeting suppressing entries during
the high-volatility periods when RSI actually fired; or the strategy is genuinely unprofitable
and was lucky across 7 trades. Without a larger sample, indistinguishable. The signal frequency
issue warrants a diagnostic pass (see Revisit trigger) but does not change the gate outcome.

## Consequences

- TASK-0020 (Bollinger Bands) cancelled. Do not build unless the RSI baseline is re-evaluated
  on a different instrument, parameter set, or threshold regime, passes the gate, and a new
  decision supersedes this one.
- The mean-reversion thesis on RELIANCE over this period is not evaluable — 7 trades is not
  enough to distinguish edge from luck in either direction.
- The low signal frequency may indicate a parameter mismatch (fixed RSI thresholds vs
  RELIANCE's volatility) rather than absence of mean-reversion edge. Adaptive thresholds or
  a different instrument could produce a more testable sample size, but that is a separate
  research question.

## Related decisions

- [Strategy proliferation gate](../algorithm/2026-04-10-strategy-proliferation-gate.md) — the rule that triggered this verdict
- [Baseline backtest period 2018–2024](../algorithm/2026-04-15-baseline-backtest-period-2018-2024.md) — period commitment that governs this evaluation
- [Target instrument declared before first run](../algorithm/2026-04-15-instrument-declared-before-first-run.md) — instrument was NSE:RELIANCE, declared before any run

## Revisit trigger

Two distinct reasons to revisit: (a) if RSI is retested with adaptive thresholds (e.g.
Wilder-style percentile-based levels) or on a different instrument where the strategy fires
more frequently — pre-commit parameters and instrument before running; (b) if a diagnostic
shows the 7 low-fire-rate was due to a bug (e.g. vol-targeting zeroing out all signals during
high-vol periods) rather than threshold mismatch — fix the bug, rerun, and record a new verdict.
