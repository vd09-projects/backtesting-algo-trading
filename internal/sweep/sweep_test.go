package sweep

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// --- Fakes ---

// staticProvider satisfies provider.DataProvider and always returns the same candles,
// ignoring all arguments. Used to inject deterministic data without network access.
type staticProvider struct {
	candles []model.Candle
}

func (s *staticProvider) FetchCandles(
	_ context.Context, _ string, _ model.Timeframe, _, _ time.Time,
) ([]model.Candle, error) {
	return s.candles, nil
}

func (s *staticProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// thresholdStrategy emits Buy when the most recent close strictly exceeds threshold,
// and Sell otherwise. Lookback is 1 (operates on any single bar).
type thresholdStrategy struct {
	threshold float64
	tf        model.Timeframe
}

func (t *thresholdStrategy) Name() string               { return fmt.Sprintf("threshold-%.0f", t.threshold) }
func (t *thresholdStrategy) Timeframe() model.Timeframe { return t.tf }
func (t *thresholdStrategy) Lookback() int              { return 1 }
func (t *thresholdStrategy) Next(candles []model.Candle) model.Signal {
	if candles[len(candles)-1].Close > t.threshold {
		return model.SignalBuy
	}
	return model.SignalSell
}

// invertedThresholdStrategy emits Sell when the most recent close strictly exceeds
// threshold, and Buy otherwise. This is the opposite of thresholdStrategy, and
// produces losing trades on alternating-candle series: it buys at high prices
// (next bar open is low) and sells at low prices (next bar open is high).
type invertedThresholdStrategy struct {
	threshold float64
	tf        model.Timeframe
}

func (s *invertedThresholdStrategy) Name() string               { return fmt.Sprintf("inverted-%.0f", s.threshold) }
func (s *invertedThresholdStrategy) Timeframe() model.Timeframe { return s.tf }
func (s *invertedThresholdStrategy) Lookback() int              { return 1 }
func (s *invertedThresholdStrategy) Next(candles []model.Candle) model.Signal {
	if candles[len(candles)-1].Close > s.threshold {
		return model.SignalSell
	}
	return model.SignalBuy
}

// makeAlternatingCandles returns n daily candles alternating between highClose and
// lowClose, starting with highClose on bar 0. All OHLC fields equal the close so
// engine fills at Open are fully deterministic.
func makeAlternatingCandles(n int, highClose, lowClose float64) []model.Candle {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]model.Candle, n)
	for i := range candles {
		c := highClose
		if i%2 == 1 {
			c = lowClose
		}
		candles[i] = model.Candle{
			Instrument: "TEST:X",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       c,
			High:       c,
			Low:        c,
			Close:      c,
			Volume:     1000,
		}
	}
	return candles
}

// testEngineConfig returns a minimal engine config for sweep tests.
func testEngineConfig() engine.Config {
	return engine.Config{
		Instrument:           "TEST:X",
		From:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		InitialCash:          100000,
		PositionSizeFraction: 0.1,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	}
}

// --- Config validation tests ---

