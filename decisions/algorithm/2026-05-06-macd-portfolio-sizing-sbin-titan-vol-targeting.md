# MACD crossover portfolio — capital allocation and kill-switch thresholds (SBIN + TITAN)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-06       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | portfolio-sizing, vol-targeting, kill-switch, SBIN, TITAN, MACD-crossover, TASK-0055 |

## Context

Marcus's GO verdict on TASK-0055 (cross-strategy correlation and portfolio construction) established the expected final portfolio as NSE:SBIN + NSE:TITAN (subject to actual pairwise Pearson r validation in TASK-0085). This decision records the capital allocation rule and kill-switch thresholds that apply to this 2-instrument portfolio.

The sizing methodology follows the vol-targeting algorithm decision (2026-04-13) and kill-switch derivation methodology (2026-04-21). All thresholds are pre-committed before live deployment and must not be changed after observing live results.

## Capital allocation

**Total capital:** ₹3,00,000 (three lakh rupees)

**Base allocation per instrument:** ₹1,50,000 notional (equal split, 2-instrument portfolio)

**If regime gate (TASK-0086) flags RegimeConcentrated=true for either instrument:** halve that instrument's base allocation to ₹75,000. The freed capital remains in cash — no reallocation to the other instrument.

**If TASK-0085 admits BAJFINANCE as a third instrument (full-period r < 0.7 AND both stress-period r < 0.6):** rebase to 3-instrument allocation at ₹1,00,000 each. Revisit kill-switch thresholds for BAJFINANCE using the same methodology below.

## Sizing rule — SizingVolatilityTarget

Per-instrument position sizing follows `SizingVolatilityTarget`:

```
fraction = volTarget / (instrumentVol × sqrt(252))
fraction = min(fraction, 1.0)  // no leverage
```

Parameters:
- `volTarget`: 0.10 (10% annualized portfolio volatility target)
- `instrumentVol`: 20-bar rolling standard deviation of log-returns of the instrument's price series
- `sqrt(252)`: annualisation factor for daily bars
- No leverage: fraction is capped at 1.0

The vol-targeting formula scales position size inversely with realized volatility. In high-volatility regimes (e.g., 2020 crash), position sizes shrink automatically. This is the primary risk management mechanism alongside the kill-switch conditions below.

## Kill-switch thresholds — NSE:SBIN

Bootstrap source: `runs/bootstrap-macd-2026-05-05/SBIN.json`
In-sample period: 2018-01-01 to 2024-01-01

| Threshold | Formula | Value |
|---|---|---|
| Rolling per-trade Sharpe | SharpeP5 from bootstrap | **0.0719** |
| Max drawdown threshold | 1.5 × MaxDrawdown (2.7315%) | **4.10%** |
| Max DD duration threshold | 2 × MaxDrawdownDuration (224 days) | **448 days** |

MaxDrawdownDuration from SBIN.json: 19,353,600,000,000,000 ns = 19,353,600 s = 224.0 days (exact).

## Kill-switch thresholds — NSE:TITAN

Bootstrap source: `runs/bootstrap-macd-2026-05-05/TITAN.json`
In-sample period: 2018-01-01 to 2024-01-01

| Threshold | Formula | Value |
|---|---|---|
| Rolling per-trade Sharpe | SharpeP5 from bootstrap | **0.0854** |
| Max drawdown threshold | 1.5 × MaxDrawdown (3.1454%) | **4.72%** |
| Max DD duration threshold | 2 × MaxDrawdownDuration (694 days) | **1,388 days** |

MaxDrawdownDuration from TITAN.json: 59,961,600,000,000,000 ns = 59,961,600 s = 693.5 days → rounded to **694 days**, threshold = **1,388 days**.

Note: TITAN's max DD duration threshold (1,388 days = ~3.8 years) is unusually long relative to the 6-year backtest window. This reflects a single deep multi-year drawdown in the in-sample period. Apply vigilance if TITAN enters drawdown — the 1,388-day threshold is technically correct per the methodology but should prompt qualitative review if duration exceeds 2 years.

## Halt procedure

When any kill-switch threshold is breached on any instrument:

1. Halt new entries on that instrument only. Do not exit existing open position mid-trade.
2. Close the open position at the next session close (or at the end of the current bar if end-of-day is already reached).
3. Record the halt in the trade log with reason: "kill-switch — [Sharpe|MaxDD|MaxDDDuration] threshold breached".
4. Do NOT retune parameters during drawdown. Evaluate from scratch with updated data if considering re-entry.
5. The other instrument in the portfolio continues operating unless its own kill-switch triggers independently.

## Pre-commitment notice

These thresholds are recorded before live deployment and are binding. Changing any threshold after observing live results constitutes overfitting to live history and invalidates the methodology. Any revision requires a new Marcus evaluation session with updated bootstrap evidence.

## Related decisions

- [Kill-switch derivation methodology (2026-04-21)](./2026-04-21-kill-switch-derivation-methodology.md) — the methodology these thresholds are derived from
- [Vol-targeting algorithm choices (2026-04-13)](./2026-04-13-vol-targeting-algorithm-choices.md) — SizingVolatilityTarget specification
- [Banking cluster decision (2026-05-06)](./2026-05-06-banking-cluster-sbin-bajfinance-icicibank.md) — instrument selection rationale; BAJFINANCE admission condition
- [Bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — source data for SharpeP5 values
- [MACD bootstrap gate results (2026-05-05)](./2026-05-05-macd-bootstrap-gate-results.md) — in-sample JSON files used for MaxDrawdown and MaxDrawdownDuration

## Revisit trigger

Revisit if:
- TASK-0085 correlation gate admits BAJFINANCE → recompute 3-way allocation
- TASK-0086 regime gate flags RegimeConcentrated=true for either instrument → halve that instrument's allocation
- Future bootstrap run with extended data window (post-2024) materially shifts SharpeP5 → update thresholds via new Marcus evaluation
