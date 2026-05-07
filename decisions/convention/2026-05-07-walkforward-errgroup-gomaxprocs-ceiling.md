# errgroup parallelism ceiling = GOMAXPROCS for walk-forward fold fan-out

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | walkforward, concurrency, errgroup, parallelism, GOMAXPROCS, memory, TASK-0059 |

## Context

`internal/walkforward.Run()` fans out fold goroutines via `errgroup`. Each fold goroutine runs two engine passes (IS + OOS), each holding a full candle series in memory for the fold's date window. The question was whether to fan out unboundedly (one goroutine per fold) or cap parallelism.

## Decision

`errgroup.SetLimit(runtime.GOMAXPROCS(0))`. The ceiling bounds the number of concurrently executing folds to the number of logical CPUs.

Rationale: each fold goroutine holds two full candle series in memory (IS + OOS). For daily-bar strategies over a 2yr IS + 1yr OOS window this is ~1,095 candles — trivially small. The ceiling is a guard against future high-frequency timeframes (e.g., 5-min bars over 2yr IS = ~26,000 candles × 2 runs × N folds) where unbounded fan-out could exhaust heap. The ceiling is conservative by default and costs nothing on current daily-bar workloads.

## Consequences

- Memory usage is bounded to GOMAXPROCS × (candle series size per fold), regardless of fold count.
- At GOMAXPROCS = 8 with 5-min daily-bar candle series (~1MB per fold run), peak memory is ~16MB — trivial. The guard matters at scale.
- The same pattern is used in `internal/universesweep` for instrument fan-out.

## Related decisions

- [errgroup with GOMAXPROCS ceiling for universe instrument fan-out](../tradeoff/2026-04-22-errgroup-universe-fan-out-gomaxprocs-ceiling.md) — same pattern, different scope (universe sweep vs. walk-forward folds)
- [IS and OOS sequential within a fold](../convention/2026-05-07-walkforward-runfold-is-oos-sequential.md) — why fold-internal concurrency is not added on top of this ceiling
