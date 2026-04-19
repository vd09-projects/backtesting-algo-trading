# RSI signal frequency diagnostic — NSE:RELIANCE 2018–2025

| Field    | Value     |
|----------|-----------|
| Date     | 2026-04-19 |
| Status   | accepted   |
| Category | algorithm  |
| Tags     | rsi-mean-reversion, signal-frequency, trade-suppression, NSE:RELIANCE, TASK-0031 |

## Context

TASK-0031 required a diagnostic pass before any RSI parameter re-test: count how many bars
RSI(14) actually spent below 30 (oversold) and above 70 (overbought) on NSE:RELIANCE daily
data, 2018-01-01 to 2025-01-01. The prior backtest (TASK-0028) produced only 7 closed trades
across 7 years. Two competing hypotheses needed resolution:

1. **Threshold miscalibration** — fixed 30/70 thresholds rarely breached on RELIANCE; signal
   count itself is low.
2. **Entry suppression** — thresholds are breached adequately, but something in the engine
   converts most signal bars into skipped entries.

Diagnostic run using `cmd/rsi-diagnostic`, commit `7de179f` signal-frequency gate tooling.

## Diagnostic results

| Metric | Value |
|---|---|
| Total bars fetched | 1,736 |
| Valid RSI bars (after 14-bar lookback) | 1,722 |
| Oversold bars (RSI < 30) | 52 |
| Overbought bars (RSI > 70) | 147 |
| Total signal bars | 199 |
| Closed trades in backtest | 7 |

## Decision

**Hypothesis 1 is false. Hypothesis 2 is true but not due to a bug.**

The 30/70 thresholds are breached on 199 of 1,722 valid bars (~11.6%). The calibration is
not the problem. The explanation for 199 signal bars producing only 7 closed trades is
structural — it follows directly from the strategy design interacting with RELIANCE's
price behaviour:

1. **The strategy is long-only with a strict exit condition.** A trade opens only on RSI < 30
   (buy) and closes only on RSI > 70 (sell). For a closed trade to exist, the RSI must
   complete the full sequence: drop below 30 → recover → exceed 70 — in that order.
   Overbought bars with no open long are silently skipped; additional oversold bars while
   already long are also skipped (no pyramiding).

2. **RELIANCE's RSI profile does not often complete this cycle.** The 147 overbought bars
   cluster into extended trending stretches (Nifty large caps trend for months at a time).
   Most of these fire while no long is open — they are missed sell signals. The 52 oversold
   bars cluster around sharp dislocations (COVID March 2020, mid-2022 selloff). After a
   dislocation, RELIANCE typically mean-reverts to neutral RSI (40–60) rather than all the
   way to overbought. The position then stays open — the exit signal never fires — until
   the next multi-month trend pushes RSI above 70.

3. **Vol-targeting is not the suppressor.** The 20-bar realized vol during oversold events
   (high-dislocation periods) would increase the vol denominator, reducing position size,
   but would not zero it out. Vol-targeting explains smaller position sizes during stress
   events, not the absence of trades.

## Consequence

The RSI(14) 30/70 long-only strategy on RELIANCE is structurally unsuited to the instrument's
regime: it exits only on overbought, but the stock's post-dislocation recoveries rarely reach
overbought before the next signal fires. This is a **strategy design mismatch**, not a
calibration failure.

Two paths to more trades, each requiring a pre-committed parameter spec and a new decision
before running:

**Option A — Neutral exit:** Change the sell condition from RSI > 70 to RSI crossing above
50 (mean-reversion complete). This would produce more closed trades per oversold entry.
Hypothesis: the edge lives in the entry timing (buying dislocations), not in holding to
overbought.

**Option B — Cross-instrument test:** Run the same strategy on instruments with symmetric
RSI profiles (e.g. mid-caps with more two-sided vol, sector ETFs). RELIANCE's
large-cap-trending behaviour may be instrument-specific rather than a universal mean-reversion
failure.

Neither option is started until a decision pre-commits parameters and instrument.

## Related decisions

- [RSI mean-reversion fails proliferation gate](../algorithm/2026-04-16-rsi-mean-reversion-proliferation-gate-failed.md) — original gate failure that triggered this diagnostic
- [Strategy proliferation gate](../algorithm/2026-04-10-strategy-proliferation-gate.md) — gate rule governing re-test conditions

## Revisit trigger

If Option A (neutral exit) is pursued: pre-commit the exit RSI level (e.g. 50 or adaptive)
and minimum trade count target before any run. Record a new decision with results.
