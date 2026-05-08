# Strategy wiring fully centralized in internal/cmdutil — Build, WalkForwardFactory, SweepFactory

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-08       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | strategy, registry, cmdutil, DRY, cmd-layer-plumbing, single-registration-point, StrategyEntry, Build, WalkForwardFactory, SweepFactory, RegisterFlags |

## Context

After TASK-0079, `StrategyRegistry` centralized strategy name validation (`MustGet`, `ListStrategies`) and zero-arg default factories. Each cmd binary — backtest, universe-sweep, walk-forward, sweep — still duplicated its own strategy construction logic: identical `strategyParams` structs, identical `buildStrategy` switch statements, and per-strategy flag declarations in every `main()`. The `cci-mean-reversion` omission from `cmd/sweep` was discovered mid-session and fixed manually — clear evidence that the per-cmd duplication was producing silent gaps.

The user observed this pattern and asked for a solution covering all cmd mains at once, not just sweep.

## Options considered

### Option A: Checklist / contributing guide
Document a step-by-step "adding a new strategy" checklist. Maintain per-cmd construction switches.
- **Pros**: Zero code change.
- **Cons**: Already proven not to work — `cci-mean-reversion` was added to `GlobalRegistry` and missed in `cmd/sweep` despite the switch existing. Checklists are not enforced.

### Option B: Sweep-only registration (first plan)
Apply a `strategyRegistration` pattern with `paramDef` slices only to `cmd/sweep`, eliminating its `fixedParams` struct and per-strategy factory functions.
- **Pros**: Smaller change; bounded to sweep.
- **Cons**: Other cmd mains still duplicate construction. Adding a strategy still requires touching backtest, universe-sweep, walk-forward, and sweep separately.

### Option C: Full centralization in cmdutil (chosen)
Extend `StrategyEntry` in `internal/cmdutil/registry.go` to carry `Params []ParamDef`, `Build func(tf, params) (Strategy, error)`, and `SweepParams []SweepParamDef`. Add `Build`, `WalkForwardFactory`, `SweepFactory`, `RegisterFlags`, `BuildParamMap`, `DefaultParams`, `ParamsMap` methods to `StrategyRegistry`. All strategy-package imports and construction logic move to `internal/cmdutil/strategies.go`.
- **Pros**: Adding a new strategy = add one `StrategyEntry` block to `strategies.go`. No other file changes. Compile-enforced — cmd mains can't call construction logic that doesn't exist in the registry.
- **Cons**: `internal/cmdutil/strategies.go` becomes a denser file. `cmd/signal-audit` and `cmd/sweep2d` excluded (signal-audit uses audit-tuned non-default params; sweep2d has 2-param joint factory that doesn't fit the 1-param sweep shape).

## Decision

Option C. `StrategyEntry` extended with `Params []ParamDef`, `Build`, and `SweepParams []SweepParamDef`. New registry methods: `Build`, `WalkForwardFactory` (validates eagerly then returns zero-arg factory per fold), `SweepFactory` (one swept param, rest fixed), `RegisterFlags` (registers all strategy params on a `*flag.FlagSet`), `BuildParamMap`, `DefaultParams`, `ParamsMap`. All strategy-package imports (`strategies/smacrossover`, `strategies/rsimeanrev`, etc.) removed from cmd mains and consolidated in `strategies.go`.

Flag names standardized: `--cci-entry`/`--cci-exit` renamed to `--cci-entry-threshold`/`--cci-exit-threshold` across all cmd mains to match the more descriptive names already in `cmd/sweep`. This is a CLI-breaking change for scripts using the old names.

`cmd/sweep2d` and `cmd/signal-audit` deliberately excluded. sweep2d has a 2-parameter joint factory (`func(float64, float64)`) that doesn't compose with the 1-param `SweepParamDef`. signal-audit uses audit-tuned parameter values (not defaults) verified by Marcus for signal frequency — auto-population from defaults would silently change audit results.

## Consequences

- Adding a strategy requires exactly one block in `internal/cmdutil/strategies.go`. No other files need changing.
- `cmd/signal-audit/allStrategyFactories()` remains manually maintained. A test (`TestSignalAuditCoversAllStrategies`) should be added to enforce that all `GlobalRegistry` strategies (minus stub) appear in the audit list — this is a follow-up task.
- `--cci-entry`/`--cci-exit` flags are gone. Scripts using them must migrate to `--cci-entry-threshold`/`--cci-exit-threshold`.
- `cmd/walk-forward`'s `localBuilders` map, `strategyParams` struct, and all `buildXxxFactory` functions are gone. `WalkForwardFactory` replaces them: validates params eagerly, returns zero-arg closure for per-fold construction.
- `ParamsMap` uses flag name keys (kebab-case) in output JSON metadata, not the old snake_case keys (`fast_period` → `fast-period`). Minor format change in run output files.

## Related decisions

- [StrategyRegistry type in internal/cmdutil](../architecture/2026-05-07-strategy-registry-in-internal-cmdutil.md) — earlier decision this extends; that decision covered name registration only
- [cmd/walk-forward local builders map retained](../architecture/2026-05-07-walk-forward-local-builders-map-retained.md) — superseded by this decision; localBuilders eliminated
- [GlobalRegistry as immutable package-level var](../architecture/2026-05-07-globalregistry-immutable-package-level-var.md) — GlobalRegistry structure unchanged; now carries richer StrategyEntry metadata

## Revisit trigger

If `cmd/sweep2d` needs to support more strategies than sma-crossover and rsi-mean-reversion, consider adding a `SweepParams2D []SweepParam2DDef` field to `StrategyEntry`. If `cmd/signal-audit` audit-tuned params are codified in the registry (e.g., as a separate `AuditFactory` field), `allStrategyFactories()` can be auto-populated and the manual sync risk is eliminated.
