# `computeReturns` extracted as package-level helper in `internal/analytics`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | refactor, returns, computeReturns, analytics, sharpe, sortino, calmar, tail-ratio, TASK-0016 |

## Context

When TASK-0016 added Sortino, Calmar, and TailRatio to the analytics report, all four metrics (Sharpe, Sortino, Calmar, TailRatio) needed to compute per-bar returns from the equity curve. Previously, `computeSharpe` computed returns inline as a local variable.

## Decision

`computeReturns([]EquityPoint) []float64` was extracted as a package-level helper. Previously `computeSharpe` computed returns inline; with four metrics now consuming the same sequence (Sharpe, Sortino, Calmar, TailRatio), the extraction eliminates duplication without adding abstraction for its own sake. `benchmark.go` uses `computeReturns` for the same reason.

This is not a premature abstraction — the helper is consumed by four callers in the same package as of this commit, and the extraction is a direct response to observed duplication, not speculative reuse.

## Consequences

All return-based metrics are guaranteed to use the same return series. Any future metric added to `internal/analytics` that needs per-bar returns calls `computeReturns` rather than re-implementing the calculation. The function is unexported — it is an implementation detail of the analytics package, not a public API.
