// Package testutil provides shared test helpers for strategy packages.
package testutil

import (
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

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
