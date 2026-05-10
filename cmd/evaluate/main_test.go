// Tests for cmd/evaluate. Written before implementation (TDD).
//
// Tests cover:
//   - All required-flag validations
//   - --params key=value parsing (valid, invalid format, unknown key ignored)
//   - Zero-survivor early exit: verdict.json with killed_at_universe_gate
//   - out-dir subdirectory creation
//   - Full pipeline path verified via unit-testable helpers
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/walkforward"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// ---------------------------------------------------------------------------
// mockProvider satisfies provider.DataProvider for tests.
// ---------------------------------------------------------------------------

type mockProvider struct {
	candles []model.Candle
	err     error
}

func (m *mockProvider) FetchCandles(_ context.Context, _ string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	return m.candles, m.err
}

func (m *mockProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

var _ provider.DataProvider = (*mockProvider)(nil)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeUniverseYAML writes a minimal universe YAML and returns the path.
func writeUniverseYAML(t *testing.T, dir string, instruments []string) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("instruments:\n")
	for _, inst := range instruments {
		sb.WriteString("  - " + inst + "\n")
	}
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("writeUniverseYAML: %v", err)
	}
	return path
}

// mockFactory returns a providerFactory that always returns mock.
func mockFactory(mock *mockProvider) func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return mock, nil
	}
}

// panicFactory is a providerFactory that panics if called — used to ensure
// flag validation errors fire before any provider construction.
func panicFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		panic("providerFactory must not be called during flag-validation errors")
	}
}

// ---------------------------------------------------------------------------
// TestParseFlags — required flags
// ---------------------------------------------------------------------------

func TestRun_MissingStrategy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --strategy")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy; got: %v", err)
	}
}

func TestRun_MissingUniverse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --universe")
	}
	if !strings.Contains(err.Error(), "--universe") {
		t.Errorf("error should mention --universe; got: %v", err)
	}
}

func TestRun_MissingFrom(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--to", "2023-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from; got: %v", err)
	}
}

func TestRun_MissingTo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to; got: %v", err)
	}
}

func TestRun_MissingOutDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --out-dir")
	}
	if !strings.Contains(err.Error(), "--out-dir") {
		t.Errorf("error should mention --out-dir; got: %v", err)
	}
}

func TestRun_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2023-01-01",
		"--to", "2020-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error when --to is before --from")
	}
}

func TestRun_InvalidTimeframe(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--timeframe", "not-a-timeframe",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid timeframe")
	}
	if !strings.Contains(err.Error(), "not-a-timeframe") {
		t.Errorf("error should mention the invalid timeframe; got: %v", err)
	}
}

func TestRun_InvalidFromDate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	universe := writeUniverseYAML(t, dir, []string{"NSE:INFY"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "not-a-date",
		"--to", "2023-01-01",
		"--out-dir", dir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid --from date")
	}
}

// ---------------------------------------------------------------------------
// TestParseParams
// ---------------------------------------------------------------------------

