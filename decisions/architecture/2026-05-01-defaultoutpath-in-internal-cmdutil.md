# DefaultOutPath placed in internal/cmdutil

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | filename, default-out, run-config, cmdutil, DefaultOutPath, TASK-0064 |

## Context

TASK-0064 required `cmd/backtest` to auto-generate a canonical output filename when `--out` is omitted. The filename follows the pattern `{strategy}-{instrument}-{timeframe}-{from}-{to}.json`. The question was where `DefaultOutPath` should live: inline in `cmd/backtest/main.go`, in a shared `internal/cmdutil` package, or somewhere else.

## Decision

`DefaultOutPath` is placed in `internal/cmdutil` alongside `ParseCommissionModel` and `BuildProvider`. These three functions share the same character: they are cmd-layer plumbing needed by cmd binaries to translate CLI inputs into typed engine inputs. `internal/cmdutil` is already the established home for this kind of helper. Adding `DefaultOutPath` there keeps `cmd/backtest/main.go` free of filename-generation logic and makes the helper available to future cmd binaries (e.g., if `cmd/sweep` or `cmd/universe-sweep` ever need a canonical output path).

The instrument name is sanitized for filesystem safety inside `DefaultOutPath` (`:` and ` ` replaced with `_`). The original instrument name (with colon) is preserved in the JSON content via `RunConfig.Instrument`.

## Consequences

`cmd/backtest` calls `cmdutil.DefaultOutPath(...)` post-parse and assigns the result to `f.outPath` only when the flag is empty. The function is pure (no side effects) and fully covered by `TestDefaultOutPath` in `cmdutil_test.go`.

## Related decisions

- [ParseCommissionModel extracted to internal/cmdutil](../convention/2026-04-29-parse-commission-model-extracted-to-cmdutil.md) — established the `internal/cmdutil` pattern for cmd-layer helpers
- [BuildProvider extracted to cmdutil](../architecture/2026-04-22-buildprovider-extracted-to-cmdutil.md) — first extraction into `internal/cmdutil`
