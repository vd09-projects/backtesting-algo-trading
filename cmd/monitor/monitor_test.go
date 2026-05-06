package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// baseTime is an arbitrary fixed time used across tests.
var baseTime = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

// makeTrade builds a model.Trade where ReturnOnNotional == r exactly.
// EntryPrice=100, Quantity=1 → notional=100 → RealizedPnL = r * 100.
func makeTrade(r float64, exitOffset time.Duration) model.Trade {
	return model.Trade{
		Instrument:  "NSE:TEST",
		Direction:   model.DirectionLong,
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   100 + r*100,
		EntryTime:   baseTime,
		ExitTime:    baseTime.Add(exitOffset),
		Commission:  0,
		RealizedPnL: r * 100,
	}
}

// writeTradesFile serializes trades to a temp JSON file and returns its path.
func writeTradesFile(t *testing.T, trades []model.Trade) string {
	t.Helper()
	data, err := json.Marshal(trades)
	if err != nil {
		t.Fatalf("marshal trades: %v", err)
	}
	path := filepath.Join(t.TempDir(), "trades.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write trades file: %v", err)
	}
	return path
}

// writeThresholdsFile writes a thresholdsFile JSON to a temp file and returns its path.
func writeThresholdsFile(t *testing.T, sharpeP5, maxDDPct float64, maxDDDurNs int64) string {
	t.Helper()
	tf := thresholdsFile{
		SharpeP5:        sharpeP5,
		MaxDrawdownPct:  maxDDPct,
		MaxDDDurationNs: maxDDDurNs,
	}
	data, err := json.Marshal(tf)
	if err != nil {
		t.Fatalf("marshal thresholds: %v", err)
	}
	path := filepath.Join(t.TempDir(), "thresholds.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write thresholds file: %v", err)
	}
	return path
}

// --- Flag validation tests ---

func TestRun_MissingTradesFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{"--thresholds", "t.json"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing --trades, got nil")
	}
	if !strings.Contains(err.Error(), "--trades") {
		t.Errorf("error %q should mention --trades", err.Error())
	}
}

func TestRun_MissingThresholdsFlag(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", "t.json"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing --thresholds, got nil")
	}
	if !strings.Contains(err.Error(), "--thresholds") {
		t.Errorf("error %q should mention --thresholds", err.Error())
	}
}

func TestRun_TradesFileNotFound(t *testing.T) {
	t.Parallel()
	tf := writeThresholdsFile(t, -1.0, 50.0, int64(100*24*time.Hour))
	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", "/nonexistent/trades.json", "--thresholds", tf}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing trades file")
	}
}

func TestRun_ThresholdsFileNotFound(t *testing.T) {
	t.Parallel()
	trades := writeTradesFile(t, []model.Trade{makeTrade(0.1, time.Hour)})
	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", trades, "--thresholds", "/nonexistent/thresholds.json"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for missing thresholds file")
	}
}

func TestRun_InvalidTradesJSON(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	badTrades := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badTrades, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	tf := writeThresholdsFile(t, -1.0, 50.0, int64(100*24*time.Hour))
	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", badTrades, "--thresholds", tf}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for invalid trades JSON")
	}
}

// --- Alert output and exit code tests ---

func TestRun_AllGreen(t *testing.T) {
	t.Parallel()
	// Four trades with positive returns → high per-trade Sharpe, no drawdown.
	trades := []model.Trade{
		makeTrade(0.10, 1*24*time.Hour),
		makeTrade(0.12, 2*24*time.Hour),
		makeTrade(0.11, 3*24*time.Hour),
		makeTrade(0.13, 4*24*time.Hour),
	}
	tf := writeThresholdsFile(t, -2.0, 50.0, int64(365*24*time.Hour))
	tp := writeTradesFile(t, trades)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output, got: %q", out)
	}
	if strings.Contains(out, "HALT") {
		t.Errorf("unexpected HALT in output: %q", out)
	}
}

