package zerodha

import (
	"fmt"
	"io"
	"net/http"
)

// doHTTP executes req using client and returns the response body.
// Returns ErrAuthRequired (wrapped) on HTTP 401 or 403.
// Returns a plain error on any other non-200 status.
func doHTTP(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // read-only; close error is non-fatal

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return body, nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, fmt.Errorf("HTTP %d: %w", resp.StatusCode, ErrAuthRequired)
	default:
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
	}
}

// toFloat64 type-asserts v to float64, returning 0 on failure.
// Used when parsing JSON number fields decoded as interface{}.
func toFloat64(v interface{}) float64 {
	f, ok := v.(float64)
	if !ok {
		return 0
	}
	return f
}
