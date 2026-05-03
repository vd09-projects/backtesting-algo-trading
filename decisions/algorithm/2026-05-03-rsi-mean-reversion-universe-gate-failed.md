# RSI mean-reversion fails universe gate — zero sufficient instruments

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | rsi-mean-reversion, universe-gate, insufficient-trades, zero-sufficient, TASK-0052, kill |

## Context

TASK-0052 ran `cmd/universe-sweep` for rsi-mean-reversion (period=14, oversold=30, overbought=70 — strategy defaults used per the fallback decision since no valid plateau was found on RELIANCE) across all 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, with `--commission zerodha_full`.

The universe gate requires: DSR-corrected average Sharpe > 0 AND >= 40% of sufficient instruments show positive Sharpe. A sufficient instrument requires trade_count >= 30.

Prior signal frequency audit (2026-05-01) had flagged this as a strong candidate to fail: trades_at_midpoint=7, and all 15 instruments were excluded at the >=30-trade threshold on NSE:RELIANCE.

## Options considered

N/A — gate application.

## Decision

**RSI mean-reversion is killed.** Gate applied as specified.

**Numeric evidence:**

All 15 instruments have `insufficient_data=true`. Maximum trade count observed: 9 (NSE:MARUTI). No instrument reaches the 30-trade floor required to be a sufficient instrument.

- Sufficient instruments: 0 of 15
- DSRAvg: undefined (no sufficient instruments)
- PassFraction: undefined (no sufficient instruments)

With zero sufficient instruments, neither condition of the universe gate can be evaluated — the strategy produces no statistically credible Sharpe estimates on any instrument in this universe at these parameters.

Gates passed before kill: none (universe gate is the first gate; strategy fails at the entry condition).

## Consequences

- RSI mean-reversion does not advance to walk-forward (TASK-0053).
- The RSI (14, 30/70) parameterization is a low-frequency mean-reversion configuration on daily bars. For Nifty50 large-caps in 2018–2024, the RSI rarely touches oversold/overbought thresholds — the 2020 crash briefly triggered oversold conditions but the recovery was fast, and the grinding 2021–2023 uptrend generated almost no mean-reversion signals.
- TASK-0053 proceeds without RSI mean-reversion.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — flagged RSI as high risk before this run (3–9 trades/instrument at defaults)
- [Fallback to strategy defaults for universe sweep when no valid plateau exists](./2026-04-29-fallback-to-defaults-no-valid-plateau.md) — default parameters used here because no valid plateau was found
- [RSI mean-reversion fails proliferation gate — NSE:RELIANCE 2018–2025](./2026-04-16-rsi-mean-reversion-proliferation-gate-failed.md) — earlier single-instrument evidence (superseded gate, consistent picture)

## Revisit trigger

RSI mean-reversion could be revisited with looser band parameters (e.g., 40/60 or 35/65 thresholds) that generate more trades. This would be a different strategy parameterization — not a continuation of the current evaluation pipeline, which used the plateau-midpoint or fallback defaults as specified.
