# Bollinger mean-reversion fails universe gate — zero sufficient instruments

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | bollinger-mean-reversion, universe-gate, insufficient-trades, zero-sufficient, TASK-0052, kill |

## Context

TASK-0052 ran `cmd/universe-sweep` for bollinger-mean-reversion (period=20, num-std-dev=2.0 — strategy defaults used per fallback decision since no valid plateau was found on RELIANCE) across all 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, with `--commission zerodha_full`.

The universe gate requires: DSR-corrected average Sharpe > 0 AND >= 40% of sufficient instruments show positive Sharpe. Sufficient = trade_count >= 30.

Prior signal frequency audit (2026-05-01) flagged this as a strong candidate to fail: trades_at_midpoint=14, all 15 instruments excluded at >=30-trade threshold on NSE:RELIANCE.

## Options considered

N/A — gate application.

## Decision

**Bollinger mean-reversion is killed.** Gate applied as specified.

**Numeric evidence:**

All 15 instruments have `insufficient_data=true`. Maximum trade count observed: 19 (NSE:HDFCBANK and NSE:TITAN). No instrument reaches the 30-trade floor.

- Sufficient instruments: 0 of 15
- DSRAvg: undefined (no sufficient instruments)
- PassFraction: undefined (no sufficient instruments)

The 2.0 standard-deviation bands are too wide to generate frequent entries on daily Nifty50 large-cap data. The strategy spends most of the backtest period in a hold state.

Gates passed before kill: none.

## Consequences

- Bollinger mean-reversion does not advance to walk-forward (TASK-0053).
- TASK-0053 proceeds without bollinger-mean-reversion.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — flagged Bollinger as high risk
- [Fallback to strategy defaults for universe sweep when no valid plateau exists](./2026-04-29-fallback-to-defaults-no-valid-plateau.md) — default parameters used here

## Revisit trigger

Bollinger mean-reversion could be revisited with tighter bands (e.g., 1.5 std dev) or a shorter period that generates more entries. That would constitute a different parameterization — a separate evaluation, not a continuation of this pipeline run.
