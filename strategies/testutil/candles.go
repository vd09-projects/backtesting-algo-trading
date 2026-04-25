// Package testutil provides shared test helpers for strategy packages.
package testutil

import (
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// MakeCandlesOHLC builds a slice of daily candles with explicit High, Low, and
// Close values. Open is set equal to Close (always within [Low, High]).
// All three slices must have equal length; the caller is responsible for
// ensuring each close is within [low, high] for that bar.
// Timestamps start at 2024-01-01 and increment by one calendar day.
func MakeCandlesOHLC(highs, lows, closes []float64) []model.Candle {
	if len(highs) != len(lows) || len(highs) != len(closes) {
		panic("MakeCandlesOHLC: highs, lows, closes must have equal length")
	}
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]model.Candle, len(closes))
	for i := range candles {
		candles[i] = model.Candle{
			Instrument: "TEST:X",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       closes[i],
			High:       highs[i],
			Low:        lows[i],
			Close:      closes[i],
			Volume:     1000,
		}
	}
	return candles
}

// MakeCandles builds a slice of daily candles from a close-price series.
// All OHLC fields are set to the close value so validation passes.
// Timestamps start at 2024-01-01 and increment by one calendar day.
func MakeCandles(closes []float64) []model.Candle {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]model.Candle, len(closes))
	for i, c := range closes {
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
