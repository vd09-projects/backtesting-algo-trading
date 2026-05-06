# run() extraction for testability in cmd/monitor

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, flag-parse, coverage, run-function, cmd/monitor, TASK-0048 |

## Context

`cmd/monitor` needed to be testable without spawning subprocesses. The pattern was already
established in `cmd/walk-forward` (2026-05-03): extract `main()`'s wiring into a
`run(args []string, stdout, stderr io.Writer) error` function and use `flag.NewFlagSet` with
`ContinueOnError` so flag-parse errors are returned rather than printed-and-exited.

`cmd/monitor` is a pure file-reading CLI — no `DataProvider`, no Kite Connect credentials, no
network calls. This makes it even more amenable to unit testing via `run()` than
`cmd/walk-forward` (which still has integration-only paths that reach the network).

## Decision

`main()` is a one-liner: `run(os.Args[1:], os.Stdout, os.Stderr)` with `errors.As` unwrapping
for `exitCodeError`. `run()` handles all flag parsing (via `flag.NewFlagSet`), file loading,
threshold evaluation, and alert output. Tests invoke `run()` directly with `bytes.Buffer` writers
and temp-file paths. Coverage: 85.9% of statements, 95% of `run()` itself.

This follows the convention established in `cmd/walk-forward` exactly. As with that package,
the paths that require a live provider remain integration-only; `cmd/monitor` has no such paths
at all, so the coverage gap is only `main()` (0%, expected) and the `exitCodeError.Error()` method
(0%, called only by `main()` which calls `os.Exit`).

## Consequences

- All non-trivial wiring paths in `cmd/monitor` are covered by unit tests.
- `main()` is intentionally untestable (calls `os.Exit`) — accepted tradeoff across all cmd/ packages.
- `flag.NewFlagSet` with `ContinueOnError` produces `--help` output to `stderr` (the passed writer),
  so tests can capture and assert on flag errors without terminal output.

## Related decisions

- [run() extraction for testability in cmd/walk-forward](2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md) — the originating decision for this pattern.
- [flags2D value struct for flag parsing testability in cmd/sweep2d](2026-04-27-flags2d-value-struct-flag-parsing-testability.md) — earlier testability pattern this refines.
