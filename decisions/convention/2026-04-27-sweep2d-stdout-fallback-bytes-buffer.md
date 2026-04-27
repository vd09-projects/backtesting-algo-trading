# sweep2d stdout fallback via writeCSVToWriter io.Writer helper

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | sweep2d, stdout, bytes.Buffer, io.Writer, testability, cmd/sweep2d, TASK-0044 |

## Context

`cmd/sweep2d` must write CSV output to either a file (`--out <path>`) or stdout (when `--out` is omitted). The `sweep2d.WriteCSV` function in `internal/sweep2d/csv.go` only accepts a file path — it cannot write to an `io.Writer`. Two approaches existed for the stdout path.

## Options considered

### Option A: `writeCSVToWriter(w io.Writer, report sweep2d.Report2D)` helper in cmd (chosen)

- **Pros**: Stdout path is testable directly in `main_test.go` without temp files or OS pipes. No duplication of CSV serialization logic if the helper is kept in sync with `sweep2d.WriteCSV` format (they share the same spec). Follows the `io.Writer`-in-Config testability pattern established for `output.Config.Stdout`.
- **Cons**: Minor duplication of the CSV format logic between `writeCSVToWriter` (in `cmd/`) and `sweep2d.WriteCSV` (in `internal/`). A format change in `sweep2d.WriteCSV` must also be applied to `writeCSVToWriter`.

### Option B: Write to a temp file then copy to stdout

- **Pros**: Reuses `sweep2d.WriteCSV` exactly — no format duplication.
- **Cons**: Introduces unnecessary filesystem I/O for the common stdout case. Temp file cleanup adds error paths. Makes the smoke test more complex.

### Option C: Extend `sweep2d.WriteCSV` to accept `io.Writer` instead of a path

- **Pros**: Single implementation of CSV format logic.
- **Cons**: Changes the public API of `internal/sweep2d` in a way that touches existing callers (e.g., tests calling `sweep2d.WriteCSV(path, report)`). The file-path API is fine for the file case. Mixing `io.Writer` and path APIs in one function would be awkward.

## Decision

`writeCSVToWriter(w io.Writer, report sweep2d.Report2D)` lives in `cmd/sweep2d/main.go`. `writeOutput` calls `writeCSVToWriter` for stdout and `sweep2d.WriteCSV` for the file path. The smoke test calls `writeCSVToWriter` directly with a `bytes.Buffer` — no temp file, no OS pipes.

The minor format duplication is accepted: both functions are short (~20 lines), and `writeCSVToWriter` has a comment referencing `sweep2d.WriteCSV` to signal the format must stay in sync.

## Consequences

Format changes to `sweep2d.WriteCSV` must also be applied to `writeCSVToWriter`. The reference comment in the function doc is the enforcement mechanism.

## Related decisions

- [io.Writer field in Config for stdout testability](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — the established pattern this follows: injectable io.Writer for testability.

## Revisit trigger

If `sweep2d.WriteCSV` gains more callers that need `io.Writer` semantics, extend its API to accept `io.Writer` and remove `writeCSVToWriter` from `cmd/`.
