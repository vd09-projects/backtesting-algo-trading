# Pre-allocated fixed-index slice writes for deterministic goroutine output

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | convention       |
| Tags     | determinism, goroutine, slice, concurrency, universesweep, TASK-0035 |

## Context

`universesweep.Run` fans out instrument runs via errgroup. The results must be collected into a slice after all goroutines complete. The collection mechanism must be race-free and produce deterministic pre-sort ordering (same instruments → same order before Sharpe sort).

## Decision

Pre-allocate `results := make([]Result, len(cfg.Instruments))` before spawning goroutines. Goroutine `i` captures `i := i` (loop variable capture) and writes `results[i]`. No mutex needed — each goroutine owns its own index, so there are no concurrent writes to the same slot.

After `g.Wait()`, `results` is fully populated. Sort descending by Sharpe. The pre-sort order is input order (same instruments slice → same indices → same results slice ordering before sort). This property makes the output deterministic: two runs with the same inputs produce bit-identical CSV files.

## Consequences

This pattern is established in `internal/walkforward` (same pattern, same reason). Future harnesses that fan out over a fixed list of independent items should use this pattern. The alternative — a channel or mutex-protected append — is either harder to reason about or introduces non-deterministic ordering.
