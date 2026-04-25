# float64 for commission arithmetic, not decimal — migration deferred

- **Date:** 2026-04-25
- **Status:** experimental
- **Category:** tradeoff
- **Tags:** commission, float64, decimal, shopspring, arithmetic, TASK-0038, TASK-0057
- **Task:** TASK-0038

## Decision

Commission arithmetic in `internal/engine/commission.go` uses `float64` throughout, consistent
with the existing engine accounting layer (`Portfolio`, equity curve, trade P&L). A migration to
`shopspring/decimal` (or equivalent) is deferred to a separate task (TASK-0057).

## Rationale

The existing accounting layer — `Portfolio.cash`, `Trade.RealizedPnL`, `model.EquityPoint.Value`
— is all `float64`. Introducing `decimal` for commission arithmetic only, while leaving the rest
as `float64`, would create an inconsistent boundary: decimal precision on the cost side, float
arithmetic on the P&L side. That inconsistency would be worse than a uniform float64 baseline.

For the backtesting use case, float64 rounding errors on per-trade commission (~₹0.001 on ₹88)
do not materially affect strategy evaluation over hundreds of trades. The error is well within
the noise of fill-price approximation and slippage modelling.

## Risk

Float64 accumulation over many trades can produce small systematic biases in total cost reporting.
For a strategy running 200 trades at ~₹88 commission, the accumulated rounding error is well
under ₹1. Acceptable for backtesting; not acceptable if this engine ever moves to live execution
accounting.

## Deferred work

**TASK-0057** — Migrate engine accounting layer to `shopspring/decimal`. Covers: commission.go,
portfolio.go (cash), pkg/model Trade.RealizedPnL, Trade.Commission, EquityPoint.Value. This is a
coordinated migration — partial decimal adoption is worse than none.

## Revisit trigger

Before any live execution accounting. Float64 is not acceptable for money arithmetic in a
live trading system, only in a backtesting engine.
