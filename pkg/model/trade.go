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
