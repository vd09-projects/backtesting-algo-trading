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

// --- Range-aware superset lookup tests ---

// wideCandles builds a synthetic candle series spanning [start, end) in daily steps.
// Timestamps are set to midnight UTC on each day; weekends are included for simplicity
// (the filter is timestamp-based, not calendar-aware).
func wideCandles(instrument string, tf model.Timeframe, start, end time.Time) []model.Candle {
	var out []model.Candle
	for d := start; d.Before(end); d = d.AddDate(0, 0, 1) {
		c, err := model.NewCandle(instrument, tf, d, 100, 110, 90, 105, 1000)
		if err != nil {
			panic(err)
		}
		out = append(out, c)
	}
	return out
}

// TestSupersetHit_Golden is the primary acceptance test: a wide cache file (2018–2024) is
// written manually, then a narrow request (2020–2022) is made. The inner provider must
// receive zero calls, and all returned candles must fall within [2020-01-01, 2022-12-31].
func TestSupersetHit_Golden(t *testing.T) {
	wideFrom := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	wideTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowTo := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC) // exclusive upper bound

	allCandles := wideCandles(testInstrument, testTF, wideFrom, wideTo)

	inner := &mockProvider{candles: allCandles}
	cp := newTestProvider(t, inner)

	// Write the wide cache file directly.
	widePath := cp.cachePath(testInstrument, testTF, wideFrom, wideTo)
	if err := cp.writeCache(widePath, allCandles); err != nil {
		t.Fatalf("writeCache wide: %v", err)
	}

	// Request a narrow range — must be served from the wide file, zero network calls.
	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, narrowFrom, narrowTo)
	if err != nil {
		t.Fatalf("FetchCandles narrow: %v", err)
	}
	if inner.callCount != 0 {
		t.Errorf("inner call count = %d, want 0 (superset hit expected)", inner.callCount)
	}
	if len(got) == 0 {
		t.Fatal("got 0 candles, want candles in [2020-01-01, 2023-01-01)")
	}
	for _, c := range got {
		if c.Timestamp.Before(narrowFrom) {
			t.Errorf("candle timestamp %v is before narrowFrom %v", c.Timestamp, narrowFrom)
		}
		if !c.Timestamp.Before(narrowTo) {
			t.Errorf("candle timestamp %v is not before narrowTo %v", c.Timestamp, narrowTo)
		}
	}
}

// TestSupersetHit_FromEqualCachedStart verifies that from == cachedFrom is treated as a superset hit.
func TestSupersetHit_FromEqualCachedStart(t *testing.T) {
	wideFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	wideTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowTo := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	allCandles := wideCandles(testInstrument, testTF, wideFrom, wideTo)

	inner := &mockProvider{candles: allCandles}
	cp := newTestProvider(t, inner)

	widePath := cp.cachePath(testInstrument, testTF, wideFrom, wideTo)
	if err := cp.writeCache(widePath, allCandles); err != nil {
		t.Fatalf("writeCache: %v", err)
	}

	// from == wideFrom, to < wideTo — superset hit.
	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, wideFrom, narrowTo)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if inner.callCount != 0 {
		t.Errorf("inner call count = %d, want 0", inner.callCount)
	}
	for _, c := range got {
		if !c.Timestamp.Before(narrowTo) {
			t.Errorf("candle %v outside requested range (to=%v)", c.Timestamp, narrowTo)
		}
	}
}

// TestSupersetHit_ToEqualCachedEnd verifies that to == cachedTo is treated as a superset hit.
func TestSupersetHit_ToEqualCachedEnd(t *testing.T) {
	wideFrom := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	wideTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	allCandles := wideCandles(testInstrument, testTF, wideFrom, wideTo)

	inner := &mockProvider{candles: allCandles}
	cp := newTestProvider(t, inner)

	widePath := cp.cachePath(testInstrument, testTF, wideFrom, wideTo)
	if err := cp.writeCache(widePath, allCandles); err != nil {
		t.Fatalf("writeCache: %v", err)
	}

	// from > wideFrom, to == wideTo — superset hit.
	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, narrowFrom, wideTo)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if inner.callCount != 0 {
		t.Errorf("inner call count = %d, want 0", inner.callCount)
	}
	for _, c := range got {
		if c.Timestamp.Before(narrowFrom) {
			t.Errorf("candle %v before narrowFrom %v", c.Timestamp, narrowFrom)
		}
	}
}

// TestSupersetMiss_NoSupersetFile verifies that when no file covers [from, to], the inner
// provider is called and the result is written to disk.
func TestSupersetMiss_NoSupersetFile(t *testing.T) {
	// Write a narrow file that does NOT cover the wider request.
	narrowFrom := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowTo := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	wideFrom := time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)
	wideTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	smallCandles := wideCandles(testInstrument, testTF, narrowFrom, narrowTo)
	bigCandles := wideCandles(testInstrument, testTF, wideFrom, wideTo)

	inner := &mockProvider{candles: bigCandles}
	cp := newTestProvider(t, inner)

	// Write the narrow file — it does NOT cover the wide request.
	smallPath := cp.cachePath(testInstrument, testTF, narrowFrom, narrowTo)
	if err := cp.writeCache(smallPath, smallCandles); err != nil {
		t.Fatalf("writeCache narrow: %v", err)
	}

	// Request the wide range — narrow file is not a superset, must hit network.
	got, err := cp.FetchCandles(context.Background(), testInstrument, testTF, wideFrom, wideTo)
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if inner.callCount != 1 {
		t.Errorf("inner call count = %d, want 1 (no superset available)", inner.callCount)
	}
	if len(got) != len(bigCandles) {
		t.Errorf("got %d candles, want %d", len(got), len(bigCandles))
	}
}

// TestNoWriteOnSupersetHit verifies that when a superset file is used to serve a request,
// no additional file is written (file count in the instrument dir stays at 1).
func TestNoWriteOnSupersetHit(t *testing.T) {
	wideFrom := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	wideTo := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowFrom := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	narrowTo := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	allCandles := wideCandles(testInstrument, testTF, wideFrom, wideTo)

	inner := &mockProvider{candles: allCandles}
	cp := newTestProvider(t, inner)

	widePath := cp.cachePath(testInstrument, testTF, wideFrom, wideTo)
	if err := cp.writeCache(widePath, allCandles); err != nil {
		t.Fatalf("writeCache wide: %v", err)
	}

	if _, err := cp.FetchCandles(context.Background(), testInstrument, testTF, narrowFrom, narrowTo); err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}

	// The instrument dir should still contain exactly one file.
	instrDir := filepath.Dir(widePath)
	entries, err := os.ReadDir(instrDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("file count = %d, want 1 (no new file written on superset hit)", len(entries))
	}
}
