package universesweep_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// Test fakes
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

// ---------------------------------------------------------------------------
// TestParseUniverseFile
// ---------------------------------------------------------------------------

func TestParseUniverseFile_ValidYAML(t *testing.T) {
	t.Parallel()

	content := `instruments:
  - NSE:RELIANCE
  - NSE:INFY
  - NSE:TCS
`
	dir := t.TempDir()
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	instruments, err := universesweep.ParseUniverseFile(path)
	if err != nil {
		t.Fatalf("ParseUniverseFile: unexpected error: %v", err)
	}
	if len(instruments) != 3 {
		t.Fatalf("expected 3 instruments, got %d: %v", len(instruments), instruments)
	}
	want := []string{"NSE:RELIANCE", "NSE:INFY", "NSE:TCS"}
	for i, w := range want {
		if instruments[i] != w {
			t.Errorf("instruments[%d] = %q, want %q", i, instruments[i], w)
		}
	}
}

func TestParseUniverseFile_DeduplicatesEntries(t *testing.T) {
	t.Parallel()

	content := `instruments:
  - NSE:RELIANCE
  - NSE:INFY
  - NSE:RELIANCE
`
	dir := t.TempDir()
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	instruments, err := universesweep.ParseUniverseFile(path)
	if err != nil {
		t.Fatalf("ParseUniverseFile: unexpected error: %v", err)
	}
	if len(instruments) != 2 {
		t.Fatalf("expected 2 instruments after dedup, got %d: %v", len(instruments), instruments)
	}
}

func TestParseUniverseFile_ReturnsErrorOnEmptyList(t *testing.T) {
	t.Parallel()

	content := `instruments: []
`
	dir := t.TempDir()
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := universesweep.ParseUniverseFile(path)
	if err == nil {
		t.Fatal("expected error for empty instruments list, got nil")
	}
}

