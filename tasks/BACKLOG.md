# Project Task Backlog

**Last updated:** 2026-04-16 | **Open tasks:** 5 | **Next up:** TASK-0024

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

### [TASK-0028] Backtest — run SMA crossover and RSI mean-rev baselines, check proliferation gate

- **Status:** in-progress
- **Priority:** high
- **Created:** 2026-04-15
- **Source:** user
- **Context:** Two strategies are implemented and the full pipeline is wired. This is the first
  live run against real NSE data and answers whether either strategy has a detectable edge over
  the 2018–2024 window. Period pre-committed (Marcus) to include the 2020 crash and 2022 choppy
  regime.
- **Acceptance criteria:**
  - [x] Instrument declared here before any run: **INSTRUMENT: NSE:RELIANCE** (Reliance Industries,
        Nifty 50 constituent, continuous trading since before 2018)
  - [x] SMA crossover run: `--strategy sma-crossover`, `NSE:RELIANCE`,
        `--from 2018-01-01 --to 2025-01-01 --timeframe daily`, `--sizing-model vol-target --vol-target 0.10`,
        Zerodha commission defaults, result saved to `runs/sma-crossover-2018-2024.json`
  - [x] RSI mean-rev run: same instrument, same period, same sizing and cost model,
        saved to `runs/rsi-mean-rev-2018-2024.json`
  - [x] Both outputs include benchmark comparison (TASK-0018 — already in output)
  - [x] Proliferation gate checked:
        SMA crossover Sharpe = 0.447 — **FAILS gate (< 0.5) → TASK-0019 cancelled**
        RSI mean-rev Sharpe = 0.469, 7 trades — **FAILS gate (< 0.5, sample too small) → TASK-0020 cancelled**
  - [x] Gate decisions recorded in `decisions/algorithm/` (one entry per strategy —
        `2026-04-16-sma-crossover-proliferation-gate-failed.md` and
        `2026-04-16-rsi-mean-reversion-proliferation-gate-failed.md`)
  - [ ] Equity curve reviewed across three regime windows: 2018–2019 (pre-crash), 2020–2021
        (crash + recovery), 2022–2024 (grind) — confirm neither strategy shows edge in only
        one window
- **Notes:** MaxDrawdown bug fixed 2026-04-16 (was accumulating P&L from 0; now uses per-bar
  equity curve via `computeMaxDrawdownDepth`). Re-run results: SMA MaxDrawdown 16.38%, RSI
  MaxDrawdown 17.36%. CalmarRatio corrected accordingly. All downstream rigor tasks (TASK-0024,
  TASK-0022, TASK-0026) depend on the trade return series this run produces.

---

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

### [TASK-0026] Rigor — kill-switch definition per strategy

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-13
- **Source:** session
- **Blocked by:** TASK-0024 (Monte Carlo bootstrap — kill-switch thresholds derived from bootstrapped distribution)
- **Context:** Before any strategy runs with real capital, a pre-committed halt condition must exist. Without it, a normal drawdown turns into parameter tweaking and re-running, which is how you overfit live. The kill-switch is what separates a system from a hobby.
- **Acceptance criteria:**
  - [ ] For each strategy, after Monte Carlo bootstrap, define and document: rolling 6-month Sharpe threshold (5th percentile of bootstrapped distribution), max drawdown threshold (1.5× worst in-sample drawdown), max drawdown recovery time threshold (2× worst in-sample recovery)
  - [ ] Kill-switch parameters written to `decisions/` alongside each strategy's backtest results
  - [ ] `internal/analytics` or `internal/output` can compare rolling live metrics against these thresholds and flag when a kill-switch is approached
- **Notes:** The rule when the line is hit: halt and re-evaluate from scratch — never retune parameters mid-drawdown. "Tweak parameters and restart while still in the drawdown" is how a single bad regime turns into a permanent overfit. This task has no implementation until TASK-0024 is done.

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

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

### [TASK-0024] Rigor — Monte Carlo bootstrap for Sharpe confidence intervals

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-10
- **Source:** session
- **Context:** A single Sharpe number from a backtest is a point estimate with unknown uncertainty. Monte Carlo bootstrap resamples the trade return sequence thousands of times to produce a confidence interval. The p5 Sharpe from this output is the kill-switch threshold — halt when live rolling Sharpe drops below it.
- **Acceptance criteria:**
  - [ ] `internal/montecarlo/` package with `Bootstrap(trades []model.Trade, nSimulations int) BootstrapResult`
  - [ ] `BootstrapResult`: mean Sharpe, Sharpe p5/p50/p95, worst drawdown p5/p50/p95, probability of positive Sharpe
  - [ ] Resampling: draw with replacement from the trade return series, recompute Sharpe each iteration
  - [ ] Default 10,000 simulations; configurable
  - [ ] Tests: known return distribution → expected confidence interval shape (statistically sound, not exact values)
- **Notes:** The p5 Sharpe from this output is the kill-switch threshold — document this explicitly in code comments. Reprioritized from low: must run before walk-forward (TASK-0022) and parameter sweep (TASK-0023), because the bootstrapped distribution is the input to the kill-switch definition (TASK-0026). Implement once at least one strategy has results worth evaluating.

---

### [TASK-0027] Rigor — strategy correlation analysis before portfolio assembly

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-13
- **Source:** session
- **Context:** Running multiple strategies together only provides diversification if they are genuinely uncorrelated. RSI mean-rev and Bollinger Band mean-rev on the same instrument will likely be 0.7+ correlated on daily returns — running both at full vol-target sizing is doubling the bet, not diversifying. Before any multi-strategy portfolio is assembled, pairwise correlations must be measured and sizing adjusted accordingly.
- **Acceptance criteria:**
  - [ ] After at least two strategy results are available, compute pairwise Pearson correlation of per-bar equity curve returns for each strategy pair
  - [ ] Test correlation in stress sub-periods (2020 crash, 2022 bear) separately from the full-period average — strategies that appear uncorrelated on average often correlate strongly in drawdowns
  - [ ] `internal/analytics` or `internal/output` produces a correlation matrix table alongside multi-strategy results
  - [ ] Tests: known equity curve pairs with known correlation → expected matrix values
- **Notes:** Do not start until at least two strategy results exist. Momentum strategies (SMA crossover, MACD) will likely correlate with each other; mean-reversion strategies (RSI, Bollinger) will correlate with each other; the interesting question is momentum vs mean-reversion cross-correlation, which should be low or negative. If two strategies are >0.7 correlated, halve the combined vol-target allocation rather than running both at full size.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
