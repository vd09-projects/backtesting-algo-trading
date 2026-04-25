# 2025 live trading is the true holdout; no historical data reserved

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-25       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | holdout, walk-forward, OOS, 2025-live, live-validation, evaluation-methodology |

## Context

Before running the cross-instrument evaluation pipeline, the question arose of whether to reserve a portion of the 2018-2024 historical data as a "holdout" set — data never touched during parameter selection or walk-forward — to be used as a final OOS test before live allocation. The evaluation pipeline uses walk-forward (2yr IS / 1yr OOS / 1yr step) over the full 2018-2024 window, which means all available historical data is consumed in rolling folds.

## Options considered

### Option A: Declare 2024 as holdout (restrict walk-forward to 2018-2023)
- **Pros**: Provides a clean final OOS year for evaluation before going live.
- **Cons**: Requires restricting the walk-forward range, losing one year of OOS evidence. More importantly, if any parameter selection or strategy design decisions were made with knowledge of 2024 market conditions (which is hard to avoid entirely), the 2024 "holdout" is partially contaminated anyway. False sense of a clean split.

### Option B: No historical holdout; 2025 live trading is the holdout (selected)
- **Pros**: Intellectually honest — once walk-forward has consumed 2018-2024, there is no truly pristine historical data remaining. The walk-forward OOS windows already provide empirical out-of-sample evidence across multiple market regimes. The first genuinely unseen data is 2025 live data.
- **Cons**: The "holdout test" is real capital at risk, not a simulation. Mitigated by: (a) starting at small ₹3 lakh experimental size, (b) weekly kill-switch monitoring, (c) pre-committed halt conditions.

### Option C: Use CPCV to generate many train-test paths, effectively multiplying OOS evidence
- **Pros**: More statistical power per unit of historical data.
- **Cons**: Not implemented; would require significant infrastructure work. Deferred to a future phase if the walk-forward evidence is insufficient.

## Decision

No historical data is reserved as a holdout. The **walk-forward OOS windows** (covering 2020-2024 across 4-5 rolling folds) are treated as the empirical out-of-sample evidence. The **2025 live trading period** at ₹3 lakh experimental capital is the true holdout test.

This is the intellectually honest position: any historical data the researchers have seen during strategy design, parameter selection, or evaluation methodology decisions is not truly "unseen." Calling 2024 a holdout when the strategies were designed with awareness of the 2018-2024 market landscape (including 2024 as the most recent year) is not a meaningful separation. The only genuinely new data is future data.

Consequence: the pre-live brief (TASK-0056) must document kill-switch thresholds and obtain sign-off before the first trade, because 2025 live performance is the validation test, not a safety net.

## Consequences

- Walk-forward runs over the full `2018-01-01` to `2024-12-31` window per the existing 2026-04-22 window-sizing decision.
- No separate holdout evaluation step in the pipeline.
- Weekly kill-switch monitoring (TASK-0048) becomes the primary mechanism for detecting live performance degradation.
- If 2025 live results diverge materially from the walk-forward OOS Sharpe distribution, that is the signal to halt and re-evaluate.

## Related decisions

- [Baseline backtest period: 2018-2024](./2026-04-15-baseline-backtest-period-2018-2024.md) — sets the outer window the walk-forward operates within
- [Walk-forward window sizing defaults (2yr IS / 1yr OOS)](./2026-04-22-walk-forward-window-sizing-default.md) — the window config that consumes the 2018-2024 range

## Revisit trigger

If 2025 live results diverge materially from the walk-forward OOS Sharpe distribution (e.g., live Sharpe is negative when all OOS folds were positive), this decision should be revisited — specifically, whether a more conservative historical holdout split would have predicted the divergence.
