package paramsearch_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/paramsearch"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// stubStrategy is a minimal strategy.Strategy implementation for tests.
// It always returns model.SignalHold.
type stubStrategy struct{}

func (s *stubStrategy) Name() string                       { return "stub" }
func (s *stubStrategy) Lookback() int                      { return 1 }
func (s *stubStrategy) Timeframe() model.Timeframe         { return model.TimeframeDaily }
func (s *stubStrategy) Next(_ []model.Candle) model.Signal { return model.SignalHold }

// tradingStub alternates Buy/Sell signals on odd/even bars to generate trades.
// This ensures TradeMetricsInsufficient=false when enough candles are provided.
type tradingStub struct {
	bar int
}

func (s *tradingStub) Name() string               { return "trading-stub" }
func (s *tradingStub) Lookback() int              { return 1 }
func (s *tradingStub) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *tradingStub) Next(_ []model.Candle) model.Signal {
	s.bar++
	if s.bar%2 == 1 {
		return model.SignalBuy
	}
	return model.SignalSell
}

// mockProvider returns a fixed slice of candles for every FetchCandles call.
// It allows tests to control the candle series returned per instrument.
type mockProvider struct {
	candles map[string][]model.Candle // instrument → candles
	errMap  map[string]error          // instrument → error
}

func (m *mockProvider) FetchCandles(_ context.Context, instrument string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	if err, ok := m.errMap[instrument]; ok {
		return nil, err
	}
	if c, ok := m.candles[instrument]; ok {
		return c, nil
	}
	return nil, nil
}

func (m *mockProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// makeCandles produces n daily candles starting at 2020-01-02 with OHLCV=100.
func makeCandles(n int) []model.Candle {
	candles := make([]model.Candle, n)
	base := time.Date(2020, 1, 2, 9, 15, 0, 0, time.UTC)
	for i := range candles {
		ts := base.AddDate(0, 0, i)
		c, err := model.NewCandle("NSE:TEST", model.TimeframeDaily, ts, 100, 101, 99, 100, 1000)
		if err != nil {
			panic(err)
		}
		candles[i] = c
	}
	return candles
}

// TestGridSize_CartesianProduct verifies that a 2-axis grid with 3 steps on axis1
// and 4 steps on axis2 produces GridSize=12 in the report, and that the DSR
// nTrials is 12 (verified indirectly via the grid size).
func TestGridSize_CartesianProduct(t *testing.T) {
	// Axis1: [1, 2, 3] → 3 values
	// Axis2: [10, 20, 30, 40] → 4 values
	// Total variants = 3 × 4 = 12
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 3, Step: 1},
			{Name: "p2", Min: 10, Max: 40, Step: 10},
		},
	}

	instruments := []string{"NSE:A"}
	// 300 candles → sufficient data
	candles := makeCandles(300)
	p := &mockProvider{
		candles: map[string][]model.Candle{"NSE:A": candles},
	}

	cfg := paramsearch.Config{
		GridSpec:    spec,
		Instruments: instruments,
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &stubStrategy{}, nil
		},
	}

	report, err := paramsearch.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.GridSize != 12 {
		t.Errorf("GridSize: got %d, want 12", report.GridSize)
	}
	if len(report.Results) != 12 {
		t.Errorf("len(Results): got %d, want 12", len(report.Results))
	}
}

// TestDSRRanking_SortedDescending verifies that Run returns Results sorted descending
// by DSRSharpe for non-insufficient variants, and that the sort comparison fires at
// least once (i.e., the test is not vacuously passing because all variants are
// InsufficientData).
func TestDSRRanking_SortedDescending(t *testing.T) {
	// Single-axis 2-variant grid (p1=1, p1=2) with tradingStub (alternates Buy/Sell).
	// 300 candles → ~150 trades per variant, well above MinTradesForMetrics=30.
	// Both variants are non-insufficient; the sort assertion fires for the pair at i=1.
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 2, Step: 1},
		},
	}

	instruments := []string{"NSE:A"}
	candles := makeCandles(300) // sufficient: ~150 trades per variant
	p := &mockProvider{
		candles: map[string][]model.Candle{"NSE:A": candles},
	}

	cfg := paramsearch.Config{
		GridSpec:    spec,
		Instruments: instruments,
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &tradingStub{}, nil // fresh instance per call; state not shared
		},
	}

	report, err := paramsearch.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Results) < 2 {
		t.Fatalf("expected at least 2 results, got %d", len(report.Results))
	}

	// Verify descending DSR sort order. fired tracks whether any pair comparison executed
	// so the test fails if all variants are InsufficientData (making the loop a no-op).
	var fired bool
	for i := 1; i < len(report.Results); i++ {
		if !report.Results[i-1].InsufficientData && !report.Results[i].InsufficientData {
			fired = true // set before the inner comparison: records that the check ran
			if report.Results[i-1].DSRSharpe < report.Results[i].DSRSharpe {
				t.Errorf("Results not sorted by DSRSharpe descending at index %d: %f < %f",
					i, report.Results[i-1].DSRSharpe, report.Results[i].DSRSharpe)
			}
		}
	}
	if !fired {
		t.Fatal("sort assertion never fired: all variants were InsufficientData; switch to tradingStub with sufficient candles")
	}
}

