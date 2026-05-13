# run() extraction for testability in cmd/param-search — providerFactory injection

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, run-function, flag-newFlagSet, providerFactory, cmd/param-search, TASK-0077 |

## Context

`cmd/param-search` needed to be testable without live Zerodha credentials. The established pattern across all cmd binaries is to extract all logic from `main()` into a `run(args []string, stdout, stderr io.Writer, providerFactory ...) error` function that can be called from tests with a mock provider.

## Decision

`main()` is a one-liner delegating to `run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider)`. All flag parsing, validation, grid loading, engine orchestration, and CSV output live in `run()`, split across `parseFlags/validateFlags`, `buildSearchPipeline`, and `runSearchPipeline` to stay within the cyclop complexity limit.

`providerFactory func(context.Context) (provider.DataProvider, error)` is injected: `main()` passes `buildProductionProvider`; tests inject `func(_ context.Context) (provider.DataProvider, error) { return mockProvider, nil }`.

Follows the pattern established in `cmd/walk-forward` (2026-05-03), `cmd/monitor` (2026-05-07), `cmd/fetch-history` (2026-05-09), `cmd/evaluate` (2026-05-10).

## Consequences

- `run()` complexity was 16 (max 15 per cyclop). Split into `buildSearchPipeline` + `runSearchPipeline` to stay within the limit — this is a structural consequence of the flag-validation pattern with many required flags.
- `registerFlags(fs *flag.FlagSet)` is exported (lowercase-package scope but package `main`) for use in `TestRun_NoOOSFlag`, which inspects the FlagSet to verify `--oos-from` and `--oos-to` are absent. The architectural enforcement test requires FlagSet inspection, not just parsing behavior.
- 15 cmd-layer tests cover: no-OOS flag enforcement, all missing-required-flag cases, invalid JSON grid, invalid date range, CSV output correctness, out-dir validation, and top-N limiting.

## Related decisions

- [run() extraction for cmd/walk-forward](../../convention/2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md) — original precedent
- [run() extraction for cmd/evaluate](../../convention/2026-05-10-run-extraction-for-testability-in-cmd-evaluate.md) — most recent predecessor with full providerFactory injection

## Revisit trigger

If `buildSearchPipeline` grows more validation steps and hits the cyclop limit again, extract date+timeframe validation into `cmdutil.ParseTrainWindow(fromStr, toStr, tfStr)` matching the DRY pattern in `cmdutil.ParseDateRange`. TASK-0110 is the first step toward this.
