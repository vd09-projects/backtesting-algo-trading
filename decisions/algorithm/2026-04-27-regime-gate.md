# Regime gate — no single regime accounts for more than 70% of total Sharpe

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | regime, sharpe, gate, evaluation-methodology, concentration-risk, TASK-0049, TASK-0052 |

## Context

A strategy can pass the universe sweep with a strong DSR-corrected average Sharpe while being entirely dependent on one market regime. A trend-following strategy might generate all of its Sharpe during the COVID crash and recovery (2020-2021) while being flat or negative in the pre-COVID and post-recovery periods. That strategy is not testing a robust edge — it is capturing one regime.

The regime gate requires that the Sharpe contribution be distributed across regimes, not concentrated in one.

## Regime windows

Three NSE equity regimes covering the 2018-2024 evaluation period:

| Regime label        | Window                              | Character                                      |
|---------------------|-------------------------------------|------------------------------------------------|
| Pre-COVID           | 2018-01-01 to 2020-01-31           | Moderate vol, mixed trending and sideways      |
| COVID crash+recovery| 2020-02-01 to 2021-06-30           | Sharp crash, V-shaped recovery, high vol       |
| Post-recovery       | 2021-07-01 to 2024-12-31           | Grinding uptrend, rate-hike bear 2022, lower vol |

These windows are fixed. They do not move based on strategy results — that would be post-hoc rationalization.

## Contribution metric

For each regime window, compute the per-trade Sharpe on the trades that fell in that window. Let `S[i]` be the per-trade Sharpe in regime `i` (can be negative).

The contribution fraction for regime `i` is:

```
contribution[i] = abs(S[i]) / sum(abs(S[j]) for all j)
```

Using absolute values in both numerator and denominator. This avoids a divide-by-zero or sign-cancellation problem that occurs when regimes have mixed signs and their sum approaches zero — using the raw sum as a denominator would produce a meaningless fraction.

The gate fails if any single `contribution[i] >= 0.70`.

## Options considered

### Option A: Full-period Sharpe as denominator
`contribution[i] = S[i] / full_period_sharpe`

- **Pros**: Intuitively appealing — what fraction of the "total edge" comes from each regime?
- **Cons**: Breaks when regimes have mixed signs. A strategy with S[pre-COVID] = -0.3, S[COVID] = 0.8, S[post-recovery] = 0.1 has full-period Sharpe = 0.6, and the pre-COVID contribution would be -0.3/0.6 = -50% — a fraction outside [0,1] that misrepresents the regime picture. Rejected.

### Option B: abs(S[i]) / sum(abs(S[j])) (chosen)
Normalizes each absolute regime Sharpe by the total absolute Sharpe mass. Values always sum to 1.0. Negative regimes appear as large contributions, not negative fractions.

- **Pros**: Well-defined for all sign combinations. A strategy that is strongly negative in one regime is flagged, not hidden.
- **Cons**: A strategy with S = [-0.4, 0.5, 0.3] would have contributions [0.33, 0.42, 0.25] — this passes the gate, even though the pre-COVID regime is a loss. The gate is specifically about concentration, not sign — the universe and walk-forward gates handle sign requirements.

### Option C: Equal-weight regime requirement (each regime must have Sharpe > 0)
- **Pros**: Directly tests robustness.
- **Cons**: Overly strict for low-frequency strategies where one regime may have too few trades to produce a reliable Sharpe estimate. Merges two questions (regime-specific sign and regime concentration) that should be evaluated separately.

## Decision

**Gate condition**: `max(contribution[i]) < 0.70` across all three regime windows.

A strategy failing this gate is **not killed outright** — it receives **half-weight** in the portfolio. Regime concentration is a concern, not a disqualifier: the strategy may be capturing real edge in a specific structural environment, and that edge may persist. The half-weight penalizes the concentration risk without discarding the edge.

The outcome is recorded per strategy:
- Pass: `RegimeConcentrated = false`, full weight eligible
- Flag: `RegimeConcentrated = true` (any regime ≥ 70% contribution), half-weight in portfolio

A strategy that simultaneously fails the universe or walk-forward gate is killed by those gates — the regime gate flag is not applied to already-killed strategies.

## Consequences

- A strategy must have at least some trades in each regime window to be evaluated. If a strategy has zero trades in one regime, that regime Sharpe is undefined — treat the strategy as regime-concentrated by default if any regime window is empty.
- Regime Sharpe is per-trade (same formula as the main evaluation: `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance, no annualization). Consistent with all other per-trade Sharpe computations.
- Half-weight applies during portfolio construction (TASK-0055). A regime-concentrated strategy that otherwise passes all gates still enters the portfolio, but receives at most 50% of the capital weight it would otherwise receive under vol-targeting.

## Related decisions

- [Cross-instrument universe gate](./2026-04-25-cross-instrument-proliferation-gate.md) — first gate; must pass before regime gate is applied
- [Walk-forward OOS/IS Sharpe threshold](./2026-04-22-walk-forward-oos-is-sharpe-threshold.md) — second gate
- [Bootstrap gate](./2026-04-27-bootstrap-gate.md) — third gate; regime gate is applied in universe sweep (TASK-0052)

## Revisit trigger

If the three regime windows are redefined (e.g., to add a 2025 live period), the contribution fractions must be recomputed. The 70% threshold was set for three windows — adding a fourth window reduces the expected contribution per window and may require lowering the threshold.
