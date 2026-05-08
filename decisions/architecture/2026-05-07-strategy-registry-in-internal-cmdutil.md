# StrategyRegistry type in internal/cmdutil — centralized strategy name registration

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | strategy, registry, cmdutil, DRY, cmd-layer-plumbing, TASK-0079 |

## Context

Every cmd binary that dispatches on a strategy name (`cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`, `cmd/sweep`) maintained its own local switch statement or map. Adding a new strategy required updating 4+ files. Missing any one registration produced silent wrong behaviour — the strategy would be unavailable with no compile-time signal. With 8 strategies registered and 2 more incoming (TASK-0074, TASK-0075), the maintenance tax was compounding.

The existing `internal/cmdutil` package already housed shared cmd-layer plumbing: `BuildProvider`, `ParseCommissionModel`, `DefaultOutPath`. The three-copy extraction rule (established in the 2026-04-22 buildProvider decision) was well-exceeded.

## Options considered

### Option A: Keep per-binary switch statements, improve documentation
No change to code structure; document the requirement to update all four binaries when adding a strategy.
- **Pros**: Zero implementation cost.
- **Cons**: Documentation doesn't prevent mistakes. Silent missing-strategy bugs are exactly the kind that survive documentation and manifest in production runs. Maintenance tax compounds with every new strategy.

### Option B: init()-based auto-registration in each strategies/ package (chosen approach in some ecosystems)
Each strategy package registers itself via `init()`, so no central file needs updating.
- **Pros**: Adding a new strategy requires zero changes outside the strategy package.
- **Cons**: Explicitly rejected by CLAUDE.md no-global-state rule. `init()` side effects make the registration order non-deterministic and the dependency graph opaque. Circular import risk between `strategies/` and `internal/cmdutil`.

### Option C: StrategyRegistry type in internal/cmdutil with explicit registration (chosen)
A `StrategyRegistry` struct with `Register`, `MustGet`, and `ListStrategies` methods, populated by an explicit `buildRegistry()` function in `internal/cmdutil/strategies.go`.
- **Pros**: Single file to edit when adding a new strategy. No `init()`. Compile-time safe — the registry is populated at package init, before any cmd logic runs. `MustGet` panics on unknown names at startup (fail-fast). Follows established cmdutil precedent.
- **Cons**: Strategy packages are still imported in `internal/cmdutil/strategies.go` — cmdutil now has strategy dependencies. This is acceptable: cmdutil is cmd-layer plumbing, not a pure utility package.

## Decision

Option C: `StrategyRegistry` struct in `internal/cmdutil/registry.go` with exported `Register`, `MustGet`, and `ListStrategies` methods. The `GlobalRegistry` package-level var in `internal/cmdutil/strategies.go` is the single authoritative source. Each cmd binary calls `GlobalRegistry.MustGet(name)` for name validation at startup, then constructs the strategy with its parsed CLI params.

The cmdutil home is correct: `internal/cmdutil` is already established as the package for shared cmd-layer plumbing that multiple binaries consume. `StrategyRegistry` is of that same character.

## Consequences

- Adding a new strategy now requires exactly one file change: an entry in `internal/cmdutil/strategies.go`.
- `internal/cmdutil` now imports all `strategies/` packages — it is no longer a zero-strategy-dependency package. This is the expected cost of centralisation.
- `cmd/walk-forward` retains a local `localBuilders` map for per-fold factory construction (see the companion decision on local dispatch table). Name validation is delegated to `GlobalRegistry.MustGet`; the local map handles construction only.
- `cmd/sweep` has a latent inconsistency: `GlobalRegistry.MustGet` accepts `cci-mean-reversion` and `stub` but the local `factoryRegistry` switch has no case for them — they fall through to a default error. This is a pre-existing gap; TASK-0061 will resolve it.

## Related decisions

- [buildProvider extracted to internal/cmdutil](../architecture/2026-04-22-buildprovider-extracted-to-cmdutil.md) — established the three-copy rule and cmdutil as home for shared cmd-layer plumbing
- [ParseCommissionModel extracted to internal/cmdutil](../convention/2026-04-29-parse-commission-model-extracted-to-cmdutil.md) — same extraction pattern applied at two callers; cmdutil precedent solidified
- [strategyFactory dispatch table replaces flat switch in cmd/walk-forward](../architecture/2026-05-03-strategyfactory-table-dispatch-replaces-flat.md) — the local dispatch table this decision partially supersedes; local construction logic retained, name validation centralised
- [GlobalRegistry as immutable package-level var](../architecture/2026-05-07-globalregistry-immutable-package-level-var.md) — companion decision on how GlobalRegistry is populated

## Revisit trigger

If `strategies/` is restructured under `internal/` (removing the cross-boundary import concern), or if a third cmd layer type emerges that makes cmdutil the wrong home — revisit the placement.
