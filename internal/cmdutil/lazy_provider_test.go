package cmdutil

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// stubDataProvider is a minimal provider.DataProvider for testing.
type stubDataProvider struct {
	candles []model.Candle
	err     error
}

func (s *stubDataProvider) FetchCandles(_ context.Context, _ string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	return s.candles, s.err
}
func (s *stubDataProvider) SupportedTimeframes() []model.Timeframe { return nil }

func TestLazyProvider_InitCalledOnFirstFetch(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	stub := &stubDataProvider{candles: []model.Candle{{Instrument: "X"}}}

	lp := &lazyProvider{
		initFn: func() (provider.DataProvider, error) {
			calls.Add(1)
			return stub, nil
		},
	}

	// initFn must not fire before FetchCandles.
	if calls.Load() != 0 {
		t.Fatal("initFn called before FetchCandles")
	}

	got, err := lp.FetchCandles(context.Background(), "X", model.TimeframeDaily, time.Time{}, time.Time{})
	if err != nil {
		t.Fatalf("FetchCandles: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 candle, got %d", len(got))
	}
	if calls.Load() != 1 {
		t.Fatalf("initFn called %d times, want 1", calls.Load())
	}
}

func TestLazyProvider_InitCalledOnlyOnce(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	stub := &stubDataProvider{}

	lp := &lazyProvider{
		initFn: func() (provider.DataProvider, error) {
			calls.Add(1)
			return stub, nil
		},
	}

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := lp.FetchCandles(context.Background(), "X", model.TimeframeDaily, time.Time{}, time.Time{}); err != nil {
				t.Errorf("FetchCandles: %v", err)
			}
		}()
	}
	wg.Wait()

	if calls.Load() != 1 {
		t.Fatalf("initFn called %d times under concurrency, want exactly 1", calls.Load())
	}
}

func TestLazyProvider_InitErrorPropagated(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("auth failed")

	lp := &lazyProvider{
		initFn: func() (provider.DataProvider, error) {
			return nil, wantErr
		},
	}

	_, err := lp.FetchCandles(context.Background(), "X", model.TimeframeDaily, time.Time{}, time.Time{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("want %v, got %v", wantErr, err)
	}
}

func TestLazyProvider_SupportedTimeframes(t *testing.T) {
	t.Parallel()
	lp := &lazyProvider{initFn: func() (provider.DataProvider, error) { panic("must not init") }}
	tfs := lp.SupportedTimeframes()
	if len(tfs) == 0 {
		t.Fatal("SupportedTimeframes returned empty slice")
	}
	// initFn must NOT have fired (no panic means we're good).
}
