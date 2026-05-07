# Walk-forward Run() accepts a factory, not a single strategy instance

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | walkforward, strategy, factory, stateful-strategies, API-change, TASK-0059 |

## Context

`TimedExit` (added in TASK-0039) is the first `Strategy` implementation with mutable state — it tracks `entryBar` and `inPosition` between `Next()` calls. The walk-forward harness previously accepted a single `strategy.Strategy` instance, relying on all strategies being stateless. Using a shared `TimedExit` across parallel fold goroutines silently corrupts fold results: fold N's trailing position state bleeds into fold N+1, causing `barsSinceEntry` to be computed against a stale `entryBar` from the prior fold, which prevents the timer from ever firing in subsequent folds.

The revisit trigger in `decisions/tradeoff/2026-04-22-walkforward-strategy-single-instance.md` ("when the first mutable-state strategy is added") fired with TASK-0039.

## Decision

`internal/walkforward.Run()` accepts `factory func() strategy.Strategy` instead of a single `strategy.Strategy` instance. `runFold()` calls `factory()` twice per fold — once for the IS run and once for the OOS run — ensuring each engine run starts with a fresh strategy instance regardless of whether the strategy carries inter-bar mutable state.

Stateless strategy callers migrate trivially: `myStrategy` becomes `func() strategy.Strategy { return myStrategy }`. Stateful callers like `TimedExit` construct a new instance on each call.

## Consequences

- Silent cross-fold state corruption is eliminated at the API level — it is no longer possible to pass a shared stateful instance.
- Stateless strategy callers pay a trivial closure allocation per fold run (negligible).
- All existing `cmd/walk-forward` callers already use the factory builder pattern (dispatch table of `strategyBuilder` functions), so the API change was absorbed without behavioral change at the CLI layer.
- Adds a fold-isolation regression test (`TestRun_TimedExitFoldStateIsolation`) that would fail if a single shared `TimedExit` were passed — documents the broken behavior the factory prevents.

## Related decisions

- [Walk-forward accepts a single strategy instance, not a factory](../tradeoff/2026-04-22-walkforward-strategy-single-instance.md) — superseded by this decision; revisit trigger now resolved
- [TimedExit statefulness — first stateful wrapper in pkg/strategy](../convention/2026-04-27-timed-exit-statefulness-pkg-strategy.md) — the trigger for this change

## Revisit trigger

If a strategy constructor becomes expensive (e.g., loads a model from disk), the double-call pattern (once for validation, once per fold run) should be replaced with pre-constructing a config struct that `New()` accepts, so the factory closure performs no validation and cannot fail.
