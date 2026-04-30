# ParseCommissionModel extracted to internal/cmdutil

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-29       |
| Status   | experimental     |
| Category | convention       |
| Tags     | commission, DRY, cmdutil, flag-parsing, cmd/backtest, cmd/sweep, cmd/universe-sweep, TASK-0051, TASK-0060 |

## Context

TASK-0051 added a `--commission` flag to `cmd/backtest` and `cmd/sweep` to allow callers to specify the commission model (zerodha | zerodha_full | zerodha_full_mis | flat | percentage) rather than having it hardcoded. Both binaries need the same string-to-`model.CommissionModel` parsing logic.

The prior `buildProvider` extraction decision (2026-04-22) established a rule: extract shared cmd-layer logic to `internal/cmdutil` when a third binary would require a third copy. `cmd/universe-sweep` will need the same `--commission` flag when TASK-0060 is completed.

The question: should each binary contain its own `parseCommissionModel` private function, or should it be extracted immediately?

## Options considered

### Option A: Duplicate in each binary for now, extract at the third copy
Consistent with the established three-copy threshold.
- **Pros**: No premature abstraction. Two copies is the stated acceptable threshold.
- **Cons**: TASK-0060 (universe-sweep) is already in the backlog. The third copy is imminent — extraction will happen within 1-2 sessions. Delaying adds churn for a trivially small function.

### Option B: Extract immediately to internal/cmdutil (chosen)
The function is six lines: a switch statement mapping strings to `model.CommissionModel` constants. The extraction cost is near-zero; the third caller is already scheduled.
- **Pros**: No churn when TASK-0060 lands. Help text documentation and error messages are consistent across all three binaries automatically. The `buildProvider` precedent explicitly names cmdutil as the home for shared cmd-layer plumbing.
- **Cons**: Technically extracts at two callers, not three. Minor departure from the stated rule.

## Decision

Extract immediately. `cmdutil.ParseCommissionModel(s string) (model.CommissionModel, error)` lives in `internal/cmdutil/cmdutil.go`. Both `cmd/backtest` and `cmd/sweep` call it. `cmd/universe-sweep` will call it when TASK-0060 is implemented.

The three-copy threshold is a guideline, not a hard rule. A near-zero extraction cost and an imminent third caller make extraction correct even at two callers here.

## Consequences

- `internal/cmdutil` now imports `pkg/model` in addition to the zerodha/cache packages it already imported. This is a legitimate dependency: cmdutil is the cmd-layer plumbing package, and commission model parsing is cmd-layer plumbing.
- Any future `cmd/` binary needing commission model parsing calls `cmdutil.ParseCommissionModel` — no new copies.
- Accepted values, error message wording, and behavior (case-sensitive, empty string rejected) are canonically defined in one place.

## Related decisions

- [buildProvider extracted to internal/cmdutil](../architecture/2026-04-22-buildprovider-extracted-to-cmdutil.md) — established the three-copy rule and cmdutil as home for shared cmd plumbing; this decision applies the same pattern one copy early
