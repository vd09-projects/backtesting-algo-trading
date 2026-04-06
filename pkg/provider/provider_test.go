package provider_test

import (
	"context"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// stubProvider is a minimal no-op implementation used only to verify the interface
// signature at compile time. It is not a real provider.
type stubProvider struct{}

func (s *stubProvider) FetchCandles(_ context.Context, _ string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	return nil, nil
}

func (s *stubProvider) SupportedTimeframes() []model.Timeframe {
	return nil
}

// Compile-time assertion: stubProvider must satisfy DataProvider.
var _ provider.DataProvider = (*stubProvider)(nil)
