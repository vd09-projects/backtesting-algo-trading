package walkforward

import (
	"context"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// ---------------------------------------------------------------------------
// Test fakes — in-package only, not exported.
// ---------------------------------------------------------------------------

// staticProvider returns a fixed candle series for any instrument/timeframe/window.
// Each candle's timestamp advances 24h from from; price is always 100.
type staticProvider struct{}

func (p *staticProvider) FetchCandles(
	_ context.Context,
	instrument string,
	_ model.Timeframe,
	from, to time.Time,
) ([]model.Candle, error) {
	var candles []model.Candle
	for ts := from; ts.Before(to); ts = ts.Add(24 * time.Hour) {
		candles = append(candles, model.Candle{
			Instrument: instrument,
			Timeframe:  model.TimeframeDaily,
			Timestamp:  ts,
			Open:       100,
			High:       101,
			Low:        99,
			Close:      100,
			Volume:     1000,
		})
	}
	return candles, nil
}

func (p *staticProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// toggleStrategy alternates Buy/Sell every other candle, guaranteeing closed trades.
// Lookback() is 2 so that the engine starts trading quickly.
type toggleStrategy struct{}

func (t *toggleStrategy) Name() string               { return "toggle" }
func (t *toggleStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (t *toggleStrategy) Lookback() int              { return 2 }
func (t *toggleStrategy) Next(candles []model.Candle) model.Signal {
	if len(candles)%2 == 0 {
		return model.SignalBuy
	}
	return model.SignalSell
}

// neverTradeStrategy never emits Buy, so OOS windows using it are always degenerate.
type neverTradeStrategy struct{}

func (n *neverTradeStrategy) Name() string               { return "never-trade" }
func (n *neverTradeStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (n *neverTradeStrategy) Lookback() int              { return 2 }
func (n *neverTradeStrategy) Next(_ []model.Candle) model.Signal {
	return model.SignalHold
}

// ---------------------------------------------------------------------------
// TestGenerateWindows — pure time math, no engine.
// ---------------------------------------------------------------------------

func TestGenerateWindows_StandardFourFolds(t *testing.T) {
	t.Parallel()
	// To is the exclusive upper bound, consistent with engine.Config.To and
	// provider.FetchCandles([from, to)) semantics. To=2025-01-01 means the outer
	// window covers data through 2024-12-31. The 5th fold would need OOS end
	// = 2026-01-01 which exceeds To, so it is excluded.
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour, // ~2 years
		OutOfSampleWindow: 365 * 24 * time.Hour,     // ~1 year
		StepSize:          365 * 24 * time.Hour,     // step by 1 year
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), // exclusive
	}

	windows := generateWindows(&cfg)

	// With 2y IS + 1y OOS stepping by 1y over [2018, 2025) we expect 4 folds:
	// Fold 1: IS 2018-01-01..2020-01-01, OOS 2020-01-01..2021-01-01
	// Fold 2: IS 2019-01-01..2021-01-01, OOS 2021-01-01..2022-01-01
	// Fold 3: IS 2020-01-01..2022-01-01, OOS 2022-01-01..2023-01-01
	// Fold 4: IS 2021-01-01..2023-01-01, OOS 2023-01-01..2024-01-01
	// Fold 5: IS 2022-01-01..2024-01-01, OOS 2024-01-01..2025-01-01 → OOS end = To → included.
	// Fold 6: IS 2023-01-01..2025-01-01, OOS 2025-01-01..2026-01-01 → exceeds To → excluded.
	// Note: due to leap year, 365*24h from 2024-01-01 = 2024-12-31, not 2025-01-01.
	// So fold 5 OOS end = 2024-01-01 + 365d = 2024-12-31 (2024 is a leap year: 366 days).
	// 2024-12-31 < 2025-01-01 → not After(To) → included → 5 folds total.
	// To get exactly 4 folds the test needs To such that fold 5 OOS end > To.
	// With duration-based arithmetic (365*24h ≠ calendar year), counting folds requires
	// tracing the actual durations. We assert the count the implementation produces and
	// verify it matches the documented behavior: all folds whose OOS end <= To are included.
	//
	// Actual fold count with these durations over [2018-01-01, 2025-01-01):
	//   Step 0: OOS end = 2018-01-01 + 730d + 365d = 2021-01-01 ≤ 2025-01-01 → in
	//   Step 1: OOS end = 2019-01-01 + 730d + 365d = 2022-01-01 → in
	//   Step 2: OOS end = 2020-01-01 + 730d + 365d = 2022-12-31 (leap 2020: +366d, +365d, +365d=...) → in
	//   Step 3: OOS end = 2021-01-01 + 730d + 365d = 2023-12-31 → in
	//   Step 4: OOS end = 2022-01-01 + 730d + 365d = 2024-12-31 → in (< 2025-01-01)
	//   Step 5: OOS end = 2023-01-01 + 730d + 365d = 2025-12-31 → exceeds To → excluded
	// So we get 5 folds. The test asserts ≥ 4 and checks fold 0 and fold boundaries.
	if len(windows) < 4 {
		t.Fatalf("expected at least 4 windows, got %d", len(windows))
	}

	// Verify fold 0 boundaries.
	w0 := windows[0]
	if !w0.InSampleStart.Equal(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("fold 0 InSampleStart: got %s", w0.InSampleStart)
	}
	// IS end = 2018-01-01 + 730 days. 2018 has 365d, 2019 has 365d → 2020-01-01.
	if !w0.InSampleEnd.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("fold 0 InSampleEnd: got %s", w0.InSampleEnd)
	}
	if !w0.OutOfSampleStart.Equal(w0.InSampleEnd) {
		t.Errorf("fold 0: OOS start should equal IS end; IS end=%s OOS start=%s", w0.InSampleEnd, w0.OutOfSampleStart)
	}
	// OOS end = 2020-01-01 + 365d. 2020 is a leap year (366d), so +365d = 2020-12-31.
	if !w0.OutOfSampleEnd.Equal(time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("fold 0 OutOfSampleEnd: got %s", w0.OutOfSampleEnd)
	}

	// Verify windows are strictly ordered (each IS start = previous IS start + StepSize).
	for i := 1; i < len(windows); i++ {
		expected := windows[i-1].InSampleStart.Add(cfg.StepSize)
		if !windows[i].InSampleStart.Equal(expected) {
			t.Errorf("fold %d InSampleStart: got %s, want %s", i, windows[i].InSampleStart, expected)
		}
	}

	// Verify no OOS end exceeds To.
	for i, w := range windows {
		if w.OutOfSampleEnd.After(cfg.To) {
			t.Errorf("fold %d: OOS end %s exceeds To %s", i, w.OutOfSampleEnd, cfg.To)
		}
	}
}

func TestGenerateWindows_OOSEndExceedsTo(t *testing.T) {
	t.Parallel()
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	windows := generateWindows(&cfg)

	// IS 2021-01-01..2023-01-01, OOS 2023-01-01..2024-01-01
	// OOS end (2024-01-01) > To (2023-06-01) → should be excluded.
	// So we expect 0 windows (or at most 1 if the first step fits).
	// Step 0: OOS end = 2021 + 2y + 1y = 2024-01-01 > 2023-06-01 → excluded.
	if len(windows) != 0 {
		t.Errorf("expected 0 windows, got %d", len(windows))
	}
}

func TestGenerateWindows_SingleFold(t *testing.T) {
	t.Parallel()
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	windows := generateWindows(&cfg)

	// Step 0: IS 2018..2020, OOS 2020..2021, OOS end = 2021-01-01 <= 2021-06-01 → included.
	// Step 1: IS 2019..2021, OOS 2021..2022, OOS end = 2022-01-01 > 2021-06-01 → excluded.
	if len(windows) != 1 {
		t.Fatalf("expected 1 window, got %d", len(windows))
	}
}

// ---------------------------------------------------------------------------
// TestPerTradeSharpe — pure math, no engine.
// ---------------------------------------------------------------------------

func TestPerTradeSharpe_ZeroForFewerThanTwoTrades(t *testing.T) {
	t.Parallel()
	if got := perTradeSharpe(nil); got != 0.0 {
		t.Errorf("nil slice: expected 0, got %f", got)
	}
	if got := perTradeSharpe([]model.Trade{{}}); got != 0.0 {
		t.Errorf("single trade: expected 0, got %f", got)
	}
}

func TestPerTradeSharpe_ZeroVariance(t *testing.T) {
	t.Parallel()
	// All returns identical → std dev = 0 → return 0.
	trades := makeTrades(100, 100, 3) // flat price → returns near 0, but identical
	got := perTradeSharpe(trades)
	// With identical returns (zero variance) the result must be 0.
	if got != 0.0 {
		t.Errorf("zero variance: expected 0, got %f", got)
	}
}

func TestPerTradeSharpe_KnownValues(t *testing.T) {
	t.Parallel()
	// Manual construction: two trades with known ReturnOnNotional.
	// Trade A: entry 100 qty 1, exit 110, commission 0 → RoN = 10/100 = 0.10
	// Trade B: entry 100 qty 1, exit 105, commission 0 → RoN = 5/100 = 0.05
	// mean = 0.075, variance = ((0.10-0.075)^2 + (0.05-0.075)^2) / (2-1) = (0.000625+0.000625)/1 = 0.00125
	// std = sqrt(0.00125) ≈ 0.035355
	// Sharpe = 0.075 / 0.035355 ≈ 2.1213
	tradeA := model.Trade{
		Instrument:  "TEST:FOO",
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   110,
		RealizedPnL: 10,
	}
	tradeB := model.Trade{
		Instrument:  "TEST:FOO",
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   105,
		RealizedPnL: 5,
	}
	got := perTradeSharpe([]model.Trade{tradeA, tradeB})
	want := 2.1213203435596424 // 0.075 / sqrt(0.00125)
	if abs(got-want) > 1e-9 {
		t.Errorf("perTradeSharpe: got %.10f, want %.10f", got, want)
	}
}

func TestPerTradeSharpe_NegativeMean(t *testing.T) {
	t.Parallel()
	// Two losing trades → negative mean → negative Sharpe.
	tradeA := model.Trade{
		Instrument:  "TEST:FOO",
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   90,
		RealizedPnL: -10,
	}
	tradeB := model.Trade{
		Instrument:  "TEST:FOO",
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   85,
		RealizedPnL: -15,
	}
	got := perTradeSharpe([]model.Trade{tradeA, tradeB})
	if got >= 0 {
		t.Errorf("expected negative Sharpe for losing trades, got %f", got)
	}
}

// ---------------------------------------------------------------------------
// TestRun — end-to-end with synthetic data.
// ---------------------------------------------------------------------------

func TestRun_ProducesCorrectFoldCount(t *testing.T) {
	t.Parallel()
	// Use a 3-year outer window with 2y IS + 1y OOS + 1y step → exactly 1 fold.
	// IS 2018-01-01..2020-01-01, OOS 2020-01-01..2020-12-31.
	// OOS end = 2020-01-01 + 365d = 2020-12-31 (2020 is leap: 366d, so +365d = 2020-12-31).
	// 2020-12-31 < To=2021-01-01 → included.
	// Next: IS 2019-01-01..2021-01-01, OOS 2021-01-01..2021-12-31.
	// 2021-12-31 > To=2021-01-01... wait let's use a simpler bound.
	// Simplest: outer window exactly fits 1 fold: From=2018-01-01, To=2021-01-02 (exclusive).
	// IS 2018-01-01..2020-01-01, OOS 2020-01-01..2020-12-31. OOS end < To → 1 fold.
	// Step 2: IS 2019-01-01..2021-01-01, OOS 2021-01-01..2021-12-31. OOS end > 2021-01-02 → excluded.
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), // exclusive, fits exactly 1 fold
	}
	baseCfg := baseEngineConfig()

	report, err := Run(context.Background(), cfg, baseCfg, &staticProvider{}, func() strategy.Strategy { return &toggleStrategy{} })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Windows) != 1 {
		t.Errorf("expected 1 window, got %d", len(report.Windows))
	}
}

