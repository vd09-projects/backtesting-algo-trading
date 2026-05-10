# Skip OHLC-invalid candles in parseKiteCandles; hard-fail on structural errors

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | bad-candle, skip, OHLC-validation, structural-error, parseKiteCandles, TASK-0100 |

## Context

`parseKiteCandles` iterates the Kite API's array-of-arrays candle response. Before this change, any error — whether a malformed row, an unparseable timestamp, or an OHLC validation failure from `model.NewCandle` — caused the entire chunk to abort with an error.

The Zerodha artifact produces a well-formed row (6 elements, valid timestamp, numeric OHLC values) where the open price sits marginally below the low (e.g. open=835.6, low=837.4). The row can be interpreted; only one price relationship is violated. Aborting the entire chunk discards ~98,900 valid candles because of one bad bar.

## Decision

Two classes of error are distinguished within `parseKiteCandles`:

**Structural errors → hard fail the chunk:**
- Row has fewer than 6 elements (`len(row) < 6`)
- Timestamp field is not a string
- Timestamp is unparseable by both the Kite format and RFC 3339 fallback

In these cases, the row cannot be meaningfully interpreted. Proceeding would produce a candle series with a missing or corrupted position — potentially corrupting the index-based `SkippedCandle.Index` for downstream consumers. The chunk returns `nil, nil, err`.

**OHLC validation errors → skip the candle and continue:**
- `model.NewCandle` returns a non-nil error (any validation failure: open outside [low,high], close outside [low,high], negative volume, zero OHLC, etc.)

In these cases the row is structurally intact — only the price relationship is violated. The candle is added to the `[]SkippedCandle` accumulator with `{Index: i, Reason: "zerodha: candle[N]: " + err.Error()}` and the loop continues.

The `completeness check` (90% of weekday estimate) acts as the safety net: if too many candles are skipped, `*ErrIncompleteData` fires and the entire fetch is treated as a data gap.

## Consequences

- `parseKiteCandles` signature changed from `([]model.Candle, error)` to `([]model.Candle, []SkippedCandle, error)`. All callers within the package (`fetchChunk`, `FetchCandles`) updated accordingly.
- The existing test `TestParseKiteCandles_non_float_value_returns_error` (which expected a hard error when Open=0) was updated to `TestParseKiteCandles_non_float_value_skips_candle` — Open=0 is now a skipped candle, not a hard abort.
- Structural errors (short row, non-string timestamp, unparseable timestamp) are still hard errors — their tests are unchanged.

## Revisit trigger

If Zerodha introduces a new class of API artifact that corrupts the row structure (e.g. missing volume field with 5 elements instead of 6), the structural-error bucket handles it correctly. If a new price-relationship artifact appears that cannot be isolated to a single candle (e.g. a chunk where 100% of candles have bad OHLC), the completeness check catches it.
