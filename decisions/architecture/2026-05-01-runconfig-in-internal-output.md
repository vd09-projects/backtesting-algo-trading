# RunConfig placed in internal/output as serialization DTO

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | metadata, JSON, run-config, output, serialization-dto, TASK-0064 |

## Context

Backtest result JSON files contained only performance metrics with no metadata about the run itself — no instrument, timeframe, date range, strategy name, parameters, or commission model. TASK-0064 added a `RunConfig` metadata block to the JSON output. The question was where `RunConfig` should live: `pkg/model` (alongside `Trade`, `Candle`, etc.) or `internal/output` (alongside the formatter that writes it).

## Options considered

### Option A: RunConfig in pkg/model
- **Pros**: Alongside other domain types; importable by multiple packages if needed
- **Cons**: `RunConfig` is purely a serialization concern — it describes how a run looks in output form, not a domain primitive. Putting it in `pkg/model` would widen the model package with a type that no non-output code needs, and would couple the JSON shape to domain type changes.

### Option B: RunConfig in internal/output
- **Pros**: Collocated with the formatter that consumes it; describes serialized form, not domain; doesn't pollute `pkg/model`; zero external importers needed
- **Cons**: If a future consumer outside `internal/output` needs `RunConfig`, it can't import it (internal package restriction)

## Decision

`RunConfig` lives in `internal/output` as a serialization DTO. It describes how a backtest run is represented in serialized output form, not a domain primitive that other packages need. All fields are plain strings (not `model.Timeframe` or `model.CommissionModel`) to decouple the JSON shape from `pkg/model` type changes. The cmd layer converts typed values to strings before constructing `RunConfig`.

## Consequences

If a future consumer outside `internal/output` needs `RunConfig` (e.g., a sweep runner that embeds run metadata in sweep CSV output), it cannot import the type directly. At that point, either promote `RunConfig` to `pkg/model` with a targeted migration, or define a parallel type in the consumer. This is a bridge-when-needed tradeoff — premature promotion to `pkg/model` is worse than a future migration.

## Related decisions

- [ParseCommissionModel extracted to internal/cmdutil](../convention/2026-04-29-parse-commission-model-extracted-to-cmdutil.md) — same principle: cmd-layer plumbing stays out of pkg/model
- [BuildProvider extracted to cmdutil](../architecture/2026-04-22-buildprovider-extracted-to-cmdutil.md) — established pattern for cmd-layer helpers

## Revisit trigger

If `RunConfig` is needed by more than one non-output package, promote to `pkg/model` at that point.
