package zerodha

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// loadInstrumentsCSV downloads the Kite Connect instruments master and returns
// a map keyed by "EXCHANGE:TRADINGSYMBOL" → instrument_token (int64).
// One HTTP call per process start; callers hold the result for the process lifetime.
func loadInstrumentsCSV(ctx context.Context, client *http.Client, baseURL, apiKey, accessToken string) (map[string]int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/instruments", http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build instruments request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", apiKey, accessToken))
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doHTTP(client, req)
	if err != nil {
		return nil, fmt.Errorf("download instruments: %w", err)
	}

	tokens, err := parseInstrumentsCSV(body)
	if err != nil {
		return nil, fmt.Errorf("parse instruments: %w", err)
	}
	return tokens, nil
}

// parseInstrumentsCSV parses the Kite Connect instruments CSV and returns a
// map keyed by "EXCHANGE:TRADINGSYMBOL" → instrument_token.
//
// The CSV must have a header row containing at least the columns:
// instrument_token, tradingsymbol, exchange.
// Rows where instrument_token is not a valid integer are skipped.
func parseInstrumentsCSV(data []byte) (map[string]int64, error) {
	r := csv.NewReader(bytes.NewReader(data))
	r.FieldsPerRecord = -1 // allow variable column count

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read csv header: %w", err)
	}

	// Build column index from header names for resilience against column reordering.
	colIdx := make(map[string]int, len(header))
	for i, h := range header {
		colIdx[h] = i
	}

	tokenCol, ok1 := colIdx["instrument_token"]
	exchangeCol, ok2 := colIdx["exchange"]
	symbolCol, ok3 := colIdx["tradingsymbol"]
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("instruments csv: missing required columns (instrument_token, exchange, tradingsymbol); got header: %v", header)
	}

	tokens := make(map[string]int64)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read csv row: %w", err)
		}

		maxCol := tokenCol
		if exchangeCol > maxCol {
			maxCol = exchangeCol
		}
		if symbolCol > maxCol {
			maxCol = symbolCol
		}
		if len(row) <= maxCol {
			continue // skip malformed rows silently
		}

		tok, err := strconv.ParseInt(row[tokenCol], 10, 64)
		if err != nil {
			continue // skip rows with non-numeric instrument_token
		}

		key := row[exchangeCol] + ":" + row[symbolCol]
		tokens[key] = tok
	}
	return tokens, nil
}
