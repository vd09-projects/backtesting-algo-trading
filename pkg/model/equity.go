package model

import "time"

// EquityPoint is a single snapshot of total portfolio value at a point in time.
// Value = cash + mark-to-market value of any open positions at the bar's close price.
type EquityPoint struct {
	Timestamp time.Time
	Value     float64
}
