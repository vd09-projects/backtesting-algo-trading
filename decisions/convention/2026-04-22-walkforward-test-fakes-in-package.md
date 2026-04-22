# In-package test fakes defined in _test.go only

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | convention       |
| Tags     | walkforward, test-fake, provider, strategy, _test.go, TASK-0022 |

## Context

`internal/walkforward` tests need a fake `provider.DataProvider` and a fake `strategy.Strategy`. The question was where to define them — in-package (unexported, _test.go only), or exported from `pkg/provider` or `pkg/strategy` for reuse by other test files.

## Decision

In-package fakes (`staticProvider`, `toggleStrategy`, `neverTradeStrategy`) are defined in `walkforward_test.go` only and are not exported. Exporting test infrastructure from production packages (`pkg/provider`, `pkg/strategy`) for the benefit of a single test file sends dependencies in the wrong direction — production packages would depend on test concerns. The fakes are simple enough to duplicate if another package needs similar behavior.

## Consequences

If a future package needs the same fake data provider or strategy, it defines its own local version. This is acceptable duplication; shared test helpers belong in a dedicated `testutil` package only if the pattern becomes sufficiently widespread (3+ callers).
