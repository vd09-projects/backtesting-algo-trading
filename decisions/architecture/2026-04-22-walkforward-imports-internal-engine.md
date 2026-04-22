# internal/walkforward imports internal/engine (orchestration harness, not stats package)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | walkforward, package-boundary, engine-dependency, montecarlo, TASK-0022 |

## Context

`internal/montecarlo` imports only `pkg/model` — it resamples a trade list and is a pure statistics package with no engine dependency. When designing `internal/walkforward`, the question was whether it should follow the same narrow dependency pattern.

## Decision

`internal/walkforward` imports `internal/engine`, `pkg/model`, `pkg/strategy`, and `pkg/provider`. This is correct and intentional. Walk-forward is not a statistics package — it is an orchestration harness that runs the engine repeatedly across fold windows. The engine dependency is the whole point: each fold executes a complete engine run. The montecarlo analogy breaks down because montecarlo operates on an already-computed trade list; walk-forward computes the trade lists by running the engine.

## Consequences

`internal/walkforward` sits at a higher layer in the dependency graph than `internal/montecarlo`. Both are in `internal/` but they are not peers in terms of abstraction level. Any future "run something N times over data windows" harness should similarly import the engine rather than attempting to operate at the trade-list level.
