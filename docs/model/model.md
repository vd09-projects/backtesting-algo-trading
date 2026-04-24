# Domain Model

**Package:** `pkg/model/`

All shared types live here. Nothing outside `pkg/model` imports another package
in `pkg/model`. It is the leaf of the dependency graph.

---

## Candle

```go
type Candle struct {
    Instrument string        // e.g. "NSE:RELIANCE"
    Timestamp  time.Time
    Timeframe  Timeframe
    Open       float64
    High       float64
    Low        float64
    Close      float64
    Volume     float64
}
```

Candles are created via `model.NewCandle()` which validates:
- `High >= max(Open, Close)`
- `Low <= min(Open, Close)`
- `Low <= High`
- All prices > 0

Invalid candles return an error. The provider never silently passes bad data downstream.

---

## Signal

```go
type Signal string

const (
    SignalBuy  Signal = "buy"
    SignalSell Signal = "sell"
    SignalHold Signal = "hold"
)
```

Hold signals are never passed to the portfolio — the engine discards them.
Only Buy and Sell trigger portfolio actions.

---

## Trade

```go
type Trade struct {
    Instrument string
    Direction  Direction     // DirectionLong (only long trades currently)
    Quantity   float64       // fractional shares allowed
    EntryPrice float64
    EntryTime  time.Time
    ExitPrice  float64
    ExitTime   time.Time     // zero if trade is still open
    Commission float64       // total commission (entry + exit)
    RealizedPnL float64      // set on close: (exitPrice - entryPrice) × qty - commission
}

// ReturnOnNotional returns RealizedPnL / (EntryPrice × Quantity)
// Used by Monte Carlo bootstrap and kill-switch monitoring.
func (t Trade) ReturnOnNotional() float64
```

A trade record exists in two states:
- **Open sentinel:** `ExitTime.IsZero() == true` — position still active
- **Closed:** `ExitTime` set — round-trip complete, `RealizedPnL` populated

Break-even trades (`RealizedPnL == 0`) count as losses in win-rate calculations.

---

## Position

```go
type Position struct {
    Instrument string
    Direction  Direction
    Quantity   float64
    EntryPrice float64
}
```

Only one position per instrument at a time. The Portfolio holds a
`map[string]Position` keyed by instrument string. Attempting to open a second
position in the same instrument while one is already open is silently ignored.

---

## EquityPoint

```go
type EquityPoint struct {
    Timestamp time.Time
    Value     float64   // total portfolio value = Cash + mark-to-market positions
}
```

One point is appended per bar, at that bar's Close. The equity curve covers
every bar from the first candle, including the lookback warmup period.

---

## OrderConfig

```go
type OrderConfig struct {
    SlippagePct     float64          // e.g. 0.0005 = 0.05%
    CommissionModel CommissionModel
    CommissionValue float64          // meaning depends on CommissionModel
}

type CommissionModel int
const (
    CommissionNone       CommissionModel = 0
    CommissionFlat       CommissionModel = 1  // CommissionValue = fixed ₹ per trade
    CommissionPercentage CommissionModel = 2  // CommissionValue = fraction of trade value
    CommissionZerodha    CommissionModel = 3  // 0.03% capped at ₹20 per trade
)
```

---

## SizingModel

```go
type SizingModel int
const (
    SizingFixed            SizingModel = 0  // fixed fraction of cash (default)
    SizingVolatilityTarget SizingModel = 1  // size to target annualized vol
)
```

---

## Timeframe

```go
type Timeframe string
const (
    Timeframe1Min   Timeframe = "1min"
    Timeframe5Min   Timeframe = "5min"
    Timeframe15Min  Timeframe = "15min"
    TimeframeDaily  Timeframe = "daily"
    TimeframeWeekly Timeframe = "weekly"
)
```

Used by the strategy to declare which bar frequency it operates on,
and by analytics to select the correct annualization factor (e.g. 252 for daily,
252 × 25 for 15-minute NSE bars).

---

## Type Relationships

```
Engine.Run receives:
  DataProvider  → produces []Candle
  Strategy      → consumes []Candle, produces Signal

Portfolio holds:
  Cash (float64)
  map[instrument]Position
  []Trade
  []EquityPoint

Analytics.Compute receives:
  []Trade       (closed trades)
  []EquityPoint (equity curve)
  Timeframe     (for annualization)

montecarlo.Bootstrap receives:
  []Trade       (calls .ReturnOnNotional() per trade)
```
