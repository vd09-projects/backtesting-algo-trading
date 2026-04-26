# Project Task Backlog

**Last updated:** 2026-04-26 | **Open tasks:** 19 | **Next up:** TASK-0042

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

_No tasks in progress._

---

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0042] Strategy — Bollinger Band Mean Reversion

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** Supersedes TASK-0020 (cancelled 2026-04-16 when RSI failed single-instrument proliferation gate). Bollinger is a vol-normalized mean reversion — buys when price is unusually extended below the mean, more adaptive to regime changes than fixed RSI thresholds.
- **Acceptance criteria:**
  - [ ] `strategies/bollinger/` package implementing `pkg/strategy.Strategy`
  - [ ] Parameters: `period int` (20), `numStdDev float64` (2.0), `timeframe model.Timeframe`
  - [ ] Uses `talib.Bbands(closes, period, numStdDev, numStdDev, talib.SMA)` returning upper, middle, lower
  - [ ] Buy signal: Close crosses below lower band (was ≥ lower previous bar, is < lower current bar)
  - [ ] Sell signal: Close crosses above upper band (was ≤ upper previous bar, is > upper current bar)
  - [ ] Strict crossover detection — consistent with SMA and MACD strategy patterns
  - [ ] `Lookback()` returns `period`
  - [ ] Registered in all three CLIs; sweep params: `bb-period`, `bb-num-std-dev`
  - [ ] Tests written before implementation (TDD)
- **Notes:** Exit is on upper band touch (not return to middle). Positions can be held for weeks in trending markets. A "return-to-mean" exit variant is worth a separate sweep after initial validation — not in scope here.

---

### [TASK-0043] Strategy — 12-Month Rate-of-Change Momentum

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** New strategy family. The classic academic equity momentum factor (Jegadeesh & Titman 1993). Edge thesis: institutional herding — stocks that outperformed in the past 11 months continue to outperform for another month on average. The "skip last month" convention avoids short-term reversal contamination and is the standard AQR/academic implementation.
- **Acceptance criteria:**
  - [ ] `strategies/momentum/` package implementing `pkg/strategy.Strategy`
  - [ ] Parameters: `lookback int` (231 = 252 − 21 bars, skip-last-month convention), `threshold float64` (10.0, representing 10%), `timeframe model.Timeframe`
  - [ ] Uses `talib.Roc(closes, lookback)`
  - [ ] Buy signal: `ROC(lookback) > threshold`
  - [ ] Sell signal: `ROC(lookback) < -threshold`
  - [ ] `Lookback()` returns `lookback + 1`
  - [ ] Registered in all three CLIs; sweep param: `momentum-lookback`, `momentum-threshold`
  - [ ] Tests written before implementation (TDD)
- **Notes:** The 231-bar lookback (not 252) is an algorithm decision pre-committed by Marcus in this session. The skip-last-month convention is the Jegadeesh-Titman standard — produces a cleaner momentum signal than raw 12-month return. Symmetric threshold for buy/sell keeps the strategy balanced.

---

### [TASK-0039] Strategy — time-based exit wrapper (`pkg/strategy.TimedExit`)

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** All current strategies hold until a reversal signal fires, which can mean months. For 1-7 day hold strategies, the engine needs a way to force-close after N bars regardless of the inner strategy's signal. A wrapper in `pkg/strategy` achieves this without any engine changes.
- **Acceptance criteria:**
  - [ ] `TimedExit` struct in `pkg/strategy/` implements the `Strategy` interface
  - [ ] Constructor: `NewTimedExit(inner Strategy, maxHoldBars int) Strategy`
  - [ ] `Name()` returns `"timed-exit(" + inner.Name() + ")"`
  - [ ] `Lookback()` returns `inner.Lookback()`
  - [ ] `Timeframe()` returns `inner.Timeframe()`
  - [ ] `Next()`: if inner strategy emits `SignalBuy`, record entry bar index; if `barsSinceEntry >= maxHoldBars` and in position, emit `SignalSell` regardless of inner signal; inner `SignalSell` honored before timer fires
  - [ ] Tests written before implementation: (a) inner sell fires before timer — timer resets; (b) timer fires before inner sell; (c) exact boundary — sell fires at bar maxHoldBars, not maxHoldBars-1 or maxHoldBars+1; (d) no position held during warmup
