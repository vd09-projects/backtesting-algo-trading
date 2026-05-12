# ExitCodeError promoted to internal/cmdutil, shared across cmd binaries

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | exit-code, errors, DRY, ErrIncompleteData, TASK-0083 |

## Context

`cmd/walk-forward` had a local unexported `exitCodeError{code int}` type that `run()` returned to signal `main()` to call `os.Exit(1)` for walk-forward gate failures. When TASK-0083 added exit-code-2 handling for `*ErrIncompleteData` across `cmd/backtest` and `cmd/walk-forward`, each binary would have needed its own copy of the same three-line struct.

## Options considered

### Option A: Keep exitCodeError local in each cmd binary
- **Pros**: No cross-package coupling; each binary is self-contained.
- **Cons**: Three identical three-line structs in three `package main` files with no way to share tests.

### Option B: Promote to internal/cmdutil as ExitCodeError (chosen)
- **Pros**: Single definition; all cmd binaries import it from `cmdutil` (already imported by all of them). `main()` pattern is uniform: `errors.As(err, &ee) → os.Exit(ee.Code)`.
- **Cons**: `cmdutil` now exports an error type it owns rather than the calling binary owning it.

## Decision

`ExitCodeError{Code int}` is defined once in `internal/cmdutil/errors.go`. `cmd/walk-forward`'s local `exitCodeError` is removed. All single-instrument cmd binaries (`backtest`, `walk-forward`) use `*cmdutil.ExitCodeError` from their `main()` functions. The exported `Code` field (PascalCase) replaces the unexported `code` field.

## Consequences

- Any new cmd binary that needs non-1 exit codes imports `cmdutil.ExitCodeError` directly — no boilerplate.
- `walk-forward`'s `TestExitCodeError_Error` test moved to `internal/cmdutil/errors_test.go`.
- `cmd/universe-sweep` does not use `ExitCodeError` — it never aborts on `*ErrIncompleteData`, so no exit-code-2 path exists there.

## Related decisions

- [HandleIncompleteDataError extracted to internal/cmdutil](2026-05-13-handleincompletedataerror-extracted-to-cmdutil.md) — companion decision; the helper that returns `*ExitCodeError{Code:2}`
