# Donchian breakout fails universe gate — DSR-corrected average Sharpe negative

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | rejected         |
| Category | algorithm        |
| Tags     | donchian-breakout, universe-gate, DSR, insufficient-trades, TASK-0052, kill |

## Context

TASK-0052 ran `cmd/universe-sweep` for donchian-breakout (period=10, the plateau-midpoint from TASK-0051) across all 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, with `--commission zerodha_full`. The universe gate requires: DSR-corrected average Sharpe > 0 AND >= 40% of sufficient instruments show positive Sharpe.

A sufficient instrument is one where `insufficient_data=false` (trade_count >= 30).

Prior signal frequency audit (2026-05-01) had flagged this as borderline: donchian-period=10 was a single-value plateau with 7 of 15 instruments excluded at the >=30-trade threshold on NSE:RELIANCE.

## Options considered

N/A — gate application, not a design decision. The gate criteria are fixed (2026-04-25).

## Decision

**Donchian-breakout is killed.** Gate applied as specified.

**Numeric evidence:**

- Sufficient instruments (trade_count >= 30): 7 of 15
  - NSE:INFY (30 trades, Sharpe=0.672)
  - NSE:SBIN (31 trades, Sharpe=0.597)
  - NSE:RELIANCE (32 trades, Sharpe=0.436)
  - NSE:KOTAKBANK (30 trades, Sharpe=0.242)
  - NSE:ITC (33 trades, Sharpe=-0.022)
  - NSE:AXISBANK (36 trades, Sharpe=-0.172)
  - NSE:MARUTI (34 trades, Sharpe=-0.370)
- Excluded (insufficient_data=true): NSE:BAJFINANCE(24), NSE:LT(25), NSE:TITAN(29), NSE:WIPRO(27), NSE:TCS(27), NSE:HINDUNILVR(25), NSE:ICICIBANK(29), NSE:HDFCBANK(29)
- PositiveSharpe (raw>0): 4/7 = 57.1% — would pass the >= 40% condition
- DSR-corrected average Sharpe (nTrials=15): **-0.1194** — **fails DSRAvg > 0 condition**

The three negative-Sharpe sufficient instruments (ITC, AXISBANK, MARUTI) drag the DSR-corrected average below zero. The strategy generates meaningful signal frequency only on the portion of the universe that shows no edge; the instruments with apparent edge are excluded because they have insufficient trades.

Gates passed before kill: none (universe gate is the first gate in the TASK-0052 pipeline step).

## Consequences

- Donchian-breakout does not advance to walk-forward (TASK-0053).
- The pattern here is structural: period=10 is too short for most Nifty50 large-caps to generate 30+ trades in a 6-year window. A longer period would increase trade frequency but would require re-running the sweep to find a valid plateau — and longer periods on a breakout strategy may have different edge characteristics.
- TASK-0053 (walk-forward) proceeds without donchian-breakout in the candidate set.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [Plateau procedure for trade-count-constrained strategies](./2026-04-29-plateau-procedure-trade-count-constrained.md) — explains why period=10 was selected despite low trade count
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — flagged donchian as borderline before this run

## Revisit trigger

If the project revisits 15-instrument universe sweep with a different donchian period (e.g., period=20-30 where more instruments reach 30 trades), this kill decision should be reconsidered. The edge thesis (breakout persistence) is not disproved — the specific period=10 parameter cannot produce sufficient sample size across this universe.
