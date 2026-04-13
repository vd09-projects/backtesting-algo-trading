package model

// SizingModel determines how position size is calculated when opening a long.
type SizingModel int

const (
	// SizingFixed deploys a fixed fraction of available cash per trade (current default).
	SizingFixed SizingModel = iota
	// SizingVolatilityTarget sizes each trade so the expected annualized dollar volatility
	// of the position equals cash × VolatilityTarget.
	// position notional = (cash × volTarget) / (instrumentVol × sqrt(252))
	// where instrumentVol is the 20-bar realized std dev of daily log returns.
	SizingVolatilityTarget
)
