package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/walkforward"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// TestParseAndValidateFlags
// ---------------------------------------------------------------------------

func TestParseAndValidateFlags_ValidInput(t *testing.T) {
	t.Parallel()
	from, to, err := parseAndValidateFlags(
		"NSE:TCS", "2018-01-01", "2024-12-31", "sma-crossover",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("from: got %s", from)
	}
	if !to.Equal(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("to: got %s", to)
	}
}

func TestParseAndValidateFlags_MissingInstrument(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("", "2018-01-01", "2024-12-31", "sma-crossover")
	if err == nil {
		t.Fatal("expected error for missing instrument")
	}
	if !strings.Contains(err.Error(), "--instrument") {
		t.Errorf("error should mention --instrument, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingFrom(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "", "2024-12-31", "sma-crossover")
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingTo(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "2018-01-01", "", "sma-crossover")
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingStrategy(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "2018-01-01", "2024-12-31", "")
	if err == nil {
		t.Fatal("expected error for missing --strategy")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "not-a-date", "2024-12-31", "sma-crossover")
	if err == nil {
		t.Fatal("expected error for invalid --from")
	}
}

func TestParseAndValidateFlags_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "2018-01-01", "not-a-date", "sma-crossover")
	if err == nil {
		t.Fatal("expected error for invalid --to")
	}
}

func TestParseAndValidateFlags_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "2024-01-01", "2018-01-01", "sma-crossover")
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_EqualFromTo(t *testing.T) {
	t.Parallel()
	_, _, err := parseAndValidateFlags("NSE:TCS", "2018-01-01", "2018-01-01", "sma-crossover")
	if err == nil {
		t.Fatal("expected error when --from == --to")
	}
}

// ---------------------------------------------------------------------------
// TestWalkForwardFactory (replaces TestStrategyFactory — delegates to registry)
// ---------------------------------------------------------------------------

func TestWalkForwardFactory_KnownStrategies(t *testing.T) {
	t.Parallel()
	knownStrategies := []string{
		"sma-crossover",
		"rsi-mean-reversion",
		"donchian-breakout",
		"macd-crossover",
		"bollinger-mean-reversion",
		"momentum",
		"cci-mean-reversion",
	}
	for _, name := range knownStrategies {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := cmdutil.GlobalRegistry.DefaultParams(name)
			factory, err := cmdutil.GlobalRegistry.WalkForwardFactory(name, model.TimeframeDaily, p)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", name, err)
			}
			if factory == nil {
				t.Fatalf("%s: factory must not be nil", name)
			}
			s := factory()
			if s == nil {
				t.Fatalf("%s: factory() returned nil", name)
			}
		})
	}
}

func TestWalkForwardFactory_UnknownStrategy(t *testing.T) {
	t.Parallel()
	// WalkForwardFactory returns an error for unknown names (unlike MustGet which panics).
	// The panic path is handled by the explicit MustGet call in run() before
	// WalkForwardFactory is invoked.
	_, err := cmdutil.GlobalRegistry.WalkForwardFactory("not-a-strategy", model.TimeframeDaily, nil)
	if err == nil {
		t.Fatal("expected error for unknown strategy, got nil")
	}
	if !strings.Contains(err.Error(), "not-a-strategy") {
		t.Errorf("error should mention the unknown strategy name, got: %v", err)
	}
}

func TestWalkForwardFactory_InvalidParams_EagerValidation(t *testing.T) {
	t.Parallel()
	// fast-period >= slow-period is invalid for sma-crossover; should error at
	// factory construction time (eager validation), not panic inside the closure.
	bad := map[string]float64{
		"fast-period": 50,
		"slow-period": 10,
	}
	_, err := cmdutil.GlobalRegistry.WalkForwardFactory("sma-crossover", model.TimeframeDaily, bad)
	if err == nil {
		t.Fatal("expected error for invalid sma-crossover params (fast >= slow), got nil")
	}
}

// ---------------------------------------------------------------------------
// TestBuildWalkForwardConfig
// ---------------------------------------------------------------------------

func TestBuildWalkForwardConfig_YearsToDuration(t *testing.T) {
	t.Parallel()
	from := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := buildWalkForwardConfig("NSE:TCS", from, to, 2, 1, 1)

	wantIS := 2 * 365 * 24 * time.Hour
	wantOOS := 365 * 24 * time.Hour
	wantStep := 365 * 24 * time.Hour

	if cfg.InSampleWindow != wantIS {
		t.Errorf("InSampleWindow: got %v, want %v", cfg.InSampleWindow, wantIS)
	}
	if cfg.OutOfSampleWindow != wantOOS {
		t.Errorf("OutOfSampleWindow: got %v, want %v", cfg.OutOfSampleWindow, wantOOS)
	}
	if cfg.StepSize != wantStep {
		t.Errorf("StepSize: got %v, want %v", cfg.StepSize, wantStep)
	}
	if cfg.Instrument != "NSE:TCS" {
		t.Errorf("Instrument: got %s", cfg.Instrument)
	}
	if !cfg.From.Equal(from) {
		t.Errorf("From: got %s", cfg.From)
	}
	if !cfg.To.Equal(to) {
		t.Errorf("To: got %s", cfg.To)
	}
}

