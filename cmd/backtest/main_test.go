package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
)

// ---------------------------------------------------------------------------
// TestParseAndValidateFlags
// ---------------------------------------------------------------------------

func TestParseAndValidateFlags_ValidInput(t *testing.T) {
	t.Parallel()
	f := &flags{
		fromStr: "2018-01-01",
		toStr:   "2024-12-31",
		tfStr:   "daily",
	}
	from, to, tf, err := parseAndValidateFlags(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !from.Equal(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
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
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "", toStr: "2024-12-31", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_MissingTo(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "not-a-date", toStr: "2024-12-31", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for invalid --from date")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestParseAndValidateFlags_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "bad", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error for invalid --to date")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestParseAndValidateFlags_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2024-01-01", toStr: "2018-01-01", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to must be strictly after") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseAndValidateFlags_ToEqualFrom(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2024-01-01", toStr: "2024-01-01", tfStr: "daily"})
	if err == nil {
		t.Fatal("expected error when --to equals --from")
	}
}

func TestParseAndValidateFlags_InvalidTimeframe(t *testing.T) {
	t.Parallel()
	_, _, _, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "2024-12-31", tfStr: "monthly"})
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
			_, _, tf, err := parseAndValidateFlags(&flags{fromStr: "2018-01-01", toStr: "2024-12-31", tfStr: tfStr})
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
// TestParseSizingConfig
// ---------------------------------------------------------------------------

func TestParseSizingConfig_Fixed(t *testing.T) {
	t.Parallel()
	sm, err := parseSizingConfig("fixed", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm != model.SizingFixed {
		t.Errorf("got %v, want SizingFixed", sm)
	}
}

func TestParseSizingConfig_VolTargetValid(t *testing.T) {
	t.Parallel()
	sm, err := parseSizingConfig("vol-target", 0.10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sm != model.SizingVolatilityTarget {
		t.Errorf("got %v, want SizingVolatilityTarget", sm)
	}
}

func TestParseSizingConfig_VolTargetZeroVolTarget(t *testing.T) {
	t.Parallel()
	_, err := parseSizingConfig("vol-target", 0)
	if err == nil {
		t.Fatal("expected error for vol-target with vol-target=0")
	}
	if !strings.Contains(err.Error(), "--vol-target") {
		t.Errorf("error should mention --vol-target, got: %v", err)
	}
}

func TestParseSizingConfig_UnknownModel(t *testing.T) {
	t.Parallel()
	_, err := parseSizingConfig("unknown", 0)
	if err == nil {
		t.Fatal("expected error for unknown sizing model")
	}
	if !strings.Contains(err.Error(), "--sizing-model") {
		t.Errorf("error should mention --sizing-model, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// TestRun_ErrIncompleteData — providerFactory injection + exit-code tests
// ---------------------------------------------------------------------------

// incompleteDataFactory returns a providerFactory whose FetchCandles always
// returns *zerodha.ErrIncompleteData for the given instrument.
func incompleteDataFactory(instrument string) func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return &incompleteDataProvider{instrument: instrument}, nil
	}
}

type incompleteDataProvider struct{ instrument string }

func (p *incompleteDataProvider) FetchCandles(
	_ context.Context, _ string, _ model.Timeframe, from, to time.Time,
) ([]model.Candle, error) {
	return nil, &zerodha.ErrIncompleteData{
		Instrument: p.instrument,
		From:       from,
		To:         to,
		Expected:   261,
		Got:        20,
	}
}

func (p *incompleteDataProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// genericErrorFactory returns a providerFactory whose FetchCandles always returns
// a plain, non-typed error — used to verify exit code 1 path.
func genericErrorFactory() func(context.Context) (provider.DataProvider, error) {
	return func(_ context.Context) (provider.DataProvider, error) {
		return &genericErrorProvider{}, nil
	}
}

type genericErrorProvider struct{}

func (p *genericErrorProvider) FetchCandles(
	_ context.Context, _ string, _ model.Timeframe, _, _ time.Time,
) ([]model.Candle, error) {
	return nil, fmt.Errorf("generic provider failure")
}

func (p *genericErrorProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// TestRun_IncompleteData_ReturnsExitCodeError2 verifies that when the provider
// returns *ErrIncompleteData, run() returns *cmdutil.ExitCodeError with Code==2.
func TestRun_IncompleteData_ReturnsExitCodeError2(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--instrument", "NSE:RELIANCE",
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, incompleteDataFactory("NSE:RELIANCE"))

	if err == nil {
		t.Fatal("run() returned nil, want *cmdutil.ExitCodeError")
	}

	var ee *cmdutil.ExitCodeError
	if !errors.As(err, &ee) {
		t.Fatalf("run() error is %T (%v), want *cmdutil.ExitCodeError", err, err)
	}
	if ee.Code != 2 {
		t.Errorf("ExitCodeError.Code = %d, want 2", ee.Code)
	}

	// Diagnostic must appear on stderr.
	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "incomplete data:") {
		t.Errorf("stderr missing 'incomplete data:' diagnostic; got: %q", stderrOut)
	}
	if !strings.Contains(stderrOut, "NSE:RELIANCE") {
		t.Errorf("stderr missing instrument name; got: %q", stderrOut)
	}
}

// TestRun_GenericError_NotExitCodeError2 verifies that a plain (non-typed) provider
// error does NOT produce *cmdutil.ExitCodeError with Code==2 — generic errors must
// keep the standard exit-code-1 path so the codes remain distinct.
func TestRun_GenericError_NotExitCodeError2(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--instrument", "NSE:RELIANCE",
		"--strategy", "stub",
		"--from", "2022-01-01",
		"--to", "2023-01-01",
	}, &stdout, &stderr, genericErrorFactory())

	if err == nil {
		t.Fatal("run() returned nil, want an error for generic provider failure")
	}

	var ee *cmdutil.ExitCodeError
	if errors.As(err, &ee) && ee.Code == 2 {
		t.Errorf("run() returned ExitCodeError{Code:2} for a generic error — must not happen")
	}
}