func TestRun_SharpeBreached_ExitCode1(t *testing.T) {
	t.Parallel()
	// Four trades with negative returns → negative per-trade Sharpe.
	trades := []model.Trade{
		makeTrade(-0.10, 1*24*time.Hour),
		makeTrade(-0.12, 2*24*time.Hour),
		makeTrade(-0.11, 3*24*time.Hour),
		makeTrade(-0.13, 4*24*time.Hour),
	}
	// SharpeP5=0.0 means any negative Sharpe breaches.
	tf := writeThresholdsFile(t, 0.0, 50.0, int64(365*24*time.Hour))
	tp := writeTradesFile(t, trades)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf}, &stdout, &stderr)

	var ee *exitCodeError
	if err == nil {
		t.Fatal("expected exitCodeError for Sharpe breach")
	}
	if !isExitCodeError(err, &ee) || ee.code != 1 {
		t.Fatalf("expected exitCodeError{1}, got: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "HALT (Sharpe breached)") {
		t.Errorf("expected 'HALT (Sharpe breached)' in output, got: %q", out)
	}
}

func TestRun_DrawdownBreached_ExitCode1(t *testing.T) {
	t.Parallel()
	// Trades: big gain then big loss → large drawdown. Returns are positive on average
	// so Sharpe is positive, but equity drops 60% from peak.
	// Trade 1: +50%, trade 2: -60% → equity goes 100→150→60 → drawdown = 60%.
	trades := []model.Trade{
		makeTrade(0.50, 1*24*time.Hour),
		makeTrade(-0.60, 2*24*time.Hour),
	}
	// MaxDrawdownPct=20% (will be breached by 60% drawdown). SharpeP5=very negative so Sharpe OK.
	// Synthetic curve: initialEquity=100 → after trade1: 100+50=150 → after trade2: 150-60=90
	// peak=150, last=90 → drawdown=(150-90)/150*100=40%. Set threshold to 20%.
	tf := writeThresholdsFile(t, -10.0, 20.0, int64(365*24*time.Hour))
	tp := writeTradesFile(t, trades)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf, "--initial-equity", "100"}, &stdout, &stderr)

	var ee *exitCodeError
	if err == nil {
		t.Fatal("expected exitCodeError for drawdown breach")
	}
	if !isExitCodeError(err, &ee) || ee.code != 1 {
		t.Fatalf("expected exitCodeError{1}, got: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "HALT (drawdown breached)") {
		t.Errorf("expected 'HALT (drawdown breached)' in output, got: %q", out)
	}
}

func TestRun_DurationBreached_ExitCode1(t *testing.T) {
	t.Parallel()
	// Equity peaks at trade 1 then never recovers.
	// Trade 1 +10%, trade 2 -5% separated by 200 days.
	// Duration threshold = 100 days → duration of 200 days breaches it.
	t1 := baseTime
	t2 := baseTime.Add(200 * 24 * time.Hour)

	trades := []model.Trade{
		{
			Instrument:  "NSE:TEST",
			Direction:   model.DirectionLong,
			Quantity:    1,
			EntryPrice:  100,
			ExitPrice:   110,
			EntryTime:   t1,
			ExitTime:    t1,
			RealizedPnL: 10,
		},
		{
			Instrument:  "NSE:TEST",
			Direction:   model.DirectionLong,
			Quantity:    1,
			EntryPrice:  100,
			ExitPrice:   95,
			EntryTime:   t2,
			ExitTime:    t2,
			RealizedPnL: -5,
		},
	}
	// MaxDDDurationNs = 100 days in nanoseconds. Sharpe and drawdown thresholds set loose.
	threshold100Days := int64(100 * 24 * time.Hour)
	tf := writeThresholdsFile(t, -10.0, 50.0, threshold100Days)
	tp := writeTradesFile(t, trades)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf, "--initial-equity", "100"}, &stdout, &stderr)

	var ee *exitCodeError
	if err == nil {
		t.Fatal("expected exitCodeError for duration breach")
	}
	if !isExitCodeError(err, &ee) || ee.code != 1 {
		t.Fatalf("expected exitCodeError{1}, got: %v", err)
	}
	out := stdout.String()
	if !strings.Contains(out, "HALT (duration breached)") {
		t.Errorf("expected 'HALT (duration breached)' in output, got: %q", out)
	}
}

