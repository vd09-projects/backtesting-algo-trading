# Sweep executes parameter steps sequentially — no `errgroup` parallelism

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | concurrency, sequential, errgroup, dependencies, internal/sweep, TASK-0023 |

## Context

The parameter sweep runs one full backtest per parameter value. For a 20–50 step sweep on daily bars, the execution time is bounded and small. Adding parallelism requires `golang.org/x/sync` (for `errgroup`), which is not in `go.mod`. CLAUDE.md explicitly prohibits new dependencies without discussion.

## Options considered

### Option A: Sequential execution (chosen)
- **Pros**: No new dependencies. Simple, readable for-loop. For the realistic sweep sizes (20–50 steps, daily bars), completes in seconds.
- **Cons**: Does not exploit multi-core for larger sweeps.

### Option B: Parallel execution via `errgroup`
- **Pros**: Faster for large sweeps or high-frequency bars.
- **Cons**: Requires `golang.org/x/sync` (a new dependency). Adds complexity. Not needed for current use cases.

## Decision

Sequential execution rather than parallel via `errgroup`. The constraint is `golang.org/x/sync` not being in `go.mod` and CLAUDE.md's no-new-dependencies rule. For 20–50 step single-parameter sweeps on daily bars, sequential completes in seconds.

## Consequences

The upgrade path to parallel is localized: replace the for-loop with an errgroup pool, write results into a pre-sized slice at fixed indices. The API (`Run` signature and `SweepReport` return type) does not change when parallelism is added.

## Revisit trigger

If sweeps routinely exceed 100 steps, or if intraday bars make sweep runtime material (>30 seconds), revisit and add `golang.org/x/sync` with discussion.
