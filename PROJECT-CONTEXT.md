# PROJECT-CONTEXT.md — Backtesting Algo Trading

## Project Identity

A Go-based backtesting engine for evaluating trading strategies against historical market data. The engine is strategy-agnostic — it provides the execution framework, data pipeline, and analytics; strategies are plugged in as implementations of a common interface. The data source is abstracted behind an interface, with Zerodha Kite Connect as the first concrete implementation. This is a backtesting-only project — no live trading, no paper trading, no forward testing. If that changes in the future, the architecture should not fight it, but we are not designing for it today.

**Language:** Go  
**Lifecycle:** Bootstrapping  

---

## Architecture Overview

The system is composed of five core components that communicate through well-defined interfaces, not concrete types.

**1. Data Provider** — Fetches and serves historical OHLCV candle data. Abstract interface with Zerodha Kite Connect as the first implementation. Supports multiple timeframes (1min, 5min, 15min, daily, weekly) — the strategy decides which timeframe it needs. Data is fetched once and cached locally to avoid repeated API calls during iterative backtesting.

**2. Strategy** — The trading logic. Receives candle data, computes indicators, emits signals (buy/sell/hold). Every strategy implements the same interface. Strategies declare what timeframe and how much historical lookback they need. Strategies must be stateless between runs — all state lives in the engine, not the strategy.

**3. Engine** — The backtest executor. Feeds historical candles to the strategy one at a time (event-driven, not vectorized). Manages simulated portfolio state: positions, cash, order execution. Applies realistic constraints: slippage model, commission/brokerage fees, position sizing. Produces a complete trade log.

**4. Analytics** — Computes performance metrics from the trade log. Start simple: P&L, win rate, max drawdown. Extensible for later: Sharpe ratio, Sortino ratio, drawdown curves, trade distribution, equity curves. Analytics is a separate package — it takes a trade log as input, not a reference to the engine.

**5. Results/Output** — Formats and presents backtest results. Start with structured console output and JSON export. Extensible for later: HTML reports, charts, comparison tables across strategies.

**Data flow:**
```
DataProvider → Engine ← Strategy
                ↓
            Trade Log
                ↓
            Analytics
                ↓
            Results/Output
```

The engine is the orchestrator. It pulls data from the provider, feeds it to the strategy, records trades, and hands the trade log to analytics.

---

## Active Conventions

### Development Approach
- **Test-Driven Development (TDD).** Write the test first, then the implementation. No exceptions. Every public function has a test. Table-driven tests are the default pattern in Go — use them.
- **Interface-first design.** Define the interface before writing any implementation. This applies to DataProvider, Strategy, Analytics, and any future component.
- **Small, focused packages.** Each component is its own Go package. No God packages. No circular dependencies.

### Code Patterns
- Use Go interfaces for all component boundaries. Keep interfaces small — prefer 1-3 methods.
- Errors are returned, not panicked. Use typed errors where the caller needs to distinguish error kinds.
- No global state. No init() functions with side effects. Dependencies are injected explicitly.
- Configuration is passed as structs, not scattered across function parameters.
- Prefer composition over inheritance (Go doesn't have inheritance anyway, but avoid deep embedding chains).

### Library Usage
- **Technical indicators:** Use `github.com/markcheno/go-talib` (pure Go port of TA-Lib). Do NOT reimplement indicator math — use the library. If go-talib doesn't have an indicator we need, wrap it cleanly so we can swap implementations later.
- **General principle:** If a well-maintained, popular Go library exists for something, use it. Don't write complex logic from scratch. But always wrap external libraries behind our own interface so we're not coupled to them.

### Naming
- Packages: lowercase single words (`engine`, `strategy`, `provider`, `analytics`)
- Interfaces: describe behavior, not implementation (`DataProvider`, not `ZerodhaProvider`)
- Test files: `*_test.go` in the same package (whitebox) or `*_test` package (blackbox) depending on what's being tested
- Config structs: `Config` suffix (`EngineConfig`, `BacktestConfig`)

### File/Folder Structure
```
backtesting-algo-trading/
├── cmd/                    # CLI entrypoints
│   └── backtest/
│       └── main.go
├── internal/               # Private application code
│   ├── engine/             # Backtest execution engine
│   ├── analytics/          # Performance metrics computation
│   └── output/             # Result formatting and export
├── pkg/                    # Public reusable packages
│   ├── strategy/           # Strategy interface + base helpers
│   ├── provider/           # DataProvider interface + implementations
│   │   ├── provider.go     # Interface definition
│   │   └── zerodha/        # Zerodha Kite Connect implementation
│   └── model/              # Shared domain types (Candle, Trade, Signal, etc.)
├── strategies/             # Concrete strategy implementations
│   └── sma_crossover/      # Example: SMA crossover strategy
├── testdata/               # Sample CSV/JSON data for tests
├── decisions/              # Decision journal entries (if skill linked)
├── sessions/               # Session logs (if skill linked)
├── PROJECT-CONTEXT.md      # This file
├── go.mod
├── go.sum
└── README.md
```

### Testing Conventions
- Every package has tests. No untested public API.
- Use `testdata/` for fixture files (sample candle data in CSV or JSON).
- Mock external dependencies (Zerodha API) using interfaces — never call real APIs in tests.
- Benchmark tests for the engine's candle processing loop — performance matters when running thousands of backtests.
- Test strategies against known datasets with known expected outcomes (golden file tests).

---

## Current State

**Status:** Not yet started. This document is the starting point.

**Immediate next steps:**
1. Initialize Go module, set up folder structure
2. Define core domain types in `pkg/model/` (Candle, Trade, Signal, Position, etc.)
3. Define `DataProvider` interface in `pkg/provider/`
4. Define `Strategy` interface in `pkg/strategy/`
5. Build engine with a trivial test strategy and hardcoded candle data
6. Add analytics for basic metrics
7. Implement Zerodha data provider

---

## Key Constraints & Boundaries

### Data Source Abstraction
The `DataProvider` interface must be fully abstract. Nothing in the engine, strategy, or analytics packages should know about Zerodha. The interface should look roughly like:
```go
type DataProvider interface {
    FetchCandles(instrument string, timeframe Timeframe, from, to time.Time) ([]Candle, error)
    SupportedTimeframes() []Timeframe
}
```
This means we can later add CSV file provider (for testing), another broker's API, or a database-backed provider without touching any other code.

### Scope Boundaries
- **Single instrument backtesting** for now. The engine processes one instrument per run. But design the `Candle` and `Trade` types to carry an instrument identifier so portfolio-level backtesting can be added later without type changes.
- **No live trading.** No WebSocket connections, no order placement, no real money. The engine is purely a simulator.
- **No optimization/parameter sweep** in v1. The engine runs one strategy with one parameter set per invocation. Optimization (grid search, walk-forward) is a future addition — but keep the engine's API clean enough that wrapping it in a parameter sweep loop would be straightforward.

### Realistic Simulation
- **Slippage model:** Configurable slippage (default: a fixed percentage). Even a simple model is better than none.
- **Brokerage/commission:** Configurable per-trade cost. Default should model Zerodha's fee structure (₹20 per order or 0.03% of turnover, whichever is lower, for intraday equity).
- **Position sizing:** Configurable. Start with fixed-quantity, design for percentage-of-equity and risk-based sizing later.
- **No lookahead bias.** The strategy only ever sees candles up to the current bar. The engine must enforce this — never pass future data to the strategy.
- **Fill assumptions:** Market orders fill at next candle's open price (not current candle's close). This is more realistic.

