package zerodha

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// newTestServer builds an httptest.Server that routes:
//   - GET /instruments            → testdata/instruments.csv
//   - GET /instruments/historical/... → the provided candleJSON body
//
// requestCount is incremented on each historical endpoint hit.
func newTestServer(t *testing.T, candleJSON []byte, requestCount *atomic.Int32) *httptest.Server {
	t.Helper()
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read instruments testdata: %v", err)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/instruments":
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write(csvData)
		case len(r.URL.Path) > len("/instruments/historical/"):
			if requestCount != nil {
				requestCount.Add(1)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(candleJSON)
		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))
}

func TestNewProvider_success(t *testing.T) {
	srv := newTestServer(t, []byte(`{"data":{"candles":[]}}`), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:      "key",
		AccessToken: "token",
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
	})
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if len(p.tokens) == 0 {
		t.Error("tokens map is empty after construction")
	}
	if p.tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", p.tokens["NSE:NIFTY 50"])
	}
}

func TestNewProvider_empty_api_key(t *testing.T) {
	_, err := NewProvider(t.Context(), Config{AccessToken: "tok"})
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}

func TestNewProvider_empty_access_token(t *testing.T) {
	_, err := NewProvider(t.Context(), Config{APIKey: "key"})
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}

func TestNewProvider_instruments_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
	})
	if err == nil {
		t.Fatal("want error from instruments download failure, got nil")
	}
}

func TestSupportedTimeframes(t *testing.T) {
	srv := newTestServer(t, []byte(`{"data":{"candles":[]}}`), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tfs := p.SupportedTimeframes()
	if len(tfs) != 4 {
		t.Errorf("want 4 supported timeframes, got %d", len(tfs))
	}
	for _, tf := range tfs {
		if tf == model.TimeframeWeekly {
			t.Error("TimeframeWeekly must not be in SupportedTimeframes")
		}
	}
}

// TestFetchCandles_daily_fixture is the integration test that uses the recorded
// API response in testdata/candles_daily.json.
func TestFetchCandles_daily_fixture(t *testing.T) {
	fixtureJSON, err := os.ReadFile(filepath.Join("testdata", "candles_daily.json"))
	if err != nil {
		t.Fatalf("read candles fixture: %v", err)
	}

	srv := newTestServer(t, fixtureJSON, nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
		Sleep: func(time.Duration) {}, // no-op in tests
	})
	if err != nil {
		t.Fatal(err)
	}

	from := time.Date(2026, 3, 29, 18, 30, 0, 0, time.UTC)
	to := time.Date(2026, 4, 7, 18, 30, 0, 0, time.UTC)

	candles, err := p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}

	if len(candles) != 5 {
		t.Fatalf("want 5 candles, got %d", len(candles))
	}

	// All candles carry the instrument identifier.
	for i, c := range candles {
		if c.Instrument != "NSE:NIFTY 50" {
			t.Errorf("candle[%d].Instrument = %q, want %q", i, c.Instrument, "NSE:NIFTY 50")
		}
		if c.Timeframe != model.TimeframeDaily {
			t.Errorf("candle[%d].Timeframe = %q, want %q", i, c.Timeframe, model.TimeframeDaily)
		}
		if c.Timestamp.Location() != time.UTC {
			t.Errorf("candle[%d].Timestamp not UTC: %v", i, c.Timestamp)
		}
	}

	// Spot-check first candle against fixture values.
	first := candles[0]
	// Fixture: ["2026-03-30T00:00:00+0530", 22549.65, 22714.1, 22283.85, 22331.4, 0]
	// UTC: 2026-03-29T18:30:00Z
	wantTS := time.Date(2026, 3, 29, 18, 30, 0, 0, time.UTC)
	if !first.Timestamp.Equal(wantTS) {
		t.Errorf("first candle timestamp: want %v, got %v", wantTS, first.Timestamp)
	}
	if first.Open != 22549.65 {
		t.Errorf("first candle Open: want 22549.65, got %f", first.Open)
	}
	if first.Close != 22331.4 {
		t.Errorf("first candle Close: want 22331.4, got %f", first.Close)
	}
}

