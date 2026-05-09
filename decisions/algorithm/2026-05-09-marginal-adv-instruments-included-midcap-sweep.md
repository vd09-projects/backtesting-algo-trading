# Marginal-ADV instruments included in midcap gate computation — not pre-excluded

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | midcap, ADV, liquidity, marginal-adv, universe-gate, BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM, SAIL, TASK-0091 |

## Context

Five instruments in `universes/nifty-midcap-liquid.yaml` were flagged as marginal-ADV (Rs 50–100 crore range, below the Rs 100 crore comfort threshold): NSE:BHEL, NSE:SCHAEFFLER, NSE:EXIDEIND, NSE:NATIONALUM, NSE:SAIL. Per the 2026-05-07 ADV floor decision, these were included in the universe but marked for scrutiny after sweep results — the decision was to flag-not-exclude, with replacement considered if `insufficient_data=true` appeared for a majority of them.

## Decision

**All five marginal-ADV instruments are retained in the eligible-for-walk-forward set.** No replacements needed for universe v1.

Sweep results for the five flagged instruments:

| Instrument | Raw Sharpe | Trades | DSR | Positive |
|---|---|---|---|---|
| NSE:SAIL | 0.4727 | 59 | 0.1759 | YES |
| NSE:BHEL | 0.3376 | 56 | 0.0328 | YES |
| NSE:EXIDEIND | 0.2923 | 55 | -0.0154 | YES |
| NSE:SCHAEFFLER | 0.2847 | 63 | -0.0024 | YES |
| NSE:NATIONALUM | 0.2136 | 58 | -0.0858 | YES |

All five: `insufficient_data=false`, trade_count >= 30 (55–63 trades), positive raw Sharpe. The 2026-05-07 replacement trigger condition (majority showing `insufficient_data=true`) was not met. The ADV floor decision's concern — that daily-bar fill quality could be materially wrong for these names — is not yet supported or refuted by sweep results; walk-forward will provide more evidence via fold-level performance variance.

The DSR correction brings EXIDEIND (-0.015), SCHAEFFLER (-0.002), and NATIONALUM (-0.086) below zero, but these are negative-DSR with positive raw Sharpe — they count toward the pass fraction (raw Sharpe > 0) even though their DSR contribution is negative.

## Consequences

- All five marginal-ADV instruments advance to walk-forward alongside the 38 non-marginal positive-Sharpe instruments.
- If walk-forward reveals high fold-level variance (large OOS IS ratio spread) for PSU names (BHEL, SAIL, NATIONALUM) specifically, this is consistent with the ADV concern — liquidity compression during stress periods shows up as regime instability in walk-forward, not in daily-bar universe sweep.
- EXIDEIND and SCHAEFFLER are private-sector names with more stable ADV; their marginal DSR is attributed to lower raw Sharpe, not liquidity risk.

## Related decisions

- [ADV floor for midcap daily-bar universe](./2026-05-07-adv-floor-midcap-daily-bar-universe.md) — the decision establishing the flag-not-exclude policy
- [Nifty Midcap 150 universe — instrument list](./2026-05-07-nifty-midcap-150-universe-instrument-list.md) — lists these instruments with marginal-ADV annotations
- [MACD crossover passes midcap universe gate](./2026-05-09-macd-crossover-midcap-universe-gate-passed.md) — the gate run providing this data

## Revisit trigger

If walk-forward shows that BHEL, SAIL, or NATIONALUM have majority-negative folds (NegFoldFlag) while private-sector names of similar raw Sharpe do not, revisit whether PSU liquidity compression is introducing a regime-instability signal not visible in daily-bar universe sweep.