func TestParseParams_Valid(t *testing.T) {
	t.Parallel()
	p, err := parseParams([]string{"fast-period=10", "slow-period=20"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p["fast-period"] != 10.0 {
		t.Errorf("fast-period: got %v, want 10", p["fast-period"])
	}
	if p["slow-period"] != 20.0 {
		t.Errorf("slow-period: got %v, want 20", p["slow-period"])
	}
}

func TestParseParams_Empty(t *testing.T) {
	t.Parallel()
	p, err := parseParams(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p) != 0 {
		t.Errorf("expected empty map, got %v", p)
	}
}

func TestParseParams_InvalidFormat(t *testing.T) {
	t.Parallel()
	_, err := parseParams([]string{"no-equals-sign"})
	if err == nil {
		t.Fatal("expected error for missing '=' in param")
	}
	if !strings.Contains(err.Error(), "no-equals-sign") {
		t.Errorf("error should mention the bad param; got: %v", err)
	}
}

func TestParseParams_NonNumericValue(t *testing.T) {
	t.Parallel()
	_, err := parseParams([]string{"fast-period=abc"})
	if err == nil {
		t.Fatal("expected error for non-numeric value")
	}
}

func TestParseParams_FloatValue(t *testing.T) {
	t.Parallel()
	p, err := parseParams([]string{"bb-num-std-dev=2.5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p["bb-num-std-dev"] != 2.5 {
		t.Errorf("bb-num-std-dev: got %v, want 2.5", p["bb-num-std-dev"])
	}
}

// ---------------------------------------------------------------------------
// TestMakeStageDir
// ---------------------------------------------------------------------------

func TestMakeStageDir_CreatesSubdirectory(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	stageDir, err := makeStageDir(outDir, "sma-crossover", "daily")
	if err != nil {
		t.Fatalf("makeStageDir: %v", err)
	}
	if _, err := os.Stat(stageDir); os.IsNotExist(err) {
		t.Errorf("stage directory was not created: %s", stageDir)
	}
	// Directory name should contain strategy and timeframe.
	base := filepath.Base(stageDir)
	if !strings.Contains(base, "sma-crossover") {
		t.Errorf("stage dir name should contain strategy; got: %s", base)
	}
	if !strings.Contains(base, "daily") {
		t.Errorf("stage dir name should contain timeframe; got: %s", base)
	}
}

func TestMakeStageDir_InvalidParent(t *testing.T) {
	t.Parallel()
	_, err := makeStageDir("/nonexistent/path/that/does/not/exist", "sma", "daily")
	if err == nil {
		t.Fatal("expected error for non-existent parent directory")
	}
}

// ---------------------------------------------------------------------------
// TestWriteVerdictJSON
// ---------------------------------------------------------------------------

func TestWriteVerdictJSON_KilledAtUniverseGate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	v := &verdictJSON{
		Result:   "killed_at_universe_gate",
		Strategy: "sma-crossover",
		From:     "2020-01-01",
		To:       "2023-01-01",
	}
	if err := writeVerdictJSON(dir, v); err != nil {
		t.Fatalf("writeVerdictJSON: %v", err)
	}
	path := filepath.Join(dir, "verdict.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\ndata: %s", err, data)
	}
	if decoded["result"] != "killed_at_universe_gate" {
		t.Errorf("result: got %v, want killed_at_universe_gate", decoded["result"])
	}
	if decoded["strategy"] != "sma-crossover" {
		t.Errorf("strategy: got %v, want sma-crossover", decoded["strategy"])
	}
}

func TestWriteVerdictJSON_ValidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	v := &verdictJSON{
		Result:    "pass",
		Strategy:  "macd-crossover",
		From:      "2018-01-01",
		To:        "2025-01-01",
		Survivors: []survivorRecord{{Instrument: "NSE:SBIN", UniverseSharpe: 0.5, WFPassed: true, BootstrapP5: 0.05, BootstrapProbPositive: 0.92}},
		Kills:     []killRecord{{Instrument: "NSE:INFY", Stage: "universe_gate", Reason: "InsufficientData"}},
	}
	if err := writeVerdictJSON(dir, v); err != nil {
		t.Fatalf("writeVerdictJSON: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "verdict.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\ndata: %s", err, data)
	}
	if decoded["result"] != "pass" {
		t.Errorf("result: got %v", decoded["result"])
	}
}

// ---------------------------------------------------------------------------
// TestApplyWFGate — WF instrument-count gate (60% floor, min 6)
// ---------------------------------------------------------------------------

func TestApplyWFGate_Passes(t *testing.T) {
	t.Parallel()
	// 9 of 14 = 64.3% >= 60% AND >= 6 → PASS
	passed, failed := applyWFGate(14, []string{
		"A", "B", "C", "D", "E", "F", "G", "H", "I",
	}, []string{"J", "K", "L", "M", "N"})
	if len(passed) != 9 {
		t.Errorf("passed: got %d, want 9", len(passed))
	}
	if len(failed) != 5 {
		t.Errorf("failed: got %d, want 5", len(failed))
	}
}

func TestApplyWFGate_FailsBelowFloor(t *testing.T) {
	t.Parallel()
	// 4 of 12 = 33% < 60% → FAIL
	passed, failed := applyWFGate(12, []string{"A", "B", "C", "D"}, []string{"E", "F", "G", "H", "I", "J", "K", "L"})
	if len(passed) != 0 {
		t.Errorf("should return no survivors when gate fails; got %d", len(passed))
	}
	if len(failed) != 12 {
		t.Errorf("all instruments should be in failed list; got %d", len(failed))
	}
}

func TestApplyWFGate_FailsBelowMin6(t *testing.T) {
	t.Parallel()
	// 5 of 7 = 71% >= 60%, but 5 < 6 min → FAIL
	passed, failed := applyWFGate(7, []string{"A", "B", "C", "D", "E"}, []string{"F", "G"})
	if len(passed) != 0 {
		t.Errorf("should return no survivors when WF pass count < 6; got %d", len(passed))
	}
	_ = failed
}

func TestApplyWFGate_ExactlyAtFloor(t *testing.T) {
	t.Parallel()
	// floor(0.60 * 10) = 6 → need 6; have exactly 6 → PASS
	passed, _ := applyWFGate(10, []string{"A", "B", "C", "D", "E", "F"}, []string{"G", "H", "I", "J"})
	if len(passed) != 6 {
		t.Errorf("exactly at floor: got %d survivors, want 6", len(passed))
	}
}

// ---------------------------------------------------------------------------
// TestApplyBootstrapGate
// ---------------------------------------------------------------------------

func TestApplyBootstrapGate_Passes(t *testing.T) {
	t.Parallel()
	// SharpeP5 > 0 AND ProbPositiveSharpe > 0.80 → PASS
	ok := applyBootstrapGate(0.05, 0.95)
	if !ok {
		t.Error("expected gate pass for SharpeP5=0.05, ProbPositiveSharpe=0.95")
	}
}

func TestApplyBootstrapGate_FailsP5Zero(t *testing.T) {
	t.Parallel()
	ok := applyBootstrapGate(0.0, 0.95)
	if ok {
		t.Error("expected gate fail for SharpeP5=0.0 (must be strictly > 0)")
	}
}

func TestApplyBootstrapGate_FailsP5Negative(t *testing.T) {
	t.Parallel()
	ok := applyBootstrapGate(-0.1, 0.95)
	if ok {
		t.Error("expected gate fail for SharpeP5=-0.1")
	}
}

func TestApplyBootstrapGate_FailsProbBelowThreshold(t *testing.T) {
	t.Parallel()
	ok := applyBootstrapGate(0.05, 0.79)
	if ok {
		t.Error("expected gate fail for ProbPositiveSharpe=0.79 (must be > 0.80)")
	}
}

func TestApplyBootstrapGate_FailsBoth(t *testing.T) {
	t.Parallel()
	ok := applyBootstrapGate(-0.1, 0.5)
	if ok {
		t.Error("expected gate fail when both conditions fail")
	}
}

// ---------------------------------------------------------------------------
// TestRun_ZeroSurvivorEarlyExit
// ---------------------------------------------------------------------------

// flatCandles returns n daily candles at a constant price for the given instrument.
// A flat price series produces zero SMA-crossover signals → zero trades →
// TradeMetricsInsufficient=true → InsufficientData=true → gate fails.
func flatCandles(t *testing.T, instrument string, n int) []model.Candle {
	t.Helper()
	candles := make([]model.Candle, n)
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range n {
		c, err := model.NewCandle(instrument, model.TimeframeDaily,
			start.AddDate(0, 0, i),
			100.0, 101.0, 99.0, 100.0, 1000.0)
		if err != nil {
			t.Fatalf("flatCandles: %v", err)
		}
		candles[i] = c
	}
	return candles
}

// TestRun_ZeroSurvivors verifies that when the universe sweep produces zero
// survivors (all instruments have InsufficientData or negative Sharpe, causing
// DSRAverageSharpe <= 0), the pipeline halts and writes verdict.json with
// result="killed_at_universe_gate". It does NOT proceed to walk-forward.
func TestRun_ZeroSurvivors(t *testing.T) {
	t.Parallel()

	outDir := t.TempDir()
	universeDir := t.TempDir()
	universe := writeUniverseYAML(t, universeDir, []string{"NSE:INFY"})

	// mockProvider returns flat candles. sma-crossover (fast=10, slow=20) on a
	// constant price line produces zero crossovers → zero trades →
	// TradeMetricsInsufficient=true → InsufficientData=true →
	// DSRAverageSharpe=0 → GatePass=false.
	mock := &mockProvider{candles: flatCandles(t, "NSE:INFY", 60)}
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--out-dir", outDir,
	}, &stdout, &stderr, mockFactory(mock))
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	// Stage directory is created.
	entries, readErr := os.ReadDir(outDir)
	if readErr != nil {
		t.Fatalf("ReadDir: %v", readErr)
	}
	if len(entries) == 0 {
		t.Fatal("expected stage directory to be created in out-dir")
	}

	// Find verdict.json — it may be at outDir/verdict.json or inside stage dir.
	// We look for it at the stage dir level.
	stageDir := filepath.Join(outDir, entries[0].Name())
	verdictPath := filepath.Join(stageDir, "verdict.json")
	data, err := os.ReadFile(verdictPath)
	if err != nil {
		t.Fatalf("verdict.json not found at %s: %v", verdictPath, err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("verdict.json is not valid JSON: %v\ndata: %s", err, data)
	}
	result, ok := decoded["result"].(string)
	if !ok {
		t.Fatalf("result field missing or not a string in verdict.json; got: %v", decoded)
	}
	if result != "killed_at_universe_gate" {
		t.Errorf("result: got %q, want %q", result, "killed_at_universe_gate")
	}
}

