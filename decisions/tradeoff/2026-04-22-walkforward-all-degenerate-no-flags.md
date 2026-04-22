# All-degenerate walk-forward result: both flags false, DeduplicatedFoldCount=0

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | walkforward, degenerate, scoring, overfitting-gate, TASK-0022 |

## Context

When every fold in a walk-forward run has zero OOS trades (all folds degenerate), what should the Report flags say? `OverfitFlag=true` would be technically defensible (the strategy produced nothing), but semantically wrong.

## Decision

When all folds are degenerate, `OverfitFlag=false`, `NegativeFoldFlag=false`, and `DeduplicatedFoldCount=0`. "No trades produced" is not overfitting — it is a dead strategy, a different problem entirely. Flagging it as overfit conflates two distinct failure modes. Callers that want to detect "completely dead strategy" check `DeduplicatedFoldCount == 0`; that is the correct signal for "strategy never traded."

## Consequences

Callers must check `DeduplicatedFoldCount == 0` as a separate condition from the pass/fail flags. This is a small additional check but it is semantically correct. CLI output should surface this case explicitly ("0 scoreable folds — strategy produced no OOS trades").
