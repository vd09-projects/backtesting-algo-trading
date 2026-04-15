# Sweep plateau uses non-contiguous min/max range over all qualifying entries

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | sweep, parameter-sweep, plateau, non-contiguous, internal/sweep, TASK-0023 |

## Context

The plateau detection needs to identify the range of parameter values that produce "good enough" Sharpe ratios. A real parameter response curve is rarely perfectly smooth — there may be one slightly-worse parameter value in the middle of an otherwise good region.

## Options considered

### Option A: Collect min/max ParamValue of all qualifying entries, regardless of contiguity (chosen)
- **Pros**: Conservative interpretation. Tells the trader "there are good values inside this range, verify each one." Simpler to compute. Resistant to noise in the sweep results.
- **Cons**: A non-contiguous gap widens the reported range, potentially overstating the stable region.

### Option B: Split into multiple contiguous ranges
- **Pros**: More precise reporting of separate stable regions.
- **Cons**: More complex to implement. Adds output complexity (multiple plateau ranges). The threshold for "what counts as a gap" would need an arbitrary parameter itself.

## Decision

The plateau calculation receives results pre-sorted descending by Sharpe and collects min/max ParamValue of all qualifying entries, regardless of contiguity. A non-contiguous gap widens the reported range rather than splitting it into two — the conservative interpretation, which matches what a trader using this tool needs to know: "there are good values inside this range, but verify each one."

`output.WriteSweep` lives in `internal/output` rather than `internal/sweep` because output formatting is the output package's responsibility, and the sweep package has no reason to know about `io.Writer`. The import direction is `output → sweep`, not the reverse.

## Consequences

The plateau is reported as a single `[MinParam, MaxParam]` range with a count and minimum Sharpe. If two separate parameter islands both qualify, they are reported as one wider range — a known limitation. For v1, this is acceptable; the user can inspect the full ranked table.

## Revisit trigger

If users report that the plateau range misleads them (e.g., a clearly bad parameter value inside an otherwise good range creates a false sense of stability), revisit and implement split-range detection.