- **Notes:** No engine changes. Composable with all existing and new strategies. Not on the critical path for the Phase 1 evaluation pipeline — reordered after TASK-0040–0043 so the four strategy builds (which unblock TASK-0050) are picked up first.

---

### [TASK-0047] Engine — MIS (intraday) commission model

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** Intraday (MIS) orders have a different STT structure than delivery (CNC): 0.025% on sell leg only vs 0.10% on both legs. Without this, intraday backtests overstate costs by ~₹60 per round-trip on a ₹30K trade. Unblocked by TASK-0038 (done 2026-04-25).
- **Acceptance criteria:**
  - [x] TASK-0038 is done
  - [ ] New `CommissionModel` constant `CommissionZerodhaFullMIS` added to `pkg/model/order.go`
  - [ ] `calcCosts` switch: new case applies STT 0.025% on sell leg only; all other charges same as `CommissionZerodhaFull`
  - [ ] Round-trip cost on ₹30K intraday trade equals ~₹55-60 (hand-verified in test)
  - [ ] Golden test: known notional → exact expected per-leg cost for MIS
  - [ ] Tests written before implementation (TDD)

---

### [TASK-0044] Tooling — `cmd/sweep2d` CLI entrypoint

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** `internal/sweep2d` is complete (implemented in TASK-0032) but has no CLI entrypoint. The 2D sweep is needed to test parameter interaction surfaces (e.g. SMA fast × slow period grid) and produces DSR-corrected peak Sharpe automatically.
- **Acceptance criteria:**
  - [ ] `cmd/sweep2d/main.go` created; wires to `internal/sweep2d` package
  - [ ] Flags: `--instrument`, `--from`, `--to`, `--timeframe`, `--cash`, `--strategy`, `--p1-name/--p1-min/--p1-max/--p1-step`, `--p2-name/--p2-min/--p2-max/--p2-step`, `--out` (CSV output path; stdout if omitted)
  - [ ] Fixed-param flags for each strategy (same set as `cmd/sweep`)
  - [ ] Supports `sma-crossover` (fast × slow) and `rsi-mean-reversion` (period × oversold) initially; extend as new strategies land
  - [ ] DSR-corrected peak Sharpe printed to stderr alongside CSV path
  - [ ] End-to-end smoke test: runs on synthetic/cached data, produces valid CSV with correct column headers
- **Notes:** All the heavy lifting is in `internal/sweep2d`. This is purely CLI wiring — should be the lightest task in Phase 1. `internal/cmdutil.BuildProvider` (extracted in TASK-0035) handles provider setup.

---

### [TASK-0045] Research spike — NIFTY TRI benchmark data availability

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** Marcus established that the correct benchmark is NIFTY 50 Total Return Index (TRI), not the price index. TRI includes dividend reinvestment (~1.3-1.5% per year), which matters over multi-year comparisons. Before any benchmark code is written, we need to know whether Zerodha Kite provides TRI data.
- **Acceptance criteria:**
  - [ ] Timebox: 2 hours
  - [ ] Inspect the downloaded Zerodha instruments CSV for any instrument matching patterns: `NIFTY.*TOTAL`, `NIFTY.*TRI`, `NIFTY.*RETURN`
  - [ ] If found: document the exact instrument string and verify a candle fetch returns plausible TRI values
  - [ ] If not found: document the two implementation options — (a) external CSV loader pointing at NSE's published TRI data, (b) a second `DataProvider` implementation for NSE data
  - [ ] Outcome recorded as a decision in `decisions/infrastructure/`
  - [ ] No code written until the decision is recorded
- **Notes:** NSE publishes historical TRI data at nseindia.com/products/content/equities/indices/historical_total_returns.htm as downloadable CSV. If Zerodha doesn't have it, the external CSV loader option is simpler than a second provider.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

