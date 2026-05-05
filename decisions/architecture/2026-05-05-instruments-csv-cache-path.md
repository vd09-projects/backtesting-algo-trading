# Instruments CSV cached at {cacheDir}/instruments.csv alongside candle cache

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-05       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | zerodha, provider, instruments-csv, caching, InstrumentsCacheDir, TASK-0081 |

## Context

`zerodha.NewProvider` previously downloaded the instruments CSV on every construction via `kite.Instruments()`, requiring a live Zerodha access token unconditionally. This blocked the TASK-0069 bootstrap session mid-run after token expiry — all candle data was cached but the provider failed to construct. The fix required deciding where to store the cached CSV and how to structure the cache path API.

## Options considered

### Option A: Fixed path — always `.cache/zerodha/instruments.csv`
- **Pros**: Simple; no Config change needed.
- **Cons**: Hardcodes path in library; different cmd/ binaries using different cache dirs would write to the same file, causing staleness races.

### Option B: Separate config field — `InstrumentsCacheDir` on `zerodha.Config`
- **Pros**: Caller controls location; no hardcoded paths in library; collocates instruments.csv with candle cache under the same root that `cmdutil.BuildProvider` already knows about.
- **Cons**: API surface increase (one new Config field).

### Option C: Separate directory — `InstrumentsCacheDir` distinct from candle cache dir
- **Pros**: Decouples instruments cache lifecycle from candle cache lifecycle.
- **Cons**: Unnecessary complexity — both caches belong to the same provider; splitting them adds two distinct Config fields to manage.

## Decision

Option B. `InstrumentsCacheDir` on `zerodha.Config` controls where `instruments.csv` is written and read. The path is `{InstrumentsCacheDir}/instruments.csv` — intentionally simple (no subdirectory). `cmdutil.BuildProvider` passes `cacheDir` as `InstrumentsCacheDir`, so all cmd/ binaries collocate the instruments CSV with the candle cache automatically.

Cache freshness: file age < 24h = fresh (skip network). Stale or absent = fetch from Kite, write to disk before returning.

## Consequences

- Empty `InstrumentsCacheDir` preserves original uncached behavior (backward compatible with all existing tests).
- Non-empty `InstrumentsCacheDir` enables token-free construction when instruments CSV is fresh — fully-cached eval runs no longer require a live token.
- `loadOrCacheInstrumentsCSV` in `instruments_cache.go` owns the read/write/TTL logic; `NewProvider` delegates to it.

## Related decisions

- [Zerodha instrument token lookup at init](2026-04-07-zerodha-instrument-token-lookup.md) — prior decision: download /instruments CSV at init, no disk cache. This decision supersedes the "no disk cache" element.
- [InstrumentsCacheDir as explicit Config field](2026-05-05-instruments-cache-dir-explicit-config-field.md) — companion decision on the Config API design.

## Revisit trigger

If the instruments CSV format changes (Kite API version bump) and the cached file must be invalidated by schema version rather than age alone — 24h TTL may be insufficient in that case.