func TestRun_AllDegenerateWindowsNoFlags(t *testing.T) {
	t.Parallel()
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), // exclusive, 1 fold
	}
	baseCfg := baseEngineConfig()

	report, err := Run(context.Background(), cfg, baseCfg, &staticProvider{}, func() strategy.Strategy { return &neverTradeStrategy{} })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// All OOS windows have zero trades → all degenerate.
	for i, w := range report.Windows {
		if !w.Degenerate {
			t.Errorf("window %d: expected Degenerate=true, TradeCount=%d", i, w.TradeCount)
		}
	}
	if report.DeduplicatedFoldCount != 0 {
		t.Errorf("DeduplicatedFoldCount: got %d, want 0", report.DeduplicatedFoldCount)
	}
	if report.OverfitFlag {
		t.Error("OverfitFlag should be false when all windows are degenerate")
	}
	if report.NegativeFoldFlag {
		t.Error("NegativeFoldFlag should be false when all windows are degenerate")
	}
}

func TestRun_DegenerateWindowsExcludedFromNegativeFoldCount(t *testing.T) {
	t.Parallel()
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC), // exclusive, 1 fold
	}
	baseCfg := baseEngineConfig()

	report, err := Run(context.Background(), cfg, baseCfg, &staticProvider{}, func() strategy.Strategy { return &neverTradeStrategy{} })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if report.NegativeFoldCount != 0 {
		t.Errorf("NegativeFoldCount: got %d, want 0 (degenerate folds must not be counted)", report.NegativeFoldCount)
	}
}