### [TASK-0049] Evaluation — pre-commit gate definitions before any backtest runs

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0047 (MIS commission model) — cost model family should be complete before cost-adjusted Sharpe gates are committed. TASK-0038 done 2026-04-25.
- **Context:** Every acceptance gate for the evaluation pipeline must be written and committed before any strategy result is seen. Gates set after seeing results are post-hoc rationalization, not gates. This task produces six decision documents in `decisions/algorithm/`, all dated before any Stage 3+ run begins.
- **Acceptance criteria:**
  - [ ] Decision record: universe gate — DSR-corrected average Sharpe > 0 AND ≥ 40% of instruments show positive Sharpe with ≥ 30 trades
  - [ ] Decision record: walk-forward gate — OverfitFlag = false AND NegativeFoldFlag = false (both required, not either)
  - [ ] Decision record: bootstrap gate — SharpeP5 > 0 AND P(Sharpe > 0) > 80%
  - [ ] Decision record: regime gate — no single regime accounts for more than 70% of full-period Sharpe
  - [ ] Decision record: correlation gate — full-period r < 0.7 AND stress-period r < 0.6 between any two portfolio strategies
  - [ ] Decision record: kill-switch gate — thresholds derived from bootstrap p5 Sharpe (not round numbers), documented before live capital
  - [ ] All six decision files in `decisions/algorithm/` dated before any Stage 3+ run begins
- **Notes:** Marcus superseded the single-instrument proliferation gate (2026-04-10) in this session (2026-04-25). The new cross-instrument gate formally replaces it and must be recorded here. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0050] Evaluation — signal frequency audit (all 6 strategies × 15 instruments)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0040, TASK-0041, TASK-0042, TASK-0043 (all four new strategies must be built before this can run; TASK-0038 done 2026-04-25)
- **Context:** Before running any full backtest, audit how many trades each strategy generates per instrument in the 2018-2023 window with default parameters. Any combination producing < 30 trades is excluded from further analysis on that instrument. The 2026-04-19 RSI diagnostic showed this problem is real — scaling it to 6 strategies × 15 instruments prevents wasted analysis runs downstream.
- **Acceptance criteria:**
  - [ ] Run signal frequency diagnostic for each strategy on all 15 instruments in `universes/nifty50-large-cap.yaml` over 2018-01-01 to 2024-01-01
  - [ ] Produce and save matrix: strategy × instrument → trade count to `runs/signal-frequency-audit-YYYY-MM-DD.csv`
  - [ ] Flag any cell with < 30 trades as EXCLUDED
  - [ ] If any strategy fires < 30 trades across the ENTIRE 15-instrument universe combined, kill that strategy before any full backtest runs
- **Notes:** Not optional. Skipping this step means discovering insufficient trades after running the full validation pipeline, which wastes significant compute and analysis time. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0051] Evaluation — in-sample baseline and parameter sensitivity (all 6 strategies, RELIANCE, 2018-2023)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0049, TASK-0050
- **Context:** Orientation run on a single instrument to understand each strategy's behavior and find the robust parameter region via 1D sweep. Not a gate — the universe sweep is the gate. The output of this task is the plateau-midpoint parameter for each strategy, which is what gets used in the universe sweep.
- **Acceptance criteria:**
  - [ ] Run `cmd/backtest` for all 6 strategies on NSE:RELIANCE 2018-01-01 to 2024-01-01 with `CommissionZerodhaFull` and default parameters
  - [ ] Record for each: Sharpe, regime splits (3 NSE windows), trade count, max drawdown, bootstrap p5 Sharpe
  - [ ] Run 1D parameter sweep (`cmd/sweep`) for each strategy on its key parameter (rsi-period, sma-fast/slow, macd-fast/slow, bb-period, donchian-period, momentum-lookback)
  - [ ] Identify plateau range (within 80% of peak Sharpe) for each strategy
  - [ ] Select plateau-midpoint parameter for each strategy for use in universe sweep; if no plateau exists, flag strategy as "sensitivity concern" — this does not disqualify it, but the flag is recorded
  - [ ] Results saved to `runs/baseline-YYYY-MM-DD/` with JSON and CSV outputs
