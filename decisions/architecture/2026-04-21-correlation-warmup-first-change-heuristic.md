# Warmup detection by first-change heuristic in `alignAndTrim`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | warmup-detection, correlation, alignAndTrim, TASK-0027 |

## Context

TASK-0027 required trimming warmup bars (leading bars where the equity curve is flat at the initial value) before computing Pearson correlation. The question was how `ComputeCorrelation` learns where each strategy's warmup ends.

Two approaches were considered: an explicit `warmup int` parameter threaded through the call chain, or detecting warmup automatically from the curve itself.

## Options considered

### Option A: Explicit warmup parameter

Pass each strategy's lookback period into `ComputeCorrelation` as an integer. Trim that many bars from the front of each curve.

- **Pros**: Precise, no inference.
- **Cons**: Requires callers to know each strategy's lookback. `cmd/correlate` loads CSV files — it has no access to strategy config. Ties correlation analysis to strategy internals, violating the clean file-boundary architecture (Go writes CSVs, caller loads them).

### Option B: First-change heuristic (chosen)

Scan each curve from index 0. The warmup end is the first index where `curve[i].Value != curve[0].Value`. Take the maximum across all curves being compared.

- **Pros**: `ComputeCorrelation` is completely decoupled from strategy configuration. Works for any strategy without extra wiring. A strategy that never trades produces an all-flat series → NaN (correct behavior — undefined correlation).
- **Cons**: If a strategy's first bar happens to differ from the initial value for a reason other than trading (e.g., a pricing adjustment), the heuristic could trim less than intended. Acceptable risk given the CSVs are engine output and the initial value is always the starting capital.

## Decision

First-change heuristic. `firstActiveIndex` returns the index of the first bar where `curve[i].Value != curve[0].Value`. `alignAndTrim` takes the max across both curves, then trims and aligns to equal length.

## Consequences

- `ComputeCorrelation` signature stays `(a, b NamedCurve) PairCorrelation` — no config leak.
- Strategies that never trade → constant series → NaN (not a bug; correct signal that the strategy had no activity).
- Walk-forward folds and CSV-loaded curves both work without extra parameters.

## Related decisions

- [math.NaN() sentinel for undefined correlation](../convention/2026-04-21-correlation-nan-sentinel-undefined.md) — the downstream convention for constant-series output.
- [LoadCurveCSV in internal/output](../architecture/2026-04-21-load-curve-csv-in-output-package.md) — the boundary that makes explicit warmup parameters impractical.
