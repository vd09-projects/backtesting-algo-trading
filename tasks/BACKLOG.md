# Project Task Backlog

**Last updated:** 2026-04-29 | **Open tasks:** 18 | **Next up:** TASK-0051

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

### [TASK-0051] Evaluation — in-sample baseline and parameter sensitivity (all 6 strategies, RELIANCE, 2018-2023)

- **Status:** in-progress
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Context:** Orientation run on a single instrument to understand each strategy's behavior and find the robust parameter region via 1D sweep. Not a gate — the universe sweep is the gate. The output of this task is the plateau-midpoint parameter for each strategy, which is what gets used in the universe sweep.
- **Acceptance criteria:**
  - [x] Extend `cmd/backtest` and `cmd/sweep` to support all 6 strategies with `--commission` flag (done 2026-04-29; `ParseCommissionModel` extracted to `internal/cmdutil`)
  - [x] `internal/sweep.computePlateau` updated to apply 80% floor against valid-region peak (TradeCount ≥ 30) per Marcus's verdict; `SensitivityConcern` field added to `Report`
  - [ ] Run `cmd/backtest` for all 6 strategies on NSE:RELIANCE 2018-01-01 to 2024-01-01 with `--commission zerodha_full` and default parameters
  - [ ] Record for each: Sharpe, regime splits (3 NSE windows), trade count, max drawdown, bootstrap p5 Sharpe
  - [ ] Run 1D parameter sweep (`cmd/sweep`) for each strategy on its key parameter (rsi-period, sma-fast/slow, macd-fast/slow, bb-period, donchian-period, momentum-lookback)
  - [ ] Identify plateau range (within 80% of peak Sharpe in ≥30-trade valid region) for each strategy
  - [ ] Select plateau-midpoint parameter for each strategy for use in universe sweep; if no valid plateau exists, flag strategy as "sensitivity concern" and use defaults for universe sweep
  - [ ] Results saved to `runs/baseline-2026-04-29/` with JSON and CSV outputs; `plateau-params.json` produced
- **Notes:** Tooling gate complete as of 2026-04-29. Remaining work: execute the CLI runs (requires live Zerodha token). Use `--commission zerodha_full --bootstrap` for baseline runs. Plateau logic now applies ≥30 trade filter per Marcus's verdict. Owner: Marcus (algo-trading-veteran).

---

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

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

### [TASK-0062] Tooling — NIFTY 50 TRI CSV loader for benchmark comparison

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-28
- **Source:** decision
- **Context:** TASK-0045 confirmed that NIFTY 50 TRI is not available via Zerodha Kite Connect. Decision `2026-04-28-nifty-tri-benchmark-data-source.md` chose Option A: load TRI from NSE-published CSV. This task implements the loader so benchmark comparisons use TRI (with dividend reinvestment) rather than the price index, which understates buy-and-hold by ~8–10% over 6 years.
- **Acceptance criteria:**
  - [ ] Download NIFTY 50 TRI historical CSV from NSE (`nseindia.com/products/content/equities/indices/historical_total_returns.htm`) covering 2015-01-01 to present; store at `data/benchmarks/nifty50-tri.csv`
  - [ ] `LoadNSETRICSV(path string) ([]model.Candle, error)` function implemented — parses the NSE CSV, returns daily candles with `Instrument: "NSE:NIFTY50-TRI"` and Close set to the TRI value; Open/High/Low set equal to Close (TRI is close-only)
  - [ ] Loader placed in `internal/analytics` alongside `ComputeBenchmark`, or a dedicated `internal/benchmark` package if benchmark logic is being extracted — decide at build time
  - [ ] A sample TRI fixture CSV (20 rows minimum, covering a verifiable date range) added to testdata
  - [ ] `TestLoadNSETRICSV` covers: happy path, missing file, malformed row (skipped vs. error), empty file
  - [ ] `ComputeBenchmark` updated to accept the TRI candle series instead of fetching `NSE:NIFTY 50` from Zerodha when a TRI path is provided
  - [ ] Tests written before implementation (TDD)
- **Notes:** Per decision, TRI series is stable historical data — no live fetch needed. Manual update cadence: re-download from NSE before each full evaluation run that extends beyond the last downloaded date. The price index (`NSE:NIFTY 50`) remains the default when no TRI path is provided, for backwards compatibility. Blocked by nothing — can start immediately.

