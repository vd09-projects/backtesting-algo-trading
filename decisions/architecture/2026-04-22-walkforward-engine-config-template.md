# Run() accepts EngineConfigTemplate, not engine.Config directly

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | walkforward, engine-config, dependency-injection, config-template, TASK-0022 |

## Context

`internal/walkforward.Run()` needs engine configuration (initial cash, position sizing, order config) that is constant across all folds, plus fold-specific fields (From, To, Instrument) that the harness stamps per fold. The question was how to pass these two kinds of configuration without coupling the public API to `engine.Config` directly.

## Options considered

### Option A: Accept engine.Config and overwrite From/To/Instrument per fold
- **Pros**: One fewer type in the API
- **Cons**: Makes it implicit which fields the harness controls; a caller could set engine.Config.From thinking it controls the outer window, not realizing the harness overwrites it per fold

### Option B: EngineConfigTemplate separate type (selected)
- **Pros**: Explicit boundary — the harness owns From/To/Instrument; the caller owns cost model and sizing. Prevents accidental misuse. WalkForwardConfig stays clean (walk-forward params only)
- **Cons**: One extra struct type in the API

## Decision

`Run()` accepts a separate `EngineConfigTemplate` holding only the fields that are constant across folds (InitialCash, OrderConfig, PositionSizeFraction, SizingModel, VolatilityTarget). Per fold, the harness derives a fresh `engine.Config` by copying the template and stamping Instrument, From, and To. This makes the API surface explicit about which caller-controlled fields the harness stamps, and avoids WalkForwardConfig becoming a superset of engine.Config (a maintenance coupling that would require WalkForwardConfig to grow whenever engine.Config grows).

## Consequences

Callers constructing a Run() call write two structs instead of one. This is a small inconvenience justified by the clearer API contract.
