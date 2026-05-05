package output_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// failAfterFirstWriter succeeds on the first Write call, then returns an error.
// Used to exercise error paths that are only reachable on the second write in a function.
type failAfterFirstWriter struct{ wrote bool }

func (f *failAfterFirstWriter) Write(p []byte) (int, error) {
	if f.wrote {
		return 0, errors.New("write failed")
	}
	f.wrote = true
	return len(p), nil
}

func TestWrite_JSONOutput(t *testing.T) {
	tests := []struct {
		name   string
		report analytics.Report
	}{
		{
			name:   "empty_report",
			report: analytics.Report{},
		},
		{
			name: "single_winner",
			report: analytics.Report{
				TotalPnL:   100,
				WinRate:    100,
				TradeCount: 1,
				WinCount:   1,
				LossCount:  0,
			},
		},
		{
			name: "mixed_trades",
			report: analytics.Report{
				TotalPnL:    150,
				WinRate:     66.6667,
				MaxDrawdown: 50,
				TradeCount:  3,
				WinCount:    2,
				LossCount:   1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "out.json")
			cfg := output.Config{FilePath: path}

			if err := output.Write(tt.report, cfg); err != nil {
				t.Fatalf("Write: %v", err)
			}

			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}

			var got analytics.Report
			if err := json.Unmarshal(data, &got); err != nil {
				t.Fatalf("Unmarshal JSON: %v", err)
			}

			if got != tt.report {
				t.Errorf("round-trip mismatch\n got  %+v\n want %+v", got, tt.report)
			}
		})
	}
}

func TestWrite_JSONIsValid(t *testing.T) {
	// JSON output must be valid even for a zero-value report.
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := output.Write(analytics.Report{}, output.Config{FilePath: path}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !json.Valid(data) {
		t.Errorf("output is not valid JSON: %s", data)
	}
}

func TestWrite_StdoutSummary(t *testing.T) {
	report := analytics.Report{
		TotalPnL:    500,
		WinRate:     75,
		MaxDrawdown: 10,
		TradeCount:  4,
		WinCount:    3,
		LossCount:   1,
	}

	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &buf,
	}

	if err := output.Write(report, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"500", "75", "10", "4"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout summary missing %q:\n%s", want, out)
		}
	}
}

func TestWrite_StdoutIncludesSharpe(t *testing.T) {
	report := analytics.Report{SharpeRatio: 1.2345}

	var buf bytes.Buffer
	if err := output.Write(report, output.Config{PrintToStdout: true, Stdout: &buf}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !strings.Contains(buf.String(), "Sharpe") {
		t.Errorf("stdout summary missing Sharpe line:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "1.2345") {
		t.Errorf("stdout summary missing Sharpe value 1.2345:\n%s", buf.String())
	}
}

func TestWrite_StdoutDisabled(t *testing.T) {
	// When PrintToStdout is false, Stdout writer must not be touched.
	cfg := output.Config{
		PrintToStdout: false,
		Stdout:        nil, // would panic if written to
	}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestWrite_NeitherOutput(t *testing.T) {
	// No FilePath, no PrintToStdout — must succeed silently.
	if err := output.Write(analytics.Report{}, output.Config{}); err != nil {
		t.Fatalf("Write: %v", err)
	}
}

func TestWrite_FilePathOnly(t *testing.T) {
	// Stdout must not be written when PrintToStdout is false.
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	var buf bytes.Buffer
	cfg := output.Config{
		FilePath:      path,
		PrintToStdout: false,
		Stdout:        &buf,
	}

	if err := output.Write(analytics.Report{TradeCount: 1, WinCount: 1, WinRate: 100}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no stdout output, got: %q", buf.String())
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("JSON file not created: %v", err)
	}
}

func TestWrite_BadFilePath(t *testing.T) {
	cfg := output.Config{FilePath: "/nonexistent/dir/out.json"}
	if err := output.Write(analytics.Report{}, cfg); err == nil {
		t.Error("expected error for bad file path, got nil")
	}
}

func TestWrite_StdoutWithBenchmark(t *testing.T) {
	report := analytics.Report{
		TotalPnL:    500,
		WinRate:     75,
		MaxDrawdown: 10,
		TradeCount:  4,
		SharpeRatio: 1.2345,
	}
	benchmark := &analytics.BenchmarkReport{
		TotalReturn:      18.50,
		AnnualizedReturn: 12.30,
		MaxDrawdown:      8.75,
		SharpeRatio:      0.9876,
	}

	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &buf,
		Benchmark:     benchmark,
	}

	if err := output.Write(report, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	out := buf.String()
	for _, want := range []string{"Buy-and-Hold Benchmark", "18.50", "12.30", "8.75", "0.9876"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout missing %q:\n%s", want, out)
		}
	}
}

func TestWrite_BenchmarkWriteError(t *testing.T) {
	// The writer succeeds on the first fmt.Fprintf (strategy summary) and fails on the
	// second (benchmark section), exercising the "output: write benchmark summary" error path.
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &failAfterFirstWriter{},
		Benchmark:     &analytics.BenchmarkReport{TotalReturn: 10},
	}
	if err := output.Write(analytics.Report{}, cfg); err == nil {
		t.Error("expected error when benchmark write fails, got nil")
	}
}

func TestWrite_StdoutNoBenchmarkSection(t *testing.T) {
	var buf bytes.Buffer
	cfg := output.Config{PrintToStdout: true, Stdout: &buf}

	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if strings.Contains(buf.String(), "Benchmark") {
		t.Errorf("expected no benchmark section when Benchmark is nil:\n%s", buf.String())
	}
}

// --- Insufficient-sample warnings ---

func TestWrite_StdoutWarning_TradeMetrics(t *testing.T) {
	report := analytics.Report{TradeMetricsInsufficient: true, TradeCount: 7}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{PrintToStdout: true, Stdout: &buf}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected WARNING in output, got:\n%s", out)
	}
	if !strings.Contains(out, "7") {
		t.Errorf("expected trade count 7 in warning, got:\n%s", out)
	}
}

