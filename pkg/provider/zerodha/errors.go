// Package zerodha implements the DataProvider interface using the Kite Connect API.
package zerodha

import "errors"

// Sentinel errors returned by the Zerodha provider. All are compatible with errors.Is.
var (
	// ErrAuthRequired is returned when the access token is missing, expired, or rejected by the API.
	ErrAuthRequired = errors.New("zerodha: auth required — run the login flow to get a fresh access token")

	// ErrInstrumentNotFound is returned when the requested instrument is not in the instruments map.
	ErrInstrumentNotFound = errors.New("zerodha: instrument not found in instruments map")

	// ErrUnsupportedTimeframe is returned when the requested timeframe has no Kite Connect interval mapping.
	ErrUnsupportedTimeframe = errors.New("zerodha: timeframe not supported by Kite Connect")
)
