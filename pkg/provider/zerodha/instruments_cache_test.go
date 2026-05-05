package zerodha

import (
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

// ── loadOrCacheInstrumentsCSV ─────────────────────────────────────────────────

// TestLoadOrCacheInstrumentsCSV_freshCacheSkipsNetwork verifies that when a
// fresh instruments CSV already exists on disk, no network call is made and
// the cached contents are returned.
func TestLoadOrCacheInstrumentsCSV_freshCacheSkipsNetwork(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	dir := t.TempDir()
	cachePath := filepath.Join(dir, "instruments.csv")
	if err := os.WriteFile(cachePath, csvData, 0o644); err != nil {
		t.Fatalf("write cache file: %v", err)
	}
	// File is freshly written — well within 24h TTL.

	var networkCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		networkCalls.Add(1)
		w.WriteHeader(http.StatusInternalServerError) // should never be reached
	}))
	defer srv.Close()

	tokens, err := loadOrCacheInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "", "", dir)
	if err != nil {
		t.Fatalf("loadOrCacheInstrumentsCSV: %v", err)
	}
	if networkCalls.Load() != 0 {
		t.Errorf("network was called %d times; want 0 (cache hit)", networkCalls.Load())
	}
	if tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", tokens["NSE:NIFTY 50"])
	}
}

// TestLoadOrCacheInstrumentsCSV_staleCacheFetchesAndUpdates verifies that when
// the cached CSV is older than 24h, a network fetch is made and the cache file
// is overwritten with fresh data.
func TestLoadOrCacheInstrumentsCSV_staleCacheFetchesAndUpdates(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	dir := t.TempDir()
	cachePath := filepath.Join(dir, "instruments.csv")
	// Write a stale file (mtime 25h ago).
	if err := os.WriteFile(cachePath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale cache: %v", err)
	}
	staleTime := time.Now().Add(-25 * time.Hour)
	if err := os.Chtimes(cachePath, staleTime, staleTime); err != nil {
		t.Fatalf("set stale mtime: %v", err)
	}

	var networkCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		networkCalls.Add(1)
		if r.URL.Path == "/instruments" {
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write(csvData)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	tokens, err := loadOrCacheInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "key", "token", dir)
	if err != nil {
		t.Fatalf("loadOrCacheInstrumentsCSV: %v", err)
	}
	if networkCalls.Load() != 1 {
		t.Errorf("network calls = %d, want 1", networkCalls.Load())
	}
	if tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", tokens["NSE:NIFTY 50"])
	}

	// Cache file must be updated with fresh content.
	updated, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatalf("read updated cache: %v", err)
	}
	if string(updated) == "stale" {
		t.Error("cache file not updated after stale fetch")
	}
}

// TestLoadOrCacheInstrumentsCSV_noCacheNoToken verifies that when no cache
// exists and the network returns 401, ErrAuthRequired is returned.
func TestLoadOrCacheInstrumentsCSV_noCacheNoToken(t *testing.T) {
	dir := t.TempDir() // empty — no instruments.csv

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := loadOrCacheInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "key", "badtoken", dir)
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}

// TestLoadOrCacheInstrumentsCSV_noCacheFreshFetch verifies that when no cache
// exists, a network fetch succeeds and the result is written to disk.
func TestLoadOrCacheInstrumentsCSV_noCacheFreshFetch(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	dir := t.TempDir()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instruments" {
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write(csvData)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	tokens, err := loadOrCacheInstrumentsCSV(t.Context(), srv.Client(), srv.URL, "key", "token", dir)
	if err != nil {
		t.Fatalf("loadOrCacheInstrumentsCSV: %v", err)
	}
	if tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", tokens["NSE:NIFTY 50"])
	}

	// Cache must be written.
	if _, statErr := os.Stat(filepath.Join(dir, "instruments.csv")); statErr != nil {
		t.Errorf("cache file not written after fresh fetch: %v", statErr)
	}
}

// ── NewProvider with InstrumentsCacheDir ──────────────────────────────────────

// TestNewProvider_cachedInstruments_noTokenRequired verifies that NewProvider
// succeeds without a live network call when a fresh instruments CSV is cached.
func TestNewProvider_cachedInstruments_noTokenRequired(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	dir := t.TempDir()
	cachePath := filepath.Join(dir, "instruments.csv")
	if err := os.WriteFile(cachePath, csvData, 0o644); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	// Server returns 401 for all requests — simulating expired token.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:              "key",
		AccessToken:         "expired-token",
		BaseURL:             srv.URL,
		HTTPClient:          srv.Client(),
		InstrumentsCacheDir: dir,
	})
	if err != nil {
		t.Fatalf("NewProvider with fresh instruments cache: %v", err)
	}
	if p.tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", p.tokens["NSE:NIFTY 50"])
	}
}

