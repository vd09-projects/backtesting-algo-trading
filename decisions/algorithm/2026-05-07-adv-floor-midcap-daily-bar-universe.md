# ADV floor for midcap daily-bar universe — Rs 50 crore minimum, Rs 100 crore comfort threshold

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | [midcap, liquidity, ADV, instrument-screening, universe, slippage, TASK-0072] |

## Context

Midcap stocks have thinner order books than Nifty50 large-caps. A minimum ADV threshold is needed to ensure that daily-bar strategy fills are clean and that the zero-slippage or fixed-slippage assumption in backtesting holds to a reasonable approximation. The Nifty50 large-cap universe had no explicit ADV screen (all large-caps comfortably exceed any reasonable threshold); the midcap universe required one.

## Options considered

### Option A: Rs 100 crore ADV minimum (hard floor)
- **Pros**: Strong liquidity guarantee; eliminates PSU names with seasonally compressed liquidity.
- **Cons**: Significantly narrows the midcap universe; excludes names like BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM which may have genuine edge but occasionally trade below Rs 100 crore.

### Option B: Rs 50 crore ADV minimum with Rs 100 crore comfort flag (chosen)
- **Pros**: Wider universe; Rs 50 crore is sufficient for clean fills at modest position sizes (< 0.5% of ADV). Marginal names are included but flagged for closer review after sweep results.
- **Cons**: Includes names whose liquidity can compress during market stress — fills in stress periods may be worse than modelled.

### Option C: No explicit ADV screen (let the strategy's trade count gate handle it)
- **Pros**: No manual judgment required.
- **Cons**: A strategy might technically generate 30+ trades on a thin instrument while the fill model is materially wrong. ADV screen is a data-quality gate, not a strategy-quality gate.

## Decision

Rs 50 crore ADV as the hard minimum floor for inclusion. Instruments in the Rs 50–100 crore range are flagged with an inline YAML comment in the universe file — they are not pre-excluded, but are marked for scrutiny after sweep results. The Rs 100 crore level is a comfort threshold, not a gate.

Reasoning: at typical position sizes for this backtest (10% of Rs 1 lakh = Rs 10,000 per trade), Rs 50 crore ADV means the position is ≤ 0.02% of daily volume — well inside the range where market impact is negligible on daily bars. Slippage from order-book thinness is a larger concern at intraday bar frequencies, not daily.

## Consequences

- BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM, SAIL are included with marginal-ADV flags. If they consistently show `insufficient_data=true` or negative Sharpe in the sweep, they are the first replacement candidates.
- PSU names (BHEL, SAIL, NATIONALUM) are specifically noted as prone to liquidity compression — worth monitoring across market regimes.
- If the project moves to intraday (5-min) bars on this universe, the ADV floor should be revisited upward (Rs 150–200 crore minimum for 5-min fills).

## Related decisions

- [Nifty Midcap 150 universe — instrument list and gate thresholds](./2026-05-07-nifty-midcap-150-universe-instrument-list.md) — the instrument list this threshold screens

## Revisit trigger

If intraday bar frequencies (5-min or 15-min) are used on the midcap universe, raise the ADV floor to Rs 150 crore minimum. Also revisit if any flagged instrument shows consistently poor fill quality in live execution.
