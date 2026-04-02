package model

// Signal is the trading instruction emitted by a strategy for a given bar.
type Signal string

const (
	SignalBuy  Signal = "buy"
	SignalSell Signal = "sell"
	SignalHold Signal = "hold"
)

func (s Signal) String() string {
	return string(s)
}
