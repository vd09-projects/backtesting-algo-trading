# Bootstrap Sharpe: non-annualized per-trade computation

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | bootstrap, montecarlo, sharpe, annualization, per-trade, kill-switch, TASK-0024, TASK-0026 |

## Context

When adding Monte Carlo bootstrap for Sharpe confidence intervals (TASK-0024), the annualization question was explicit: should `Bootstrap()` compute annualized Sharpe (to match the main `Report.SharpeRatio`) or per-trade Sharpe? The two yield different numeric ranges and are not directly comparable.

The constraint is what can be computed honestly from the resampled data. Bootstrap resamples individual trades, not bars. Each row of the resample is one `Trade`, carrying `ReturnOnNotional`.

## Options considered

### Option A: Annualized Sharpe
Annualize the per-simulation Sharpe by multiplying by √(annualization_factor / avg_trades_per_year).

- **Pros**: Comparable to main report's SharpeRatio; familiar units for practitioners.
- **Cons**: The annualization factor depends on trading frequency which varies per-simulation. More critically, annualized Sharpe requires preserving the autocorrelation structure of the return series — achieved via block bootstrap on bar-level returns, not trade-level resampling. Applying an annualization factor to a trade-resampled statistic is statistically incorrect.

### Option B: Per-trade Sharpe (chosen)
`mean(ReturnOnNotional) / std(ReturnOnNotional)` across the simulation's trades. No annualization.

- **Pros**: Correctly computed from what is actually being resampled. Consistent with how the kill-switch (TASK-0026) will evaluate live performance. Honest about what the number represents.
- **Cons**: Not directly comparable to the annualized Sharpe in the main report. Requires clear labeling in output ("Per-trade Sharpe") to prevent confusion.

## Decision

Per-trade Sharpe only. `mean(r) / std(r)` on return-on-notional, sample variance (n-1), no annualization factor applied. Annualizing a trade-resampled statistic would require preserving bar-level autocorrelation structure (block bootstrap) — that machinery is out of scope for TASK-0024.

The output header labels this explicitly as "Per-trade Sharpe" to distinguish it from the annualized Sharpe in the main backtest report.

## Consequences

1. The bootstrap CI (`p5`, `p50`, `p95`) is not directly comparable to `Report.SharpeRatio`. Users must understand they are different metrics on different bases.
2. TASK-0026 (kill-switch definition) **must** use the identical per-trade computation. Using annualized Sharpe in live monitoring would make the kill-switch measure a different quantity than what bootstrap tested — the threshold would be meaningless.
3. If block-bootstrap on bar-level returns is ever added (to produce annualized CIs), it is a separate function, not a modification of `Bootstrap()`. The existing signature and behavior must remain stable.

## Related decisions

- [NSE annualization factors](../convention/2026-04-10-nse-annualization-factors.md) — sets the annualization factors used by the main report; this decision explicitly does NOT use them in bootstrap.
- [Sharpe uses sample variance (n-1)](../algorithm/2026-04-10-sharpe-sample-variance.md) — sample variance convention adopted here too.

## Revisit trigger

When TASK-0026 is built: verify that live kill-switch computation uses `mean(r)/std(r)` on return-on-notional (not annualized Sharpe). If they diverge, this decision must be updated.
