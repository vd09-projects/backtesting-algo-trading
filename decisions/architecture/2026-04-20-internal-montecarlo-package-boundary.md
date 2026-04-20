# `internal/montecarlo` as a standalone package, isolated from analytics

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | montecarlo, bootstrap, package-boundary, internal/montecarlo, internal/analytics, dependency-direction, TASK-0024 |

## Context

TASK-0024 required a Monte Carlo bootstrap package. The question was where to place it in the dependency graph. The candidates were: attach to `internal/analytics`, attach to `internal/engine`, create `pkg/montecarlo`, or create `internal/montecarlo`.

Bootstrap only needs `[]model.Trade` as input and produces `BootstrapResult` as output. It has no need for engine state, portfolio state, or the analytics compute graph.

## Options considered

### Option A: Attach to `internal/analytics`
Add bootstrap functions directly into the analytics package.

- **Pros**: One less package; fewer import paths.
- **Cons**: Analytics is already the computation layer for Sharpe, Sortino, Calmar, TailRatio, drawdown. Adding simulation logic (PRNG, resampling) would muddy its responsibility. Analytics computes deterministic metrics from a fixed curve; bootstrap adds non-determinism (seeded but still structurally different). Hard to test in isolation.

### Option B: `pkg/montecarlo`
Export the package so external callers can use it.

- **Pros**: Usable from outside the module if we ever publish it.
- **Cons**: Bootstrap depends only on `pkg/model.Trade` — it's a backtesting-specific concern, not a general library. Exporting it creates a public surface we'd have to maintain. CLAUDE.md principle: no speculative abstraction.

### Option C: `internal/montecarlo` (chosen)
New package under `internal/`, importable only from within this module.

- **Pros**: Clean dependency graph. `internal/montecarlo` imports only `pkg/model`. `cmd/backtest` and `internal/output` can import it. No cycle. Can be tested in isolation. Can be evolved independently of analytics.
- **Cons**: One more package to navigate.

## Decision

`internal/montecarlo` as a standalone package. Imports only `pkg/model`. The dependency arrows are:

```
cmd/backtest        → internal/montecarlo  (calls Bootstrap())
internal/output     → internal/montecarlo  (reads BootstrapResult)
internal/montecarlo → pkg/model            (reads model.Trade)
```

This is a diamond (two paths into montecarlo), not a cycle. All dependency directions flow inward toward `pkg/model`. No engine coupling; no analytics coupling.

Key test: `internal/montecarlo` can be imported and tested without pulling in `internal/engine`, `internal/analytics`, or `internal/output`.

## Consequences

- Analytics package stays pure: deterministic metrics only.
- Bootstrap can evolve independently (e.g., add block bootstrap later) without touching analytics.
- `BootstrapResult` is an output-layer concern; `internal/output` imports it for formatting.
- If block-bootstrap on bar-level returns is added, it belongs in `internal/montecarlo` as a sibling function, not in analytics.

## Related decisions

- [Analytics `computeReturns` helper extracted](../architecture/2026-04-15-analytics-compute-returns-extracted-helper.md) — shows the analytics package's design philosophy; bootstrap is intentionally kept separate.
- [Bootstrap Sharpe non-annualized per-trade](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — the algorithm decision that drove the input type choice ([]Trade, not []EquityPoint).
