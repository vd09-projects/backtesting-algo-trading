# Insufficient instruments excluded per variant in param-search; all-insufficient variant flagged

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | InsufficientData, aggregation, filtering, param-search, universesweep, TASK-0077 |

## Context

`internal/paramsearch` runs each parameter variant across a universe of instruments. At extreme parameter values — e.g., a very long SMA period, a very tight CCI threshold — some instruments generate too few trades or too short an equity curve to produce meaningful Sharpe estimates. Without filtering, these instruments contribute zero Sharpe to the variant's aggregate, dragging the average down and corrupting the variant ranking. A variant that generates sufficient trades on 10 instruments but has 2 insufficient instruments could score lower than a variant with fewer trades on all 12 instruments due to the zero-Sharpe drag.

## Decision

Mirrors `universesweep.runInstrument`: instruments where `analytics.Report.TradeMetricsInsufficient || CurveMetricsInsufficient` are excluded from per-variant Sharpe and DSR aggregation. `VariantResult.InsufficientData=true` when zero sufficient instruments exist for a variant — these variants are sorted last in the output and present in the CSV with `insufficient_data=true` so the caller can see the full grid rather than a silently truncated view.

## Consequences

- Per-variant `RawSharpe` and `DSRSharpe` are means over sufficient instruments only. `VariantResult.TradeCount` is the sum of trades from sufficient instruments only.
- `VariantResult.InsufficientData=true` does not mean the variant was not run — it means all instruments it was run against produced insufficient metrics. The CSV row is present so the caller knows the search space was exhausted.
- Variants with some but not all insufficient instruments produce valid DSR estimates from the sufficient subset. The quality of the estimate degrades as the sufficient subset shrinks, but the flag for this is the raw trade count and DSR value itself, not a separate indicator.

## Related decisions

- [universesweep runInstrument InsufficientData handling](../../convention/2026-05-11-config-newstrategy-as-factory-instead-of-shared.md) — the established precedent this mirrors
- [param-search DSR aggregation matches ApplyUniverseGate](./2026-05-13-param-search-dsr-aggregation-matches-applyuniverse.md) — companion decision on how the aggregate is computed once filtering is applied
