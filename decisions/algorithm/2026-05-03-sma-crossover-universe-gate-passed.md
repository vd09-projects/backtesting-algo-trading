# SMA crossover passes universe gate — advances to walk-forward

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | sma-crossover, universe-gate, DSR, walk-forward, TASK-0052, survivor |

## Context

TASK-0052 universe sweep for sma-crossover (fast=10, slow=20 — plateau-midpoint from TASK-0051) across 15 Nifty50 large-cap instruments, 2018-01-01 to 2024-01-01, `--commission zerodha_full`. Applied the DSR-corrected universe gate (2026-04-25 decision).

## Decision

**SMA crossover passes the universe gate.** Advances to walk-forward (TASK-0053).

**Numeric evidence:**

- NSE:BAJFINANCE excluded (insufficient_data=true, trade_count=29 — one below the 30-trade floor)
- Sufficient instruments: 14/15
- PositiveSharpe (raw>0): 12/14 = 85.7%
- DSR-corrected average Sharpe (nTrials=15): **0.0969** — passes DSRAvg > 0
- PassFraction: 0.857 — passes >= 0.40

**Per-instrument results (sorted by Sharpe, sufficient instruments only):**

| Instrument | Raw Sharpe | Trades | DSR |
|---|---|---|---|
| NSE:LT | 0.9068 | 30 | 0.5780 |
| NSE:RELIANCE | 0.6179 | 37 | 0.3228 |
| NSE:TITAN | 0.6044 | 36 | 0.3051 |
| NSE:INFY | 0.6010 | 35 | 0.2974 |
| NSE:HINDUNILVR | 0.5558 | 30 | 0.2270 |
| NSE:ICICIBANK | 0.5463 | 38 | 0.2552 |
| NSE:SBIN | 0.5351 | 37 | 0.2400 |
| NSE:WIPRO | 0.4952 | 39 | 0.2079 |
| NSE:TCS | 0.3806 | 41 | 0.1006 |
| NSE:AXISBANK | 0.1695 | 42 | -0.1070 |
| NSE:ITC | 0.1343 | 37 | -0.1608 |
| NSE:KOTAKBANK | 0.0670 | 41 | -0.2129 |
| NSE:HDFCBANK | -0.0499 | 42 | -0.3264 |
| NSE:MARUTI | -0.0826 | 39 | -0.3699 |

Excluded from sufficient set: NSE:BAJFINANCE (trade_count=29, insufficient_data=true).

**Eligible for walk-forward (positive raw Sharpe, sufficient instruments):**
NSE:LT, NSE:RELIANCE, NSE:TITAN, NSE:INFY, NSE:HINDUNILVR, NSE:ICICIBANK, NSE:SBIN, NSE:WIPRO, NSE:TCS, NSE:AXISBANK, NSE:ITC, NSE:KOTAKBANK (12 instruments)

NSE:HDFCBANK and NSE:MARUTI excluded (negative Sharpe). NSE:BAJFINANCE excluded (insufficient data).

**Regime gate:** Deferred (same reason as MACD — no per-period trade logs in CSV).

## Consequences

- SMA crossover advances to TASK-0053 (walk-forward) on 12 eligible instruments.
- DSRAvg of 0.0969 is a thin but positive margin. The strategy passes the gate but is the weaker of the two survivors — the multiple-testing correction takes a bigger bite relative to MACD because per-instrument raw Sharpes are lower. Walk-forward will be the more important validation for SMA than for MACD.
- NSE:BAJFINANCE was the top-Sharpe instrument (1.329) but had 29 trades — one below the floor. Its exclusion from the sufficient set does not affect the gate outcome (DSRAvg still positive, PassFrac still above 40%) but would have significantly boosted the DSRAvg if included.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here
- [SMA crossover fails proliferation gate — NSE:RELIANCE (superseded)](./2026-04-16-sma-crossover-proliferation-gate-failed.md) — earlier single-instrument failure under the old gate; that gate was superseded and this run supersedes it as the current evidence base
- [TASK-0051 signal frequency gate](./2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md) — SMA plateau at slow=20 produced trades_at_midpoint=37, no sensitivity concern

## Revisit trigger

If walk-forward (TASK-0053) fails on the majority of the 12 eligible instruments, the thin DSRAvg here (0.0969) warrants scrutiny — this margin is sensitive to a few negative-instrument results. If fewer than 6 instruments pass walk-forward, consider whether SMA crossover has sufficient universe breadth for portfolio inclusion.
