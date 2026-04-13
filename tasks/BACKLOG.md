# Project Task Backlog

**Last updated:** 2026-04-13 | **Open tasks:** 10 | **Next up:** TASK-0015

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

_Nothing in progress._

---

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0015] Strategy — RSI mean-reversion

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-10
- **Source:** session
- **Context:** RSI mean-reversion (buy when RSI < 30, sell when RSI > 70) is the canonical mean-reversion baseline and the complement to SMA crossover's trend-following approach. Running both gives the first real signal on whether this market regime rewards momentum or mean reversion. Treat this as a dirty test — the goal is calibration and regime signal, not a live edge.
- **Acceptance criteria:**
  - [ ] `strategies/rsimeanrev/` package implementing the `Strategy` interface
  - [ ] Uses `github.com/markcheno/go-talib` RSI — no hand-rolled calculation
  - [ ] Configurable period (default 14), overbought threshold (default 70), oversold threshold (default 30)
  - [ ] Lookback returns `period + 1` (talib RSI needs one extra bar for the initial smoothing)
  - [ ] Tests: known OHLCV sequence with known RSI values → expected signals (table-driven)
  - [ ] Long-only: buy on oversold, sell/exit on overbought (no shorting in v1)
- **Notes:** After running, compare Sharpe to SMA crossover and buy-and-hold (TASK-0018). If RSI mean-rev is consistently better than SMA momentum, that's information about the regime. If both underperform buy-and-hold, that's also information — and the gate for TASK-0019/0020 should be applied.

---

### [TASK-0018] Analytics — buy-and-hold benchmark comparison

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** Every strategy result needs a baseline. A strategy that returns 12% annually sounds good until you learn the index returned 18%. Buy-and-hold on the same instrument over the same period is the minimum viable benchmark — if the strategy can't beat it after costs, it has no edge worth pursuing. This must be available when the first strategy results come in, not later.
- **Acceptance criteria:**
  - [ ] `analytics.BenchmarkReport` struct: `TotalReturn float64`, `AnnualizedReturn float64`, `SharpeRatio float64`, `MaxDrawdown float64`
  - [ ] `analytics.ComputeBenchmark(candles []model.Candle, initialCash float64) BenchmarkReport` — buys at first open, sells at last close, no costs
  - [ ] `output.printSummary` prints strategy vs benchmark side-by-side
  - [ ] Tests: known candle sequence → hand-verified benchmark metrics
- **Notes:** Moved up from Todo backlog. The first question after any strategy run is "did this beat buy-and-hold?" — not "what's the Sortino?" Keep it simple: one instrument, fully invested, no rebalancing.

---

### [TASK-0021] Engine — volatility-based position sizing

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** The current portfolio uses a fixed fraction of cash per trade (`PositionSizeFraction`). This ignores instrument volatility — a quiet stock gets the same dollar risk as a volatile one. Volatility targeting sizes each trade so the expected dollar risk is constant, which improves risk-adjusted returns on any strategy. Must be introduced before additional strategies are built so results are comparable across strategies and sizing doesn't need to be retrofitted.
- **Acceptance criteria:**
  - [ ] `model.SizingModel` typed enum: `SizingFixed` (current behavior), `SizingVolatilityTarget`
  - [ ] `engine.Config` gains `SizingModel` and `VolatilityTarget float64` (annualized, e.g. 0.10 = 10%)
  - [ ] When `SizingVolatilityTarget`: position notional (₹) = `(cash * volTarget) / (instrumentVol * sqrt(252))` where `instrumentVol` is the 20-bar realized std dev of daily returns (not annualized); position quantity = notional / fillPrice
  - [ ] Existing `SizingFixed` behavior unchanged — backward compatible
  - [ ] Tests: given known vol, expected position size is correct; vol=0 edge case handled
- **Notes:** Moved up from Todo backlog. Sizing dominates strategy selection — a mediocre strategy with proper sizing beats a great strategy with bad sizing. Build this before MACD and Bollinger Bands, not after.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

_No blocked tasks (TASK-0014 is in Up Next with its blocker noted there)._

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

