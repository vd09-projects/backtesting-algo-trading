# Provider validates API responses via model.NewCandle at parse time

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-08 |
| Status   | accepted   |
| Category | convention |
| Tags     | zerodha, provider, candle, validation, model, NewCandle, parseKiteCandles, convention |

## Context

`parseKiteCandles` decodes the Kite Connect array-of-arrays JSON response into `[]model.Candle`.
Two construction approaches were available: direct struct literal assignment (bypassing
`model.Candle.Validate()`) or calling `model.NewCandle(...)` (which calls `Validate()` internally).

The Kite API has been observed (in practice and in community reports) to return OHLC values of 0
for suspended instruments, instruments with no trades in the period, or edge cases in the API.
These would produce `model.Candle{Open: 0, ...}` which fails `Validate()` ("OHLC values must be
positive").

The question: should the provider accept whatever the API returns and let the engine deal with it,
or reject invalid candles at parse time?

## Options considered

### Option A — Direct struct literal (bypass validation)

```go
c := model.Candle{
    Instrument: instrument,
    Timeframe:  tf,
    Timestamp:  ts.UTC(),
    Open: toFloat64(row[1]), ...
}
candles = append(candles, c)
```

- **Pros:** Provider is "pure" — it just maps API data to structs without opinions about validity.
  Lenient parsing handles malformed data silently.
- **Cons:** A candle with Open=0 propagates into the engine, where a 0-open bar produces
  nonsensical signal calculations and analytics results. The error surfaces far from its source —
  hard to diagnose. The engine and analytics layers trust that candles are valid; adding defensive
  checks there violates their single responsibility.

### Option B — `model.NewCandle` with validation (chosen)

```go
c, err := model.NewCandle(instrument, tf, ts.UTC(), open, high, low, close, volume)
if err != nil {
    return nil, fmt.Errorf("zerodha: candle[%d]: %w", i, err)
}
```

- **Pros:** Invalid data is caught at the boundary where it enters the system — the provider —
  with a clear error pointing to the row index and validation failure. The engine and analytics
  layers can rely on candles being valid, simplifying their logic. The `candle[3]: OHLC values
  must be positive` error is immediately actionable.
- **Cons:** `FetchCandles` returns an error for API responses with invalid rows, even if most rows
  are valid. A future caller that wants partial results would need a different approach. (Not a
  current requirement.)

## Decision

**Option B — `model.NewCandle` with validation.**

Every parsed candle is validated before being added to the result slice. Errors include the row
index and the validation failure reason. The provider is the system's external data boundary —
validation at the boundary is correct regardless of how trustworthy the source is.

This is the general convention: **providers validate; engine and analytics trust**.

## Consequences

- NIFTY 50 index candles with `volume=0` are valid — `Validate()` allows `Volume >= 0`.
- If Zerodha returns OHLC=0 for a suspended instrument, `FetchCandles` returns an error rather
  than silently injecting a zero-bar. The caller must handle this (retry, skip the date, etc.).
- This convention applies to all future `DataProvider` implementations — CSV providers, mock
  providers, etc. Every provider should call `model.NewCandle` rather than constructing the struct
  directly, so that validation is uniform across all data sources.
- If partial results ever become necessary (e.g. "return valid candles, skip invalid ones"),
  add a `strict bool` field to `Config`. Default strict=true preserves current behavior.

## Related decisions

- [Zerodha instrument token lookup](../architecture/2026-04-07-zerodha-instrument-token-lookup.md) — part of the same provider init/fetch design
- [Every Candle/Trade/Position must carry an instrument identifier](../convention/2026-04-02-trade-pnl-stored-not-computed.md) — CLAUDE.md rule that model.NewCandle enforces (Instrument != "")

## Revisit trigger

If a data source legitimately produces candles that fail `Validate()` for non-error reasons
(e.g. a CSV provider where volume is always 0 for a valid reason that Validate doesn't allow),
relax the validation rule in `model.Candle.Validate()` rather than bypassing it in the provider.
