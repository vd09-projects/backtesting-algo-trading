# run() extraction for testability in cmd/walk-forward

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, cmd/walk-forward, run-function, flag-newFlagSet, exitCodeError, TASK-0066 |

## Context

`cmd/walk-forward/main.go` initially had all wiring in `main()`: flag parsing, strategy factory construction, config building, `walkforward.Run()` call, and exit-code determination. This made the package untestable at the unit level — `main()` cannot be called from tests without spawning a subprocess. The quality gate flagged `cmd/walk-forward` at 47.2% coverage, and `parseAndValidateFlags` was unreachable from tests that compiled in the same package.

## Decision

Extract main's wiring into `run(args []string, stdout, stderr io.Writer) error`. `main()` becomes a one-liner that calls `run(os.Args[1:], os.Stdout, os.Stderr)` and converts the error to an exit code. `run()` uses `flag.NewFlagSet(name, flag.ContinueOnError)` so flag-parse errors are returned (not printed-and-exited), allowing tests to assert on error text and exit-code behavior without subprocesses.

Exit-code routing uses a typed `exitCodeError` sentinel: when `walkforward.Run()` returns a report with `OverfitFlag` or `NegativeFoldFlag` set, `run()` wraps the code in `exitCodeError{code: 1}`; `main()` unwraps it via `errors.As` and calls `os.Exit`. All other errors produce exit code 1 via the normal non-nil error path.

This pattern was established for `cmd/sweep2d` (flags2D value struct) and extended here. Coverage reached 73.4% (above 70% threshold) after adding tests for flag-parse failure, unknown strategy, invalid commission, and to-before-from paths via `run()`.

## Consequences

- `main()` is untestable by design (it calls `os.Exit`); this is the accepted tradeoff across all cmd packages.
- All non-trivial wiring paths are reachable from unit tests via `run()`.
- `exitCodeError` is private to the package — it is not a general convention, just a local dispatch mechanism.
- Paths that call `walkforward.Run()` remain uncovered at unit level (require a live DataProvider); these are integration-only paths, not a gap in the gate.

## Related decisions

- [flags2D value struct for flag parsing testability in cmd/sweep2d](../convention/2026-04-27-flags2d-value-struct-flag-parsing-testability.md) — same testability convention applied to 2D sweep CLI
- [io.Writer field in Config for stdout testability](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — foundational io.Writer injection convention this builds on
