# ErrBadCandles as non-fatal typed warning returned alongside valid candles

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | bad-candle, skip, OHLC-validation, Zerodha-artifact, ErrBadCandles, TASK-0100 |

## Context

`cmd/fetch-history` hard-failed the entire instrument fetch on any OHLC validation error from `model.NewCandle`. 10 instruments (HDFCBANK, ICICIBANK + 8 midcap) hit candle[450] OHLC error — a Zerodha tick-vs-aggregation artifact where the raw API returns an open price marginally outside [low, high] on a specific date. Marcus's ruling: bad candle → warn, skip, continue. The validator in `pkg/model/candle.go` must NOT be relaxed.

The `DataProvider` interface signature is `FetchCandles(...) ([]model.Candle, error)`. To surface skip information to callers, the error return must carry both the valid candles AND the skip details — the interface cannot be changed without touching every implementation and every caller.

## Options considered

### Option A: Return non-nil typed warning alongside candles (`ErrBadCandles`)
- **Pros**: No interface change. Fail-safe default — callers not checking `errors.As` see a hard error. Typed for structured inspection. CachedProvider can selectively handle it.
- **Cons**: Returning a non-nil error alongside a non-nil candle slice is unconventional in Go. Callers must know to check `errors.As` before discarding candles.

### Option B: Return candles silently (skip internally, no error surfaced)
- **Pros**: Simple, no interface changes, no caller changes.
- **Cons**: `cmd/fetch-history` cannot log per-candle warnings or update the manifest with skip details. Silent data loss — callers cannot audit what was dropped.

### Option C: Change the `DataProvider` interface to return a result struct
- **Pros**: Clean, explicit, type-safe.
- **Cons**: Breaking change requiring all implementations and callers to update. Disproportionate for a single-provider project.

## Decision

Option A: `FetchCandles` returns `([]model.Candle, *ErrBadCandles)` when one or more candles are skipped. The `*ErrBadCandles` type carries the instrument name and a `[]SkippedCandle` slice (each with `Index int` and `Reason string`).

The fail-safe property is intentional: callers that write `if err != nil { return err }` will treat this as a hard failure and surface it. Only callers explicitly handling bad-candle skipping (currently only `cmd/fetch-history`) use `errors.As` to detect `*ErrBadCandles` and treat the fetch as a partial success. `CachedProvider` handles the type check to cache the valid candles and propagate the warning.

## Consequences

- A non-nil error alongside a non-nil candle slice is unusual Go. The `ErrBadCandles` godoc must be clear that callers should use `errors.As` before discarding candles.
- `CachedProvider` now needs explicit `errors.As` handling in its network-fetch branch. It must cache the candles (valid, worth caching) and propagate the warning.
- Any future DataProvider caller that fetches candles in a bulk loop must explicitly handle `*ErrBadCandles` or lose the valid data.

## Related decisions

- [manifestSkippedCandle as local DTO in cmd/fetch-history](./2026-05-10-manifest-skipped-candle-local-dto-not-provider-type.md) — DTO in cmd-layer for JSON serialization; decoupled from this type
- [cache package imports zerodha for ErrBadCandles](./2026-05-10-cache-imports-zerodha-for-errbadcandles-not-circular.md) — consequence of this decision: cache must know ErrBadCandles

## Revisit trigger

If a second `DataProvider` implementation is added that also needs to surface skipped data, consider whether `ErrBadCandles` should move to `pkg/provider/` (interface package) so the cache package does not depend on the Zerodha-specific package.