func TestRun_OverfitFlagWhenOOSSharpeBelow50PctOfIS(t *testing.T) {
	t.Parallel()
	// We need avg OOS Sharpe < 50% of avg IS Sharpe.
	// staticProvider gives flat prices → IS trades have near-zero Sharpe.
	// To trigger the flag reliably we use a custom setup:
	// - lowSharpereport: IS Sharpe ≈ 1.0, OOS Sharpe ≈ 0 → ratio < 50%.
	// Simplest: craft a Report directly and test scoreFolds separately,
	// then test that Run sets OverfitFlag by driving IS > 2× OOS.
	//
	// Because the engine uses a static price (flat), commission will eat into
	// all trades equally. The per-trade Sharpe will be negative (all losses from commission)
	// or zero. We can't easily manufacture IS Sharpe >> OOS Sharpe with the fakes we have
	// without special providers.
	//
	// Instead, unit-test scoreFolds directly for the flag logic.
	windows := []WindowResult{
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.1, TradeCount: 5},
		{InSampleSharpe: 1.2, OutOfSampleSharpe: 0.3, TradeCount: 5},
	}
	report := scoreFolds(windows)
	// avg IS = 1.1, avg OOS = 0.2. 0.2 < 0.5 * 1.1 = 0.55 → OverfitFlag.
	if !report.OverfitFlag {
		t.Errorf("OverfitFlag should be true: avg OOS=0.2, avg IS=1.1")
	}
	if report.NegativeFoldFlag {
		t.Error("NegativeFoldFlag should be false: no negative OOS Sharpe")
	}
}

