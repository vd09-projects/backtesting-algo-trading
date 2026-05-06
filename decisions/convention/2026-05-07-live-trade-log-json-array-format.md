# Live trade log format: JSON array of model.Trade

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | live-trade-log, JSON, file-format, kill-switch, cmd/monitor, TASK-0048 |

## Context

`cmd/monitor` needs to read a live trade history to evaluate kill-switch thresholds. The
engine already produces `model.Trade` values in backtest runs. The question was how to
represent accumulated live trades on disk.

Two options were considered:

### Option A: JSON array of model.Trade (chosen)

The live trade log is a JSON file containing a `[]model.Trade` array. Each trade is
serialised with Go's default JSON encoder — no custom marshaler:

- `Direction` is `model.Direction`, a `string` type — serialises as `"long"` or `"short"`.
- `EntryTime` and `ExitTime` are `time.Time` — serialise as RFC 3339 UTC strings.
- All numeric fields are plain `float64`.

Example:
```json
[
  {
    "Instrument": "NSE:SBIN",
    "Direction": "long",
    "Quantity": 10,
    "EntryPrice": 620.50,
    "ExitPrice": 635.00,
    "EntryTime": "2025-01-15T09:15:00Z",
    "ExitTime": "2025-01-22T15:30:00Z",
    "Commission": 48.20,
    "RealizedPnL": 96.80
  }
]
```

**Pros:** No new schema. Backtest output (from `cmd/backtest --out`) can be used directly
as a seed for the live log. `json.Unmarshal` handles everything. Adding a live trade means
appending a JSON object to the array — straightforward for any tooling.

**Cons:** JSON arrays require re-writing the whole file to append atomically. For weekly
monitoring cadence this is not a concern — the file is small and writes are infrequent.

### Option B: Append-only CSV

A custom CSV schema (instrument, direction, quantity, entry_price, exit_price, entry_time,
exit_time, commission, realized_pnl).

**Rejected:** Requires a custom parser and a separate schema decision. The CSV format
adds no expressive benefit over JSON for this domain. JSON already handles time, direction
strings, and numeric precision without custom quoting logic.

## Decision

JSON array of `model.Trade`. One file per strategy+instrument, named by convention
(e.g. `live-trades-macd-SBIN.json`). The user appends closed trades after each
session. `cmd/monitor --trades <path>` reads the full file.

## Consequences

- `cmd/monitor` uses `json.Unmarshal` with `[]model.Trade` — no custom parser.
- Backtest output JSON (the `trades` key from `cmd/backtest --out`) can be extracted
  and used directly as the live trade log seed.
- If the trade log grows large (hundreds of trades over months), the user can prune it
  to the monitoring window (e.g. last 6 months) manually. `cmd/monitor` applies no
  rolling window — the caller controls the window via the file contents.
- If an append-only format becomes necessary (e.g. concurrent writers), migrate to
  newline-delimited JSON (NDJSON) at that point. The change is localised to `loadTrades`
  in `cmd/monitor`.

## Related decisions

- [Kill-switch derivation methodology](../algorithm/2026-04-21-kill-switch-derivation-methodology.md) — defines the three thresholds cmd/monitor evaluates.
- [Kill-switch API keeps analytics free of montecarlo](../architecture/2026-04-21-kill-switch-analytics-to-montecarlo-boundary.md) — architecture context for the analytics package cmd/monitor imports.
