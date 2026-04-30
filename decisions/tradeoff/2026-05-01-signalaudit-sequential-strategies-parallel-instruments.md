# signalaudit: strategies run sequentially, instruments fan out in parallel

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | signalaudit, concurrency, errgroup, zerodha-api, rate-limit, TASK-0050 |

## Context

`internal/signalaudit.Run()` loops over 6 strategies × 15 instruments = 90 (strategy, instrument) engine runs. Two parallelism models were available: (a) run all 90 concurrently, (b) run strategies sequentially with instruments fanned out per strategy.

## Options considered

### Option A: Full parallel (all 90 concurrent)
- **Pros**: Maximum throughput in theory.
- **Cons**: 90 concurrent goroutines each holding a Zerodha API connection on a 4–8 core machine; Zerodha rate-limits historical data requests. Likely to produce throttle errors and degrade performance relative to sequential.

### Option B: Sequential strategies, parallel instruments (chosen)
- **Pros**: Each strategy's 15-instrument fan-out gets full hardware utilisation (GOMAXPROCS ceiling via errgroup). Only 15 concurrent Zerodha requests at a time — well within rate limits. Deterministic execution order simplifies debugging.
- **Cons**: Total wall time is 6× the time for one strategy's sweep rather than theoretically 1×.

## Decision

Strategies are run sequentially — one `universesweep.Run()` call per strategy. Within each strategy, instruments fan out in parallel via the existing errgroup inside `universesweep.Run()`. This gives full hardware utilisation on the inner loop without multiplying Zerodha API pressure by 6.

## Consequences

- Wall time for a full 6-strategy audit is ~6× the single-strategy instrument sweep time. Acceptable for a one-off or periodic run.
- If Zerodha ever raises rate limits substantially or we move to a local data cache, Option A becomes worth revisiting.

## Revisit trigger

If Zerodha API rate limits are relaxed, or if data is served from a local cache (no HTTP), full parallelism becomes safe to try.