---

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

### [TASK-0060] Tooling — `--commission` flag for `cmd/universe-sweep`

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-27
- **Source:** discovery
- **Context:** `CommissionZerodhaFull` and `CommissionZerodhaFullMIS` exist in `pkg/model/order.go`. `cmd/backtest` and `cmd/sweep` received the `--commission` flag in TASK-0051 (2026-04-29). `cmd/universe-sweep` still hardcodes `CommissionZerodha` and needs the same treatment. `ParseCommissionModel` is already in `internal/cmdutil` — wiring it in is a small change.
- **Acceptance criteria:**
  - [x] `--commission` flag added to `cmd/backtest` (done 2026-04-29, TASK-0051)
  - [x] `--commission` flag added to `cmd/sweep` (done 2026-04-29, TASK-0051)
  - [x] `ParseCommissionModel` extracted to `internal/cmdutil` (done 2026-04-29, TASK-0051)
  - [ ] `--commission` flag added to `cmd/universe-sweep`; wire `cmdutil.ParseCommissionModel` into `model.OrderConfig.CommissionModel`
  - [ ] Invalid value returns a clear error at startup, not a silent fallback
  - [ ] Help text documents the accepted values
- **Notes:** `cmd/backtest` and `cmd/sweep` done in TASK-0051. Only `cmd/universe-sweep` remains. Required before any MIS intraday backtest run (Phase 2+). Low friction to implement now that `ParseCommissionModel` is shared.

---

### [TASK-0058] Tooling — fix cyclomatic complexity in `cmd/rsi-diagnostic/main.go`

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-27
- **Source:** discovery
- **Context:** `cmd/rsi-diagnostic/main.go` `main()` function has cyclomatic complexity 17, exceeding the project's golangci-lint cyclop limit of 15. Discovered during TASK-0043 build session — the file was pre-existing, not introduced by TASK-0043. The fix pattern is established: extract strategy-dispatch and parameter-parsing logic into named helper functions, matching the refactor applied to `cmd/sweep/main.go` in TASK-0043 (smaFactory, rsiFactory, donchianFactory extraction).
- **Acceptance criteria:**
  - [ ] `golangci-lint run ./cmd/rsi-diagnostic/...` reports 0 issues
  - [ ] `go1.25.0 test -race ./...` still passes
  - [ ] No behavioral changes — refactor only
- **Notes:** The same cyclop issue does NOT exist in cmd/backtest or cmd/sweep after TASK-0043 refactored sweep's factoryRegistry. rsi-diagnostic is the only remaining offender.

---

### [TASK-0059] Engine — walk-forward `Run()` factory API for stateful strategy wrappers

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-27
- **Source:** session
- **Context:** `TimedExit` (added in TASK-0039) is the first stateful `Strategy` implementation — it tracks `entryBar` and `inPosition` between `Next()` calls. The walk-forward harness (`internal/walkforward`) currently accepts a single `strategy.Strategy` instance, documented as safe only when the strategy is stateless. Using `TimedExit` with walk-forward today would silently produce corrupted results across fold boundaries because the shared state is never reset between folds.
- **Acceptance criteria:**
  - [ ] `internal/walkforward.Run()` signature changed to accept `factory func() strategy.Strategy` instead of a single `strategy.Strategy` instance
  - [ ] Each fold constructs a fresh strategy instance via `factory()` — no shared state across folds
  - [ ] Existing callers updated: stateless strategies pass `func() strategy.Strategy { return myStrategy }` closures
  - [ ] All 17 existing walk-forward tests still pass with race detector
  - [ ] New test: `TimedExit`-wrapped strategy used in walk-forward — verify fold 2 starts with clean position state
  - [ ] Godoc on `Run()` updated to remove the concurrent-safety caveat (factory eliminates the concern)
- **Notes:** Triggered by `2026-04-27-timed-exit-statefulness-pkg-strategy` decision. The `decisions/tradeoff/2026-04-22-walkforward-strategy-single-instance.md` revisit trigger fires here: "If stateful strategies are added, the signature changes to `func() strategy.Strategy`." This is that moment. Breaking API change — scan all callers in `cmd/` before implementing.

---

