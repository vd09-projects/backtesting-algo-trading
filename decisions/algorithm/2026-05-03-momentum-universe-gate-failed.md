# Momentum fails universe gate — zero sufficient instruments

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | momentum, universe-gate, insufficient-trades, zero-sufficient, TASK-0052, kill |

## Context

TASK-0052 ran `cmd/universe-sweep` for momentum (lookback=231, threshold=10.0 — strategy defaults used per fallback decision; no valid plateau was found on RELIANCE) across all 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, with `--commission zerodha_full`.

The universe gate requires: DSR-corrected average Sharpe > 0 AND >= 40% of sufficient instruments show positive Sharpe. Sufficient = trade_count >= 30.

Prior signal frequency audit (2026-05-01) flagged momentum as the highest-risk candidate: trades_at_midpoint=4, all 15 instruments excluded at >=30-trade threshold. A 231-day lookback with a 10% ROC threshold on daily bars produces near-zero signal frequency on this universe.

## Options considered

N/A — gate application.

## Decision

**Momentum is killed.** Gate applied as specified.

**Numeric evidence:**

All 15 instruments have `insufficient_data=true`. Maximum trade count observed: 4 (NSE:RELIANCE and NSE:KOTAKBANK). This is the lowest signal frequency of all six strategies evaluated — the 231-day lookback window means the strategy barely completes one full turnover in the 6-year evaluation period.

- Sufficient instruments: 0 of 15
- DSRAvg: undefined (no sufficient instruments)
- PassFraction: undefined (no sufficient instruments)

Gates passed before kill: none.

## Consequences

- Momentum does not advance to walk-forward (TASK-0053).
- The lookback=231 default (252 - 21, the skip-last-month convention from Jegadeesh-Titman style momentum) is designed for annual rebalancing cycles. Applied to daily-bar entry/exit on individual NSE large-caps over 6 years, it produces 1-4 trades per instrument — far below the statistical minimum for any meaningful Sharpe estimate.
- TASK-0053 proceeds without momentum.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — flagged momentum as highest-risk candidate (1–4 trades/instrument at defaults)
- [Fallback to strategy defaults for universe sweep when no valid plateau exists](./2026-04-29-fallback-to-defaults-no-valid-plateau.md) — default parameters used here

## Revisit trigger

Momentum as a strategy thesis (buying relative strength leaders) is valid but requires a different implementation for this universe: either (a) a cross-sectional ranking approach across all 15 instruments simultaneously, or (b) a much shorter lookback (20–60 days) for single-instrument time-series momentum. Neither is this strategy as currently implemented. A future task could revisit with a shorter lookback sweep.
