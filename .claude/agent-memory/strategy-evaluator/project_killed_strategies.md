---
name: Kill modes by strategy type
description: Which edge categories and failure modes have killed strategies in this project — pattern recognition for new evaluations
type: project
---

Strategies killed as of 2026-05-06, with specific failure modes:

| Strategy | Edge category | Failure gate | Specific failure |
|---|---|---|---|
| SMA crossover (10/20) | Trend-following | Walk-forward | 33% retention rate (4/12 instruments), gate required 60% |
| SMA crossover (20/50) | Trend-following | Universe gate | Zero sufficient instruments — 2-3 trades/year, far below 30-trade minimum |
| RSI mean-reversion | Mean-reversion | Signal frequency (proliferation gate) | Too few signals on NSE daily bars |
| Bollinger mean-reversion | Mean-reversion | Universe gate | DSR-corrected Sharpe gate failed |
| CCI mean-reversion | Mean-reversion | Universe gate | DSR-corrected Sharpe gate failed |
| Donchian breakout | Trend-following | Universe gate | DSR-corrected Sharpe gate failed |
| Momentum | Momentum | Universe gate | DSR-corrected Sharpe gate failed |

**Pattern:** Mean-reversion strategies (RSI, Bollinger, CCI) all fail on NSE large-cap daily bars. The universe gate is the most common kill point (5 of 7 kills). Signal frequency on daily bars is the first structural risk for any new strategy.

**Why this matters:** Any new strategy in the mean-reversion edge bucket faces strong prior evidence of failure on NSE large-cap daily bars. New trend-following strategies need to demonstrate > 35 trades/year per instrument before proceeding. Portfolio stage: only MACD crossover survives.

**How to apply:** In Marcus's Step 2 evaluation, flag immediately if: (a) edge bucket is mean-reversion — strong prior against this universe/timeframe combination, (b) estimated annual trade frequency < 35/year on daily bars — high-risk for signal audit, or (c) edge bucket is trend-following and the strategy has slow parameters — universe gate is likely binding.
