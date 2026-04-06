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

`MaxDrawdown` is computed from the equity curve: `(peak - trough) / peak * 100`. The equity curve is built inline during the single pass over `[]Trade` by accumulating `RealizedPnL`. Peak is updated whenever equity exceeds the prior peak; drawdown is computed at every point where equity falls below peak.

Implementation:
```go
equity += t.RealizedPnL
if equity > peak { peak = equity }
if peak > 0 {
    dd := (peak - equity) / peak * 100
    if dd > maxDD { maxDD = dd }
}
```

The `peak > 0` guard means drawdown is only computed once the strategy has ever been profitable. If the equity curve never goes positive (all losses from the start), `MaxDrawdown` is 0 — there is no measurable percentage drawdown from a non-existent peak.

## Consequences

- `MaxDrawdown` is the standard metric that most practitioners expect. It can be compared directly to published strategy benchmarks.
- A strategy that loses from bar one will show `MaxDrawdown = 0` — which may be counterintuitive. This is a known limitation of the peak-to-trough definition when there is no positive peak.
- The metric is path-dependent: reordering the same trades can produce a different `MaxDrawdown`.

## Related decisions

- [Trade.RealizedPnL stored on the struct, not computed on-demand](../convention/2026-04-02-trade-pnl-stored-not-computed.md) — the equity curve is built by summing pre-computed `RealizedPnL` fields; no commission or slippage recalculation needed here

## Revisit trigger

If we add intra-trade drawdown tracking (e.g. using candle Low prices to estimate unrealised drawdown during open positions), `MaxDrawdown` would need to be extended beyond the closed-trade equity curve.
