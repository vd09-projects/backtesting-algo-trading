// Package donchian implements a Donchian Channel breakout strategy.
// It emits Buy when the current close breaks above the N-bar high of prior bars,
// and Sell when it breaks below the N-bar low. All other bars emit Hold.
package donchian

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on Donchian Channel breakouts.
type Strategy struct {
	period    int
	timeframe model.Timeframe
}

// New constructs a Donchian Channel breakout Strategy.
// Returns an error if period is less than 1.
func New(tf model.Timeframe, period int) (*Strategy, error) {
	if period < 1 {
		return nil, fmt.Errorf("donchian: period must be >= 1, got %d", period)
	}
	return &Strategy{period: period, timeframe: tf}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "donchian-breakout" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns period + 1.
// One extra bar beyond period is required so that the window of prior bars
// contains exactly period bars when the current bar is excluded.
func (s *Strategy) Lookback() int { return s.period + 1 }

// Next returns the signal for the current bar.
// Buy is emitted when Close strictly exceeds the rolling max High of the prior
// period bars. Sell is emitted when Close falls strictly below the rolling min
// Low of the prior period bars. Hold is returned otherwise.
//
// The current bar's own High and Low are intentionally excluded from the channel
// calculation; including them would produce lookahead bias.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n < s.Lookback() {
		return model.SignalHold
	}

	// Slice to prior bars only — current bar's high/low must not enter the window.
	prior := candles[:n-1]
	highs := make([]float64, len(prior))
	lows := make([]float64, len(prior))
	for i, c := range prior {
		highs[i] = c.High
		lows[i] = c.Low
	}

	maxHighs := talib.Max(highs, s.period)
	minLows := talib.Min(lows, s.period)

	channelHigh := maxHighs[len(maxHighs)-1]
	channelLow := minLows[len(minLows)-1]

	curr := candles[n-1]
	switch {
	case curr.Close > channelHigh:
		return model.SignalBuy
	case curr.Close < channelLow:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
