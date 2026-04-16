# Equity curve wired into output.Config fields, not a second Write parameter

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-16       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | config-shape, equity-curve, output, internal/output, TASK-0029 |

## Context

TASK-0029 required `output.Write` to accept a `[]model.EquityPoint` equity curve alongside the existing `analytics.Report`, so it could serialize it to a CSV file when `--output-curve` is set. The question was how to thread the new data into `Write` — which already has a `Config` struct as its second parameter.

## Options considered

### Option A: New parameter `Write(report, curve, cfg)`
- **Pros**: Explicit — the curve is a first-class argument, not buried in a config field.
- **Cons**: Breaks all existing callers (`cmd/backtest`, `cmd/sweep`, all output tests). Higher churn for zero behavioral benefit. `Config` already acts as the extension point for optional outputs — adding a required positional parameter contradicts that design.

### Option B: Separate `WriteCurve(path string, curve []model.EquityPoint)` function
- **Pros**: Single-responsibility; the curve write is its own call.
- **Cons**: Splits one logical "write all backtest outputs" operation into two call sites that callers must manually coordinate. Any future caller that forgets to call `WriteCurve` silently skips curve export with no compile-time warning.

### Option C: `Config.CurvePath` + `Config.Curve` fields (chosen)
- **Pros**: `Write` call site stays `output.Write(report, cfg)`. New behavior is opt-in by setting two Config fields — callers that don't set them get the old behavior unchanged. Consistent with how `Config.FilePath`, `Config.Benchmark`, and `Config.Stdout` already work. All existing callers and tests compile and pass without modification.
- **Cons**: `Config` gains a slice field (`[]model.EquityPoint`), making it slightly heavier when copied by value. Slice header copy is 24 bytes — not a real cost at this call frequency (once per backtest run).

## Decision

Config fields (Option C). `Config` is already the established extension point for optional output destinations in this package. Adding `CurvePath string` and `Curve []model.EquityPoint` follows the existing pattern exactly, keeps `Write`'s signature stable, and makes new behavior opt-in at zero cost to existing callers.

## Consequences

- All existing output tests pass unchanged.
- `writeCurveCSV` is unexported and reached exclusively through `Write`, keeping the public API surface flat.
- Future output destinations (e.g., regime split CSV, per-trade CSV) follow the same pattern: new fields on `Config`, guarded by an empty-path check in `Write`.

## Related decisions

- [io.Writer field in Config for stdout testability](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — established the Config-as-extension-point pattern this decision extends.
- [EquityPoint defined in pkg/model, not internal/engine](../convention/2026-04-10-equitypoint-in-pkg-model.md) — why `[]model.EquityPoint` is importable by `internal/output` without a circular dependency.

## Revisit trigger

If `output.Write` accumulates more than 3-4 optional output destinations via Config fields, reconsider whether a builder pattern or explicit sub-writers would be cleaner.
