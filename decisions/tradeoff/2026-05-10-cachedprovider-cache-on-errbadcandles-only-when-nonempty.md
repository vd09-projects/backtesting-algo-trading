# CachedProvider caches candles on ErrBadCandles only when len(candles) > 0

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | bad-candle, ErrBadCandles, cache, empty-candle-guard, TASK-0100 |

## Context

`CachedProvider.FetchCandles` wraps any `DataProvider`. When the inner provider returns a non-nil error, the existing behavior was: propagate error, write nothing to cache. With `*ErrBadCandles` as a non-fatal warning that carries valid candles, `CachedProvider` must decide:

1. Whether to cache the partial candles (preventing a re-fetch of the same bad candle next time)
2. What to do when the inner provider returns an empty candle slice alongside `*ErrBadCandles` (all candles bad)

## Options considered

### Option A: Always write to cache when inner returns `*ErrBadCandles`
- **Pros**: Prevents re-fetching the same date range.
- **Cons**: If the candle slice is empty (all bad), writing `{"candles":null}` creates a phantom cache entry. A subsequent `readCache` returns `(nil, true)` — valid hit, nil candles, nil error — silently bypassing the network forever.

### Option B: Guard `writeCache` with `len(candles) > 0`
- **Pros**: Prevents phantom empty-candle cache entries. A second call on an all-bad response will re-fetch (a subsequent fetch might succeed if the artifact is transient, or return the same result).
- **Cons**: When ≥1 valid candle exists, the cache is written as expected. When all candles are bad, the network will be retried. Slightly more network traffic in the degenerate case.

### Option C: Never cache when inner returns `*ErrBadCandles`
- **Pros**: Simpler logic.
- **Cons**: Forces a re-fetch of the same date range on every subsequent call, even when most candles are valid. The typical scenario (1 bad out of ~98,900) would re-fetch 98,900 candles unnecessarily.

## Decision

Option B: guard `writeCache` with `if len(candles) > 0`. The implementation:

```go
if len(candles) > 0 {
    _ = c.writeCache(path, candles)
}
return candles, err  // propagate ErrBadCandles warning (or nil) to caller
```

The empty-candle guard protects against phantom cache entries. The non-empty path caches the valid candles and propagates the warning unchanged to the caller.

## Consequences

- An all-bad instrument fetch will retry the network on every `cmd/fetch-history` re-run (no manifest `completed` entry is written, and no cache file is created). This is correct — the instrument should be retried.
- A partial-bad fetch (≥1 valid candle) behaves like a normal fetch for caching purposes. The `ErrBadCandles` warning is still propagated so `cmd/fetch-history` can log warnings and update the manifest.
- The `TestCachedProvider_DoesNotCacheEmptyOnErrBadCandles` test verifies the empty-guard path and confirms the second call re-hits the network.

## Related decisions

- [ErrBadCandles as non-fatal typed warning](../architecture/2026-05-10-errbadcandles-non-fatal-warning-alongside-valid-candles.md) — defines the error type this decision handles
