# param-search concurrency: outer variants sequential, inner instruments parallel

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | concurrency, errgroup, grid-search, GOMAXPROCS, param-search, TASK-0077 |

## Context

`internal/paramsearch.Run` must execute a (variants × instruments) matrix of engine runs. Two parallelism strategies were considered: outer-parallel (variants run concurrently, each spawning inner goroutines) and outer-sequential (variants run one at a time, each spawning inner goroutines for its instruments).

## Options considered

### Option A: Outer-parallel (variants parallel, instruments parallel)
- **Pros**: Maximum theoretical throughput — all engine runs execute concurrently.
- **Cons**: `gridSize × instruments` goroutines launched simultaneously. With a 200-variant grid and 15 instruments, that's 3,000 concurrent goroutines, each holding a full candle series in memory. Memory blow-up is the real risk. Requires a 2D pre-allocated result matrix indexed by (variant, instrument) to avoid shared writes — significantly more complex than the 1D pattern established by universesweep and sweep2d.

### Option B: Outer-sequential, inner-parallel — chosen
Outer loop over variants runs sequentially. For each variant, a single errgroup fans out over instruments with `GOMAXPROCS` ceiling, matching the universesweep.Run pattern.
- **Pros**: Memory bounded — at most `GOMAXPROCS × one candle series` live at any time per the inner errgroup ceiling. Code follows the established universesweep/walkforward pattern: pre-allocated 1D result slice, index-owned-per-goroutine, no mutex. Simple to reason about.
- **Cons**: Variants cannot benefit from parallelism across each other. Total wall time for a 200-variant grid is 200 × (time to run 15 instruments in parallel). For most practical grids, the dominant cost is network fetch per instrument; the outer loop adds only the sequential overhead of launching each inner errgroup.

## Decision

Option B. The network fetch cost dominates; outer sequential / inner parallel is sufficient to saturate cores. The complexity reduction (no 2D pre-allocation, bounded memory) is significant. Matches the established pattern in `universesweep.Run` and `internal/walkforward/Run`.

## Consequences

- Wall time scales linearly in `gridSize` for fixed instrument set and timeframe. For a 50-variant grid with 15 instruments, this is acceptable on a developer machine.
- If a very large grid (500+ variants) requires faster turnaround, a bounded outer parallel loop could be added (outer errgroup with GOMAXPROCS ceiling, 2D grid pre-allocated). That is an optimization, not a correctness change.
- `gctx` from the inner errgroup is not propagated to the outer loop — if one instrument fails within a variant, that variant's inner errgroup aborts but the outer loop receives the error and stops the entire search. This is the correct semantics: a provider error mid-search should abort rather than silently skip a variant.

## Revisit trigger

If typical grid sizes exceed 200 variants and wall time becomes a bottleneck, consider outer-parallel with a bounded errgroup and a 2D results grid. Benchmark first; the inner parallelism already saturates cores and outer parallelism may not reduce wall time proportionally.
