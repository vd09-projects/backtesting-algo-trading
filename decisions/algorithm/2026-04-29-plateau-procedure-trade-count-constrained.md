# Plateau procedure for trade-count-constrained strategies

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-29       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | sweep, plateau, trade-count, valid-region, sensitivity-concern, parameter-sensitivity, TASK-0051 |

## Context

TASK-0051 runs a 1D parameter sweep for each of six strategies on NSE:RELIANCE 2018–2024 to identify the plateau-midpoint parameter for use in the universe sweep. Five of the six strategies (SMA, RSI, Donchian, Bollinger, Momentum) were excluded by the TASK-0050 signal frequency audit — they produce fewer than 30 trades per instrument at default parameters.

The existing `computePlateau` function in `internal/sweep` applied the 80% Sharpe floor against the global peak Sharpe across all parameter values in the sweep range. For thin-signal strategies, the global peak often comes from a parameter value that generates fewer than 30 trades — making the peak Sharpe statistically meaningless. Applying the floor against this peak would misrepresent what "80% of peak" means.

The question: when a strategy has sparse signal counts, should the plateau analysis use the global peak or only parameters that generate a minimum number of trades?

## Options considered

### Option A: Global peak (existing behaviour)
- **Pros**: Simple. Consistent with existing tests.
- **Cons**: A parameter generating 5 trades with Sharpe 2.0 sets the bar against which parameters generating 40 trades at Sharpe 1.0 are measured. The 1.0 appears to be 50% of peak and fails the 80% floor — even though it's the only statistically credible point. This produces misleading plateau output for thin-signal strategies.

### Option B: Valid-region peak (chosen)
Apply the 80% floor against the peak Sharpe within the "valid region" — the subset of parameters where TradeCount ≥ 30.
- **Pros**: Floor is anchored to a statistically credible baseline. A parameter with 40 trades at Sharpe 1.0 is correctly evaluated against other ≥30-trade parameters, not against a noisy 5-trade peak. Conceptually correct: "robust parameter region" requires both performance AND minimum sample size.
- **Cons**: Requires filtering results before finding the peak. Slightly more complex logic. Existing tests needed updating to pass `minTrades=0` (backward-compatible) or `minTrades=30` (new behaviour).

### Option C: Adjust the 80% threshold for thin-signal strategies
Reduce the floor (e.g., to 70%) when strategies have low signal counts.
- **Cons**: Rejected. A noisier Sharpe estimate requires a stricter plateau definition, not a looser one. Reducing the threshold would widen the apparent stable region precisely when the data is least trustworthy.

## Decision

Apply the 80% Sharpe floor against the peak within the valid region (TradeCount ≥ 30), not the global peak. The threshold constant (80%) stays unchanged.

Implemented in `internal/sweep.computePlateauWithMinTrades(results []Result, minTrades int)`. When `minTrades=0` the function behaves identically to the old `computePlateau` (backward compatible). Production callers pass `MinTradesForPlateau = 30`.

If the valid region is empty (no parameter achieves ≥ 30 trades in the sweep range): `Report.SensitivityConcern` is set to `"no parameter achieves >= 30 trades in sweep range"` and `Report.Plateau` is nil.

If the valid region is non-empty but all peak Sharpes are non-positive: `SensitivityConcern` is set to `"no viable parameter region: valid-region peak Sharpe is non-positive"` and `Report.Plateau` is nil.

## Consequences

- `internal/sweep.Report` gains a `SensitivityConcern string` field. Callers that previously checked only `report.Plateau == nil` should also check `report.SensitivityConcern` to distinguish "no data" from "valid trades exist but no positive Sharpe region".
- The plateau-midpoint parameter selected for the universe sweep is always drawn from the valid region. If the valid region is empty or all-negative, strategy defaults are used for the universe sweep (see companion decision on fallback behaviour).
- `MinTradesForPlateau = 30` is exported from `internal/sweep` so callers can reference the threshold explicitly rather than hardcoding 30.

## Related decisions

- [Fallback to defaults for universe sweep when no valid plateau exists](./2026-04-29-fallback-to-defaults-no-valid-plateau.md) — companion decision specifying what to use when SensitivityConcern fires
- [Sweep plateau uses non-contiguous min/max range](../architecture/2026-04-15-sweep-plateau-non-contiguous-min-max-range.md) — prior decision on how qualifying entries define the plateau range; still applies within the valid region

## Revisit trigger

If `MinTradesForPlateau = 30` causes more than half of the universe instruments to always fall into the invalid region for a given strategy, revisit the threshold. The universe sweep gate already handles insufficient-trade instruments via the DSR correction — a lower threshold here (e.g., 20) may be appropriate if 30 proves unachievable across most of the universe.
