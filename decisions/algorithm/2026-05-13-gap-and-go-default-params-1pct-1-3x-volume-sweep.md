# Gap-and-Go default parameters: gapThreshold=1.0%, volumeMultiplier=1.3×; sweep axes defined

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, volume-threshold, gap-threshold, signal-audit, TASK-0075 |

## Context

Gap-and-Go requires two primary filter parameters: the minimum gap size to qualify as a gap event, and the minimum volume required to confirm institutional participation. Both are tested at the 09:15 bar. Setting these too aggressively kills signal frequency (fewer than 30 trades per instrument fails the signal audit gate). Setting them too loosely admits low-quality gap events where the price action is random noise rather than institutional follow-through. Marcus defined the defaults and sweep axes during the 2026-05-13 evaluation session.

## Decision

**Default parameters:**
- `gapThreshold = 1.0%`: minimum gap size (`(Open09:15 - PreviousSessionClose) / PreviousSessionClose >= 0.01`)
- `volumeMultiplier = 1.3×`: minimum volume at the 09:15 bar relative to the 20-day average daily volume

**Parameter sweep axes** (run on NSE:INDHOTEL as orientation instrument):

| Axis | Values | Default |
|---|---|---|
| Gap threshold | 0.75%, 1.0%, 1.5% | 1.0% |
| Volume multiplier | 1.0×, 1.3×, 1.5× | 1.3× |
| SL multiplier | 0.5×, 0.8×, 1.0× gapPct | 0.8× |
| Hold bars | 75, 150, 225 | 150 |

Total grid: 81 variants (3×3×3×3). Plateau procedure applied: 80% Sharpe floor within valid region (>= 30 trades) per `decisions/algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md`.

**Signal frequency expectations at 1.0% threshold:** 8-12% of NSE midcap trading days qualify. At ~250 trading days/year, this gives 20-30 qualifying events per year per instrument, 50-75 over the 2021-2023 evaluation window. This is above the 30-trade signal audit floor but the margin is thin.

**Signal audit risk flag:** if more than 20% of the 48 midcap instruments produce fewer than 30 trades at default parameters (1.0% gap, 1.3× volume), re-run the signal audit at 0.75% gap threshold before declaring a kill. The lower threshold may be necessary for lower-volatility midcap names in sectors like Consumer/FMCG.

## Consequences

- Volume tracking requires a 20-day rolling average of daily volume. This must be pre-allocated at strategy construction (ring buffer pattern) to avoid hot-loop allocation.
- The signal audit gate (TASK-0119) uses default parameters. If signal audit passes, the sweep then identifies the plateau-midpoint for universe sweep.
- If the sweep produces no valid plateau at any combination of these axes, fallback-to-defaults applies per `decisions/algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md`.

## Revisit trigger

If more than 20% of the 48 midcap instruments are EXCLUDED at signal audit with 1.0% gap threshold, test 0.75% threshold before kill decision. If 0.75% also fails the signal audit, the strategy cannot achieve adequate signal frequency on the midcap universe and should be killed.

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document
- [Plateau procedure for trade-count-constrained strategies](../algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md) — governs plateau selection from sweep
- [Fallback to defaults when no valid plateau](../algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md) — fallback if sweep has no valid region
