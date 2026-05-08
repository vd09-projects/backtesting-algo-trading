# cmd/walk-forward retains local dispatch table for per-fold factory construction

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | superseded       |
| Category | architecture     |
| Tags     | strategy, factory, walk-forward, per-fold, local-dispatch, cmdutil, TASK-0079 |

## Context

`cmd/walk-forward` has a different strategy dispatch requirement from other cmd binaries. Where `cmd/backtest` and `cmd/universe-sweep` need a single `strategy.Strategy` instance constructed with parsed CLI params, `cmd/walk-forward` needs a zero-arg `func() strategy.Strategy` closure — a per-fold factory that captures parsed params and produces a fresh strategy instance on each call. This is the factory-per-fold API established in decision 2026-05-07-walkforward-strategy-factory-per-fold.

Before TASK-0079, `cmd/walk-forward` had a `var strategyRegistry = map[string]strategyBuilder{...}` that served as both the name-validation authority and the construction dispatch. After centralization, the name-validation authority moves to `cmdutil.GlobalRegistry` but the per-fold construction logic must remain local.

## Options considered

### Option A: Move per-fold factory logic into cmdutil.GlobalRegistry
Register per-fold factories (parameterized closures) directly in `GlobalRegistry` rather than zero-arg default-param factories.
- **Pros**: Single location for all construction logic.
- **Cons**: Per-fold factories are parameterized — they need CLI params that are only known at cmd startup. `GlobalRegistry` would need to accept params somehow, changing its API from `func() strategy.Strategy` to `func(params) func() strategy.Strategy`. This breaks the clean zero-arg factory contract that `ListStrategies` and `MustGet` are designed around.

### Option B: Retain local dispatch table, delegate name validation to GlobalRegistry (chosen)
Keep the local `localBuilders map[string]strategyBuilder` for per-fold construction logic. Call `cmdutil.GlobalRegistry.MustGet(name)` at the top of `strategyFactory` for name validation; then look up the local builder.
- **Pros**: Clean separation: GlobalRegistry owns the authoritative name list and default factories; local map owns the per-fold construction logic. Name validation is centralised (unknown names panic via MustGet); construction is local. The local map no longer claims to be the source of truth for known names.
- **Cons**: Two maps must be kept in sync: if a strategy is added to `GlobalRegistry` but not to `localBuilders`, `MustGet` succeeds but the local lookup fails with an "unreachable" error. This is a maintenance gap, but a less dangerous one than the original problem (silent unavailability).

## Decision

Option B: `localBuilders` map renamed from `strategyRegistry` to make explicit that it is not the name-validation authority. `strategyFactory` calls `cmdutil.GlobalRegistry.MustGet(name)` first (panics on unknown name), then looks up `localBuilders[name]` for the actual per-fold factory builder. If `MustGet` succeeds but `localBuilders` lookup fails, the code returns `fmt.Errorf("unknown strategy %q", name)` — a path that should be unreachable if both maps are kept in sync.

The `localBuilders` and `GlobalRegistry` must have the same strategy names. The "unreachable" comment at the default case serves as the sync reminder.

## Consequences

- When a new strategy is added to `GlobalRegistry` in `internal/cmdutil/strategies.go`, a corresponding entry must be added to `localBuilders` in `cmd/walk-forward/main.go`. The build compiles without it, but the runtime will error. This is a weaker guarantee than a compile-time check.
- A future improvement could validate at startup that `localBuilders` covers all names in `GlobalRegistry`; currently no such check exists.

## Related decisions

- [Walk-forward Run() accepts a factory, not a single strategy instance](../architecture/2026-05-07-walkforward-strategy-factory-per-fold.md) — established why per-fold factories are needed
- [strategyFactory dispatch table replaces flat switch in cmd/walk-forward](../architecture/2026-05-03-strategyfactory-table-dispatch-replaces-flat.md) — the dispatch table this decision partially evolves
- [StrategyRegistry type in internal/cmdutil](../architecture/2026-05-07-strategy-registry-in-internal-cmdutil.md) — the central registry that now owns name validation
- [Strategy wiring fully centralized in internal/cmdutil](../architecture/2026-05-08-strategy-wiring-fully-centralized-in-cmdutil.md) — supersedes this decision; localBuilders eliminated, WalkForwardFactory added to registry

## Revisit trigger

Superseded 2026-05-08. See 2026-05-08-strategy-wiring-fully-centralized-in-cmdutil.
