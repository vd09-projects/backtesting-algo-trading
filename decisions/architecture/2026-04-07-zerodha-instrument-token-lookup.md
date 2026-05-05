# Zerodha instrument token lookup — CSV download at provider init

| Field    | Value        |
|----------|--------------|
| Date     | 2026-04-07   |
| Status   | superseded (no-disk-cache element superseded 2026-05-05 by TASK-0081) |
| Category | architecture |
| Tags     | zerodha, provider, instrument-token, instruments-csv, lookup, init, kite-connect |

## Context

The `DataProvider` interface accepts instrument identifiers as human-readable strings (e.g.
`"NSE:NIFTY50"`). The Kite Connect historical data endpoint requires a **numeric
`instrument_token`** (e.g. `256265`), obtained from a separate instruments master CSV.

The instruments CSV is available at `GET /instruments` (no authentication required) and
contains one row per tradable instrument with fields including `instrument_token`,
`tradingsymbol`, `exchange`, `name`, and others.

There is no endpoint to look up a single instrument by symbol — the only option is to
download the full CSV and filter it in memory.

## Options considered

### Option A — Download CSV on every FetchCandles call
- **Pros:** Always fresh.
- **Cons:** ~5MB download on every call. Completely wasteful — the instruments list changes
  at most once per day (new listings/delistings). Would dominate latency and rate limit budget.

### Option B — Download once at provider init, cache in memory for the process lifetime (chosen)
- **Pros:** One download per process start. Fast lookups thereafter (map[string]int).
  Instruments change infrequently — daily re-download is sufficient.
- **Cons:** If an instrument is listed mid-session, the in-memory map won't reflect it
  until the process restarts. Acceptable for a backtesting tool — we're not tracking live
  listings.

### Option C — Download once per day, persist to file
- **Pros:** Survives process restarts without re-downloading.
- **Cons:** Adds file I/O, cache invalidation, and complexity. The instruments CSV is ~5MB
  and takes <1s to download — not worth caching to disk in v1.

## Decision

**Option B — download at init, hold in memory.**

`NewZerodhaProvider` downloads the instruments CSV from `/instruments`, parses it into a
`map[string]int` keyed by `"{exchange}:{tradingsymbol}"` → `instrument_token`, and stores it
on the provider struct. `FetchCandles` looks up the token from this map before making the
historical API call.

```go
type ZerodhaProvider struct {
    apiKey      string
    accessToken string
    tokens      map[string]int // "NSE:NIFTY50" → 256265
    httpClient  *http.Client
}

func NewZerodhaProvider(apiKey, accessToken string) (*ZerodhaProvider, error) {
    p := &ZerodhaProvider{...}
    if err := p.loadInstruments(); err != nil {
        return nil, fmt.Errorf("zerodha: load instruments: %w", err)
    }
    return p, nil
}
```

If an instrument symbol is not found in the map, `FetchCandles` returns a typed error:
`ErrInstrumentNotFound` — so the caller gets a clear message rather than a 400 from the API.

## Consequences

- Provider construction (`NewZerodhaProvider`) makes one HTTP call and takes ~1s. This is
  acceptable at startup. The caller must handle the error from the constructor.
- The instruments map is read-only after init — no concurrency concerns.
- Key format `"{exchange}:{tradingsymbol}"` must exactly match what callers pass to
  `FetchCandles`. Document this in the provider package.
- If an instrument is not in the map (delisted, wrong symbol format, wrong exchange prefix),
  the error surfaces at `FetchCandles` call time with a meaningful message — not as a cryptic
  API 400.

## Related decisions

- [Zerodha auth strategy](./2026-04-07-zerodha-auth-strategy.md) — access token required for
  the instruments CSV download (passed in the Authorization header)
- [Zerodha cache strategy](../infrastructure/2026-04-07-zerodha-cache-strategy.md) — the
  `CachedProvider` wraps `ZerodhaProvider`; instrument token lookup happens inside the inner
  provider, transparent to the cache layer
- [Instruments CSV cached at cacheDir/instruments.csv](./2026-05-05-instruments-csv-cache-path.md) — **supersedes** the "no disk cache" element of this decision; disk caching with 24h TTL implemented in TASK-0081
- [InstrumentsCacheDir as explicit Config field](./2026-05-05-instruments-cache-dir-explicit-config-field.md) — companion API design decision

## Revisit trigger

If startup latency from the instruments download becomes a problem (e.g. running many
backtests in a tight loop), persist the instruments CSV to `.cache/zerodha/instruments.csv`
with a 24-hour TTL and load from file on subsequent starts.
