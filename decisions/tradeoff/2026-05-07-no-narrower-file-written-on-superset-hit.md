# No narrower cache file written on superset hit

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | cache, CachedProvider, write-behavior, no-duplicate, tradeoff, TASK-0089 |

## Context

TASK-0089 introduced superset-aware lookup to `CachedProvider.FetchCandles`. When a request for `[from, to)` is served from an existing wider cache file `[cachedFrom, cachedTo)`, a choice must be made about whether to also write a new cache file for the exact `[from, to)` range.

## Options considered

### Option A: Do not write a new file on superset hit (chosen)

Return the filtered slice immediately without writing to disk.

- **Pros**: No duplicate storage; avoids a write syscall on every superset hit; the superset file will serve all future requests within its range via `findSupersetFile`; simpler code path (no conditional write logic).
- **Cons**: Future requests for the same narrow range must scan the directory to find the superset file each time (O(n) `ReadDir`) rather than hitting an exact-match file (O(1) stat). In practice, n is tiny and `FetchCandles` is called once per run, so this is not a performance concern.

### Option B: Write a new exact-match file on superset hit (rejected)

After serving from the superset file, also persist the filtered slice as `{tf}_{from}_{to}.json`.

- **Pros**: Subsequent exact-match requests for `[from, to)` would hit the O(1) exact-match path instead of the O(n) superset scan.
- **Cons**: Duplicate disk storage (the narrow range data exists twice); adds a write call on every superset hit, even when unnecessary; if the superset file is later updated or deleted, the narrow file becomes stale without any mechanism to detect it; complexity increase for a negligible performance gain.

## Decision

Option A — no write on superset hit. The superset file is a durable record of the full fetched range and will serve all future requests within that range. Writing a narrower file would create redundant storage and a new class of stale-data risk (narrow file diverges from superset). The O(n) `ReadDir` cost on subsequent superset hits is irrelevant given call frequency (once per backtest run) and n (1–10 files).

## Consequences

- Cache directory contains one canonical file per full fetch, not one file per unique requested range. This keeps the cache directory clean and avoids unbounded storage growth from incremental range variations.
- `findSupersetFile` is always called for any request that misses the exact-match path. For a typical run with a wide cache file and multiple narrow requests, all narrow requests go through the O(n) scan. This is acceptable given the call pattern.
- TASK-0080 (incremental manifest for `LastCachedTime`) is a write-side concern and coexists cleanly with this decision — the manifest tracks the latest candle time, not the set of cached ranges.

## Revisit trigger

If a future use case requires many narrow ranges to be fetched and re-used frequently in a tight loop (e.g., a parallel walk-forward that fetches each fold's range as an individual `FetchCandles` call), the O(n) superset scan per fold may become meaningful. In that scenario, re-evaluate Option B or add a lightweight in-process range index (not on-disk).
