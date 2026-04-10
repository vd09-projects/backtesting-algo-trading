# Equity curve records every bar, including warmup

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | convention       |
| Tags     | equity-curve, engine, portfolio, lookback, warmup, analytics |

## Context

`Portfolio.RecordEquity` is called once per bar in the engine hot loop. The engine enforces a lookback period — strategy signals are not generated for the first `lookback-1` bars. The question was: should equity snapshots also be skipped during warmup, or should they start from bar 0?

## Options considered

### Option A: Record only from strategy-active bars (i >= lookback)
- **Pros**: Curve starts exactly when the first signal is possible; length matches `len(barResults)`.
- **Cons**: Curve length varies by strategy lookback — callers must coordinate. Breaks the invariant that the curve and the candle slice are the same length. The equity baseline (initial cash, no position) is lost.

### Option B: Record every bar, including warmup (chosen)
- **Pros**: `len(EquityCurve()) == len(candles)` always, regardless of strategy. The pre-trading equity baseline (cash at close of warmup bars) is preserved — useful for Sharpe and drawdown where you want to anchor on initial equity. No coordination needed between engine and analytics.
- **Cons**: Warmup-bar snapshots will always show cash-only equity (no position can be open yet). Minor — this is expected and correct behaviour.

## Decision

`RecordEquity(candles[i])` is called unconditionally inside the loop, before the `if i+1 < lookback { continue }` guard. Every candle, including warmup candles, gets a snapshot.

The invariant `len(portfolio.EquityCurve()) == len(candles)` is intentional and documented in the whitebox tests. Analytics code can rely on it without inspecting strategy lookback.

## Consequences

- Equity curve length always equals the number of candles fetched — a stable invariant for analytics.
- The first `lookback-1` points will always be cash-only (no fill can happen during warmup). Callers computing per-bar returns should be aware that these bars have zero volatility, which will slightly suppress annualized Sharpe. Acceptable — this mirrors real-world behaviour where a strategy is still computing its initial indicator state.
- `RecordEquity` is called before the strategy but after any pending fill — so fills at bar `i`'s open are already reflected in bar `i`'s close snapshot.

## Related decisions

- [MaxDrawdown computed from equity curve, not per-trade losses](../algorithm/2026-04-07-max-drawdown-from-equity-curve.md) — both decisions commit to an equity curve as the authoritative time series for all analytics metrics.
