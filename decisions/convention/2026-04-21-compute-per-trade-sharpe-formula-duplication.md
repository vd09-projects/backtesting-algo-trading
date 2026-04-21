# `computePerTradeSharpe` intentionally duplicates `montecarlo.sampleSharpe`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-21       |
| Status   | experimental     |
| Category | convention       |
| Tags     | kill-switch, analytics, montecarlo, formula-duplication, dependency-direction, TASK-0026 |

## Context

`CheckKillSwitch` needs to compute per-trade Sharpe from a rolling window of trades. `internal/montecarlo` already has `sampleSharpe([]float64) float64` that does exactly this — `mean(r)/std(r)`, sample variance (n-1), no annualization. The TASK-0024 algorithm decision requires that the kill-switch live computation and the bootstrap computation use identical formulas; otherwise the threshold is measured in different units than the live metric.

The question was how to share the formula across the two packages.

## Options considered

### Option A: Export `sampleSharpe` from `internal/montecarlo` (rejected)

Make `sampleSharpe` exported (`SampleSharpe`) and have `internal/analytics` import it.

- **Pros**: Single implementation, DRY.
- **Cons**: Exports a helper that is only meaningful in context — callers outside the package have no obvious use for it. More critically, it requires `internal/analytics` to import `internal/montecarlo`, which the architecture decision (`2026-04-21-kill-switch-analytics-to-montecarlo-boundary`) explicitly prohibits. One formula is not worth a cross-layer import.

### Option B: Move the formula to `pkg/model` or a shared utility (rejected)

Extract the formula into a shared package that both `internal/analytics` and `internal/montecarlo` import.

- **Pros**: Single implementation.
- **Cons**: A three-line statistical helper does not justify a shared package. `pkg/model` is for domain types, not math helpers. Creating a `pkg/stats` or `internal/mathutil` package for one function is premature abstraction that violates the "five gates" principle (dumber version would work).

### Option C: Duplicate the formula (chosen)

Implement `computePerTradeSharpe([]float64) float64` in `internal/analytics/killswitch.go` with identical code to `montecarlo.sampleSharpe`.

- **Pros**: No new imports, no new packages, no exported helpers. Both packages remain self-contained.
- **Cons**: Two implementations of the same three-line formula. If the formula ever changes (e.g., a convention decision to use population variance), both must be updated.

## Decision

Duplicate the formula. `computePerTradeSharpe` in `internal/analytics` uses the identical `mean(r)/std(r)` with `n-1` sample variance as `montecarlo.sampleSharpe`. An inline comment in both functions cross-references the other and cites the TASK-0024 algorithm decision as the authority on what "per-trade Sharpe" means.

The duplication is load-bearing: the formula identity is the correctness guarantee that the kill-switch threshold and the live metric are comparable. If the two ever diverge, the kill-switch becomes meaningless. The cross-references in code make this visible; the algorithm decision in `decisions/` makes the constraint durable.

Three lines of duplicated math is not a maintenance burden. A shared package for three lines is.

## Consequences

If the per-trade Sharpe formula changes (e.g., switching to population variance), both `montecarlo.sampleSharpe` and `analytics.computePerTradeSharpe` must be updated in the same commit. The algorithm decision file for TASK-0024 must also be updated. The cross-references in comments make this a discoverable change, not a silent one.

## Related decisions

- [Bootstrap Sharpe: non-annualized per-trade computation](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — the authority on the formula; both implementations must match it.
- [Kill-switch API keeps analytics free of montecarlo](2026-04-21-kill-switch-analytics-to-montecarlo-boundary.md) — the import constraint that made option A unavailable.
- [internal/montecarlo as a standalone package](../architecture/2026-04-20-internal-montecarlo-package-boundary.md) — the package boundary these decisions collectively enforce.

## Revisit trigger

If a third package needs the same per-trade Sharpe formula, revisit whether a shared package is justified. One duplication is acceptable; three copies are a smell.
