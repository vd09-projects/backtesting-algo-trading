# Bootstrap gate pass/fail thresholds

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | bootstrap, montecarlo, sharpe, gate, evaluation-methodology, TASK-0049, TASK-0054 |

## Context

The Monte Carlo bootstrap (TASK-0024, `internal/montecarlo.Bootstrap`) produces a distribution of per-trade Sharpe values from 10,000 resampled trade sequences. A strategy with a high point-estimate Sharpe but a wide bootstrap distribution — meaning low p5 — has too much sampling variance to trust with real capital. Two complementary thresholds are needed to capture both the lower tail and the mass of the distribution.

This gate is applied after universe sweep and walk-forward validation (TASK-0054). Strategies that pass this gate proceed to correlation screening and portfolio construction.

## Options considered

### Option A: p5 > 0 only
A single threshold on the left tail. A strategy passes if fewer than 5% of bootstrap simulations show negative per-trade Sharpe.

- **Pros**: Simple. Directly answers "is there a floor of evidence for positive edge?"
- **Cons**: A strategy with p5 = 0.01 and p50 = 0.02 passes — the whole distribution is barely above zero, which is not a convincing edge.

### Option B: p5 > 0 AND P(Sharpe > 0) > 80% (chosen)
Two conditions: (1) the 5th percentile of the bootstrap Sharpe distribution is positive, and (2) at least 80% of the 10,000 simulated Sharpe values are positive.

- **Pros**: p5 > 0 guards the left tail; the 80% probability mass condition guards against strategies that have a positive p5 but a wide, barely-positive distribution. Together they require both a meaningful floor and sufficient probability concentration above zero.
- **Cons**: Two thresholds are harder to explain. The 80% threshold is a pre-committed number, not derived from data — but it was declared before any bootstrap results were seen.

### Option C: p5 > 0.1 (absolute level)
Require the left tail to land at an absolute Sharpe level, not just positive.

- **Pros**: Rules out very weak edge.
- **Cons**: The relevant comparison is whether the edge is real, not whether it is strong enough — strength is captured by the universe gate and the portfolio allocation. A minimum-level gate conflates two separate questions.

## Decision

A strategy passes the bootstrap gate if **SharpeP5 > 0 AND P(Sharpe > 0) > 80%**.

- `SharpeP5`: the 5th percentile of the bootstrap per-trade Sharpe distribution from `montecarlo.BootstrapResult.SharpeP5`. Must be strictly positive.
- `P(Sharpe > 0)`: fraction of the 10,000 bootstrap simulations that produced positive per-trade Sharpe, from `montecarlo.BootstrapResult.ProbPositiveSharpe`. Must exceed 80%.

Both conditions must hold. Failing either condition kills the strategy — record the kill decision in `decisions/algorithm/`.

**Sharpe computation**: per-trade non-annualized, `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance (n-1), no annualization factor. Consistent with the standing order in `2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md`.

**Bootstrap parameters**: 10,000 simulations, seed logged with every result for reproducibility.

## Consequences

- A strategy with a high in-sample Sharpe but fewer than 30 trades will almost certainly fail this gate — the bootstrap distribution on a 20-trade history is essentially noise. This is intentional: sparse data provides no reliable bootstrap signal.
- The `SharpeP5` value from a passing bootstrap run feeds directly into the kill-switch threshold derivation (TASK-0056). This is the mechanism that ties pre-deployment evidence to live monitoring — not a round number.
- The live kill-switch computation must use the identical per-trade Sharpe formula. Using annualized Sharpe in live monitoring would make the threshold meaningless (the bootstrap Sharpe and the live Sharpe would be on different scales).

## Related decisions

- [Bootstrap Sharpe: non-annualized per-trade computation](./2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — defines the Sharpe formula used here
- [Kill-switch derivation methodology](./2026-04-21-kill-switch-derivation-methodology.md) — consumes SharpeP5 from this gate
- [Cross-instrument universe gate](./2026-04-25-cross-instrument-proliferation-gate.md) — first gate; bootstrap is the third

## Revisit trigger

If a strategy regularly produces ≥ 100 trades in the evaluation window, the 80% probability mass threshold may be too permissive — a high-frequency strategy should be able to demonstrate P(Sharpe > 0) > 90%. Revisit when the first high-frequency strategy enters evaluation.
