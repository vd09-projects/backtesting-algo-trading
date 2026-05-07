# Superset-lookup via filename parsing — no separate index file

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | cache, CachedProvider, filename-parsing, no-index, architecture, TASK-0089 |

## Context

TASK-0089 added range-aware superset lookup to `CachedProvider.FetchCandles`. Before hitting the network, the cache now scans the instrument's subdirectory for any file whose date range is a superset of the requested `[from, to]`. The implementation needed a way to identify candidate superset files without reading each one. Two structural options existed: parse existing filenames in-band, or maintain a separate index file mapping (instrument, timeframe) → list of cached ranges.

## Options considered

### Option A: Parse filenames in-band (chosen)

Scan `{cacheDir}/{instrument}/` with `os.ReadDir`, filter by timeframe prefix, parse `{tf}_{YYYY-MM-DD}_{YYYY-MM-DD}.json` filename structure to extract `cachedFrom` and `cachedTo`. One `ReadDir` syscall per `FetchCandles` invocation; no file I/O per candidate.

- **Pros**: No additional files to maintain; existing filename format already encodes the full range; zero write-path changes; index can never get out of sync with actual files; simple to implement and reason about.
- **Cons**: O(n) over cache files per `FetchCandles` call (n = number of cached files per instrument per timeframe, typically 1–10).

### Option B: Maintain a separate index file (rejected)

Write a JSON or YAML index (e.g., `{cacheDir}/{instrument}/index.json`) that maps timeframe → list of cached `[from, to]` ranges. Update the index on every `writeCache` call. Read only the index during superset lookup.

- **Pros**: O(1) lookup if the index is kept in memory; avoids per-call `ReadDir`.
- **Cons**: Adds write-path complexity; index must be updated atomically alongside every cache write; if the index is missing or stale (e.g., files manually deleted), superset lookup silently fails; introduces a new class of failure mode (index out of sync with actual files) that doesn't exist in Option A; performance advantage is irrelevant given n ≈ 1–10 and non-hot-loop call frequency.

## Decision

Option A — filename parsing, no index file. The existing `{tf}_{YYYY-MM-DD}_{YYYY-MM-DD}.json` format encodes everything needed for a superset check. `os.ReadDir` is called once per `FetchCandles` invocation; this call happens at run setup time (once per instrument×timeframe pair), not inside the candle-processing hot loop. The failure mode of Option B (index out of sync) is a meaningful correctness risk; Option A fails only if the filesystem is unreadable, which is already a hard failure in the existing code.

## Consequences

- `findSupersetFile` is O(n) over files in the instrument's cache directory. With n typically 1–10, this is not a performance concern.
- If a caller has accumulated many cache files per instrument (unusual), the scan degrades linearly. A future TASK-0080 manifest (write-side incremental tracking) and this read-side superset lookup are complementary but independent — TASK-0080 does not replace this pattern.
- Filename format is now load-bearing in two places: `cachePath` (writer) and `findSupersetFile` (reader). Any change to the filename format must update both.

## Related decisions

- [Instruments CSV cached at cacheDir/instruments.csv](2026-05-05-instruments-csv-cache-path.md) — companion caching architecture decision; same principle of explicit Config fields over auto-detection.

## Revisit trigger

If the number of cache files per instrument per timeframe grows above ~50 (e.g., from many different date ranges accumulated over many runs), consider switching to Option B with an atomic index write. Current usage pattern (1 wide file per instrument×timeframe) does not approach this threshold.
