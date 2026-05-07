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

// dateLayout is the YYYY-MM-DD format used in cache filenames.
const dateLayout = "2006-01-02"

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
//
// Lookup order:
//  1. Exact-match file: {tf}_{from}_{to}.json — checked first (O(1) stat).
//  2. Superset file: any {tf}_{cachedFrom}_{cachedTo}.json where cachedFrom <= from
//     and cachedTo >= to — candles are filtered to [from, to) in memory. No network call.
//  3. Network fetch: inner provider called, result written to disk.
//
// A write failure silently degrades to uncached behavior.
//
// Decision (superset-lookup via filename parsing — no index file) — architecture: experimental
// Decision (no narrower file written on superset hit — superset already covers future requests) — tradeoff: experimental
func (c *CachedProvider) FetchCandles(ctx context.Context, instrument string, tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
	// 1. Exact-match path — unchanged behavior.
	path := c.cachePath(instrument, tf, from, to)
	if candles, ok := c.readCache(path, to); ok {
		return candles, nil
	}

	// 2. Superset lookup — scan instrument dir for any file whose range covers [from, to].
	if superPath, ok := c.findSupersetFile(instrument, tf, from, to); ok {
		if candles, ok := c.readCache(superPath, to); ok {
			return filterCandles(candles, from, to), nil
		}
		// Superset file exists but failed to read (corrupt/expired) — fall through to network.
	}

	// 3. Network fetch.
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

// sanitizeInstrument converts an instrument name to a filesystem-safe lowercase string.
// Colons and spaces are replaced with underscores.
func sanitizeInstrument(instrument string) string {
	return strings.NewReplacer(":", "_", " ", "_").Replace(strings.ToLower(instrument))
}

// instrDir returns the cache subdirectory path for the given instrument.
func (c *CachedProvider) instrDir(instrument string) string {
	return filepath.Join(c.cacheDir, sanitizeInstrument(instrument))
}

// cachePath returns the filesystem path for the given (instrument, timeframe, from, to) tuple.
// Format: {cacheDir}/{sanitized_instrument}/{timeframe}_{from}_{to}.json
// Instrument sanitisation: colon and spaces → underscore, all lowercase.
// from/to are formatted as YYYY-MM-DD (UTC).
func (c *CachedProvider) cachePath(instrument string, tf model.Timeframe, from, to time.Time) string {
	filename := fmt.Sprintf("%s_%s_%s.json",
		string(tf),
		from.UTC().Format(dateLayout),
		to.UTC().Format(dateLayout),
	)
	return filepath.Join(c.instrDir(instrument), filename)
}

// findSupersetFile scans the instrument's cache directory for any file whose
// filename-encoded date range is a superset of [from, to] for the given timeframe.
// Returns (path, true) if a superset file is found; ("", false) otherwise.
// Only filename parsing is used — no file I/O per candidate.
//
// Filename format: {tf}_{from}_{to}.json (e.g. "daily_2018-01-01_2025-01-01.json")
func (c *CachedProvider) findSupersetFile(instrument string, tf model.Timeframe, from, to time.Time) (string, bool) {
	dir := c.instrDir(instrument)
	entries, err := os.ReadDir(dir)
	if err != nil {
		// Directory absent — no cached files.
		return "", false
	}

	prefix := string(tf) + "_"
	fromUTC := from.UTC().Truncate(24 * time.Hour)
	toUTC := to.UTC().Truncate(24 * time.Hour)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".json") {
			continue
		}
		// Strip prefix and suffix to get "YYYY-MM-DD_YYYY-MM-DD".
		body := name[len(prefix) : len(name)-len(".json")]
		parts := strings.SplitN(body, "_", 2)
		if len(parts) != 2 {
			continue
		}
		cachedFrom, err1 := time.Parse(dateLayout, parts[0])
		cachedTo, err2 := time.Parse(dateLayout, parts[1])
		if err1 != nil || err2 != nil {
			continue
		}
		// Superset check: cachedFrom <= from AND cachedTo >= to (date-granular, UTC).
		if !cachedFrom.After(fromUTC) && !cachedTo.Before(toUTC) {
			return filepath.Join(dir, name), true
		}
	}
	return "", false
}

// filterCandles returns the subset of candles whose Timestamp falls in [from, to).
// The returned slice shares no backing array with the input.
//
// Decision (candle subset filter on Timestamp field, half-open [from, to) interval) — convention: experimental
func filterCandles(candles []model.Candle, from, to time.Time) []model.Candle {
	var out []model.Candle
	for _, c := range candles {
		ts := c.Timestamp.UTC()
		if !ts.Before(from.UTC()) && ts.Before(to.UTC()) {
			out = append(out, c)
		}
	}
	return out
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
