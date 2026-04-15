# Baseline backtest period set to 2018–2024 for NSE strategy evaluation

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | backtest, period, NSE, regime, training-window, TASK-0028 |

## Context

Before running the first baseline strategies (SMA crossover, RSI mean-rev) on NSE daily bar data,
we needed to commit to a test period that would produce meaningful, trustworthy results. The period
selection needed to cover multiple market regimes — a single-regime test would either inflate or
deflate results depending on which regime was chosen. The choice needed to be made before seeing
any strategy output.

## Options considered

### Option A: 2020–2024
- **Pros**: More recent data; 5 years still reasonable length
- **Cons**: Starts with the COVID crash recovery — a sustained bull run that is maximally favorable
  for trend-following. Testing SMA crossover only on 2020–2021 is testing it in its best possible
  environment. Any Sharpe result from this window cannot be trusted.

### Option B: 2015–2024
- **Pros**: 9 years, longer history, more regimes
- **Cons**: Introduces more structural-change risk for Indian equities; the 2016–2017 demonetization
  period adds noise that is unlikely to recur. Longer is not always better if the extra history
  contains non-recurring regimes.

### Option C: 2018–2024
- **Pros**: 6 years covering a pre-crash bull (2018–2019), crash and V-recovery (2020), sustained
  bull (2021), flat-to-down grind (2022), and subsequent recovery (2023–2024). Both trend-following
  and mean-reversion strategies face hostile regimes within the window.
- **Cons**: Excludes 2015–2017 history, but those years add complexity without adding coverage of
  the regimes most likely to recur.

## Decision

**2018-01-01 to 2024-12-31.** The window must include the 2020 crash and the 2022 flat-to-down
regime. These two periods are the real stress tests: the 2020 crash is where mean-reversion
strategies face tail risk (buying dips into a sustained selloff), and 2022 is where trend-following
strategies face their worst environment (directional noise with no sustained trends). A strategy
that can't survive both is not robust.

Testing only on 2020–2021 is the most common mistake in NSE retail backtesting — the post-COVID
recovery was a near-uninterrupted uptrend that makes any buy-the-dip or trend-following strategy
look good. That test proves nothing.

## Consequences

- The 2018-01-01 to 2024-12-31 window is the full evaluation window for TASK-0028 (both SMA
  crossover and RSI mean-rev).
- Sub-period analysis within this window (2018–2019, 2020–2021, 2022–2024) is required as part of
  TASK-0028 to confirm neither strategy shows edge only in one sub-window.
- Walk-forward (TASK-0022) and any CPCV work will use this outer window.
- If strategies are expanded to additional instruments in the future, the same outer window applies
  for comparability.

## Related decisions

- [Strategy proliferation gate — Sharpe ≥ 0.5 vs buy-and-hold before variation strategies](./2026-04-10-strategy-proliferation-gate.md) — the gate check uses results from this window

## Revisit trigger

If the strategy passes the proliferation gate on 2018–2024 but noticeably degrades on 2025+ live
or out-of-sample data, reconsider whether the training window should be rolled forward or extended.
Also revisit if a future strategy requires a longer history to generate sufficient trades.
