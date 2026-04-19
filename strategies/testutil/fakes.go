package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// StaticProvider satisfies provider.DataProvider and always returns the same
// candles, ignoring instrument, timeframe, and date range arguments.
// Use it to inject deterministic data without network access.
type StaticProvider struct {
	Candles []model.Candle
}

// FetchCandles returns the fixed candle slice, ignoring all arguments.
func (s *StaticProvider) FetchCandles(
	_ context.Context, _ string, _ model.Timeframe, _, _ time.Time,
) ([]model.Candle, error) {
	return s.Candles, nil
}

// SupportedTimeframes returns daily as the only supported timeframe.
func (s *StaticProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// ThresholdStrategy emits Buy when the most recent close strictly exceeds
// Threshold, and Sell otherwise. Lookback is 1.
type ThresholdStrategy struct {
	Threshold float64
	TF        model.Timeframe
}

// Name returns a human-readable name including the threshold value.
func (t *ThresholdStrategy) Name() string { return fmt.Sprintf("threshold-%.0f", t.Threshold) }

// Timeframe returns the configured timeframe.
func (t *ThresholdStrategy) Timeframe() model.Timeframe { return t.TF }

// Lookback returns 1 — the strategy operates on a single bar.
func (t *ThresholdStrategy) Lookback() int { return 1 }

// Next emits Buy when the most recent close strictly exceeds Threshold, Sell otherwise.
func (t *ThresholdStrategy) Next(candles []model.Candle) model.Signal {
	if candles[len(candles)-1].Close > t.Threshold {
		return model.SignalBuy
	}
	return model.SignalSell
}

// MakeAlternatingCandles returns n daily candles on instrument "TEST:X"
// alternating between highClose and lowClose, starting with highClose on bar 0.
// All OHLC fields equal the close so engine fills at Open are fully deterministic.
func MakeAlternatingCandles(n int, highClose, lowClose float64) []model.Candle {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]model.Candle, n)
	for i := range candles {
		c := highClose
		if i%2 == 1 {
			c = lowClose
		}
		candles[i] = model.Candle{
			Instrument: "TEST:X",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       c,
			High:       c,
			Low:        c,
			Close:      c,
			Volume:     1000,
		}
	}
	return candles
}

// TestEngineConfig returns a minimal engine.Config for sweep tests.
// Instrument is "TEST:X"; date range spans all of 2024.
func TestEngineConfig() engine.Config {
	return engine.Config{
		Instrument:           "TEST:X",
		From:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		InitialCash:          100000,
		PositionSizeFraction: 0.1,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	}
}

// compile-time check: ThresholdStrategy satisfies strategy.Strategy.
var _ strategy.Strategy = (*ThresholdStrategy)(nil)
