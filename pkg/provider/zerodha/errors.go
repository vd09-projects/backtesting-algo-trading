// Package zerodha implements the DataProvider interface using the Kite Connect API.
package zerodha

import (
	"errors"
	"fmt"
	"time"
)

// Sentinel errors returned by the Zerodha provider. All are compatible with errors.Is.
var (
	// ErrAuthRequired is returned when the access token is missing, expired, or rejected by the API.
	ErrAuthRequired = errors.New("zerodha: auth required — run the login flow to get a fresh access token")

	// ErrInstrumentNotFound is returned when the requested instrument is not in the instruments map.
	ErrInstrumentNotFound = errors.New("zerodha: instrument not found in instruments map")

	// ErrUnsupportedTimeframe is returned when the requested timeframe has no Kite Connect interval mapping.
	ErrUnsupportedTimeframe = errors.New("zerodha: timeframe not supported by Kite Connect")
)

// ErrIncompleteData is returned by FetchCandles when the merged candle slice
// across all chunked requests is below 95% of the expected weekday count for
// the requested date range. This distinguishes "partial data returned" from
// "instrument not found" or "no data for this period".
//
// Callers can inspect Expected and Got to decide whether to proceed with the
// available data or treat the result as an error.
type ErrIncompleteData struct {
	Instrument string
	From       time.Time
	To         time.Time
	Expected   int // weekday-based estimate × candlesPerDay
	Got        int // actual candles returned
}

func (e *ErrIncompleteData) Error() string {
	return fmt.Sprintf(
		"zerodha: incomplete data for %s [%s, %s): expected ~%d candles, got %d",
		e.Instrument,
		e.From.Format("2006-01-02"),
		e.To.Format("2006-01-02"),
		e.Expected,
		e.Got,
	)
}
