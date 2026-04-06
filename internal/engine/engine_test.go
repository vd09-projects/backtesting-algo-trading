package engine_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// ── Stubs ─────────────────────────────────────────────────────────────────────

// stubProvider returns a fixed slice of candles regardless of arguments.
type stubProvider struct {
	candles []model.Candle
	err     error
}

func (s *stubProvider) FetchCandles(_ context.Context, _ string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	return s.candles, s.err
}
func (s *stubProvider) SupportedTimeframes() []model.Timeframe { return nil }

var _ provider.DataProvider = (*stubProvider)(nil)

// stubStrategy records every call to Next so tests can inspect what it received.
type stubStrategy struct {
	lookback int
	signal   model.Signal
	calls    [][]model.Candle // copy of candles slice passed on each call
}

func (s *stubStrategy) Name() string               { return "stub" }
func (s *stubStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *stubStrategy) Lookback() int              { return s.lookback }
func (s *stubStrategy) Next(candles []model.Candle) model.Signal {
	cp := make([]model.Candle, len(candles))
	copy(cp, candles)
	s.calls = append(s.calls, cp)
	return s.signal
}

var _ strategy.Strategy = (*stubStrategy)(nil)

// ── Helpers ───────────────────────────────────────────────────────────────────

var base = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

func makeCandles(n int) []model.Candle {
	candles := make([]model.Candle, n)
	for i := range candles {
		candles[i] = model.Candle{
			Instrument: "TEST",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       100 + float64(i),
			High:       105 + float64(i),
			Low:        95 + float64(i),
			Close:      102 + float64(i),
			Volume:     1000,
		}
	}
	return candles
}

func defaultConfig() engine.Config {
	return engine.Config{
		Instrument:           "TEST",
		From:                 base,
		To:                   base.AddDate(0, 0, 30),
		InitialCash:          100_000,
		PositionSizeFraction: 0.1,
	}
}

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestRun_SignalsCollected(t *testing.T) {
	candles := makeCandles(10)
	p := &stubProvider{candles: candles}
	s := &stubStrategy{lookback: 1, signal: model.SignalBuy}

	e := engine.New(defaultConfig())
	require.NoError(t, e.Run(context.Background(), p, s))

	results := e.Results()
	assert.Len(t, results, 10)
	for _, r := range results {
		assert.Equal(t, model.SignalBuy, r.Signal)
	}
}

func TestRun_LookbackRespected(t *testing.T) {
	// With lookback=3 and 10 candles, strategy should be called for bars 2..9 (8 calls).
	candles := makeCandles(10)
	p := &stubProvider{candles: candles}
	s := &stubStrategy{lookback: 3, signal: model.SignalHold}

	e := engine.New(defaultConfig())
	require.NoError(t, e.Run(context.Background(), p, s))

	assert.Len(t, s.calls, 8, "strategy called once per bar starting at lookback index")
	assert.Len(t, e.Results(), 8)
}

func TestRun_NoLookahead(t *testing.T) {
	// At bar i (0-indexed), strategy must receive exactly i+1 candles.
	candles := makeCandles(5)
	p := &stubProvider{candles: candles}
	s := &stubStrategy{lookback: 1, signal: model.SignalHold}

	e := engine.New(defaultConfig())
	require.NoError(t, e.Run(context.Background(), p, s))

	for callIdx, seen := range s.calls {
		expectedLen := callIdx + 1 // bar 0 → 1 candle, bar 1 → 2 candles, …
		assert.Len(t, seen, expectedLen,
			"call %d: strategy must see exactly %d candles (no lookahead)", callIdx, expectedLen)
	}
}

func TestRun_BarResultCandleMatchesCurrentBar(t *testing.T) {
	// BarResult.Candle must be the candle at bar i, not some future bar.
	candles := makeCandles(5)
	p := &stubProvider{candles: candles}
	s := &stubStrategy{lookback: 1, signal: model.SignalHold}

	e := engine.New(defaultConfig())
	require.NoError(t, e.Run(context.Background(), p, s))

	results := e.Results()
	for i, r := range results {
		assert.Equal(t, candles[i].Timestamp, r.Candle.Timestamp,
			"result %d candle timestamp mismatch", i)
	}
}

func TestRun_LookbackEqualsNCandles(t *testing.T) {
	// Exactly one call when candles == lookback.
	candles := makeCandles(5)
	p := &stubProvider{candles: candles}
	s := &stubStrategy{lookback: 5, signal: model.SignalBuy}

	e := engine.New(defaultConfig())
	require.NoError(t, e.Run(context.Background(), p, s))

	assert.Len(t, s.calls, 1)
	assert.Len(t, s.calls[0], 5)
}

func TestRun_ProviderError(t *testing.T) {
	p := &stubProvider{err: fmt.Errorf("network timeout")}
	s := &stubStrategy{lookback: 1, signal: model.SignalHold}

	e := engine.New(defaultConfig())
	err := e.Run(context.Background(), p, s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "network timeout")
}

func TestRun_NoCandles(t *testing.T) {
	p := &stubProvider{candles: nil}
	s := &stubStrategy{lookback: 1, signal: model.SignalHold}

	e := engine.New(defaultConfig())
	err := e.Run(context.Background(), p, s)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no candles")
}

func TestRun_ValidationErrors(t *testing.T) {
	candles := makeCandles(5)

	cases := []struct {
		name    string
		cfg     engine.Config
		wantErr string
	}{
		{
			name: "empty instrument",
			cfg: engine.Config{
				Instrument: "",
				From:       base, To: base.AddDate(0, 0, 10),
				InitialCash: 100_000,
			},
			wantErr: "instrument",
		},
		{
			name: "zero From",
			cfg: engine.Config{
				Instrument: "TEST",
				From:       time.Time{}, To: base.AddDate(0, 0, 10),
				InitialCash: 100_000,
			},
			wantErr: "From and To",
		},
		{
			name: "To before From",
			cfg: engine.Config{
				Instrument: "TEST",
				From:       base.AddDate(0, 0, 10), To: base,
				InitialCash: 100_000,
			},
			wantErr: "must be after",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &stubProvider{candles: candles}
			s := &stubStrategy{lookback: 1, signal: model.SignalHold}
			err := engine.New(tc.cfg).Run(context.Background(), p, s)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}
