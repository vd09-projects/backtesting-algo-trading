# Decision Index

<!-- 
  This file is maintained by the decision-journal skill.
  Entries are in YAML format for machine-friendly querying.
  Newest entries go at the top. Do not manually reorder.
-->

```yaml
decisions:
  - id: 2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity
    title: "TASK-0051 routing: MACD proceeds on defaults; other 5 require parameter sensitivity pass first"
    date: 2026-05-01
    status: experimental
    category: algorithm
    tags: [TASK-0051, TASK-0052, signal-audit, routing, parameter-sensitivity, macd, 30-trade-floor, kill-gate, Marcus]
    path: algorithm/2026-05-01-task0051-signal-audit-routing-macd-proceeds-others-need-sensitivity.md
    summary: "MACD (44–65 trades/instrument at defaults) proceeds to TASK-0051 directly. SMA, RSI, Donchian, Bollinger, Momentum must find a ≥30-trade parameter region via sensitivity sweep before advancing. Strategies with no valid parameter at any sweep value are killed here, not held for pooled analysis."

  - id: 2026-05-01-signalaudit-strategy-factory-decoupling
    title: "signalaudit package uses StrategyFactory — no import of concrete strategy packages"
    date: 2026-05-01
    status: experimental
    category: architecture
    tags: [signalaudit, strategy-factory, decoupling, cmd-layer, dependency-direction, TASK-0050]
    path: architecture/2026-05-01-signalaudit-strategy-factory-decoupling.md
    summary: "internal/signalaudit defines a StrategyFactory type; the cmd layer owns the factory closure that imports concrete strategy packages. Keeps internal/signalaudit free of strategies/ dependency and matches the pattern in internal/sweep and internal/walkforward."

  - id: 2026-05-01-signalaudit-sequential-strategies-parallel-instruments
    title: "signalaudit: strategies run sequentially, instruments fan out in parallel"
    date: 2026-05-01
    status: experimental
    category: tradeoff
    tags: [signalaudit, concurrency, errgroup, zerodha-api, rate-limit, TASK-0050]
    path: tradeoff/2026-05-01-signalaudit-sequential-strategies-parallel-instruments.md
    summary: "6 strategies run sequentially; within each, 15 instruments fan out in parallel via errgroup. Full parallelism (90 concurrent runs) would overwhelm Zerodha rate limits. Sequential strategies give full hardware utilisation on the inner loop without multiplying API pressure by 6."

  - id: 2026-05-01-signal-audit-no-per-strategy-parameter-flags
    title: "cmd/signal-audit hardcodes strategy defaults — no per-strategy parameter override flags"
    date: 2026-05-01
    status: experimental
    category: convention
    tags: [signal-audit, cmd, flags, default-params, audit-purpose, TASK-0050]
    path: convention/2026-05-01-signal-audit-no-per-strategy-parameter-flags.md
    summary: "cmd/signal-audit accepts no strategy-specific parameter flags. Its sole purpose is auditing default-parameter trade frequency. cmd/universe-sweep and cmd/sweep serve the parameter-explorer role. Avoids 13-flag proliferation and keeps the binary's responsibility narrow."

  - id: 2026-04-29-parse-commission-model-extracted-to-cmdutil
    title: "ParseCommissionModel extracted to internal/cmdutil — shared flag-parsing helper"
    date: 2026-04-29
    status: experimental
    category: convention
    tags: [commission, DRY, cmdutil, flag-parsing, cmd/backtest, cmd/sweep, cmd/universe-sweep, TASK-0051, TASK-0060]
    path: convention/2026-04-29-parse-commission-model-extracted-to-cmdutil.md
    summary: "ParseCommissionModel(s string) (model.CommissionModel, error) extracted to internal/cmdutil rather than duplicated in cmd/backtest and cmd/sweep. Third caller (cmd/universe-sweep, TASK-0060) is imminent — extraction at two callers accepted given near-zero cost and established cmdutil precedent."

  - id: 2026-04-29-fallback-to-defaults-no-valid-plateau
    title: "Fallback to strategy defaults for universe sweep when no valid plateau exists"
    date: 2026-04-29
    status: experimental
    category: algorithm
    tags: [sweep, plateau, universe-sweep, sensitivity-concern, parameter-selection, overfitting, TASK-0051]
    path: algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md
    summary: "When the valid region (TradeCount >= 30) of a parameter sweep is empty or all-negative Sharpe, use strategy defaults for the universe sweep — not the least-bad parameter. Selecting the least-bad parameter on RELIANCE would overfit losses on one instrument. Defaults carry no RELIANCE-specific bias."

  - id: 2026-04-29-plateau-procedure-trade-count-constrained
    title: "Plateau procedure for trade-count-constrained strategies: apply 80% floor against valid-region peak"
    date: 2026-04-29
    status: experimental
    category: algorithm
    tags: [sweep, plateau, trade-count, valid-region, sensitivity-concern, parameter-sensitivity, TASK-0051]
    path: algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md
    summary: "The 80% Sharpe floor for plateau identification is applied against the peak Sharpe within the valid region (TradeCount >= MinTradesForPlateau=30), not the global peak. A global peak from a 5-trade parameter would set an unreliable bar. If valid region is empty or all-negative, Report.SensitivityConcern is set and Report.Plateau is nil."

  - id: 2026-04-28-nifty-tri-benchmark-data-source
    title: "NIFTY 50 TRI benchmark — not available via Zerodha Kite; use NSE-published CSV"
    date: 2026-04-28
    status: accepted
    category: infrastructure
    tags: [benchmark, NIFTY-TRI, total-return, data-source, zerodha, NSE, kite-connect, TASK-0045]
    path: infrastructure/2026-04-28-nifty-tri-benchmark-data-source.md
    summary: "NIFTY 50 TRI is not in Kite instruments master — only price return index (token 256265) is available. TR-named instruments (NIFTY50 TR 2X LEV, NIFTY50 TR 1X INV) are leveraged/inverse strategy indices, not the TRI benchmark. Decision: use NSE-published CSV (Option A) rather than a second DataProvider. Implementation deferred to build task."

  - id: 2026-04-27-sweep2d-csv-writer-io-writer-helper
    title: "sweep2d CSV writer via io.Writer helper (writeCSVToWriter)"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [sweep2d, io.Writer, csv, testability, cmd/sweep2d, TASK-0044]
    path: convention/2026-04-27-sweep2d-csv-writer-io-writer-helper.md
    summary: "writeCSVToWriter(io.Writer) helper in cmd/sweep2d serializes the Sharpe matrix to any writer, enabling smoke tests without temp files. Production path routes to os.Stdout (no --out) or sweep2d.WriteCSV (--out set). Follows output.Config.Stdout injectable-writer pattern."

  - id: 2026-04-27-flags2d-value-struct-flag-parsing-testability
    title: "flags2D value struct for flag parsing testability in cmd/sweep2d"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [cmd/sweep2d, flags, testability, value-struct, TASK-0044]
    path: convention/2026-04-27-flags2d-value-struct-flag-parsing-testability.md
    summary: "Parsed flag values bundled into a flags2D struct passed to parseAndValidateFlags2D. Tests construct the struct directly — no flag.FlagSet construction needed. Avoids the 10-parameter function smell and keeps the signature stable as flags are added."

  - id: 2026-04-27-sweep2d-stdout-fallback-bytes-buffer
    title: "sweep2d stdout fallback via writeCSVToWriter io.Writer helper"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [sweep2d, stdout, bytes.Buffer, io.Writer, testability, cmd/sweep2d, TASK-0044]
    path: convention/2026-04-27-sweep2d-stdout-fallback-bytes-buffer.md
    summary: "When --out is omitted, cmd/sweep2d writes CSV via writeCSVToWriter to os.Stdout rather than a temp-file-then-copy approach. Keeps the stdout path testable via bytes.Buffer. Minor format duplication with sweep2d.WriteCSV accepted; cross-reference comment enforces sync."

  - id: 2026-04-27-sweep2d-sma-crossover-axis-p1-fast-p2-slow
    title: "sweep2d sma-crossover axis mapping: p1=fast-period, p2=slow-period"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [sweep2d, sma-crossover, axis-mapping, cmd/sweep2d, TASK-0044]
    path: convention/2026-04-27-sweep2d-sma-crossover-axis-p1-fast-p2-slow.md
    summary: "For cmd/sweep2d, sma-crossover always maps p1→fast-period and p2→slow-period. Fixed by convention in factoryRegistry2D; --p1-name/--p2-name are cosmetic labels only, not axis selectors. Natural fast×slow grid orientation. Revisit if axis-swap CLI control is needed."

  - id: 2026-04-27-mis-stt-constant-placement-commission-go
    title: "MIS STT rate co-located with CNC STT rate in commission.go constant block"
    date: 2026-04-27
    status: experimental
    category: architecture
    tags: [commission, MIS, STT, constants, commission.go, TASK-0047]
    path: architecture/2026-04-27-mis-stt-constant-placement-commission-go.md
    summary: "nseMISSTTRate placed alongside nseSTTRate in the same const block, grouped by charge type (not model variant). Makes CNC vs MIS STT comparison trivial; consistent with the established pattern of grouping NSE statutory rates by charge category."

  - id: 2026-04-27-timed-exit-statefulness-pkg-strategy
    title: "TimedExit statefulness — first stateful wrapper in pkg/strategy, not concurrent-safe"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [pkg/strategy, TimedExit, statefulness, concurrency, walk-forward, TASK-0039, TASK-0059]
    path: convention/2026-04-27-timed-exit-statefulness-pkg-strategy.md
    summary: "TimedExit tracks entryBar and inPosition between Next() calls — the first stateful Strategy in pkg/strategy. A single instance must not be shared across concurrent walk-forward folds; TASK-0059 tracks the factory API change to internal/walkforward that enforces this automatically."

  - id: 2026-04-27-momentum-signal-level-comparison-not-crossover
    title: "Momentum strategy signal semantics: level-comparison, not crossover"
    date: 2026-04-27
    status: experimental
    category: convention
    tags: [momentum, signal-semantics, level-comparison, crossover, roc, TASK-0043]
    path: convention/2026-04-27-momentum-signal-level-comparison-not-crossover.md
    summary: "ROC threshold comparison is a single-indicator state assertion, not a line-crossing event. Level-comparison (Buy every bar where ROC > threshold) is semantically correct; strict crossover would add tracking state without adding meaning. Under no-pyramiding, practical behaviour is identical."

  - id: 2026-04-26-workflow-gate-pre-tool-use-hook
    title: "Workflow gate: PreToolUse hook blocks production Go writes when session-state absent"
    date: 2026-04-26
    status: accepted
    category: infrastructure
    tags: [workflow, hook, pre-tool-use, session-state, build.md, TASK-0041]
    path: infrastructure/2026-04-26-workflow-gate-pre-tool-use-hook.md
    summary: "Check 3 added to pre-tool-use-edit.sh blocks Edit/Write to *.go files under strategies/, internal/, pkg/, cmd/ unless workflows/.session-state.json exists. Enforces build.md Steps 1-4 before any implementation; pipe-tested in all three cases."

  - id: 2026-04-26-macd-guard-talib-initialization-boundary
    title: "MACD guard condition: n <= slow+signal-1 blocks talib zero-fill initialization boundary"
    date: 2026-04-26
    status: accepted
    category: convention
    tags: [macd, talib, initialization, guard, crossover, TASK-0041]
    path: convention/2026-04-26-macd-guard-talib-initialization-boundary.md
    summary: "talib.Macd zero-fills uninitialized output positions with 0. Guard if n <= slow+signal-1 prevents a spurious 0-to-nonzero crossover firing on the first real computation bar. Crossover detection starts at n = slow+signal where both current and previous values are real."

  - id: 2026-04-25-nse-charge-rates-as-unexported-constants
    title: "NSE charge rates as unexported constants in commission.go, not as OrderConfig fields"
    date: 2026-04-25
    status: experimental
    category: architecture
    tags: [commission, NSE, constants, OrderConfig, TASK-0038]
    path: architecture/2026-04-25-nse-charge-rates-as-unexported-constants.md
    summary: "STT, exchange charges, SEBI charges, stamp duty rates, and GST rate are unexported package-level constants in commission.go. These are statutory facts, not user-configurable parameters — putting them in OrderConfig would imply they are variable inputs."

  - id: 2026-04-25-calcommission-isbuy-bool-not-enum
    title: "calcCommission side parameter: isBuy bool, not an enum"
    date: 2026-04-25
    status: experimental
    category: convention
    tags: [commission, calcCommission, bool, enum, TASK-0038]
    path: convention/2026-04-25-calcommission-isbuy-bool-not-enum.md
    summary: "isBuy bool chosen over an OrderSide enum for the unexported calcCommission helper. Two states, two call sites, no cross-package visibility. Migrate to enum if the function becomes exported or gains a third call site with a different side semantic."

  - id: 2026-04-25-commission-logic-extracted-to-commission-go
    title: "Full-model commission logic extracted to commission.go, method stays on *Portfolio"
    date: 2026-04-25
    status: experimental
    category: architecture
    tags: [commission, file-organization, portfolio.go, commission.go, TASK-0038]
    path: architecture/2026-04-25-commission-logic-extracted-to-commission-go.md
    summary: "calcZerodhaFullCommission is a pure helper in commission.go (same engine package). calcCommission method stays on *Portfolio in portfolio.go as the dispatch point. Groups commission math for readability without adding a new package boundary."

  - id: 2026-04-25-gst-base-brokerage-plus-exchange-charges-only
    title: "GST base = brokerage + exchange charges only (STT, SEBI, stamp exempt)"
    date: 2026-04-25
    status: accepted
    category: convention
    tags: [commission, GST, NSE, SEBI, STT, stamp-duty, TASK-0038]
    path: convention/2026-04-25-gst-base-brokerage-plus-exchange-charges-only.md
    summary: "GST (18%) applies to brokerage + NSE exchange transaction charges only. STT, SEBI charges, and stamp duty are excluded. Confirmed against Indian tax regulation and Zerodha's brokerage calculator (used to hand-verify ₹88.24 round-trip golden test)."

  - id: 2026-04-25-float64-for-commission-arithmetic
    title: "float64 for commission arithmetic, not decimal — migration deferred"
    date: 2026-04-25
    status: experimental
    category: tradeoff
    tags: [commission, float64, decimal, shopspring, arithmetic, TASK-0038, TASK-0057]
    path: tradeoff/2026-04-25-float64-for-commission-arithmetic.md
    summary: "float64 used in commission.go consistent with the existing engine accounting layer. Partial decimal adoption (commission only) would be worse than uniform float64. Full migration deferred to TASK-0057 — coordinated change across commission.go, portfolio.go, Trade, EquityPoint."

  - id: 2026-04-25-cross-instrument-proliferation-gate
    title: "Cross-instrument universe gate supersedes single-instrument proliferation gate"
    date: 2026-04-25
    status: experimental
    category: algorithm
    tags: [proliferation-gate, cross-instrument, universe-sweep, DSR, evaluation-methodology, TASK-0049, TASK-0052]
    path: algorithm/2026-04-25-cross-instrument-proliferation-gate.md
    summary: "Single-instrument Sharpe ≥ 0.5 gate replaced by: DSR-corrected average Sharpe > 0 across 15 Nifty50 large-caps AND ≥ 40% of instruments show positive Sharpe with ≥ 30 trades. Supersedes 2026-04-10-strategy-proliferation-gate."

  - id: 2026-04-25-2025-live-data-as-true-holdout
    title: "2025 live trading is the true holdout; no historical data reserved"
    date: 2026-04-25
    status: experimental
    category: algorithm
    tags: [holdout, walk-forward, OOS, 2025-live, live-validation, evaluation-methodology]
    path: algorithm/2026-04-25-2025-live-data-as-true-holdout.md
    summary: "Walk-forward OOS windows (2020-2024 across rolling folds) are the out-of-sample evidence. 2025 live trading at ₹3 lakh is the true holdout. No historical data reserved as a separate holdout split."

  - id: 2026-04-25-intraday-session-detection-ist-cutoff
    title: "Intraday session-close detection via IST timestamp ≥ 15:15, not bar count"
    date: 2026-04-25
    status: experimental
    category: algorithm
    tags: [intraday, session-boundary, timezone, IST, TASK-0046]
    path: algorithm/2026-04-25-intraday-session-detection-ist-cutoff.md
    summary: "Session close detected by comparing bar timestamp (in IST) against a configurable cutoff of 15:15. Works for any bar frequency (5min, 15min) without per-timeframe config. Parameterized via SessionConfig.SessionCutoff."

  - id: 2026-04-25-intraday-forced-close-fill-price
    title: "Intraday forced-close fill price: 3:15 PM bar Close"
    date: 2026-04-25
    status: experimental
    category: algorithm
    tags: [intraday, session-boundary, fill-model, MIS, TASK-0046]
    path: algorithm/2026-04-25-intraday-forced-close-fill-price.md
    summary: "MIS positions force-closed at 3:15 PM bar's Close price. Conservative approximation — Zerodha auto-squareoff fills at or slightly worse than bid, so real performance will be at this price or marginally below. Backtest does not flatter live P&L."

  - id: 2026-04-22-buildprovider-extracted-to-cmdutil
    title: "`buildProvider` extracted from cmd binaries into `internal/cmdutil.BuildProvider`"
    date: 2026-04-22
    status: experimental
    category: architecture
    tags: [provider, DRY, cmd, zerodha, cmdutil, refactor, TASK-0035]
    path: architecture/2026-04-22-buildprovider-extracted-to-cmdutil.md
    summary: "buildProvider was duplicated in cmd/backtest and cmd/sweep. Adding cmd/universe-sweep would have made a third copy — the DRY threshold. Extracted to cmdutil.BuildProvider; all three cmd binaries delegate to it."

  - id: 2026-04-22-pre-allocated-fixed-index-writes-for-determinism
    title: "Pre-allocated fixed-index slice writes for deterministic goroutine output"
    date: 2026-04-22
    status: experimental
    category: convention
    tags: [determinism, goroutine, slice, concurrency, universesweep, TASK-0035]
    path: convention/2026-04-22-pre-allocated-fixed-index-writes-for-determinism.md
    summary: "Goroutine i writes to results[i] on a pre-allocated slice. No mutex needed; same instruments → same pre-sort order → same CSV. Pattern established in internal/walkforward, now applied in internal/universesweep."

  - id: 2026-04-22-errgroup-universe-fan-out-gomaxprocs-ceiling
    title: "`errgroup` with GOMAXPROCS ceiling for universe instrument fan-out"
    date: 2026-04-22
    status: experimental
    category: tradeoff
    tags: [concurrency, errgroup, parallelism, GOMAXPROCS, universesweep, TASK-0035]
    path: tradeoff/2026-04-22-errgroup-universe-fan-out-gomaxprocs-ceiling.md
    summary: "errgroup with GOMAXPROCS ceiling for parallel instrument runs. Unlike parameter sweep (sequential — no x/sync, ordering deps), instrument runs are fully independent and x/sync is already in go.mod. Ceiling prevents memory blow-up on large universes."

  - id: 2026-04-22-universe-file-yaml-format
    title: "Universe file uses YAML with top-level `instruments:` key"
    date: 2026-04-22
    status: experimental
    category: convention
    tags: [YAML, universe-file, file-format, universesweep, TASK-0035]
    path: convention/2026-04-22-universe-file-yaml-format.md
    summary: "YAML with top-level instruments: key chosen over plain text (no extensibility) and bare root sequence (fragile to adding top-level fields). gopkg.in/yaml.v3 already in go.mod."

  - id: 2026-04-22-universe-sweep-csv-schema
    title: "Universe-sweep CSV schema: 6 columns, no rank column"
    date: 2026-04-22
    status: experimental
    category: convention
    tags: [CSV, output, schema, universesweep, TASK-0035]
    path: convention/2026-04-22-universe-sweep-csv-schema.md
    summary: "Six columns: instrument, sharpe, trade_count, total_pnl, max_drawdown, insufficient_data. No rank column — row position is rank. insufficient_data is bool OR of TradeMetricsInsufficient||CurveMetricsInsufficient from analytics.Report."

  - id: 2026-04-22-universe-sweep-runner-placement
    title: "Universe-sweep runner lives in `internal/universesweep`, not inline in `cmd/`"
    date: 2026-04-22
    status: experimental
    category: architecture
    tags: [package-boundaries, cmd, internal, universesweep, TASK-0035]
    path: architecture/2026-04-22-universe-sweep-runner-placement.md
    summary: "ParseUniverseFile, Run, WriteCSV live in internal/universesweep (testable package), not main.go. Follows the cmd/sweep vs internal/sweep precedent. Acceptance criterion test lives in the package."

  - id: 2026-04-22-walkforward-strategy-single-instance
    title: "Walk-forward accepts a single strategy instance, not a factory"
    date: 2026-04-22
    status: experimental
    category: tradeoff
    tags: [walkforward, strategy, concurrency, API, factory, TASK-0022]
    path: tradeoff/2026-04-22-walkforward-strategy-single-instance.md
    summary: "Run() takes a single strategy.Strategy; all current strategies are stateless so concurrent fold runs are safe. Factory API deferred until a mutable-state strategy is added."

  - id: 2026-04-22-walkforward-to-exclusive-upper-bound
    title: "WalkForwardConfig.To is the exclusive upper bound"
    date: 2026-04-22
    status: experimental
    category: convention
    tags: [walkforward, time, boundary, half-open-interval, API, TASK-0022]
    path: convention/2026-04-22-walkforward-to-exclusive-upper-bound.md
    summary: "WalkForwardConfig.To is exclusive (half-open interval), consistent with engine.Config.To and provider.FetchCandles. Callers use 2025-01-01 to mean through end of 2024."

  - id: 2026-04-22-walkforward-all-degenerate-no-flags
    title: "All-degenerate walk-forward result: both flags false, DeduplicatedFoldCount=0"
    date: 2026-04-22
    status: experimental
    category: tradeoff
    tags: [walkforward, degenerate, scoring, overfitting-gate, TASK-0022]
    path: tradeoff/2026-04-22-walkforward-all-degenerate-no-flags.md
    summary: "When all folds have zero OOS trades, OverfitFlag=false and NegativeFoldFlag=false. 'No trades' is not overfitting. Callers detect a dead strategy via DeduplicatedFoldCount==0."

  - id: 2026-04-22-walkforward-test-fakes-in-package
    title: "In-package test fakes defined in _test.go only"
    date: 2026-04-22
    status: experimental
    category: convention
    tags: [walkforward, test-fake, provider, strategy, _test.go, TASK-0022]
    path: convention/2026-04-22-walkforward-test-fakes-in-package.md
    summary: "staticProvider, toggleStrategy, neverTradeStrategy are unexported and live in walkforward_test.go. Exporting test infrastructure from pkg/provider or pkg/strategy sends dependencies in the wrong direction."

  - id: 2026-04-22-walkforward-imports-internal-engine
    title: "internal/walkforward imports internal/engine (orchestration harness, not stats)"
    date: 2026-04-22
    status: experimental
    category: architecture
    tags: [walkforward, package-boundary, engine-dependency, montecarlo, TASK-0022]
    path: architecture/2026-04-22-walkforward-imports-internal-engine.md
    summary: "internal/walkforward imports internal/engine because it is an orchestration harness that runs the engine across fold windows, not a statistics package that resamples a trade list. The internal/montecarlo narrow-imports analogy does not apply."

  - id: 2026-04-22-walkforward-engine-config-template
    title: "Run() accepts EngineConfigTemplate, not engine.Config directly"
    date: 2026-04-22
    status: experimental
    category: architecture
    tags: [walkforward, engine-config, dependency-injection, config-template, TASK-0022]
    path: architecture/2026-04-22-walkforward-engine-config-template.md
    summary: "Run() takes EngineConfigTemplate (caller-controlled fields) + WalkForwardConfig (fold params). The harness stamps Instrument/From/To per fold. Prevents callers from accidentally setting engine.Config.From thinking it controls the outer window."

  - id: 2026-04-22-walk-forward-oos-is-sharpe-threshold
    title: "Walk-forward OOS/IS Sharpe threshold and fold-level flagging"
    date: 2026-04-22
    status: experimental
    category: algorithm
    tags: [walk-forward, IS-OOS, sharpe, threshold, overfitting-gate, TASK-0022]
    path: algorithm/2026-04-22-walk-forward-oos-is-sharpe-threshold.md
    summary: "Aggregate: OverfitFlag when avg OOS Sharpe < 50% avg IS Sharpe. Secondary: NegativeFoldFlag when 2+ non-degenerate folds have negative OOS Sharpe. Degenerate folds (zero OOS trades) excluded. Sharpe = per-trade non-annualized, consistent with bootstrap."

  - id: 2026-04-22-walk-forward-window-sizing-default
    title: "Walk-forward window sizing defaults (2yr IS / 1yr OOS)"
    date: 2026-04-22
    status: experimental
    category: algorithm
    tags: [walk-forward, window-size, rolling, IS-OOS, regime-coverage, TASK-0022]
    path: algorithm/2026-04-22-walk-forward-window-sizing-default.md
    summary: "Fixed rolling 2yr IS / 1yr OOS / 1yr step. 4 folds over 2018–2024. 2024 held out naturally. Expanding windows rejected (no fitting occurs). Per-fold Sharpe indicative only at 10–20 trades/fold."

  - id: 2026-04-22-walk-forward-purpose-stateless-strategies
    title: "Walk-forward purpose for stateless fixed-parameter strategies"
    date: 2026-04-22
    status: experimental
    category: algorithm
    tags: [walk-forward, stateless, regime-stability, overfitting-defense, TASK-0022]
    path: algorithm/2026-04-22-walk-forward-purpose-stateless-strategies.md
    summary: "Walk-forward on fixed-parameter strategies is a regime-stability test, not a parameter-overfitting test. The IS/OOS comparison reveals whether edge is regime-concentrated, not whether parameters were overfit."

  - id: 2026-04-21-cmd-correlate-new-binary
    title: "`cmd/correlate` as a new binary rather than extending `cmd/backtest`"
    date: 2026-04-21
    status: experimental
    category: architecture
    tags: [CLI, correlation, cmd/correlate, cmd/backtest, TASK-0027]
    path: architecture/2026-04-21-cmd-correlate-new-binary.md
    summary: "New binary at cmd/correlate. Correlation requires multiple pre-computed strategy results — adding multi-strategy input to cmd/backtest would complicate its single-strategy model. --curve name:path flag (repeatable). Thin: parse → load → ComputeMatrix → WriteCorrelationMatrix."

  - id: 2026-04-21-load-curve-csv-in-output-package
    title: "`LoadCurveCSV` co-located with `writeCurveCSV` in `internal/output`"
    date: 2026-04-21
    status: experimental
    category: architecture
    tags: [LoadCurveCSV, csv-reader, package-boundary, internal/output, TASK-0027]
    path: architecture/2026-04-21-load-curve-csv-in-output-package.md
    summary: "LoadCurveCSV lives in internal/output/load.go alongside writeCurveCSV. Schema changes (column names, timestamp format) touch one file. Rejected: separate curveio package — overkill for one function at this stage."

  - id: 2026-04-21-correlation-nan-sentinel-undefined
    title: "`math.NaN()` as sentinel for undefined correlation"
    date: 2026-04-21
    status: experimental
    category: convention
    tags: [NaN, sentinel, pearson, correlation, TASK-0027]
    path: convention/2026-04-21-correlation-nan-sentinel-undefined.md
    summary: "pearson() and stressPearson() return math.NaN() for constant series, <2 points, or empty stress window. Zero is a valid correlation — using it as sentinel conflates 'uncorrelated' with 'undefined'. TooCorrelated never triggered by NaN. Output prints 'n/a'."

  - id: 2026-04-21-correlation-warmup-first-change-heuristic
    title: "Warmup detection by first-change heuristic in `alignAndTrim`"
    date: 2026-04-21
    status: experimental
    category: architecture
    tags: [warmup-detection, correlation, alignAndTrim, TASK-0027]
    path: architecture/2026-04-21-correlation-warmup-first-change-heuristic.md
    summary: "firstActiveIndex scans for curve[i].Value != curve[0].Value. alignAndTrim takes max across both curves. Keeps ComputeCorrelation decoupled from strategy config. Rejected: explicit warmup int — would require threading strategy config into every CSV-based caller."

  - id: 2026-04-21-compute-per-trade-sharpe-formula-duplication
    title: "`computePerTradeSharpe` intentionally duplicates `montecarlo.sampleSharpe`"
    date: 2026-04-21
    status: experimental
    category: convention
    tags: [kill-switch, analytics, montecarlo, formula-duplication, dependency-direction, TASK-0026]
    path: convention/2026-04-21-compute-per-trade-sharpe-formula-duplication.md
    summary: "Three-line mean(r)/std(r) formula duplicated rather than shared via import. Avoids analytics→montecarlo dependency. Cross-references in both files enforce formula identity. Revisit if a third caller emerges."

  - id: 2026-04-21-kill-switch-analytics-to-montecarlo-boundary
    title: "Kill-switch API keeps `internal/analytics` free of `internal/montecarlo`"
    date: 2026-04-21
    status: experimental
    category: architecture
    tags: [kill-switch, analytics, montecarlo, dependency-direction, package-boundary, TASK-0026]
    path: architecture/2026-04-21-kill-switch-analytics-to-montecarlo-boundary.md
    summary: "DeriveKillSwitchThresholds accepts sharpeP5 float64, not BootstrapResult. Caller extracts the field. analytics imports nothing from montecarlo — pure computation layer stays dependency-free of simulation layer."

  - id: 2026-04-21-kill-switch-rsi-mean-reversion
    title: "Kill-switch parameters — RSI Mean-Reversion (NSE:RELIANCE, 2018–2024)"
    date: 2026-04-21
    status: accepted
    category: algorithm
    tags: [kill-switch, rsi-mean-reversion, NSE:RELIANCE, TASK-0026]
    path: algorithm/2026-04-21-kill-switch-rsi-mean-reversion.md
    summary: "MaxDD threshold=26.03% (1.5×17.36%), duration threshold=744 days (2×372). Bootstrap p5 Sharpe pending re-run. Strategy failed gate at 7 trades — kill-switch threshold also unreliable at this sample size."

  - id: 2026-04-21-kill-switch-sma-crossover
    title: "Kill-switch parameters — SMA Crossover (NSE:RELIANCE, 2018–2024)"
    date: 2026-04-21
    status: accepted
    category: algorithm
    tags: [kill-switch, sma-crossover, NSE:RELIANCE, TASK-0026]
    path: algorithm/2026-04-21-kill-switch-sma-crossover.md
    summary: "MaxDD threshold=24.57% (1.5×16.38%), duration threshold=2274 days (2×1137). Bootstrap p5 Sharpe pending re-run. Strategy failed gate."

  - id: 2026-04-21-kill-switch-derivation-methodology
    title: "Kill-switch derivation methodology"
    date: 2026-04-21
    status: accepted
    category: algorithm
    tags: [kill-switch, live-monitoring, sharpe, drawdown, bootstrap, methodology, TASK-0026]
    path: algorithm/2026-04-21-kill-switch-derivation-methodology.md
    summary: "Three thresholds: (1) bootstrap p5 per-trade Sharpe, (2) 1.5× in-sample max drawdown, (3) 2× in-sample max recovery duration. Implementation in internal/analytics/killswitch.go. Per-trade Sharpe must use identical formula to montecarlo bootstrap (no annualization)."

  - id: 2026-04-20-nearest-rank-percentile
    title: "Nearest-rank (floor index) percentile for bootstrap CIs"
    date: 2026-04-20
    status: experimental
    category: tradeoff
    tags: [montecarlo, bootstrap, percentile, nearest-rank, linear-interpolation, p5, p50, p95, TASK-0024]
    path: tradeoff/2026-04-20-nearest-rank-percentile.md
    summary: "Floor-index nearest-rank chosen over linear interpolation. Returns an observed simulation value. At N=10,000 the difference is undetectable; floor is deterministic; coarseness on small N is a known limitation, not a bug."

  - id: 2026-04-20-geometric-compounding-bootstrap-drawdown
    title: "Geometric compounding for per-simulation drawdown in bootstrap"
    date: 2026-04-20
    status: experimental
    category: convention
    tags: [montecarlo, bootstrap, drawdown, geometric-compounding, additive, TASK-0024]
    path: convention/2026-04-20-geometric-compounding-bootstrap-drawdown.md
    summary: "equity *= (1+r), floored at 0. Additive rejected: physically wrong unless position size is fixed in dollar terms. Geometric is more conservative on high-variance sequences — correct direction for a kill-switch estimate."

  - id: 2026-04-20-rand-newpcg-math-rand-v2
    title: "`math/rand/v2` with `rand.NewPCG` for bootstrap PRNG"
    date: 2026-04-20
    status: experimental
    category: tradeoff
    tags: [montecarlo, bootstrap, prng, pcg64, math-rand-v2, determinism, reproducibility, seed, TASK-0024]
    path: tradeoff/2026-04-20-rand-newpcg-math-rand-v2.md
    summary: "PCG64 (math/rand/v2) chosen over v1 (weak LCG, shared state) and crypto/rand (non-deterministic). Same seed → bit-identical results. Seed logged in output header. Required for comparing bootstrap CIs across sweep variants."

  - id: 2026-04-20-internal-montecarlo-package-boundary
    title: "`internal/montecarlo` as a standalone package, isolated from analytics"
    date: 2026-04-20
    status: experimental
    category: architecture
    tags: [montecarlo, bootstrap, package-boundary, internal/montecarlo, internal/analytics, dependency-direction, TASK-0024]
    path: architecture/2026-04-20-internal-montecarlo-package-boundary.md
    summary: "Bootstrap lives in internal/montecarlo, importing only pkg/model. Rejected: internal/analytics (muddies deterministic metrics with simulation logic) and pkg/montecarlo (speculative public API). Diamond dep graph, no cycles."

  - id: 2026-04-20-bootstrap-sharpe-non-annualized-per-trade
    title: "Bootstrap Sharpe: non-annualized per-trade computation"
    date: 2026-04-20
    status: experimental
    category: algorithm
    tags: [bootstrap, montecarlo, sharpe, annualization, per-trade, kill-switch, TASK-0024, TASK-0026]
    path: algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md
    summary: "mean(r)/std(r) on ReturnOnNotional, no annualization. Annualizing a trade-resampled statistic requires block bootstrap on bar-level returns — out of scope. TASK-0026 must use identical computation or the kill-switch threshold is meaningless."

  - id: 2026-04-20-tf-model-timeframe-added-to-computeregimesplits
    title: "`tf model.Timeframe` added to `ComputeRegimeSplits` signature for correct Sharpe annualization"
    date: 2026-04-20
    status: accepted
    category: tradeoff
    tags: [analytics, regime-split, timeframe, annualization, Sharpe, ComputeRegimeSplits, internal/analytics, TASK-0034]
    path: tradeoff/2026-04-20-tf-model-timeframe-added-to-computeregimesplits.md
    summary: "AC specified no-timeframe signature; hardcoding daily rejected (silently wrong for 15min/1min callers); non-annualized Sharpe rejected (incomparable to main report). Added `tf model.Timeframe` as third parameter — deviates from AC by one arg, correct for all callers."

  - id: 2026-04-16-curve-data-wired-through-config-fields-rather
    title: "Equity curve wired into output.Config fields, not a second Write parameter"
    date: 2026-04-16
    status: experimental
    category: tradeoff
    tags: [config-shape, equity-curve, output, internal/output, TASK-0029]
    path: tradeoff/2026-04-16-curve-data-wired-through-config-fields-rather.md
    summary: "Config.CurvePath + Config.Curve chosen over a new Write parameter or a separate WriteCurve function. Extends the existing Config-as-extension-point pattern; all existing callers and tests unchanged."

  - id: 2026-04-16-rsi-mean-reversion-proliferation-gate-failed
    title: "RSI mean-reversion fails proliferation gate — NSE:RELIANCE 2018–2025"
    date: 2026-04-16
    status: accepted
    category: algorithm
    tags: [rsi-mean-reversion, proliferation-gate, sharpe, trade-count, NSE:RELIANCE, TASK-0028, TASK-0020]
    path: algorithm/2026-04-16-rsi-mean-reversion-proliferation-gate-failed.md
    summary: "RSI (14/30/70) on NSE:RELIANCE 2018–2025 fails gate: Sharpe 0.469 and only 7 trades (statistically meaningless). 100% win rate is a red flag, not a green one. TASK-0020 (Bollinger Bands) cancelled."

  - id: 2026-04-16-sma-crossover-proliferation-gate-failed
    title: "SMA Crossover fails proliferation gate — NSE:RELIANCE 2018–2025"
    date: 2026-04-16
    status: accepted
    category: algorithm
    tags: [sma-crossover, proliferation-gate, sharpe, NSE:RELIANCE, TASK-0028, TASK-0019]
    path: algorithm/2026-04-16-sma-crossover-proliferation-gate-failed.md
    summary: "SMA crossover (10/50) on NSE:RELIANCE 2018–2025 fails gate: Sharpe 0.447. 22 trades, ProfitFactor 2.16 but 36% win rate. MaxDrawdownDuration 3.1 years. TASK-0019 (MACD) cancelled."

  - id: 2026-04-15-maxdrawdownduration-computed-from-per-bar-equity
    title: "MaxDrawdownDuration computed from per-bar equity curve, not trade P&L accumulation"
    date: 2026-04-15
    status: superseded
    category: tradeoff
    tags: [drawdown, duration, equity-curve, analytics, MaxDrawdownDuration, TASK-0017]
    path: tradeoff/2026-04-15-maxdrawdownduration-computed-from-per-bar-equity.md
    summary: "Superseded 2026-04-16: MaxDrawdown depth also moved to computeMaxDrawdownDepth(curve) — both metrics now walk the same per-bar EquityPoint curve. The inconsistency this decision documented no longer exists."

  - id: 2026-04-15-instrument-declared-before-first-run
    title: "Target instrument must be declared in writing before the first backtest run"
    date: 2026-04-15
    status: accepted
    category: algorithm
    tags: [backtest, instrument, anti-cherry-picking, research-hygiene, methodology, TASK-0028]
    path: algorithm/2026-04-15-instrument-declared-before-first-run.md
    summary: "Instrument name must be written into TASK-0028's acceptance criteria before any `go run cmd/backtest` is executed. Post-hoc selection invalidates the proliferation gate check."

  - id: 2026-04-15-baseline-backtest-period-2018-2024
    title: "Baseline backtest period set to 2018–2024 for NSE strategy evaluation"
    date: 2026-04-15
    status: experimental
    category: algorithm
    tags: [backtest, period, NSE, regime, training-window, TASK-0028]
    path: algorithm/2026-04-15-baseline-backtest-period-2018-2024.md
    summary: "2018-01-01 to 2024-12-31 committed before any run. Chosen to include the 2020 crash and 2022 flat-to-down regime — the two periods that stress-test trend-following and mean-reversion respectively. Testing only on 2020–2021 would be testing in the most favorable possible environment for both strategies."

  - id: 2026-04-15-analytics-compute-returns-extracted-helper
    title: "`computeReturns` extracted as package-level helper in `internal/analytics`"
    date: 2026-04-15
    status: experimental
    category: architecture
    tags: [refactor, returns, computeReturns, analytics, sharpe, sortino, calmar, tail-ratio, TASK-0016]
    path: architecture/2026-04-15-analytics-compute-returns-extracted-helper.md
    summary: "`computeReturns([]EquityPoint) []float64` extracted from `computeSharpe` after four metrics (Sharpe, Sortino, Calmar, TailRatio) all needed the same return series. Eliminates duplication without adding abstraction for its own sake. Function is unexported."

  - id: 2026-04-15-sortino-population-denominator-rollinger-hoffman
    title: "Sortino uses population-style denominator over all observations (Rollinger-Hoffman)"
    date: 2026-04-15
    status: experimental
    category: convention
    tags: [sortino, downside-deviation, convention, analytics, rollinger-hoffman, TASK-0016]
    path: convention/2026-04-15-sortino-population-denominator-rollinger-hoffman.md
    summary: "Sortino denominator is `sum(min(r,0)²) / n` over all observations (Rollinger-Hoffman), not divided by count of negative returns only. Most common in practice; agrees with Bloomberg. Alternative produces volatile estimates on small samples."

  - id: 2026-04-15-sweep-plateau-non-contiguous-min-max-range
    title: "Sweep plateau uses non-contiguous min/max range over all qualifying entries"
    date: 2026-04-15
    status: experimental
    category: architecture
    tags: [sweep, parameter-sweep, plateau, non-contiguous, internal/sweep, TASK-0023]
    path: architecture/2026-04-15-sweep-plateau-non-contiguous-min-max-range.md
    summary: "Plateau collects min/max ParamValue of all qualifying entries regardless of contiguity. A non-contiguous gap widens the range rather than splitting it — conservative interpretation. `output.WriteSweep` lives in internal/output (import direction: output→sweep)."

  - id: 2026-04-15-sweep-type-names-no-stutter-lo-hi-params
    title: "Sweep types renamed to eliminate stutter; `min/max` params renamed to `lo/hi`"
    date: 2026-04-15
    status: experimental
    category: convention
    tags: [stutter, naming, go-lint, revive, builtinShadow, internal/sweep, TASK-0023]
    path: convention/2026-04-15-sweep-type-names-no-stutter-lo-hi-params.md
    summary: "`SweepConfig/SweepResult/SweepReport` renamed to `Config/Result/Report` (no-stutter rule). `min/max` params renamed to `lo/hi` to avoid shadowing Go 1.21 builtins (gocritic builtinShadow). Applies project-wide no-stutter convention to sweep package."

  - id: 2026-04-15-sweep-sequential-execution-no-errgroup
    title: "Sweep executes parameter steps sequentially — no `errgroup` parallelism"
    date: 2026-04-15
    status: experimental
    category: tradeoff
    tags: [concurrency, sequential, errgroup, dependencies, internal/sweep, TASK-0023]
    path: tradeoff/2026-04-15-sweep-sequential-execution-no-errgroup.md
    summary: "Sequential execution chosen over errgroup parallelism. Constraint: `golang.org/x/sync` not in go.mod; CLAUDE.md prohibits new dependencies without discussion. For 20–50 step daily-bar sweeps, sequential completes in seconds. Upgrade path localized to `Run()`."

  - id: 2026-04-15-sweep-run-returns-report-not-slice
    title: "`sweep.Run` returns `SweepReport` (results + plateau) rather than `[]SweepResult`"
    date: 2026-04-15
    status: experimental
    category: tradeoff
    tags: [return-type, plateau, report, internal/sweep, TASK-0023]
    path: tradeoff/2026-04-15-sweep-run-returns-report-not-slice.md
    summary: "`Run` returns `SweepReport` so plateau analysis always runs and callers can't accidentally skip it. Cost: slightly heavier return type. Callers who only need the slice use `report.Results`."

  - id: 2026-04-15-sweep-strategy-factory-func-type
    title: "Sweep uses `StrategyFactory func(float64) (strategy.Strategy, error)` for parameterization"
    date: 2026-04-15
    status: experimental
    category: architecture
    tags: [parameterization, strategy-factory, sweep, internal/sweep, TASK-0023]
    path: architecture/2026-04-15-sweep-strategy-factory-func-type.md
    summary: "Sweep constructs strategies via a caller-supplied `func(float64) (strategy.Strategy, error)` closure. Sweep package stays agnostic to concrete strategy types. Rejected: `ParameterizableStrategy` interface requiring `WithParam` on every concrete strategy."

  - id: 2026-04-13-benchmark-report-separate-struct
    title: "BenchmarkReport is a separate struct, not an extension of Report"
    date: 2026-04-13
    status: accepted
    category: convention
    tags: [benchmark, BenchmarkReport, Report, analytics, struct-design, convention, TASK-0018]
    path: convention/2026-04-13-benchmark-report-separate-struct.md
    summary: "BenchmarkReport defined as a separate struct from Report. Report fields (TradeCount, WinRate, TotalPnL) are semantically meaningless for buy-and-hold; sharing a struct would cause silent zero-value misrepresentation. Each output type carries only valid fields for its subject."

  - id: 2026-04-13-benchmark-cagr-from-elapsed-time
    title: "BenchmarkReport annualized return uses actual elapsed calendar time, not bar-count"
    date: 2026-04-13
    status: accepted
    category: algorithm
    tags: [benchmark, CAGR, annualized-return, elapsed-time, BenchmarkReport, analytics, TASK-0018]
    path: algorithm/2026-04-13-benchmark-cagr-from-elapsed-time.md
    summary: "CAGR uses actual elapsed calendar time (timestamp delta / 365.25) rather than bar-count × annualization factor. CAGR is quoted in calendar years for external comparability; Sharpe uses bar-count annualization because volatility accumulates at bar frequency. The two metrics intentionally use different time bases."

  - id: 2026-04-13-function-parameter-injection-for-testability
    title: "Function-parameter injection for testability — complement to Config injection"
    date: 2026-04-13
    status: accepted
    category: convention
    tags: [testability, dependency-injection, http.Client, function-parameters, LoginFlow, cmdutil, convention]
    path: convention/2026-04-13-function-parameter-injection-for-testability.md
    summary: "Stateless helper functions that call injectable external code (HTTP client, base URL) should accept the dependency as a function parameter, not via a new Config struct. Pattern: callers pass real dependencies explicitly; tests pass httptest.Server clients. Complement to Config-level injection for long-lived structs."

  - id: 2026-04-13-vol-targeting-algorithm-choices
    title: "Vol-targeting sizing: algorithm choices for window, returns, zero-vol, and cap"
    date: 2026-04-13
    status: accepted
    category: algorithm
    tags: [volatility-targeting, position-sizing, SizingModel, log-returns, sample-variance, 20-bar-window, no-lookahead, TASK-0021]
    path: algorithm/2026-04-13-vol-targeting-algorithm-choices.md
    summary: "20-bar rolling sample std dev of daily log returns. Zero vol → fraction=0 → buy skipped (not fallback to fixed). Fraction capped at 1.0. Vol computed from candles[:i] at fill time (no lookahead). Sample variance consistent with Sharpe. EWMA and simple returns rejected for v1."

  - id: 2026-04-13-sma-lookback-returns-slow-period-with-guard
    title: "SMA crossover Lookback() returns slowPeriod, crossover guard handles first bar"
    date: 2026-04-13
    status: accepted
    category: algorithm
    tags: [sma-crossover, lookback, guard, slowPeriod, strategy, interface, TASK-0012]
    path: algorithm/2026-04-13-sma-lookback-returns-slow-period-with-guard.md
    summary: "Lookback() returns slowPeriod per acceptance criterion; an internal guard (n <= slowPeriod → Hold) handles the one bar where previous slow SMA is not yet valid. Pattern to reuse for any future crossover strategy."

  - id: 2026-04-13-sma-crossover-strict-crossover-vs-level-comparison
    title: "SMA crossover: strict crossover detection, not level comparison"
    date: 2026-04-13
    status: accepted
    category: algorithm
    tags: [sma-crossover, signal-semantics, crossover, level-comparison, strategy, Next, BarResult, TASK-0012]
    path: algorithm/2026-04-13-sma-crossover-strict-crossover-vs-level-comparison.md
    summary: "Strict crossover (Buy/Sell only on transition bar) chosen over level comparison (Buy every bar fast>slow). Behaviorally equivalent under no-pyramiding, but strict crossover is semantically correct: Signal means 'act now', not 'regime is bullish'. Keeps BarResult logs diagnostic."

  - id: 2026-04-10-zerodha-daily-candles-adjusted-for-corporate-actions
    title: "Zerodha daily candles are adjusted for corporate actions — no adjustment layer needed"
    date: 2026-04-10
    status: accepted
    category: infrastructure
    tags: [zerodha, data-quality, corporate-action, adjusted-prices, split, bonus, dividend, demerger, daily-candles, intraday, TASK-0025]
    path: infrastructure/2026-04-10-zerodha-daily-candles-adjusted-for-corporate-actions.md
    summary: "Kite Connect day candles are adjusted for splits, bonuses, rights, spin-offs, and extraordinary dividends. No adjustment parameter exists — it's automatic and retroactive. Regular dividends are market-adjusted only (not total-return). Intraday candles are NOT adjusted. Demergers use Zerodha's COA methodology (may differ from TradingView). TASK-0012 and TASK-0015 are unblocked."

  - id: 2026-04-10-sharpe-zero-for-degenerate-inputs
    title: "Sharpe returns 0 for degenerate inputs — Compute() stays error-free"
    date: 2026-04-10
    status: accepted
    category: tradeoff
    tags: [sharpe, analytics, error-handling, pure-function, Compute, degenerate, zero-variance, timeframe, API-design]
    path: tradeoff/2026-04-10-sharpe-zero-for-degenerate-inputs.md
    summary: "Compute() returns SharpeRatio=0 for <3 equity points, zero-variance curves, or unknown timeframes. Keeps the function error-free and consistent with other zero-default metrics. Requires sharpeAnnualizationFactor switch to stay exhaustive — 100% coverage enforces this."

  - id: 2026-04-10-sharpe-sample-variance
    title: "Sharpe ratio uses sample variance (n-1), not population variance (n)"
    date: 2026-04-10
    status: accepted
    category: algorithm
    tags: [sharpe, analytics, variance, statistics, standard-deviation, sample, population, equity-curve]
    path: algorithm/2026-04-10-sharpe-sample-variance.md
    summary: "Sample std dev (÷ n-1) chosen over population std dev (÷ n). Backtests are finite samples — population variance systematically underestimates variance and inflates Sharpe, especially on short intraday windows. Consistent with Bloomberg, QuantLib, and standard quant practice."

  - id: 2026-04-10-nse-annualization-factors
    title: "NSE annualization factors for Sharpe and volatility calculations"
    date: 2026-04-10
    status: accepted
    category: convention
    tags: [NSE, annualization, sharpe, volatility, timeframe, 15min, daily, bars-per-year, analytics, convention]
    path: convention/2026-04-10-nse-annualization-factors.md
    summary: "NSE session is 9:15–3:30 IST = 375 min/day. Annualization factors: daily→252, 15min→6300, 1min→94500. US session (390 min/day = 26 bars) must not be used for NSE strategies. All Sharpe, Sortino, and vol-targeting implementations use these constants."

  - id: 2026-04-10-corporate-action-verification-gate
    title: "Zerodha corporate action verification required before running any strategy"
    date: 2026-04-10
    status: accepted
    category: infrastructure
    tags: [zerodha, data-quality, corporate-action, adjusted-prices, unadjusted, split, dividend, gate, TASK-0025]
    path: infrastructure/2026-04-10-corporate-action-verification-gate.md
    summary: "TASK-0025 is a mandatory gate before any strategy executes. Zerodha's adjustment behaviour must be verified against a known split event. Unadjusted prices cause phantom drawdowns and corrupted Sharpe — silent failures that waste entire strategy evaluation runs."

  - id: 2026-04-10-strategy-proliferation-gate
    superseded_by: 2026-04-25-cross-instrument-proliferation-gate
    title: "Strategy proliferation gate — Sharpe ≥ 0.5 vs buy-and-hold before variation strategies"
    date: 2026-04-10
    status: superseded
    category: algorithm
    tags: [strategy, sharpe, gate, research-methodology, MACD, bollinger-bands, SMA, RSI, buy-and-hold, overfitting]
    path: algorithm/2026-04-10-strategy-proliferation-gate.md
    summary: "MACD and Bollinger Bands are only built if the baseline strategy in their thesis category (SMA crossover or RSI) achieves Sharpe ≥ 0.5 vs buy-and-hold after costs. Threshold set before seeing results to prevent post-hoc rationalisation. Low bar: filters dead strategies, not underpowered ones."

  - id: 2026-04-10-equitypoint-in-pkg-model
    title: "EquityPoint defined in pkg/model, not internal/engine"
    date: 2026-04-10
    status: accepted
    category: convention
    tags: [equity-curve, model, pkg/model, analytics, architecture, dependency-direction, EquityPoint]
    path: convention/2026-04-10-equitypoint-in-pkg-model.md
    summary: "EquityPoint lives in pkg/model so analytics and output can import it without depending on internal/engine. Engine → model is the only valid dep direction; analytics → engine would violate the architecture."

  - id: 2026-04-10-equity-curve-covers-all-bars
    title: "Equity curve records every bar, including warmup"
    date: 2026-04-10
    status: accepted
    category: convention
    tags: [equity-curve, engine, portfolio, lookback, warmup, analytics]
    path: convention/2026-04-10-equity-curve-covers-all-bars.md
    summary: "RecordEquity is called unconditionally for every candle, including warmup bars. Invariant: len(EquityCurve()) == len(candles) always. Warmup snapshots show cash-only equity (no fills possible yet). Chosen over post-lookback-only recording to give analytics a stable, length-predictable time series."

  - id: 2026-04-09-no-type-name-stutter-project-wide
    title: "No type-name stutter — project-wide convention"
    date: 2026-04-09
    status: accepted
    category: convention
    tags: [naming, revive, convention, stutter, output, Config]
    path: convention/2026-04-09-no-type-name-stutter-project-wide.md
    summary: "Exported types must not repeat their package name (output.OutputConfig → output.Config). Revive linter enforces this project-wide. Task acceptance criteria that name types describe intent, not the literal identifier — the no-stutter rule takes precedence."

  - id: 2026-04-09-io-writer-in-config-for-stdout-testability
    title: "io.Writer field in Config for stdout testability"
    date: 2026-04-09
    status: accepted
    category: convention
    tags: [output, testing, io.Writer, Config, testability, stdout, convention]
    path: convention/2026-04-09-io-writer-in-config-for-stdout-testability.md
    summary: "output.Config.Stdout io.Writer (nil → os.Stdout) allows unit tests to capture stdout via bytes.Buffer without OS-level pipe hacks. Follows the same injectable-via-Config pattern as sleep injection in the zerodha provider."

  - id: 2026-04-09-error-wrapping-required-at-every-call-site
    title: "Every error return must be wrapped with call-site context"
    date: 2026-04-09
    status: accepted
    category: convention
    tags: [error-handling, wrapping, fmt.Errorf, convention, zerodha, provider]
    path: convention/2026-04-09-error-wrapping-required-at-every-call-site.md
    summary: "All error returns must use fmt.Errorf(\"context: %w\", err) — bare return err is disallowed. Wrapping message describes the operation, not the error. Exceptions: intentional best-effort discards (nolint) and propagating already-wrapped sentinels like ErrAuthRequired."

  - id: 2026-04-08-godoc-required-on-exported-types
    title: "Godoc comments are required on all exported types and functions"
    date: 2026-04-08
    status: accepted
    category: convention
    tags: [godoc, naming, revive, convention, documentation, exported]
    path: convention/2026-04-08-godoc-required-on-exported-types.md
    summary: "All exported identifiers must have a doc comment starting with the identifier name. Enforced by revive linter (exported rule) — missing comments fail CI. Unexported identifiers are not checked but should be commented when logic is non-obvious."

  - id: 2026-04-08-no-package-name-stutter-in-zerodha
    title: "Types in pkg/provider/zerodha must not repeat the package name"
    date: 2026-04-08
    status: accepted
    category: convention
    tags: [zerodha, naming, revive, convention, provider, stutter]
    path: convention/2026-04-08-no-package-name-stutter-in-zerodha.md
    summary: "ZerodhaProvider renamed to Provider (zerodha.ZerodhaProvider stutters). Convention: all exported types in pkg/provider/zerodha omit the 'Zerodha' prefix — the package name provides the context. Revive linter enforces this."

  - id: 2026-04-08-provider-validates-via-model-newcandle
    title: "Provider validates API responses via model.NewCandle at parse time"
    date: 2026-04-08
    status: accepted
    category: convention
    tags: [zerodha, provider, candle, validation, model, NewCandle, parseKiteCandles, convention]
    path: convention/2026-04-08-provider-validates-via-model-newcandle.md
    summary: "Providers call model.NewCandle (with Validate) rather than constructing struct literals. Catches invalid API data (e.g. OHLC=0 for suspended instruments) at the data boundary. Convention: providers validate, engine and analytics trust."

  - id: 2026-04-08-dohttp-centralizes-auth-errors
    title: "`doHTTP` helper centralizes 401/403 → ErrAuthRequired mapping"
    date: 2026-04-08
    status: accepted
    category: architecture
    tags: [zerodha, provider, http, auth, error-handling, ErrAuthRequired, doHTTP, convention]
    path: architecture/2026-04-08-dohttp-centralizes-auth-errors.md
    summary: "Package-private doHTTP maps HTTP 401/403 → ErrAuthRequired in one place. Every HTTP call in the package gets correct auth error handling automatically. New call sites can't accidentally miss the mapping."

  - id: 2026-04-08-sleep-injection-via-config
    title: "Sleep injection via Config for rate-limit throttling"
    date: 2026-04-08
    status: accepted
    category: convention
    tags: [zerodha, provider, testing, sleep, config, injection, rate-limit, convention]
    path: convention/2026-04-08-sleep-injection-via-config.md
    summary: "Config.Sleep func(time.Duration) defaults to time.Sleep; tests pass a no-op. Chosen over global variable (violates no-global-state rule), build tags (hidden dependency), and Sleeper interface (overkill for one function). Preferred pattern for all injectable behaviors in this package."

  - id: 2026-04-07-zerodha-instrument-token-lookup
    title: "Zerodha instrument token lookup — CSV download at provider init"
    date: 2026-04-07
    status: accepted
    category: architecture
    tags: [zerodha, provider, instrument-token, instruments-csv, lookup, init, kite-connect]
    path: architecture/2026-04-07-zerodha-instrument-token-lookup.md
    summary: "Provider downloads /instruments CSV once at init, builds map[exchange:symbol]→token in memory. Skips per-call download (wasteful) and file caching (unnecessary complexity for v1). ErrInstrumentNotFound returned for unknown symbols."

  - id: 2026-04-07-timeframe-weekly-unsupported-in-zerodha
    title: "TimeframeWeekly excluded from Zerodha SupportedTimeframes"
    date: 2026-04-07
    status: accepted
    category: convention
    tags: [zerodha, provider, timeframe, weekly, kite-connect, model, SupportedTimeframes]
    path: convention/2026-04-07-timeframe-weekly-unsupported-in-zerodha.md
    summary: "Kite Connect has no weekly interval. TimeframeWeekly stays in pkg/model (valid type, may be served by future providers) but is omitted from zerodha.Provider.SupportedTimeframes(). Engine must validate strategy timeframe against SupportedTimeframes before calling FetchCandles."

  - id: 2026-04-07-stdlib-dotenv-no-godotenv
    title: "stdlib .env parser — no godotenv dependency"
    date: 2026-04-07
    status: accepted
    category: tradeoff
    tags: [dependencies, dotenv, credentials, prototype, stdlib, convention]
    path: tradeoff/2026-04-07-stdlib-dotenv-no-godotenv.md
    summary: "25-line stdlib implementation chosen over github.com/joho/godotenv. Covers all actual use cases (KEY=value, blank lines, # comments). Repo rule: no new dependencies without justification. Local to cmd/authtest — not a shared utility."

  - id: 2026-04-07-zerodha-cache-strategy
    title: "Zerodha provider — local file-based caching strategy"
    date: 2026-04-07
    status: accepted
    category: infrastructure
    tags: [zerodha, cache, provider, file-cache, json, invalidation, kite-connect]
    path: infrastructure/2026-04-07-zerodha-cache-strategy.md
    summary: "File-based JSON cache in .cache/zerodha/ keyed on exact (instrument, timeframe, from, to). Historical data never invalidates; recent data (to >= today) has 1-hour TTL. CachedProvider is a DataProvider decorator — cache is above the chunk loop so a hit skips all API calls."

  - id: 2026-04-07-zerodha-pagination-strategy
    title: "Zerodha historical data — pagination and rate-limit strategy"
    date: 2026-04-07
    status: accepted
    category: architecture
    tags: [zerodha, provider, pagination, chunking, rate-limit, historical-data, kite-connect]
    path: architecture/2026-04-07-zerodha-pagination-strategy.md
    summary: "Fixed 350ms sleep between chunks (≈2.85 req/sec). chunkDateRange splits [from,to) into windows using per-interval maxDays map with safety margins. Limits are community-established (not published by Zerodha) — must verify in Phase 5 prototype before finalising constants."

  - id: 2026-04-07-zerodha-auth-strategy
    title: "Zerodha auth strategy — manual token paste for v1"
    date: 2026-04-07
    status: accepted
    category: architecture
    tags: [zerodha, auth, kite-connect, access-token, provider, session]
    path: architecture/2026-04-07-zerodha-auth-strategy.md
    summary: "Manual copy-paste of request_token chosen over local HTTP server for v1. Token persisted to ~/.config/backtest/token.json with 6AM IST expiry. Auth is internal to the provider — not exposed through DataProvider interface."

  - id: 2026-04-07-hold-signal-not-passed-to-portfolio
    title: "SignalHold filtered at engine level — never passed to portfolio"
    date: 2026-04-07
    status: accepted
    category: convention
    tags: [engine, portfolio, signal, hold, applySignal, convention, coverage]
    path: convention/2026-04-07-hold-signal-not-passed-to-portfolio.md
    summary: "Engine guards with `if pendingSignal != SignalHold` before calling applySignal; Hold is filtered at the call site, not inside the portfolio layer. applySignal retains a defensive Hold branch covered by whitebox test."

  - id: 2026-04-07-max-drawdown-from-equity-curve
    title: "MaxDrawdown computed from equity curve, not per-trade losses"
    date: 2026-04-07
    status: accepted
    category: algorithm
    tags: [analytics, drawdown, equity-curve, metrics, algorithm]
    path: algorithm/2026-04-07-max-drawdown-from-equity-curve.md
    summary: "MaxDrawdown uses peak-to-trough on the per-bar EquityPoint curve via computeMaxDrawdownDepth(). Original 2026-04-07 impl used closed-trade P&L accumulation starting at 0 — buggy (could exceed 100%). Fixed 2026-04-16 to use initialCash-anchored curve."

  - id: 2026-04-07-breakeven-counts-as-loss
    title: "Break-even trades classified as losses in analytics"
    date: 2026-04-07
    status: accepted
    category: convention
    tags: [analytics, win-rate, pnl, trade, classification, convention]
    path: convention/2026-04-07-breakeven-counts-as-loss.md
    summary: "Trades with RealizedPnL <= 0 count as losses. Break-even is not a third category — it inflates win rate and adds struct complexity for a near-impossible edge case on real commission-bearing fills."

  - id: 2026-04-06-value-semantics-for-domain-types
    title: "Value semantics for domain model types (Candle, Config)"
    date: 2026-04-06
    status: accepted
    category: convention
    tags: [candle, model, value-semantics, pointer, gocritic, performance, convention]
    path: convention/2026-04-06-value-semantics-for-domain-types.md
    summary: "Domain types like Candle and Config use value semantics even when large; pointer semantics would hurt cache locality in slice-heavy hot loops. gocritic hugeParam is suppressed by convention."

  - id: 2026-04-06-context-parameter-deferred
    title: "context.Context deferred from Run() and DataProvider interface"
    date: 2026-04-06
    status: accepted
    category: tradeoff
    tags: [context, engine, provider, interface, cancellation, zerodha]
    path: tradeoff/2026-04-06-context-parameter-deferred.md
    summary: "context.Context deferred until just before Zerodha provider was written. Resolved 2026-04-07: both Run() and FetchCandles() now accept ctx as first param. Interface is correct before any real provider code landed."

  - id: 2026-04-03-no-pyramiding-v1
    title: "No pyramiding — single position per instrument enforced in v1"
    date: 2026-04-03
    status: accepted
    category: tradeoff
    tags: [portfolio, position-sizing, pyramiding, v1-scope, engine]
    path: tradeoff/2026-04-03-no-pyramiding-v1.md
    summary: "Buy signals on an already-open position are silently skipped in v1. Pyramiding deferred until a concrete strategy requires scale-in behaviour."

  - id: 2026-04-02-trade-pnl-stored-not-computed
    title: "Trade.RealizedPnL stored on the struct, not computed on-demand"
    date: 2026-04-02
    status: accepted
    category: convention
    tags: [trade, pnl, analytics, engine, commission, slippage]
    path: convention/2026-04-02-trade-pnl-stored-not-computed.md
    summary: "Engine computes and stores RealizedPnL at close time so analytics never needs to know about commission models or slippage — keeping it a pure read-only reporting layer."

  - id: 2026-04-02-strategy-lookback-as-interface-method
    title: "Strategy.Lookback() as a first-class interface method"
    date: 2026-04-02
    status: accepted
    category: architecture
    tags: [strategy, interface, engine, lookback, no-lookahead]
    path: architecture/2026-04-02-strategy-lookback-as-interface-method.md
    summary: "Lookback is a required interface method so the engine enforces the no-lookahead constraint universally, rather than delegating it to individual strategies."
```
