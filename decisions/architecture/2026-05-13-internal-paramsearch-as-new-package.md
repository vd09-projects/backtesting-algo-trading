# internal/paramsearch as a new package for grid search logic

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | package-boundary, DSR, grid-search, testability, param-search, TASK-0077 |

## Context

`cmd/param-search` needed N-dimensional parameter grid search with per-variant DSR aggregation and InsufficientData filtering. The question was whether to build this logic inline in the cmd package or extract it to a new `internal/paramsearch` package.

The `cmd/evaluate` and `cmd/monitor` patterns put their business logic inline in cmd-layer functions — justified because the logic is single-caller and not independently testable in a meaningful way. But `cmd/param-search`'s core logic — Cartesian product enumeration, InsufficientData filtering, DSR aggregation — has testable invariants that are independent of the CLI wiring: DSR monotonicity, nTrials=gridSize correctness, sort order, and the insufficient-data exclusion logic.

## Options considered

### Option A: Logic inline in cmd/param-search
- **Pros**: Matches the cmd/evaluate and cmd/monitor precedents for single-caller cmd packages.
- **Cons**: Cartesian product and DSR aggregation logic are not testable without constructing full CLI flag state. The DSR aggregation correctness (which required a plan review blocking finding to catch) is the highest-risk part of this tool — it needs isolated tests.

### Option B: Extract to internal/paramsearch — chosen
- **Pros**: Core logic (grid iteration, filtering, DSR aggregation) is independently testable with a mock provider and scripted candle series. Matches `internal/sweep`, `internal/sweep2d`, `internal/universesweep`, `internal/walkforward` precedents — all tools that have meaningful algorithms extracted for isolated testing. Decision marks, testable invariants, and methodology documentation live in the package, not the cmd binary.
- **Cons**: One more package in `internal/`. Currently single-caller (`cmd/param-search` only).

## Decision

Option B. The testable invariants in the core logic — particularly DSR aggregation correctness and InsufficientData filtering — justify extraction. Matches `internal/sweep` and `internal/sweep2d` precedents exactly.

## Consequences

- `internal/paramsearch` is today single-caller. If a second cmd binary needs grid search (e.g., multi-universe param-search, strategy comparison tooling), it gets `internal/paramsearch` for free.
- `Config.StrategyFactory` is an exported function type — callers must honor the read-only params contract (params map is shared across concurrent goroutine calls for the same variant; factory must not mutate or retain it). This contract is documented in the field godoc (TASK-0112 follow-up).
- Test coverage: `internal/paramsearch` at 92.9%, covering Cartesian product, InsufficientData filtering, DSR sort order, factory error propagation, and config validation.

## Related decisions

- [internal/sweep package structure](../../architecture/2026-04-15-sweep-strategy-factory-func-type.md) — the precedent this follows
- [Strategy wiring fully centralized in internal/cmdutil](../../architecture/2026-05-08-strategy-wiring-fully-centralized-in-cmdutil.md) — strategy construction uses GlobalRegistry.Build in the cmd-layer factory closure, not in the internal package

## Revisit trigger

If a second cmd binary needs grid search, the API is stable enough to use without changes. If the StrategyFactory read-only contract causes a bug due to an external caller (not the single internal caller), move params copying into `runVariant` rather than requiring the caller to honor the contract.
