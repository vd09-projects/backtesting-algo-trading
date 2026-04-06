package model

// Direction indicates whether a position is long or short.
type Direction string

// Supported position directions.
const (
	DirectionLong  Direction = "long"
	DirectionShort Direction = "short"
)

// Position represents an open holding in a single instrument.
type Position struct {
	Instrument string
	Direction  Direction
	Quantity   float64
	EntryPrice float64
}
