# run() extraction for testability in cmd/fetch-history

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, flag-parse, coverage, run-function, cmd/fetch-history, TASK-0070 |

## Context

`cmd/fetch-history` needed to be testable at the unit level — flag parsing, dry-run behavior,
partial-failure manifest writes, and resume-from-manifest all needed test coverage without spawning
subprocesses or requiring live Zerodha credentials.

The pattern was already established in two prior cmd packages:
- `cmd/walk-forward` (2026-05-03): `run(args []string, stdout, stderr io.Writer) error`
- `cmd/monitor` (2026-05-07): same pattern; `cmd/monitor` has no DataProvider dependency at all

`cmd/fetch-history` extends the pattern with a `providerFactory` parameter (see separate decision
`2026-05-09-provider-factory-receives-parsed-flags.md`).

## Decision

`main()` is a one-liner:
```go
func main() {
    err := run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider)
    if err != nil { ... os.Exit(1) }
}
```

`run()` uses `flag.NewFlagSet("fetch-history", flag.ContinueOnError)` with `fs.SetOutput(stderr)`
so flag-parse errors are returned as errors rather than printed-and-exited. Tests invoke `run()`
directly with `bytes.Buffer` writers, `t.TempDir()` cache directories, and mock provider factories.

Flag parsing is extracted to a separate `parseFlags(args []string, stderr io.Writer) (fetchFlags, error)`
helper to keep `run()` within the cyclomatic complexity limit (cyclop max=15).

## Consequences

- `main()` remains intentionally untestable (calls `os.Exit`) — accepted tradeoff across all cmd packages.
- Coverage: 79.5% of statements. `main()` (0%) and `buildProductionProvider` (0%, integration-only)
  account for the gap.
- 12 test functions covering: all required-flag validations, dry-run output, dry-run no-manifest-written,
  successful manifest write, partial-failure manifest (only first instrument), resume-from-manifest
  (mock call count = 1), env-var fallback, progress logging, and multiple timeframes.

## Related decisions

- [run() extraction for testability in cmd/walk-forward](2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md) — originating decision.
- [run() extraction for testability in cmd/monitor](2026-05-07-run-extraction-for-testability-in-cmd-monitor.md) — second application; the pattern this follows.
- [providerFactory receives parsed flags](2026-05-09-provider-factory-receives-parsed-flags.md) — the extension specific to cmd/fetch-history.
