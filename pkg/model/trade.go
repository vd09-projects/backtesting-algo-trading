package model

import "time"

// Trade is a completed round-trip (open + close) for a single instrument.
// RealizedPnL is stored directly after accounting for commission and slippage —
// analytics reads it without recomputing.
type Trade struct {
	Instrument  string
	Direction   Direction
	Quantity    float64
	EntryPrice  float64 // slippage-adjusted fill price on entry
	ExitPrice   float64 // slippage-adjusted fill price on exit
	EntryTime   time.Time
	ExitTime    time.Time
	Commission  float64 // total commission paid across entry and exit fills
	RealizedPnL float64
}

// ReturnOnNotional returns RealizedPnL divided by entry notional (EntryPrice × Quantity).
// This is the per-trade return the Monte Carlo bootstrap resamples from (TASK-0024).
//
// The p5 Sharpe of the bootstrapped distribution is the kill-switch threshold for live
// monitoring (TASK-0026): halt when the rolling per-trade Sharpe on ReturnOnNotional()
// values drops below it. The live metric must use the same non-annualized computation.
//
// Returns 0 when EntryPrice or Quantity is zero.
func (t Trade) ReturnOnNotional() float64 { //nolint:gocritic // value receiver is intentional; Trade is always used by value (see decisions/convention/2026-04-06)
	notional := t.EntryPrice * t.Quantity
	if notional == 0 {
		return 0
	}
	return t.RealizedPnL / notional
}