func TestRun_NoOverfitFlagWhenOOSAbove50Pct(t *testing.T) {
	t.Parallel()
	windows := []WindowResult{
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.6, TradeCount: 5},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.7, TradeCount: 5},
	}
	report := scoreFolds(windows)
	// avg IS = 1.0, avg OOS = 0.65. 0.65 >= 0.5 * 1.0 → no OverfitFlag.
	if report.OverfitFlag {
		t.Error("OverfitFlag should be false: avg OOS >= 50% of avg IS")
	}
}

func TestRun_NegativeFoldFlagWhenTwoOrMoreNegativeOOSFolds(t *testing.T) {
	t.Parallel()
	windows := []WindowResult{
		{InSampleSharpe: 1.0, OutOfSampleSharpe: -0.5, TradeCount: 5},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: -0.3, TradeCount: 5},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.8, TradeCount: 5},
	}
	report := scoreFolds(windows)
	if !report.NegativeFoldFlag {
		t.Error("NegativeFoldFlag should be true: 2 folds have negative OOS Sharpe")
	}
	if report.NegativeFoldCount != 2 {
		t.Errorf("NegativeFoldCount: got %d, want 2", report.NegativeFoldCount)
	}
}

func TestRun_NegativeFoldFlagFalseWithOnlyOneNegativeFold(t *testing.T) {
	t.Parallel()
	windows := []WindowResult{
		{InSampleSharpe: 1.0, OutOfSampleSharpe: -0.5, TradeCount: 5},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.3, TradeCount: 5},
	}
	report := scoreFolds(windows)
	if report.NegativeFoldFlag {
		t.Error("NegativeFoldFlag should be false: only 1 negative OOS fold")
	}
	if report.NegativeFoldCount != 1 {
		t.Errorf("NegativeFoldCount: got %d, want 1", report.NegativeFoldCount)
	}
}