### Performance
- The engine should handle 10 years of daily candles or 1 year of minute candles without noticeable delay.
- Avoid unnecessary allocations in the hot loop (candle processing). Pre-allocate where possible.

---

## Dependency Landscape

### Go Libraries (planned)
| Library | Purpose | Notes |
|---|---|---|
| `github.com/markcheno/go-talib` | Technical indicators (SMA, EMA, RSI, MACD, Bollinger Bands, etc.) | Pure Go port of TA-Lib. No CGo. Wrap behind our own indicator interface. |
| `github.com/stretchr/testify` | Test assertions and mocking | Standard in Go ecosystem. Use `assert` and `require` sub-packages. |
| Standard library | HTTP client, JSON, CSV, time | Prefer stdlib over third-party for basic operations. |

### External Services
| Service | Purpose | Notes |
|---|---|---|
| Zerodha Kite Connect API | Historical candle data | Requires API key + access token. Rate limited. Abstract behind DataProvider interface. Never call in tests. |

### Environment Notes
- Zerodha API requires daily token refresh (access token expires daily). The provider implementation must handle this, but the interface should not expose it.
- Historical data from Zerodha has limits on how far back you can go and how many candles per request. The provider implementation should handle pagination/chunking internally.

---

## Quick Reference

```bash
# Initialize (once)
go mod init github.com/<username>/backtesting-algo-trading

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. ./internal/engine/

# Run a backtest (once CLI exists)
go run cmd/backtest/main.go --strategy sma-crossover --instrument RELIANCE --from 2024-01-01 --to 2025-01-01

# Lint
golangci-lint run ./...
```

---

## Linked Skills

| Skill | Purpose | Data Location in This Repo |
|---|---|---|
| `decision-journal` | Track decisions, rejected approaches, experiment reasoning | `decisions/` |
| `session-continuity` | Session start/end protocol, handoff between sessions | `sessions/` |
| `project-context` | This file — living project snapshot | `PROJECT-CONTEXT.md` |

---

## Rules for AI Assistants Working on This Repo

1. **Read this file first** at the start of every session. If it looks stale, ask about it.
2. **TDD is mandatory.** Write the test before the implementation. If you're about to write a function, write its test first. Show me the test, then the implementation.
3. **Never bypass the interface.** If you need data, go through DataProvider. If you need a strategy, go through the Strategy interface. No concrete type references across package boundaries.
4. **Use go-talib for indicators.** Do not hand-roll SMA/EMA/RSI/MACD calculations. If go-talib doesn't have it, discuss before implementing.
5. **No premature optimization.** Start simple, measure, then optimize. But also no premature pessimization — don't allocate inside hot loops without reason.
6. **Record decisions.** If we discuss multiple approaches and pick one, record it in the decision journal. If something is tried and abandoned, record why.
7. **End sessions properly.** Capture what was done, what's next, and what's unresolved.
8. **Keep this file updated.** If architecture changes, conventions change, or new constraints are discovered, update this file before the session ends.
9. **Ask before adding dependencies.** Don't pull in a new library without discussing it first. We want a minimal dependency footprint.
10. **Instrument identifier on everything.** Even though we're single-instrument now, every Candle, Trade, and Position should carry an instrument identifier. This is cheap insurance for portfolio-level backtesting later.