func TestFetchCandles_instrument_not_found(t *testing.T) {
	srv := newTestServer(t, []byte(`{"data":{"candles":[]}}`), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.FetchCandles(t.Context(), "NSE:NOTEXIST", model.TimeframeDaily,
		time.Now().Add(-24*time.Hour), time.Now())
	if !errors.Is(err, ErrInstrumentNotFound) {
		t.Errorf("want ErrInstrumentNotFound, got %v", err)
	}
}

func TestFetchCandles_unsupported_timeframe(t *testing.T) {
	srv := newTestServer(t, []byte(`{"data":{"candles":[]}}`), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeWeekly,
		time.Now().Add(-24*time.Hour), time.Now())
	if !errors.Is(err, ErrUnsupportedTimeframe) {
		t.Errorf("want ErrUnsupportedTimeframe, got %v", err)
	}
}

func TestFetchCandles_auth_error(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read instruments testdata: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instruments" {
			_, _ = w.Write(csvData)
			return
		}
		// Historical endpoint returns 401
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL: srv.URL, HTTPClient: srv.Client(),
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily,
		time.Now().Add(-24*time.Hour), time.Now())
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}

func TestFetchCandles_chunked_makes_multiple_requests(t *testing.T) {
	// Range wider than maxDaysPerInterval[TimeframeDaily]=1800 days → 2 chunks.
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC) // ~2191 days

	// The completeness check requires ≥ 90% of ~1565 expected weekday candles.
	// With 2 chunks, each chunk must return ≥ 744. Use 800 per chunk so the
	// merged total (1600) comfortably passes the ≥1487 threshold.
	// buildCandleJSON is defined in instruments_cache_test.go (same package).
	chunkResponse := buildCandleJSON(800)

	var count atomic.Int32
	srv := newTestServer(t, chunkResponse, &count)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		Sleep:      func(time.Duration) {}, // no-op
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}

	if count.Load() != 2 {
		t.Errorf("want 2 HTTP requests for 2 chunks, got %d", count.Load())
	}
}

