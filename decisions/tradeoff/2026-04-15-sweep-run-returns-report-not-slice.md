# `sweep.Run` returns `SweepReport` (results + plateau) rather than `[]SweepResult`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | return-type, plateau, report, internal/sweep, TASK-0023 |

## Context

The sweep runner produces both a ranked list of results and a plateau analysis. These needed to be surfaced together to the caller. Two design options existed for the return type of `Run`.

## Options considered

### Option A: `Run` returns `SweepReport{Results, Plateau}` (chosen)
- **Pros**: Plateau analysis always runs; caller cannot accidentally skip it. Caller accesses results via `report.Results` if they only need the slice.
- **Cons**: Slightly heavier return type.

### Option B: `Run` returns `[]SweepResult`; caller calls a separate `ComputePlateau` function
- **Pros**: Lighter primary return type.
- **Cons**: Caller must coordinate two calls. Easy to skip the plateau. Increases the sweep package's surface area unnecessarily.

## Decision

`Run` returns `SweepReport` (results + plateau) rather than `[]SweepResult`. This means plateau analysis always happens and the caller doesn't coordinate two calls. If a caller wants just the slice, they take `report.Results`. No caller can accidentally skip the plateau.

## Consequences

The `SweepReport` struct is the canonical output type for the sweep package. `output.WriteSweep` accepts it directly. Any future additions to sweep output (e.g. confidence intervals, best-parameter annotation) extend `SweepReport` rather than adding new return values.
