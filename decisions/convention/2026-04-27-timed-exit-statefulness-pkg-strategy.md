# TimedExit statefulness — first stateful wrapper in pkg/strategy, not concurrent-safe

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | pkg/strategy, TimedExit, statefulness, concurrency, walk-forward, TASK-0039, TASK-0059 |

## Context

`TimedExit` (added in TASK-0039) is a strategy wrapper that forces a sell after `maxHoldBars` bars regardless of the inner strategy's signal. Unlike all previous `Strategy` implementations — which are stateless pure functions over a candle slice — `TimedExit` must track position state between `Next()` calls to compute `barsSinceEntry`.

The `internal/walkforward` harness accepts a single `strategy.Strategy` instance and passes it across parallel fold executions, documenting (in `decisions/tradeoff/2026-04-22-walkforward-strategy-single-instance.md`) that this is safe because all existing strategies are stateless. `TimedExit` invalidates that assumption.

## Options considered

### Option A: Stateless wrapper — recompute position from candle history (rejected)

Infer entry state by scanning all candles passed to `Next()` each call, looking for past Buy signals from inner.

- **Pros**: No mutable fields; safe for concurrent use without any caller changes.
- **Cons**: Requires the wrapper to re-run the inner strategy on all prior candles every bar — O(n²) in candle count, which is unacceptable in the hot loop. Also requires the inner strategy to be pure and side-effect-free across repeated calls with the same inputs, which is not guaranteed by the `Strategy` interface contract.

### Option B: Stateful wrapper — track entryBar and inPosition fields (chosen)

Maintain `entryBar int` and `inPosition bool` as struct fields, updated on each `Next()` call.

- **Pros**: O(1) per call. Correct and simple. Matches the obvious mental model.
- **Cons**: Not safe for concurrent use across walk-forward folds. Callers must construct a fresh `TimedExit` per fold. The `internal/walkforward` factory-API change is deferred to TASK-0059.

## Decision

Stateful wrapper (Option B). `TimedExit` maintains `entryBar int` and `inPosition bool` as mutable struct fields. These are reset to zero values when a position is closed (by timer or inner sell signal).

The concurrent-safety implication is documented in the type's godoc and in this decision. TASK-0059 tracks the follow-up change to `internal/walkforward.Run()` to accept a `func() strategy.Strategy` factory — the revisit trigger called out in `decisions/tradeoff/2026-04-22-walkforward-strategy-single-instance.md` is now active.

## Consequences

- `TimedExit` is the first `Strategy` implementation in `pkg/strategy` with mutable state.
- A single `TimedExit` instance **must not** be shared across concurrent walk-forward folds. Callers that do so will see fold N's trailing position state bleed into fold N+1, silently corrupting the results.
- Until TASK-0059 lands, `TimedExit` cannot safely be used with `internal/walkforward.Run()` in its current single-instance form.
- The `NewTimedExit` godoc documents the concurrency constraint explicitly.

## Related decisions

- [Walk-forward accepts a single strategy instance, not a factory](../tradeoff/2026-04-22-walkforward-strategy-single-instance.md) — the revisit trigger in that decision fires here; TASK-0059 is the follow-up
- [Strategy.Lookback() as a first-class interface method](../architecture/2026-04-02-strategy-lookback-as-interface-method.md) — the interface `TimedExit` satisfies; interface itself is unchanged

## Revisit trigger

Resolved when TASK-0059 lands: `internal/walkforward.Run()` accepts `func() strategy.Strategy` and each fold constructs a fresh instance. At that point this convention note becomes informational — the harness enforces safety automatically.
