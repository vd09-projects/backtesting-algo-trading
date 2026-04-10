# Sharpe ratio uses sample variance (n-1), not population variance (n)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | sharpe, analytics, variance, statistics, standard-deviation, sample, population, equity-curve |

## Context

When implementing the Sharpe ratio from per-bar equity curve returns, we must choose between two estimators for standard deviation:

- **Population std dev**: divide sum of squared deviations by **n**
- **Sample std dev**: divide by **n-1** (Bessel's correction)

Backtests operate on finite historical samples, often spanning 1–3 years of daily bars (~250–750 points) or much shorter intraday windows. The choice materially affects computed Sharpe for small samples.

## Options considered

### Option A: Population variance (÷ n)
- **Pros**: Simpler; slightly familiar from textbook definitions.
- **Cons**: Biased downward for finite samples — systematically underestimates true variance, inflating Sharpe. On a 60-bar intraday backtest the bias is ~1.7%; on a 20-bar window it's ~5%. Returns an optimistically high Sharpe on short backtests, which are exactly the cases where we need conservatism.

### Option B: Sample variance (÷ n-1)
- **Pros**: Unbiased estimator for the true population variance. Standard in statistical practice for finite samples. Conservative — slightly lower Sharpe, which is correct when data is limited. Consistent with how Sharpe is computed by Bloomberg, QuantLib, and most quantitative finance libraries.
- **Cons**: Undefined when n < 2 (only 1 return value). Slightly more complex guard required.

## Decision

Use **sample variance (÷ n-1)**. Backtests are always finite samples — we never have the full population of possible returns. The unbiased estimator is statistically correct. The undefined-at-n=1 case is handled by the `len(curve) < 3` guard (which ensures at least 2 returns before any computation).

## Consequences

- Sharpe values are slightly lower (more conservative) than a population-variance implementation. The difference is negligible for long daily backtests (>252 bars) but meaningful for intraday windows of <50 bars.
- `computeSharpe` returns 0 when `len(curve) < 3` (fewer than 2 returns). This threshold exists specifically because sample variance requires n ≥ 2.
- Walk-forward windows (TASK-0022) will likely use short out-of-sample windows (30–90 bars). Using sample variance is especially important there — population variance would produce misleadingly high Sharpe on those small windows.

## Related decisions

- [NSE annualization factors](../convention/2026-04-10-nse-annualization-factors.md) — defines the N in `mean(r)/stddev(r)*sqrt(N)`
- [Sharpe returns 0 for degenerate inputs](../tradeoff/2026-04-10-sharpe-zero-for-degenerate-inputs.md) — handles the n<2 edge case this decision creates

## Revisit trigger

If we ever compare backtest Sharpe values against a third-party benchmark or published strategy result, verify whether that benchmark uses population or sample variance. A mismatch of ~1-2% on daily bars will appear as a systematic discrepancy.
