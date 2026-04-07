# TimeframeWeekly excluded from Zerodha SupportedTimeframes

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | convention |
| Tags     | zerodha, provider, timeframe, weekly, kite-connect, model, SupportedTimeframes |

## Context

`pkg/model/timeframe.go` defines five timeframes:
- `Timeframe1Min`, `Timeframe5Min`, `Timeframe15Min`, `TimeframeDaily`, `TimeframeWeekly`

During Phase 1 research (TASK-0007), the complete list of Kite Connect historical API interval
strings was confirmed from the official documentation:
`minute`, `3minute`, `5minute`, `10minute`, `15minute`, `30minute`, `60minute`, `day`

There is **no `week` interval** in the Kite Connect API. Weekly candles are not available from
the historical endpoint.

## Decision

`ZerodhaProvider.SupportedTimeframes()` returns:
```go
[]model.Timeframe{
    model.Timeframe1Min,
    model.Timeframe5Min,
    model.Timeframe15Min,
    model.TimeframeDaily,
    // TimeframeWeekly intentionally omitted — not supported by Kite Connect API
}
```

`TimeframeWeekly` remains in `pkg/model/timeframe.go` — it is a valid model type and may be
served by a future provider (e.g. a CSV file provider or a different broker). The Zerodha
provider simply does not advertise it.

The engine should validate that `strategy.Timeframe()` is in `provider.SupportedTimeframes()`
before calling `FetchCandles`. If a strategy requests weekly candles and the Zerodha provider
is wired, the engine should return a descriptive error rather than making a failing API call.

## Consequences

- Any strategy requiring weekly candles cannot be backtested against Zerodha data in v1.
  This is an explicit, documented limitation — not a silent failure.
- The `TimeframeWeekly` → `Duration()` method in `model/timeframe.go` returns correctly
  (`7 * 24 * time.Hour`) — this is accurate and should not be removed.
- If Zerodha ever adds a weekly interval, add `TimeframeWeekly` back to `SupportedTimeframes()`
  and map it to the new interval string.

## Related decisions

- [Zerodha pagination strategy](../architecture/2026-04-07-zerodha-pagination-strategy.md) —
  the interval→maxDays map does not include an entry for TimeframeWeekly
