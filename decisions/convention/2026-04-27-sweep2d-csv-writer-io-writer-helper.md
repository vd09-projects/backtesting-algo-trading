# sweep2d CSV writer via io.Writer helper (writeCSVToWriter)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | sweep2d, io.Writer, csv, testability, cmd/sweep2d, TASK-0044 |

## Context

The smoke test in `cmd/sweep2d/main_test.go` needs to verify that a completed sweep produces a CSV with correct column headers. The primary CSV serialization path (`sweep2d.WriteCSV`) writes to a file path, not an `io.Writer`. The smoke test should not require a real file path or temp directory to verify the CSV format.

## Decision

`writeCSVToWriter(w io.Writer, report sweep2d.Report2D)` is an unexported helper in `cmd/sweep2d/main.go` that serializes the Sharpe matrix directly to any `io.Writer`. The smoke test calls it with a `bytes.Buffer` and inspects the string output — no filesystem, no temp directory, no cleanup.

`writeOutput` (the production path) routes to `writeCSVToWriter(os.Stdout, report)` when `--out` is omitted, and to `sweep2d.WriteCSV(path, report)` when `--out` is set.

This follows the same io.Writer testability convention as `output.Config.Stdout`: injectable writer for testing, real writer for production.

## Consequences

The CSV format is implemented twice: once in `sweep2d.WriteCSV` (canonical, file-path API) and once in `writeCSVToWriter` (io.Writer API, cmd-local). Both must produce identical output. A comment in `writeCSVToWriter` cross-references `sweep2d.WriteCSV` to signal the format dependency.

## Related decisions

- [io.Writer field in Config for stdout testability](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — the established project pattern this follows.
- [sweep2d stdout fallback via writeCSVToWriter io.Writer helper](./2026-04-27-sweep2d-stdout-fallback-bytes-buffer.md) — the stdout routing decision that uses this helper.

## Revisit trigger

If `sweep2d.WriteCSV` gains an `io.Writer` overload, remove `writeCSVToWriter` from `cmd/` and route both output paths through the package function.
