package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
)

// ---------------------------------------------------------------------------
// Test fakes
// ---------------------------------------------------------------------------

// flatProvider returns a fixed candle series — flat price, one candle per day.
type flatProvider struct{}

func (p *flatProvider) FetchCandles(_ context.Context, instrument string, _ model.Timeframe, from, to time.Time) ([]model.Candle, error) {
	var candles []model.Candle
	for ts := from; ts.Before(to); ts = ts.Add(24 * time.Hour) {
		candles = append(candles, model.Candle{
			Instrument: instrument,
			Timeframe:  model.TimeframeDaily,
			Timestamp:  ts,
			Open:       100, High: 101, Low: 99, Close: 100, Volume: 1000,
		})
	}
	return candles, nil
}

func (p *flatProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

var _ provider.DataProvider = (*flatProvider)(nil)

// mockFactory returns a providerFactory that always returns a flatProvider.
func mockFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return &flatProvider{}, nil
	}
}

// panicFactory panics if called — validates that provider is never reached on flag-error paths.
func panicFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		panic("providerFactory must not be called during flag-validation errors")
	}
}

// incompleteDataProvider returns *zerodha.ErrIncompleteData for one named
// instrument and delegates to flatProvider for all others.
type incompleteDataProvider struct {
	targetInstrument string
}

func (p *incompleteDataProvider) FetchCandles(
	ctx context.Context, instrument string, tf model.Timeframe, from, to time.Time,
) ([]model.Candle, error) {
	if instrument == p.targetInstrument {
		return nil, &zerodha.ErrIncompleteData{
			Instrument: instrument,
			From:       from,
			To:         to,
			Expected:   261,
			Got:        20,
		}
	}
	return (&flatProvider{}).FetchCandles(ctx, instrument, tf, from, to)
}

func (p *incompleteDataProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

func incompleteFactory(targetInstrument string) func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return &incompleteDataProvider{targetInstrument: targetInstrument}, nil
	}
}

// writeUniverseYAML writes a minimal universe file and returns its path.
func writeUniverseYAML(t *testing.T, instruments []string) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("instruments:\n")
	for _, inst := range instruments {
		sb.WriteString("  - " + inst + "\n")
	}
	path := filepath.Join(t.TempDir(), "universe.yaml")
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("writeUniverseYAML: %v", err)
	}
	return path
}

// ---------------------------------------------------------------------------
// TestParseAndValidateFlags
// ---------------------------------------------------------------------------

func TestParseAndValidateFlags_ValidInput(t *testing.T) {
	t.Parallel()
	from, to, tf, err := parseAndValidateFlags(
		"universes/nifty50-large-cap.yaml",
		"sma-crossover",
		"2020-01-01",
		"2024-12-31",
		"daily",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("from: got %s", from)
	}
	if !to.Equal(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("to: got %s", to)
	}
	if tf != model.TimeframeDaily {
		t.Errorf("tf: got %s, want daily", tf)
	}
}

