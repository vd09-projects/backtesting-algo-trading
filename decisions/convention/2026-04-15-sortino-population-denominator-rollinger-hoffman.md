# Sortino uses population-style denominator over all observations (Rollinger-Hoffman)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | convention       |
| Tags     | sortino, downside-deviation, convention, analytics, rollinger-hoffman, TASK-0016 |

## Context

The Sortino ratio uses downside deviation as its denominator. There are two common conventions for computing it:

1. **Rollinger-Hoffman (chosen)**: `sqrt(sum(min(r, 0)²) / n)` — divide by the count of *all* observations.
2. **Alternative**: `sqrt(sum(min(r, 0)²) / n_neg)` — divide by the count of *negative* observations only.

## Options considered

### Option A: Population-style denominator over all observations (Rollinger-Hoffman)
- **Pros**: Most common in practice; agrees with Bloomberg, most risk systems, and Rollinger & Hoffman (2010). Stable estimate even with few negative returns.
- **Cons**: Produces a smaller denominator than the alternative, resulting in a higher (more optimistic) Sortino when negative returns are rare.

### Option B: Divide by count of negative returns only
- **Pros**: More "pure" — only uses observations that constitute downside.
- **Cons**: Produces a more volatile estimate with small samples. Disagrees with what most risk systems report. Less comparable across strategies with different frequencies of negative returns.

## Decision

Sortino uses the population-style denominator over all observations: `sum(min(r, 0)²) / n`. This is the Rollinger-Hoffman convention and the most common in practice. The alternative produces a more volatile estimate with small samples and disagrees with what most risk systems report.

The Calmar ratio uses the same annualization factor as Sharpe (bars/year from the timeframe), converting max drawdown (already a percentage of equity) directly.

## Consequences

Sortino values from this engine are comparable to Bloomberg and standard quant tooling that follows the Rollinger-Hoffman convention. Users should be aware the convention choice matters when comparing to implementations that divide by negative-return count only.

## Revisit trigger

If a user brings a specific risk system or comparison benchmark that uses the alternative convention, consider exposing the convention as a configuration option.

## Related decisions

- [NSE annualization factors for Sharpe and volatility calculations](../convention/2026-04-10-nse-annualization-factors.md) — annualization factors used by Sortino and Calmar.
- [Sharpe ratio uses sample variance (n-1), not population variance (n)](../algorithm/2026-04-10-sharpe-sample-variance.md) — related variance convention choice for Sharpe.