- **Notes:** A strategy with Sharpe < 0 on RELIANCE is not killed here — the universe sweep is the primary gate. This step is for orientation and parameter selection, not elimination. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0052] Evaluation — universe sweep (cross-instrument primary gate)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0051 (plateau-midpoint parameters needed for the sweep run)
- **Context:** The primary validation gate for all six strategy families. A real edge should work on more than one instrument. Strategies that pass on RELIANCE only are not robust — they passed the wrong test. This run determines which strategies and instruments survive into walk-forward.
- **Acceptance criteria:**
  - [ ] Run `cmd/universe-sweep` for all 6 strategies using plateau-midpoint parameters from TASK-0051, across all 15 instruments in `universes/nifty50-large-cap.yaml`, 2018-01-01 to 2024-01-01
  - [ ] Apply universe gate (from TASK-0049): DSR-corrected average Sharpe > 0 AND ≥ 40% of instruments show positive Sharpe with ≥ 30 trades
  - [ ] Strategies that fail the gate are killed — record kill decision in `decisions/algorithm/`
  - [ ] Produce explicit survivor matrix: strategy × instrument (pass/fail per cell)
  - [ ] Apply regime gate (from TASK-0049): no single regime accounts for > 70% of total Sharpe; strategies failing the regime gate are flagged even if they pass the universe gate
  - [ ] Results saved to `runs/universe-sweep-YYYY-MM-DD.csv`
- **Notes:** This replaces the single-instrument proliferation gate (2026-04-10 decision, formally superseded in TASK-0049). Marcus signed off on the supersession 2026-04-25. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0053] Evaluation — walk-forward validation on universe sweep survivors

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0052
- **Context:** Walk-forward tests whether fixed parameters remain stable across time periods not used for parameter selection. A strategy that survives the universe sweep but fails walk-forward on most of its passing instruments is overfitting to the historical regime, not capturing a transferable edge. Parameters are fixed — no reoptimization per fold.
- **Acceptance criteria:**
  - [ ] Run walk-forward for each surviving strategy × instrument pair from TASK-0052, 2018-01-01 to 2024-12-31, 2yr IS / 1yr OOS / 1yr step
  - [ ] Apply walk-forward gate (from TASK-0049): OverfitFlag = false AND NegativeFoldFlag = false
  - [ ] A strategy must pass walk-forward on at least as many instruments as it passed the universe gate — if it passes universe on 8 instruments but walk-forward on only 3, the strategy is killed
  - [ ] Record surviving strategy × instrument pairs with: AvgInSampleSharpe, AvgOutOfSampleSharpe, OOS/IS ratio, NegativeFoldCount
- **Notes:** Walk-forward window per 2026-04-22 decision: 2yr IS / 1yr OOS / 1yr step. OverfitFlag fires when AvgOOSSharpe < 50% of AvgISSharpe — both flags must be false. OOS Sharpe that is positive but below the 50% floor still fails the OverfitFlag gate. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0054] Evaluation — Monte Carlo bootstrap on walk-forward survivors

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0053
- **Context:** Bootstrap produces the confidence interval on the Sharpe estimate and the kill-switch Sharpe threshold. A strategy with a high point-estimate Sharpe but wide bootstrap distribution (low p5) has too much sampling variance to trust with real capital. The p5 Sharpe from this run becomes the live kill-switch threshold, not a round number.
- **Acceptance criteria:**
  - [ ] Run `cmd/backtest --bootstrap` for each surviving strategy × instrument pair from TASK-0053, 10,000 simulations
  - [ ] Apply bootstrap gate (from TASK-0049): SharpeP5 > 0 AND P(Sharpe > 0) > 80%
  - [ ] Record for each survivor: SharpeP5, SharpeP50, SharpeP95, ProbPositiveSharpe, WorstDrawdownP95
  - [ ] SharpeP5 value recorded as the kill-switch Sharpe threshold for that strategy × instrument pair — this feeds directly into TASK-0056
  - [ ] Kill strategies failing the bootstrap gate; record kill decision in `decisions/algorithm/`
  - [ ] Bootstrap seed logged with every result for reproducibility
