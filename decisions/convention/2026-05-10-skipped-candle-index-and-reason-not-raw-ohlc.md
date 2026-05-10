# SkippedCandle carries Index and Reason, not raw OHLC values

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | bad-candle, skip, SkippedCandle, manifest, TASK-0100 |

## Context

`SkippedCandle` records a single candle that was skipped by `parseKiteCandles` due to an OHLC validation failure. The struct needs to carry enough information for:
1. `cmd/fetch-history` to log a human-readable warning per skipped candle
2. The manifest (`fetch-progress.json`) to record the skip for auditability

The question: what fields should `SkippedCandle` carry?

## Options considered

### Option A: Index + Reason (string)
- **Pros**: Lightweight. `Reason` is already the formatted validation error from `model.Candle.Validate` — it contains the exact bad values (e.g. `"candle: open (835.6000) must be within [low=837.4000, high=843.8000]"`). `Index` locates it in the chunk context. No type dependency on `model.Candle` or raw float64 fields.
- **Cons**: To get the raw OHLC values separately, a caller must parse the Reason string (fragile).

### Option B: Index + Reason + raw OHLC float64 fields (Open, High, Low, Close)
- **Pros**: Structured access to the specific values that failed.
- **Cons**: Adds 4 float64 fields. The Reason string already contains the invalid values in human-readable form. No current caller needs to do arithmetic on the bad values — they only need to log them.

### Option C: Embed a partial `model.Candle` (timestamp + OHLC)
- **Pros**: Richer context.
- **Cons**: Creates a type dependency on `model.Candle` in the error type. `SkippedCandle` lives in `pkg/provider/zerodha` which already imports `pkg/model`; the dependency exists. But the value-copy semantics of `model.Candle` (a struct with 8 fields) add weight to what is currently a compact warning type.

## Decision

Option A: `SkippedCandle` carries only `Index int` and `Reason string`.

```go
type SkippedCandle struct {
    Index  int    // zero-based index in the raw API response
    Reason string // validation error message from model.Candle.Validate
}
```

`Index` is sufficient to locate the candle in the chunk context. `Reason` already contains the specific bad values — no parsing required by log consumers. Both fields are consumed by `cmd/fetch-history` as-is for warning logs and manifest serialization.

## Consequences

- `cmd/fetch-history` uses `sc.Index` in the warning format string and `sc.Reason` in the log — no post-processing.
- `manifestSkippedCandle` (the cmd-layer DTO) has the same two fields — mirroring is trivial and no transformation is needed.
- If a future caller needs structured access to the bad OHLC values, the `Reason` string must be parsed (fragile) or the struct must be extended.