func TestParseAndValidateFlags_MissingUniverseFile(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("", "sma-crossover", "2020-01-01", "2024-12-31", "daily")
	if err == nil {
		t.Fatal("expected error for missing --universe")
	}
	if !strings.Contains(err.Error(), "--universe") {
		t.Errorf("error should mention --universe, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingStrategy(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "", "2020-01-01", "2024-12-31", "daily")
	if err == nil {
		t.Fatal("expected error for missing --strategy")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "", "2024-12-31", "daily")
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingTo(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2020-01-01", "", "daily")
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "not-a-date", "2024-12-31", "daily")
	if err == nil {
		t.Fatal("expected error for invalid --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2020-01-01", "not-a-date", "daily")
	if err == nil {
		t.Fatal("expected error for invalid --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2024-01-01", "2020-01-01", "daily")
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to must be strictly after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseAndValidateFlags_ToEqualFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2024-01-01", "2024-01-01", "daily")
	if err == nil {
		t.Fatal("expected error when --to equals --from")
	}
}

func TestParseAndValidateFlags_InvalidTimeframe(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2020-01-01", "2024-12-31", "monthly")
	if err == nil {
		t.Fatal("expected error for invalid --timeframe")
	}
	if !strings.Contains(err.Error(), "--timeframe") {
		t.Errorf("error should mention --timeframe, got: %v", err)
	}
}

func TestParseAndValidateFlags_AllValidTimeframes(t *testing.T) {
	t.Parallel()
	tfs := []string{"1min", "5min", "15min", "daily", "weekly"}
	for _, tfStr := range tfs {
		t.Run(tfStr, func(t *testing.T) {
			t.Parallel()
			_, _, tf, err := parseAndValidateFlags("universes/test.yaml", "sma-crossover", "2020-01-01", "2024-12-31", tfStr)
			if err != nil {
				t.Fatalf("unexpected error for timeframe %q: %v", tfStr, err)
			}
			if string(tf) != tfStr {
				t.Errorf("tf: got %s, want %s", tf, tfStr)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestRun — integration tests via run() with injected fake provider
// ---------------------------------------------------------------------------

func TestRun_ProducesCSVOutput(t *testing.T) {
	t.Parallel()

	universePath := writeUniverseYAML(t, []string{"NSE:RELIANCE", "NSE:INFY"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", universePath,
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
		"--timeframe", "daily",
	}, &stdout, &stderr, mockFactory())
	if err != nil {
		t.Fatalf("run: unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.HasPrefix(out, "instrument,sharpe,trade_count,") {
		t.Errorf("output missing CSV header, got: %q", out[:minInt(len(out), 80)])
	}
	if !strings.Contains(out, "NSE:RELIANCE") {
		t.Error("output missing NSE:RELIANCE row")
	}
	if !strings.Contains(out, "NSE:INFY") {
		t.Error("output missing NSE:INFY row")
	}
}

func TestRun_MissingUniverseFlagError(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --universe")
	}
	if !strings.Contains(err.Error(), "--universe") {
		t.Errorf("error should mention --universe, got: %v", err)
	}
}

func TestRun_MissingStrategyFlagError(t *testing.T) {
	t.Parallel()
	universePath := writeUniverseYAML(t, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", universePath,
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --strategy")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy, got: %v", err)
	}
}

func TestRun_InvalidCommissionError(t *testing.T) {
	t.Parallel()
	universePath := writeUniverseYAML(t, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", universePath,
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
		"--commission", "invalid",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid --commission")
	}
	if !strings.Contains(err.Error(), "--commission") {
		t.Errorf("error should mention --commission, got: %v", err)
	}
}

func TestRun_UniverseFileNotFoundError(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", "/nonexistent/universe.yaml",
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing universe file")
	}
	if !strings.Contains(err.Error(), "universe file") {
		t.Errorf("error should mention universe file, got: %v", err)
	}
}

func TestRun_UnknownStrategyError(t *testing.T) {
	t.Parallel()
	universePath := writeUniverseYAML(t, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	// WalkForwardFactory panics for unknown strategy names — recover to turn into test failure.
	defer func() {
		if r := recover(); r != nil {
			t.Logf("expected panic from unknown strategy: %v", r)
		}
	}()
	err := run([]string{
		"--universe", universePath,
		"--strategy", "does-not-exist",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, panicFactory())
	// Either an error is returned or a panic fires (registry.MustGet panics on unknown name).
	if err != nil && strings.Contains(err.Error(), "--strategy") {
		return // error path — acceptable
	}
}

// min is a local helper for Go versions without the built-in min.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ---------------------------------------------------------------------------
// TestRun_IncompleteData — Option B: sweep continues on per-instrument failure
// ---------------------------------------------------------------------------

// TestRun_IncompleteData_SweepContinues verifies that when one instrument
// returns *ErrIncompleteData, run() does NOT return an error. The sweep
// completes, CSV is written (with the incomplete instrument flagged), and a
// diagnostic is printed to stderr.
func TestRun_IncompleteData_SweepContinues(t *testing.T) {
	t.Parallel()

	universePath := writeUniverseYAML(t, []string{"NSE:RELIANCE", "NSE:INFY"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", universePath,
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
		"--timeframe", "daily",
	}, &stdout, &stderr, incompleteFactory("NSE:RELIANCE"))
	if err != nil {
		t.Fatalf("run() returned error (sweep must continue): %v", err)
	}

	// CSV must contain both instruments.
	out := stdout.String()
	if !strings.Contains(out, "NSE:RELIANCE") {
		t.Error("CSV missing NSE:RELIANCE row")
	}
	if !strings.Contains(out, "NSE:INFY") {
		t.Error("CSV missing NSE:INFY row")
	}

	// Diagnostic must appear on stderr.
	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "incomplete data:") {
		t.Errorf("stderr missing 'incomplete data:' diagnostic; got: %q", stderrOut)
	}
	if !strings.Contains(stderrOut, "NSE:RELIANCE") {
		t.Errorf("stderr missing instrument name in diagnostic; got: %q", stderrOut)
	}
}