func TestWrite_StdoutWarning_CurveMetrics(t *testing.T) {
	report := analytics.Report{CurveMetricsInsufficient: true}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{PrintToStdout: true, Stdout: &buf}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected WARNING in output, got:\n%s", buf.String())
	}
}

func TestWrite_StdoutWarning_NoFlagsNoWarning(t *testing.T) {
	report := analytics.Report{TradeCount: 100}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{PrintToStdout: true, Stdout: &buf}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected no WARNING when flags are false, got:\n%s", buf.String())
	}
}

// --- WriteCurveCSV tests ---

func TestWrite_CurvePath_RoundTrip(t *testing.T) {
	curve := []model.EquityPoint{
		{Timestamp: time.Date(2018, 1, 2, 9, 15, 0, 0, time.UTC), Value: 100000.00},
		{Timestamp: time.Date(2018, 1, 3, 9, 15, 0, 0, time.UTC), Value: 100250.50},
		{Timestamp: time.Date(2018, 1, 4, 9, 15, 0, 0, time.UTC), Value: 99875.25},
	}

	dir := t.TempDir()
	curvePath := filepath.Join(dir, "curve.csv")

	if err := output.Write(analytics.Report{}, output.Config{
		CurvePath: curvePath,
		Curve:     curve,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(curvePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != len(curve)+1 {
		t.Fatalf("expected %d lines (1 header + %d data), got %d", len(curve)+1, len(curve), len(lines))
	}
	if lines[0] != "timestamp,equity_value" {
		t.Errorf("unexpected header: %q", lines[0])
	}

	for i, pt := range curve {
		parts := strings.SplitN(lines[i+1], ",", 2)
		if len(parts) != 2 {
			t.Errorf("row %d: expected 2 fields, got %q", i, lines[i+1])
			continue
		}
		ts, err := time.Parse(time.RFC3339, parts[0])
		if err != nil {
			t.Errorf("row %d: parse timestamp %q: %v", i, parts[0], err)
			continue
		}
		if !ts.Equal(pt.Timestamp) {
			t.Errorf("row %d: timestamp got %v, want %v", i, ts, pt.Timestamp)
		}
		got, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			t.Errorf("row %d: parse value %q: %v", i, parts[1], err)
			continue
		}
		if got != pt.Value {
			t.Errorf("row %d: value got %v, want %v", i, got, pt.Value)
		}
	}
}

func TestWrite_CurvePath_EmptyPath_NoCurveFile(t *testing.T) {
	// CurvePath is empty — Write must succeed without touching any file.
	curve := []model.EquityPoint{
		{Timestamp: time.Date(2018, 1, 2, 9, 15, 0, 0, time.UTC), Value: 100000.00},
	}
	if err := output.Write(analytics.Report{}, output.Config{Curve: curve}); err != nil {
		t.Fatalf("Write with no CurvePath: %v", err)
	}
}

func TestWrite_CurvePath_BadPath(t *testing.T) {
	curve := []model.EquityPoint{
		{Timestamp: time.Date(2018, 1, 2, 9, 15, 0, 0, time.UTC), Value: 100000.00},
	}
	if err := output.Write(analytics.Report{}, output.Config{
		CurvePath: "/nonexistent/dir/curve.csv",
		Curve:     curve,
	}); err == nil {
		t.Error("expected error for bad curve path, got nil")
	}
}

// --- WriteSweep tests ---

func makeSweepReport() sweep.Report {
	return sweep.Report{
		ParameterName: "period",
		Results: []sweep.Result{
			{ParamValue: 14, SharpeRatio: 1.5, TotalPnL: 5000, TradeCount: 20, MaxDrawdown: 8.5},
			{ParamValue: 12, SharpeRatio: 1.2, TotalPnL: 4200, TradeCount: 22, MaxDrawdown: 9.1},
			{ParamValue: 20, SharpeRatio: 0.8, TotalPnL: 2100, TradeCount: 15, MaxDrawdown: 12.0},
		},
		Plateau: &sweep.PlateauRange{MinParam: 12, MaxParam: 14, Count: 2, MinSharpe: 1.2},
	}
}

func TestWriteSweep_ContainsParameterName(t *testing.T) {
	var buf bytes.Buffer
	if err := output.WriteSweep(&buf, makeSweepReport()); err != nil {
		t.Fatalf("WriteSweep: %v", err)
	}
	if !strings.Contains(buf.String(), "period") {
		t.Errorf("output missing parameter name %q:\n%s", "period", buf.String())
	}
}

func TestWriteSweep_ContainsAllResults(t *testing.T) {
	var buf bytes.Buffer
	if err := output.WriteSweep(&buf, makeSweepReport()); err != nil {
		t.Fatalf("WriteSweep: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"14", "12", "20", "1.5", "1.2", "0.8", "5000", "4200", "2100"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWriteSweep_ContainsPlateauInfo(t *testing.T) {
	var buf bytes.Buffer
	if err := output.WriteSweep(&buf, makeSweepReport()); err != nil {
		t.Fatalf("WriteSweep: %v", err)
	}
	out := buf.String()
	// Plateau must mention the range and count.
	for _, want := range []string{"Plateau", "12", "14", "2"} {
		if !strings.Contains(out, want) {
			t.Errorf("plateau section missing %q:\n%s", want, out)
		}
	}
}

func TestWriteSweep_NilPlateauOmitsSection(t *testing.T) {
	report := sweep.Report{
		ParameterName: "period",
		Results:       []sweep.Result{{ParamValue: 10, SharpeRatio: -0.5}},
		Plateau:       nil,
	}
	var buf bytes.Buffer
	if err := output.WriteSweep(&buf, report); err != nil {
		t.Fatalf("WriteSweep: %v", err)
	}
	if strings.Contains(buf.String(), "Plateau") {
		t.Errorf("expected no plateau section when Plateau is nil:\n%s", buf.String())
	}
}

func TestWriteSweep_WriteError(t *testing.T) {
	// failWriter always fails on Write — exercises the error return path.
	if err := output.WriteSweep(&failWriter{}, makeSweepReport()); err == nil {
		t.Error("expected error from failing writer, got nil")
	}
}

// failWriter always returns an error on Write.
type failWriter struct{}

func (f *failWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

// --- Proliferation gate tests ---

func TestWrite_GatePASS(t *testing.T) {
	report := analytics.Report{SharpeRatio: 0.65}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{
		PrintToStdout: true, Stdout: &buf, GateThreshold: 0.5,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "PASS") {
		t.Errorf("expected PASS in output, got:\n%s", out)
	}
	if !strings.Contains(out, "0.65") {
		t.Errorf("expected actual Sharpe in output, got:\n%s", out)
	}
	if !strings.Contains(out, "≥0.50") {
		t.Errorf("expected threshold label in output, got:\n%s", out)
	}
}

func TestWrite_GateFAIL(t *testing.T) {
	report := analytics.Report{SharpeRatio: 0.447}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{
		PrintToStdout: true, Stdout: &buf, GateThreshold: 0.5,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "FAIL") {
		t.Errorf("expected FAIL in output, got:\n%s", out)
	}
}

func TestWrite_GateDisabled_ZeroThreshold(t *testing.T) {
	report := analytics.Report{SharpeRatio: 0.3}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{
		PrintToStdout: true, Stdout: &buf, GateThreshold: 0,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "gate") {
		t.Errorf("expected no gate output when threshold is 0, got:\n%s", buf.String())
	}
}

func TestWrite_GateSkipped_TradeMetricsInsufficient(t *testing.T) {
	report := analytics.Report{SharpeRatio: 0.3, TradeMetricsInsufficient: true, TradeCount: 5}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{
		PrintToStdout: true, Stdout: &buf, GateThreshold: 0.5,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "gate") {
		t.Errorf("expected no gate output when TradeMetricsInsufficient, got:\n%s", buf.String())
	}
}

func TestWrite_GateSkipped_CurveMetricsInsufficient(t *testing.T) {
	report := analytics.Report{SharpeRatio: 0.3, CurveMetricsInsufficient: true}
	var buf bytes.Buffer
	if err := output.Write(report, output.Config{
		PrintToStdout: true, Stdout: &buf, GateThreshold: 0.5,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "gate") {
		t.Errorf("expected no gate output when CurveMetricsInsufficient, got:\n%s", buf.String())
	}
}

func TestWrite_GateWriteError(t *testing.T) {
	// failAfterFirstWriter: write 1 (main summary) succeeds, write 2 (gate line) fails.
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &failAfterFirstWriter{},
		GateThreshold: 0.5,
	}
	if err := output.Write(analytics.Report{SharpeRatio: 0.65}, cfg); err == nil {
		t.Error("expected error when gate write fails, got nil")
	}
}

// --- Regime split table tests ---

func TestWrite_RegimeSplits_TablePrinted(t *testing.T) {
	regimes := []analytics.RegimeReport{
		{
			Name:        "Pre-COVID (2018–2019)",
			From:        time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			To:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			SharpeRatio: 0.312,
			MaxDrawdown: 9.5,
		},
		{
			Name:        "COVID Crash + Recovery (2020–2021)",
			From:        time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			To:          time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			SharpeRatio: -0.140,
			MaxDrawdown: 16.3,
		},
	}

	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &buf,
		RegimeSplits:  regimes,
	}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Regime Split",
		"Pre-COVID",
		"COVID",
		"0.3120",
		"-0.1400",
		"9.50",
		"16.30",
		"2018-01",
		"2020-01",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("regime table missing %q:\n%s", want, out)
		}
	}
}

// --- Bootstrap section tests ---

func TestWrite_Bootstrap_Printed(t *testing.T) {
	result := &montecarlo.BootstrapResult{
		MeanSharpe:         0.4600,
		SharpeP5:           -0.1234,
		SharpeP50:          0.4567,
		SharpeP95:          1.2345,
		WorstDrawdownP5:    3.21,
		WorstDrawdownP50:   12.43,
		WorstDrawdownP95:   28.91,
		ProbPositiveSharpe: 0.734,
	}
	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout:  true,
		Stdout:         &buf,
		Bootstrap:      result,
		BootstrapSeed:  42,
		BootstrapNSims: 10_000,
	}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"Bootstrap",
		"seed=42",
		"-0.1234", // SharpeP5
		"0.4567",  // SharpeP50
		"1.2345",  // SharpeP95
		"73.4",    // ProbPositiveSharpe as %
	} {
		if !strings.Contains(out, want) {
			t.Errorf("bootstrap output missing %q:\n%s", want, out)
		}
	}
}