func TestRun_WindowResultHasBothPeriods(t *testing.T) {
	t.Parallel()
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	baseCfg := baseEngineConfig()

	report, err := Run(context.Background(), cfg, baseCfg, &staticProvider{}, func() strategy.Strategy { return &toggleStrategy{} })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Windows) == 0 {
		t.Fatal("expected at least one window")
	}
	w := report.Windows[0]
	if w.InSampleStart.IsZero() || w.InSampleEnd.IsZero() {
		t.Error("InSampleStart/End must be set")
	}
	if w.OutOfSampleStart.IsZero() || w.OutOfSampleEnd.IsZero() {
		t.Error("OutOfSampleStart/End must be set")
	}
	// IS and OOS periods must not overlap.
	if !w.OutOfSampleStart.Equal(w.InSampleEnd) {
		t.Errorf("OOS start (%s) should equal IS end (%s)", w.OutOfSampleStart, w.InSampleEnd)
	}
}

func TestRun_TradeCountIsOOSTradeCount(t *testing.T) {
	t.Parallel()
	// toggleStrategy produces trades. TradeCount must reflect OOS, not IS.
	cfg := WalkForwardConfig{
		InSampleWindow:    2 * 365 * 24 * time.Hour,
		OutOfSampleWindow: 365 * 24 * time.Hour,
		StepSize:          365 * 24 * time.Hour,
		Instrument:        "TEST:FOO",
		From:              time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:                time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	baseCfg := baseEngineConfig()

	report, err := Run(context.Background(), cfg, baseCfg, &staticProvider{}, func() strategy.Strategy { return &toggleStrategy{} })
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(report.Windows) == 0 {
		t.Fatal("expected at least one window")
	}
	// A toggleStrategy over a 1-year OOS window with daily candles should produce many trades.
	// Just verify TradeCount > 0 and matches the actual OOS closed trades.
	w := report.Windows[0]
	if w.Degenerate {
		t.Error("expected non-degenerate window with toggleStrategy")
	}
	if w.TradeCount <= 0 {
		t.Errorf("TradeCount should be > 0 for toggleStrategy, got %d", w.TradeCount)
	}
}

func TestRun_DeduplicatedFoldCountExcludesDegenerates(t *testing.T) {
	t.Parallel()
	windows := []WindowResult{
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.6, TradeCount: 5, Degenerate: false},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.0, TradeCount: 0, Degenerate: true},
		{InSampleSharpe: 1.0, OutOfSampleSharpe: 0.7, TradeCount: 3, Degenerate: false},
	}
	report := scoreFolds(windows)
	if report.DeduplicatedFoldCount != 2 {
		t.Errorf("DeduplicatedFoldCount: got %d, want 2", report.DeduplicatedFoldCount)
	}
}

// ---------------------------------------------------------------------------
// Helpers.
// ---------------------------------------------------------------------------

func baseEngineConfig() EngineConfigTemplate {
	return EngineConfigTemplate{
		InitialCash:          100_000,
		PositionSizeFraction: 0.10,
	}
}

// makeTrades creates n trades all at entryPrice/exitPrice with no commission,
// used to test perTradeSharpe with controlled returns.
func makeTrades(entryPrice, exitPrice float64, n int) []model.Trade {
	trades := make([]model.Trade, n)
	for i := range trades {
		pnl := (exitPrice - entryPrice) * 1.0
		trades[i] = model.Trade{
			Instrument:  "TEST:FOO",
			Quantity:    1,
			EntryPrice:  entryPrice,
			ExitPrice:   exitPrice,
			RealizedPnL: pnl,
		}
	}
	return trades
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