func TestRun_MultipleBreaches_AllReported(t *testing.T) {
	t.Parallel()
	// Trades trigger both Sharpe AND drawdown breach.
	trades := []model.Trade{
		makeTrade(-0.10, 1*24*time.Hour),
		makeTrade(-0.50, 2*24*time.Hour),
	}
	// SharpeP5=0 (negative Sharpe breaches), MaxDrawdownPct=5% (will breach).
	tf := writeThresholdsFile(t, 0.0, 5.0, int64(365*24*time.Hour))
	tp := writeTradesFile(t, trades)

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf, "--initial-equity", "100"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected error for breach")
	}
	out := stdout.String()
	if !strings.Contains(out, "HALT (Sharpe breached)") {
		t.Errorf("expected Sharpe HALT in output: %q", out)
	}
	if !strings.Contains(out, "HALT (drawdown breached)") {
		t.Errorf("expected drawdown HALT in output: %q", out)
	}
}

func TestRun_EmptyTradeLog_OK(t *testing.T) {
	t.Parallel()
	// Zero trades: Sharpe undefined (no breach), curve empty (no drawdown).
	tf := writeThresholdsFile(t, 0.0, 5.0, int64(24*time.Hour))
	tp := writeTradesFile(t, []model.Trade{})

	var stdout, stderr bytes.Buffer
	err := run([]string{"--trades", tp, "--thresholds", tf}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("unexpected error for empty trade log: %v", err)
	}
	if !strings.Contains(stdout.String(), "OK") {
		t.Errorf("expected OK for empty trade log, got: %q", stdout.String())
	}
}

// TestBuildSyntheticCurve_Order verifies curve is built in ExitTime order,
// not insertion order.
func TestBuildSyntheticCurve_Order(t *testing.T) {
	t.Parallel()
	// Trades given out-of-order by ExitTime.
	t1 := baseTime
	t2 := baseTime.Add(24 * time.Hour)
	t3 := baseTime.Add(48 * time.Hour)

	trades := []model.Trade{
		{ExitTime: t3, RealizedPnL: 30},
		{ExitTime: t1, RealizedPnL: 10},
		{ExitTime: t2, RealizedPnL: 20},
	}
	initialEquity := 100.0
	curve := buildSyntheticCurve(trades, initialEquity)

	if len(curve) != 3 {
		t.Fatalf("expected 3 curve points, got %d", len(curve))
	}
	// After sorting: t1 → equity=110, t2 → equity=130, t3 → equity=160.
	wantValues := []float64{110, 130, 160}
	wantTimes := []time.Time{t1, t2, t3}
	for i, pt := range curve {
		if pt.Value != wantValues[i] {
			t.Errorf("curve[%d].Value: got %.2f, want %.2f", i, pt.Value, wantValues[i])
		}
		if !pt.Timestamp.Equal(wantTimes[i]) {
			t.Errorf("curve[%d].Timestamp: got %v, want %v", i, pt.Timestamp, wantTimes[i])
		}
	}
}

func TestBuildSyntheticCurve_Empty(t *testing.T) {
	t.Parallel()
	curve := buildSyntheticCurve(nil, 100.0)
	if len(curve) != 0 {
		t.Errorf("expected empty curve for nil trades, got %d points", len(curve))
	}
}

// isExitCodeError type-asserts err to *exitCodeError, writing to dst.
func isExitCodeError(err error, dst **exitCodeError) bool {
	if err == nil {
		return false
	}
	e, ok := err.(*exitCodeError)
	if ok {
		*dst = e
	}
	return ok
}