- **Notes:** Bootstrap Sharpe is per-trade non-annualized per 2026-04-20 decision. Kill-switch comparison must use the identical formula. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0055] Evaluation — cross-strategy correlation and portfolio construction

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0054
- **Context:** Selects the final portfolio from bootstrap survivors. Two correlated strategies do not provide diversification — they add correlated risk, especially in the stress periods (2020, 2022) when diversification is most needed. The portfolio at ₹3 lakh targets ~10% annualized vol using vol-targeting sizing per strategy.
- **Acceptance criteria:**
  - [ ] Run `cmd/correlate` for all surviving strategy × instrument pairs from TASK-0054 (full equity curves)
  - [ ] Apply correlation gate (from TASK-0049): full-period r < 0.7 AND stress-period r < 0.6 for every pair in the final portfolio
  - [ ] If two strategies are correlated and from the same edge bucket, keep only the higher-DSR-Sharpe one
  - [ ] Select 2-4 uncorrelated survivors for the portfolio
  - [ ] Define capital allocation per strategy: combined portfolio targets ~10% annualized vol using `SizingVolatilityTarget`
  - [ ] Record portfolio composition and sizing rule in `decisions/algorithm/`
  - [ ] Record excluded strategies with reasons (correlation, gate failure, or sizing constraint)
- **Notes:** At ₹3 lakh total capital with vol-targeting, each strategy typically receives 20-50% of capital depending on realized volatility. No leverage at this stage. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0056] Evaluation — pre-live brief: kill-switch thresholds and go/no-go sign-off

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0055, TASK-0048 (cmd/monitor must exist for weekly monitoring cadence)
- **Context:** The final checkpoint before any live capital is allocated. Documents specific kill-switch thresholds, capital allocation, and the explicit go/no-go verdict for each portfolio strategy. No strategy goes live without this document existing and dated before the first trade. This is what the algo reviewer signs.
- **Acceptance criteria:**
  - [ ] For each portfolio strategy: kill-switch thresholds recorded — SharpeP5 threshold (from TASK-0054), MaxDrawdownPct (1.5× in-sample worst), MaxDDDuration (2× in-sample worst)
  - [ ] Thresholds written to `decisions/algorithm/YYYY-MM-DD-kill-switch-{strategy}.md` before first trade
  - [ ] Capital allocation per strategy documented in ₹ and % of ₹3 lakh total
  - [ ] Monitoring cadence documented: weekly kill-switch check via `cmd/monitor`
  - [ ] Explicit go/no-go verdict per strategy: APPROVED FOR LIVE or NOT APPROVED with specific reason
  - [ ] Algo reviewer acknowledgement noted in the brief
- **Notes:** Aggregates outputs from TASK-0049 through TASK-0055. Cannot be done in isolation. No strategy goes live without this document. The live period (2025 onward) is the true holdout — the walk-forward OOS windows are the proxy for out-of-sample evidence. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0046] Engine — session-boundary support for intraday backtesting

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** Phase 2 sequencing — intraday engine work starts after Phase 1 daily-bar evaluation pipeline is complete (TASK-0049–0056). Methodology questions resolved: Marcus answered both in session 2026-04-25 (Decision 2026-04.3.0: forced-close at 3:15 PM bar Close; Decision 2026-04.3.1: session detection via IST timestamp ≥ 15:15). Ready to build when Phase 2 begins.
- **Context:** The engine event loop has no concept of a trading session. For intraday (MIS) strategies, any open position must be closed by 3:15 PM IST or Zerodha auto-squares it at a random market price. Without this logic, intraday backtests are invalid.
- **Acceptance criteria:**
  - [ ] `SessionConfig` struct added: `Exchange string`, `Timezone *time.Location`, `SessionEndTime time.Time` (local time-of-day)
  - [ ] `engine.Config` gains optional `Session *SessionConfig` (nil = no session boundary, current behavior preserved)
  - [ ] `isLastBarOfSession(bar model.Candle, cfg *SessionConfig) bool` helper in `internal/engine/`
  - [ ] Event loop: after applying pending signal, if `isLastBarOfSession` returns true and a position is open, force-close at the configured fill price
  - [ ] Golden test: 2-day intraday candle series with position open at session end → forced close on day 1, correct equity and trade log
  - [ ] Timezone-aware tests covering IST session boundaries
  - [ ] Tests written before implementation (TDD)
- **Notes:** This is a significant engine change. Golden tests are mandatory for any event loop modification per repo standards. The `Session *SessionConfig` being optional (nil pointer) preserves all existing daily-bar tests without modification.

---

