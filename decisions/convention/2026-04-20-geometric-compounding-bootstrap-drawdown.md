# Geometric compounding for per-simulation drawdown in bootstrap

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | experimental     |
| Category | convention       |
| Tags     | montecarlo, bootstrap, drawdown, geometric-compounding, additive, TASK-0024 |

## Context

`worstDrawdown()` in `internal/montecarlo` simulates a portfolio equity path from a resampled sequence of trade returns to estimate max drawdown. The question was whether to model the equity path as geometric (compounding) or additive (linear).

The main backtest uses per-bar EquityPoint snapshots from the engine for drawdown — that simulation is inherently geometric because fills are sized against current equity. The bootstrap uses trade returns on notional (ReturnOnNotional), so we must reconstruct the equity path from scratch.

## Options considered

### Option A: Additive (linear) simulation
`equity += r * initial_capital` for each trade return r.

- **Pros**: Simple; guarantees floor at 0 trivially when initial_capital > 0.
- **Cons**: Physically wrong except when position size is fixed in dollar terms. Vol-targeting explicitly varies position size with equity — additive understates drawdown on large upswings followed by large losses. On sequences with high-variance returns, additive systematically understates risk.

### Option B: Geometric compounding (chosen)
`equity *= (1 + r)` for each trade return r. Floor at 0.

- **Pros**: Physically correct. Each trade's return compounds on the running portfolio value, matching how position sizing actually works (especially under vol-targeting). More conservative for drawdown estimation on sequences with large returns — which is when the estimate matters most for the kill-switch.
- **Cons**: Slightly more complex; requires explicit floor-at-zero guard.

## Decision

Geometric compounding: `equity *= (1 + r)`, floored at 0. The floor is realistic — a long-only book can't go below zero (exchange margin call happens first). Starting equity for the simulation is 1.0 (normalized), so the absolute drawdown % is directly comparable across simulations with different notional sizes.

```go
equity := 1.0
peak := 1.0
worstDD := 0.0
for _, r := range returns {
    equity *= (1 + r)
    if equity < 0 {
        equity = 0
    }
    if equity > peak {
        peak = equity
    }
    if dd := (peak - equity) / peak; dd > worstDD {
        worstDD = dd
    }
}
```

## Consequences

- Bootstrap drawdown estimates will be slightly higher than additive estimates on high-variance return sequences. This is the correct direction for a conservative kill-switch.
- The convention is consistent with how real portfolios compound — no special casing needed for vol-targeting vs. fixed sizing.
- If fixed-dollar position sizing is ever the only use case, additive would give identical results asymptotically (for small r, `1+r ≈ 1`, so the two converge). Geometric is not wrong in that case; it's just not necessary.

## Related decisions

- [MaxDrawdown from equity curve](../algorithm/2026-04-07-max-drawdown-from-equity-curve.md) — the main backtest uses per-bar equity snapshots; bootstrap reconstructs an equivalent path using geometric compounding to stay consistent.
- [Bootstrap Sharpe non-annualized per-trade](../algorithm/2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md) — the bootstrap algorithm this convention serves.
