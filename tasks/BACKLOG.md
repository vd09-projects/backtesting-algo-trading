# Project Task Backlog

**Last updated:** 2026-04-19 | **Open tasks:** 11 | **Next up:** TASK-0031

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

### [TASK-0031] Research — RSI signal frequency diagnostic on NSE:RELIANCE

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-16
- **Source:** session
- **Context:** Pre-condition for re-evaluating the mean-reversion thesis. 7 trades in 7 years
  likely means RSI(14) at fixed 30/70 thresholds rarely breaches those levels on RELIANCE daily
  — not that mean-reversion has no edge. Confirming the calibration failure is the diagnostic
  step before any re-test. Marcus pre-committed this as the required check.
- **Acceptance criteria:**
  - [ ] Count how many times RSI(14) actually breached 30 and 70 on RELIANCE daily 2018–2025
  - [ ] If fewer than 30 signals: record decision that fixed 30/70 is miscalibrated for RELIANCE;
        propose adaptive threshold re-test (pre-commit parameters before running)
  - [ ] If > 30 signals: investigate whether a signal generation bug suppressed entries
        (vol-targeting zeroing out entries during high-vol events is the likely cause)
  - [ ] Outcome documented in `decisions/algorithm/`
- **Notes:** This is a research step, not infrastructure. May need a diagnostic output mode or a
  small script to count signal triggers without requiring closed trades. Pre-commit the exact
  diagnostic spec before running. See Marcus's decision: rsi-signal-frequency-diagnostic (2026-04-16).

---

### [TASK-0032] Tooling — 2D parameter sweep with DSR calculation

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-16
- **Source:** session
- **Context:** The 1D sweep shows the best parameter value but not whether there's a plateau of
  working values (robustness) or just a narrow peak (overfit). A 2D grid sweep produces the
  robustness surface. DSR (Deflated Sharpe Ratio) correction for multiple testing is computed
  from the variant count — tells you how much the "best" Sharpe is inflated by the search.
- **Acceptance criteria:**
  - [ ] `internal/sweep2d` package with `Run(cfg Config2D, provider, factory) Report2D`
  - [ ] `Config2D`: two parameter ranges (name, min, max, step each) + engine config
  - [ ] `Report2D`: `[i][j]` grid of results (Sharpe, trade count, max drawdown), variant count, DSR-corrected peak Sharpe
  - [ ] Runs use `errgroup` with `GOMAXPROCS` concurrency; results stored at fixed indices (deterministic output order)
  - [ ] `DSR(observedSharpe, nTrials, nObservations float64) float64` function in `internal/analytics`
  - [ ] `sweep.Report` gains `VariantCount int`; `WriteSweep` output shows DSR-corrected Sharpe alongside raw
  - [ ] Output: CSV matrix written to `runs/` for Python heatmap consumption
  - [ ] Tests: 2×2 grid → 4 results; DSR returns lower Sharpe for higher trial counts
- **Notes:** DSR formula from Bailey & López de Prado. CSV is the handoff boundary to Python
  notebooks. Priya's decision: sweep2d-package-structure (2026-04-16).

---

### [TASK-0033] Tooling — automated proliferation gate PASS/FAIL in CLI output

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-16
- **Source:** session
- **Context:** The proliferation gate check is currently manual — read the Sharpe from JSON and
  compare to 0.5. The CLI should print PASS or FAIL automatically so the gate is never
  accidentally skipped and the result is unambiguous in the output.
- **Acceptance criteria:**
  - [ ] `--proliferation-gate-threshold` flag on `cmd/backtest` (default 0.0 = disabled; set to 0.5 for NSE daily)
  - [ ] When threshold > 0, `output.printSummary` prints `Proliferation gate (≥0.50): PASS` or `FAIL` with actual Sharpe
  - [ ] Gate check skipped (not printed) when threshold is 0, or when `TradeMetricsInsufficient`/`CurveMetricsInsufficient` is set
  - [ ] Tests: PASS/FAIL printed correctly; no output when threshold is 0 or sample insufficient
- **Notes:** Insufficient sample and gate failure are different outcomes and must print differently.
  Depends on TASK-0030 (signal frequency gate) for the insufficient-sample suppression.

---

### [TASK-0034] Analytics — regime-split report in backtest output

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-16
- **Source:** session
- **Context:** Per-regime Sharpe should be automatic in every backtest output so regime
  concentration is visible without manual analysis. NSE went through distinct regimes:
  2018–2019 (pre-COVID steady), 2020–2021 (crash + recovery), 2022–2024 (grind).
- **Acceptance criteria:**
  - [ ] `ComputeRegimeSplits(curve []model.EquityPoint, regimes []Regime) []RegimeReport` in `internal/analytics`
  - [ ] `Regime`: name string, from/to `time.Time`; `RegimeReport`: name, period, Sharpe, max drawdown
  - [ ] Pre-defined `NSERegimes2018_2024` constant slice with the three windows above
  - [ ] `cmd/backtest` prints regime table when `--output-curve` path is set (curve required for splits)
  - [ ] Tests: known equity curve split across regime boundaries → expected per-regime Sharpe values
- **Notes:** Depends on TASK-0029 (equity curve output) — regime splits require the timestamped
  equity curve. Directly unblocks TASK-0028's final acceptance criterion once both are built.

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
