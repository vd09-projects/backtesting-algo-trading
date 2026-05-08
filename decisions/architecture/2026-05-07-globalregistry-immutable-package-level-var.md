# GlobalRegistry as immutable package-level var populated by buildRegistry(), not init()

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | strategy, registry, no-init, no-global-mutable-state, cmdutil, package-level-var, TASK-0079 |

## Context

`StrategyRegistry` needed to be accessible as a shared package-level value in `internal/cmdutil` without requiring callers to construct it themselves. Go provides several patterns for this: package-level `var` with explicit construction, `init()` auto-population, or `sync.Once`-based lazy init. The choice had to be compatible with CLAUDE.md's "no global mutable state, no `init()` with side effects" rule.

The `ist` timezone var in `pkg/strategy/session.go` (decision 2026-05-07-ist-package-level-var-time-fixedzone) established a working precedent: a package-level `var` populated by a pure constructor call, treated as read-only after initialization. The question was whether the same pattern extends to a heavier object (a registry containing 8 strategy factories).

## Options considered

### Option A: init() auto-population
Each `buildRegistry()` call inside an `init()` function populates `GlobalRegistry`.
- **Pros**: Familiar Go idiom for package-level singletons.
- **Cons**: Explicitly rejected by CLAUDE.md. `init()` side effects are non-obvious, execution order is non-deterministic across packages, errors from constructor calls cannot surface cleanly. The rule "no `init()` with side effects" exists for exactly this reason.

### Option B: sync.Once-based lazy initialisation
`GlobalRegistry` is populated on first access via `sync.Once`.
- **Pros**: Defers cost to first use; technically compatible with the no-global-mutable-state rule if the Once is used only once.
- **Cons**: Unnecessary complexity — the registry contains strategy factories that have trivially low construction cost. Lazy init exists for expensive resources (network connections, file I/O). Overhead here is adding two `sync` paths for zero benefit.

### Option C: var GlobalRegistry = buildRegistry() (chosen)
`buildRegistry()` is a pure function that creates a `StrategyRegistry`, populates it with all strategy factories, and returns it. The package-level var is assigned once by the initializer.
- **Pros**: Explicit — the construction site is readable at the point of declaration. No hidden side effects. Immutable after the var initializer returns. Safe for concurrent reads without synchronization. Compatible with the no-global-mutable-state rule (the rule targets vars mutated at runtime; this var is read-only after initialization). Follows the `ist` timezone precedent.
- **Cons**: `internal/cmdutil` must import all `strategies/` packages. This is the expected cost of centralisation (see StrategyRegistry decision).

## Decision

`var GlobalRegistry = buildRegistry()` in `internal/cmdutil/strategies.go`. `buildRegistry()` is a pure factory function — no side effects, no network calls, no I/O. The var is set once at package initialization and never mutated after. All reads from any goroutine are safe without synchronization.

This is not "global mutable state" under CLAUDE.md's intent. The rule targets vars that accumulate runtime mutations (loggers, caches, connection pools). A read-only registry set at compile-time is structurally identical to a constant — just heavier to initialize.

## Consequences

- The package initialization graph now includes `strategies/` → `internal/cmdutil` — if a strategy package ever needed to import cmdutil, a cycle would result. Currently no strategy package imports cmdutil; this constraint should be noted.
- `buildRegistry()` panics at startup if any default-param constructor call fails. This is intentional: if the default parameters for a built-in strategy are invalid, the binary should refuse to start rather than silently continuing with a broken registry.
- Any test that imports `internal/cmdutil` will implicitly load all strategy packages (via `GlobalRegistry` initialization). This is acceptable test overhead given the package count and constructor cost.

## Related decisions

- [IST timezone as package-level var](../convention/2026-05-07-ist-package-level-var-time-fixedzone.md) — establishes the precedent for package-level vars that are constructed once and never mutated
- [StrategyRegistry type in internal/cmdutil](../architecture/2026-05-07-strategy-registry-in-internal-cmdutil.md) — the registry type this var holds

## Revisit trigger

If the strategy constructor cost becomes non-trivial (e.g., loading a model from disk), revisit whether lazy init or explicit injection is warranted.
