# BenchmarkReport annualized return uses actual elapsed calendar time, not bar-count

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | benchmark, CAGR, annualized-return, elapsed-time, BenchmarkReport, analytics, TASK-0018 |

## Context

`ComputeBenchmark` needed to annualize the buy-and-hold total return into a CAGR figure. Two approaches were available: (1) count bars and multiply by the timeframe's annualization factor, or (2) compute actual elapsed calendar time from the first and last candle timestamps.

## Options considered

### Option A: Bar-count × annualization factor
- **Pros**: Consistent with how `computeSharpe` annualizes volatility; uses the same `sharpeAnnualizationFactor` lookup.
- **Cons**: Overcounts on sparse data (a 252-bar window spanning 18 months of daily bars includes weekends, holidays, and gaps — the bar count and calendar time diverge). For daily NSE data, 252 bars is a full trading year, not a calendar year, but CAGR is conventionally quoted in calendar years.

### Option B: Actual elapsed time from timestamps
- **Pros**: Matches how CAGR is quoted in practice (calendar years). Robust to gaps, holidays, and varying session density. Directly comparable to benchmarks quoted externally (e.g., Nifty 50 CAGR). Formula: `years = duration.Hours() / (365.25 * 24)`.
- **Cons**: Requires valid `Timestamp` fields on candles; returns 0 if duration is zero (single-day or same-day candles).

## Decision

Option B — elapsed calendar time. CAGR is a compounded annualized growth rate expressed in calendar terms. Expressing it in trading-year terms would make results non-comparable with any external reference. The formula uses 365.25 days per year to account for leap years.

Sharpe annualization is kept on bar-count (Option A equivalent) because Sharpe measures the ratio of return to volatility — both of which are accumulated at the bar frequency, so bar-count annualization is mathematically correct there.

## Consequences

- CAGR and Sharpe annualization use different time bases (calendar vs. bar-frequency). This is intentional and consistent with how quantitative finance treats the two metrics.
- If candle timestamps have gaps (e.g., a backtest starting mid-year and ending mid-year), CAGR is still accurate because it uses actual wall-clock duration.
- A degenerate input where first and last candles share the same timestamp returns `AnnualizedReturn = 0`.

## Related decisions

- [NSE annualization factors for Sharpe and volatility calculations](../convention/2026-04-10-nse-annualization-factors.md) — covers bar-count annualization used by Sharpe; CAGR deliberately diverges from this.
- [Sharpe returns 0 for degenerate inputs](../tradeoff/2026-04-10-sharpe-zero-for-degenerate-inputs.md) — same zero-default pattern applied here for zero-duration input.
