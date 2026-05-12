# cmd/backtest and cmd/walk-forward: exit code 2 for *ErrIncompleteData, exit code 1 for generic errors

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | exit-code, ErrIncompleteData, main, cmd/backtest, cmd/walk-forward, TASK-0083 |

## Context

Before TASK-0083, `cmd/backtest/main()` called `cmdutil.Fatalf` (exit 1) for all errors. `cmd/walk-forward/main()` had an `errors.As` check for `*exitCodeError` to handle exit 1 for WF gate failures, but no path for exit 2. Neither binary distinguished "partial data from provider" from "generic failure" at exit-code level.

## Decision

Both binaries now follow this pattern in `main()`:

```go
if err := run(...); err != nil {
    var ee *cmdutil.ExitCodeError
    if errors.As(err, &ee) {
        os.Exit(ee.Code)
    }
    cmdutil.Fatalf("%v", err)
}
```

Exit code mapping:
- **2** — `*ErrIncompleteData` from FetchCandles (incomplete candle data for the instrument). `HandleIncompleteDataError` returns `*ExitCodeError{Code:2}` and prints the diagnostic.
- **1** — all other errors via `cmdutil.Fatalf`.
- **1** — `cmd/walk-forward` WF gate failure (`*ExitCodeError{Code:1}` returned when `determineExitCode(report) != 0`).

Unix convention for exit codes: 1 = generic error, 2 = misuse/data problem. Exit code 2 for incomplete data follows the same reasoning as `cmd/walk-forward`'s use of exit code 1 for gate failure — it enables scripting and CI to distinguish error categories without parsing stderr.

## Consequences

- Callers can `case $? in 2) echo "incomplete data — check fetch-history cache" ;; esac` in shell scripts.
- The `errors.As` check in `main()` must remain before `cmdutil.Fatalf` — otherwise all errors get exit 1 regardless of type.
- `cmd/universe-sweep` does NOT use exit code 2 — it continues the sweep on incomplete data (see `decisions/convention/2026-05-13-universe-sweep-option-b-per-instrument-warning.md`).