func TestWrite_Bootstrap_OmittedWhenNil(t *testing.T) {
	var buf bytes.Buffer
	if err := output.Write(analytics.Report{}, output.Config{PrintToStdout: true, Stdout: &buf}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "Bootstrap") {
		t.Errorf("expected no bootstrap section when Bootstrap is nil:\n%s", buf.String())
	}
}

func TestWrite_Bootstrap_DefaultNSims(t *testing.T) {
	// BootstrapNSims=0 should display 10000 in the header.
	result := &montecarlo.BootstrapResult{ProbPositiveSharpe: 0.5}
	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout:  true,
		Stdout:         &buf,
		Bootstrap:      result,
		BootstrapSeed:  1,
		BootstrapNSims: 0,
	}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if !strings.Contains(buf.String(), "10000") {
		t.Errorf("expected default 10000 in bootstrap header:\n%s", buf.String())
	}
}

func TestWrite_RegimeSplits_OmittedWhenEmpty(t *testing.T) {
	var buf bytes.Buffer
	cfg := output.Config{
		PrintToStdout: true,
		Stdout:        &buf,
		RegimeSplits:  nil,
	}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if strings.Contains(buf.String(), "Regime Split") {
		t.Errorf("expected no regime section when RegimeSplits is nil:\n%s", buf.String())
	}
}