// ---------------------------------------------------------------------------
// TestWfGateFailed2Kills
// ---------------------------------------------------------------------------

func TestWfGateFailed2Kills_ReturnsNonPassed(t *testing.T) {
	t.Parallel()
	universe := []string{"A", "B", "C", "D", "E"}
	passed := []string{"B", "D"}
	failed := wfGateFailed2Kills(universe, passed)
	if len(failed) != 3 {
		t.Fatalf("got %d failures, want 3; failed=%v", len(failed), failed)
	}
	want := map[string]struct{}{"A": {}, "C": {}, "E": {}}
	for _, f := range failed {
		if _, ok := want[f]; !ok {
			t.Errorf("unexpected instrument in failed: %s", f)
		}
	}
}

func TestWfGateFailed2Kills_AllPassed(t *testing.T) {
	t.Parallel()
	universe := []string{"A", "B"}
	passed := []string{"A", "B"}
	failed := wfGateFailed2Kills(universe, passed)
	if len(failed) != 0 {
		t.Errorf("expected empty failed list; got %v", failed)
	}
}

func TestWfGateFailed2Kills_NonePassed(t *testing.T) {
	t.Parallel()
	universe := []string{"A", "B", "C"}
	passed := []string{}
	failed := wfGateFailed2Kills(universe, passed)
	if len(failed) != 3 {
		t.Errorf("expected all 3 in failed; got %d", len(failed))
	}
}

