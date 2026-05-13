package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// stubStrategy is a minimal Strategy for cmd-layer tests.
type stubStrategy struct{}

func (s *stubStrategy) Name() string                       { return "stub" }
func (s *stubStrategy) Lookback() int                      { return 1 }
func (s *stubStrategy) Timeframe() model.Timeframe         { return model.TimeframeDaily }
func (s *stubStrategy) Next(_ []model.Candle) model.Signal { return model.SignalHold }

// mockProvider returns scripted candles by instrument.
type mockProvider struct {
	candles map[string][]model.Candle
}

func (m *mockProvider) FetchCandles(_ context.Context, instrument string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	if c, ok := m.candles[instrument]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("instrument not found: %s", instrument)
}

func (m *mockProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// makeCandles produces n daily candles with OHLCV=100.
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

// makeGridFile writes a param-grid JSON to a temp file and returns its path.
func makeGridFile(t *testing.T, spec any) string {
	t.Helper()
	b, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal grid spec: %v", err)
	}
	f, err := os.CreateTemp(t.TempDir(), "grid*.json")
	if err != nil {
		t.Fatalf("create temp grid file: %v", err)
	}
	if _, err := f.Write(b); err != nil {
		t.Fatalf("write grid file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

// makeUniverseFile writes a universe YAML to a temp file and returns its path.
func makeUniverseFile(t *testing.T, instruments []string) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("instruments:\n")
	for _, inst := range instruments {
		fmt.Fprintf(&sb, "  - %s\n", inst)
	}
	f, err := os.CreateTemp(t.TempDir(), "universe*.yaml")
	if err != nil {
		t.Fatalf("create temp universe file: %v", err)
	}
	if _, err := f.WriteString(sb.String()); err != nil {
		t.Fatalf("write universe file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

// TestRun_NoOOSFlag verifies that --oos-from and --oos-to flags do not exist.
// This is the architectural enforcement test: the binary must not accept OOS flags
// per Marcus standing order (param search on training window only).
func TestRun_NoOOSFlag(t *testing.T) {
	fs := flag.NewFlagSet("param-search", flag.ContinueOnError)
	registerFlags(fs)

	// Attempt to look up --oos-from and --oos-to by parsing them; they must be rejected.
	err := fs.Parse([]string{"--oos-from", "2024-01-01"})
	if err == nil {
		t.Error("expected error when passing --oos-from (flag must not exist), got nil")
	}

	err = fs.Parse([]string{"--oos-to", "2024-12-31"})
	if err == nil {
		t.Error("expected error when passing --oos-to (flag must not exist), got nil")
	}
}

// TestRun_MissingRequiredFlags verifies that missing required flags produce errors.
func TestRun_MissingRequiredFlags(t *testing.T) {
	outDir := t.TempDir()
	gridFile := makeGridFile(t, map[string]any{
		"axes": []map[string]any{
			{"name": "p1", "min": 1.0, "max": 2.0, "step": 1.0},
		},
	})
	universeFile := makeUniverseFile(t, []string{"NSE:A"})
	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	tests := []struct {
		name        string
		args        []string
		wantErrFrag string
	}{
		{
			name: "missing strategy",
			args: []string{
				"--param-grid", gridFile,
				"--universe", universeFile,
				"--train-from", "2020-01-02",
				"--train-to", "2021-06-01",
				"--out-dir", outDir,
			},
			wantErrFrag: "--strategy",
		},
		{
			name: "missing param-grid",
			args: []string{
				"--strategy", "macd-crossover",
				"--universe", universeFile,
				"--train-from", "2020-01-02",
				"--train-to", "2021-06-01",
				"--out-dir", outDir,
			},
			wantErrFrag: "--param-grid",
		},
		{
			name: "missing universe",
			args: []string{
				"--strategy", "macd-crossover",
				"--param-grid", gridFile,
				"--train-from", "2020-01-02",
				"--train-to", "2021-06-01",
				"--out-dir", outDir,
			},
			wantErrFrag: "--universe",
		},
		{
			name: "missing train-from",
			args: []string{
				"--strategy", "macd-crossover",
				"--param-grid", gridFile,
				"--universe", universeFile,
				"--train-to", "2021-06-01",
				"--out-dir", outDir,
			},
			wantErrFrag: "--train-from",
		},
		{
			name: "missing train-to",
			args: []string{
				"--strategy", "macd-crossover",
				"--param-grid", gridFile,
				"--universe", universeFile,
				"--train-from", "2020-01-02",
				"--out-dir", outDir,
			},
			wantErrFrag: "--train-to",
		},
		{
			name: "missing out-dir",
			args: []string{
				"--strategy", "macd-crossover",
				"--param-grid", gridFile,
				"--universe", universeFile,
				"--train-from", "2020-01-02",
				"--train-to", "2021-06-01",
			},
			wantErrFrag: "--out-dir",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := run(tc.args, &stdout, &stderr, factory)
			if err == nil {
				t.Errorf("expected error containing %q, got nil", tc.wantErrFrag)
				return
			}
			if !strings.Contains(err.Error(), tc.wantErrFrag) {
				t.Errorf("error %q does not contain %q", err.Error(), tc.wantErrFrag)
			}
		})
	}
}

// TestRun_InvalidGridJSON verifies that a malformed param-grid JSON file returns
// an error at parse time, before any engine runs.
func TestRun_InvalidGridJSON(t *testing.T) {
	outDir := t.TempDir()

	// Write invalid JSON
	gridFile := filepath.Join(t.TempDir(), "bad-grid.json")
	if err := os.WriteFile(gridFile, []byte("{not valid json"), 0o644); err != nil {
		t.Fatal(err)
	}
	universeFile := makeUniverseFile(t, []string{"NSE:A"})
	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "macd-crossover",
		"--param-grid", gridFile,
		"--universe", universeFile,
		"--train-from", "2020-01-02",
		"--train-to", "2021-06-01",
		"--out-dir", outDir,
	}, &stdout, &stderr, factory)
	if err == nil {
		t.Error("expected error for invalid grid JSON, got nil")
	}
}

// TestRun_InvalidTrainDateRange verifies that train-to <= train-from produces an error.
func TestRun_InvalidTrainDateRange(t *testing.T) {
	outDir := t.TempDir()
	gridFile := makeGridFile(t, map[string]any{
		"axes": []map[string]any{
			{"name": "p1", "min": 1.0, "max": 2.0, "step": 1.0},
		},
	})
	universeFile := makeUniverseFile(t, []string{"NSE:A"})
	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "macd-crossover",
		"--param-grid", gridFile,
		"--universe", universeFile,
		"--train-from", "2021-06-01",
		"--train-to", "2020-01-02", // to before from
		"--out-dir", outDir,
	}, &stdout, &stderr, factory)
	if err == nil {
		t.Error("expected error for invalid date range, got nil")
	}
}

