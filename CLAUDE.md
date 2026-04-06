# Project: backtesting-algo-trading

A Go-based backtesting engine for evaluating trading strategies against historical market data. The engine is strategy-agnostic — it provides the execution framework, data pipeline, and analytics; strategies are plugged in as implementations of a common interface. The data source is abstracted behind an interface, with Zerodha Kite Connect as the first concrete implementation. Backtesting-only — no live trading, no paper trading.

---

## Quality standards

This project uses the go-quality-review skill for code review. The following standards apply to all code in this repository.

### Review overrides

<!-- Uncomment and modify lines below to override default review levels -->
<!-- - behavior: always run at full depth, even for standard review -->
<!-- - architecture: skip for all levels -->
<!-- - test-quality: minimum coverage threshold is 85% (not 70%) -->

### Repo-specific rules

- All strategies must implement the `Strategy` interface defined in `pkg/strategy/`. Never reference a concrete strategy type across package boundaries.
- All data access must go through the `DataProvider` interface. No package outside `pkg/provider/` should know about Zerodha.
- Use `github.com/markcheno/go-talib` for technical indicators. Do not hand-roll SMA/EMA/RSI/MACD or other indicator math.
- TDD is mandatory. Write the test before the implementation. Every public function must have a test.
- No global state. No `init()` functions with side effects. All dependencies are injected explicitly.
- Errors are returned, not panicked. Use typed errors where the caller needs to distinguish error kinds.
- Every `Candle`, `Trade`, and `Position` must carry an instrument identifier — even though we're single-instrument now.
- No allocations inside the hot loop (candle processing) without pre-allocation justification.
- Do not add new dependencies without discussion. Keep the dependency footprint minimal.

### Commands

```bash
# Lint
golangci-lint run ./...

# Test with race detection
go test -race ./...

# Test with coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Benchmarks (engine hot loop)
go test -bench=. ./internal/engine/

# Run all checks before committing
golangci-lint run ./... && go test -race -coverprofile=coverage.out ./...
```

### Architecture

Event-driven backtesting engine. Data flows one direction: provider → engine ← strategy → trade log → analytics → output.

```
DataProvider → Engine ← Strategy
                ↓
            Trade Log
                ↓
            Analytics
                ↓
            Results/Output
```

- `cmd/backtest/`       → CLI entrypoint and wiring
- `internal/engine/`    → backtest executor, portfolio state, order simulation
- `internal/analytics/` → performance metrics (P&L, drawdown, win rate, Sharpe)
- `internal/output/`    → result formatting and JSON export
- `pkg/strategy/`       → Strategy interface + base helpers
- `pkg/provider/`       → DataProvider interface + Zerodha implementation
- `pkg/model/`          → shared domain types (Candle, Trade, Signal, Position)
- `strategies/`         → concrete strategy implementations

Dependencies flow inward toward `pkg/model/`. No circular imports. No concrete type references across package boundaries.
