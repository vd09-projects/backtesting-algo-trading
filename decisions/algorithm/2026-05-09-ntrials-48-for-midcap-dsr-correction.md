# nTrials=48 for DSR correction in Nifty Midcap 150 universe sweep

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | DSR, nTrials, multiple-testing, midcap, nifty-midcap-150, universe-gate, TASK-0091 |

## Context

The Deflated Sharpe Ratio (DSR) corrects an observed Sharpe for the expected maximum Sharpe from testing nTrials independent strategies/instruments. The nTrials parameter controls the severity of the correction: higher nTrials → higher E[max SR] → lower (more conservative) DSR.

For the large-cap universe sweep (TASK-0052), nTrials=15 was implicit — the universe had 15 instruments. The midcap universe (TASK-0091) has 48 instruments. The question is whether nTrials should equal the number of instruments tested or some other count.

## Decision

**nTrials=48 for the Nifty Midcap 150 universe sweep DSR computation**, equal to the number of instruments in the universe.

This is the same convention used for the large-cap sweep (nTrials=15). We tested 48 instruments; we correct for 48 independent comparisons. The DSR formula from analytics.DSR(sharpe, nTrials, tradeCount) uses nTrials as the multiple-testing count.

**Numerical impact:**

- E[max SR] at nTrials=15: 2.123 (Φ⁻¹(14/15)·(1−γ) + Φ⁻¹(14/(15·e))·γ)
- E[max SR] at nTrials=48: 2.261 (Φ⁻¹(47/48)·(1−γ) + Φ⁻¹(47/(48·e))·γ)
- Difference: +0.138 in E[max SR], which compresses DSR by approximately 0.138 × SE per instrument

This is why MACD's DSR avg dropped from 0.2715 (large-cap, nTrials=15) to 0.0885 (midcap, nTrials=48) even though the raw Sharpe distribution on midcaps is strong — the penalty for searching over 48 instruments is substantially heavier.

## Consequences

- The midcap gate is a harder bar in DSR terms than the large-cap gate, even with the same DSR > 0 threshold, because nTrials=48 applies a larger correction.
- A strategy must have meaningfully stronger per-instrument Sharpe to achieve the same DSRAvg on a 48-instrument universe vs a 15-instrument universe.
- This is methodologically correct: we did run 48 tests. Reducing nTrials to 15 (to match the large-cap convention directly) would underestimate the multiple-testing penalty and is not defensible.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — gate using DSR correction
- [MACD crossover passes midcap universe gate](./2026-05-09-macd-crossover-midcap-universe-gate-passed.md) — the gate run this decision covers

## Revisit trigger

If the midcap universe is expanded or contracted significantly (e.g., to 100 instruments or trimmed to 25), revisit whether nTrials should equal the full universe size or only the instruments actually tested in a given run.
