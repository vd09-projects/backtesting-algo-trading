# manifestSkippedCandle as local DTO in cmd/fetch-history, not embedding zerodha.SkippedCandle

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | bad-candle, manifest, DTO, serialization, cmd-layer, TASK-0100 |

## Context

`cmd/fetch-history` writes skipped-candle information to `fetch-progress.json` under a `skipped_candles` key on each `progressEntry`. This information originates from `zerodha.SkippedCandle` (via `*ErrBadCandles`). The manifest needs a JSON-serializable struct for the skipped candle records.

Two options for the struct type:
1. Directly embed / reference `zerodha.SkippedCandle` in `progressEntry`
2. Define a local DTO `manifestSkippedCandle` with identical fields

## Options considered

### Option A: Reference `zerodha.SkippedCandle` directly in progressEntry
- **Pros**: No duplication — one struct definition.
- **Cons**: The manifest format (JSON schema) becomes coupled to the provider package's internal type. If `zerodha.SkippedCandle` gains new fields, the manifest serializes them automatically (may be unwanted). If `zerodha.SkippedCandle` is renamed or restructured, the manifest schema changes without a deliberate decision. The cmd-layer manifest is a stable external interface; its schema should be controlled explicitly.

### Option B: Local `manifestSkippedCandle` DTO in cmd/fetch-history (chosen)
- **Pros**: Manifest schema is decoupled from provider type. Adding fields to `zerodha.SkippedCandle` does not affect the manifest. The cmd-layer controls exactly which fields appear in JSON output.
- **Cons**: One extra struct definition with identical fields (`Index int`, `Reason string`). Manual synchronization if `zerodha.SkippedCandle` changes shape.

## Decision

Option B: define `manifestSkippedCandle` locally in `cmd/fetch-history/main.go`:

```go
type manifestSkippedCandle struct {
    Index  int    `json:"index"`
    Reason string `json:"reason"`
}
```

This follows the existing pattern in this codebase: `cmd/monitor` uses a `thresholdsFile` DTO rather than embedding `analytics.KillSwitchThresholds` (decision `2026-05-07-thresholdsfile-dto-in-cmd-monitor.md`). The manifest is a cmd-layer serialization concern; provider types should not carry JSON tags.

Conversion from `zerodha.SkippedCandle` to `manifestSkippedCandle` is a trivial field copy in `fetchOne`.

## Consequences

- If `zerodha.SkippedCandle` gains a `Timestamp` field in the future, the manifest does not automatically include it — a deliberate update to `manifestSkippedCandle` and its population code is required.
- `fetch-progress.json` schema: `skipped_candles` is an array of `{"index": N, "reason": "..."}` objects. This is stable regardless of upstream type changes.

## Related decisions

- [thresholdsFile DTO in cmd/monitor](../architecture/2026-05-07-thresholdsfile-dto-in-cmd-monitor.md) — same pattern: cmd-layer DTO over provider/analytics type
