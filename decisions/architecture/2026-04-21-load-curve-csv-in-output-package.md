# `LoadCurveCSV` co-located with `writeCurveCSV` in `internal/output`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | LoadCurveCSV, csv-reader, package-boundary, internal/output, TASK-0027 |

## Context

TASK-0027 required a CSV reader for equity curve files written by `writeCurveCSV`. The question was where to place `LoadCurveCSV` in the package hierarchy.

## Options considered

### Option A: `internal/curveio` — symmetric read/write package

A dedicated package for the CSV schema contract.

- **Pros**: Explicitly names the concern.
- **Cons**: One function. Overkill at this scale. Adds a package to navigate for a trivial operation.

### Option B: `internal/output` (chosen)

Co-locate with `writeCurveCSV`, the function that writes the format `LoadCurveCSV` reads.

- **Pros**: Schema changes (column names, timestamp format, decimal precision) touch one file. Symmetric read/write boundary is self-documenting. No new package overhead.
- **Cons**: `internal/output` is now both a writer and a reader; callers importing it for formatting get the reader too. Acceptable for now.

## Decision

`LoadCurveCSV` lives in `internal/output/load.go`, alongside `writeCurveCSV` in `internal/output/output.go`. The CSV schema (RFC 3339 UTC timestamps, `timestamp,equity_value` header, two decimal places) is the contract between writer and reader.

## Consequences

- `cmd/correlate` imports `internal/output` for `LoadCurveCSV` and for `WriteCorrelationMatrix`. One import, two uses — clean.
- If the CSV schema is ever extended (e.g., additional columns), both functions are in the same package and the change is localized.
- Revisit if a third CSV type is added — at that point a `curveio` package may earn its existence.

## Related decisions

- [`cmd/correlate` as a new binary](../architecture/2026-04-21-cmd-correlate-new-binary.md) — the caller that drives the LoadCurveCSV usage pattern.