// TestRun_WritesCSV verifies the happy path: run produces param-search-results.csv
// with the expected headers and at least one row.
func TestRun_WritesCSV(t *testing.T) {
	outDir := t.TempDir()
	gridFile := makeGridFile(t, map[string]any{
		"axes": []map[string]any{
			{"name": "fast-period", "min": 5.0, "max": 10.0, "step": 5.0},
		},
	})
	universeFile := makeUniverseFile(t, []string{"NSE:A"})

	// Use a strategy that generates real trades so we get non-insufficient results.
	// Override the strategy factory via providerFactory injection.
	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "macd-crossover",
		"--param-grid", gridFile,
		"--universe", universeFile,
		"--train-from", "2020-01-02",
		"--train-to", "2021-06-01",
		"--out-dir", outDir,
		"--top-n", "10",
	}, &stdout, &stderr, factory)
	if err != nil {
		t.Fatalf("run: %v (stderr: %s)", err, stderr.String())
	}

	// Verify output file exists.
	csvPath := filepath.Join(outDir, "param-search-results.csv")
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open results CSV: %v", err)
	}
	defer f.Close() //nolint:errcheck // read-only file opened in test; close error is non-fatal

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read results CSV: %v", err)
	}

	// Must have header + at least 1 data row.
	if len(records) < 2 {
		t.Fatalf("expected at least 2 rows (header + data), got %d", len(records))
	}

	// Verify header contains required columns.
	header := records[0]
	requiredCols := []string{"params", "dsr_sharpe", "raw_sharpe", "trade_count", "insufficient_data"}
	for _, col := range requiredCols {
		found := false
		for _, h := range header {
			if h == col {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CSV header missing required column %q; got: %v", col, header)
		}
	}
}

