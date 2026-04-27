package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep2d"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// minimalReport returns a small Report2D with two p1 values and one p2 value,
// sufficient to exercise writeOutput and writeCSVToWriter without running a sweep.
func minimalReport() sweep2d.Report2D {
	return sweep2d.Report2D{
		Param1Name:   "fast",
		Param2Name:   "slow",
		Param1Values: []float64{5, 10},
		Param2Values: []float64{20},
		Grid: [][]sweep2d.GridCell{
			{{Param1Value: 5, Param2Value: 20, SharpeRatio: 1.2}},
			{{Param1Value: 10, Param2Value: 20, SharpeRatio: 0.8}},
		},
		VariantCount:           2,
		PeakSharpe:             1.2,
		DSRCorrectedPeakSharpe: 1.1,
	}
}

// TestWriteOutput_File verifies that writeOutput writes a valid CSV to the given path.
func TestWriteOutput_File(t *testing.T) {
	t.Parallel()
	report := minimalReport()
	outPath := filepath.Join(t.TempDir(), "out.csv")
	if err := writeOutput(os.Stdout, report, outPath); err != nil {
		t.Fatalf("writeOutput: %v", err)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output file: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "fast") {
		t.Errorf("CSV missing param1 name 'fast'; got:\n%s", got)
	}
	if !strings.Contains(got, "slow") {
		t.Errorf("CSV missing param2 name 'slow'; got:\n%s", got)
	}
}

