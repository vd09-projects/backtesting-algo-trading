# Project: backtesting-algo-trading

A Go-based backtesting engine for evaluating trading strategies against historical market data. The engine is strategy-agnostic — it provides the execution framework, data pipeline, and analytics; strategies are plugged in as implementations of a common interface. The data source is abstracted behind an interface, with Zerodha Kite Connect as the first concrete implementation. Backtesting-only — no live trading, no paper trading.

---

## Mandatory process gates

Skipping any of these requires explicit user instruction.

### Before marking any task done

1. **TDD** — tests were written before the implementation they test (not after)
2. **Quality gate** — `/go-quality-review` at standard level was run if `internal/` or `pkg/` was touched; the `.quality-gate/last-pass` sentinel must be current
3. **Acceptance criteria** — every criterion in the task block is checked off

### During every turn

- **Skill terminal state** → read `workflows/INDEX.md` immediately to determine the next step; do not proceed without checking it
- **Design choice made** → mark inline in the response: `**Decision (topic) — category: status**` so the session-end harvest captures it
- **Editing production code in `internal/` or `pkg/`** → write the failing test first, before the implementation

### Session end

Both must fire before closing — the session-end hook will remind you:

- `/task-manager` — harvest implicit tasks from this session
- `/decision-journal` — harvest all inline decision marks from this session

---

## Skill coordination

This project uses 5 skills (`task-manager`, `algo-trading-veteran`, `algo-trading-lead-dev`, `go-quality-review`, `decision-journal`) coordinated via workflow files — see `workflows/INDEX.md`. After any skill reaches a terminal state, check `workflows/INDEX.md` for the matching trigger and follow it.

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
- Every public function must have a test.
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
go1.25.0 test -race ./...

# Test with coverage
go1.25.0 test -coverprofile=coverage.out ./...
go1.25.0 tool cover -func=coverage.out

# Benchmarks (engine hot loop)
go1.25.0 test -bench=. ./internal/engine/

# Run all checks before committing
golangci-lint run ./... && go1.25.0 test -race -coverprofile=coverage.out ./...
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