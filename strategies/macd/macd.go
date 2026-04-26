// Package macd implements a MACD crossover strategy.
// It emits Buy when the MACD line crosses above the signal line, and Sell
// when it crosses below. All other bars emit Hold.
package macd

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on MACD line / signal line crossovers.
type Strategy struct {
	fastPeriod   int
	slowPeriod   int
	signalPeriod int
	timeframe    model.Timeframe
}

// New constructs a MACD crossover Strategy. Returns an error if any period is
// less than 1 or if fastPeriod is not strictly less than slowPeriod.
func New(tf model.Timeframe, fastPeriod, slowPeriod, signalPeriod int) (*Strategy, error) {
	if fastPeriod < 1 {
		return nil, fmt.Errorf("macd: fastPeriod must be >= 1, got %d", fastPeriod)
	}
	if slowPeriod < 1 {
		return nil, fmt.Errorf("macd: slowPeriod must be >= 1, got %d", slowPeriod)
	}
	if signalPeriod < 1 {
		return nil, fmt.Errorf("macd: signalPeriod must be >= 1, got %d", signalPeriod)
	}
	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("macd: fastPeriod (%d) must be < slowPeriod (%d)", fastPeriod, slowPeriod)
	}
	return &Strategy{
		fastPeriod:   fastPeriod,
		slowPeriod:   slowPeriod,
		signalPeriod: signalPeriod,
		timeframe:    tf,
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "macd-crossover" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns slowPeriod + signalPeriod - 1.
// This is the minimum number of bars for the signal line to have one valid
// value. Crossover detection requires one additional bar (handled in Next).
func (s *Strategy) Lookback() int { return s.slowPeriod + s.signalPeriod - 1 }

// Next returns the signal for the current bar. Buy is emitted when the MACD
// line strictly crosses above the signal line (was ≤ previous bar, is >
// current bar). Sell is emitted on the inverse crossover. Hold is returned
// otherwise.
//
// The guard n <= slowPeriod+signalPeriod-1 returns Hold when only one valid
// signal-line value exists; crossover detection requires two consecutive
// valid values.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n <= s.slowPeriod+s.signalPeriod-1 {
		return model.SignalHold
	}

	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	macdLine, signalLine, _ := talib.Macd(closes, s.fastPeriod, s.slowPeriod, s.signalPeriod)

	prevMacd, currMacd := macdLine[n-2], macdLine[n-1]
	prevSig, currSig := signalLine[n-2], signalLine[n-1]

	switch {
	case prevMacd <= prevSig && currMacd > currSig:
		return model.SignalBuy
	case prevMacd >= prevSig && currMacd < currSig:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
