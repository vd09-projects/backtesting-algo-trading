# Architecture Overview

## End-to-end Data Flow

Data moves in one direction. The engine is the chokepoint — nothing bypasses it.

```
┌──────────────────────────────────────────────────────────────────────────┐
│                            CLI Entrypoints                               │
│                                                                          │
│   cmd/backtest        — single instrument, full analytics               │
│   cmd/sweep           — 1D parameter sweep                              │
│   cmd/universe-sweep  — fixed strategy across many instruments          │
│   cmd/correlate       — pairwise strategy correlation from saved CSVs   │
│   cmd/rsi-diagnostic  — RSI signal frequency debugger                   │
│   cmd/authtest        — verify Zerodha OAuth credentials                │
│   cmd/providertest    — smoke-test the data provider                    │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │
                               │ 1. Build provider (Zerodha)
                               │ 2. Build strategy
                               │ 3. Build engine config
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                         DataProvider (interface)                         │
│                       pkg/provider/provider.go                           │
│                                                                          │
│   FetchCandles(ctx, instrument, timeframe, from, to) ([]Candle, error)  │
│                                                                          │
│   Concrete: pkg/provider/zerodha/                                        │
│   ┌──────────────────────────────────────────────────────────────────┐   │
│   │  Auth         → OAuth flow, token persistence                   │   │
│   │  Instruments  → CSV download, token lookup (NSE:X → int token)  │   │
│   │  HTTP layer   → chunked requests (Kite max 2000 candles/call)   │   │
│   │  Cache        → .cache/zerodha/{instrument}/{tf}_{from}_{to}.json│  │
│   └──────────────────────────────────────────────────────────────────┘   │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ []model.Candle (full series, validated)
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                       Engine (event loop)                                │
│                    internal/engine/engine.go                             │
│                                                                          │
│   for i, candle := range candles:                                        │
│     ① Fill pending signal at candle[i].Open  ←── next-open fill rule   │
│     ② Snapshot equity at candle[i].Close     ←── mark-to-market        │
│     ③ if i+1 >= lookback:                                               │
│           signal = strategy.Next(candles[:i+1])  ← no lookahead        │
│           pendingSignal = signal                                         │
│                                                                          │
│   ┌──────────────────────┐   ┌──────────────────────────────────────┐   │
│   │  Portfolio           │   │  Strategy (interface)                │   │
│   │  • Cash              │   │  • Name() string                     │   │
│   │  • Positions (map)   │   │  • Timeframe() Timeframe             │   │
│   │  • Trade log         │   │  • Lookback() int                    │   │
│   │  • Equity curve      │◄──│  • Next([]Candle) Signal             │   │
│   │  • Commission calc   │   │                                      │   │
│   │  • Slippage calc     │   │  strategies/smacrossover             │   │
│   │  • Sizing calc       │   │  strategies/rsimeanrev               │   │
│   └──────────────────────┘   │  strategies/stub                     │   │
│                               └──────────────────────────────────────┘   │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ []model.Trade + []model.EquityPoint
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                           Analytics                                      │
│                      internal/analytics/                                 │
│                                                                          │
│   analytics.Compute()           → Report (all performance metrics)      │
│   analytics.ComputeBenchmark()  → buy-and-hold baseline                │
│   analytics.ComputeRegimeSplits() → per-market-regime breakdown        │
│   analytics.ComputeCorrelation()  → Pearson r between strategy curves  │
│   analytics.CheckKillSwitch()     → live monitoring alerts              │
│   analytics.DSR()                 → deflated Sharpe (multi-testing)    │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ Report / analytics.RegimeReport / etc.
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                      Monte Carlo Bootstrap                               │
│                      internal/montecarlo/                                │
│                                                                          │
│   montecarlo.Bootstrap() → BootstrapResult                             │
│   • 10,000 resamples of per-trade ReturnOnNotional                      │
│   • Sharpe p5/p50/p95, drawdown p5/p50/p95                             │
│   • P(Sharpe > 0)                                                        │
│   • SharpeP5 → fed into DeriveKillSwitchThresholds()                   │
└──────────────────────────────┬───────────────────────────────────────────┘
                               │ BootstrapResult + Report
                               ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                         Output / Export                                  │
│                        internal/output/                                  │
│                                                                          │
│   output.Write()      → stdout table + optional JSON file              │
│   output.WriteSweep() → sweep rankings table to stdout                 │
│   universesweep.WriteCSV() → universe results CSV                      │
│   sweep2d CSV         → 2D grid results CSV                            │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Validation Harnesses (sit alongside the core loop)

```
┌───────────────────────────────────┐   ┌───────────────────────────────────┐
│      1D Parameter Sweep           │   │      Walk-Forward Validation       │
│      internal/sweep/              │   │      internal/walkforward/         │
│                                   │   │                                   │
│  Sweep one param over [Min,Max]   │   │  Rolling IS/OOS windows           │
│  → ranked Sharpe table            │   │  → OverfitFlag / NegativeFoldFlag │
│  → plateau range (robustness)     │   │  Folds run in parallel            │
└───────────────────────────────────┘   └───────────────────────────────────┘

┌───────────────────────────────────┐   ┌───────────────────────────────────┐
│      2D Grid Sweep                │   │      Universe Sweep               │
│      internal/sweep2d/            │   │      internal/universesweep/      │
│                                   │   │                                   │
│  Sweep two params as a grid       │   │  Same strategy, many instruments  │
│  Parallel via errgroup            │   │  → CSV ranked by Sharpe           │
│  DSR-corrected peak Sharpe        │   │  Parallel via errgroup            │
└───────────────────────────────────┘   └───────────────────────────────────┘
```

---

## Package Dependency Rules

```
cmd/*
  └─► internal/*
        └─► pkg/strategy, pkg/provider, pkg/model
              └─► pkg/model  (leaf — no outbound deps)

strategies/*
  └─► pkg/strategy (interface), pkg/model

pkg/provider/zerodha
  └─► pkg/provider (interface), pkg/model

NO circular imports.
NO concrete type references across package boundaries.
pkg/provider/zerodha is invisible to anything outside pkg/provider/.
```

---

## Core Invariants

### 1. No lookahead — enforced by the engine
`strategy.Next` receives `candles[:i+1]` — exactly what was visible at bar `i`.
Strategies cannot reach candle `i+1`. This is structural, not a convention.

### 2. Next-open fill — always
Signals generated at bar N are filled at bar N+1's **Open**.
You cannot accidentally get same-bar fills.

### 3. Determinism — bit-identical results
- No `time.Now()` in the engine or analytics.
- Monte Carlo RNG seed is always logged with results.
- Grid and universe results are written to pre-allocated slices at fixed indices before sorting — goroutine scheduling cannot change output order.

### 4. Analytics guards against false precision
- `TradeMetricsInsufficient` = true when trade count < 30. Win rate, profit factor, avg win/loss are **zeroed**, not estimated.
- `CurveMetricsInsufficient` = true when equity curve length < 252 bars (< 1 NSE trading year). Sharpe, Sortino, Calmar, tail ratio are **zeroed**.

### 5. No pyramiding
Second buy signal while a position is open is silently skipped. The engine does not error; it simply ignores the signal.

### 6. Engine is the source of truth
Go computes returns, fills, costs, and metrics. No metric is computed outside `internal/analytics` or `internal/montecarlo`. Python/notebooks consume engine output files — they never feed back into the engine.
