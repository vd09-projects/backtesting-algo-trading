package zerodha

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// compile-time check that Provider satisfies the DataProvider interface.
var _ provider.DataProvider = (*Provider)(nil)

// Config configures the Provider.
type Config struct {
	APIKey      string
	AccessToken string
	// BaseURL overrides the Kite Connect API base URL. Empty defaults to the real API.
	// Set to an httptest.Server URL in tests.
	BaseURL string
	// HTTPClient overrides the default HTTP client. Nil defaults to a 15-second timeout client.
	HTTPClient *http.Client
	// Sleep overrides time.Sleep for rate-limit throttling between chunk requests.
	// Nil defaults to time.Sleep. Set to a no-op in tests to avoid slow tests.
	Sleep func(time.Duration)
	// InstrumentsCacheDir is the directory where instruments.csv is cached.
	// When non-empty, NewProvider loads the instruments master from disk if the
	// file is less than 24h old, skipping the kite.Instruments() network call
	// entirely — no valid access token is required for cached runs.
	// When empty, the network call is always made (original behavior).
	//
	// **Decision (InstrumentsCacheDir as explicit Config field) — architecture: experimental**
	// scope: pkg/provider/zerodha.Config
	// tags: instruments, caching, token-free, explicit-dependency
	// owner: priya
	//
	// The cache path is an explicit Config field rather than an env-var default
	// inside NewProvider. Callers (cmdutil.BuildProvider) control it; the provider
	// does not assume a path. This follows the CachedProvider precedent: cacheDir
	// is always explicit, never inferred. The field is optional (empty = no cache)
	// to preserve backward compatibility for callers that don't pass a cache dir.
	InstrumentsCacheDir string
}

// Provider implements provider.DataProvider using the Kite Connect API.
// Instrument token lookup is resolved once at construction time from the instruments master CSV.
// Auth token management (login flow, token persistence) is the caller's responsibility;
// pass a valid AccessToken in Config. FetchCandles returns ErrAuthRequired on 401/403.
type Provider struct {
	apiKey      string
	accessToken string
	baseURL     string
	httpClient  *http.Client
	sleep       func(time.Duration)
	tokens      map[string]int64 // "EXCHANGE:TRADINGSYMBOL" → instrument_token
}

// NewProvider creates a Provider and resolves the instruments master CSV.
//
// If Config.InstrumentsCacheDir is set and {cacheDir}/instruments.csv is less
// than 24h old, the file is loaded from disk — no network call, no valid token
// required. If the file is absent or stale, the CSV is fetched from the Kite
// API (token required) and written to cache for future runs.
//
// If Config.InstrumentsCacheDir is empty (the original behavior), the
// instruments CSV is always fetched from the network. Returns ErrAuthRequired
// if APIKey or AccessToken is empty in that case.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) { //nolint:gocritic // hugeParam: Config is a value type at the API boundary; changing to pointer would require all callers to take address of literals
	if cfg.BaseURL == "" {
		cfg.BaseURL = kiteAPIBaseURL
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.Sleep == nil {
		cfg.Sleep = time.Sleep
	}

	var tokens map[string]int64
	var err error

	if cfg.InstrumentsCacheDir != "" {
		// Cache-aware path: may skip network if a fresh file exists.
		tokens, err = loadOrCacheInstrumentsCSV(ctx, cfg.HTTPClient, cfg.BaseURL, cfg.APIKey, cfg.AccessToken, cfg.InstrumentsCacheDir)
		if err != nil {
			return nil, fmt.Errorf("zerodha: load instruments: %w", err)
		}
	} else {
		// Original uncached path: always fetches from network; token required.
		if cfg.APIKey == "" || cfg.AccessToken == "" {
			return nil, ErrAuthRequired
		}
		tokens, err = loadInstrumentsCSV(ctx, cfg.HTTPClient, cfg.BaseURL, cfg.APIKey, cfg.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("zerodha: load instruments: %w", err)
		}
	}

	return &Provider{
		apiKey:      cfg.APIKey,
		accessToken: cfg.AccessToken,
		baseURL:     cfg.BaseURL,
		httpClient:  cfg.HTTPClient,
		sleep:       cfg.Sleep,
		tokens:      tokens,
	}, nil
}

// SupportedTimeframes returns the timeframes supported by the Kite Connect historical API.
// TimeframeWeekly is intentionally omitted — Kite Connect has no weekly interval.
func (p *Provider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{
		model.Timeframe1Min,
		model.Timeframe5Min,
		model.Timeframe15Min,
		model.TimeframeDaily,
	}
}