### [TASK-0061] Tooling — extend `cmd/sweep2d` factoryRegistry to all 6 strategies

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-27
- **Source:** session
- **Context:** `cmd/sweep2d/main.go` was built in TASK-0044 with `sma-crossover` and `rsi-mean-reversion` only ("extend as new strategies land"). The remaining four strategies (donchian-breakout, macd-crossover, bollinger-mean-reversion, momentum) need 2D axis mappings added to `factoryRegistry2D`. The `fixedParams` struct is also duplicated between `cmd/sweep` and `cmd/sweep2d` — each new strategy requires updating both files. Consider extracting to `internal/cmdutil` or a shared cmd-layer type at this point.
- **Acceptance criteria:**
  - [ ] `factoryRegistry2D` in `cmd/sweep2d/main.go` handles all 6 strategies
  - [ ] Axis mappings documented in code comments: donchian (p1=period, p2=tbd), macd (p1=fast, p2=slow), bollinger (p1=period, p2=num-std-dev), momentum (p1=lookback, p2=threshold)
  - [ ] `fixedParams` struct duplication between `cmd/sweep` and `cmd/sweep2d` resolved — either extracted to shared location or duplication accepted with a comment
  - [ ] All new factory paths covered by `TestFactoryRegistry2D_KnownStrategies`
  - [ ] `golangci-lint run ./cmd/sweep2d/...` still passes
- **Notes:** Donchian has only one meaningful sweep parameter (period) — its p2 axis is less obvious; defer the axis mapping decision until this task is picked up. The `fixedParams` duplication is a low-friction issue for now (2 files to update per new strategy) but compounds at 6 strategies.

---

### [TASK-0062] Tooling — NIFTY 50 TRI benchmark: download CSV and implement StaticCSVProvider

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-28
- **Source:** decision
- **Context:** TASK-0045 (research spike) confirmed NIFTY 50 TRI is not available via Zerodha Kite Connect. Decision `2026-04-28-nifty-tri-benchmark-data-source.md` chose Option A: NSE-published CSV loader. This task implements that decision — download the authoritative TRI CSV from NSE and build a minimal `StaticCSVProvider` so the benchmark computation path is provider-agnostic.
- **Acceptance criteria:**
  - [ ] `data/benchmarks/nifty50-tri.csv` downloaded from NSE (nseindia.com/products/content/equities/indices/historical_total_returns.htm) covering 2015-01-01 to present; committed to repo
  - [ ] `pkg/provider/csv/` package created with `StaticCSVProvider` implementing `provider.DataProvider` for a single instrument (daily timeframe only)
  - [ ] `StaticCSVProvider` returns `ErrUnsupportedTimeframe` for non-daily timeframes and `ErrInstrumentNotFound` for instruments not in the loaded file
  - [ ] `BenchmarkReport` computation wired to use `StaticCSVProvider` for the TRI benchmark when `--benchmark-tri` flag is set (or equivalent)
  - [ ] Tests written before implementation (TDD); `go1.25.0 test -race ./pkg/provider/csv/...` passes
  - [ ] `golangci-lint run ./pkg/provider/csv/...` passes
- **Notes:** `StaticCSVProvider` should satisfy `provider.DataProvider` at compile time via a `var _ provider.DataProvider = (*StaticCSVProvider)(nil)` check. NSE CSV columns: Date, Open, High, Low, Close (or just Index Value for TRI — inspect the actual download first). TRI values will be in the 9,000–28,000 range for 2015–2024. No chunking, no auth, no rate limits needed. Follow-up to decision `2026-04-28-nifty-tri-benchmark-data-source.md`.

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

### [TASK-0063] Tooling — update `cmd/backtest` package doc comment to list all 6 strategies

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-29
- **Source:** discovery
- **Context:** The package-level doc comment in `cmd/backtest/main.go` lists only `stub`, `sma-crossover`, and `rsi-mean-reversion` under "Available strategies". The `strategyRegistry` and `--strategy` flag help text now correctly list all 6, but the doc comment at the top of the file is stale and would mislead someone reading the source. Discovered during TASK-0051 quality review.
- **Acceptance criteria:**
  - [ ] `cmd/backtest/main.go` package doc comment "Available strategies" section updated to list all 6 strategies with their flag descriptions
  - [ ] `golangci-lint run ./cmd/backtest/...` still passes
- **Notes:** Pure documentation change — no logic, no tests needed. Low priority; do alongside any other `cmd/backtest` touch.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
