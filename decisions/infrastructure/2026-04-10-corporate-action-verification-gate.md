# Zerodha corporate action verification required before running any strategy

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | infrastructure   |
| Tags     | zerodha, data-quality, corporate-action, adjusted-prices, unadjusted, split, dividend, gate, TASK-0025 |

## Context

Zerodha's Kite Connect historical candle API behaviour regarding corporate actions (stock splits, bonus shares, dividends) is not definitively documented for all endpoints. Some endpoints return unadjusted prices by default. In unadjusted price data, a 1:5 stock split appears as an 80% price drop in a single bar — a strategy will fire a spurious sell signal, the equity curve will record a phantom drawdown, and Sharpe/Calmar metrics will be corrupted. This is a silent failure: the backtest runs to completion with no error, but the output is wrong.

For daily-bar strategies covering 3–5 years of NSE data, most NIFTY 50 constituents have at least one corporate action (split, bonus, or large dividend) in that window.

## Options considered

### Option A: Proceed with first strategy and verify data quality empirically
- **Pros**: Faster to first result.
- **Cons**: Results may be corrupted; corrupted results are worse than no results because they waste time evaluating a phantom signal. If the equity curve shows an unexplained crash, root-causing it after the fact is expensive.

### Option B: Verify Zerodha API adjustment behaviour before running any strategy (TASK-0025)
- **Pros**: One-time analysis; result is definitive; either the data is clean (proceed) or we know exactly what to fix before wasting a single strategy run.
- **Cons**: Small upfront time cost.

## Decision

Option B. TASK-0025 is a mandatory gate before TASK-0012 (SMA crossover) or TASK-0015 (RSI mean-reversion) are executed.

Verification method: fetch a NIFTY 50 constituent with a known split or bonus in the last 5 years and check price continuity around that date. If prices show a sharp discontinuity consistent with an unadjusted split, Zerodha returns unadjusted data on that endpoint and a corporate action adjustment layer must be added before any strategy is run. If prices are smooth through the event, the data is adjusted.

The outcome is recorded as a decision regardless of result. If unadjusted: a new task for a corporate action layer is created and becomes a blocker for TASK-0012 and TASK-0015. If adjusted: the verification is documented with the specific endpoint tested so future engineers don't re-ask the question.

## Consequences

- Adds one analysis task (TASK-0025) before the first strategy can run.
- If Zerodha returns unadjusted prices, the project timeline extends by the time required to build an adjustment layer — but this is unavoidable. Running strategies on unadjusted data and then discovering the problem produces worthless results that must be thrown away.
- The check is cheap (one API call + visual inspection) relative to the cost of a corrupted multi-week backtest run.

## Related decisions

- [Zerodha pagination strategy](architecture/2026-04-07-zerodha-pagination-strategy.md) — rate limits and chunking strategy for the same API

## Revisit trigger

If the Zerodha API is upgraded or a different data endpoint is adopted (e.g., switching to NSE's own data feed or a third-party vendor), re-run this verification for the new source. The adjustment status of one endpoint does not transfer to another.
