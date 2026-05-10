# ErrBadCandles all-candles-bad path returns error and skips completed mark

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | bad-candle, all-bad, error-path, manifest, TASK-0100 |

## Context

`cmd/fetch-history/fetchOne` handles `*ErrBadCandles` from `FetchCandles`. When `*ErrBadCandles` is detected alongside a non-empty candle slice (≥1 valid candle), the fetch is treated as a partial success: warnings logged, manifest updated with `skipped_candles`, instrument added to `Completed`.

The degenerate case: all candles in the response fail OHLC validation, so `len(candles) == 0`. What should `fetchOne` do?

## Options considered

### Option A: Treat all-bad as success (append to Completed, return nil)
- **Pros**: Consistent with the partial-bad success path; no special casing.
- **Cons**: The instrument is marked as `Completed` with zero useful candles cached. A subsequent backtest run that picks up this instrument will silently operate on an empty candle series. The original bug (hard-fail) is replaced with a silent-data-loss bug. The initial implementation took this path — caught by the multi-perspective review.

### Option B: Treat all-bad as failure (do NOT append to Completed, return error)
- **Pros**: The instrument will be retried on the next `cmd/fetch-history` run (no manifest entry to skip it). If the Zerodha artifact is transient (unlikely but possible), a retry may succeed. If it is permanent, the operator is clearly informed via the error count.
- **Cons**: The operator sees the instrument in the failure count, which might seem alarming. The log message must be clear that the failure is due to all candles being bad, not a network error.

## Decision

Option B: when `len(candles) == 0` after detecting `*ErrBadCandles`, `fetchOne` logs a descriptive error to stderr and returns the error:

```
<inst> × <tf>: all N candles were invalid, nothing cached — instrument will retry on next run
```

The instrument is NOT appended to `manifest.Completed`. `saveManifest` is still called to preserve progress from any prior successful fetches in the same run.

This was the second blocking issue found by the multi-perspective review's Error Handling Inspector and fixed in iterate round 1.

## Consequences

- In the real-world Zerodha artifact scenario (1 bad candle out of ~98,900 valid), `len(candles)` is never 0 — the partial-bad success path applies. The all-bad guard is a defensive correctness check.
- The all-bad guard prevents a new class of silent data loss where an instrument is marked complete but unusable.
- Operators running `cmd/fetch-history` against the 10 previously failing instruments (HDFCBANK, ICICIBANK + 8 midcap) will see warnings for the skipped candle[450] but the instrument will succeed (partial-bad path), not fail (all-bad path).

## Related decisions

- [ErrBadCandles as non-fatal typed warning](../architecture/2026-05-10-errbadcandles-non-fatal-warning-alongside-valid-candles.md) — defines the parent error type
- [CachedProvider cache guard for empty candles](../tradeoff/2026-05-10-cachedprovider-cache-on-errbadcandles-only-when-nonempty.md) — complementary guard at the cache layer
