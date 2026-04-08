package cache

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// compile-time check that CachedProvider satisfies the DataProvider interface.
var _ provider.DataProvider = (*CachedProvider)(nil)

// mockProvider is a test double for provider.DataProvider.
type mockProvider struct {
	candles    []model.Candle
	err        error
	callCount  int
	timeframes []model.Timeframe
}

func (m *mockProvider) FetchCandles(_ context.Context, _ string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	m.callCount++
	return m.candles, m.err
}

func (m *mockProvider) SupportedTimeframes() []model.Timeframe {
	return m.timeframes
}

// fixedClock returns a function that always returns t.
func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

// mustCandle builds a Candle, panicking on validation error (test helper only).
func mustCandle(instrument string, tf model.Timeframe, ts time.Time) model.Candle {
	c, err := model.NewCandle(instrument, tf, ts, 100, 110, 90, 105, 1000)
	if err != nil {
		panic(err)
	}
	return c
}

var (
	testInstrument = "NSE:NIFTY50"
	testTF         = model.TimeframeDaily
	testFrom       = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	testTo         = time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	testCandles    = []model.Candle{
		mustCandle(testInstrument, testTF, time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)),
		mustCandle(testInstrument, testTF, time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)),
	}
	// testNow is a "today" that makes testTo clearly historical (to < today).
	testNow = time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)
)

func newTestProvider(t *testing.T, inner *mockProvider) *CachedProvider {
	t.Helper()
	dir := t.TempDir()
	cp := NewCachedProvider(inner, dir)
	cp.now = fixedClock(testNow)
	return cp
}

// TestCacheMiss verifies that a cache miss fetches from the inner provider,
// writes a cache file, and returns the candles.
func TestCacheMiss(t *testing.T) {
	inner := &mockProvider{candles: testCandles}
	cp := newTestProvider(t, inner)

	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, testTo)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("inner call count = %d, want 1", inner.callCount)
	}
	if len(got) != len(testCandles) {
		t.Errorf("got %d candles, want %d", len(got), len(testCandles))
	}

	// Cache file must exist after a miss.
	path := cp.cachePath(testInstrument, testTF, testFrom, testTo)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("cache file not written after miss")
	}
}

// TestCacheHit_Historical verifies that a second fetch for historical data
// is served from cache without hitting the inner provider.
func TestCacheHit_Historical(t *testing.T) {
	inner := &mockProvider{candles: testCandles}
	cp := newTestProvider(t, inner)

	// First call — cache miss, populates the file.
	if _, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, testTo); err != nil {
		t.Fatalf("first FetchCandles: %v", err)
	}

	// Second call — must hit cache.
	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, testTo)
	if err != nil {
		t.Fatalf("second FetchCandles: %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("inner call count = %d after cache hit, want 1", inner.callCount)
	}
	if len(got) != len(testCandles) {
		t.Errorf("got %d candles, want %d", len(got), len(testCandles))
	}
}

// TestCacheHit_Recent_WithinTTL verifies that recent data (to >= today) within
// the 1-hour TTL is served from cache.
func TestCacheHit_Recent_WithinTTL(t *testing.T) {
	// to is "today" — recent data subject to TTL.
	recentTo := testNow.Truncate(24 * time.Hour) // 2026-04-09

	inner := &mockProvider{candles: testCandles}
	cp := newTestProvider(t, inner)

	// Prime the cache.
	if _, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, recentTo); err != nil {
		t.Fatalf("prime FetchCandles: %v", err)
	}

	// Advance clock by 30 minutes — still within TTL.
	cp.now = fixedClock(testNow.Add(30 * time.Minute))

	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, recentTo)
	if err != nil {
		t.Fatalf("second FetchCandles: %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("inner call count = %d, want 1 (cache hit expected)", inner.callCount)
	}
	if len(got) != len(testCandles) {
		t.Errorf("got %d candles, want %d", len(got), len(testCandles))
	}
}

