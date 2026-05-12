# universesweep.Run and runInstrument gain stderr io.Writer for per-instrument diagnostic logging

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | ErrIncompleteData, per-instrument-warning, stderr, io.Writer, TASK-0083 |

## Context

`internal/universesweep.Run` fans out per-instrument engine runs via errgroup. When an instrument's `FetchCandles` returns `*ErrIncompleteData`, the correct behavior (Option B) is to log a per-instrument diagnostic to stderr and continue the sweep. The question is how to get the caller's stderr into the goroutines.

## Options considered

### Option A: Write directly to os.Stderr inside runInstrument
- **Pros**: No signature change.
- **Cons**: Tests cannot capture the diagnostic output — they'd have to redirect os.Stderr via subprocess. Inconsistent with the project's established pattern of injecting stderr as `io.Writer` for testability.

### Option B: Add stderr io.Writer to Run and runInstrument (chosen)
- **Pros**: Tests pass `*bytes.Buffer` and assert the diagnostic content. Consistent with `cmd/walk-forward`, `cmd/universe-sweep`, `cmd/evaluate`, `cmd/monitor` — all inject stderr as io.Writer.
- **Cons**: Signature change requires updating all callers (`cmd/universe-sweep`, `cmd/evaluate`). Two callers total — minor churn.

## Decision

`Run(ctx, cfg, p, stderr io.Writer)` and `runInstrument(ctx, cfg, p, instrument, stderr io.Writer)` gain the `stderr` parameter. Callers pass their own `stderr io.Writer`. Tests pass `*bytes.Buffer` and assert `strings.Contains(stderrOut, "incomplete data:")`.

This follows the pattern established by `cmd/walk-forward`'s `run(args, stdout, stderr)` extraction (decision `2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md`).

## Consequences

- Two callers updated: `cmd/universe-sweep/main.go` passes its `stderr`; `cmd/evaluate/main.go` passes its `stderr` in `runUniverseSweep`.
- Note: the `stderr io.Writer` is shared across all goroutines spawned by errgroup. `os.Stderr` (production) is safe — small write(2) syscalls are atomic on Linux/macOS. `*bytes.Buffer` (tests) is not goroutine-safe for concurrent writes. TASK-0107 tracks adding a `syncWriter` wrapper.

## Revisit trigger

When TASK-0107 is implemented (syncWriter wrapper for stderr in Run), update this decision to reflect that the writer is synchronized before goroutine launch.