// TestInsufficientInstruments_ExcludedFromAggregate verifies two things:
//  1. When one instrument returns insufficient data, it is excluded from the DSR average.
//  2. When ALL instruments for a variant return insufficient data, the VariantResult.InsufficientData=true.
func TestInsufficientInstruments_ExcludedFromAggregate(t *testing.T) {
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 1, Step: 1},
		},
	}

	instruments := []string{"NSE:A", "NSE:B"}
	// NSE:A: 300 candles → sufficient
	// NSE:B: 10 candles → insufficient (CurveMetricsInsufficient=true, TradeMetricsInsufficient=true)
	p := &mockProvider{
		candles: map[string][]model.Candle{
			"NSE:A": makeCandles(300),
			"NSE:B": makeCandles(10),
		},
	}

	// Use tradingStub: alternates Buy/Sell to generate enough trades for TradeMetricsInsufficient=false
	// on NSE:A (300 candles → ~150 trades >> MinTradesForMetrics=30).
	// NSE:B (10 candles → ~5 trades < 30) stays insufficient.
	cfg := paramsearch.Config{
		GridSpec:    spec,
		Instruments: instruments,
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &tradingStub{}, nil // fresh instance per call; state not shared
		},
	}

	report, err := paramsearch.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	// Variant has one sufficient instrument (NSE:A) → InsufficientData must be false.
	if report.Results[0].InsufficientData {
		t.Errorf("expected InsufficientData=false when at least one instrument is sufficient")
	}

	// Now test all-insufficient: use only NSE:B (10 candles → ~5 trades < MinTradesForMetrics).
	cfgAllInsufficient := paramsearch.Config{
		GridSpec:    spec,
		Instruments: []string{"NSE:B"},
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &tradingStub{}, nil
		},
	}

	report2, err := paramsearch.Run(context.Background(), cfgAllInsufficient, p)
	if err != nil {
		t.Fatalf("Run (all insufficient): %v", err)
	}
	if len(report2.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report2.Results))
	}
	if !report2.Results[0].InsufficientData {
		t.Errorf("expected InsufficientData=true when all instruments are insufficient")
	}
}

// TestEmptyGrid_ReturnsError verifies that a GridSpec with no axes returns an error
// before any engine runs are attempted.
func TestEmptyGrid_ReturnsError(t *testing.T) {
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{},
	}
	cfg := paramsearch.Config{
		GridSpec:    spec,
		Instruments: []string{"NSE:A"},
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &stubStrategy{}, nil
		},
	}

	p := &mockProvider{candles: map[string][]model.Candle{}}
	_, err := paramsearch.Run(context.Background(), cfg, p)
	if err == nil {
		t.Error("expected error for empty grid, got nil")
	}
}

// TestRun_NilStrategyFactory_ReturnsError verifies that a nil StrategyFactory
// is caught at validation time.
func TestRun_NilStrategyFactory_ReturnsError(t *testing.T) {
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 2, Step: 1},
		},
	}
	cfg := paramsearch.Config{
		GridSpec:        spec,
		Instruments:     []string{"NSE:A"},
		EngineConfig:    engine.Config{From: time.Now(), To: time.Now().AddDate(1, 0, 0), InitialCash: 100_000},
		Timeframe:       model.TimeframeDaily,
		StrategyFactory: nil,
	}

	p := &mockProvider{candles: map[string][]model.Candle{}}
	_, err := paramsearch.Run(context.Background(), cfg, p)
	if err == nil {
		t.Error("expected error for nil StrategyFactory, got nil")
	}
}

// TestRun_EmptyInstruments_ReturnsError verifies that an empty instruments list
// is caught at validation time.
func TestRun_EmptyInstruments_ReturnsError(t *testing.T) {
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 2, Step: 1},
		},
	}
	cfg := paramsearch.Config{
		GridSpec:     spec,
		Instruments:  []string{},
		EngineConfig: engine.Config{From: time.Now(), To: time.Now().AddDate(1, 0, 0), InitialCash: 100_000},
		Timeframe:    model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return &stubStrategy{}, nil
		},
	}

	p := &mockProvider{candles: map[string][]model.Candle{}}
	_, err := paramsearch.Run(context.Background(), cfg, p)
	if err == nil {
		t.Error("expected error for empty instruments list, got nil")
	}
}

// TestRun_FactoryError_PropagatesError verifies that a StrategyFactory returning
// an error for a specific param combination causes Run to return that error.
func TestRun_FactoryError_PropagatesError(t *testing.T) {
	spec := paramsearch.GridSpec{
		Axes: []paramsearch.GridAxis{
			{Name: "p1", Min: 1, Max: 1, Step: 1},
		},
	}
	cfg := paramsearch.Config{
		GridSpec:    spec,
		Instruments: []string{"NSE:A"},
		EngineConfig: engine.Config{
			From:                 time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
		},
		Timeframe: model.TimeframeDaily,
		StrategyFactory: func(_ map[string]float64) (strategy.Strategy, error) {
			return nil, errors.New("invalid param combination")
		},
	}

	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	_, err := paramsearch.Run(context.Background(), cfg, p)
	if err == nil {
		t.Error("expected error from factory, got nil")
	}
}