// ---------------------------------------------------------------------------
// TestWfKillReason
// ---------------------------------------------------------------------------

func TestWfKillReason_OverfitOnly(t *testing.T) {
	t.Parallel()
	r := walkforward.Report{OverfitFlag: true, NegativeFoldFlag: false}
	got := wfKillReason(r)
	if got != "OverfitFlag" {
		t.Errorf("got %q, want OverfitFlag", got)
	}
}

func TestWfKillReason_NegativeFoldOnly(t *testing.T) {
	t.Parallel()
	r := walkforward.Report{OverfitFlag: false, NegativeFoldFlag: true}
	got := wfKillReason(r)
	if got != "NegativeFoldFlag" {
		t.Errorf("got %q, want NegativeFoldFlag", got)
	}
}

func TestWfKillReason_BothFlags(t *testing.T) {
	t.Parallel()
	r := walkforward.Report{OverfitFlag: true, NegativeFoldFlag: true}
	got := wfKillReason(r)
	if got != "OverfitFlag+NegativeFoldFlag" {
		t.Errorf("got %q, want OverfitFlag+NegativeFoldFlag", got)
	}
}

// ---------------------------------------------------------------------------
// TestSanitizeInstrument
// ---------------------------------------------------------------------------

func TestSanitizeInstrument_ReplacesColon(t *testing.T) {
	t.Parallel()
	got := sanitizeInstrument("NSE:INFY")
	if got != "NSE_INFY" {
		t.Errorf("got %q, want NSE_INFY", got)
	}
}

func TestSanitizeInstrument_ReplacesSpaceAndSlash(t *testing.T) {
	t.Parallel()
	got := sanitizeInstrument("NSE/NIFTY 50")
	if got != "NSE_NIFTY_50" {
		t.Errorf("got %q, want NSE_NIFTY_50", got)
	}
}

func TestSanitizeInstrument_NoChange(t *testing.T) {
	t.Parallel()
	got := sanitizeInstrument("INFY")
	if got != "INFY" {
		t.Errorf("got %q, want INFY", got)
	}
}

// ---------------------------------------------------------------------------
// TestWriteFoldsCSV
// ---------------------------------------------------------------------------

