// Package smacrossover implements a simple moving average crossover strategy.
// It emits a Buy signal when the fast SMA crosses above the slow SMA, and a
// Sell signal when it crosses below. All other bars emit Hold.
package smacrossover

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Strategy emits signals based on SMA crossovers.
type Strategy struct {
	fastPeriod int
	slowPeriod int
	timeframe  model.Timeframe
}

// New constructs an SMA crossover Strategy. Returns an error if fastPeriod is
// not strictly less than slowPeriod, or if either period is less than 1.
func New(tf model.Timeframe, fastPeriod, slowPeriod int) (*Strategy, error) {
	if fastPeriod < 1 {
		return nil, fmt.Errorf("smacrossover: fastPeriod must be >= 1, got %d", fastPeriod)
	}
	if slowPeriod < 1 {
		return nil, fmt.Errorf("smacrossover: slowPeriod must be >= 1, got %d", slowPeriod)
	}
	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("smacrossover: fastPeriod (%d) must be < slowPeriod (%d)", fastPeriod, slowPeriod)
	}
	return &Strategy{
		fastPeriod: fastPeriod,
		slowPeriod: slowPeriod,
		timeframe:  tf,
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "sma-crossover" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns the minimum candle history required before Next is called.
// The engine guarantees len(candles) >= Lookback() on every Next call.
func (s *Strategy) Lookback() int { return s.slowPeriod }

// Next returns the signal for the current bar. A Buy is emitted on the bar
// where the fast SMA first crosses above the slow SMA; a Sell is emitted when
// it crosses below. Hold is returned on all other bars.
//
// Crossover detection requires the previous bar's SMAs, so one extra bar
// beyond slowPeriod is needed. Until that bar is available, Hold is returned.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	// Need slowPeriod+1 bars so the previous bar's slow SMA is valid.
	if n <= s.slowPeriod {
		return model.SignalHold
	}

	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	fastVals := talib.Sma(closes, s.fastPeriod)
	slowVals := talib.Sma(closes, s.slowPeriod)

	currFast, prevFast := fastVals[n-1], fastVals[n-2]
	currSlow, prevSlow := slowVals[n-1], slowVals[n-2]

	switch {
	case prevFast <= prevSlow && currFast > currSlow:
		return model.SignalBuy
	case prevFast >= prevSlow && currFast < currSlow:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
