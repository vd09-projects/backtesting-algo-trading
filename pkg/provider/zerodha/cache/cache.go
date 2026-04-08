// Package cache provides a file-based caching decorator for provider.DataProvider.
// It wraps any DataProvider and transparently caches FetchCandles results as JSON
// files under a configurable directory, eliminating redundant API calls during
// iterative backtesting sessions.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// CachedProvider wraps a DataProvider and caches FetchCandles results to disk.
// Each unique (instrument, timeframe, from, to) tuple is stored as one JSON file.
// Historical ranges (to < today UTC) are cached indefinitely; recent ranges use a 1-hour TTL.
type CachedProvider struct {
	inner    provider.DataProvider
	cacheDir string
	now      func() time.Time
}

// NewCachedProvider returns a CachedProvider backed by inner, writing cache files to cacheDir.
// cacheDir is created on first write; it need not exist at construction time.
func NewCachedProvider(inner provider.DataProvider, cacheDir string) *CachedProvider {
	return &CachedProvider{
		inner:    inner,
		cacheDir: cacheDir,
		now:      time.Now,
	}
}

// cacheEntry is the on-disk JSON format for a cached candle set.
type cacheEntry struct {
	CachedAt time.Time      `json:"cached_at"`
	Candles  []model.Candle `json:"candles"`
}

// FetchCandles returns candles for [from, to), served from the local cache when valid.
// On a cache miss, the inner provider is called, the result is persisted to disk, and
// the candles are returned. A write failure silently degrades to uncached behavior.
func (c *CachedProvider) FetchCandles(ctx context.Context, instrument string, tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
	path := c.cachePath(instrument, tf, from, to)

	if candles, ok := c.readCache(path, to); ok {
		return candles, nil
	}

	candles, err := c.inner.FetchCandles(ctx, instrument, tf, from, to)
	if err != nil {
		return nil, err
	}

	// Best-effort write: a cache failure must not fail the fetch.
	_ = c.writeCache(path, candles) //nolint:errcheck // best-effort; cache failure must not fail the caller
	return candles, nil
}

// SupportedTimeframes delegates to the inner provider.
func (c *CachedProvider) SupportedTimeframes() []model.Timeframe {
	return c.inner.SupportedTimeframes()
}

// cachePath returns the filesystem path for the given (instrument, timeframe, from, to) tuple.
// Format: {cacheDir}/{sanitized_instrument}/{timeframe}_{from}_{to}.json
// Instrument sanitisation: colon and spaces → underscore, all lowercase.
// from/to are formatted as YYYY-MM-DD (UTC).
func (c *CachedProvider) cachePath(instrument string, tf model.Timeframe, from, to time.Time) string {
	sanitized := strings.NewReplacer(":", "_", " ", "_").Replace(strings.ToLower(instrument))
	filename := fmt.Sprintf("%s_%s_%s.json",
		string(tf),
		from.UTC().Format("2006-01-02"),
		to.UTC().Format("2006-01-02"),
	)
	return filepath.Join(c.cacheDir, sanitized, filename)
}

// readCache reads and validates a cache file at path.
// Returns (candles, true) on a valid hit, (nil, false) on a miss, corrupt file, or TTL expiry.
// TTL (1 hour) applies only when to >= today UTC; historical ranges never expire.
func (c *CachedProvider) readCache(path string, to time.Time) ([]model.Candle, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Apply TTL only to recent/intraday data (to >= today).
	today := c.now().UTC().Truncate(24 * time.Hour)
	if !to.UTC().Before(today) && c.now().UTC().Sub(entry.CachedAt.UTC()) > time.Hour {
		_ = os.Remove(path) //nolint:errcheck // best-effort; stale file left on disk is harmless
		return nil, false
	}

	return entry.Candles, true
}

// writeCache serializes candles to a JSON file at path, creating parent directories as needed.
func (c *CachedProvider) writeCache(path string, candles []model.Candle) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	entry := cacheEntry{
		CachedAt: c.now().UTC(),
		Candles:  candles,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}