### [TASK-0016] Analytics — profit factor, average win/loss, Sortino, and Calmar

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** Sharpe alone punishes upside volatility. Profit factor (gross profit / gross loss) and average win vs average loss together tell you whether the edge is in hit rate or in payoff ratio — critical for understanding how a strategy will behave in a drawdown. Sortino complements Sharpe by measuring only downside deviation.
- **Acceptance criteria:**
  - [ ] `Report` gains: `ProfitFactor float64`, `AvgWin float64`, `AvgLoss float64`, `SortinoRatio float64`, `CalmarRatio float64`
  - [ ] ProfitFactor = sum(winning trade P&L) / abs(sum(losing trade P&L)); returns 0 if no losing trades
  - [ ] AvgWin and AvgLoss are per-trade averages (not total)
  - [ ] Sortino uses downside deviation of per-bar returns (same equity curve as Sharpe, target return = 0)
  - [ ] Calmar = annualized return / max drawdown (%); returns 0 if max drawdown is zero
  - [ ] Tests: known trade sequences → hand-verified expected values for all five new fields
  - [ ] `output.printSummary` updated to include new fields
- **Notes:** A 35% win rate with PF 1.8 is a fine strategy. A 70% win rate with PF 1.1 is a time bomb. Both print a positive Sharpe. Calmar is particularly revealing for Indian equities, which can sit underwater for 12-18 months after corrections — it directly answers "how much pain per unit of return?"

---

### [TASK-0017] Analytics — drawdown duration tracking

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** Max drawdown depth (%) is in the report, but duration is not. A 15% drawdown that recovers in 3 weeks is survivable; a 15% drawdown that lasts 9 months is not. Drawdown duration is an under-appreciated tell — if out-of-sample recovery time diverges from in-sample, the strategy is decaying.
- **Acceptance criteria:**
  - [ ] `Report` gains: `MaxDrawdownDuration time.Duration` (wall time from peak to recovery or end of test)
  - [ ] Computed from the equity curve time series (requires TASK-0013)
  - [ ] If the equity curve never fully recovers by end of test, duration = time from peak to last bar
  - [ ] Tests: equity curve with known peak/trough/recovery → expected duration
- **Notes:** Blocked on TASK-0013 for equity curve. Can be implemented immediately after.

---

### [TASK-0019] Strategy — MACD trend-following

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** MACD (12/26/9 defaults) is the third baseline trend strategy, complementing SMA crossover. It adds signal-line smoothing which reduces whipsaws compared to raw SMA crossover. Running all three momentum strategies together reveals whether the edge (if any) is robust to parameterization or specific to one configuration.
- **Acceptance criteria:**
  - [ ] `strategies/macdtrend/` package implementing `Strategy` interface
  - [ ] Uses `github.com/markcheno/go-talib` MACD — no hand-rolled computation
  - [ ] Configurable fast (12), slow (26), signal (9) periods; defaults baked in
  - [ ] Signal: buy when MACD line crosses above signal line; sell when crosses below
  - [ ] Lookback = slow + signal periods
  - [ ] Tests: known OHLCV sequence → expected signals
- **Notes:** **Conditional gate — do not start until TASK-0012 (SMA crossover) and TASK-0015 (RSI) results are in and reviewed against TASK-0018 (buy-and-hold benchmark).** If neither SMA crossover nor RSI beats buy-and-hold after costs with Sharpe >= 0.5, cancel this task and go looking for a different edge thesis. MACD will not rescue a regime that doesn't reward trend-following.

---

### [TASK-0020] Strategy — Bollinger Band mean-reversion

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** Bollinger Band mean-reversion (buy at lower band, sell at upper band or middle) is the volatility-adaptive counterpart to RSI mean-reversion. It adapts its thresholds to current volatility, which makes it more robust across different market regimes than fixed RSI thresholds.
- **Acceptance criteria:**
  - [ ] `strategies/bollingermeanrev/` package implementing `Strategy` interface
  - [ ] Uses `github.com/markcheno/go-talib` Bollinger Bands — no hand-rolled computation
  - [ ] Configurable period (20), std-dev multiplier (2.0); defaults baked in
  - [ ] Buy signal: close touches or crosses below lower band; sell signal: close touches or crosses above upper band
  - [ ] Lookback = period
  - [ ] Tests: known OHLCV sequence → expected signals
