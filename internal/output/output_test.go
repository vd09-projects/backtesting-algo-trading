package output_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
)

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
