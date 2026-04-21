# Kill-switch API keeps `internal/analytics` free of `internal/montecarlo`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | kill-switch, analytics, montecarlo, dependency-direction, package-boundary, TASK-0026 |

## Context

`DeriveKillSwitchThresholds` needs the bootstrap p5 Sharpe to build a `KillSwitchThresholds` struct. The natural call site has `montecarlo.BootstrapResult` in scope. The question was whether to accept a `BootstrapResult` directly or accept `sharpeP5 float64`.

`internal/analytics` currently imports only `pkg/model`, `math`, `sort`, and `time`. It is a pure, deterministic computation layer with no simulation dependency.

## Options considered

### Option A: Accept `montecarlo.BootstrapResult` (rejected)

```go
func DeriveKillSwitchThresholds(b montecarlo.BootstrapResult, inSample Report) KillSwitchThresholds
```

- **Pros**: Caller doesn't need to extract the field; one argument instead of two.
- **Cons**: Adds `internal/montecarlo` as a dependency of `internal/analytics`. The analytics package, which is a pure metrics layer, would now import a Monte Carlo simulation package. This violates the dependency direction established by `2026-04-20-internal-montecarlo-package-boundary` and pulls RNG machinery into a package that has none.

### Option B: Accept `sharpeP5 float64` (chosen)

```go
func DeriveKillSwitchThresholds(sharpeP5 float64, inSample Report) KillSwitchThresholds
```

- **Pros**: `internal/analytics` stays dependency-free of `internal/montecarlo`. Caller extracts the field explicitly — one line of code, zero import cost. Analytics remains a pure computation layer.
- **Cons**: Caller must know to pass `BootstrapResult.SharpeP5`, not the full struct. Godoc on the function must make this clear (it does).

## Decision

Accept `sharpeP5 float64`. The caller passes `montecarlo.BootstrapResult.SharpeP5` directly. `internal/analytics` imports nothing from `internal/montecarlo`.

The analytics package is a deterministic, pure-function layer. Adding a simulation dependency to it would blur the boundary between "what metrics does a strategy produce?" and "what would a random resampling say about those metrics?" Those are different questions and should stay in different packages.

## Consequences

Callers with a `BootstrapResult` in scope write `DeriveKillSwitchThresholds(result.SharpeP5, report)`. The extra field access is trivial. The import boundary is preserved: `output` imports `montecarlo` (already the case); `analytics` does not.

## Related decisions

- [internal/montecarlo as a standalone package](2026-04-20-internal-montecarlo-package-boundary.md) — established the package boundary this decision enforces from the analytics side.
- [Bootstrap Sharpe: non-annualized per-trade computation](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — defines what SharpeP5 means; TASK-0026 must use the same formula.

## Revisit trigger

If `internal/analytics` ever needs other fields from `BootstrapResult` (e.g., `WorstDrawdownP5` for a drawdown-based kill-switch threshold), reconsider whether to accept the full struct. At that point the import cost may be worth the ergonomic improvement.
