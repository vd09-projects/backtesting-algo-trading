# Engine

**Package:** `internal/engine/`
**Files:** `engine.go`, `portfolio.go`, `sizing.go`

The engine is the event loop that feeds candles to a strategy one bar at a time,
simulates order fills, and tracks portfolio state. It is the only place in the
codebase where a signal becomes a trade.

---

## How a Single Backtest Run Works

```
engine.New(cfg)
       │
       ▼
engine.Run(ctx, provider, strategy)
       │
       ├── 1. Fetch full candle series from provider
       │         p.FetchCandles(ctx, instrument, timeframe, from, to)
       │
       ├── 2. Validate: lookback >= 1, instrument set, from < to
       │
       ├── 3. Create Portfolio (initial cash, order config, capacity hint)
       │
       └── 4. Event loop: for i, candle := range candles
                  │
                  ├── A. Apply pendingSignal at candle[i].Open
                  │        (the signal stored from the previous bar)
                  │        → Portfolio.applySignal(signal, instrument, Open, time, sizeFrac)
                  │        → pendingSignal = Hold
                  │
                  ├── B. Snapshot equity at candle[i].Close
                  │        → Portfolio.RecordEquity(candle)
                  │        value = Cash + (position.Quantity × candle.Close)
                  │
                  ├── C. Skip if i+1 < lookback  (warmup period)
                  │
                  └── D. Call strategy.Next(candles[:i+1])
                           → pendingSignal = returned signal
                           → append BarResult{Candle, Signal}
```

**Key point:** the signal from bar `i` is stored and applied at bar `i+1`'s Open.
You never get a fill at the same bar's Close that generated the signal.

---

## Engine Config

```go
type Config struct {
    Instrument           string        // e.g. "NSE:RELIANCE"
    From                 time.Time     // inclusive
    To                   time.Time     // exclusive
    InitialCash          float64       // starting ₹
    OrderConfig          model.OrderConfig
    PositionSizeFraction float64       // fraction of cash per trade (SizingFixed)
    SizingModel          model.SizingModel
    VolatilityTarget     float64       // annualized vol target (SizingVolatilityTarget only)
}
```

---

## Portfolio

The Portfolio is the engine's internal ledger. It is not directly accessible during the
run — the engine exposes it via `engine.Portfolio()` after `Run` completes.

```
Portfolio state after Run:
┌─────────────────────────────────────────────────────────────────┐
│  Cash         float64                                           │
│  Positions    map[instrument]model.Position                     │
│  Trades       []model.Trade    (open + closed, in order)        │
│  equityCurve  []model.EquityPoint  (one point per bar)          │
└─────────────────────────────────────────────────────────────────┘

Portfolio methods available after Run:
  .EquityCurve()    → []model.EquityPoint
  .ClosedTrades()   → []model.Trade  (completed round-trips only)
  .OpenTrades()     → []model.Trade  (still open at end of run)
```

### Signal → Portfolio action

```
SignalBuy  → openLong()
  1. Skip if position already open (no pyramiding)
  2. fillPrice = basePrice × (1 + slippagePct)  [buy pays more]
  3. cost = Cash × sizeFraction
  4. quantity = cost / fillPrice
  5. entryCommission = calcCommission(quantity × fillPrice)
  6. Skip if totalCost > Cash
  7. Cash -= totalCost
  8. Positions[instrument] = Position{...}
  9. Trades = append(Trades, open sentinel Trade)

SignalSell → closeLong()
  1. Skip if no open position
  2. fillPrice = basePrice × (1 - slippagePct)  [sell receives less]
  3. exitCommission = calcCommission(fillPrice × quantity)
  4. Find open sentinel trade, complete it:
       ExitPrice = fillPrice
       ExitTime  = fillTime
       Commission += exitCommission
       RealizedPnL = (exitPrice - entryPrice) × quantity - totalCommission
  5. Cash += fillPrice × quantity - exitCommission
  6. delete(Positions, instrument)

SignalHold → no-op
```

---

## Commission Models

| Model | Behavior |
|---|---|
| `CommissionZerodha` | `min(tradeValue × 0.0003, 20)` — 0.03% capped at ₹20 |
| `CommissionFlat` | Fixed ₹ amount per trade (set in `CommissionValue`) |
| `CommissionPercentage` | `tradeValue × CommissionValue` |
| none (default) | Zero commission |

---

## Slippage

```
Buy fill price  = basePrice × (1 + SlippagePct)   // you pay more
Sell fill price = basePrice × (1 - SlippagePct)   // you receive less
```

Default in `cmd/backtest`: `SlippagePct = 0.0005` (0.05%).

---

## Sizing Models

### Fixed (`SizingFixed`)

```
positionSizeFraction = cfg.PositionSizeFraction   (constant, e.g. 0.10 = 10% of cash)
cost = Cash × 0.10
```

### Volatility-Target (`SizingVolatilityTarget`)

```
volTarget = cfg.VolatilityTarget   (e.g. 0.10 = 10% annualized)

At each buy signal:
  1. Compute 20-bar realized vol of log returns on candles seen so far
     (uses the last min(20, available) bars)
  2. fraction = volTarget / (instrumentVol × √252)
  3. fraction = min(fraction, 1.0)   [cap at 100% of cash]
  4. If vol = 0 or too few bars → fraction = 0 → skip buy

Effect: larger position when the instrument is calm; smaller when it's volatile.
        All positions target the same annualized dollar volatility regardless of
        instrument or market regime.
```

---

## What the Engine Returns

After `Run` completes:

```go
eng.Portfolio().EquityCurve()     // []EquityPoint — one per bar, mark-to-market at Close
eng.Portfolio().ClosedTrades()    // []Trade — completed round-trips
eng.Portfolio().OpenTrades()      // []Trade — positions still open at end of run
eng.Candles()                     // []Candle — full series including lookback bars
eng.Results()                     // []BarResult — {Candle, Signal} per active bar
```

The equity curve covers **every bar** from the first candle — not just bars after the
lookback ends. During the lookback period, equity is simply Cash (no position open),
so the curve is flat at the start.
