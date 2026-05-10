# run() extraction for testability in cmd/evaluate — providerFactory injection

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, run-function, flag-newFlagSet, providerFactory, cmd/evaluate, TASK-0073 |

## Context

`cmd/evaluate` is a three-stage pipeline orchestrator (universe sweep → walk-forward → bootstrap). Like all other cmd/ binaries in this repo, its logic needs to be unit-testable without spawning a subprocess or holding live Zerodha credentials. The established pattern across `cmd/walk-forward`, `cmd/monitor`, and `cmd/fetch-history` is to extract all wiring into a `run()` function with injectable dependencies.

## Decision

`main()` is a one-liner calling `run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider)`. All pipeline logic lives in `run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error`. The `providerFactory` parameter is the key injection point: production code passes `buildProductionProvider` (which calls `cmdutil.BuildProvider`); tests inject a mock factory returning a `mockProvider`.

This split means all non-provider code paths — flag validation, param parsing, stage directory creation, gate logic, verdict writing — are fully testable by calling `run()` directly with `bytes.Buffer` writers and a mock factory. Provider-dependent paths (universe sweep, WF, bootstrap) remain integration-only, consistent with the pattern across all cmd/ packages.

## Consequences

- `main()` is intentionally untestable (it calls `os.Exit`) — accepted tradeoff in all cmd/ packages.
- All flag validation, date parsing, commission model parsing, and zero-survivor early exit are unit-testable.
- Provider-dependent stages (`runWalkForward`, `runBootstrap`, `collectTrades`) are integration-only paths — same as `walkforward.Run` path in `cmd/walk-forward`.
- Coverage: 65% (integration-only gap in provider-dependent functions; all pure helpers at 100%).

## Related decisions

- [run() extraction for testability in cmd/walk-forward](2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md) — originating pattern
- [run() extraction for testability in cmd/monitor](2026-05-07-run-extraction-for-testability-in-cmd-monitor.md) — second application
- [run() extraction for testability in cmd/fetch-history](2026-05-09-run-extraction-for-testability-in-cmd-fetch-history.md) — providerFactory injection pattern this follows
