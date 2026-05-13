---
name: DSR aggregation plan review pattern — Domain Logic always catches it
description: Plan reviews on DSR-ranking tools consistently catch wrong aggregation order; always verify avg(DSR) vs DSR(avg) before build
type: feedback
---

In TASK-0077 plan review round 1, Domain Logic Reviewer raised two blocking findings about DSR aggregation:

1. DSR(avgSharpe, totalTradeCount) is wrong: totalTradeCount inflates nObs across instruments; DSR(avg) ≠ avg(DSR).
2. InsufficientData filtering was missing from the plan.

**Why:** The correct formula (avg(DSR per instrument)) is non-obvious. It's easy to write DSR(avgSharpe) and have it feel correct while being methodologically inconsistent with ApplyUniverseGate.

**How to apply:** Whenever planning any cross-instrument metric aggregation involving DSR:
- Verify the aggregation is avg(DSR per instrument), not DSR(avg metric)
- Verify InsufficientData filtering is explicitly planned (instruments below MinTradesForMetrics or MinCurvePointsForMetrics excluded from aggregate)
- Cross-check against universesweep.ApplyUniverseGate as the reference implementation

This pattern will repeat for any future tool that computes per-variant quality metrics across a universe of instruments.
