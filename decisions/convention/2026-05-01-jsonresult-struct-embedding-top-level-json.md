# jsonResult struct embedding for top-level JSON merge

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | convention       |
| Tags     | JSON, embedding, serialization, struct-embedding, omitempty, internal/output, TASK-0064 |

## Context

TASK-0064 required embedding `RunConfig` metadata fields at the top level of the backtest result JSON — not nested under a sub-key like `"run_config": {...}`. `analytics.Report` is the existing type written to JSON. The question was how to merge `RunConfig` and `analytics.Report` at the JSON top level without changing `analytics.Report`'s signature or breaking existing callers.

## Options considered

### Option A: Add RunConfig fields directly to analytics.Report
- **Pros**: Simple; no extra type
- **Cons**: `analytics.Report` is a domain type in `internal/analytics`; adding serialization-only metadata fields violates its responsibility. Also breaks the principle that analytics knows nothing about run configuration.

### Option B: Nested JSON sub-key ({"run_config": {...}, "trade_count": ...})
- **Pros**: No merging needed; caller just passes `RunConfig` as a field
- **Cons**: Violates the acceptance criterion (top-level embedding required); nested keys are less ergonomic for downstream tooling

### Option C: unexported jsonResult struct with embedded fields (chosen)
- **Pros**: Go's field promotion merges both types at the JSON top level without modifying either; `omitempty` on `RunConfig` fields means zero-valued `RunConfig` emits no extra keys — backward compatible; the merge type is internal to `writeJSON` and has no surface area
- **Cons**: Slight indirection; the merged type is invisible in the exported API

## Decision

An unexported `jsonResult` struct embeds both `RunConfig` and `analytics.Report` as anonymous fields. Go's `encoding/json` promotes all exported fields from both embedded types to the top level of the JSON object. `RunConfig` fields carry `omitempty` JSON tags so that callers passing a zero-valued `RunConfig` get identical JSON output to before — no backward compatibility break.

```go
type jsonResult struct {
    RunConfig
    analytics.Report
}
```

## Consequences

This is a zero-surface-area merge: `jsonResult` is unexported and confined to `writeJSON`. Existing callers of `output.Write` are unaffected unless they pass a non-zero `RunConfig`. The `omitempty` guarantee means the JSON schema is additive, not breaking.

If `analytics.Report` ever adds a field with the same JSON key name as a `RunConfig` field, there will be a silent collision at the JSON level. The build won't catch it — it requires a test asserting both fields appear. The existing `TestWrite_RunConfig_MetricsFieldsStillPresent` test guards against this.

## Revisit trigger

If `analytics.Report` adds fields with JSON key names that collide with `RunConfig` field names.
