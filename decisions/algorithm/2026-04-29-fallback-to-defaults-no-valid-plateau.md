# Fallback to strategy defaults for universe sweep when no valid plateau exists

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-29       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | sweep, plateau, universe-sweep, sensitivity-concern, parameter-selection, overfitting, TASK-0051 |

## Context

TASK-0051 selects a plateau-midpoint parameter for each strategy to use as input to the TASK-0052 universe sweep. For strategies where the valid region (TradeCount ≥ 30) is empty or all-negative Sharpe on RELIANCE, there is no statistically credible plateau to select from.

The question: what parameter should be passed to the universe sweep when no valid plateau exists — the "least bad" parameter from the valid region, strategy defaults, or something else?

## Options considered

### Option A: Least-bad parameter from the valid region
Select the parameter with the highest Sharpe in the valid region, even if that Sharpe is negative.
- **Pros**: Uses observed data rather than defaults.
- **Cons**: Selecting a parameter tuned to minimize losses on RELIANCE is a form of overfitting — the parameter is selected based on RELIANCE-specific results, which is exactly what the universe sweep is supposed to test against. Even if the Sharpe is negative, choosing the "least bad" parameter on one instrument biases the universe sweep. A strategy that loses least on RELIANCE may still lose on all other instruments.

### Option B: Strategy defaults (chosen)
Use the strategy's default parameters for the universe sweep when no valid plateau exists.
- **Pros**: Default parameters are pre-specified before any data is seen — they carry no RELIANCE-specific bias. The universe sweep tests the strategy as it would be deployed, not as it was tuned to fail least on one instrument.
- **Cons**: Default parameters may also produce sparse signals (< 30 trades) on RELIANCE. That is acceptable — the universe sweep's DSR correction and ≥40% instrument pass rate handle low-frequency strategies explicitly.

### Option C: Skip the strategy in the universe sweep
Exclude strategies with no valid plateau from TASK-0052.
- **Cons**: The universe sweep is the primary gate (per cross-instrument proliferation gate decision 2026-04-25). A strategy failing on RELIANCE does not mean it fails the universe gate. Excluding it preemptively would reintroduce the single-instrument gate that was explicitly superseded. Rejected.

## Decision

When `Report.SensitivityConcern` is set (valid region empty or all-negative Sharpe), the universe sweep receives the strategy's default parameters. The least-bad parameter is recorded in orientation data (e.g., the sweep CSV) but is not selected as the plateau-midpoint.

The orientation data note for these strategies reads: `"plateau_midpoint": null, "universe_sweep_params": "defaults", "sensitivity_concern": "<reason>"`.

## Consequences

- TASK-0052 (universe sweep) will run strategies with sensitivity concerns using their canonical default parameters. This is the correct baseline: if a strategy has any edge, defaults should capture it; if defaults fail, the strategy is killed cleanly without RELIANCE-specific tuning contaminating the result.
- The `plateau-params.json` output of TASK-0051 must include a `sensitivity_concern` field alongside the null `plateau_midpoint` so downstream tasks know why the default is being used.
- This does not pre-kill any strategy. A strategy flagged here can still pass the universe gate — the flag is purely an orientation note.

## Related decisions

- [Plateau procedure for trade-count-constrained strategies](./2026-04-29-plateau-procedure-trade-count-constrained.md) — defines when SensitivityConcern fires
- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate that processes the universe sweep results; a sensitivity concern here does not trigger a kill

## Revisit trigger

If a strategy with a sensitivity concern on RELIANCE consistently fails the universe gate across multiple instruments AND the failure pattern clearly correlates with low trade count (not edge absence), revisit whether the universe sweep should also apply a minimum-trades gate before recording a pass.
