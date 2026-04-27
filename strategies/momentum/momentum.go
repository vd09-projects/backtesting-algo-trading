// Package momentum implements a 12-month rate-of-change momentum strategy based
// on the Jegadeesh-Titman (1993) academic equity momentum factor. The "skip last
// month" convention (231 bars = 252 − 21) avoids short-term reversal
// contamination and is the standard AQR/academic implementation.
package momentum

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on a rate-of-change momentum threshold.
// A Buy signal is emitted when ROC(lookback) strictly exceeds threshold;
// a Sell signal is emitted when ROC(lookback) is strictly below -threshold.
// All other bars emit Hold.
type Strategy struct {
	lookback  int
	threshold float64
	timeframe model.Timeframe
}

// New constructs a momentum Strategy. Returns an error if lookback is less than 1
// or threshold is not strictly positive.
func New(tf model.Timeframe, lookback int, threshold float64) (*Strategy, error) {
	if lookback < 1 {
		return nil, fmt.Errorf("momentum: lookback must be >= 1, got %d", lookback)
	}
	if threshold <= 0 {
		return nil, fmt.Errorf("momentum: threshold must be > 0, got %g", threshold)
	}
	return &Strategy{
		lookback:  lookback,
		threshold: threshold,
		timeframe: tf,
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "momentum" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns lookback + 1 — the minimum number of bars required for
// talib.Roc to produce one valid output value at the last position.
// talib.Roc zero-fills the first lookback positions; lookback+1 bars ensures
// roc[n-1] is always a real computation.
func (s *Strategy) Lookback() int { return s.lookback + 1 }

// Next returns the signal for the current bar.
// Buy is emitted when ROC(lookback) is strictly greater than threshold.
// Sell is emitted when ROC(lookback) is strictly less than -threshold.
// Hold is returned otherwise, including when the bar count is below Lookback().
//
// Signal semantics are level-comparison (not crossover): each bar's ROC
// independently asserts momentum state. Under the engine's no-pyramiding
// constraint, this produces one entry per regime transition in practice.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n < s.Lookback() {
		return model.SignalHold
	}

	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	roc := talib.Roc(closes, s.lookback)
	curr := roc[n-1]

	switch {
	case curr > s.threshold:
		return model.SignalBuy
	case curr < -s.threshold:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
