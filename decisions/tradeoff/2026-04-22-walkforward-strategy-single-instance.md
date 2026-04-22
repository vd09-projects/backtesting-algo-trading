# Walk-forward accepts a single strategy instance, not a factory

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | walkforward, strategy, concurrency, API, factory, TASK-0022 |

## Context

`Run()` runs folds in parallel via errgroup. Each fold calls `strategy.Next()` concurrently across folds. The question was whether `Run()` should accept a single `strategy.Strategy` instance (callers ensure concurrent safety) or a factory `func() strategy.Strategy` (each fold gets a fresh copy).

## Decision

`Run()` accepts a single `strategy.Strategy` instance. All current strategies are stateless (no mutable fields between `Next()` calls), so concurrent access is safe. A factory API adds overhead and API complexity for a problem that doesn't yet exist. The concurrency assumption is documented in the `Run()` godoc: callers are responsible for ensuring the provided strategy is safe for concurrent use, or must pass distinct instances per fold. If stateful strategies are added, the signature changes to `func() strategy.Strategy`.

## Consequences

Any future strategy with mutable state (e.g., a running EMA buffer updated inside `Next()`) must either be made concurrent-safe or the `Run()` API must be updated to accept a factory. This is a known future work item if stateful strategies are introduced.

## Revisit trigger

When the first mutable-state strategy is added to `strategies/`.