### [TASK-0048] Tooling — weekly kill-switch monitor (`cmd/monitor`)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** Trade log file format decision required. The engine currently outputs `model.Trade` structs to JSON via `cmd/backtest --out`. A weekly monitor needs a file format for live trades that accumulates across sessions. Decision needed: reuse the existing JSON format, or define a separate append-only CSV schema for live trade records.
- **Context:** Marcus specified weekly kill-switch monitoring (not quarterly re-validation). The existing `analytics.CheckKillSwitch()` and `DeriveKillSwitchThresholds()` functions are ready. This task wires them into a runnable binary that reads live trade history and thresholds, outputs alert status, and can be cron-scheduled.
- **Acceptance criteria:**
  - [ ] Trade log file format decided and documented (decision record in `decisions/`)
  - [ ] `cmd/monitor/main.go` reads: (1) live trade log in the agreed format, (2) kill-switch thresholds JSON produced by `DeriveKillSwitchThresholds`
  - [ ] Calls `analytics.CheckKillSwitch(recentTrades, liveCurve, thresholds)`
  - [ ] Prints clear alert status: `OK`, `HALT (Sharpe breached)`, `HALT (drawdown breached)`, `HALT (duration breached)`
  - [ ] Exit code 0 if OK, non-zero if any threshold breached (enables shell scripting / cron alerting)
  - [ ] Tests: known trade sequence crossing each threshold type → correct alert output and exit code

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

### [TASK-0057] Engine — migrate accounting layer from float64 to shopspring/decimal

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-25
- **Source:** decision
- **Context:** The commission arithmetic in `commission.go` and the broader accounting layer (Portfolio.cash, Trade.RealizedPnL, Trade.Commission, EquityPoint.Value) all use float64. Accumulated rounding errors are negligible for backtesting but not acceptable for live execution accounting. This migration must be coordinated — partial decimal adoption creates a worse inconsistency than uniform float64.
- **Acceptance criteria:**
  - [ ] `shopspring/decimal` added to `go.mod` (requires explicit approval per CLAUDE.md no-new-deps rule — confirm before implementation)
  - [ ] `commission.go` migrated: all intermediate calculations use `decimal.Decimal`; final return values converted to float64 only at the portfolio accounting boundary
  - [ ] `portfolio.go`: `cash` field migrated to `decimal.Decimal`
  - [ ] `pkg/model/trade.go`: `RealizedPnL`, `Commission` fields migrated to `decimal.Decimal`
  - [ ] `pkg/model/equity.go`: `EquityPoint.Value` migrated to `decimal.Decimal`
  - [ ] All existing tests pass with race detector after migration
  - [ ] Golden tests in `commission_zerodha_full_test.go` updated to use exact decimal comparisons
  - [ ] Benchmark (`BenchmarkEngineRun`) remains within 1ms/op budget after migration
- **Notes:** This is a coordinated migration — do not migrate commission.go alone. Deferred from TASK-0038 per decision `2026-04-25-float64-for-commission-arithmetic`. The `shopspring/decimal` dependency must be discussed with the user before implementation per the no-new-dependencies rule in CLAUDE.md. Do not start until that discussion has happened.

---

### [TASK-0037] Rigor — bootstrap re-run to fill kill-switch p5 Sharpe thresholds

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-21
- **Source:** session
- **Context:** TASK-0026 documented drawdown and duration kill-switch thresholds for SMA crossover and RSI mean-reversion, but the bootstrap p5 Sharpe threshold is PENDING for both. The CLI commands are ready; the Zerodha token needs to be refreshed to run them.
- **Acceptance criteria:**
  - [ ] Run `go run ./cmd/backtest --strategy sma-crossover ... --bootstrap` (full command in `decisions/algorithm/2026-04-21-kill-switch-sma-crossover.md`)
  - [ ] Run `go run ./cmd/backtest --strategy rsi-mean-reversion ... --bootstrap` (full command in `decisions/algorithm/2026-04-21-kill-switch-rsi-mean-reversion.md`)
  - [ ] Paste the `Per-trade Sharpe p5` value from each run into the respective decision file, replacing `PENDING`
  - [ ] Update decision file status from `accepted` (PENDING) to reflect actual values
- **Notes:** Both strategies failed the proliferation gate — these thresholds are reference values, not live deployment approval. With only 7 and 22 trades respectively, the p5 Sharpe will have wide confidence intervals. Document that caveat alongside the values.

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
