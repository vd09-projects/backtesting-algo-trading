---
id: 2026-05-17-5min-primary-timeframe-strategy-direction
title: "5-min bars are primary strategy development timeframe; daily bars secondary"
date: 2026-05-17
status: accepted
category: algorithm
tags: [timeframe, 5min, strategy-direction, daily-bars, 1min]
---

## Decision

New strategy development prioritizes 5-minute bars. Daily-bar strategies already in the live portfolio (MACD midcap — SBIN, TITAN, PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN) are not retired — they remain live and are monitored via their existing kill-switch thresholds. No new daily-bar strategies will be developed until the 5-min strategy portfolio is established.

1-min bars: contingent on empirical verification of Kite historical data depth (TASK-0129). If 1-min data extends past 60 days, evaluate feasibility. If limited to 60 days total, drop 1-min as a development target.

## Rationale

- 5-min bars provide ~75× more evaluation points per year than daily bars (18,900 vs 252), giving statistical tests significantly more power
- Existing daily strategies are indicator-crossover heavy (MACD, SMA, RSI, Bollinger, CCI, Momentum); all share the same edge bucket and most failed the universe gate
- NSE 5-min data is confirmed available for 5+ years (per TASK-0099/0101 — 15/15 large-cap, 48/48 midcap instruments cached)
- ORB and Gap-and-Go (both 5-min, Marcus GO verdicts) provide proof of concept for intraday edge on NSE

## Scope of shift

Existing daily strategies at 5-min: run all six existing strategy types (MACD, RSI mean-reversion, SMA crossover, Bollinger mean-reversion, CCI mean-reversion, Momentum) on 5-min data with recalibrated parameters and multiple variants (TASK-0127). Results decide which survive — no pre-judgement.

New strategy research (TASK-0126): canvas must target 5-min-feasible strategies. Daily-bar strategies not part of the research canvas unless they have a clear 5-min adaptation path.

## Infrastructure status

Generalized (works at any timeframe): engine, walk-forward, analytics (Sharpe annualization), provider/cache, TimedExit, PriceExit, session utilities.

Daily-specific (tickets filed): MinCurvePointsForMetrics constant (TASK-0130), NSERegimes2018_2024 missing 5-min coverage (TASK-0130).
