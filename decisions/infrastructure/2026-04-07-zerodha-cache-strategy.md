# Zerodha provider — local file-based caching strategy

| Field    | Value          |
|----------|----------------|
| Date     | 2026-04-07     |
| Status   | accepted       |
| Category | infrastructure |
| Tags     | zerodha, cache, provider, file-cache, json, invalidation, kite-connect |

## Context

Iterative backtesting means `FetchCandles` is called repeatedly for the same instrument and
date range — often dozens of times per session as strategy parameters are tuned. Without a
cache, each run hammers the Kite API, burns through the rate limit, and adds 2–30 seconds of
network latency per backtest. A local cache eliminates all of that on repeat runs.

The cache must be:
- Transparent to callers — `DataProvider` interface is unchanged
- Persistent across process restarts
- Trivially inspectable and manually clearable
- Correct: must not serve stale data for open/recent sessions

## Options considered

### Option A — File-based JSON cache (chosen)

One JSON file per unique (`instrument`, `timeframe`, `from`, `to`) tuple. Files live in
`.cache/zerodha/` at the project root (already gitignored).

- **Pros:** Persists across restarts. Human-readable — you can `cat` a cache file to inspect
  raw candles. Zero dependencies — `encoding/json` is stdlib. Easy to nuke: `rm -rf .cache/`.
  Filesystem as the index — no separate metadata DB.
- **Cons:** Slow for very large files (thousands of candles in one JSON blob), but
  `encoding/json` handles 2,000 daily candles in <5ms — acceptable.

### Option B — SQLite cache

- **Pros:** Supports partial range queries (fetch only the subset you need from a large cached range).
- **Cons:** Requires a CGo dependency or a pure-Go SQLite driver. Adds significant complexity.
  Partial range queries are not a use case for v1 — the pattern is always "same range, different
  strategy parameters". Overkill.

### Option C — In-memory cache (per process lifetime only)

- **Pros:** Fastest possible. Zero I/O.
- **Cons:** Evaporates on process exit. Useless for the actual use case (repeat runs across
  sessions). Rejected immediately.

## Decision

**File-based JSON cache, wrapping the Zerodha provider as a decorator.**

### Cache location

Default: `.cache/zerodha/` at the repo root (already in `.gitignore`).
Override: `BACKTEST_CACHE_DIR` environment variable.

```
.cache/
  zerodha/
    NSE_NIFTY50/           ← instrument name, colon replaced with underscore
      daily_2024-01-01_2025-01-01.json
      minute_2024-06-01_2024-07-30.json
    NSE_RELIANCE/
      daily_2023-01-01_2024-01-01.json
```

### Cache key

`{sanitized_instrument}/{timeframe}_{from}_{to}.json`

- `from` and `to` formatted as `YYYY-MM-DD` (date only — time component is not meaningful for
  the cache key since candle data is day-bounded)
- Instrument sanitized: `:` → `_`, spaces → `_`, all lowercase
- Example: `nse_nifty50/daily_2024-01-01_2025-01-01.json`

The key is derived from the **exact parameters passed to `FetchCandles`**, not the chunk
boundaries. This means:
- `FetchCandles("NSE:NIFTY50", daily, 2024-01-01, 2025-01-01)` → one cache file
- `FetchCandles("NSE:NIFTY50", daily, 2023-01-01, 2025-01-01)` → different cache file (miss)

Partial range reuse is not implemented in v1. The primary use case — running the same strategy
multiple times on the same instrument and date range — always produces identical keys.

### Cache file format

```json
{
  "cached_at": "2026-04-07T10:30:00Z",
  "candles": [
    {
      "Instrument": "NSE:NIFTY50",
      "Timeframe": "daily",
      "Timestamp": "2024-01-01T09:15:00Z",
      "Open": 21500.5,
      "High": 21650.0,
      "Low": 21480.0,
      "Close": 21620.0,
      "Volume": 123456
    }
  ]
}
```

`cached_at` is used for TTL-based invalidation of recent data (see below). Timestamps in the
`candles` array are stored in UTC (the provider converts from IST +0530 on ingest).

### Invalidation rules

| Condition | Rule | Rationale |
|---|---|---|
| `to` < today's date (UTC) | **Never invalidate** | Historical sessions are closed; candle data never changes |
| `to` >= today's date (UTC) | **TTL = 1 hour** | Intraday data updates during market hours |

TTL check: on cache read, if `now - cached_at > 1h` and `to >= today`, delete the file and
re-fetch. This means an intraday cache file is automatically refreshed at most once per hour.

### Cache layer architecture

`CachedProvider` is a decorator that wraps any `DataProvider`:

```go
type CachedProvider struct {
    inner    provider.DataProvider
    cacheDir string
}

func NewCachedProvider(inner provider.DataProvider, cacheDir string) *CachedProvider

func (c *CachedProvider) FetchCandles(ctx context.Context, instrument string,
    tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
    // 1. Compute cache path
    // 2. Check if file exists and is valid (TTL check)
    // 3. On hit: unmarshal and return
    // 4. On miss: call c.inner.FetchCandles, marshal and write, return
}

func (c *CachedProvider) SupportedTimeframes() []model.Timeframe {
    return c.inner.SupportedTimeframes()
}
```

`CachedProvider` satisfies `DataProvider` at compile time (verified by a compile-time check
in the test file). The wiring in `cmd/backtest/main.go` is:

```go
zerodha := zerodha.NewProvider(apiKey, accessToken)
provider := cache.NewCachedProvider(zerodha, cacheDir)
engine.Run(ctx, provider, strategy)
```

### Concurrency

This is a single-process CLI tool. No concurrent cache access. No file locking needed in v1.

## Consequences

- A cache miss on the first run of a new date range incurs full API latency + chunking overhead.
  Every subsequent run for the same range is a fast local file read (<10ms for daily data,
  <100ms for large intraday files).
- Cache files are plain JSON — a developer can inspect, edit, or patch them manually for
  testing purposes.
- The cache is keyed on exact (`instrument`, `timeframe`, `from`, `to`) — changing any
  parameter produces a miss. This is the expected behaviour for backtesting workflow.
- Cache must be cleared manually when testing the fetch pipeline (`rm -rf .cache/zerodha/`).
  No programmatic clear API in v1.
- `CachedProvider` lives in `pkg/provider/zerodha/cache/` — it is part of the Zerodha
  provider package group, not a generic caching utility. If a second provider is added later,
  the cache can be promoted to `pkg/provider/cache/` as a generic decorator.

## Related decisions

- [Zerodha pagination strategy](./2026-04-07-zerodha-pagination-strategy.md) — the cache
  wraps at the `FetchCandles` level (full requested range), above the chunk loop. A cache hit
  skips chunking and all API calls entirely.
- [Zerodha auth strategy](./2026-04-07-zerodha-auth-strategy.md) — auth is handled by the
  inner Zerodha provider; the cache layer is auth-agnostic.

## Revisit trigger

- If partial range reuse becomes valuable (e.g. caching 5 years of data and slicing arbitrary
  subsets), consider chunking the cache by calendar month and assembling slices from multiple
  files — or switch to SQLite.
- If multiple strategies are run in parallel against the same data (concurrent processes),
  add file locking (`flock` or a `.lock` sidecar file).
