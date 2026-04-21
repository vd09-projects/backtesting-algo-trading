# `cmd/correlate` as a new binary rather than extending `cmd/backtest`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | CLI, correlation, cmd/correlate, cmd/backtest, TASK-0027 |

## Context

TASK-0027 required a CLI entry point for computing pairwise correlation across multiple strategy equity curves. The choice was whether to extend `cmd/backtest` with multi-strategy input or create a dedicated binary.

## Decision

New binary at `cmd/correlate`. Correlation requires multiple pre-computed strategy results; it is a post-processing analysis step, not a single-strategy backtest operation. Adding multi-strategy input to `cmd/backtest` would complicate its single-strategy model — `cmd/backtest` is built around one strategy, one instrument, one run.

The `--curve name:path` flag pattern (repeatable, takes parsed CSVs) is a natural fit for a standalone tool. The binary is deliberately thin: parse flags → load CSVs → `analytics.ComputeMatrix` → `output.WriteCorrelationMatrix`.

## Consequences

- `cmd/backtest` stays single-strategy, single-instrument. No flag proliferation.
- `cmd/correlate` can be run across any set of CSV files regardless of how they were produced (backtest, sweep, external tool).
- If multi-strategy portfolio tooling grows (allocation optimizer, combined risk report), it belongs in `cmd/correlate` or a sibling binary, not in `cmd/backtest`.

## Related decisions

- [LoadCurveCSV in internal/output](../architecture/2026-04-21-load-curve-csv-in-output-package.md) — the reader that `cmd/correlate` uses to load curves.