// TestRun_OutDirNotExist verifies that a non-existent --out-dir returns an error
// before engine runs begin.
func TestRun_OutDirNotExist(t *testing.T) {
	gridFile := makeGridFile(t, map[string]any{
		"axes": []map[string]any{
			{"name": "p1", "min": 1.0, "max": 2.0, "step": 1.0},
		},
	})
	universeFile := makeUniverseFile(t, []string{"NSE:A"})
	p := &mockProvider{candles: map[string][]model.Candle{}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "macd-crossover",
		"--param-grid", gridFile,
		"--universe", universeFile,
		"--train-from", "2020-01-02",
		"--train-to", "2021-06-01",
		"--out-dir", "/nonexistent/path/that/cannot/exist/abc123",
	}, &stdout, &stderr, factory)
	if err == nil {
		t.Error("expected error for non-existent out-dir, got nil")
	}
}

// TestRun_TopNLimitsOutput verifies that --top-n=1 produces at most 2 rows in CSV
// (header + 1 data row).
func TestRun_TopNLimitsOutput(t *testing.T) {
	outDir := t.TempDir()
	gridFile := makeGridFile(t, map[string]any{
		"axes": []map[string]any{
			// 3 variants
			{"name": "fast-period", "min": 5.0, "max": 15.0, "step": 5.0},
		},
	})
	universeFile := makeUniverseFile(t, []string{"NSE:A"})
	p := &mockProvider{candles: map[string][]model.Candle{"NSE:A": makeCandles(300)}}
	factory := func(_ context.Context) (provider.DataProvider, error) { return p, nil }

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "macd-crossover",
		"--param-grid", gridFile,
		"--universe", universeFile,
		"--train-from", "2020-01-02",
		"--train-to", "2021-06-01",
		"--out-dir", outDir,
		"--top-n", "1",
	}, &stdout, &stderr, factory)
	if err != nil {
		t.Fatalf("run: %v", err)
	}

	csvPath := filepath.Join(outDir, "param-search-results.csv")
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open results CSV: %v", err)
	}
	defer f.Close() //nolint:errcheck // read-only file opened in test; close error is non-fatal

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("read CSV: %v", err)
	}
	// header + at most 1 data row
	if len(records) > 2 {
		t.Errorf("expected at most 2 rows with --top-n=1, got %d", len(records))
	}
}

// TestRegisterFlags_RequiredFlagsExist verifies that all required flags are registered.
func TestRegisterFlags_RequiredFlagsExist(t *testing.T) {
	fs := flag.NewFlagSet("param-search", flag.ContinueOnError)
	registerFlags(fs)

	required := []string{"strategy", "param-grid", "universe", "timeframe", "train-from", "train-to", "out-dir", "top-n", "commission"}
	for _, name := range required {
		if fs.Lookup(name) == nil {
			t.Errorf("required flag --%s not registered", name)
		}
	}
}

// TestUnusedInterfaceCompilationCheck ensures mockProvider satisfies provider.DataProvider
// at compile time. This is a compile-time check — no runtime test needed.
var _ provider.DataProvider = (*mockProvider)(nil)

// TestUnusedInterfaceCompilationCheck2 ensures stubStrategy satisfies strategy.Strategy
// at compile time.
var _ strategy.Strategy = (*stubStrategy)(nil)