func TestFetchCandles_context_canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())

	// Cancel after the first chunk is served.
	instrumentsCSV, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read instruments testdata: %v", err)
	}
	var requestsDone atomic.Int32
	cancelSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instruments" {
			_, _ = w.Write(instrumentsCSV)
			return
		}
		n := requestsDone.Add(1)
		if n == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"candles":[]}}`))
			cancel() // cancel after first chunk
		} else {
			http.Error(w, "should not reach", http.StatusInternalServerError)
		}
	}))
	defer cancelSrv.Close()

	// Re-construct with the canceling server.
	p2, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL:    cancelSrv.URL,
		HTTPClient: cancelSrv.Client(),
		Sleep:      func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 2-chunk range; cancel fires after first chunk.
	from := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)

	_, err = p2.FetchCandles(ctx, "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("want context.Canceled, got %v", err)
	}
	if requestsDone.Load() > 1 {
		t.Errorf("second request was made after context cancellation")
	}
}

func TestParseKiteCandles_fixture(t *testing.T) {
	fixtureJSON, err := os.ReadFile(filepath.Join("testdata", "candles_daily.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	candles, skipped, err := parseKiteCandles("NSE:NIFTY 50", model.TimeframeDaily, fixtureJSON)
	if err != nil {
		t.Fatalf("parseKiteCandles: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("want 0 skipped candles from valid fixture, got %d", len(skipped))
	}
	if len(candles) != 5 {
		t.Fatalf("want 5 candles, got %d", len(candles))
	}

	// Volume=0 is valid for NIFTY 50 index.
	for i, c := range candles {
		if c.Volume != 0 {
			t.Errorf("candle[%d] Volume: want 0, got %f", i, c.Volume)
		}
		if c.Timestamp.Location() != time.UTC {
			t.Errorf("candle[%d] timestamp not UTC", i)
		}
	}
}

func TestParseKiteCandles_bad_timestamp(t *testing.T) {
	body := []byte(`{"data":{"candles":[["not-a-time",100,110,90,105,1000]]}}`)
	_, _, err := parseKiteCandles("NSE:RELIANCE", model.TimeframeDaily, body)
	if err == nil {
		t.Fatal("want error for bad timestamp, got nil")
	}
}

func TestParseKiteCandles_short_row(t *testing.T) {
	body := []byte(`{"data":{"candles":[["2024-01-01T00:00:00+0530",100,110]]}}`)
	_, _, err := parseKiteCandles("NSE:RELIANCE", model.TimeframeDaily, body)
	if err == nil {
		t.Fatal("want error for row with <6 elements, got nil")
	}
}

func TestParseKiteCandles_rfc3339_fallback(t *testing.T) {
	// RFC 3339 uses "+05:30" (with colon); the primary layout expects "+0530" (no colon).
	// This tests the fallback path for future API changes.
	body := []byte(`{"data":{"candles":[["2026-03-30T00:00:00+05:30",22549.65,22714.1,22283.85,22331.4,0]]}}`)
	candles, skipped, err := parseKiteCandles("NSE:NIFTY 50", model.TimeframeDaily, body)
	if err != nil {
		t.Fatalf("parseKiteCandles with RFC3339 timestamp: %v", err)
	}
	if len(skipped) != 0 {
		t.Errorf("want 0 skipped candles, got %d", len(skipped))
	}
	if len(candles) != 1 {
		t.Fatalf("want 1 candle, got %d", len(candles))
	}
	wantTS := time.Date(2026, 3, 29, 18, 30, 0, 0, time.UTC)
	if !candles[0].Timestamp.Equal(wantTS) {
		t.Errorf("timestamp: want %v, got %v", wantTS, candles[0].Timestamp)
	}
}

func TestParseKiteCandles_non_float_value_skips_candle(t *testing.T) {
	// When a numeric field is a JSON string, toFloat64 returns 0.
	// Open=0 fails model.Candle validation; the candle must be skipped and
	// collected into ErrBadCandles rather than hard-failing the entire fetch.
	body := []byte(`{"data":{"candles":[["2026-03-30T00:00:00+0530","notanumber",22714.1,22283.85,22331.4,0]]}}`)
	candles, skipped, err := parseKiteCandles("NSE:NIFTY 50", model.TimeframeDaily, body)
	if err != nil {
		t.Fatalf("want nil error (bad candle should be skipped), got %v", err)
	}
	if len(candles) != 0 {
		t.Errorf("want 0 valid candles (only bad candle in body), got %d", len(candles))
	}
	if len(skipped) != 1 {
		t.Fatalf("want 1 skipped candle, got %d", len(skipped))
	}
	if skipped[0].Index != 0 {
		t.Errorf("skipped[0].Index = %d, want 0", skipped[0].Index)
	}
	if skipped[0].Reason == "" {
		t.Error("skipped[0].Reason must not be empty")
	}
}

func TestParseKiteCandles_bad_ohlc_skips_candle(t *testing.T) {
	// Simulate the Zerodha tick-vs-aggregation artifact: open (835.6) < low (837.4).
	// This is the real error seen at candle[450] for HDFCBANK/ICICIBANK.
	// Bad candle must be skipped; valid candles before and after must be returned.
	body := []byte(`{"data":{"candles":[
		["2021-01-08T09:15:00+0530",100,110,90,105,1000],
		["2021-01-08T09:20:00+0530",835.6,843.8,837.4,840.0,5000],
		["2021-01-08T09:25:00+0530",841.0,845.0,839.0,843.0,4500]
	]}}`)
	candles, skipped, err := parseKiteCandles("NSE:HDFCBANK", model.Timeframe5Min, body)
	if err != nil {
		t.Fatalf("want nil error (bad candle should be skipped), got %v", err)
	}
	if len(candles) != 2 {
		t.Fatalf("want 2 valid candles (middle one is bad), got %d", len(candles))
	}
	if len(skipped) != 1 {
		t.Fatalf("want 1 skipped candle, got %d", len(skipped))
	}
	if skipped[0].Index != 1 {
		t.Errorf("skipped[0].Index = %d, want 1", skipped[0].Index)
	}
	if skipped[0].Reason == "" {
		t.Error("skipped[0].Reason must not be empty")
	}
}

func TestFetchCandles_bad_ohlc_returns_ErrBadCandles_with_valid_candles(t *testing.T) {
	// One bad candle (open < low) plus one valid candle in the response.
	// Date range: Mon 2021-01-11 → Tue 2021-01-12 (1 weekday).
	// weekdayCount = 1, candlesPerDay (daily) = 1, expected = 1, threshold = int(0.9*1) = 0.
	// Got = 1 valid (bad one skipped). 1 >= 0 → completeness passes.
	// FetchCandles must return: 1 valid candle + *ErrBadCandles{Skipped: 1 entry}.
	mixedCandleBody := `{"data":{"candles":[
		["2021-01-11T00:00:00+0530",835.6,843.8,837.4,840.0,5000],
		["2021-01-12T00:00:00+0530",841.0,845.0,839.0,843.0,4500]
	]}}`
	// First candle has open=835.6 < low=837.4 → OHLC validation error → skipped.
	// Second candle is valid.

	srv := newTestServer(t, []byte(mixedCandleBody), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		Sleep:      func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	// 1-weekday range: Mon 2021-01-11 to Tue 2021-01-12.
	from := time.Date(2021, 1, 11, 0, 0, 0, 0, time.UTC)
	to := time.Date(2021, 1, 12, 0, 0, 0, 0, time.UTC)
	candles, fetchErr := p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)

	// Must return a non-nil *ErrBadCandles warning.
	var badCandles *ErrBadCandles
	if !errors.As(fetchErr, &badCandles) {
		t.Fatalf("want *ErrBadCandles from FetchCandles, got %T: %v", fetchErr, fetchErr)
	}
	if len(badCandles.Skipped) != 1 {
		t.Errorf("want 1 skipped candle in ErrBadCandles, got %d", len(badCandles.Skipped))
	}
	if badCandles.Instrument != "NSE:NIFTY 50" {
		t.Errorf("ErrBadCandles.Instrument = %q, want %q", badCandles.Instrument, "NSE:NIFTY 50")
	}
	if badCandles.Skipped[0].Index != 0 {
		t.Errorf("skipped[0].Index = %d, want 0 (first candle in response)", badCandles.Skipped[0].Index)
	}
	// Valid candles: second candle (open=841, within [low=839, high=845]) is returned.
	if len(candles) != 1 {
		t.Errorf("want 1 valid candle, got %d", len(candles))
	}
}

func TestFetchCandles_all_valid_returns_nil_error(t *testing.T) {
	// Regression: when all candles are valid, FetchCandles must not return ErrBadCandles.
	// Use the existing daily fixture which has 5 valid candles and no bad ones.
	fixtureJSON, err := os.ReadFile(filepath.Join("testdata", "candles_daily.json"))
	if err != nil {
		t.Fatalf("read candles fixture: %v", err)
	}

	srv := newTestServer(t, fixtureJSON, nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey: "key", AccessToken: "tok",
		BaseURL:    srv.URL,
		HTTPClient: srv.Client(),
		Sleep:      func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	from := time.Date(2026, 3, 29, 18, 30, 0, 0, time.UTC)
	to := time.Date(2026, 4, 7, 18, 30, 0, 0, time.UTC)

	_, fetchErr := p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	if fetchErr != nil {
		t.Errorf("want nil error for all-valid candles, got %v", fetchErr)
	}
}
