# Sweep uses `StrategyFactory func(float64) (strategy.Strategy, error)` for parameterization

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | parameterization, strategy-factory, sweep, internal/sweep, TASK-0023 |

## Context

Building the parameter sweep runner (TASK-0023), the sweep package needs to construct a fresh strategy instance for each parameter value being tested. The sweep package must remain agnostic to which strategy is being swept — it should not import any concrete strategy type.

## Options considered

### Option A: `StrategyFactory func(float64) (strategy.Strategy, error)` (chosen)
- **Pros**: Sweep package stays unaware of concrete strategy types. Callers write a closure over the strategy constructor. Every concrete strategy is unchanged and unaware of sweep infrastructure.
- **Cons**: Caller must write a closure; marginally more setup code at the call site.

### Option B: `ParameterizableStrategy` interface with `WithParam(float64) Strategy`
- **Pros**: Explicit contract in the type system.
- **Cons**: Requires every concrete strategy to implement extra interface methods for a concern that belongs entirely to the sweep, not the strategy. Pollutes the Strategy interface for a single caller's need.

## Decision

`StrategyFactory func(float64) (strategy.Strategy, error)` is how the sweep constructs a strategy for each parameter value. The sweep package never knows the concrete strategy type; the caller writes a closure over the strategy constructor. The closure approach keeps every concrete strategy unchanged and unaware of the sweep infrastructure.

## Consequences

Callers must provide a factory closure, but this is minimal boilerplate and is isolated to the `cmd/sweep` entry point. Adding a new strategy to a sweep requires no changes to the sweep package itself.
