# EquityPoint defined in pkg/model, not internal/engine

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | convention       |
| Tags     | equity-curve, model, pkg/model, analytics, architecture, dependency-direction, EquityPoint |

## Context

When adding `EquityPoint` (a `{Timestamp time.Time; Value float64}` pair used to represent a single equity snapshot), the question was where to define the type. The natural candidate was `internal/engine` alongside `Portfolio`, which owns the slice. The alternative was `pkg/model`, alongside the other domain types (`Candle`, `Trade`, `Signal`).

## Options considered

### Option A: Define EquityPoint in internal/engine
- **Pros**: Co-located with `Portfolio` and `RecordEquity`, the producer of the data.
- **Cons**: Violates the project's dependency direction rule. `internal/analytics` and `internal/output` both need to consume equity curves (for Sharpe ratio, drawdown duration, printed reports). If `EquityPoint` lives in `internal/engine`, analytics and output must import engine — creating a dependency from analytics on engine. The architecture mandates that dependencies flow toward `pkg/model`, not toward internal packages.

### Option B: Define EquityPoint in pkg/model (chosen)
- **Pros**: Consistent with all other domain types (`Candle`, `Trade`, `Signal`, `Position`). Any package can import it — analytics, output, engine — without creating cross-package cycles. Follows the existing dep direction: engine depends on model, analytics depends on model, engine does not depend on analytics.
- **Cons**: Minor — the type is defined separately from its primary producer (`Portfolio`). Acceptable since all domain types live in model regardless of which layer produces them.

## Decision

`EquityPoint` is defined in `pkg/model/equity.go`. This is the only choice consistent with the project's dependency direction rule and existing pattern of placing all domain types in `pkg/model`.

## Consequences

- The dep direction invariant is preserved: `internal/analytics` can import `model.EquityPoint` directly without importing `internal/engine`.
- All future analytics metrics that consume the equity curve (Sharpe, Sortino, drawdown duration) import `pkg/model`, not `internal/engine`.
- New domain types that represent computed outputs (not just raw market data) belong in `pkg/model` by this same principle.

## Revisit trigger

If the project moves to a multi-instrument portfolio where equity snapshots require more complex structure (e.g., per-instrument breakdown), the `EquityPoint` type may need to be extended. At that point, consider whether the single `Value float64` field is still sufficient.

## Related decisions

- [MaxDrawdown computed from equity curve, not per-trade losses](../algorithm/2026-04-07-max-drawdown-from-equity-curve.md) — established that the equity curve is the canonical data source for drawdown; this decision ensures the type is accessible to analytics without a dep cycle.
- [Value semantics for domain model types (Candle, Config)](../convention/2026-04-06-value-semantics-for-domain-types.md) — `EquityPoint` follows the same convention: value type, passed by value.