func TestBuildWalkForwardConfig_Defaults(t *testing.T) {
	t.Parallel()
	from := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := buildWalkForwardConfig("NSE:INFY", from, to, 2, 1, 1)
	if cfg.InSampleWindow != 2*365*24*time.Hour {
		t.Errorf("default IS window wrong: got %v", cfg.InSampleWindow)
	}
	if cfg.OutOfSampleWindow != 365*24*time.Hour {
		t.Errorf("default OOS window wrong: got %v", cfg.OutOfSampleWindow)
	}
	if cfg.StepSize != 365*24*time.Hour {
		t.Errorf("default step wrong: got %v", cfg.StepSize)
	}
}

// ---------------------------------------------------------------------------
// TestDetermineExitCode
// ---------------------------------------------------------------------------

func TestDetermineExitCode_Clean(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{
		OverfitFlag:      false,
		NegativeFoldFlag: false,
	}
	if code := determineExitCode(report); code != 0 {
		t.Errorf("clean report: expected exit 0, got %d", code)
	}
}

func TestDetermineExitCode_OverfitFlag(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{
		OverfitFlag:      true,
		NegativeFoldFlag: false,
	}
	if code := determineExitCode(report); code != 1 {
		t.Errorf("OverfitFlag=true: expected exit 1, got %d", code)
	}
}

func TestDetermineExitCode_NegativeFoldFlag(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{
		OverfitFlag:      false,
		NegativeFoldFlag: true,
	}
	if code := determineExitCode(report); code != 1 {
		t.Errorf("NegativeFoldFlag=true: expected exit 1, got %d", code)
	}
}

func TestDetermineExitCode_BothFlags(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{
		OverfitFlag:      true,
		NegativeFoldFlag: true,
	}
	if code := determineExitCode(report); code != 1 {
		t.Errorf("both flags: expected exit 1, got %d", code)
	}
}

// ---------------------------------------------------------------------------
// TestFormatFoldsCSV
// ---------------------------------------------------------------------------

func TestFormatFoldsCSV_HeaderRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	windows := []walkforward.WindowResult{}
	if err := writeFoldsCSV(&buf, windows); err != nil {
		t.Fatalf("writeFoldsCSV: %v", err)
	}
	output := buf.String()
	wantHeader := "fold_index,is_start,is_end,oos_start,oos_end,is_sharpe,oos_sharpe,trade_count,degenerate"
	if !strings.Contains(output, wantHeader) {
		t.Errorf("missing header; got:\n%s", output)
	}
}

func TestFormatFoldsCSV_DataRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	windows := []walkforward.WindowResult{
		{
			InSampleStart:     time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleEnd:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleStart:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleEnd:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleSharpe:    1.2,
			OutOfSampleSharpe: 0.8,
			TradeCount:        15,
			Degenerate:        false,
		},
	}
	if err := writeFoldsCSV(&buf, windows); err != nil {
		t.Fatalf("writeFoldsCSV: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "0,") {
		t.Errorf("expected fold_index=0 in output; got:\n%s", output)
	}
	if !strings.Contains(output, "2018-01-01") {
		t.Errorf("expected IS start date in output; got:\n%s", output)
	}
	if !strings.Contains(output, "false") {
		t.Errorf("expected degenerate=false in output; got:\n%s", output)
	}
}

