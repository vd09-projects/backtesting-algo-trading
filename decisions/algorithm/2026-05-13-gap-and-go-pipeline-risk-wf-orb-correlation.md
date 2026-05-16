# Gap-and-Go pipeline risks: walk-forward OverfitFlag (2022) and ORB/Gap-and-Go correlation gate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, walk-forward, correlation-gate, ORB, TASK-0075 |

## Context

Marcus identified two structural risks for Gap-and-Go during the 2026-05-13 evaluation session. Both are post-implementation pipeline concerns — they do not affect the GO verdict but must be tracked to inform evaluation monitoring and portfolio construction decisions.

## Decision

**Risk 1 (primary): Walk-forward OverfitFlag in 2022 choppy regime.**

The 2022 market regime (rate hike uncertainty, FII selling, range-bound NSE) is materially different from the 2021 trending regime. Gap events in 2022 were more likely to reverse intraday — stocks would gap up on overnight news and reverse within the first session as macro uncertainty dominated single-stock flow. If IS Sharpe is dominated by 2021 performance, OOS Sharpe in 2022 will degrade, triggering OverfitFlag (OOSSharpe < 50% of ISSharpe) at walk-forward.

Monitoring: during TASK-0122 (walk-forward), pay particular attention to fold cohorts with OOS windows covering 2022. If OverfitFlag clusters on Consumer/FMCG names (GODREJCP, COLPAL, DABUR, EMAMILTD, TATACONSUM) — which are more macro-sensitive in gap behavior — that is informative for universe construction but does not change the gate application. The gate applies uniformly.

The 60% instrument retention floor (floor(0.60 × universe_gate_passes), minimum 6) is the quantitative test. If Gap-and-Go clears 60% retention despite 2022 pressure, the edge is sufficiently robust.

**Risk 2 (secondary): ORB/Gap-and-Go mutual correlation gate if both pass bootstrap.**

ORB and Gap-and-Go are structurally similar: both are 5-min CNC strategies, both are long-only, both enter on morning opening behavior (within the first 30-60 minutes of the 09:15 IST session), and both hold for 2-3 sessions. They are the highest-risk pair in the portfolio from a correlation standpoint.

In trending markets, both strategies will fire frequently and in the same direction (gap-up on ORB-confirming breakout, gap-up on Gap-and-Go catalyst). In crash regimes (2020 COVID, 2022 rate shock), both will take losses simultaneously — long-only exposure to the same morning session dynamics. Stress-period correlation (COVID crash 2020, rate correction 2022) may exceed 0.6, breaching the correlation gate threshold.

Resolution path per `decisions/algorithm/2026-04-27-correlation-gate.md`: if both pass bootstrap and fail the mutual correlation gate (stress-period r >= 0.6), retain the one with higher DSR-corrected Sharpe from the universe sweep. The dropped strategy is recorded as "excluded (correlation)" — not a gate failure. TASK-0124 handles this explicitly.

## Consequences

- TASK-0122 (walk-forward) must track which fold OOS windows cover 2022 and whether OverfitFlag failures cluster in that period. This is informative context, not a rescue mechanism — the gate still applies.
- If ORB and Gap-and-Go both pass bootstrap, TASK-0124 must compute mutual correlation before portfolio construction. This gate may eliminate one of them even if both have strong individual bootstrap results.
- MACD/Gap-and-Go correlation (MACD is daily-bar trend-following, Gap-and-Go is intraday event-driven) is the lower-risk pair. Both are long-only, which creates some co-movement in crash periods, but mechanistic divergence is high. Stress-period r is likely in 0.3-0.5 range, below the 0.6 threshold.

## Revisit trigger

If ORB is killed before reaching bootstrap (universe gate or walk-forward gate failure), the ORB/Gap-and-Go correlation risk is moot. The correlation check in TASK-0124 would then only cover Gap-and-Go vs MACD midcap portfolio.

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document including pipeline risk section
- [Correlation gate thresholds](../algorithm/2026-04-27-correlation-gate.md) — the gate applied at TASK-0124
- [Walk-forward instrument-count gate relaxed to 60%](../algorithm/2026-05-05-walk-forward-instrument-count-gate-relaxed.md) — the retention floor at TASK-0122
- [ORB Marcus Rules](./2026-05-13-orb-marcus-rules.md) — ORB evaluation; the correlation counterpart to Gap-and-Go
