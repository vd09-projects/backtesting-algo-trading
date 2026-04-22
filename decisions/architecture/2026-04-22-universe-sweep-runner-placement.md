# Universe-sweep runner lives in `internal/universesweep`, not inline in `cmd/`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | package-boundaries, cmd, internal, universesweep, TASK-0035 |

## Context

Adding `cmd/universe-sweep` required deciding where the fan-out and result aggregation logic should live. The precedent is `internal/sweep` (runner with testable domain logic) vs a thin `cmd/` binary with all logic inline.

## Options considered

### Option A: Logic inline in `cmd/universe-sweep/main.go`
- **Pros**: Fewer files; no package boundary to cross.
- **Cons**: `ParseUniverseFile`, `Run`, and `WriteCSV` are independently testable units. Putting them in `main.go` makes them untestable without a subprocess harness.

### Option B: `internal/universesweep` package (chosen)
- **Pros**: Follows the `cmd/sweep` vs `internal/sweep` precedent. Acceptance criterion ("Tests: synthetic 2-instrument universe → 2-row output CSV") is naturally a package-level test. cmd stays thin.
- **Cons**: One more package.

## Decision

Runner logic lives in `internal/universesweep`. `ParseUniverseFile`, `Run`, and `WriteCSV` are all exported functions with unit tests. `cmd/universe-sweep/main.go` is thin: parse flags, construct provider, call the package, write to stdout.

The deciding factor: the acceptance criterion explicitly requires a test that calls `Run` and checks the CSV output. That test belongs in the package, not in a `cmd/` integration test. The `internal/sweep` pattern exists for exactly this reason.

## Consequences

`cmd/universe-sweep` imports `internal/universesweep`. Any caller that wants to run a universe sweep programmatically can import the package directly. The cmd binary is not the only entry point.