// TestCacheHit_Recent_ExpiredTTL verifies that recent data older than 1 hour
// is invalidated and re-fetched from the inner provider.
func TestCacheHit_Recent_ExpiredTTL(t *testing.T) {
	recentTo := testNow.Truncate(24 * time.Hour) // 2026-04-09

	inner := &mockProvider{candles: testCandles}
	cp := newTestProvider(t, inner)

	// Prime the cache at testNow.
	if _, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, recentTo); err != nil {
		t.Fatalf("prime FetchCandles: %v", err)
	}

	// Advance clock by 2 hours — TTL expired.
	cp.now = fixedClock(testNow.Add(2 * time.Hour))

	if _, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, recentTo); err != nil {
		t.Fatalf("second FetchCandles: %v", err)
	}
	if inner.callCount != 2 {
		t.Errorf("inner call count = %d, want 2 (re-fetch expected after TTL expiry)", inner.callCount)
	}

	// Stale cache file must have been deleted (re-created with new timestamp).
	path := cp.cachePath(testInstrument, testTF, testFrom, recentTo)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cache file missing after re-fetch: %v", err)
	}
	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal cache after re-fetch: %v", err)
	}
	want := testNow.Add(2 * time.Hour).UTC()
	if !entry.CachedAt.Equal(want) {
		t.Errorf("cached_at = %v, want %v", entry.CachedAt, want)
	}
}

// TestCorruptCacheFile verifies that a corrupt cache file is treated as a miss
// and the inner provider is called.
func TestCorruptCacheFile(t *testing.T) {
	inner := &mockProvider{candles: testCandles}
	cp := newTestProvider(t, inner)

	// Write a corrupt cache file manually.
	path := cp.cachePath(testInstrument, testTF, testFrom, testTo)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("not valid json {{{"), 0o644); err != nil {
		t.Fatalf("write corrupt file: %v", err)
	}

	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, testTo)
	if err != nil {
		t.Fatalf("FetchCandles with corrupt cache: %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("inner call count = %d, want 1 (miss expected on corrupt file)", inner.callCount)
	}
	if len(got) != len(testCandles) {
		t.Errorf("got %d candles, want %d", len(got), len(testCandles))
	}
}

// TestInnerError verifies that an error from the inner provider is propagated
// and nothing is written to the cache.
func TestInnerError(t *testing.T) {
	sentinel := errors.New("api down")
	inner := &mockProvider{err: sentinel}
	cp := newTestProvider(t, inner)

	_, err := cp.FetchCandles(context.Background(), testInstrument, testTF, testFrom, testTo)
	if !errors.Is(err, sentinel) {
		t.Fatalf("got err %v, want %v", err, sentinel)
	}

	path := cp.cachePath(testInstrument, testTF, testFrom, testTo)
	if _, statErr := os.Stat(path); !os.IsNotExist(statErr) {
		t.Error("cache file must not be written when inner provider returns an error")
	}
}

// TestSupportedTimeframes verifies delegation to the inner provider.
func TestSupportedTimeframes(t *testing.T) {
	want := []model.Timeframe{model.TimeframeDaily, model.Timeframe1Min}
	inner := &mockProvider{timeframes: want}
	cp := newTestProvider(t, inner)

	got := cp.SupportedTimeframes()
	if len(got) != len(want) {
		t.Fatalf("SupportedTimeframes len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SupportedTimeframes[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestCachePath verifies the derived cache file path format.
func TestCachePath(t *testing.T) {
	cp := NewCachedProvider(nil, "/base")

	got := cp.cachePath(
		"NSE:NIFTY 50",
		model.TimeframeDaily,
		time.Date(2024, 1, 1, 9, 15, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	want := "/base/nse_nifty_50/daily_2024-01-01_2025-01-01.json"
	if got != want {
		t.Errorf("cachePath = %q, want %q", got, want)
	}
}