func TestWriteFoldsCSV_HeaderOnly(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := writeFoldsCSV(&buf, nil); err != nil {
		t.Fatalf("writeFoldsCSV: %v", err)
	}
	out := buf.String()
	wantHeader := "fold_index,is_start,is_end,oos_start,oos_end,is_sharpe,oos_sharpe,trade_count,degenerate"
	if !strings.Contains(out, wantHeader) {
		t.Errorf("missing CSV header; got:\n%s", out)
	}
}

func TestWriteFoldsCSV_DataRow(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	windows := []walkforward.WindowResult{
		{
			InSampleStart:     time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleEnd:       time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleStart:  time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleEnd:    time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleSharpe:    1.2,
			OutOfSampleSharpe: 0.8,
			TradeCount:        10,
			Degenerate:        false,
		},
	}
	if err := writeFoldsCSV(&buf, windows); err != nil {
		t.Fatalf("writeFoldsCSV: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "2020-01-01") {
		t.Errorf("expected IS start date in output; got:\n%s", out)
	}
	if !strings.Contains(out, "false") {
		t.Errorf("expected degenerate=false in output; got:\n%s", out)
	}
}

func TestWriteFoldsCSV_WriteError(t *testing.T) {
	t.Parallel()
	err := writeFoldsCSV(&alwaysErrWriter{}, nil)
	if err == nil {
		t.Fatal("expected error from write failure, got nil")
	}
}

// alwaysErrWriter always returns an error from Write.
type alwaysErrWriter struct{}

func (w *alwaysErrWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("simulated write error")
}

// ---------------------------------------------------------------------------
// TestWriteWFJSON
// ---------------------------------------------------------------------------

func TestWriteWFJSON_WritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	report := walkforward.Report{
		AvgInSampleSharpe:     1.0,
		AvgOutOfSampleSharpe:  0.7,
		DeduplicatedFoldCount: 2,
		OverfitFlag:           false,
		NegativeFoldFlag:      false,
	}
	if err := writeWFJSON(dir, "NSE:INFY", report); err != nil {
		t.Fatalf("writeWFJSON: %v", err)
	}
	// File should exist with sanitized name.
	path := filepath.Join(dir, "wf-NSE_INFY.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\ndata: %s", err, data)
	}
	if _, ok := decoded["OverfitFlag"]; !ok {
		t.Errorf("JSON missing OverfitFlag field; got:\n%s", data)
	}
}

