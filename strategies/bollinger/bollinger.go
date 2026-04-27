// Package bollinger implements a Bollinger Band mean-reversion strategy.
// It emits Buy when close crosses below the lower band, and Sell when close
// crosses above the upper band. All other bars emit Hold.
package bollinger

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on Bollinger Band crossovers.
type Strategy struct {
	period    int
	numStdDev float64
	timeframe model.Timeframe
}

// New constructs a Bollinger Band mean-reversion Strategy. Returns an error if
// period is less than 1 or numStdDev is not strictly positive.
func New(tf model.Timeframe, period int, numStdDev float64) (*Strategy, error) {
	if period < 1 {
		return nil, fmt.Errorf("bollinger: period must be >= 1, got %d", period)
	}
	if numStdDev <= 0 {
		return nil, fmt.Errorf("bollinger: numStdDev must be > 0, got %g", numStdDev)
	}
	return &Strategy{
		period:    period,
		numStdDev: numStdDev,
		timeframe: tf,
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "bollinger-mean-reversion" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns period — the minimum number of bars for one valid BB value.
// Crossover detection requires one additional bar (handled in Next).
func (s *Strategy) Lookback() int { return s.period }

// Next returns the signal for the current bar. Buy is emitted when close
// strictly crosses below the lower Bollinger Band (was >= lower previous bar,
// is < lower current bar). Sell is emitted when close strictly crosses above
// the upper band. Hold is returned otherwise.
//
// The guard n <= period returns Hold when only one valid BB value exists;
// crossover detection requires two consecutive valid values.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n <= s.period {
		return model.SignalHold
	}

	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	upper, _, lower := talib.BBands(closes, s.period, s.numStdDev, s.numStdDev, talib.SMA)

	prevClose, currClose := closes[n-2], closes[n-1]
	prevLower, currLower := lower[n-2], lower[n-1]
	prevUpper, currUpper := upper[n-2], upper[n-1]

	switch {
	case prevClose >= prevLower && currClose < currLower:
		return model.SignalBuy
	case prevClose <= prevUpper && currClose > currUpper:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