func TestValidateConfig(t *testing.T) {
	t.Parallel()
	validFactory := func(float64) (strategy.Strategy, error) {
		return &thresholdStrategy{threshold: 100, tf: model.TimeframeDaily}, nil
	}
	base := Config{
		ParameterName:   "threshold",
		Min:             80,
		Max:             130,
		Step:            10,
		Timeframe:       model.TimeframeDaily,
		EngineConfig:    testEngineConfig(),
		StrategyFactory: validFactory,
	}

	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr string // substring expected in error; empty means no error
	}{
		{"valid", func(*Config) {}, ""},
		{"min equals max", func(c *Config) { c.Min = 100; c.Max = 100 }, ""}, // single step is valid
		{"empty parameter name", func(c *Config) { c.ParameterName = "" }, "ParameterName"},
		{"zero step", func(c *Config) { c.Step = 0 }, "Step"},
		{"negative step", func(c *Config) { c.Step = -1 }, "Step"},
		{"max less than min", func(c *Config) { c.Max = 70 }, "Max"},
		{"nil factory", func(c *Config) { c.StrategyFactory = nil }, "StrategyFactory"},
		{"empty timeframe", func(c *Config) { c.Timeframe = "" }, "Timeframe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := base
			tt.modify(&cfg)
			err := validateConfig(cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("validateConfig: unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("validateConfig: expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("validateConfig: error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// --- paramSteps tests ---

func TestParamSteps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		min, max, step float64
		want           []float64
	}{
		{"single step", 10, 10, 5, []float64{10}},
		{"three integer steps", 10, 20, 5, []float64{10, 15, 20}},
		{"fractional step", 0.1, 0.3, 0.1, []float64{0.1, 0.2, 0.3}},
		{"step divides range exactly", 1, 10, 3, []float64{1, 4, 7, 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := paramSteps(tt.min, tt.max, tt.step)
			if len(got) != len(tt.want) {
				t.Fatalf("len: got %d %v, want %d %v", len(got), got, len(tt.want), tt.want)
			}
			for i, v := range got {
				if math.Abs(v-tt.want[i]) > 1e-9 {
					t.Errorf("step[%d]: got %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

// --- Plateau unit tests ---

func TestComputePlateau(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		results []Result // must be sorted Sharpe descending for realistic input
		want    *PlateauRange
	}{
		{
			name:    "empty",
			results: nil,
			want:    nil,
		},
		{
			name: "all non-positive sharpe",
			results: []Result{
				{ParamValue: 10, SharpeRatio: 0},
				{ParamValue: 20, SharpeRatio: -0.5},
			},
			want: nil,
		},
		{
			name: "single result above threshold",
			results: []Result{
				{ParamValue: 14, SharpeRatio: 1.0},
			},
			want: &PlateauRange{MinParam: 14, MaxParam: 14, Count: 1, MinSharpe: 1.0},
		},
		{
			// Peak=1.0, threshold=0.8. Qualifying: Sharpe 1.0, 0.85, 0.82 (ParamValues 14, 13, 15).
			// ParamValue 12 (Sharpe 0.70) is below threshold.
			name: "clear plateau of three",
			results: []Result{
				{ParamValue: 14, SharpeRatio: 1.0},
				{ParamValue: 13, SharpeRatio: 0.85},
				{ParamValue: 15, SharpeRatio: 0.82},
				{ParamValue: 12, SharpeRatio: 0.70},
			},
			want: &PlateauRange{MinParam: 13, MaxParam: 15, Count: 3, MinSharpe: 0.82},
		},
		{
			name: "all results in plateau",
			results: []Result{
				{ParamValue: 10, SharpeRatio: 1.0},
				{ParamValue: 20, SharpeRatio: 0.95},
				{ParamValue: 30, SharpeRatio: 0.90},
			},
			want: &PlateauRange{MinParam: 10, MaxParam: 30, Count: 3, MinSharpe: 0.90},
		},
		{
			// ParamValues 10 (Sharpe 1.0) and 30 (Sharpe 0.85) qualify; ParamValue 20 (0.75) does not.
			// The plateau spans the min and max of qualifying values regardless of contiguity.
			name: "non-contiguous qualifying values",
			results: []Result{
				{ParamValue: 10, SharpeRatio: 1.0},
				{ParamValue: 30, SharpeRatio: 0.85},
				{ParamValue: 20, SharpeRatio: 0.75},
			},
			want: &PlateauRange{MinParam: 10, MaxParam: 30, Count: 2, MinSharpe: 0.85},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := computePlateau(tt.results)
			if tt.want == nil {
				if got != nil {
					t.Errorf("expected nil plateau, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil plateau, got nil")
			}
			if got.MinParam != tt.want.MinParam {
				t.Errorf("MinParam: got %.4f, want %.4f", got.MinParam, tt.want.MinParam)
			}
			if got.MaxParam != tt.want.MaxParam {
				t.Errorf("MaxParam: got %.4f, want %.4f", got.MaxParam, tt.want.MaxParam)
			}
			if got.Count != tt.want.Count {
				t.Errorf("Count: got %d, want %d", got.Count, tt.want.Count)
			}
			if math.Abs(got.MinSharpe-tt.want.MinSharpe) > 1e-9 {
				t.Errorf("MinSharpe: got %.9f, want %.9f", got.MinSharpe, tt.want.MinSharpe)
			}
		})
	}
}

// --- Integration golden test ---

// TestRun_IntegrationGolden verifies the sweep end-to-end against a deterministic synthetic setup.
//
// Setup: 300 alternating candles (high=120, low=80). The thresholdStrategy emits Buy when
// close > threshold and Sell otherwise. Because the engine fills signals at the NEXT bar's
// open price:
//
//   - High bar (close=120) → Buy signal → fill at next bar's open=80. Buys low.
//   - Low bar (close=80)   → Sell signal → fill at next bar's open=120. Sells high.
//
// For threshold in (80, 120] exclusive: every cycle the strategy buys at 80 and sells at 120.
// Profitable. threshold=80 is included because close=80 is NOT strictly > 80 (Sell emitted).
// For threshold ≤ 80: close=80 > threshold is true, so strategy never exits — stays long forever.
// Sharpe ≈ 0 (oscillating equity, no completed profitable trades).
// For threshold ≥ 120: close=120 is NOT > threshold — never enters. 0 trades. Sharpe = 0.
//
// 300 candles used so the equity curve exceeds MinCurvePointsForMetrics (252), allowing
// Sharpe to be reported rather than suppressed by the signal-frequency gate.
//
// Sweep: 60..130 step 10 → 8 values: 60, 70, 80, 90, 100, 110, 120, 130.
// Expected plateau: {80, 90, 100, 110} all produce identical profitable trade sequences.
func TestRun_IntegrationGolden(t *testing.T) { //nolint:cyclop // golden integration test; all assertions must remain co-located to be reviewable
	candles := makeAlternatingCandles(300, 120, 80)
	p := &staticProvider{candles: candles}

	factory := func(v float64) (strategy.Strategy, error) {
		return &thresholdStrategy{threshold: v, tf: model.TimeframeDaily}, nil
	}

	cfg := Config{
		ParameterName:   "threshold",
		Min:             60,
		Max:             130,
		Step:            10,
		Timeframe:       model.TimeframeDaily,
		EngineConfig:    testEngineConfig(),
		StrategyFactory: factory,
	}

	report, err := Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// 8 parameter values: 60, 70, 80, 90, 100, 110, 120, 130.
	if len(report.Results) != 8 {
		t.Fatalf("expected 8 results, got %d: %v", len(report.Results), report.Results)
	}

	// Results must be sorted descending by Sharpe.
	for i := 1; i < len(report.Results); i++ {
		if report.Results[i].SharpeRatio > report.Results[i-1].SharpeRatio {
			t.Errorf("results not sorted at [%d,%d]: %.4f > %.4f",
				i, i-1, report.Results[i].SharpeRatio, report.Results[i-1].SharpeRatio)
		}
	}

	// Top result must have a positive Sharpe (profitable range is [80, 110]).
	if report.Results[0].SharpeRatio <= 0 {
		t.Errorf("top result Sharpe = %.4f; expected positive (profitable range 80–110)",
			report.Results[0].SharpeRatio)
	}

	// Top result's ParamValue must be in the profitable range.
	top := report.Results[0].ParamValue
	if top < 80 || top > 110 {
		t.Errorf("top ParamValue %.0f not in profitable range [80, 110]", top)
	}

	// Thresholds 120 and 130 never enter — expect 0 trades and Sharpe=0.
	for _, r := range report.Results {
		if r.ParamValue == 120 || r.ParamValue == 130 {
			if r.TradeCount != 0 {
				t.Errorf("threshold %.0f: expected 0 trades, got %d", r.ParamValue, r.TradeCount)
			}
			if r.SharpeRatio != 0 {
				t.Errorf("threshold %.0f: expected Sharpe=0, got %.4f", r.ParamValue, r.SharpeRatio)
			}
		}
	}

	// Plateau must be non-nil and identify exactly the profitable range [80, 110].
	if report.Plateau == nil {
		t.Fatal("expected non-nil plateau, got nil")
	}
	if report.Plateau.MinParam != 80 {
		t.Errorf("Plateau.MinParam: got %.0f, want 80", report.Plateau.MinParam)
	}
	if report.Plateau.MaxParam != 110 {
		t.Errorf("Plateau.MaxParam: got %.0f, want 110", report.Plateau.MaxParam)
	}
	if report.Plateau.Count != 4 {
		t.Errorf("Plateau.Count: got %d, want 4 (values 80, 90, 100, 110)", report.Plateau.Count)
	}

	// All plateau members satisfy Sharpe >= 80% of peak.
	peakSharpe := report.Results[0].SharpeRatio
	floorSharpe := plateauThreshold * peakSharpe
	for _, r := range report.Results {
		if r.ParamValue >= 80 && r.ParamValue <= 110 {
			if r.SharpeRatio < floorSharpe {
				t.Errorf("ParamValue=%.0f Sharpe %.4f < plateau floor %.4f",
					r.ParamValue, r.SharpeRatio, floorSharpe)
			}
		}
	}

	// ParameterName is propagated to the report.
	if report.ParameterName != cfg.ParameterName {
		t.Errorf("ParameterName: got %q, want %q", report.ParameterName, cfg.ParameterName)
	}
}

// TestComputePlateauWithMinTrades tests computePlateau with a positive minTrades
// constraint — the valid region is restricted to entries with TradeCount >= minTrades.
func TestComputePlateauWithMinTrades(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		results   []Result
		minTrades int
		want      *PlateauRange
		wantNil   bool
	}{
		{
			// Global peak Sharpe=2.0 comes from a 5-trade entry (below minTrades=30).
			// Valid region contains only the 35-trade entry (Sharpe=1.0).
			// 80% of valid-region peak (1.0) = 0.8; only the 35-trade entry qualifies.
			name: "global peak below minTrades; valid region peak lower",
			results: []Result{
				{ParamValue: 10, SharpeRatio: 2.0, TradeCount: 5},  // invalid: < 30 trades
				{ParamValue: 20, SharpeRatio: 1.0, TradeCount: 35}, // valid
				{ParamValue: 30, SharpeRatio: 0.5, TradeCount: 40}, // valid but below 80% of 1.0
			},
			minTrades: 30,
			want:      &PlateauRange{MinParam: 20, MaxParam: 20, Count: 1, MinSharpe: 1.0},
		},
		{
			// All entries have TradeCount < 30. Valid region is empty.
			// Expect nil plateau with SensitivityConcern set.
			name: "no entry meets minTrades — empty valid region",
			results: []Result{
				{ParamValue: 10, SharpeRatio: 1.5, TradeCount: 5},
				{ParamValue: 20, SharpeRatio: 1.0, TradeCount: 10},
			},
			minTrades: 30,
			wantNil:   true,
		},
		{
			// Valid region exists but all valid Sharpes are negative.
			// Expect nil plateau.
			name: "valid region is all-negative Sharpe",
			results: []Result{
				{ParamValue: 10, SharpeRatio: -0.5, TradeCount: 35},
				{ParamValue: 20, SharpeRatio: -1.0, TradeCount: 40},
			},
			minTrades: 30,
			wantNil:   true,
		},
		{
			// Multiple entries in valid region; 80% floor applied to valid-region peak.
			// Valid region: Sharpe 2.0 (TradeCount=40), 1.8 (TradeCount=50), 1.0 (TradeCount=35).
			// Valid-region peak = 2.0; floor = 0.8*2.0 = 1.6.
			// Qualifying: Sharpe 2.0 (param=20), 1.8 (param=30). Not 1.0 (param=10, below 1.6).
			name: "plateau within valid region, floor from valid-region peak",
			results: []Result{
				{ParamValue: 20, SharpeRatio: 2.0, TradeCount: 40},
				{ParamValue: 30, SharpeRatio: 1.8, TradeCount: 50},
				{ParamValue: 10, SharpeRatio: 1.0, TradeCount: 35},
			},
			minTrades: 30,
			want:      &PlateauRange{MinParam: 20, MaxParam: 30, Count: 2, MinSharpe: 1.8},
		},
		{
			// minTrades=0 — equivalent to old behavior. Same as existing tests.
			// All entries qualify; global peak is used.
			name: "minTrades=0 behaves as no-filter (backward compatibility)",
			results: []Result{
				{ParamValue: 14, SharpeRatio: 1.0, TradeCount: 0},
				{ParamValue: 13, SharpeRatio: 0.85, TradeCount: 0},
				{ParamValue: 15, SharpeRatio: 0.82, TradeCount: 0},
				{ParamValue: 12, SharpeRatio: 0.70, TradeCount: 0},
			},
			minTrades: 0,
			want:      &PlateauRange{MinParam: 13, MaxParam: 15, Count: 3, MinSharpe: 0.82},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := computePlateauWithMinTrades(tt.results, tt.minTrades)
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil plateau, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected non-nil plateau, got nil")
			}
			if got.MinParam != tt.want.MinParam {
				t.Errorf("MinParam: got %.4f, want %.4f", got.MinParam, tt.want.MinParam)
			}
			if got.MaxParam != tt.want.MaxParam {
				t.Errorf("MaxParam: got %.4f, want %.4f", got.MaxParam, tt.want.MaxParam)
			}
			if got.Count != tt.want.Count {
				t.Errorf("Count: got %d, want %d", got.Count, tt.want.Count)
			}
			if math.Abs(got.MinSharpe-tt.want.MinSharpe) > 1e-9 {
				t.Errorf("MinSharpe: got %.9f, want %.9f", got.MinSharpe, tt.want.MinSharpe)
			}
		})
	}
}

// TestReport_SensitivityConcern verifies that Run sets SensitivityConcern on the Report
// when the sweep's valid region (TradeCount >= MinTradesForPlateau) is empty or all-negative.
// Uses a synthetic strategy that generates 0 trades — all results will have TradeCount=0,
// so with minTrades=30 the valid region is empty.
func TestReport_SensitivityConcern_EmptyValidRegion(t *testing.T) {
	t.Parallel()
	// Use a strategy that always holds (no trades) to produce TradeCount=0 for every param.
	// With minTrades=30, the valid region is empty — SensitivityConcern must be set.
	candles := makeAlternatingCandles(300, 120, 80)
	p := &staticProvider{candles: candles}

	// A strategy that never trades (always returns Hold).
	factory := func(_ float64) (strategy.Strategy, error) {
		return &thresholdStrategy{threshold: 200, tf: model.TimeframeDaily}, nil // threshold above any close
	}

	cfg := Config{
		ParameterName:   "threshold",
		Min:             10,
		Max:             20,
		Step:            10,
		Timeframe:       model.TimeframeDaily,
		EngineConfig:    testEngineConfig(),
		StrategyFactory: factory,
	}

	report, err := Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// With MinTradesForPlateau=30, all TradeCount=0 entries are invalid.
	// Plateau must be nil and SensitivityConcern must be set.
	if report.Plateau != nil {
		t.Errorf("expected nil Plateau when valid region is empty, got %+v", report.Plateau)
	}
	if report.SensitivityConcern == "" {
		t.Error("expected SensitivityConcern to be set when valid region is empty, got empty string")
	}
}

// TestReport_SensitivityConcern_AllNegativeValidRegion verifies that Run sets
// SensitivityConcern when the valid region (TradeCount >= MinTradesForPlateau) exists
// but every entry in it has non-positive Sharpe.
//
// Setup: invertedThresholdStrategy on alternating candles (high=120, low=80).
// The engine fills at the next bar's open price:
//   - Low bar (close=80 ≤ threshold=100) → Buy signal → fills at next bar's open=120. Buys at 120.
//   - High bar (close=120 > threshold=100) → Sell signal → fills at next bar's open=80. Sells at 80.
//
// Each round-trip: buy at 120, sell at 80 → loss. 300 candles produce ~150 trades
// (well above MinTradesForPlateau=30), all losing. SensitivityConcern must be set.
func TestReport_SensitivityConcern_AllNegativeValidRegion(t *testing.T) {
	t.Parallel()
	candles := makeAlternatingCandles(300, 120, 80)
	p := &staticProvider{candles: candles}

	// invertedThresholdStrategy with threshold=100 on alternating candles:
	// buys at 120 (on low bars), sells at 80 (on high bars) — always loses.
	factory := func(_ float64) (strategy.Strategy, error) {
		return &invertedThresholdStrategy{threshold: 100, tf: model.TimeframeDaily}, nil
	}

	cfg := Config{
		ParameterName:   "threshold",
		Min:             10,
		Max:             20,
		Step:            10,
		Timeframe:       model.TimeframeDaily,
		EngineConfig:    testEngineConfig(),
		StrategyFactory: factory,
	}

	report, err := Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// All valid-region entries have negative Sharpe — Plateau must be nil.
	if report.Plateau != nil {
		t.Errorf("expected nil Plateau for all-negative valid region, got %+v", report.Plateau)
	}

	// At least one entry must have TradeCount >= MinTradesForPlateau to confirm
	// the valid region is non-empty (the !hasPositive branch, not !hasValid).
	hasValidEntry := false
	for _, r := range report.Results {
		if r.TradeCount >= MinTradesForPlateau {
			hasValidEntry = true
			break
		}
	}
	if !hasValidEntry {
		t.Fatalf("no result has TradeCount >= %d; test setup is wrong — invertedThresholdStrategy should generate many losing trades", MinTradesForPlateau)
	}

	// SensitivityConcern must be the all-negative-region message, not the empty-region message.
	if report.SensitivityConcern == "" {
		t.Error("expected SensitivityConcern to be set for all-negative valid region, got empty string")
	}
	if report.SensitivityConcern == "no parameter achieves >= 30 trades in sweep range" {
		t.Errorf("SensitivityConcern is the wrong message (empty-region): got %q; expected the all-negative message", report.SensitivityConcern)
	}
}

// TestRun_FactoryError verifies that a factory error stops the sweep and surfaces the
// failing parameter value in the returned error message.
func TestRun_FactoryError(t *testing.T) {
	p := &staticProvider{candles: makeAlternatingCandles(20, 120, 80)}
	cfg := Config{
		ParameterName: "threshold",
		Min:           80,
		Max:           100,
		Step:          10,
		Timeframe:     model.TimeframeDaily,
		EngineConfig:  testEngineConfig(),
		StrategyFactory: func(v float64) (strategy.Strategy, error) {
			return nil, fmt.Errorf("injected factory error")
		},
	}
	_, err := Run(context.Background(), cfg, p)
	if err == nil {
		t.Fatal("expected error from factory, got nil")
	}
	if !strings.Contains(err.Error(), "factory") {
		t.Errorf("error %q does not mention factory context", err.Error())
	}
}
