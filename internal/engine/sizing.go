package engine

import (
	"math"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// sizeFractionForBar returns the position size fraction to use for a buy signal
// at the current bar, given the historical candles seen so far.
//
// For SizingFixed it returns cfg.PositionSizeFraction unchanged.
// For SizingVolatilityTarget it sizes to target annualized dollar volatility:
//
//	fraction = volTarget / (instrumentVol × √252)
//
// capped at 1.0. Returns 0 if vol cannot be computed (too few bars or vol=0),
// which causes the buy to be skipped.
func sizeFractionForBar(cfg Config, candles []model.Candle) float64 { //nolint:gocritic // Config value semantics intentional; consistent with engine.New
	if cfg.SizingModel != model.SizingVolatilityTarget {
		return cfg.PositionSizeFraction
	}

	vol := computeInstrumentVol(candles)
	if vol == 0 {
		return 0
	}

	fraction := cfg.VolatilityTarget / (vol * math.Sqrt(252))
	if fraction > 1 {
		return 1
	}
	return fraction
}

// computeInstrumentVol returns the 20-bar realized standard deviation of daily
// log returns using the tail of candles. Returns 0 if there are fewer than 3
// candles (need at least 2 returns for sample std dev).
//
// The window is capped at 20 bars: only the last 20 candles are used.
func computeInstrumentVol(candles []model.Candle) float64 {
	n := len(candles)
	if n < 3 {
		// Need at least 2 returns (3 price points) for a meaningful sample std dev.
		return 0
	}

	const window = 20
	start := n - window
	if start < 0 {
		start = 0
	}

	// Compute log returns for bars start+1 .. n-1.
	returns := make([]float64, 0, n-start-1)
	for i := start + 1; i < n; i++ {
		if candles[i-1].Close <= 0 || candles[i].Close <= 0 {
			continue
		}
		returns = append(returns, math.Log(candles[i].Close/candles[i-1].Close))
	}

	if len(returns) < 2 {
		return 0
	}

	// Sample mean.
	mean := 0.0
	for _, r := range returns {
		mean += r
	}
	mean /= float64(len(returns))

	// Sample variance (n-1 denominator).
	variance := 0.0
	for _, r := range returns {
		d := r - mean
		variance += d * d
	}
	variance /= float64(len(returns) - 1)

	return math.Sqrt(variance)
}
