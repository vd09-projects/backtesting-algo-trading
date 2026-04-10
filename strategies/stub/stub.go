// Package stub provides a no-op Strategy that always emits SignalHold.
// It exists solely to allow cmd/backtest CLI wiring to be exercised before any
// real strategy is implemented. It must not be used in production backtests.
package stub

import "github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"

// Strategy is a no-op strategy that always holds.
type Strategy struct {
	timeframe model.Timeframe
}

// New returns a stub Strategy for the given timeframe.
func New(tf model.Timeframe) *Strategy {
	return &Strategy{timeframe: tf}
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "stub" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns the minimum candle history required before Next is called.
func (s *Strategy) Lookback() int { return 1 }

// Next always returns SignalHold regardless of candle history.
func (s *Strategy) Next(_ []model.Candle) model.Signal { return model.SignalHold }
