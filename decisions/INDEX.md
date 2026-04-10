# Decision Index

<!-- 
  This file is maintained by the decision-journal skill.
  Entries are in YAML format for machine-friendly querying.
  Newest entries go at the top. Do not manually reorder.
-->

```yaml
decisions:
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
    title: "Strategy proliferation gate — Sharpe ≥ 0.5 vs buy-and-hold before variation strategies"
    date: 2026-04-10
    status: accepted
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
    summary: "MaxDrawdown uses the standard peak-to-trough definition on the running equity curve. Per-trade drawdown was rejected as it understates risk from consecutive losers. A peak>0 guard means all-loss sequences report 0% drawdown."

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
