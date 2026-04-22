# `errgroup` with GOMAXPROCS ceiling for universe instrument fan-out

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | concurrency, errgroup, parallelism, GOMAXPROCS, universesweep, TASK-0035 |

## Context

The universe sweep runs one engine per instrument. The standing order for the parameter sweep (`internal/sweep`) chose sequential execution because `golang.org/x/sync` wasn't in `go.mod` and parameter steps can sometimes have ordering dependencies. Neither constraint applies here: `x/sync` is already in `go.mod` (added for walk-forward), and instrument runs are fully independent.

## Decision

`errgroup.WithContext` with `g.SetLimit(runtime.GOMAXPROCS(0))`. Each goroutine writes to `results[i]` at a fixed pre-allocated index (see [pre-allocated fixed-index writes](2026-04-22-pre-allocated-fixed-index-writes-for-determinism.md)).

The GOMAXPROCS ceiling prevents spawning N goroutines for a 500-stock universe on a 4-core machine — each goroutine holds a full candle series in memory.

## Consequences

The standing order for `internal/sweep` (sequential) does not apply here — that decision was driven by a missing dependency and a different concurrency model. Universe sweep has no ordering dependency between instruments. The errgroup pattern is consistent with `internal/walkforward`, which runs fold windows in parallel for the same reason.

## Revisit trigger

If a stateful strategy is ever added whose `Next()` call mutates internal state, all concurrent goroutines sharing a single strategy instance will race. The fix is the factory pattern (same trigger condition as in `internal/walkforward`).