// TestNewProvider_noCacheDir_requiresNetwork verifies that NewProvider without
// an InstrumentsCacheDir still makes a network call (existing behavior preserved).
func TestNewProvider_noCacheDir_requiresNetwork(t *testing.T) {
	csvData, err := os.ReadFile(filepath.Join("testdata", "instruments.csv"))
	if err != nil {
		t.Fatalf("read testdata: %v", err)
	}

	var networkCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/instruments" {
			networkCalls.Add(1)
			w.Header().Set("Content-Type", "text/csv")
			_, _ = w.Write(csvData)
		}
	}))
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:      "key",
		AccessToken: "token",
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		// InstrumentsCacheDir intentionally empty — no caching
	})
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	if networkCalls.Load() != 1 {
		t.Errorf("network calls = %d, want 1 (no cache dir)", networkCalls.Load())
	}
	if p.tokens["NSE:NIFTY 50"] != 256265 {
		t.Errorf("tokens[NSE:NIFTY 50] = %d, want 256265", p.tokens["NSE:NIFTY 50"])
	}
}

// ── ErrIncompleteData / chunk completeness ────────────────────────────────────

// TestFetchCandles_incompleteData_returnsTypedError verifies that when the
// merged candle slice is far below the expected weekday count, FetchCandles
// returns *ErrIncompleteData (not a generic error).
func TestFetchCandles_incompleteData_returnsTypedError(t *testing.T) {
	// 10 candles served, but a 365-day daily range expects ~261 trading days.
	// 10/261 ≈ 3.8% — well below the 95% threshold.
	tinyCandles := buildCandleJSON(10)

	srv := newTestServer(t, tinyCandles, nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:      "key",
		AccessToken: "token",
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		Sleep:       func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC) // ~261 trading days

	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	var incErr *ErrIncompleteData
	if !errors.As(err, &incErr) {
		t.Fatalf("want *ErrIncompleteData, got %T: %v", err, err)
	}
	if incErr.Instrument != "NSE:NIFTY 50" {
		t.Errorf("Instrument = %q, want %q", incErr.Instrument, "NSE:NIFTY 50")
	}
	if incErr.Got != 10 {
		t.Errorf("Got = %d, want 10", incErr.Got)
	}
	if incErr.Expected <= 0 {
		t.Errorf("Expected = %d, want > 0", incErr.Expected)
	}
}

// TestFetchCandles_sufficientData_noError verifies that when candle count meets
// the 90% threshold no ErrIncompleteData is returned.
func TestFetchCandles_sufficientData_noError(t *testing.T) {
	// 5-day Mon–Fri range → 5 weekdays expected. Serve 5 candles.
	fiveCandles := buildCandleJSON(5)
	srv := newTestServer(t, fiveCandles, nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:      "key",
		AccessToken: "token",
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		Sleep:       func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	from := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC) // Monday
	to := time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC)    // Saturday → 5 weekdays

	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, from, to)
	if err != nil {
		t.Fatalf("FetchCandles with sufficient candles: unexpected error: %v", err)
	}
}

// TestFetchCandles_emptyRange_noCompletenessCheck verifies that a zero-length
// date range (from == to) does not trigger ErrIncompleteData.
func TestFetchCandles_emptyRange_noCompletenessCheck(t *testing.T) {
	srv := newTestServer(t, []byte(`{"data":{"candles":[]}}`), nil)
	defer srv.Close()

	p, err := NewProvider(t.Context(), Config{
		APIKey:      "key",
		AccessToken: "token",
		BaseURL:     srv.URL,
		HTTPClient:  srv.Client(),
		Sleep:       func(time.Duration) {},
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	_, err = p.FetchCandles(t.Context(), "NSE:NIFTY 50", model.TimeframeDaily, now, now)
	if err != nil {
		t.Fatalf("empty range: unexpected error: %v", err)
	}
}

// TestErrIncompleteData_errorInterface verifies the typed error satisfies the
// error interface and produces a human-readable message containing key fields.
func TestErrIncompleteData_errorInterface(t *testing.T) {
	e := &ErrIncompleteData{
		Instrument: "NSE:INFY",
		From:       time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		To:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Expected:   261,
		Got:        10,
	}
	msg := e.Error()
	for _, want := range []string{"NSE:INFY", "261", "10"} {
		if !containsStr(msg, want) {
			t.Errorf("Error() = %q, missing %q", msg, want)
		}
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// buildCandleJSON builds a minimal Kite candle JSON response with n candles.
// Timestamps are sequential IST days starting from 2023-01-02 (Monday).
func buildCandleJSON(n int) []byte {
	ist := time.FixedZone("IST", 5*3600+30*60)
	base := time.Date(2023, 1, 2, 9, 15, 0, 0, ist)
	var rows []byte
	for i := 0; i < n; i++ {
		ts := base.Add(time.Duration(i) * 24 * time.Hour)
		if i > 0 {
			rows = append(rows, ',')
		}
		tsStr := ts.Format("2006-01-02T15:04:05-0700")
		row := `["` + tsStr + `",100,110,90,105,1000]`
		rows = append(rows, []byte(row)...)
	}
	return []byte(`{"data":{"candles":[` + string(rows) + `]}}`)
}

func containsStr(s, sub string) bool {
	if sub == "" {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
