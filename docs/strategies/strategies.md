# Strategies

**Packages:** `pkg/strategy/`, `strategies/`

---

## The Strategy Interface

Every strategy must implement this interface. The engine knows nothing about
the concrete type — it only calls these four methods.

```go
// pkg/strategy/strategy.go
type Strategy interface {
    Name()      string           // human-readable identifier
    Timeframe() model.Timeframe  // declares which candle frequency this strategy needs
    Lookback()  int              // minimum bars required before Next is called
    Next(candles []model.Candle) model.Signal
}
```

### Lookback contract

The engine enforces the lookback:
- `Next` is **not called** until `len(candles) >= Lookback()`
- During warmup, bars are fed to the equity curve (cash is flat, no signals)
- `Lookback()` returning < 1 is a hard error

### No-lookahead contract

`Next` receives `candles[:i+1]` — exactly the bars visible at bar `i`.
The strategy cannot access future bars. This is enforced by the engine,
not by convention.

---

## SMA Crossover

**Package:** `strategies/smacrossover/`

### What it does

```
Buy  when: fastSMA crosses ABOVE slowSMA  (fast was ≤ slow last bar, fast > slow now)
Sell when: fastSMA crosses BELOW slowSMA  (fast was ≥ slow last bar, fast < slow now)
Hold otherwise
```

Strict crossover detection — a crossover requires a direction change between
consecutive bars. If the fast SMA is already above the slow SMA and stays above,
that does not trigger another buy.

### Parameters

| Parameter | Flag | Default | Notes |
|---|---|---|---|
| `fastPeriod` | `--fast-period` | 10 | Fast SMA window in bars |
| `slowPeriod` | `--slow-period` | 50 | Slow SMA window in bars |
| `tf` | `--timeframe` | `daily` | Bar frequency |

### Lookback

`slowPeriod` — the engine waits until at least `slowPeriod` bars have been seen before calling `Next`.
In practice the strategy returns `Hold` for the first `slowPeriod - 1` bars even after the engine
starts calling it, because `go-talib` SMA returns NaN until the window fills.

### Indicator

Uses `github.com/markcheno/go-talib` — `talib.Sma(closes, period)`.
Do not hand-roll SMA math.

---

## RSI Mean Reversion

**Package:** `strategies/rsimeanrev/`

### What it does

```
Buy  when: RSI < oversold threshold   (e.g. RSI < 30 → oversold, expect bounce)
Sell when: RSI > overbought threshold (e.g. RSI > 70 → overbought, take profit)
Hold otherwise
```

### Parameters

| Parameter | Flag | Default | Notes |
|---|---|---|---|
| `rsiPeriod` | `--rsi-period` | 14 | RSI calculation window |
| `oversold` | `--oversold` | 30 | Buy threshold (RSI below this) |
| `overbought` | `--overbought` | 70 | Sell threshold (RSI above this) |
| `tf` | `--timeframe` | `daily` | Bar frequency |

### Lookback

`rsiPeriod + 1` — RSI needs one extra bar for its initial smoothing step.

### Indicator

Uses `github.com/markcheno/go-talib` — `talib.Rsi(closes, period)`.

### Known behavior

RSI mean reversion on NSE daily bars fires relatively infrequently —
overbought/oversold conditions may occur only a handful of times per year,
especially on large-cap instruments with lower volatility. See `cmd/rsi-diagnostic`
for a tool to count signal frequency before committing to a backtest.

---

## Stub

**Package:** `strategies/stub/`

Always returns `SignalHold`. Never opens a position.

**Use case:** smoke-testing the full pipeline. Lets you verify the engine,
provider, and output code work end-to-end without any strategy behavior
interfering with the results.

---

## How to Add a New Strategy

1. Create a new package under `strategies/your-strategy/`
2. Implement `pkg/strategy.Strategy` — all four methods
3. Write tests (TDD: test before implementation)
4. Register it in `cmd/backtest/main.go` `strategyRegistry()`
5. Register it in `cmd/sweep/main.go` `factoryRegistry()` if it has sweepable parameters
6. Register it in `cmd/universe-sweep/main.go` `strategyRegistry()`

### Rules

- Use `go-talib` for all indicators. Do not hand-roll SMA, EMA, RSI, MACD.
- Never reference a concrete strategy type from outside `strategies/your-strategy/`.
- Strategy must be **stateless across calls** if it will be used in walk-forward
  (the same instance is reused across folds). If your strategy carries mutable
  history between `Next` calls you must document this clearly.

---

## Test Utilities

**Package:** `strategies/testutil/`

Provides:
- `testutil.MakeCandles(n, startPrice)` — generates a synthetic candle series for unit tests
- `testutil.FakeStrategy` — a configurable fake that returns preset signals, used in engine tests

These are shared across strategy and engine test packages to avoid duplicating
candle-generation logic.