- **Notes:** **Conditional gate — do not start until TASK-0015 (RSI mean-reversion) results are reviewed against TASK-0018 (buy-and-hold benchmark).** If RSI mean-reversion doesn't beat buy-and-hold after costs with Sharpe >= 0.5, cancel this task. Bollinger Bands are a variation of the same mean-reversion thesis — if the thesis has no pulse, the variation won't save it.

---

### [TASK-0022] Rigor — walk-forward validation framework

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-10
- **Source:** session
- **Context:** Running a strategy over the full historical period and reporting the result is in-sample evaluation — it tells you nothing about whether the edge is real. Walk-forward validation splits the data into rolling train/test windows and measures out-of-sample performance independently. This is the minimum viable defense against overfitting.
- **Acceptance criteria:**
  - [ ] `internal/walkforward/` package with `Run(cfg WalkForwardConfig, provider, strategy) []WindowResult`
  - [ ] `WalkForwardConfig`: in-sample window duration, out-of-sample window duration, step size, instrument, from/to
  - [ ] Each `WindowResult`: in-sample period, out-of-sample period, in-sample Sharpe, out-of-sample Sharpe, trade count
  - [ ] Report flags if avg out-of-sample Sharpe < 50% of avg in-sample Sharpe (likely overfit)
  - [ ] Tests: synthetic candle data with known signal → expected window results
- **Notes:** Strategy interface is stateless (takes `[]Candle`, returns signal), so walk-forward doesn't require strategy re-fitting. This is validation-only for rule-based strategies.

---

### [TASK-0023] Rigor — parameter sweep runner

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-10
- **Source:** session
- **Context:** A robust edge has a plateau of working parameter values, not a peak. Parameter sweeps reveal whether a strategy's Sharpe is robust to small parameter changes (real edge) or collapses when you nudge the lookback by 2 bars (curve-fitted noise).
- **Acceptance criteria:**
  - [ ] `internal/sweep/` package with `Run(cfg SweepConfig) []SweepResult`
  - [ ] `SweepConfig`: parameter name, range (min, max, step), fixed engine config, instrument, date range
  - [ ] Each `SweepResult`: parameter value, Sharpe, total P&L, trade count, max drawdown
  - [ ] Output: ranked table of results + identification of the "plateau" (parameter range where Sharpe stays within 80% of peak)
  - [ ] Tests: synthetic strategy with known optimal parameter → sweep correctly identifies the peak
- **Notes:** Single-parameter first. Multi-dimensional grid search is combinatorially expensive and a path to overfitting. Implement after at least one strategy passes walk-forward (TASK-0022).

---

### [TASK-0024] Rigor — Monte Carlo bootstrap for Sharpe confidence intervals

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-10
- **Source:** session
- **Context:** A single Sharpe number from a backtest is a point estimate with unknown uncertainty. Monte Carlo bootstrap resamples the trade return sequence thousands of times to produce a confidence interval. The p5 Sharpe from this output is the kill-switch threshold — halt when live rolling Sharpe drops below it.
- **Acceptance criteria:**
  - [ ] `internal/montecarlo/` package with `Bootstrap(trades []model.Trade, nSimulations int) BootstrapResult`
  - [ ] `BootstrapResult`: mean Sharpe, Sharpe p5/p50/p95, worst drawdown p5/p50/p95, probability of positive Sharpe
  - [ ] Resampling: draw with replacement from the trade return series, recompute Sharpe each iteration
  - [ ] Default 10,000 simulations; configurable
  - [ ] Tests: known return distribution → expected confidence interval shape (statistically sound, not exact values)
- **Notes:** The p5 Sharpe from this output is the kill-switch threshold — document this explicitly in code comments. Implement last — it only adds value once you have a strategy that's earned a real Sharpe to evaluate.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
