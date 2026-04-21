package output_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

func TestLoadCurveCSV_RoundTrip(t *testing.T) {
	pts := []model.EquityPoint{
		{Timestamp: time.Date(2020, 1, 1, 18, 30, 0, 0, time.UTC), Value: 100000.00},
		{Timestamp: time.Date(2020, 1, 2, 18, 30, 0, 0, time.UTC), Value: 101500.50},
		{Timestamp: time.Date(2020, 1, 3, 18, 30, 0, 0, time.UTC), Value: 99800.25},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "curve.csv")

	cfg := output.Config{CurvePath: path, Curve: pts}
	if err := output.Write(analytics.Report{}, cfg); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := output.LoadCurveCSV(path)
	if err != nil {
		t.Fatalf("LoadCurveCSV: %v", err)
	}

	if len(got) != len(pts) {
		t.Fatalf("len: got %d, want %d", len(got), len(pts))
	}
	for i, want := range pts {
		if !got[i].Timestamp.Equal(want.Timestamp) {
			t.Errorf("[%d] Timestamp: got %v, want %v", i, got[i].Timestamp, want.Timestamp)
		}
		const tol = 0.005 // CSV rounds to 2 decimal places
		if diff := got[i].Value - want.Value; diff < -tol || diff > tol {
			t.Errorf("[%d] Value: got %.2f, want %.2f", i, got[i].Value, want.Value)
		}
	}
}

func TestLoadCurveCSV_HeaderOnly_EmptySlice(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.csv")
	if err := os.WriteFile(path, []byte("timestamp,equity_value\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := output.LoadCurveCSV(path)
	if err != nil {
		t.Fatalf("LoadCurveCSV: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("len: got %d, want 0", len(got))
	}
}

func TestLoadCurveCSV_MalformedTimestamp_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.csv")
	content := "timestamp,equity_value\nnot-a-timestamp,100000.00\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := output.LoadCurveCSV(path)
	if err == nil {
		t.Error("LoadCurveCSV: got nil error, want parse error")
	}
}

func TestLoadCurveCSV_MalformedValue_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "badval.csv")
	content := "timestamp,equity_value\n2020-01-01T18:30:00Z,not-a-float\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	_, err := output.LoadCurveCSV(path)
	if err == nil {
		t.Error("LoadCurveCSV: got nil error, want parse error")
	}
}

func TestLoadCurveCSV_FileNotFound_ReturnsError(t *testing.T) {
	_, err := output.LoadCurveCSV("/nonexistent/path/curve.csv")
	if err == nil {
		t.Error("LoadCurveCSV: got nil error, want file-not-found error")
	}
}
