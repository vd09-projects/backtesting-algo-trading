# Intraday session-close detection via IST timestamp ≥ 15:15, not bar count

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-25       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | intraday, session-boundary, timezone, IST, TASK-0046 |

## Context

When implementing session-boundary detection in the engine (TASK-0046), the question was how to identify the last bar of a trading day — specifically, whether to use a fixed bar count from the start of session or a clock-time comparison. The implementation needs to work correctly across different bar frequencies (5min and 15min bars) without requiring separate configuration per timeframe.

## Options considered

### Option A: IST timestamp comparison against a configurable cutoff (selected)
- **Pros**: The same cutoff time (15:15 IST) works for any bar frequency. A 5min bar at 15:15 and a 15min bar at 15:15 are both detected by the same condition. Timezone-aware and DST-safe. Parameterized via `SessionConfig.SessionCutoff time.Time`.
- **Cons**: Requires timezone conversion per bar — a minor runtime cost that is negligible for daily-bar-scale data volumes.

### Option B: Fixed bar count from the start of session
- **Pros**: Simple arithmetic.
- **Cons**: The bar count per day differs by frequency (75 bars/day for 5min, 25 bars/day for 15min). Hardcoding bar count requires separate configuration per timeframe, which is fragile and error-prone. Does not generalize to other session structures or partial trading days.

### Option C: Bar count from the end (last N bars of the day)
- **Pros**: Works regardless of session length.
- **Cons**: "Last bar" requires knowing the total bar count first — a two-pass approach. Still frequency-dependent.

## Decision

Session-close detection uses an **IST timestamp comparison**: any bar where `timestamp (in IST) >= 15:15` is treated as a session-close candidate. The cutoff 15:15 (not 15:00 or 15:30) gives a buffer — Zerodha begins MIS squareoff at 3:15 PM and the engine must close positions at or before that moment.

The cutoff is parameterized in `SessionConfig.SessionCutoff time.Time` (local IST time-of-day). Priya implements `isLastBarOfSession(bar model.Candle, cfg *SessionConfig) bool` as a timezone-aware comparison that converts bar timestamp to IST before comparing against the cutoff. This makes the logic testable with injected timestamps and correct through DST transitions (though NSE as an IST exchange does not observe DST, IST being UTC+5:30 year-round).

## Consequences

- 5min bars, 15min bars, and any future intraday frequency all use the same session detection logic without configuration changes.
- All existing daily-bar tests are unaffected — `SessionConfig` is a nil-pointer optional field on `engine.Config`; nil means no session boundary enforcement.
- Integration tests must cover both 5min and 15min bar frequencies to confirm the cutoff triggers correctly at the expected bar.

## Related decisions

- [Intraday forced-close fill price: 3:15 PM bar Close](./2026-04-25-intraday-forced-close-fill-price.md) — companion decision specifying what price is used at session close
