# providerFactory injection added to cmd/backtest and cmd/walk-forward for ErrIncompleteData testability

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, providerFactory, ErrIncompleteData, run-function, TASK-0083 |

## Context

`cmd/backtest/run()` and `cmd/walk-forward/run()` previously called `cmdutil.BuildProvider` directly, constructing the Zerodha provider inside the testable entry point. The `*ErrIncompleteData` handling added by TASK-0083 requires a test that injects a mock provider returning `*ErrIncompleteData` and asserts `run()` returns `*cmdutil.ExitCodeError{Code:2}`. Without provider injection, this path is integration-only.

## Decision

Both binaries add `providerFactory func(context.Context) (provider.DataProvider, error)` as the last parameter to `run()`. `main()` passes `buildProductionProvider`, a local function that calls `cmdutil.LoadDotEnv` and `cmdutil.BuildProvider`. Tests inject a mock factory.

This follows the pattern established by `cmd/universe-sweep` (providerFactory since initial build) and `cmd/fetch-history` (TASK-0070, decision `2026-05-09-run-extraction-for-testability-in-cmd-fetch-history.md`).

`buildProductionProvider` is currently duplicated in `cmd/backtest/main.go` and `cmd/walk-forward/main.go` (identical 3-line functions). Extraction to `cmdutil.BuildProductionProvider` is deferred until `cmd/fetch-history` creates a third copy (TASK-0109) — consistent with the 3-copy extraction rule applied to `ParseCommissionModel` and `BuildProvider`.

## Consequences

- New tests: `TestRun_IncompleteData_ReturnsExitCodeError2` and `TestRun_GenericError_NotExitCodeError2` in both test files.
- `LoadDotEnv` moved from `run()` into `buildProductionProvider` — production behavior unchanged.
- All existing flag-validation tests still call `run(args, stdout, stderr, nil)` since the provider factory is never reached during flag parse failures.

## Related decisions

- [run() extraction for testability in cmd/fetch-history](2026-05-09-run-extraction-for-testability-in-cmd-fetch-history.md) — pattern this follows
