---
name: Kill modes by strategy type
description: Which edge categories and failure modes have killed strategies in this project — pattern recognition for new evaluations
type: project
---

Strategies killed as of 2026-05-13, with specific failure modes:

| Strategy | Edge category | Failure gate | Specific failure |
|---|---|---|---|
| SMA crossover (10/20) | Trend-following | Walk-forward | 33% retention rate (4/12 instruments), gate required 60% |
| SMA crossover (20/50) | Trend-following | Universe gate | Zero sufficient instruments — 2-3 trades/year, far below 30-trade minimum |
| RSI mean-reversion | Mean-reversion | Signal frequency (proliferation gate) | Too few signals on NSE daily bars (7 trades/7 years on RELIANCE) |
| Bollinger mean-reversion | Mean-reversion | Universe gate | Zero sufficient instruments (max 19 trades/instrument) |
| CCI mean-reversion | Mean-reversion | Universe gate | DSR-corrected Sharpe negative (thin sample ~31-36 trades, nTrials penalty) |
| Donchian breakout | Trend-following | Universe gate | DSR-corrected Sharpe negative; only 7/15 sufficient instruments |
| Momentum | Momentum | Universe gate | Zero sufficient instruments (1-4 trades per instrument) |

**Surviving strategies:**
- MACD crossover (daily bars, trend-following): pipeline complete on both large-cap (SBIN+TITAN) and midcap (6 instruments)
- ORB (5-min, intraday breakout): GO verdict 2026-05-13, awaiting implementation
- Gap-and-Go (5-min, event-driven gap continuation): GO verdict 2026-05-13, awaiting implementation

**Patterns:**
- Mean-reversion strategies (RSI, Bollinger, CCI) all fail on NSE large-cap daily bars — NSE large-caps trend rather than mean-revert at daily resolution
- Universe gate is the most common kill point (5 of 7 kills). Signal frequency is the binding constraint on daily bars.
- The only trend-following strategy that survived (MACD) has the highest signal frequency of all tested (45-65 trades/instrument vs 30-35 for Donchian).
- Daily-bar breakout strategies (Donchian) fail where intraday breakout strategies (ORB) are expected to pass — different competition sets and timeframes.

**How to apply:** In Marcus's Step 2 evaluation:
- Flag immediately if edge bucket is mean-reversion on NSE daily bars — strong prior against this universe/timeframe
- For daily-bar strategies: estimated trade frequency must be > 35/year or signal audit risk is high
- For intraday strategies: different failure modes apply — walk-forward regime stability (2022 choppy) is now the primary identified risk, not signal frequency
- Gap-and-Go and ORB share a mutual correlation risk: if both pass bootstrap, stress-period r may exceed 0.6 (both are long-only 5-min morning-event CNC strategies)
