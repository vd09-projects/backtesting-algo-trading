# Walk-forward purpose for stateless fixed-parameter strategies

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | walk-forward, stateless, regime-stability, overfitting-defense, TASK-0022 |

## Context

TASK-0022 builds a walk-forward validation framework. The question was whether walk-forward provides a meaningful overfitting defense for rule-based strategies with fixed parameters (e.g., SMA crossover with fixed lookback, RSI mean-reversion with fixed thresholds). No parameter fitting occurs — the strategy is stateless, taking a candle slice and returning a signal.

## Decision

Walk-forward on stateless fixed-parameter strategies is a regime-stability test, not a parameter-overfitting test. Since no fitting occurs, the IS/OOS Sharpe comparison does not measure parameter inflation. Instead, the fold-level breakdown reveals whether the strategy's reported aggregate Sharpe is concentrated in a specific regime (e.g., the 2020 V-recovery) or distributed across regimes. A strategy whose IS folds happen to cover a favorable regime and whose OOS folds cover an unfavorable one will show IS/OOS degradation without having overfit anything — which is the correct finding: the edge is regime-sensitive, not robust. The framework is still worth running for exactly this reason.

## Consequences

Callers and results-consumers must not interpret a walk-forward pass as "parameters are not overfit" for these strategies. The correct interpretation is "edge is reasonably distributed across the regimes covered by the fold window." This distinction should be documented in any CLI output that surfaces walk-forward results.

## Revisit trigger

If the project adds strategies with in-sample parameter fitting (e.g., optimized lookback windows), this decision needs revisiting — for those strategies, walk-forward regains its parameter-overfitting role.