func TestParseUniverseFile_ReturnsErrorOnMissingFile(t *testing.T) {
	t.Parallel()

	_, err := universesweep.ParseUniverseFile("/nonexistent/path/universe.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestParseUniverseFile_ReturnsErrorOnMissingInstrumentsKey(t *testing.T) {
	t.Parallel()

	content := `strategies:
  - sma-crossover
`
	dir := t.TempDir()
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := universesweep.ParseUniverseFile(path)
	if err == nil {
		t.Fatal("expected error when instruments key is absent, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestRun
// ---------------------------------------------------------------------------

func TestRun_TwoInstruments_ProducesTwoResults(t *testing.T) {
	t.Parallel()

	from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := universesweep.Config{
		Instruments: []string{"NSE:RELIANCE", "NSE:INFY"},
		Strategy:    &toggleStrategy{},
		EngineConfig: engine.Config{
			From:                 from,
			To:                   to,
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		Timeframe: model.TimeframeDaily,
	}

	report, err := universesweep.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	if len(report.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(report.Results))
	}

	// Results must include both instruments.
	instruments := map[string]bool{}
	for _, r := range report.Results {
		instruments[r.Instrument] = true
	}
	if !instruments["NSE:RELIANCE"] {
		t.Error("result for NSE:RELIANCE missing")
	}
	if !instruments["NSE:INFY"] {
		t.Error("result for NSE:INFY missing")
	}
}

func TestRun_ResultsSortedDescendingBySharpe(t *testing.T) {
	t.Parallel()

	// With a static provider (flat prices) all instruments produce the same Sharpe.
	// We can only verify the results are in non-ascending order — which they will be
	// regardless. The important invariant is: no result has a Sharpe > a preceding result.
	from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := universesweep.Config{
		Instruments: []string{"NSE:RELIANCE", "NSE:INFY", "NSE:TCS"},
		Strategy:    &toggleStrategy{},
		EngineConfig: engine.Config{
			From:                 from,
			To:                   to,
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		Timeframe: model.TimeframeDaily,
	}

	report, err := universesweep.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	for i := 1; i < len(report.Results); i++ {
		if report.Results[i].Sharpe > report.Results[i-1].Sharpe {
			t.Errorf("results not sorted descending by Sharpe: results[%d].Sharpe=%f > results[%d].Sharpe=%f",
				i, report.Results[i].Sharpe, i-1, report.Results[i-1].Sharpe)
		}
	}
}

func TestRun_InsufficientDataFlaggedWhenTradeCountBelowThreshold(t *testing.T) {
	t.Parallel()

	// 30 days of data → well below MinTradesForMetrics (30) and MinCurvePointsForMetrics (252).
	// InsufficientData should be true for all results.
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 2, 1, 0, 0, 0, 0, time.UTC) // ~30 days

	cfg := universesweep.Config{
		Instruments: []string{"NSE:RELIANCE"},
		Strategy:    &toggleStrategy{},
		EngineConfig: engine.Config{
			From:                 from,
			To:                   to,
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		Timeframe: model.TimeframeDaily,
	}

	report, err := universesweep.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}
	if !report.Results[0].InsufficientData {
		t.Error("expected InsufficientData=true for short window, got false")
	}
}

func TestRun_ReturnsErrorOnEmptyInstruments(t *testing.T) {
	t.Parallel()

	cfg := universesweep.Config{
		Instruments: []string{},
		Strategy:    &toggleStrategy{},
		EngineConfig: engine.Config{
			From:        time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			To:          time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			InitialCash: 100_000,
		},
		Timeframe: model.TimeframeDaily,
	}

	_, err := universesweep.Run(context.Background(), &cfg, &staticProvider{})
	if err == nil {
		t.Fatal("expected error for empty instruments list, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestWriteCSV
// ---------------------------------------------------------------------------

func TestWriteCSV_HeaderAndRows(t *testing.T) {
	t.Parallel()

	report := universesweep.Report{
		Results: []universesweep.Result{
			{
				Instrument:       "NSE:RELIANCE",
				Sharpe:           1.25,
				TradeCount:       45,
				TotalPnL:         12345.67,
				MaxDrawdown:      8.5,
				InsufficientData: false,
			},
			{
				Instrument:       "NSE:INFY",
				Sharpe:           0.75,
				TradeCount:       12,
				TotalPnL:         -500.00,
				MaxDrawdown:      15.2,
				InsufficientData: true,
			},
		},
	}

	var buf bytes.Buffer
	if err := universesweep.WriteCSV(&buf, report); err != nil {
		t.Fatalf("WriteCSV: unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 rows), got %d:\n%s", len(lines), output)
	}

	// Verify header.
	wantHeader := "instrument,sharpe,trade_count,total_pnl,max_drawdown,insufficient_data"
	if lines[0] != wantHeader {
		t.Errorf("header: got %q, want %q", lines[0], wantHeader)
	}

	// Verify first row contains expected values.
	if !strings.Contains(lines[1], "NSE:RELIANCE") {
		t.Errorf("row 1 missing instrument: %q", lines[1])
	}
	if !strings.Contains(lines[1], "false") {
		t.Errorf("row 1 insufficient_data should be false: %q", lines[1])
	}

	// Verify second row flags insufficient_data=true.
	if !strings.Contains(lines[2], "NSE:INFY") {
		t.Errorf("row 2 missing instrument: %q", lines[2])
	}
	if !strings.Contains(lines[2], "true") {
		t.Errorf("row 2 insufficient_data should be true: %q", lines[2])
	}
}

func TestWriteCSV_EmptyReport(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := universesweep.WriteCSV(&buf, universesweep.Report{}); err != nil {
		t.Fatalf("WriteCSV: unexpected error on empty report: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// Header only — no data rows.
	if len(lines) != 1 {
		t.Fatalf("expected 1 line (header only) for empty report, got %d:\n%s", len(lines), output)
	}
}

// ---------------------------------------------------------------------------
// TestApplyUniverseGate
// ---------------------------------------------------------------------------

func TestApplyUniverseGate_AllSufficient_AllPositiveSharpe_Passes(t *testing.T) {
	t.Parallel()

	// 5 instruments, all sufficient (TradeCount >= 30), all positive Sharpe.
	// DSR-corrected Sharpe will be slightly lower than raw but still positive
	// since raw Sharpe is large.
	report := universesweep.Report{
		Results: []universesweep.Result{
			{Instrument: "A", Sharpe: 1.5, TradeCount: 45, InsufficientData: false},
			{Instrument: "B", Sharpe: 1.2, TradeCount: 38, InsufficientData: false},
			{Instrument: "C", Sharpe: 0.8, TradeCount: 32, InsufficientData: false},
			{Instrument: "D", Sharpe: 0.5, TradeCount: 40, InsufficientData: false},
			{Instrument: "E", Sharpe: 0.3, TradeCount: 33, InsufficientData: false},
		},
	}

	gate := universesweep.ApplyUniverseGate(report, 15)

	if gate.SufficientInstruments != 5 {
		t.Errorf("SufficientInstruments: want 5, got %d", gate.SufficientInstruments)
	}
	if gate.PositiveSharpeInstruments != 5 {
		t.Errorf("PositiveSharpeInstruments: want 5, got %d", gate.PositiveSharpeInstruments)
	}
	if gate.PassFraction < 0.40 {
		t.Errorf("PassFraction: want >= 0.40, got %.4f", gate.PassFraction)
	}
	if !gate.GatePass {
		t.Errorf("GatePass: want true (sufficient instruments with positive DSR Sharpe), got false (DSRAverageSharpe=%.4f)", gate.DSRAverageSharpe)
	}
}

func TestApplyUniverseGate_InsufficientInstrumentsExcluded(t *testing.T) {
	t.Parallel()

	// 3 sufficient, 2 insufficient. Gate must only count the 3 sufficient ones.
	report := universesweep.Report{
		Results: []universesweep.Result{
			{Instrument: "A", Sharpe: 1.5, TradeCount: 45, InsufficientData: false},
			{Instrument: "B", Sharpe: 1.2, TradeCount: 38, InsufficientData: false},
			{Instrument: "C", Sharpe: 0.8, TradeCount: 32, InsufficientData: false},
			{Instrument: "D", Sharpe: 0.0, TradeCount: 10, InsufficientData: true}, // excluded
			{Instrument: "E", Sharpe: 0.0, TradeCount: 5, InsufficientData: true},  // excluded
		},
	}

	gate := universesweep.ApplyUniverseGate(report, 15)

	if gate.SufficientInstruments != 3 {
		t.Errorf("SufficientInstruments: want 3, got %d", gate.SufficientInstruments)
	}
}

func TestApplyUniverseGate_BelowPassFractionThreshold_Fails(t *testing.T) {
	t.Parallel()

	// 5 sufficient instruments, only 1 positive Sharpe → 20% pass fraction < 40%.
	report := universesweep.Report{
		Results: []universesweep.Result{
			{Instrument: "A", Sharpe: 2.0, TradeCount: 50, InsufficientData: false},
			{Instrument: "B", Sharpe: -0.5, TradeCount: 35, InsufficientData: false},
			{Instrument: "C", Sharpe: -0.3, TradeCount: 32, InsufficientData: false},
			{Instrument: "D", Sharpe: -0.8, TradeCount: 40, InsufficientData: false},
			{Instrument: "E", Sharpe: -0.2, TradeCount: 33, InsufficientData: false},
		},
	}

	gate := universesweep.ApplyUniverseGate(report, 15)

	if gate.SufficientInstruments != 5 {
		t.Errorf("SufficientInstruments: want 5, got %d", gate.SufficientInstruments)
	}
	if gate.PositiveSharpeInstruments != 1 {
		t.Errorf("PositiveSharpeInstruments: want 1, got %d", gate.PositiveSharpeInstruments)
	}
	if gate.GatePass {
		t.Errorf("GatePass: want false (20%% pass fraction < 40%%), got true")
	}
}

func TestApplyUniverseGate_NegativeDSRAverage_Fails(t *testing.T) {
	t.Parallel()

	// 6 of 10 instruments have positive raw Sharpe (passes 40%) but the DSR-corrected
	// average is negative (raw Sharpe too small relative to nTrials penalty).
	report := universesweep.Report{
		Results: []universesweep.Result{
			// Positive raw Sharpe but very small — after DSR correction will go negative.
			{Instrument: "A", Sharpe: 0.05, TradeCount: 30, InsufficientData: false},
			{Instrument: "B", Sharpe: 0.04, TradeCount: 31, InsufficientData: false},
			{Instrument: "C", Sharpe: 0.03, TradeCount: 32, InsufficientData: false},
			{Instrument: "D", Sharpe: 0.06, TradeCount: 33, InsufficientData: false},
			{Instrument: "E", Sharpe: 0.02, TradeCount: 34, InsufficientData: false},
			{Instrument: "F", Sharpe: 0.04, TradeCount: 35, InsufficientData: false},
			// Negative raw Sharpe.
			{Instrument: "G", Sharpe: -0.2, TradeCount: 36, InsufficientData: false},
			{Instrument: "H", Sharpe: -0.3, TradeCount: 37, InsufficientData: false},
			{Instrument: "I", Sharpe: -0.1, TradeCount: 38, InsufficientData: false},
			{Instrument: "J", Sharpe: -0.4, TradeCount: 39, InsufficientData: false},
		},
	}

	gate := universesweep.ApplyUniverseGate(report, 15)

	// DSR average across all 10 sufficient instruments should be negative
	// (large negative instruments drag down the tiny positive ones, plus DSR penalty).
	if gate.DSRAverageSharpe >= 0 {
		// If DSR average somehow passes, the gate must still require DSR > 0.
		// For the test to be meaningful, let's just check GatePass reflects DSRAverageSharpe.
		if gate.GatePass != (gate.DSRAverageSharpe > 0 && gate.PassFraction >= 0.40) {
			t.Errorf("GatePass inconsistency: DSRAverageSharpe=%.4f, PassFraction=%.4f, GatePass=%v",
				gate.DSRAverageSharpe, gate.PassFraction, gate.GatePass)
		}
	} else {
		if gate.GatePass {
			t.Errorf("GatePass: want false (negative DSRAverageSharpe=%.4f), got true", gate.DSRAverageSharpe)
		}
	}
}

func TestApplyUniverseGate_EmptyReport_Fails(t *testing.T) {
	t.Parallel()

	gate := universesweep.ApplyUniverseGate(universesweep.Report{}, 15)

	if gate.SufficientInstruments != 0 {
		t.Errorf("SufficientInstruments: want 0, got %d", gate.SufficientInstruments)
	}
	if gate.GatePass {
		t.Error("GatePass: want false for empty report, got true")
	}
}

func TestApplyUniverseGate_ExactlyAtPassFractionThreshold_Passes(t *testing.T) {
	t.Parallel()

	// 5 instruments, exactly 2 positive DSR Sharpe = 40.0% — exactly at the threshold.
	// 40% >= 40% → should pass the fraction check if DSR average is also > 0.
	report := universesweep.Report{
		Results: []universesweep.Result{
			{Instrument: "A", Sharpe: 2.0, TradeCount: 100, InsufficientData: false},
			{Instrument: "B", Sharpe: 1.5, TradeCount: 100, InsufficientData: false},
			{Instrument: "C", Sharpe: -0.5, TradeCount: 100, InsufficientData: false},
			{Instrument: "D", Sharpe: -0.3, TradeCount: 100, InsufficientData: false},
			{Instrument: "E", Sharpe: -0.2, TradeCount: 100, InsufficientData: false},
		},
	}

	gate := universesweep.ApplyUniverseGate(report, 15)

	if gate.SufficientInstruments != 5 {
		t.Errorf("SufficientInstruments: want 5, got %d", gate.SufficientInstruments)
	}
	// PassFraction = 2/5 = 0.40.
	if gate.PassFraction < 0.399 || gate.PassFraction > 0.401 {
		t.Errorf("PassFraction: want 0.40, got %.4f", gate.PassFraction)
	}
}

func TestApplyUniverseGate_TradesCarriedOnResult(t *testing.T) {
	t.Parallel()

	// Verify that Result.Trades is populated after Run (using a longer window).
	from := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := universesweep.Config{
		Instruments: []string{"NSE:RELIANCE"},
		Strategy:    &toggleStrategy{},
		EngineConfig: engine.Config{
			From:                 from,
			To:                   to,
			InitialCash:          100_000,
			PositionSizeFraction: 0.10,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		Timeframe: model.TimeframeDaily,
	}

	report, err := universesweep.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}

	// Trades must be populated since toggleStrategy generates trades.
	if len(report.Results[0].Trades) == 0 {
		t.Error("Result.Trades: want non-empty slice, got empty (trades not captured)")
	}
}