func TestFormatFoldsCSV_DegenerateRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	windows := []walkforward.WindowResult{
		{
			InSampleStart:    time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleEnd:      time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleStart: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleEnd:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			TradeCount:       0,
			Degenerate:       true,
		},
	}
	if err := writeFoldsCSV(&buf, windows); err != nil {
		t.Fatalf("writeFoldsCSV: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "true") {
		t.Errorf("expected degenerate=true in output; got:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// TestWriteReportJSON
// ---------------------------------------------------------------------------

func TestWriteReportJSON_ValidReport(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{
		AvgInSampleSharpe:     1.5,
		AvgOutOfSampleSharpe:  0.9,
		DeduplicatedFoldCount: 3,
		OverfitFlag:           false,
		NegativeFoldFlag:      false,
		NegativeFoldCount:     0,
		Windows: []walkforward.WindowResult{
			{
				InSampleStart:     time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
				InSampleEnd:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				OutOfSampleStart:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				OutOfSampleEnd:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				InSampleSharpe:    1.5,
				OutOfSampleSharpe: 0.9,
				TradeCount:        10,
				Degenerate:        false,
			},
		},
	}

	var buf bytes.Buffer
	if err := writeReportJSON(&buf, report); err != nil {
		t.Fatalf("writeReportJSON: %v", err)
	}

	output := buf.String()
	var decoded map[string]any
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput:\n%s", err, output)
	}
	if output[len(output)-1] != '\n' {
		t.Errorf("output must end with newline; got last byte %q", output[len(output)-1])
	}
	if _, ok := decoded["OverfitFlag"]; !ok {
		t.Errorf("JSON missing OverfitFlag field; got:\n%s", output)
	}
}

func TestWriteReportJSON_WriteError(t *testing.T) {
	t.Parallel()
	report := walkforward.Report{}
	if err := writeReportJSON(&errWriter{}, report); err == nil {
		t.Fatal("expected error from write failure, got nil")
	}
}

type errWriter struct{}

func (e *errWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

// ---------------------------------------------------------------------------
// TestWriteFoldsCSVFile
// ---------------------------------------------------------------------------

func TestWriteFoldsCSVFile_HappyPath(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp(t.TempDir(), "folds-*.csv")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		t.Fatalf("Close temp file: %v", err)
	}

	windows := []walkforward.WindowResult{
		{
			InSampleStart:     time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleEnd:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleStart:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleEnd:    time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleSharpe:    1.1,
			OutOfSampleSharpe: 0.7,
			TradeCount:        12,
			Degenerate:        false,
		},
	}

	if err := writeFoldsCSVFile(path, windows); err != nil {
		t.Fatalf("writeFoldsCSVFile: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	out := string(content)

	wantHeader := "fold_index,is_start,is_end,oos_start,oos_end,is_sharpe,oos_sharpe,trade_count,degenerate"
	if !strings.Contains(out, wantHeader) {
		t.Errorf("missing CSV header; got:\n%s", out)
	}
	if !strings.Contains(out, "2018-01-01") {
		t.Errorf("expected IS start date 2018-01-01 in CSV; got:\n%s", out)
	}
	if !strings.Contains(out, "false") {
		t.Errorf("expected degenerate=false in CSV; got:\n%s", out)
	}
}

func TestWriteFoldsCSVFile_CreateFailure(t *testing.T) {
	t.Parallel()
	err := writeFoldsCSVFile("/nonexistent/dir/out.csv", nil)
	if err == nil {
		t.Fatal("expected error for invalid path, got nil")
	}
}

// ---------------------------------------------------------------------------
// TestRun
// ---------------------------------------------------------------------------

func TestRun_MissingRequiredFlags(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when required flags are missing, got nil")
	}
	if !strings.Contains(err.Error(), "--instrument") {
		t.Errorf("error should mention --instrument; got: %v", err)
	}
}

func TestRun_UnknownStrategy(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown strategy, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value is not a string: %T: %v", r, r)
		}
		if !strings.Contains(msg, "no-such-strategy") {
			t.Errorf("panic message should mention the strategy name; got: %v", msg)
		}
	}()
	var stdout, stderr bytes.Buffer
	args := []string{
		"--instrument", "NSE:TCS",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--strategy", "no-such-strategy",
	}
	_ = run(args, &stdout, &stderr) //nolint:errcheck // expect panic before return
}

func TestRun_InvalidCommission(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	args := []string{
		"--instrument", "NSE:TCS",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--strategy", "sma-crossover",
		"--commission", "not-a-model",
	}
	err := run(args, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid commission model, got nil")
	}
}

func TestRun_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	args := []string{
		"--instrument", "NSE:TCS",
		"--from", "2023-01-01",
		"--to", "2020-01-01",
		"--strategy", "sma-crossover",
	}
	err := run(args, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error when --to is before --from, got nil")
	}
}

func TestRun_InvalidStrategyParams(t *testing.T) {
	t.Parallel()
	// fast-period >= slow-period is rejected by smacrossover.New; WalkForwardFactory
	// validates params eagerly and returns an error before any I/O.
	var stdout, stderr bytes.Buffer
	args := []string{
		"--instrument", "NSE:TCS",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--strategy", "sma-crossover",
		"--fast-period", "50",
		"--slow-period", "10",
	}
	err := run(args, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid sma-crossover params (fast >= slow), got nil")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy; got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// exitCodeError
// ---------------------------------------------------------------------------

func TestExitCodeError_Error(t *testing.T) {
	t.Parallel()
	e := &exitCodeError{code: 1}
	got := e.Error()
	if !strings.Contains(got, "1") {
		t.Errorf("Error() = %q, want it to contain the exit code", got)
	}
}

// ---------------------------------------------------------------------------
// writeReportJSON newline error path
// ---------------------------------------------------------------------------

// afterFirstWriter succeeds on the first Write call and fails on all subsequent.
type afterFirstWriter struct{ calls int }

func (w *afterFirstWriter) Write(p []byte) (int, error) {
	w.calls++
	if w.calls > 1 {
		return 0, fmt.Errorf("simulated write error on call %d", w.calls)
	}
	return len(p), nil
}

func TestWriteReportJSON_NewlineError(t *testing.T) {
	t.Parallel()
	// First Write (JSON body) succeeds; second Write (trailing newline) fails.
	w := &afterFirstWriter{}
	err := writeReportJSON(w, walkforward.Report{})
	if err == nil {
		t.Fatal("expected error when newline write fails, got nil")
	}
}

// ---------------------------------------------------------------------------
// writeFoldsCSV write-error path
// ---------------------------------------------------------------------------

func TestWriteFoldsCSV_WriteError(t *testing.T) {
	t.Parallel()
	err := writeFoldsCSV(&errWriter{}, []walkforward.WindowResult{})
	if err == nil {
		t.Fatal("expected error from write failure, got nil")
	}
}
