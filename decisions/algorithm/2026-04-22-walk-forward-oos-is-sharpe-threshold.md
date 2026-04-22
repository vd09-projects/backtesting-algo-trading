# Walk-forward OOS/IS Sharpe threshold and fold-level flagging

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | walk-forward, IS-OOS, sharpe, threshold, overfitting-gate, TASK-0022 |

## Context

The walk-forward framework needs pass/fail rules. Two questions: (1) what aggregate OOS/IS Sharpe ratio constitutes a "likely overfit" flag, and (2) should there be fold-level secondary checks beyond the aggregate?

## Decision

**Aggregate pass rule**: avg OOS Sharpe >= 50% of avg IS Sharpe. Below this threshold, `OverfitFlag = true`. For stateless strategies this threshold is permissive — without fitting, IS and OOS Sharpe should converge, so a strategy showing only 51% OOS/IS retention still has unexplained degradation worth investigating. The 50% floor is not a green light; it is the minimum floor below which the result is clearly concerning.

**Secondary fold-level rule**: Flag any fold with negative OOS Sharpe. If 2+ non-degenerate folds have negative OOS Sharpe, `NegativeFoldFlag = true` regardless of the aggregate ratio. One negative fold is uncomfortable; two or more is a kill signal. A single negative OOS fold is documented but does not trigger the flag alone.

**Degenerate windows**: Folds with zero OOS trades are marked `Degenerate=true` and excluded from all scoring. They do not contribute to averages or negative fold counts. Zero trades is not a pass or a fail — it is no data.

**Sharpe consistency**: Per-trade non-annualized Sharpe throughout — `mean(ReturnOnNotional) / std(ReturnOnNotional)` with sample variance (n-1 denominator). Consistent with the bootstrap standing order (2026-04-20-bootstrap-sharpe-non-annualized-per-trade). Do not switch to annualized Sharpe for walk-forward OOS windows just because a window spans a year.

## Consequences

The 50% threshold may need tightening if a future strategy has high enough trade frequency that fold-level Sharpe becomes reliable. The fold-level negative-OOS check is the more actionable signal for low-frequency strategies where per-fold aggregate is noisy.

## Revisit trigger

If strategy trade frequency exceeds 50 trades/year, fold-level Sharpe becomes reliable enough to tighten the aggregate threshold above 50%.
