# Project Task Backlog

**Last updated:** 2026-04-20 | **Open tasks:** 6 | **Next up:** TASK-0024

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

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
  - [ ] `Trade.ReturnOnNotional() float64` method on `pkg/model/trade.go` — returns `RealizedPnL / (EntryPrice * Quantity)`; this is the per-trade return the bootstrap resamples from
  - [ ] `internal/montecarlo/` package with `Bootstrap(trades []model.Trade, cfg BootstrapConfig) BootstrapResult`
  - [ ] `BootstrapConfig`: `NSimulations int` (default 10,000), `Seed int64` (explicit; logged in output for reproducibility)
  - [ ] `BootstrapResult`: mean Sharpe, Sharpe p5/p50/p95, worst drawdown p5/p50/p95, probability of positive Sharpe
  - [ ] Resampling: draw with replacement from the trade return series via `ReturnOnNotional()`, recompute Sharpe each iteration; RNG seeded from `cfg.Seed` using `math/rand/v2`
  - [ ] Tests: known return distribution → expected confidence interval shape (statistically sound, not exact values)
- **Notes:** The p5 Sharpe from this output is the kill-switch threshold — document this explicitly in code comments. Reprioritized from low: must run before walk-forward (TASK-0022), because the bootstrapped distribution is the input to the kill-switch definition (TASK-0026). Implement once at least one strategy has results worth evaluating. Updated 2026-04-16: added `Trade.ReturnOnNotional()` requirement and explicit seed for determinism.

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

### [TASK-0035] Tooling — multi-instrument sweep CLI (`cmd/universe-sweep`)

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-16
- **Source:** session
- **Context:** Both SMA and RSI results are from a single instrument. Cross-instrument evidence
  is needed to determine whether the lack of edge is RELIANCE-specific or thesis-wide. Running
  the same strategy across 10-15 Nifty 50 large caps automatically surfaces whether there's
  clustering of positive results elsewhere.
- **Acceptance criteria:**
  - [ ] `cmd/universe-sweep` CLI: `--universe <file>`, `--strategy`, `--from`, `--to`, `--timeframe`, standard cost flags
  - [ ] Universe file is plain text or YAML list of instrument strings; `universes/nifty50-large-cap.yaml` created with 10-15 liquid Nifty 50 constituents
  - [ ] Runs per instrument via `errgroup` concurrency; output is CSV ranked by Sharpe
  - [ ] Signal frequency gate (TASK-0030) applied per instrument; insufficient-sample results flagged in output
  - [ ] Tests: synthetic 2-instrument universe → 2-row output CSV
- **Notes:** Do not start until TASK-0030 (signal frequency gate) is done — the per-instrument
  output is misleading without the gate applied automatically.

---

### [TASK-0036] Research tooling — Python notebooks layer + file contract

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-16
- **Source:** session
- **Context:** The 2D heatmap, equity curve plots, and regime visualizations have nowhere to live.
  A `notebooks/` directory with a documented file contract is the prerequisite for any
  visualization work and establishes the Go-writes/Python-reads boundary explicitly.
- **Acceptance criteria:**
  - [ ] `notebooks/` directory at project root, version-controlled
  - [ ] `notebooks/README.md` documents file contract: equity curve CSV schema, sweep CSV schema, analytics JSON schema, column names, timestamp format
  - [ ] `notebooks/requirements.txt` with pyarrow, pandas, matplotlib pinned
  - [ ] At least one working notebook: `notebooks/equity-curve.ipynb` reads `runs/<name>-curve.csv` and plots equity curve with regime shading
- **Notes:** Depends on TASK-0029 (equity curve CSV output) for the first working notebook.
  The file contract in README.md is the formal boundary — Python never feeds back into Go inputs.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