// FetchCandles returns candles for [from, to) for the given instrument and timeframe.
// Large date ranges are transparently chunked into multiple API requests.
// 350ms is slept between chunks to respect the 3 req/sec rate limit.
//
// After all chunks are merged, a completeness check is applied: if the returned
// candle count is below 95% of the weekday-based estimate, *ErrIncompleteData is
// returned. Callers can use errors.As to inspect Expected and Got and decide
// whether to proceed with partial data.
//
// instrument must be in the format "EXCHANGE:TRADINGSYMBOL" (e.g. "NSE:NIFTY 50").
// Returns ErrInstrumentNotFound if the instrument is not in the instruments map.
// Returns ErrUnsupportedTimeframe for TimeframeWeekly.
// Returns ErrAuthRequired if the API returns 401 or 403.
func (p *Provider) FetchCandles(ctx context.Context, instrument string, tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
	token, ok := p.tokens[instrument]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrInstrumentNotFound, instrument)
	}

	interval, err := timeframeToInterval(tf)
	if err != nil {
		return nil, err
	}

	windows, err := chunkDateRange(from, to, tf)
	if err != nil {
		return nil, err
	}

	var all []model.Candle
	var allSkipped []SkippedCandle
	for i, w := range windows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		candles, skipped, err := p.fetchChunk(ctx, instrument, tf, token, interval, w.from, w.to)
		if err != nil {
			return nil, err
		}
		all = append(all, candles...)
		allSkipped = append(allSkipped, skipped...)

		if i < len(windows)-1 {
			p.sleep(350 * time.Millisecond)
		}
	}

	// Completeness check: reject silently-short slices from chunked fetches.
	// Skip when the expected count is zero (empty or same-day range).
	// The check runs against the count of valid (non-skipped) candles; losing
	// a handful of bad candles does not affect the 90% threshold in practice
	// (1 skipped out of ~98,900 valid candles ≈ 0.001% loss).
	if err := checkCompleteness(instrument, tf, from, to, len(all)); err != nil {
		return nil, err
	}

	// Return a non-fatal ErrBadCandles warning alongside the valid candles.
	// Callers that do not check errors.As will see a non-nil error (fail-safe).
	// cmd/fetch-history uses errors.As to detect this and log/record in manifest.
	if len(allSkipped) > 0 {
		return all, &ErrBadCandles{Instrument: instrument, Skipped: allSkipped}
	}
	return all, nil
}

// checkCompleteness validates that got is within 90% of the weekday-based
// expected candle count for [from, to). Returns *ErrIncompleteData if below
// threshold, nil otherwise. A zero expected count (empty range) is always valid.
//
// **Decision (completeness threshold 90%) — tradeoff: experimental**
// scope: pkg/provider/zerodha.Provider.FetchCandles
// tags: chunk-completeness, ErrIncompleteData, NSE-holidays
// owner: priya
//
// 90% floor (±10% tolerance). Weekday count over-estimates trading days because
// it ignores NSE public holidays. Annually ~10 holidays on ~252 weekdays ≈ 4%
// gap — 5% was sufficient in aggregate, but short windows (2–4 weeks) with a
// holiday cluster (e.g. Diwali) can drop to 80–85%, tripping a stricter gate.
// 10% tolerance absorbs realistic holiday density in any sub-range while still
// catching genuine data gaps (>10% missing is a real problem).
func checkCompleteness(instrument string, tf model.Timeframe, from, to time.Time, got int) error {
	expected := weekdayCount(from, to) * candlesPerDay(tf)
	if expected == 0 {
		return nil // empty range; nothing to validate
	}
	threshold := int(float64(expected) * 0.90)
	if got < threshold {
		return &ErrIncompleteData{
			Instrument: instrument,
			From:       from,
			To:         to,
			Expected:   expected,
			Got:        got,
		}
	}
	return nil
}

// fetchChunk fetches candles for a single date window from the Kite API.
// from and to are converted to IST before formatting, as the API interprets
// query parameters in IST time.
func (p *Provider) fetchChunk(ctx context.Context, instrument string, tf model.Timeframe, token int64, interval string, from, to time.Time) ([]model.Candle, []SkippedCandle, error) {
	endpoint := fmt.Sprintf("%s/instruments/historical/%d/%s", p.baseURL, token, interval)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, nil, err
	}

	// Kite Connect interprets from/to as IST. Convert from UTC before formatting.
	ist := time.FixedZone("IST", 5*3600+30*60)
	q := req.URL.Query()
	q.Set("from", from.In(ist).Format("2006-01-02 15:04:05"))
	q.Set("to", to.In(ist).Format("2006-01-02 15:04:05"))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", p.apiKey, p.accessToken))
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doHTTP(p.httpClient, req)
	if err != nil {
		return nil, nil, fmt.Errorf("fetch chunk [%s, %s): %w",
			from.Format(time.RFC3339), to.Format(time.RFC3339), err)
	}

	return parseKiteCandles(instrument, tf, body)
}

