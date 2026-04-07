# Zerodha historical data — pagination and rate-limit strategy

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | architecture |
| Tags     | zerodha, provider, pagination, chunking, rate-limit, historical-data, kite-connect |

## Context

The Kite Connect historical data endpoint (`GET /instruments/historical/{token}/{interval}`)
enforces a server-side cap on the date range allowed per request. The limits vary by interval
and are **not published in the official documentation** — they are enforced silently (the API
returns fewer candles or an error when the range is too wide). The official SDK (pykiteconnect,
gokiteconnect) makes a single API call per invocation with no chunking — the chunking
responsibility falls entirely on the caller.

`DataProvider.FetchCandles` must support arbitrary date ranges (e.g. 10 years of daily data,
3 months of minute data). Without chunking, a single `FetchCandles` call for a large range
would fail or return partial data.

Separately, the historical endpoint is rate-limited to **3 requests/second**. Sending chunks
without throttling will trigger HTTP 429 responses.

## Confirmed API facts (from official docs and SDK source)

- Rate limit: **3 req/sec** on historical endpoint (HTTP 429 on breach)
- Interval strings: `minute`, `3minute`, `5minute`, `10minute`, `15minute`, `30minute`,
  `60minute`, `day` — no `week` interval
- Date format: `"2006-01-02 15:04:05"` in IST (+05:30)
- Response timestamps: ISO 8601 with +0530 offset — must convert to UTC in the provider

## Date range limits per interval (community-established, verify in Phase 5 prototype)

These are empirically known values. The API enforces them server-side with no explicit error
message beyond returning truncated results or a generic error. **Treat as best estimates until
confirmed against your account/plan tier.**

| model.Timeframe   | Kite interval | Max days/request | Historical depth |
|-------------------|---------------|------------------|------------------|
| Timeframe1Min     | `minute`      | 60 days          | 60 days          |
| Timeframe5Min     | `5minute`     | 100 days         | 100 days         |
| Timeframe15Min    | `15minute`    | 200 days         | 200 days         |
| TimeframeDaily    | `day`         | 2,000 days       | ~20 years        |
| TimeframeWeekly   | —             | N/A              | N/A (unsupported)|

**Prototype verification required:** In Phase 5, make requests at the boundary (e.g. exactly
60 days and 61 days for `minute`) and confirm whether the API returns an error or silently
truncates.

## Decision — chunking strategy

Implement a `chunkDateRange` function inside `pkg/provider/zerodha/` that splits `[from, to)`
into windows no wider than `maxDays[timeframe]` days. The final window may be smaller.

```go
var maxDaysPerInterval = map[model.Timeframe]int{
    model.Timeframe1Min:  55,  // 55 days — 5-day safety margin under the 60-day limit
    model.Timeframe5Min:  90,
    model.Timeframe15Min: 180,
    model.TimeframeDaily: 1800,
}

type dateWindow struct{ from, to time.Time }

func chunkDateRange(from, to time.Time, tf model.Timeframe) []dateWindow {
    maxDays := maxDaysPerInterval[tf]
    step := time.Duration(maxDays) * 24 * time.Hour
    var windows []dateWindow
    cur := from
    for cur.Before(to) {
        end := cur.Add(step)
        if end.After(to) {
            end = to
        }
        windows = append(windows, dateWindow{cur, end})
        cur = end
    }
    return windows
}
```

Safety margins (5–10 days under the published limit) are applied on intraday intervals to
avoid hitting the boundary due to timezone edge cases when `from`/`to` straddle a session
boundary. Daily intervals use a larger window with no safety margin needed.

## Decision — rate limiting strategy

Sleep **350ms between chunk requests** (≈2.85 req/sec, safely under the 3 req/sec cap).
Implemented as a `time.Sleep` after each chunk fetch, skipped after the last chunk.

```go
for i, w := range windows {
    candles, err := c.fetchChunk(ctx, token, interval, w.from, w.to)
    if err != nil { return nil, err }
    all = append(all, candles...)
    if i < len(windows)-1 {
        time.Sleep(350 * time.Millisecond)
    }
}
```

350ms is a fixed constant — not configurable in v1. If the API plan has a higher rate limit,
this can be exposed as a `ProviderConfig.RequestIntervalMs` field later.

## Options considered

### Option A — Fixed sleep between chunks (chosen)
- **Pros:** Simple. No external dependency. Deterministic — easy to test by mocking time.
  Respects the 3 req/sec limit with a comfortable buffer.
- **Cons:** Does not adapt to transient 429s. A burst of 429s will not slow down automatically.

### Option B — Token bucket / leaky bucket rate limiter
- **Pros:** More precise. Can burst up to the limit and recover gracefully.
- **Cons:** Requires `golang.org/x/time/rate` or a hand-rolled implementation. The added
  complexity is not justified for a tool that makes at most ~30 chunk requests per backtest run.

### Option C — Exponential backoff on 429
- **Pros:** Automatically adapts if the limit changes or if we exceed it.
- **Cons:** Adds retry logic, complicates error handling. A well-chosen sleep interval should
  make 429s essentially impossible in practice.

**Option A is sufficient for v1.** Add exponential backoff (Option C) in TASK-0008 if 429s
are observed during the Phase 5 prototype.

## Consequences

- `FetchCandles` for a large date range makes multiple sequential HTTP requests. For 1 year of
  minute data: ~7 chunks × 350ms = ~2.5 seconds. For 10 years of daily data: ~2 chunks ×
  350ms = ~700ms. Both are acceptable for a backtesting tool (and the cache layer in TASK-0009
  eliminates repeat fetches).
- `chunkDateRange` must be unit-tested with known inputs and expected window counts — it must
  not miss any data or double-count at window boundaries. Boundary condition: windows are
  `[from, end)` half-open; the next window starts at `end` exactly.
- The 55/90/180/1800 day constants are in one place (`maxDaysPerInterval`) — easy to update
  after prototype verification without touching the chunking logic.
- Context cancellation (`ctx.Done()`) must be checked between chunk fetches so a long
  multi-chunk fetch can be interrupted cleanly.

## Related decisions

- [Zerodha auth strategy](./2026-04-07-zerodha-auth-strategy.md) — auth happens before the
  first chunk fetch; the access token is passed on every chunk request
- [context.Context deferred from Run() and DataProvider interface](../tradeoff/2026-04-06-context-parameter-deferred.md) — context already threads through FetchCandles; chunk loop must honour ctx.Done()

## Revisit trigger

If the Phase 5 prototype confirms the community limits are wrong, update `maxDaysPerInterval`
before implementing TASK-0008. If 429 errors are observed during prototype testing, implement
Option C (exponential backoff) in TASK-0008 rather than tuning the sleep interval.
