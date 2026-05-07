# lazyProvider: defer Zerodha auth and client init to first cache miss

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | lazy-init, auth, CachedProvider, cmdutil, BuildProvider, token, zerodha |

## Context

`BuildProvider` in `internal/cmdutil` constructs the Zerodha-backed `CachedProvider` used by all `cmd/` entrypoints. Before this change, `BuildProvider` loaded the saved access token and initialised the Zerodha HTTP client unconditionally at startup — even when all requested candle data was already on disk. The result: every backtest invocation printed "Loaded saved token from ..." and hit the token file regardless of cache state. For repeated runs on historical data this is pure overhead.

The `CachedProvider` wraps an inner `provider.DataProvider` and calls `inner.FetchCandles` only on a cache miss. If the inner provider's initialisation is deferred, auth becomes a no-op on full cache hits.

## Options considered

### Option A: Eager init (prior behaviour)
- **Pros**: Simple; token load failure surfaces at startup before any work begins.
- **Cons**: Auth happens unconditionally even when the cache covers the entire request; "Loaded saved token" prints on every cached run.

### Option B: `lazyProvider` wrapper with `sync.Once`
- **Pros**: Token load and Zerodha client construction are deferred until the first `FetchCandles` that reaches the inner provider (i.e. a cache miss). Full cache hits bypass auth entirely. The lazy init is thread-safe via `sync.Once`.
- **Cons**: Token load failure surfaces on first cache miss, not at startup. Harder to distinguish "auth failed" from "fetch failed" in logs if the error message isn't explicit.

### Option C: Cache-first check in `BuildProvider` before constructing Zerodha client
- **Pros**: Could skip auth without changing the provider interface.
- **Cons**: `BuildProvider` would need to know the request parameters (instrument, from, to, timeframe) at construction time — those are only known at call site. Tight coupling, wrong abstraction layer.

## Decision

Introduced a private `lazyProvider` type in `internal/cmdutil/cmdutil.go` that implements `provider.DataProvider`. It holds a `sync.Once` and an `initFn func() (provider.DataProvider, error)`. `FetchCandles` calls `initFn` on the first invocation, then delegates to the initialised inner provider. `SupportedTimeframes` returns a hardcoded list of Zerodha-supported timeframes to avoid triggering init on what is effectively a static query.

`BuildProvider` now constructs a `lazyProvider` (with `initFn` closing over apiKey, apiSecret, tokenPath, cacheDir) and wraps it in `CachedProvider`. The Zerodha token is not read until a real network fetch is needed.

## Consequences

- Cached runs no longer print "Loaded saved token from ..." and do not touch the token file.
- Auth errors surface at the point of first network fetch, not at process start. Error messages from `initFn` are wrapped with context so the cause is still clear.
- `SupportedTimeframes` on `lazyProvider` is hardcoded to Zerodha's five timeframes. This is acceptable because `BuildProvider` is explicitly Zerodha-specific; if a second provider is added, `BuildProvider` will be replaced or extended, not reused.
- The `sync.Once` guard means concurrent goroutines racing on first fetch are safe; only one will run `initFn`.

## Related decisions

- [Superset-lookup via filename parsing — no separate index file](2026-05-07-superset-lookup-via-filename-parsing-no-index.md) — the lazy pattern is only useful because `CachedProvider` can serve full ranges from disk; both changes together eliminate auth on cached runs.

## Revisit trigger

If a second `DataProvider` backend is added (non-Zerodha), `lazyProvider.SupportedTimeframes()` will need to either delegate to the inner provider after init or be removed from the interface contract.
