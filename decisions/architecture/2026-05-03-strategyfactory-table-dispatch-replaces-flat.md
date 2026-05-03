# strategyFactory dispatch table replaces flat switch in cmd/walk-forward

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | factory, dispatch-table, panic-free, cyclop, error-handling, cmd/walk-forward, TASK-0066 |

## Context

TASK-0066 introduced `cmd/walk-forward`, which needs to construct fresh strategy instances per fold via a factory closure. The initial implementation used a flat `switch` over strategy names inside `strategyFactory`, with individual `*Factory` helper functions (e.g., `smaFactory`, `rsiFactory`). Each helper returned a closure that called `New()` a second time at fold-execution time — meaning constructor errors (e.g., `fastPeriod >= slowPeriod`) could panic inside a fold goroutine mid-run. The flat switch also pushed cyclomatic complexity to 16, above the project's golangci-lint cyclop limit of 15.

## Decision

Replace the flat switch and per-strategy helper functions with a dispatch table: `map[string]strategyBuilder` where each entry is a `func(model.Timeframe, *strategyParams) (func() strategy.Strategy, error)`. Each builder validates parameters eagerly at startup (before any fold runs) and returns a clean closure on success, or a descriptive error on failure. The returned closure calls `New()` with already-validated parameters and uses an explicit `panic` with a hardcoded invariant message (not a `_` discard) on the unreachable second error — satisfying errcheck while documenting that the path should never be reached.

This pattern means: bad params → fail fast at CLI startup with a clean error; good params → closures that cannot panic at fold-execution time.

## Consequences

- Cyclomatic complexity drops from 16 to ~2 (dispatch table is a map lookup).
- All constructor error paths are tested at the validation step, not the closure step — the 0% coverage on panic branches is structurally eliminated.
- Adding a new strategy requires one entry in the dispatch table; no switch case, no separate helper function.
- The closure calls `New()` twice (once for validation, once per fold). For current constructors this is negligible; if a constructor becomes expensive, pre-constructing a config struct is the natural next step.

## Related decisions

- [Walk-forward accepts a single strategy instance, not a factory](../tradeoff/2026-04-22-walkforward-strategy-single-instance.md) — the factory pattern context this decision builds on
- [signalaudit uses StrategyFactory — no import of concrete strategy packages](../architecture/2026-05-01-signalaudit-strategy-factory-decoupling.md) — same pattern applied at the cmd layer boundary

## Revisit trigger

If a strategy constructor becomes expensive (e.g., loads a model from disk), the double-call pattern should be replaced with pre-constructing a config struct that `New()` accepts directly, so the closure performs no validation and cannot fail.
