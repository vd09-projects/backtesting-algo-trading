package provider

import (
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// DataProvider fetches historical candle data for a given instrument and timeframe.
type DataProvider interface {
	// FetchCandles returns candles for the given instrument and timeframe over [from, to).
	FetchCandles(instrument string, timeframe model.Timeframe, from, to time.Time) ([]model.Candle, error)

	// SupportedTimeframes returns the list of timeframes this provider can serve.
	SupportedTimeframes() []model.Timeframe
}