// --- WriteCorrelationMatrix tests ---

func TestWriteCorrelationMatrix_BasicValues(t *testing.T) {
	m := analytics.CorrelationMatrix{
		Pairs: []analytics.PairCorrelation{
			{NameA: "sma", NameB: "donchian", FullPeriod: 0.4321, Crash2020: 0.5678, Correction2022: 0.2345},
			{NameA: "macd", NameB: "bollinger", FullPeriod: -0.1234, Crash2020: 0.3210, Correction2022: -0.0987},
		},
	}
	var buf bytes.Buffer
	if err := output.WriteCorrelationMatrix(&buf, m); err != nil {
		t.Fatalf("WriteCorrelationMatrix: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"Strategy Correlation Matrix", "Strategy A", "Strategy B",
		"Full-Period", "2020-Crash", "2022-Corr",
		"sma", "donchian", "0.4321",
		"macd", "bollinger", "-0.1234",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestWriteCorrelationMatrix_NaN(t *testing.T) {
	m := analytics.CorrelationMatrix{
		Pairs: []analytics.PairCorrelation{
			{NameA: "a", NameB: "b", FullPeriod: math.NaN(), Crash2020: math.NaN(), Correction2022: 0.5},
		},
	}
	var buf bytes.Buffer
	if err := output.WriteCorrelationMatrix(&buf, m); err != nil {
		t.Fatalf("WriteCorrelationMatrix: %v", err)
	}
	if !strings.Contains(buf.String(), "n/a") {
		t.Errorf("expected NaN to render as n/a:\n%s", buf.String())
	}
}

func TestWriteCorrelationMatrix_TooCorrelated(t *testing.T) {
	m := analytics.CorrelationMatrix{
		Pairs: []analytics.PairCorrelation{
			{NameA: "sma", NameB: "ema", FullPeriod: 0.85, Crash2020: 0.90, Correction2022: 0.88, TooCorrelated: true},
		},
	}
	var buf bytes.Buffer
	if err := output.WriteCorrelationMatrix(&buf, m); err != nil {
		t.Fatalf("WriteCorrelationMatrix: %v", err)
	}
	if !strings.Contains(buf.String(), "too correlated") {
		t.Errorf("expected TooCorrelated warning note in output:\n%s", buf.String())
	}
}

func TestWriteCorrelationMatrix_WriteError_Header(t *testing.T) {
	m := analytics.CorrelationMatrix{
		Pairs: []analytics.PairCorrelation{{NameA: "a", NameB: "b", FullPeriod: 0.5}},
	}
	if err := output.WriteCorrelationMatrix(&failWriter{}, m); err == nil {
		t.Error("expected error when writer fails, got nil")
	}
}

func TestWriteCorrelationMatrix_WriteError_Row(t *testing.T) {
	m := analytics.CorrelationMatrix{
		Pairs: []analytics.PairCorrelation{{NameA: "a", NameB: "b", FullPeriod: 0.5}},
	}
	if err := output.WriteCorrelationMatrix(&failAfterFirstWriter{}, m); err == nil {
		t.Error("expected error when row write fails, got nil")
	}
}

// --- RunConfig metadata in JSON ---

func TestWrite_RunConfig_AppearsAtTopLevelInJSON(t *testing.T) {
	report := analytics.Report{TradeCount: 5, SharpeRatio: 1.23}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	rc := output.RunConfig{
		Instrument:      "NSE:RELIANCE",
		Timeframe:       "daily",
		From:            "2018-01-01",
		To:              "2024-01-01",
		Strategy:        "sma-crossover",
		CommissionModel: "zerodha_full",
		Parameters:      map[string]string{"fast_period": "10", "slow_period": "50"},
	}
	cfg := output.Config{FilePath: path, RunConfig: rc}

	if err := output.Write(report, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Parse into a generic map to verify top-level keys.
	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal JSON: %v", err)
	}

	for _, wantKey := range []string{"instrument", "timeframe", "from", "to", "strategy", "commission_model", "parameters"} {
		if _, ok := got[wantKey]; !ok {
			t.Errorf("JSON missing top-level key %q; full JSON:\n%s", wantKey, data)
		}
	}
	if got["instrument"] != "NSE:RELIANCE" {
		t.Errorf("instrument = %q, want %q", got["instrument"], "NSE:RELIANCE")
	}
	if got["strategy"] != "sma-crossover" {
		t.Errorf("strategy = %q, want %q", got["strategy"], "sma-crossover")
	}
}

func TestWrite_RunConfig_ZeroValue_NoExtraKeys(t *testing.T) {
	// When RunConfig is zero-valued, the JSON output must not contain metadata keys.
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := output.Write(analytics.Report{TradeCount: 1}, output.Config{FilePath: path}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal JSON: %v", err)
	}

	for _, key := range []string{"instrument", "timeframe", "from", "to", "strategy", "commission_model"} {
		if _, ok := got[key]; ok {
			t.Errorf("JSON should not contain key %q when RunConfig is zero; full JSON:\n%s", key, data)
		}
	}
}

func TestWrite_RunConfig_MetricsFieldsStillPresent(t *testing.T) {
	// Existing metrics fields must remain accessible after RunConfig is added.
	report := analytics.Report{TradeCount: 42, SharpeRatio: 0.75, TotalPnL: 12345.67}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	rc := output.RunConfig{Instrument: "NSE:INFY", Timeframe: "daily", Strategy: "momentum"}
	if err := output.Write(report, output.Config{FilePath: path, RunConfig: rc}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// The report fields must still deserialize correctly.
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal JSON: %v", err)
	}
	tradeCountRaw, ok := raw["TradeCount"]
	if !ok {
		t.Fatalf("JSON missing TradeCount field; keys: %v", keysOf(raw))
	}
	tradeCountF, ok := tradeCountRaw.(float64)
	if !ok {
		t.Fatalf("TradeCount is not a float64; got %T: %v", tradeCountRaw, tradeCountRaw)
	}
	if int(tradeCountF) != 42 {
		t.Errorf("TradeCount = %d, want 42", int(tradeCountF))
	}
}

func TestWrite_RunConfig_ParametersMap(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	rc := output.RunConfig{
		Strategy:   "rsi-mean-reversion",
		Parameters: map[string]string{"rsi_period": "14", "oversold": "30", "overbought": "70"},
	}
	if err := output.Write(analytics.Report{}, output.Config{FilePath: path, RunConfig: rc}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	params, ok := got["parameters"].(map[string]any)
	if !ok {
		t.Fatalf("parameters field missing or not an object; full JSON:\n%s", data)
	}
	if params["rsi_period"] != "14" {
		t.Errorf("parameters.rsi_period = %v, want %q", params["rsi_period"], "14")
	}
}

// --- Bootstrap JSON serialization tests (TASK-0082) ---

// TestWrite_Bootstrap_FieldsInJSON verifies that when Bootstrap is non-nil, the JSON
// output contains a "bootstrap" nested object with all 7 required fields.
func TestWrite_Bootstrap_FieldsInJSON(t *testing.T) {
	result := &montecarlo.BootstrapResult{
		SharpeP5:           0.0719,
		SharpeP50:          0.3195,
		SharpeP95:          0.5551,
		ProbPositiveSharpe: 0.980,
		WorstDrawdownP95:   18.42,
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	cfg := output.Config{
		FilePath:       path,
		Bootstrap:      result,
		BootstrapSeed:  42,
		BootstrapNSims: 10_000,
	}

	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	bsRaw, ok := raw["bootstrap"]
	if !ok {
		t.Fatalf("JSON missing 'bootstrap' key; full JSON:\n%s", data)
	}
	bs, ok := bsRaw.(map[string]any)
	if !ok {
		t.Fatalf("'bootstrap' is not an object; got %T", bsRaw)
	}

	for _, wantKey := range []string{
		"sharpe_p5", "sharpe_p50", "sharpe_p95",
		"prob_positive_sharpe", "worst_drawdown_p95",
		"n", "seed",
	} {
		if _, ok := bs[wantKey]; !ok {
			t.Errorf("bootstrap object missing key %q; keys present: %v", wantKey, keysOf(bs))
		}
	}

	assertFloat64 := func(key string, want float64) {
		t.Helper()
		v, ok := bs[key].(float64)
		if !ok {
			t.Errorf("bootstrap[%q] is not float64; got %T", key, bs[key])
			return
		}
		if v != want {
			t.Errorf("bootstrap[%q] = %v, want %v", key, v, want)
		}
	}
	assertFloat64("sharpe_p5", result.SharpeP5)
	assertFloat64("sharpe_p50", result.SharpeP50)
	assertFloat64("sharpe_p95", result.SharpeP95)
	assertFloat64("prob_positive_sharpe", result.ProbPositiveSharpe)
	assertFloat64("worst_drawdown_p95", result.WorstDrawdownP95)

	seedRaw, ok := bs["seed"].(float64)
	if !ok {
		t.Fatalf("bootstrap.seed is not float64; got %T", bs["seed"])
	}
	if int64(seedRaw) != 42 {
		t.Errorf("seed = %d, want 42", int64(seedRaw))
	}

	nRaw, ok := bs["n"].(float64)
	if !ok {
		t.Fatalf("bootstrap.n is not float64; got %T", bs["n"])
	}
	if int(nRaw) != 10_000 {
		t.Errorf("n = %d, want 10000", int(nRaw))
	}
}

// TestWrite_Bootstrap_AbsentFromJSONWhenNil verifies that when Bootstrap is nil,
// the JSON output has no "bootstrap" key — existing consumers must not be broken.
func TestWrite_Bootstrap_AbsentFromJSONWhenNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := output.Write(analytics.Report{TradeCount: 5}, output.Config{FilePath: path}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if _, ok := raw["bootstrap"]; ok {
		t.Errorf("JSON must not contain 'bootstrap' key when Bootstrap is nil; full JSON:\n%s", data)
	}
}

// TestWrite_Bootstrap_NDefaultsTo10000 verifies that when BootstrapNSims is 0,
// the JSON bootstrap.n is written as 10000 (the montecarlo default).
func TestWrite_Bootstrap_NDefaultsTo10000(t *testing.T) {
	result := &montecarlo.BootstrapResult{ProbPositiveSharpe: 0.5}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	cfg := output.Config{
		FilePath:       path,
		Bootstrap:      result,
		BootstrapSeed:  7,
		BootstrapNSims: 0, // should default to 10000
	}

	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	bsRaw2, ok := raw["bootstrap"]
	if !ok {
		t.Fatalf("JSON missing 'bootstrap' key; full JSON:\n%s", data)
	}
	bs, ok := bsRaw2.(map[string]any)
	if !ok {
		t.Fatalf("'bootstrap' is not an object; got %T", bsRaw2)
	}
	nRaw, ok := bs["n"].(float64)
	if !ok {
		t.Fatalf("bootstrap.n is not float64; got %T", bs["n"])
	}
	if int(nRaw) != 10_000 {
		t.Errorf("bootstrap.n = %d, want 10000 when BootstrapNSims is 0", int(nRaw))
	}
}

// TestWrite_Bootstrap_ZeroSharpeP5NotSuppressed verifies that a bootstrap result
// where SharpeP5 == 0.0 is still serialized (pointer-to-struct prevents omitempty
// from suppressing a valid zero result).
func TestWrite_Bootstrap_ZeroSharpeP5NotSuppressed(t *testing.T) {
	result := &montecarlo.BootstrapResult{
		SharpeP5:           0.0, // valid zero result — must not be omitted
		ProbPositiveSharpe: 0.5,
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")
	cfg := output.Config{
		FilePath:       path,
		Bootstrap:      result,
		BootstrapSeed:  1,
		BootstrapNSims: 1000,
	}

	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	bs, ok := raw["bootstrap"].(map[string]any)
	if !ok {
		t.Fatalf("'bootstrap' missing from JSON; full JSON:\n%s", data)
	}
	if _, ok := bs["sharpe_p5"]; !ok {
		t.Errorf("sharpe_p5 missing even though value is 0.0 — omitempty must not suppress valid zero")
	}
}

// TestWrite_Bootstrap_ExistingMetricsUnchanged verifies that adding bootstrap fields
// does not affect existing analytics.Report fields in the JSON output.
func TestWrite_Bootstrap_ExistingMetricsUnchanged(t *testing.T) {
	report := analytics.Report{TradeCount: 42, SharpeRatio: 1.23, TotalPnL: 9999.50}
	result := &montecarlo.BootstrapResult{SharpeP5: 0.15, ProbPositiveSharpe: 0.95}
	dir := t.TempDir()
	path := filepath.Join(dir, "out.json")

	if err := output.Write(report, output.Config{
		FilePath:       path,
		Bootstrap:      result,
		BootstrapSeed:  42,
		BootstrapNSims: 5000,
	}); err != nil {
		t.Fatalf("Write: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	tcRaw, ok := raw["TradeCount"].(float64)
	if !ok {
		t.Fatalf("TradeCount missing or not float64; keys: %v", keysOf(raw))
	}
	if int(tcRaw) != 42 {
		t.Errorf("TradeCount = %d, want 42 after bootstrap fields added", int(tcRaw))
	}
	srRaw, ok := raw["SharpeRatio"].(float64)
	if !ok {
		t.Fatalf("SharpeRatio missing or not float64; keys: %v", keysOf(raw))
	}
	if srRaw != 1.23 {
		t.Errorf("SharpeRatio = %v, want 1.23 after bootstrap fields added", srRaw)
	}
}

// keysOf returns the keys of a map for diagnostic error messages.
func keysOf(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// --- WriteSweep DSR section ---

func TestWriteSweep_DSRSection(t *testing.T) {
	report := sweep.Report{
		ParameterName: "period",
		Results: []sweep.Result{
			{ParamValue: 14, SharpeRatio: 1.5, TotalPnL: 5000, TradeCount: 20, MaxDrawdown: 8.5},
			{ParamValue: 12, SharpeRatio: 1.2, TotalPnL: 4200, TradeCount: 22, MaxDrawdown: 9.1},
		},
		VariantCount:  5,
		NObservations: 252,
	}
	var buf bytes.Buffer
	if err := output.WriteSweep(&buf, report); err != nil {
		t.Fatalf("WriteSweep: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"DSR", "variants", "obs"} {
		if !strings.Contains(out, want) {
			t.Errorf("DSR section missing %q:\n%s", want, out)
		}
	}
}

// --- Bootstrap write-error paths ---

// failAfterNWriter fails after exactly N successful writes.
type failAfterNWriter struct {
	n   int
	saw int
}

func (f *failAfterNWriter) Write(p []byte) (int, error) {
	if f.saw >= f.n {
		return 0, errors.New("write failed")
	}
	f.saw++
	return len(p), nil
}

func TestWrite_Bootstrap_WriteError(t *testing.T) {
	tests := []struct {
		name       string
		failAfterN int // number of writes that succeed before failure
	}{
		{"header_fails", 2},   // summary + (bootstrap header fails)
		{"sharpe_fails", 3},   // summary + bootstrap header + (sharpe line fails)
		{"drawdown_fails", 4}, // summary + header + sharpe + (drawdown line fails)
	}
	result := &montecarlo.BootstrapResult{ProbPositiveSharpe: 0.7}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := output.Config{
				PrintToStdout:  true,
				Stdout:         &failAfterNWriter{n: tt.failAfterN},
				Bootstrap:      result,
				BootstrapSeed:  1,
				BootstrapNSims: 1000,
			}
			if err := output.Write(analytics.Report{}, cfg); err == nil {
				t.Errorf("failAfterN=%d: expected error, got nil", tt.failAfterN)
			}
		})
	}
}
