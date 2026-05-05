# FetchCandles chunk-merge completeness threshold set at 90% of weekday estimate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-05       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | zerodha, provider, FetchCandles, ErrIncompleteData, completeness-threshold, data-quality, TASK-0081 |

## Context

`FetchCandles` issues multiple chunked requests to cover long date ranges and merges the results. Before TASK-0081, a partial response (e.g., from a mid-run API failure) was returned silently as a short slice — callers had no way to distinguish "no data for this period" from "data was fetched but incomplete." The fix added a completeness check after merge, returning a typed `*ErrIncompleteData` when the result is too short. The threshold determines how much shortfall is acceptable before the error fires.

## Options considered

### Option A: 95% threshold
- **Pros**: Tighter; catches data gaps earlier.
- **Cons**: NSE has ~250 trading days per year; the weekday estimate (Mon–Fri count) overcounts by roughly 10–15 holidays, i.e., 4–6%. A 95% floor would fire on fully-valid fetches with a normal holiday calendar.

### Option B: 90% threshold
- **Pros**: Absorbs the typical NSE holiday gap (~4–5% of weekdays) with margin; still catches meaningful gaps (e.g., a 10%+ missing window from an API failure or market closure).
- **Cons**: Allows up to 10% shortfall before triggering — a genuine 10-day gap on a 100-day request would pass silently.

### Option C: No threshold / always return what we have
- **Pros**: Zero false positives.
- **Cons**: Silently returns partial data; backtests on short slices produce misleading results with no diagnostic.

## Decision

90% threshold (Option B). The expected candle count is computed as `weekdayCount(from, to) * candlesPerDay`, which overcounts NSE trading days by ~4–5% (holidays). A 95% threshold would produce false positives on valid fetches. 90% was initially drafted at 95% during planning and revised to 90% post-build after verifying the holiday calendar math. The typed error `*ErrIncompleteData{instrument, from, to, expected, got}` carries enough context for callers to log a precise diagnostic.

## Consequences

- Fetches missing 10–25% of data (e.g., from partial API responses or long Kite outages) will return `*ErrIncompleteData` rather than a silently short slice.
- Normal fetches with NSE holiday patterns will not trigger the error.
- Callers currently receive `*ErrIncompleteData` as a generic `error` (TASK-0083 tracks adding typed-error handling at the cmd/ layer).
- The weekday estimate is intentionally simple (no NSE holiday calendar dependency) — a more precise estimate would tighten the effective gap but requires maintaining a holiday list.

## Experiments

Threshold reasoning:
- NSE ~252 trading days/year; weekday count ≈ 261 → overcount ~9 days ≈ 3.4%.
- Adding occasional exchange holidays (Republic Day, Diwali, etc.) ≈ 10–15 extra non-trading weekdays/year → total overcount 4–6%.
- 90% threshold leaves 4% headroom above the overcount ceiling.

## Revisit trigger

If a future task introduces an NSE holiday calendar (e.g., for session-boundary detection), the weekday estimate can be replaced with a precise trading-day count, allowing the threshold to be raised to 95–97% without false positives.
