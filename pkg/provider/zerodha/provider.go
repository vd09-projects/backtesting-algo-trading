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

// compile-time check that ZerodhaProvider satisfies the DataProvider interface.
var _ provider.DataProvider = (*ZerodhaProvider)(nil)

// Config configures the ZerodhaProvider.
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
}

// ZerodhaProvider implements provider.DataProvider using the Kite Connect API.
// Instrument token lookup is resolved once at construction time from the instruments master CSV.
// Auth token management (login flow, token persistence) is the caller's responsibility;
// pass a valid AccessToken in Config. FetchCandles returns ErrAuthRequired on 401/403.
type ZerodhaProvider struct {
	apiKey      string
	accessToken string
	baseURL     string
	httpClient  *http.Client
	sleep       func(time.Duration)
	tokens      map[string]int64 // "EXCHANGE:TRADINGSYMBOL" → instrument_token
}

// NewZerodhaProvider creates a ZerodhaProvider and downloads the instruments CSV.
// Returns ErrAuthRequired if APIKey or AccessToken is empty.
// The instruments download makes one HTTP call and takes ~1s at startup.
func NewZerodhaProvider(ctx context.Context, cfg Config) (*ZerodhaProvider, error) {
	if cfg.APIKey == "" || cfg.AccessToken == "" {
		return nil, ErrAuthRequired
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = kiteAPIBaseURL
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.Sleep == nil {
		cfg.Sleep = time.Sleep
	}

	tokens, err := loadInstrumentsCSV(ctx, cfg.HTTPClient, cfg.BaseURL, cfg.APIKey, cfg.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("zerodha: load instruments: %w", err)
	}

	return &ZerodhaProvider{
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
func (p *ZerodhaProvider) SupportedTimeframes() []model.Timeframe {
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
// instrument must be in the format "EXCHANGE:TRADINGSYMBOL" (e.g. "NSE:NIFTY 50").
// Returns ErrInstrumentNotFound if the instrument is not in the instruments map.
// Returns ErrUnsupportedTimeframe for TimeframeWeekly.
// Returns ErrAuthRequired if the API returns 401 or 403.
func (p *ZerodhaProvider) FetchCandles(ctx context.Context, instrument string, tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
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
	for i, w := range windows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		candles, err := p.fetchChunk(ctx, instrument, tf, token, interval, w.from, w.to)
		if err != nil {
			return nil, err
		}
		all = append(all, candles...)

		if i < len(windows)-1 {
			p.sleep(350 * time.Millisecond)
		}
	}
	return all, nil
}

// fetchChunk fetches candles for a single date window from the Kite API.
// from and to are converted to IST before formatting, as the API interprets
// query parameters in IST time.
func (p *ZerodhaProvider) fetchChunk(ctx context.Context, instrument string, tf model.Timeframe, token int64, interval string, from, to time.Time) ([]model.Candle, error) {
	endpoint := fmt.Sprintf("%s/instruments/historical/%d/%s", p.baseURL, token, interval)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("fetch chunk [%s, %s): %w",
			from.Format(time.RFC3339), to.Format(time.RFC3339), err)
	}

	return parseKiteCandles(instrument, tf, body)
}

// parseKiteCandles parses the Kite Connect array-of-arrays candle response.
// Each row is [timestamp, open, high, low, close, volume].
// Timestamps are returned in UTC. Each candle is validated via model.NewCandle.
func parseKiteCandles(instrument string, tf model.Timeframe, body []byte) ([]model.Candle, error) {
	var envelope struct {
		Data struct {
			Candles [][]interface{} `json:"candles"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("parse candles response: %w", err)
	}

	// Use a fixed-offset IST zone as fallback; time.LoadLocation may fail in
	// minimal containers lacking the timezone database.
	ist := time.FixedZone("IST", 5*3600+30*60)

	candles := make([]model.Candle, 0, len(envelope.Data.Candles))
	for i, row := range envelope.Data.Candles {
		if len(row) < 6 {
			return nil, fmt.Errorf("zerodha: candle[%d]: expected 6 elements, got %d", i, len(row))
		}

		tsStr, ok := row[0].(string)
		if !ok {
			return nil, fmt.Errorf("zerodha: candle[%d]: timestamp is not a string", i)
		}

		// Kite returns "2024-01-01T09:15:00+0530" — the +0530 offset has no colon,
		// which is NOT valid RFC 3339. Use the explicit layout instead.
		ts, err := time.ParseInLocation("2006-01-02T15:04:05-0700", tsStr, ist)
		if err != nil {
			// Try RFC 3339 as a fallback for future API changes.
			ts, err = time.Parse(time.RFC3339, tsStr)
			if err != nil {
				return nil, fmt.Errorf("zerodha: candle[%d]: parse timestamp %q: %w", i, tsStr, err)
			}
		}

		c, err := model.NewCandle(
			instrument, tf, ts.UTC(),
			toFloat64(row[1]), toFloat64(row[2]), toFloat64(row[3]),
			toFloat64(row[4]), toFloat64(row[5]),
		)
		if err != nil {
			return nil, fmt.Errorf("zerodha: candle[%d]: %w", i, err)
		}
		candles = append(candles, c)
	}
	return candles, nil
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
