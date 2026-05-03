# MACD crossover passes universe gate — advances to walk-forward

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, universe-gate, DSR, walk-forward, TASK-0052, survivor |

## Context

TASK-0052 universe sweep for macd-crossover (fast=17, slow=26, signal=9 — plateau-midpoint from TASK-0051) across 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, `--commission zerodha_full`. Applied the DSR-corrected universe gate (2026-04-25 decision).

## Decision

**MACD crossover passes the universe gate.** Advances to walk-forward (TASK-0053).

**Numeric evidence:**

- Sufficient instruments: 15/15 (all instruments have trade_count >= 30, insufficient_data=false)
- PositiveSharpe (raw>0): 14/15 = 93.3%
- DSR-corrected average Sharpe (nTrials=15): **0.2715** — passes DSRAvg > 0
- PassFraction: 0.933 — passes >= 0.40

**Per-instrument results (sorted by Sharpe):**

| Instrument | Raw Sharpe | Trades | DSR |
|---|---|---|---|
| NSE:TCS | 1.0120 | 39 | 0.7247 |
| NSE:SBIN | 0.9915 | 39 | 0.7042 |
| NSE:BAJFINANCE | 0.8676 | 49 | 0.6120 |
| NSE:TITAN | 0.8605 | 48 | 0.6022 |
| NSE:LT | 0.6539 | 47 | 0.3929 |
| NSE:ICICIBANK | 0.6399 | 48 | 0.3816 |
| NSE:INFY | 0.6215 | 48 | 0.3632 |
| NSE:RELIANCE | 0.6005 | 45 | 0.3336 |
| NSE:HINDUNILVR | 0.5938 | 47 | 0.3328 |
| NSE:WIPRO | 0.5739 | 46 | 0.3100 |
| NSE:AXISBANK | 0.2931 | 49 | 0.0375 |
| NSE:ITC | 0.2533 | 48 | -0.0050 |
| NSE:KOTAKBANK | 0.1153 | 59 | -0.1172 |
| NSE:HDFCBANK | 0.0221 | 59 | -0.2104 |
| NSE:MARUTI | -0.1281 | 47 | -0.3892 |

**Eligible for walk-forward (positive raw Sharpe, sufficient instruments):**
NSE:TCS, NSE:SBIN, NSE:BAJFINANCE, NSE:TITAN, NSE:LT, NSE:ICICIBANK, NSE:INFY, NSE:RELIANCE, NSE:HINDUNILVR, NSE:WIPRO, NSE:AXISBANK, NSE:ITC, NSE:KOTAKBANK, NSE:HDFCBANK (14 instruments)

NSE:MARUTI excluded from walk-forward (negative Sharpe = no basis for further validation).

**Regime gate:** Deferred. The universe sweep CSV does not contain per-regime trade data. Will be applied when per-period trade logs are available. Per the regime gate decision (2026-04-27), failure would result in half-weight in portfolio, not a kill.

## Consequences

- MACD crossover advances to TASK-0053 (walk-forward validation) on 14 eligible instruments.
- This is the strongest universe sweep result of the six strategies. MACD fast=17/slow=26/signal=9 generates sufficient trade frequency across all 15 instruments and shows positive Sharpe on 14 of them.
- The DSR correction is meaningful here: raw average Sharpe would be ~0.49; DSR-corrected average is 0.2715, reflecting the 15-instrument multiple-testing penalty. The strategy needs to hold up in walk-forward to confirm this isn't in-sample regime concentration.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [TASK-0051 routing decision](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — MACD flagged as clear frontrunner before this run
- [Regime gate](./2026-04-27-regime-gate.md) — applies at portfolio construction stage; deferred from this CSV

## Revisit trigger

If walk-forward (TASK-0053) fails on the majority of the 14 eligible instruments, return here to review whether the universe sweep DSRAvg of 0.2715 was meaningful or whether regime concentration accounts for most of it.
