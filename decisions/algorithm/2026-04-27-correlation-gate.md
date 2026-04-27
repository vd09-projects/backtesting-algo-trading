# Correlation gate — maximum inter-strategy correlation for portfolio construction

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | correlation, portfolio-construction, diversification, stress-period, equity-curve, TASK-0049, TASK-0055 |

## Context

Two strategies with highly correlated equity curves do not provide diversification — in a drawdown, they fall together. The portfolio at ₹3 lakh targeting ~10% annualized vol relies on strategies providing genuinely different return sources. The correlation gate enforces this requirement before any two strategies are combined.

The gate operates on the equity curves of bootstrap survivors (TASK-0054) before portfolio weights are assigned (TASK-0055).

## Return series definition

**Series**: daily log-returns of each strategy's equity curve.

`log_return[t] = ln(equity[t] / equity[t-1])`

Log-returns are used (not simple returns) because they are additive across time and their sum is the total log-return of the strategy over any window — this makes the stress-period subsetting consistent with the full-period comparison.

**Correlation metric**: Pearson r between two strategy log-return series, computed on days where both strategies have equity curve observations. Days where either strategy has no trades and equity is unchanged (flat curve segment) are included — a flat segment is a real return of zero, not missing data.

## Evaluation windows

**Full period**: 2018-01-01 to 2024-12-31 (the entire backtest range). The full-period correlation captures the average co-movement across regimes.

**Stress periods** (both must be evaluated separately):
- COVID crash: 2020-02-01 to 2020-06-30
- Rate-hike bear: 2022-01-01 to 2022-12-31

Stress-period correlation is evaluated on each stress window separately. If either stress window produces r >= 0.6, the pair fails the stress-period test.

These windows are fixed before any strategy results are seen.

## Thresholds

| Window         | Pass condition     |
|----------------|--------------------|
| Full period    | Pearson r < 0.7    |
| Stress period  | Pearson r < 0.6    |

Both conditions must hold for a strategy pair to coexist in the portfolio. Failing either triggers the tiebreaker.

## Options considered

### Option A: Full-period correlation only, threshold 0.7
- **Pros**: Simpler. A single number.
- **Cons**: Two strategies can be uncorrelated on average but highly correlated during drawdowns — which is exactly when diversification is most needed. Average correlation hides the stress-period picture.

### Option B: Full-period r < 0.7 AND stress-period r < 0.6 (chosen)
Two conditions: looser full-period threshold, tighter stress-period threshold. The stress-period threshold is tighter because the relevant question is: "do these strategies fall together when markets are worst?" A portfolio that diversifies on average but concentrates in crises provides false security.

- **Pros**: Directly tests the worst-case scenario. The tighter stress-period threshold (0.6 vs 0.7) is deliberate — if strategies are correlated at 0.65 during a crash, the diversification benefit is marginal.
- **Cons**: Stress-period windows have fewer days (~100 for COVID crash, ~250 for rate-hike bear). Pearson r on short windows is noisy. However, the question is directional — is the correlation high in a crisis? — not a precise estimate.

### Option C: Dynamic conditional correlation (DCC-GARCH)
- **Pros**: Properly models time-varying correlation structure.
- **Cons**: Requires a Python or R dependency; far out of scope for this evaluation stage. Not appropriate.

## Tiebreaker

If a strategy pair fails the gate (either condition), one must be dropped. Selection rule:

1. **Primary**: keep the strategy with the higher DSR-corrected Sharpe from the universe sweep.
2. **Tiebreaker** (if DSR-corrected Sharpe within 5%): prefer the strategy from a different edge bucket than the retained strategy. Edge buckets: trend-following (SMA, Donchian, MACD, Momentum) vs. mean-reversion (RSI, Bollinger Bands). Retaining two strategies from different buckets provides structural diversification even if their Sharpe estimates are similar.

The dropped strategy is recorded with its reason in `decisions/algorithm/`.

## Consequences

- The correlation gate is applied to all surviving strategy × instrument pairs from TASK-0054. If the same strategy survives on multiple instruments, the correlation is computed between the portfolio-level equity curves (all instruments for strategy A vs. all instruments for strategy B), not per-instrument pairs.
- This gate can result in fewer than the target 2-4 strategies in the portfolio. If only one strategy survives the correlation gate, that single strategy is the portfolio — record the outcome with reasons in `decisions/algorithm/`. Do not add a correlated strategy to meet the target count.
- The correlation gate does not revisit the universe or walk-forward gate. A strategy that passes those gates but is dropped by the correlation gate is recorded as "excluded (correlation)" — not a gate failure. It was a valid strategy that happened to be correlated with a better alternative.

## Related decisions

- [Bootstrap gate](./2026-04-27-bootstrap-gate.md) — correlation gate is applied to bootstrap survivors
- [Vol-targeting algorithm choices](./2026-04-13-vol-targeting-algorithm-choices.md) — portfolio sizing method applied after the correlation gate

## Revisit trigger

If a future evaluation adds more than 6 strategy families, the number of pairs grows quadratically. At 10+ strategies, a hierarchical clustering approach (group by correlation, pick one from each cluster) becomes more practical than pairwise comparison. The current pairwise approach is appropriate for up to ~6 strategies.
