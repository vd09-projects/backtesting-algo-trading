# param-search DSR aggregation: mean(DSR per instrument) matching ApplyUniverseGate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | DSR, aggregation, methodology, universesweep, param-search, TASK-0077 |

## Context

`internal/paramsearch` needed to rank parameter variants by DSR-corrected Sharpe. The initial plan specified `analytics.DSR(avgSharpe, gridSize, totalTradeCount)` — computing DSR once on the cross-instrument average. The plan review (Domain Logic Reviewer) identified two correctness problems with this approach and required alignment with the existing `universesweep.ApplyUniverseGate` methodology.

## Options considered

### Option A: DSR(avgSharpe, gridSize, totalTradeCount) — rejected
- **Pros**: Simpler — one DSR call per variant.
- **Cons**: Two correctness problems. (1) `totalTradeCount` as `nObservations` inflates nObs: if 5 instruments each produce 40 trades, totalTradeCount=200, but that does not represent 200 independent observations of the same strategy. (2) `DSR(avg) ≠ avg(DSR)` — the aggregation order matters for the ranking when SE varies across instruments. Rankings produced by this method are inconsistent with how `ApplyUniverseGate` computes DSR.

### Option B: mean(DSR per sufficient instrument) — chosen
For each variant, compute `analytics.DSR(instrumentSharpe, float64(gridSize), float64(instrumentTradeCount))` per sufficient instrument, then average across sufficient instruments.
- **Pros**: Matches `universesweep.ApplyUniverseGate` exactly. Each DSR call uses per-instrument trade count as `nObservations`, which correctly represents the statistical base for that instrument's Sharpe estimate. Ranking produced by param-search is directly comparable to the universe gate ranking.
- **Cons**: One DSR call per (variant × instrument) instead of one per variant. Negligible cost given the engine run dominates.

## Decision

Option B: `mean(analytics.DSR(instrumentSharpe, float64(gridSize), float64(instrumentTradeCount)))` across sufficient instruments per variant. This matches `universesweep.ApplyUniverseGate` exactly and keeps the DSR methodology consistent across all evaluation tools in the pipeline.

## Consequences

- The `gridSize` used in every DSR call is the full Cartesian product count, computed once before the outer variant loop and passed as a constant to `runVariant`. This is correct — nTrials represents the total search space, not the number of sufficient variants.
- Instruments with `TradeMetricsInsufficient || CurveMetricsInsufficient` are excluded from each variant's aggregate. See companion decision on InsufficientData filtering.
- The DSR convention in param-search now joins three established conventions: sweep2d (uses equity curve length as nObs), universesweep (uses per-instrument trade count as nObs), and param-search (uses per-instrument trade count as nObs, matching universesweep). The sweep2d divergence is a known inconsistency tracked in patterns.md.

## Related decisions

- [ApplyUniverseGate DSR computation](../../algorithm/2026-05-03-universe-gate-dsr-correction.md) — the existing precedent param-search matches
- [nTrials=48 for midcap DSR correction](../../algorithm/2026-05-09-ntrials-48-for-midcap-dsr-correction.md) — context on nTrials semantics in the broader pipeline

## Revisit trigger

If `internal/sweep2d` is updated to use per-instrument trade count as nObs (aligning with universesweep), revisit whether to backfill the same change in param-search. Currently param-search matches universesweep; sweep2d is the outlier.
