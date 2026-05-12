# HandleIncompleteDataError extracted to internal/cmdutil — not duplicated across cmd binaries

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | ErrIncompleteData, exit-code, DRY, diagnostic, TASK-0083 |

## Context

Three cmd binaries (`cmd/backtest`, `cmd/walk-forward`, `cmd/fetch-history`) need identical typed-error inspection: `errors.As(err, &ie *zerodha.ErrIncompleteData)`, print `"incomplete data: instrument=%s from=%s to=%s expected≈%d got=%d\n"` to stderr, return `*ExitCodeError{Code:2}`. The diagnostic format string is part of the acceptance criteria and must be consistent across all three binaries.

## Decision

`HandleIncompleteDataError(err error, stderr io.Writer) *ExitCodeError` lives in `internal/cmdutil/errors.go`. The function does `errors.As` for `*zerodha.ErrIncompleteData`; if matched, prints the exact diagnostic format to stderr and returns `&ExitCodeError{Code: 2}`; if not matched, returns `nil`. Callers test for nil and fall through to their normal error-wrapping path.

This follows the same DRY principle that drove `ParseCommissionModel`, `ParseDateRange`, and `ParseTimeframe` into `cmdutil` — the three-copy threshold was crossed immediately at task design time (backtest + walk-forward + fetch-history).

## Consequences

- The diagnostic format string `"incomplete data: instrument=%s from=%s to=%s expected≈%d got=%d\n"` is defined and tested exactly once in `internal/cmdutil/errors_test.go`.
- `cmd/fetch-history` will use the same helper when TASK-0109 is built — per-instrument failure semantics, not process exit.
- `//nolint:errcheck` on the `fmt.Fprintf` call is documented: writing a diagnostic to stderr is non-fatal; a failed stderr write cannot be meaningfully handled.

## Related decisions

- [ExitCodeError promoted to internal/cmdutil](2026-05-13-exitcodeerror-promoted-to-internal-cmdutil.md) — the error type this helper returns
