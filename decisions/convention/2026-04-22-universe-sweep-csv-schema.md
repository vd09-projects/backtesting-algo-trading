# Universe-sweep CSV schema: 6 columns, no rank column

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | convention       |
| Tags     | CSV, output, schema, universesweep, TASK-0035 |

## Context

Defining the output CSV schema for `cmd/universe-sweep`. The schema must expose the signal frequency gate result (TASK-0030) and rank by Sharpe without a separate rank column.

## Decision

Six columns: `instrument`, `sharpe`, `trade_count`, `total_pnl`, `max_drawdown`, `insufficient_data`.

- No `rank` column — rows are sorted descending by Sharpe; row position is the rank. Adding a `rank` column would be redundant and would require updating it on any sort change.
- `insufficient_data` is the boolean OR of `analytics.Report.TradeMetricsInsufficient` and `CurveMetricsInsufficient`. Either flag means the result should not be compared against others — the gate is already computed by `analytics.Compute`; no new threshold logic is needed here.
- `sharpe` is zeroed by `analytics.Compute` when `insufficient_data` is true, but the `insufficient_data` column makes this explicit for CSV consumers who don't know the zeroing convention.

## Consequences

Consumers (Python notebooks, shell scripts) can filter `insufficient_data == true` rows before ranking. The schema is intentionally minimal — adding columns later is non-breaking for append-only CSV consumers.
