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

// SkippedCandle records a single candle that was skipped during parsing due to
// a validation error. Index is the zero-based position in the raw API response.
//
// **Decision (SkippedCandle carries Index and Reason, not the raw OHLC values) — convention: experimental**
// scope: pkg/provider/zerodha.SkippedCandle
// tags: bad-candle, skip, manifest, TASK-0100
// owner: priya
//
// Index is sufficient to locate the candle in context (the caller logged it from --from).
// Including raw OHLC in the struct would require a model.Candle field or float64 fields —
// both add weight with no benefit: the Reason string already contains the invalid values
// (formatted by model.Candle.Validate's error message).
type SkippedCandle struct {
	Index  int    // zero-based index in the raw API response
	Reason string // validation error message from model.Candle.Validate
}

// ErrBadCandles is returned by FetchCandles when one or more candles in the
// API response failed OHLC validation and were skipped. The valid candles are
// still returned alongside this error; callers must use errors.As to detect it
// and treat the fetch as a warning rather than a failure.
//
// This is a non-fatal typed warning — callers that do not check errors.As will
// see a non-nil error and treat the fetch as failed (safe default). Only callers
// that explicitly handle bad-candle skipping (e.g. cmd/fetch-history) should
// treat this as a success.
//
// **Decision (ErrBadCandles as non-fatal typed warning returned alongside valid candles) — architecture: experimental**
// scope: pkg/provider/zerodha.Provider.FetchCandles
// tags: bad-candle, skip, OHLC-validation, Zerodha-artifact, TASK-0100
// owner: priya
//
// The DataProvider interface returns ([]model.Candle, error). A non-nil error
// with a non-nil candle slice is unusual — but the alternative (a separate
// notification channel, or a changed interface signature) would require touching
// every DataProvider implementation and every caller. The errors.As pattern is
// the standard Go idiom for typed error inspection; callers that don't opt in
// see a hard error (fail-safe). CachedProvider handles this by caching the
// candles and propagating the warning.
type ErrBadCandles struct {
	Instrument string
	Skipped    []SkippedCandle
}

func (e *ErrBadCandles) Error() string {
	return fmt.Sprintf(
		"zerodha: %s: skipped %d bad candle(s) due to OHLC validation errors",
		e.Instrument,
		len(e.Skipped),
	)
}
