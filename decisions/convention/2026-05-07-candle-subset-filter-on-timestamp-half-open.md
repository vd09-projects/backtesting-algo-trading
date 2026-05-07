# Candle subset filter uses Timestamp field with half-open [from, to) interval

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | cache, CachedProvider, filterCandles, Timestamp, half-open-interval, convention, TASK-0089 |

## Context

TASK-0089 introduced `filterCandles(candles []model.Candle, from, to time.Time) []model.Candle` to slice a superset candle set down to the requested `[from, to)` window. Two choices needed to be made: which `Candle` field to compare against the window, and whether the interval is closed or half-open.

## Options considered

### Option A: Filter on Candle.Timestamp, half-open [from, to) (chosen)

`!ts.Before(from) && ts.Before(to)` — include candles where `from <= Timestamp < to`.

- **Pros**: `Timestamp` is the bar's identity field — it is how the engine and strategies locate a bar in time. Half-open interval matches the documented semantics of `FetchCandles(ctx, instrument, tf, from, to)` throughout the codebase (callers treat `to` as exclusive). Consistent convention avoids off-by-one confusion at call sites.
- **Cons**: None identified.

### Option B: Filter on Candle.Timestamp, closed [from, to] (rejected)

`!ts.Before(from) && !ts.After(to)` — include candles where `from <= Timestamp <= to`.

- **Pros**: Might feel more natural if `to` is specified as an inclusive date.
- **Cons**: Inconsistent with existing `FetchCandles` documentation and usage across `cmd/backtest`, `cmd/universe-sweep`, and `cmd/walk-forward`, where `to` is always exclusive. Mixing conventions would require callers to adjust date arithmetic at the superset-hit path vs the network-fetch path.

## Decision

`Candle.Timestamp` with half-open `[from, to)` interval. `Timestamp` is the primary identity field on `Candle` (all engine and strategy code navigates bars by timestamp). The half-open convention matches the existing `FetchCandles` API semantics everywhere in the codebase — `to` is always the exclusive upper bound. Using a closed interval would introduce a subtle inconsistency that callers would need to paper over.

## Consequences

- `filterCandles` always returns `nil` (not an empty slice) when no candles fall in the window. Callers must handle `nil` as a valid empty result. This is consistent with Go slice idioms.
- The filter is timestamp-based, not calendar-aware — it does not know about NSE trading days or holidays. This is correct: the cache layer is data-agnostic; calendar awareness belongs to the provider or strategy layer.
- Any future refactor that makes `FetchCandles` use closed intervals would also need to update `filterCandles`.

## Revisit trigger

If the `FetchCandles` API is ever changed to use a closed `[from, to]` interval, update `filterCandles` to match.
