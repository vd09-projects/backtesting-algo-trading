// Package strategy defines the Strategy interface that all trading strategies must implement.
package strategy

import "github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"

// Strategy produces a trading signal given the candles seen so far.
// It is called once per bar with all candles up to and including the current bar.
type Strategy interface {
	// Name returns a human-readable identifier for this strategy.
	Name() string

	// Timeframe declares which candle timeframe this strategy operates on.
	Timeframe() model.Timeframe

	// Lookback declares how many historical candles are required before the
	// strategy can emit a meaningful signal. The engine will not call Next until
	// at least Lookback() candles are available.
	Lookback() int

	// Next receives all candles seen so far (length >= Lookback()) and returns
	// the signal for the current bar.
	Next(candles []model.Candle) model.Signal
}
