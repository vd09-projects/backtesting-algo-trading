// Package rsimeanrev implements an RSI mean-reversion strategy.
// It emits a Buy signal when RSI falls below the oversold threshold and a
// Sell signal when RSI rises above the overbought threshold. All other bars
// emit Hold. Long-only: no short positions are taken.
package rsimeanrev

import (
	"fmt"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on RSI levels.
type Strategy struct {
	period     int
	oversold   float64
	overbought float64
	timeframe  model.Timeframe
}

// New constructs an RSI mean-reversion Strategy. Returns an error if period
// is less than 1, if oversold is negative, if overbought exceeds 100, or if
// oversold is not strictly less than overbought.
func New(tf model.Timeframe, period int, oversold, overbought float64) (*Strategy, error) {
	if period < 1 {
		return nil, fmt.Errorf("rsimeanrev: period must be >= 1, got %d", period)
	}
	if oversold < 0 {
		return nil, fmt.Errorf("rsimeanrev: oversold must be >= 0, got %g", oversold)
	}
	if overbought > 100 {
		return nil, fmt.Errorf("rsimeanrev: overbought must be <= 100, got %g", overbought)
	}
	if oversold >= overbought {
		return nil, fmt.Errorf("rsimeanrev: oversold (%g) must be < overbought (%g)", oversold, overbought)
	}
	return &Strategy{
		period:     period,
		oversold:   oversold,
		overbought: overbought,
		timeframe:  tf,
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "rsi-mean-reversion" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns period + 1.
// talib RSI requires one extra bar beyond the period to produce its first
// valid output (initial Wilder average needs period differences, which needs
// period+1 closes).
func (s *Strategy) Lookback() int { return s.period + 1 }

// Next returns the signal for the current bar.
// Buy is emitted when RSI < oversold. Sell is emitted when RSI > overbought.
// Hold is returned for all other bars, including before Lookback() bars are
// available.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n < s.Lookback() {
		return model.SignalHold
	}

	closes := make([]float64, n)
	for i, c := range candles {
		closes[i] = c.Close
	}

	rsiVals := talib.Rsi(closes, s.period)
	current := rsiVals[n-1]

	switch {
	case current < s.oversold:
		return model.SignalBuy
	case current > s.overbought:
		return model.SignalSell
	default:
		return model.SignalHold
	}
}