func TestWriteWFJSON_InvalidDir(t *testing.T) {
	t.Parallel()
	err := writeWFJSON("/nonexistent/dir", "NSE:INFY", walkforward.Report{})
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

// ---------------------------------------------------------------------------
// TestWriteWFFoldsCSV
// ---------------------------------------------------------------------------

func TestWriteWFFoldsCSV_WritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	windows := []walkforward.WindowResult{
		{
			InSampleStart:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			InSampleEnd:      time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleStart: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			OutOfSampleEnd:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			TradeCount:       5,
			Degenerate:       false,
		},
	}
	if err := writeWFFoldsCSV(dir, "NSE:SBIN", windows); err != nil {
		t.Fatalf("writeWFFoldsCSV: %v", err)
	}
	path := filepath.Join(dir, "wf-NSE_SBIN-folds.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	out := string(data)
	if !strings.Contains(out, "fold_index") {
		t.Errorf("expected CSV header; got:\n%s", out)
	}
	if !strings.Contains(out, "2020-01-01") {
		t.Errorf("expected IS start date; got:\n%s", out)
	}
}

func TestWriteWFFoldsCSV_InvalidDir(t *testing.T) {
	t.Parallel()
	err := writeWFFoldsCSV("/nonexistent/dir", "NSE:INFY", nil)
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

// ---------------------------------------------------------------------------
// TestWriteBootstrapJSON
// ---------------------------------------------------------------------------

func TestWriteBootstrapJSON_WritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	result := montecarlo.BootstrapResult{
		MeanSharpe:         0.3,
		SharpeP5:           0.05,
		SharpeP50:          0.28,
		SharpeP95:          0.65,
		WorstDrawdownP5:    2.1,
		WorstDrawdownP50:   5.3,
		WorstDrawdownP95:   12.7,
		ProbPositiveSharpe: 0.92,
	}
	if err := writeBootstrapJSON(dir, "NSE:TCS", result); err != nil {
		t.Fatalf("writeBootstrapJSON: %v", err)
	}
	path := filepath.Join(dir, "bootstrap-NSE_TCS.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("invalid JSON: %v\ndata: %s", err, data)
	}
	p5, ok := decoded["sharpe_p5"].(float64)
	if !ok {
		t.Fatalf("sharpe_p5 missing or wrong type; got: %v", decoded["sharpe_p5"])
	}
	if p5 != 0.05 {
		t.Errorf("sharpe_p5: got %v, want 0.05", p5)
	}
	probPos, ok2 := decoded["prob_positive_sharpe"].(float64)
	if !ok2 {
		t.Fatalf("prob_positive_sharpe missing or wrong type; got: %v", decoded["prob_positive_sharpe"])
	}
	if probPos != 0.92 {
		t.Errorf("prob_positive_sharpe: got %v, want 0.92", probPos)
	}
}

func TestWriteBootstrapJSON_InvalidDir(t *testing.T) {
	t.Parallel()
	err := writeBootstrapJSON("/nonexistent/dir", "NSE:INFY", montecarlo.BootstrapResult{})
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

// ---------------------------------------------------------------------------
// TestWriteUniverseSweepCSV
// ---------------------------------------------------------------------------

func TestWriteUniverseSweepCSV_WritesFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	report := universesweep.Report{
		Results: []universesweep.Result{
			{Instrument: "NSE:INFY", Sharpe: 0.5, TradeCount: 35, TotalPnL: 5000, MaxDrawdown: 0.08, InsufficientData: false},
			{Instrument: "NSE:TCS", Sharpe: -0.1, TradeCount: 30, TotalPnL: -1000, MaxDrawdown: 0.15, InsufficientData: true},
		},
	}
	if err := writeUniverseSweepCSV(dir, report); err != nil {
		t.Fatalf("writeUniverseSweepCSV: %v", err)
	}
	path := filepath.Join(dir, "universe-sweep.csv")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	out := string(data)
	if !strings.Contains(out, "instrument") {
		t.Errorf("expected CSV header; got:\n%s", out)
	}
	if !strings.Contains(out, "NSE:INFY") {
		t.Errorf("expected NSE:INFY in output; got:\n%s", out)
	}
}

func TestWriteUniverseSweepCSV_InvalidDir(t *testing.T) {
	t.Parallel()
	err := writeUniverseSweepCSV("/nonexistent/dir", universesweep.Report{})
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

// ---------------------------------------------------------------------------
// TestRun_ProviderError
// ---------------------------------------------------------------------------

// TestRun_ProviderError verifies that a provider construction failure is
// propagated as an error from run(). This exercises buildPipeline's provider
// error path without requiring a live provider.
func TestRun_ProviderError(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	universeDir := t.TempDir()
	universe := writeUniverseYAML(t, universeDir, []string{"NSE:INFY"})

	errFactory := func(_ context.Context) (provider.DataProvider, error) {
		return nil, fmt.Errorf("injected provider construction failure")
	}
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--out-dir", outDir,
	}, &stdout, &stderr, errFactory)
	if err == nil {
		t.Fatal("expected error from provider failure, got nil")
	}
	if !strings.Contains(err.Error(), "provider") {
		t.Errorf("error should mention provider; got: %v", err)
	}
}

// TestRun_InvalidCommission verifies that an invalid --commission flag
// is caught before provider construction.
func TestRun_InvalidCommission(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	universeDir := t.TempDir()
	universe := writeUniverseYAML(t, universeDir, []string{"NSE:INFY"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--commission", "not-a-model",
		"--out-dir", outDir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid commission model")
	}
}

// TestRun_InvalidParams verifies that a malformed --params value is caught.
func TestRun_InvalidParams(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	universeDir := t.TempDir()
	universe := writeUniverseYAML(t, universeDir, []string{"NSE:INFY"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--params", "no-equals-sign",
		"--out-dir", outDir,
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for malformed --params")
	}
}

// ---------------------------------------------------------------------------
// TestRun_UnknownStrategy — panics (MustGet behavior)
// ---------------------------------------------------------------------------

func TestRun_UnknownStrategy(t *testing.T) {
	t.Parallel()
	outDir := t.TempDir()
	universeDir := t.TempDir()
	universe := writeUniverseYAML(t, universeDir, []string{"NSE:INFY"})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for unknown strategy, got none")
		}
	}()
	var stdout, stderr bytes.Buffer
	_ = run([]string{ //nolint:errcheck // expect panic before return
		"--strategy", "no-such-strategy",
		"--universe", universe,
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--out-dir", outDir,
	}, &stdout, &stderr, panicFactory())
}
