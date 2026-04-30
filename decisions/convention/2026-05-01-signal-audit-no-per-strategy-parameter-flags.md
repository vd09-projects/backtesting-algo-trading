# cmd/signal-audit hardcodes strategy defaults — no per-strategy parameter override flags

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | convention       |
| Tags     | signal-audit, cmd, flags, default-params, audit-purpose, TASK-0050 |

## Context

`cmd/signal-audit` runs all 6 strategies at their default parameters to count trades per instrument. The question was whether to add per-strategy parameter flags (e.g. `--rsi-period`, `--fast-period`) so operators could audit at non-default parameters.

## Decision

`cmd/signal-audit` accepts no strategy-specific parameter flags. It hardcodes the same defaults as `cmd/universe-sweep` inside the factory closure. The audit's single purpose is "how many trades do **default** parameters produce?" Adding override flags would imply it is a general parameter explorer, which it is not — `cmd/universe-sweep` and `cmd/sweep` already serve that purpose.

## Consequences

- To audit at non-default parameters, use `cmd/universe-sweep` directly with explicit flags.
- The binary stays simple: no 13-flag proliferation. One responsibility, one set of defaults.
- If the evaluation methodology ever moves to auditing at plateau-midpoint parameters (post-TASK-0051), a new binary or a flag addition would be the right path — not retrofitting this one.

## Revisit trigger

If the signal audit is ever run at non-default parameters as a standard step in the evaluation pipeline, add an override flag at that point.