// parseKiteCandles parses the Kite Connect array-of-arrays candle response.
// Each row is [timestamp, open, high, low, close, volume].
// Timestamps are returned in UTC. Each candle is validated via model.NewCandle.
//
// Structural errors (malformed JSON, missing rows, unparseable timestamps) are
// returned as hard errors — the chunk cannot be used. OHLC validation failures
// from model.NewCandle are collected into the returned []SkippedCandle and the
// candle is skipped rather than aborting the entire fetch. This handles the
// Zerodha tick-vs-aggregation artifact where open sits marginally outside [low, high].
//
// **Decision (skip OHLC-invalid candles in parseKiteCandles, hard-fail on structural errors) — convention: experimental**
// scope: pkg/provider/zerodha.parseKiteCandles
// tags: bad-candle, skip, OHLC-validation, structural-error, TASK-0100
// owner: priya
//
// Structural errors (malformed row, bad timestamp) are unrecoverable — the row
// cannot be interpreted, and proceeding would corrupt the candle series order.
// OHLC validation errors are recoverable — the row is well-formed, the timestamp
// and surrounding candles are valid; only one price relationship is violated.
// This matches the Zerodha artifact: open=835.6 with low=837.4 is a tick-boundary
// misalignment, not a protocol error.
func parseKiteCandles(instrument string, tf model.Timeframe, body []byte) ([]model.Candle, []SkippedCandle, error) {
	var envelope struct {
		Data struct {
			Candles [][]interface{} `json:"candles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, nil, fmt.Errorf("parse candles response: %w", err)
	}

	// Use a fixed-offset IST zone as fallback; time.LoadLocation may fail in
	// minimal containers lacking the timezone database.
	ist := time.FixedZone("IST", 5*3600+30*60)

	candles := make([]model.Candle, 0, len(envelope.Data.Candles))
	var skipped []SkippedCandle
	for i, row := range envelope.Data.Candles {
		if len(row) < 6 {
			// Structural error: row is truncated, cannot determine timestamp or OHLC.
			return nil, nil, fmt.Errorf("zerodha: candle[%d]: expected 6 elements, got %d", i, len(row))
		}

		tsStr, ok := row[0].(string)
		if !ok {
			// Structural error: timestamp field is not a string.
			return nil, nil, fmt.Errorf("zerodha: candle[%d]: timestamp is not a string", i)
		}

		// Kite returns "2024-01-01T09:15:00+0530" — the +0530 offset has no colon,
		// which is NOT valid RFC 3339. Use the explicit layout instead.
		ts, err := time.ParseInLocation("2006-01-02T15:04:05-0700", tsStr, ist)
		if err != nil {
			// Try RFC 3339 as a fallback for future API changes.
			ts, err = time.Parse(time.RFC3339, tsStr)
			if err != nil {
				// Structural error: timestamp is unparseable.
				return nil, nil, fmt.Errorf("zerodha: candle[%d]: parse timestamp %q: %w", i, tsStr, err)
			}
		}

		c, err := model.NewCandle(
			instrument, tf, ts.UTC(),
			toFloat64(row[1]), toFloat64(row[2]), toFloat64(row[3]),
			toFloat64(row[4]), toFloat64(row[5]),
		)
		if err != nil {
			// OHLC validation error: skip this candle, collect reason, continue.
			// The completeness check will catch the case where too many are skipped.
			skipped = append(skipped, SkippedCandle{
				Index:  i,
				Reason: fmt.Sprintf("zerodha: candle[%d]: %s", i, err.Error()),
			})
			continue
		}
		candles = append(candles, c)
	}
	return candles, skipped, nil
}

// timeframeToInterval maps model.Timeframe to the Kite Connect interval string.
func timeframeToInterval(tf model.Timeframe) (string, error) {
	intervals := map[model.Timeframe]string{
		model.Timeframe1Min:  "minute",
		model.Timeframe5Min:  "5minute",
		model.Timeframe15Min: "15minute",
		model.TimeframeDaily: "day",
	}
	iv, ok := intervals[tf]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedTimeframe, tf)
	}
	return iv, nil
}
