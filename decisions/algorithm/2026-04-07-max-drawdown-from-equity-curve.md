# MaxDrawdown computed from equity curve, not per-trade losses

| Field    | Value     |
|----------|-----------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | algorithm  |
| Tags     | analytics, drawdown, equity-curve, metrics, algorithm |

## Context

When implementing `MaxDrawdown` in `analytics.Compute`, there are two common interpretations of what "maximum drawdown" means:

1. The largest single losing trade as a percentage of the account at entry
2. The largest peak-to-trough decline on the running equity curve across the full trade sequence

Both are valid metrics, but they measure different things. The choice here determines what `Report.MaxDrawdown` actually represents.

## Options considered

### Option A: Per-trade drawdown (largest single-trade loss as % of equity at entry)
- **Pros**: Simple to compute — one pass, no peak tracking.
- **Cons**: Misses the effect of consecutive losses compounding. A strategy with five 2% losers in a row has a 10% drawdown on the equity curve, but per-trade drawdown would report only 2%. Understates real risk.

### Option B: Equity-curve drawdown (peak-to-trough on cumulative PnL)
- **Pros**: The standard industry definition. Captures the worst stretch of consecutive losses. Reflects the actual pain a strategy inflicted on the account — how much you would have lost if you entered at the worst possible time relative to the subsequent trough.
- **Cons**: Slightly more complex — requires tracking a running peak and the maximum observed decline from that peak.

## Decision

`MaxDrawdown` is computed from the equity curve: `(peak - trough) / peak * 100`. The implementation
uses `computeMaxDrawdownDepth(curve []model.EquityPoint)` — the per-bar equity curve passed into
`Compute()`, which starts at `initialCash` and records mark-to-market equity at every bar.

The original implementation (2026-04-07) built an inline equity series by accumulating `RealizedPnL`
over closed trades, starting from zero. This was a bug: when cumulative P&L went negative (losses
exceeded all prior gains), the numerator exceeded the denominator and `MaxDrawdown` exceeded 100%.
The fix (2026-04-16) moved to the per-bar `EquityPoint` curve, which starts at `initialCash`, so
the denominator (peak) is always at least the initial capital and the result is bounded to [0, 100].

The `peak > 0` guard means drawdown is only computed once the curve has a positive peak. If equity
never exceeds zero (degenerate scenario), `MaxDrawdown` is 0.

## Consequences

- `MaxDrawdown` is the standard metric that most practitioners expect. It can be compared directly to published strategy benchmarks.
- A strategy that loses from bar one will show `MaxDrawdown = 0` — which may be counterintuitive. This is a known limitation of the peak-to-trough definition when there is no positive peak.
- The metric is path-dependent: reordering the same trades can produce a different `MaxDrawdown`.

## Related decisions

- [Trade.RealizedPnL stored on the struct, not computed on-demand](../convention/2026-04-02-trade-pnl-stored-not-computed.md) — the equity curve is built by summing pre-computed `RealizedPnL` fields; no commission or slippage recalculation needed here

## Revisit trigger

If we add intra-trade drawdown tracking (e.g. using candle Low prices to estimate unrealised drawdown during open positions), `MaxDrawdown` would need to be extended beyond the closed-trade equity curve.
