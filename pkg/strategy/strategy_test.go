package strategy_test

import (
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// stubStrategy is a minimal no-op implementation used only to verify the interface
// signature at compile time. It is not a real strategy.
type stubStrategy struct{}

func (s *stubStrategy) Name() string               { return "stub" }
func (s *stubStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *stubStrategy) Lookback() int              { return 1 }
func (s *stubStrategy) Next(_ []model.Candle) model.Signal {
	return model.SignalHold
}

// Compile-time assertion: stubStrategy must satisfy Strategy.
var _ strategy.Strategy = (*stubStrategy)(nil)
