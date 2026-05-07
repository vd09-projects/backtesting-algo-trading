# IS and OOS runs within a single walk-forward fold execute sequentially, not in parallel

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | walkforward, concurrency, fold-internal-sequencing, errgroup, IS-OOS, TASK-0059 |

## Context

`internal/walkforward.runFold()` executes two engine runs per fold: an in-sample (IS) run and an out-of-sample (OOS) run. These two runs are independent — IS and OOS windows do not overlap and neither depends on the other's output. The question was whether to run them concurrently (nested goroutines within the fold goroutine) or sequentially.

## Decision

IS and OOS runs within a fold execute sequentially. Each fold goroutine is already one unit of parallelism in the outer `errgroup` (folds fan out across available CPUs). Adding a second goroutine inside each fold would double the concurrency count without halving the wall-clock runtime, because:

1. The outer errgroup already saturates available CPU cores with fold goroutines (ceiling = GOMAXPROCS).
2. Candle processing in the engine is CPU-bound, not I/O-bound. Nested goroutines add scheduling overhead without freeing a CPU for useful work.
3. Sequential IS → OOS ordering is natural and readable; the IS result is not inspected before running OOS, so there is no logical dependency between them.

Nested errgroups are avoided entirely; the code stays flat.

## Consequences

- Code remains simple: one goroutine per fold, two sequential engine.Run() calls inside.
- No nested synchronization primitives to reason about.
- If a future fold ever needs to parallelize IS and OOS (e.g., for very large candle series where I/O dominates), adding an inner errgroup is straightforward.

## Related decisions

- [errgroup with GOMAXPROCS ceiling for walk-forward fold parallelism](../convention/2026-05-07-walkforward-errgroup-gomaxprocs-ceiling.md) — outer parallelism model that makes inner parallelism unnecessary
