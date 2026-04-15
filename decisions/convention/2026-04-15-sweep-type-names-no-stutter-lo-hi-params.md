# Sweep types renamed to eliminate stutter; `min/max` params renamed to `lo/hi`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | convention       |
| Tags     | stutter, naming, go-lint, revive, builtinShadow, internal/sweep, TASK-0023 |

## Context

During TASK-0023 implementation, two naming issues surfaced:

1. The initial type names `SweepConfig`, `SweepResult`, `SweepReport` repeat the package name — the revive linter flags these as stutter when used outside the package as `sweep.SweepConfig`.
2. The `paramSteps(min, max float64, n int)` helper function used `min` and `max` as parameter names, shadowing Go 1.21's builtin `min` and `max` functions, which the gocritic `builtinShadow` check catches.

## Decision

`SweepConfig/SweepResult/SweepReport` were renamed to `Config/Result/Report` to eliminate stutter. Internal (test) code in `package sweep` is unaffected since it accesses them without the package prefix. The parameters `min/max` were renamed to `lo/hi` throughout `paramSteps` and its callers.

This follows the project-wide no-stutter convention established in `convention/2026-04-09-no-type-name-stutter-project-wide.md`.

## Consequences

External callers use `sweep.Config`, `sweep.Result`, `sweep.Report` — consistent with how every other package in the repo names its public types (`output.Config`, `analytics.Report`). The rename is consistent and enforceable by linter.

## Related decisions

- [No type-name stutter — project-wide convention](../convention/2026-04-09-no-type-name-stutter-project-wide.md) — this decision applies the repo-wide rule to the sweep package.
