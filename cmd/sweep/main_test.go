package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// ---------------------------------------------------------------------------
// Test fakes
// ---------------------------------------------------------------------------

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

func mockFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return &flatProvider{}, nil
	}
}

func panicFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		panic("providerFactory must not be called during flag-validation errors")
	}
}

// ---------------------------------------------------------------------------
// TestParseAndValidateFlags
// ---------------------------------------------------------------------------

func TestParseAndValidateFlags_ValidInput(t *testing.T) {
	t.Parallel()
	from, to, tf, err := parseAndValidateFlags(
		"2020-01-01", "2024-12-31", "daily",
		"sma-crossover", "fast-period",
		1.0, 5.0, 50.0,
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

func TestParseAndValidateFlags_MissingFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("", "2024-12-31", "daily", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingTo(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "", "daily", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingStrategy(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "2024-12-31", "daily", "", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for missing --strategy")
	}
	if !strings.Contains(err.Error(), "--strategy") {
		t.Errorf("error should mention --strategy, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingSweepParam(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "2024-12-31", "daily", "sma-crossover", "", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for missing --sweep-param")
	}
	if !strings.Contains(err.Error(), "--sweep-param") {
		t.Errorf("error should mention --sweep-param, got: %v", err)
	}
}

func TestParseAndValidateFlags_ZeroStep(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "2024-12-31", "daily", "sma-crossover", "fast-period", 0.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for --step=0")
	}
	if !strings.Contains(err.Error(), "--step") {
		t.Errorf("error should mention --step, got: %v", err)
	}
}

func TestParseAndValidateFlags_MinNotLessThanMax(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "2024-12-31", "daily", "sma-crossover", "fast-period", 1.0, 50.0, 5.0)
	if err == nil {
		t.Fatal("expected error when --min >= --max")
	}
	if !strings.Contains(err.Error(), "--min") {
		t.Errorf("error should mention --min, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("not-a-date", "2024-12-31", "daily", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for invalid --from date")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "bad-date", "daily", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for invalid --to date")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2024-01-01", "2020-01-01", "daily", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to must be strictly after") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidTimeframe(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags("2020-01-01", "2024-12-31", "monthly", "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
	if err == nil {
		t.Fatal("expected error for invalid --timeframe")
	}
	if !strings.Contains(err.Error(), "--timeframe") {
		t.Errorf("error should mention --timeframe, got: %v", err)
	}
}

func TestParseAndValidateFlags_AllValidTimeframes(t *testing.T) {
	t.Parallel()
	for _, tfStr := range []string{"1min", "5min", "15min", "daily", "weekly"} {
		t.Run(tfStr, func(t *testing.T) {
			t.Parallel()
			_, _, tf, err := parseAndValidateFlags("2020-01-01", "2024-12-31", tfStr, "sma-crossover", "fast-period", 1.0, 5.0, 50.0)
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

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--instrument", "NSE:RELIANCE",
		"--strategy", "sma-crossover",
		"--sweep-param", "fast-period",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--timeframe", "daily",
		"--min", "5",
		"--max", "20",
		"--step", "5",
	}, &stdout, &stderr, mockFactory())
	if err != nil {
		t.Fatalf("run: unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "fast-period") {
		t.Errorf("CSV output missing sweep param column, got: %q", out[:minInt(len(out), 120)])
	}
}

func TestRun_MissingFromFlagError(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--sweep-param", "fast-period",
		"--to", "2023-01-01",
		"--min", "5", "--max", "20", "--step", "5",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestRun_InvalidCommissionError(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--strategy", "sma-crossover",
		"--sweep-param", "fast-period",
		"--from", "2020-01-01",
		"--to", "2023-01-01",
		"--min", "5", "--max", "20", "--step", "5",
		"--commission", "invalid",
	}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid --commission")
	}
	if !strings.Contains(err.Error(), "--commission") {
		t.Errorf("error should mention --commission, got: %v", err)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