// TestWriteOutput_Stdout verifies that writeOutput writes valid CSV to the provided writer
// when no output path is given.
func TestWriteOutput_Stdout(t *testing.T) {
	t.Parallel()
	report := minimalReport()
	var buf bytes.Buffer
	if err := writeOutput(&buf, report, ""); err != nil {
		t.Fatalf("writeOutput: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "fast") {
		t.Errorf("CSV missing param1 name 'fast'; got:\n%s", got)
	}
	if !strings.Contains(got, "1.200000") {
		t.Errorf("CSV missing peak Sharpe cell '1.200000'; got:\n%s", got)
	}
}

// TestParseAndValidateFlags2D_RequiredFlags verifies that missing required flags are rejected.
func TestParseAndValidateFlags2D_RequiredFlags(t *testing.T) {
	t.Parallel()
	base := flags2D{
		fromStr:   "2024-01-01",
		toStr:     "2024-12-31",
		tfStr:     "daily",
		stratName: "sma-crossover",
		p1Min:     5,
		p1Max:     20,
		p1Step:    5,
		p2Min:     20,
		p2Max:     60,
		p2Step:    10,
	}

	cases := []struct {
		name    string
		modify  func(*flags2D)
		wantErr string
	}{
		{"missing from", func(f *flags2D) { f.fromStr = "" }, "--from"},
		{"missing to", func(f *flags2D) { f.toStr = "" }, "--to"},
		{"missing strategy", func(f *flags2D) { f.stratName = "" }, "--strategy"},
		{"p1 step zero", func(f *flags2D) { f.p1Step = 0 }, "--p1-step"},
		{"p1 step negative", func(f *flags2D) { f.p1Step = -1 }, "--p1-step"},
		{"p2 step zero", func(f *flags2D) { f.p2Step = 0 }, "--p2-step"},
		{"p1 max < min", func(f *flags2D) { f.p1Max = 1; f.p1Min = 10 }, "--p1-max"},
		{"p2 max < min", func(f *flags2D) { f.p2Max = 1; f.p2Min = 10 }, "--p2-max"},
		{"to not after from", func(f *flags2D) { f.toStr = "2024-01-01" }, "--to must be strictly after"},
		{"bad from date", func(f *flags2D) { f.fromStr = "not-a-date" }, "--from"},
		{"bad to date", func(f *flags2D) { f.toStr = "not-a-date" }, "--to"},
		{"invalid timeframe", func(f *flags2D) { f.tfStr = "hourly" }, "--timeframe"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := base
			tc.modify(&f)
			_, _, _, err := parseAndValidateFlags2D(f)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

// TestParseAndValidateFlags2D_Valid verifies that well-formed flags parse successfully.
func TestParseAndValidateFlags2D_Valid(t *testing.T) {
	t.Parallel()
	f := flags2D{
		fromStr:   "2024-01-01",
		toStr:     "2024-12-31",
		tfStr:     "daily",
		stratName: "sma-crossover",
		p1Min:     5,
		p1Max:     20,
		p1Step:    5,
		p2Min:     20,
		p2Max:     60,
		p2Step:    10,
	}
	from, to, tf, err := parseAndValidateFlags2D(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("from: got %v, want 2024-01-01", from)
	}
	if !to.Equal(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("to: got %v, want 2024-12-31", to)
	}
	if tf != model.TimeframeDaily {
		t.Errorf("tf: got %q, want daily", tf)
	}
}

// TestFactoryRegistry2D_KnownStrategies verifies that sma-crossover and
// rsi-mean-reversion produce callable factories.
func TestFactoryRegistry2D_KnownStrategies(t *testing.T) {
	t.Parallel()

	cases := []struct {
		strategyName string
		p1           float64
		p2           float64
	}{
		{"sma-crossover", 10, 50},
		{"rsi-mean-reversion", 14, 30},
	}

	for _, tc := range cases {
		t.Run(tc.strategyName, func(t *testing.T) {
			t.Parallel()
			factory, err := factoryRegistry2D(tc.strategyName, model.TimeframeDaily)
			if err != nil {
				t.Fatalf("factoryRegistry2D(%q): %v", tc.strategyName, err)
			}
			s, err := factory(tc.p1, tc.p2)
			if err != nil {
				t.Fatalf("factory(%g, %g): %v", tc.p1, tc.p2, err)
			}
			if s == nil {
				t.Fatal("factory returned nil strategy")
			}
		})
	}
}

// TestFactoryRegistry2D_UnknownStrategy verifies that unsupported strategy names
// return an error.
func TestFactoryRegistry2D_UnknownStrategy(t *testing.T) {
	t.Parallel()
	_, err := factoryRegistry2D("does-not-exist", model.TimeframeDaily)
	if err == nil {
		t.Fatal("expected error for unknown strategy, got nil")
	}
	if !strings.Contains(err.Error(), "does-not-exist") {
		t.Errorf("error %q does not mention the unknown strategy name", err.Error())
	}
}

// TestRunAndWriteCSV_SmokeTest is the end-to-end smoke test using a static provider.
// It verifies that a 2D sweep over sma-crossover (fast × slow) completes and
// produces a CSV with the correct column headers.
func TestRunAndWriteCSV_SmokeTest(t *testing.T) {
	t.Parallel()

	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}

	factory, err := factoryRegistry2D("sma-crossover", model.TimeframeDaily)
	if err != nil {
		t.Fatalf("factoryRegistry2D: %v", err)
	}

	cfg := sweep2d.Config2D{
		Param1:    sweep2d.ParamRange{Name: "fast-period", Min: 5, Max: 10, Step: 5},
		Param2:    sweep2d.ParamRange{Name: "slow-period", Min: 20, Max: 40, Step: 20},
		Timeframe: model.TimeframeDaily,
		EngineConfig: engine.Config{
			Instrument:           "TEST:X",
			From:                 time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			To:                   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
			InitialCash:          100000,
			PositionSizeFraction: 0.1,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		StrategyFactory: factory,
	}

	report, err := sweep2d.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("sweep2d.Run: %v", err)
	}

	// Write CSV to a buffer and verify column headers.
	var buf bytes.Buffer
	if err := writeCSVToWriter(&buf, report); err != nil {
		t.Fatalf("writeCSVToWriter: %v", err)
	}

	csv := buf.String()
	lines := strings.Split(strings.TrimRight(csv, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("CSV has %d lines, want >= 2", len(lines))
	}

	// Header must contain both parameter names.
	header := lines[0]
	if !strings.Contains(header, "fast-period") {
		t.Errorf("header %q does not contain 'fast-period'", header)
	}
	if !strings.Contains(header, "slow-period") {
		t.Errorf("header %q does not contain 'slow-period'", header)
	}

	// DSR should be populated.
	if report.DSRCorrectedPeakSharpe == 0 && report.VariantCount > 1 {
		t.Error("DSRCorrectedPeakSharpe is zero for multi-variant sweep")
	}
}
