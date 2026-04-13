# Vol-targeting sizing: algorithm choices for window, returns, zero-vol, and cap

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | volatility-targeting, position-sizing, SizingModel, log-returns, sample-variance, 20-bar-window, no-lookahead, TASK-0021 |

## Context

TASK-0021 introduced `SizingVolatilityTarget`, a new sizing model where position notional is
sized so the expected annualised dollar volatility of the position equals `cash × volTarget`.
Several non-obvious implementation choices had to be made: how to estimate instrument vol, what
to do at degenerate inputs, and when to compute it within the engine's fill loop.

## Options considered

### Window length for realized vol

#### Option A: 20-bar rolling window (chosen)
- **Pros**: Industry standard for "1-month historical vol" on daily data; short enough to track
  recent regime; consistent with how most vol-targeting quant literature defines instrument vol.
- **Cons**: Hardcoded — not configurable in v1.

#### Option B: EWMA (exponentially weighted) vol
- **Pros**: More responsive to recent observations; used in RiskMetrics.
- **Cons**: Requires choosing a decay parameter (lambda); adds complexity for v1; no evidence
  yet that responsiveness matters more than simplicity.

### Return type

#### Option A: Log returns — ln(close[i] / close[i-1]) (chosen)
- **Pros**: Time-additive; symmetric under normal distribution assumption; standard in quant
  finance for vol estimation.
- **Cons**: Marginally slower to compute than simple returns.

#### Option B: Simple returns — (close[i] - close[i-1]) / close[i-1]
- **Pros**: Slightly simpler.
- **Cons**: Not time-additive; biased for compounding; inconsistent with how annualized vol is
  conventionally defined in portfolio theory.

### Variance estimator

#### Option A: Sample variance (n-1 denominator) (chosen)
- **Pros**: Unbiased estimator for a finite return series; consistent with the Sharpe ratio
  implementation (see related decisions).
- **Cons**: Marginally inflated variance vs population estimator on small windows.

#### Option B: Population variance (n denominator)
- **Pros**: None — systematically underestimates variance on short windows, inflating the vol
  estimate in the wrong direction (smaller vol → larger position → more unintended risk).

### Behaviour when vol = 0 or history is insufficient

#### Option A: Return fraction = 0, skip the buy (chosen)
- **Pros**: Conservative; prevents division by zero cleanly; a strategy that fires a Buy signal
  before the vol window is warm simply waits — no partial fill at a wrong size.
- **Cons**: A Buy signal is silently dropped for the first ~20 bars of any run.

#### Option B: Fall back to `PositionSizeFraction` when vol = 0
- **Pros**: Never silently skips a signal.
- **Cons**: Produces inconsistent sizing depending on history length — some fills are vol-targeted,
  others are fixed-fraction. Hard to attribute results to the sizing model during analysis.

### Fraction cap

#### Option A: Cap at 1.0 (chosen)
- **Pros**: A very low-vol instrument at a moderate vol target can produce `fraction > 1`.
  Capping prevents deploying more than available cash. Preserves "deploy this fraction of cash"
  semantics.
- **Cons**: In low-vol regimes the position is under-sized relative to the vol target — but this
  is correct behaviour (you can't lever up).

### Timing: which candles to use for vol at fill time

#### Option A: `candles[:i]` — bars the strategy saw when it emitted the signal (chosen)
- **Pros**: No lookahead. The fill executes at `candles[i].Open`; using `candles[:i]` means vol
  is computed from the same history the strategy had. Strictly correct.
- **Cons**: The vol estimate is one bar stale relative to the fill price — acceptable.

#### Option B: `candles[:i+1]` — include current bar up to its open
- **Pros**: Slightly fresher estimate.
- **Cons**: `candles[i].Close` is not yet available at fill time (fill is at open). Using it
  would be a lookahead violation.

## Decision

20-bar rolling sample std dev of daily log returns. Zero vol (insufficient history or constant
prices) → fraction = 0 → buy silently skipped. Fraction capped at 1.0. Vol computed from
`candles[:i]` at fill time, not `candles[:i+1]`, to preserve no-lookahead.

Formula:

```
instrumentVol = sample_stddev(log_returns(candles[n-20 : n]))   // n = bars seen so far
fraction      = volTarget / (instrumentVol × sqrt(252))
fraction      = min(fraction, 1.0)
```

The 20-bar window is hardcoded in v1 (`const window = 20` in `sizing.go`). If future strategies
need a configurable window, add `VolWindow int` to `engine.Config` at that point — not now.

## Consequences

- Buy signals in the first ~20 bars of any backtest are silently skipped under
  `SizingVolatilityTarget`. Strategy results with short data windows will show fewer trades
  than expected.
- Results are not comparable between `SizingFixed` and `SizingVolatilityTarget` runs — the
  position size per trade will differ even on the same signal sequence. Always note which model
  was used when comparing Sharpe ratios.
- Very low-vol instruments at moderate vol targets will hit the `fraction > 1` cap, meaning the
  vol target is not actually achieved. This is a known limitation of the per-trade sizing approach
  (vs a portfolio-level vol budget).

## Related decisions

- [Sharpe uses sample variance (n-1)](2026-04-10-sharpe-sample-variance.md) — same estimator
  choice applied consistently.
- [NSE annualization factors](../convention/2026-04-10-nse-annualization-factors.md) — sqrt(252)
  comes from this convention; do not substitute sqrt(252) with sqrt(252 × something) for intraday.
- [No pyramiding in v1](../tradeoff/2026-04-03-no-pyramiding-v1.md) — positions are all-in/all-out,
  so vol targeting sizes the whole position, not an incremental add.

## Revisit trigger

If a strategy with a 20-bar lookback (e.g., RSI-14 plus engine overhead) shows systematically
fewer trades than expected, consider whether the 20-bar vol window is eating into signal
generation. May need a configurable `VolWindow` or a shorter default.
