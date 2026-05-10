# Walk-forward dispatch sequential per instrument in cmd/evaluate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | concurrency, walk-forward, orchestrator, sequential, errgroup, cmd/evaluate, TASK-0073 |

## Context

`cmd/evaluate`'s stage 2 (`runWalkForward`) dispatches `walkforward.Run` for each universe survivor sequentially. An alternative would be to fan out the per-instrument WF runs in parallel (e.g., via errgroup) at the orchestrator level.

## Options considered

### Option A: Parallel WF dispatch via errgroup at orchestrator level
- **Pros**: Faster wall-clock time when N survivors > GOMAXPROCS.
- **Cons**: `walkforward.Run` already fans out its folds internally via errgroup with a GOMAXPROCS ceiling. Outer parallelism doubles the goroutine count without halving runtime on CPU-bound fold processing. Progress logs become interleaved and harder to read.

### Option B: Sequential dispatch per instrument (chosen)
- **Pros**: Clean progress logging (`evaluate: WF NSE:SBIN — OverfitFlag=false`). Each `walkforward.Run` call saturates available CPUs internally. Simpler code. Matches the pattern in `cmd/walk-forward`.
- **Cons**: Slightly slower if N survivors is large and fold count is small — the rare case.

## Decision

Sequential per-instrument dispatch. `runWalkForward` iterates `instruments` with a plain `for` loop; each `walkforward.Run` call handles its own internal concurrency. Bootstrap stage (`runBootstrap`) follows the same sequential pattern.

## Consequences

For typical evaluation runs (5–50 survivors, 4–6 folds per instrument), sequential outer dispatch with parallelism inside `walkforward.Run` saturates CPUs effectively. The tradeoff is acceptable for a research tool where run time is measured in minutes, not milliseconds.

## Revisit trigger

If the evaluation pipeline is used with very large universes (100+ instruments) and consistently shows low CPU utilization due to sequential WF dispatch, add outer errgroup with a goroutine limit to cap concurrent WF runs.
