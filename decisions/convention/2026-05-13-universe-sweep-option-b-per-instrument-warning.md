# cmd/universe-sweep *ErrIncompleteData handling — Option B: per-instrument warning, sweep continues

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | ErrIncompleteData, universe-sweep, per-instrument-warning, sweep-continues, TASK-0083 |

## Context

TASK-0083 required deciding how `cmd/universe-sweep` handles `*ErrIncompleteData` from `FetchCandles`. The original plan (Option A) was to abort the sweep with exit code 2. The alternative (Option B) is to treat incomplete data as a per-instrument deficiency and continue.

## Options considered

### Option A: Abort the sweep on first *ErrIncompleteData (original plan)
- **Pros**: Simple; operator always knows the sweep completed with full data.
- **Cons**: A single instrument with partial data ruins the entire sweep. For a 50-instrument universe, one bad fetch forces a re-run of the entire sweep.

### Option B: Per-instrument warning, sweep continues (chosen)
- **Pros**: The sweep produces a complete report; incomplete instruments are flagged `insufficient_data=true` in the CSV and excluded from the DSR gate calculation. The operator gets actionable output even when some instruments have partial data.
- **Cons**: The operator must check the CSV's `insufficient_data` column or stderr for per-instrument diagnostics.

## Decision

Option B, per user decision. `*ErrIncompleteData` in `internal/universesweep.runInstrument` logs the per-instrument diagnostic to stderr and returns `Result{Instrument: instrument, InsufficientData: true}` with nil error — the sweep continues. The check belongs in `internal/universesweep.runInstrument()`, not at the `cmd/universe-sweep/run()` level.

`cmd/backtest` and `cmd/walk-forward` (single-instrument, not bulk sweep) retain fatal exit code 2 — partial data on a targeted single-instrument run is not recoverable.

## Consequences

- The diagnostic format `"incomplete data: instrument=%s from=%s to=%s expected≈%d got=%d\n"` is written to the caller's stderr once per incomplete instrument.
- Incomplete instruments appear in the CSV with `insufficient_data=true`; their rows have zero Sharpe and zero trade count.
- `ApplyUniverseGate` already excludes `InsufficientData=true` results from the DSR calculation — no gate-logic changes needed.
- The operator must inspect stderr or the CSV's `insufficient_data` column to identify data quality issues.
