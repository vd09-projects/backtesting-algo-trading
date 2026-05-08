# Learned Patterns

## Project Conventions
- Error wrapping: `fmt.Errorf("context: %w", err)` — always wrap with context string
- Tests: table-driven with named subtests (`t.Run(tc.name, ...)`)
- Decimal money: `github.com/shopspring/decimal` — never float64 for P&L or position sizing
- Indicators: `github.com/markcheno/go-talib` only — hand-rolled SMA/EMA/RSI/MACD is a blocker
- Instrument identifier: every `Candle`, `Trade`, `Position` carries `.Instrument` string field
- Hot loop: `internal/engine/executor.go` — heap allocs per candle are blockers unless pre-allocated

## False Positive Suppressions
- `strategies/stub/` → skip all reviewers (test stub, intentionally minimal)
- `strategies/testutil/` → skip all reviewers (test utility, not production code)
- Independent indicator math in `*_test.go` files → suppress Dependency Reviewer complaint (tests may compute expected values without go-talib to verify against it)
- `cmd/` packages → suppress API & Contract Reviewer for internal wiring (CLI flags are not public API)

## Known Hot Spots
- `internal/engine/` — core hot loop; any change here has high blast radius; Concurrency & State Safety always warranted
- `pkg/model/` — shared domain types; field renames break all consumers silently; Ripple Effect always warranted
- `internal/analytics/` — math-heavy; Domain Logic Reviewer always warranted
- `cmd/universe-sweep/main.go` — strategy registry map; Ripple Effect Analyst always warranted

## Recurring Issues
<!-- Populated as reviews run — add entries here when patterns emerge across tasks -->
