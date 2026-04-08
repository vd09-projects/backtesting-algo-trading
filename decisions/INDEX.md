# Decision Index

<!-- 
  This file is maintained by the decision-journal skill.
  Entries are in YAML format for machine-friendly querying.
  Newest entries go at the top. Do not manually reorder.
-->

```yaml
decisions:
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
