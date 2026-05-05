package zerodha

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

const (
	instrumentsCacheFilename = "instruments.csv"
	instrumentsCacheTTL      = 24 * time.Hour
)

// loadOrCacheInstrumentsCSV returns the instrument token map, loading from a
// local cache file when fresh (age < 24h) or fetching from the Kite API otherwise.
//
// cacheDir is the directory where instruments.csv is stored. If the file exists
// and is less than 24h old, no network call is made — apiKey and accessToken
// are not required.
//
// If the cached file is absent or stale, fetchs from the API (token required)
// and writes the result to {cacheDir}/instruments.csv for future runs.
//
// **Decision (instruments CSV cache path) — architecture: experimental**
// scope: pkg/provider/zerodha
// tags: instruments, caching, token-free
// owner: priya
//
// Instruments CSV is written to {cacheDir}/instruments.csv alongside the candle
// cache, keeping all provider-level disk state under one root. The caller
// (cmdutil.BuildProvider) controls cacheDir; NewProvider does not assume a
// default path. This matches the CachedProvider pattern where cacheDir is an
// explicit parameter, not an env-var default inside the provider.
func loadOrCacheInstrumentsCSV(ctx context.Context, client *http.Client, baseURL, apiKey, accessToken, cacheDir string) (map[string]int64, error) {
	cachePath := filepath.Join(cacheDir, instrumentsCacheFilename)

	// Check for a fresh cache file.
	if info, err := os.Stat(cachePath); err == nil {
		age := time.Since(info.ModTime())
		if age < instrumentsCacheTTL {
			data, readErr := os.ReadFile(cachePath)
			if readErr == nil {
				tokens, parseErr := parseInstrumentsCSV(data)
				if parseErr == nil {
					return tokens, nil
				}
				// Corrupt cache: fall through to network fetch.
			}
		}
	}

	// Cache miss or stale: fetch from network.
	data, err := fetchInstrumentsRaw(ctx, client, baseURL, apiKey, accessToken)
	if err != nil {
		return nil, err
	}

	tokens, err := parseInstrumentsCSV(data)
	if err != nil {
		return nil, fmt.Errorf("parse instruments: %w", err)
	}

	// Best-effort write to cache; failure must not prevent the caller from proceeding.
	_ = writeToCacheFile(cachePath, data) //nolint:errcheck // best-effort; cache write failure is non-fatal

	return tokens, nil
}

// fetchInstrumentsRaw downloads the raw instruments CSV bytes from the Kite API.
func fetchInstrumentsRaw(ctx context.Context, client *http.Client, baseURL, apiKey, accessToken string) ([]byte, error) {
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
	return body, nil
}

// writeToCacheFile writes data to cachePath, creating parent directories as needed.
func writeToCacheFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir cache dir: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// weekdayCount returns the number of weekdays (Mon–Fri) in [from, to).
// Saturdays and Sundays are excluded. This is a conservative proxy for NSE
// trading days — the ±10% tolerance in the completeness check absorbs holidays.
func weekdayCount(from, to time.Time) int {
	count := 0
	cur := from.UTC().Truncate(24 * time.Hour)
	end := to.UTC().Truncate(24 * time.Hour)
	for cur.Before(end) {
		wd := cur.Weekday()
		if wd != time.Saturday && wd != time.Sunday {
			count++
		}
		cur = cur.Add(24 * time.Hour)
	}
	return count
}

// candlesPerDay returns the expected number of candles per trading day for the
// given timeframe. Used to compute the completeness check threshold.
func candlesPerDay(tf model.Timeframe) int {
	switch tf {
	case model.Timeframe1Min:
		return 375 // 09:15–15:30 = 375 minutes
	case model.Timeframe5Min:
		return 75 // 375 / 5
	case model.Timeframe15Min:
		return 25 // 375 / 15
	case model.TimeframeDaily:
		return 1
	default:
		return 1
	}
}
