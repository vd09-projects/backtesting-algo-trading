package sweep2d_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep2d"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// factory2D ignores p2 and wires p1 as the threshold. Used for the 2×2 grid
// where p1 drives profitable vs. unprofitable behavior.
func thresholdFactory(p1, _ float64) (strategy.Strategy, error) {
	return &testutil.ThresholdStrategy{Threshold: p1, TF: model.TimeframeDaily}, nil
}

func baseConfig() sweep2d.Config2D {
	return sweep2d.Config2D{
		Param1:          sweep2d.ParamRange{Name: "threshold", Min: 80, Max: 100, Step: 20},
		Param2:          sweep2d.ParamRange{Name: "dummy", Min: 1, Max: 2, Step: 1},
		Timeframe:       model.TimeframeDaily,
		EngineConfig:    testutil.TestEngineConfig(),
		StrategyFactory: thresholdFactory,
	}
}

// TestRun_2x2Grid verifies that a 2×2 sweep produces exactly 4 populated cells
// at the correct (param1, param2) positions.
func TestRun_2x2Grid(t *testing.T) {
	t.Parallel()
	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}

	report, err := sweep2d.Run(context.Background(), baseConfig(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(report.Param1Values) != 2 {
		t.Fatalf("Param1Values: got %d, want 2", len(report.Param1Values))
	}
	if len(report.Param2Values) != 2 {
		t.Fatalf("Param2Values: got %d, want 2", len(report.Param2Values))
	}
	if len(report.Grid) != 2 {
		t.Fatalf("Grid rows: got %d, want 2", len(report.Grid))
	}
	for i, row := range report.Grid {
		if len(row) != 2 {
			t.Fatalf("Grid[%d] cols: got %d, want 2", i, len(row))
		}
	}
	if report.VariantCount != 4 {
		t.Errorf("VariantCount: got %d, want 4", report.VariantCount)
	}
}

// TestRun_GridIndicesStable verifies that Grid[i][j].Param1Value equals
// Param1Values[i] and Grid[i][j].Param2Value equals Param2Values[j] for all cells.
func TestRun_GridIndicesStable(t *testing.T) {
	t.Parallel()
	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}

	report, err := sweep2d.Run(context.Background(), baseConfig(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	for i, v1 := range report.Param1Values {
		for j, v2 := range report.Param2Values {
			cell := report.Grid[i][j]
			if cell.Param1Value != v1 {
				t.Errorf("Grid[%d][%d].Param1Value = %g, want %g", i, j, cell.Param1Value, v1)
			}
			if cell.Param2Value != v2 {
				t.Errorf("Grid[%d][%d].Param2Value = %g, want %g", i, j, cell.Param2Value, v2)
			}
		}
	}
}

// TestRun_Determinism verifies that running the same config twice produces
// bit-identical grids.
func TestRun_Determinism(t *testing.T) {
	t.Parallel()
	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}
	cfg := baseConfig()

	r1, err := sweep2d.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	r2, err := sweep2d.Run(context.Background(), cfg, p)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}

	for i := range r1.Grid {
		for j := range r1.Grid[i] {
			c1, c2 := r1.Grid[i][j], r2.Grid[i][j]
			if c1.SharpeRatio != c2.SharpeRatio {
				t.Errorf("Grid[%d][%d].SharpeRatio: run1=%.6f run2=%.6f", i, j, c1.SharpeRatio, c2.SharpeRatio)
			}
			if c1.TradeCount != c2.TradeCount {
				t.Errorf("Grid[%d][%d].TradeCount: run1=%d run2=%d", i, j, c1.TradeCount, c2.TradeCount)
			}
		}
	}
}

// TestRun_DSRLowerForMoreTrials verifies the DSR property directly via analytics.DSR.
func TestRun_DSRLowerForMoreTrials(t *testing.T) {
	t.Parallel()
	// Use a fixed observed Sharpe and nObservations (300 candles, same as sweep).
	dsrFew := analytics.DSR(1.0, 4, 300)
	dsrMany := analytics.DSR(1.0, 100, 300)
	if dsrMany >= dsrFew {
		t.Errorf("DSR(1.0, 100, 300)=%.6f should be < DSR(1.0, 4, 300)=%.6f", dsrMany, dsrFew)
	}
}

// TestRun_PeakSharpePopulated verifies PeakSharpe and DSRCorrectedPeakSharpe are
// computed and the corrected value is lower than the raw peak (>1 variant).
func TestRun_PeakSharpePopulated(t *testing.T) {
	t.Parallel()
	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}

	report, err := sweep2d.Run(context.Background(), baseConfig(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if report.PeakSharpe == 0 && report.VariantCount > 0 {
		t.Error("PeakSharpe is zero — at least one profitable cell expected")
	}
	if report.DSRCorrectedPeakSharpe >= report.PeakSharpe {
		t.Errorf("DSRCorrectedPeakSharpe (%.6f) should be < PeakSharpe (%.6f)",
			report.DSRCorrectedPeakSharpe, report.PeakSharpe)
	}
}

// TestRun_ValidationErrors verifies that malformed configs are rejected.
func TestRun_ValidationErrors(t *testing.T) {
	t.Parallel()
	p := &testutil.StaticProvider{Candles: testutil.MakeAlternatingCandles(10, 120, 80)}
	valid := baseConfig()

	tests := []struct {
		name    string
		modify  func(*sweep2d.Config2D)
		wantErr string
	}{
		{"empty param1 name", func(c *sweep2d.Config2D) { c.Param1.Name = "" }, "Param1.Name"},
		{"empty param2 name", func(c *sweep2d.Config2D) { c.Param2.Name = "" }, "Param2.Name"},
		{"zero step1", func(c *sweep2d.Config2D) { c.Param1.Step = 0 }, "Param1.Step"},
		{"negative step2", func(c *sweep2d.Config2D) { c.Param2.Step = -1 }, "Param2.Step"},
		{"max < min param1", func(c *sweep2d.Config2D) { c.Param1.Max = 50 }, "Param1.Max"},
		{"nil factory", func(c *sweep2d.Config2D) { c.StrategyFactory = nil }, "StrategyFactory"},
		{"empty timeframe", func(c *sweep2d.Config2D) { c.Timeframe = "" }, "Timeframe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := valid
			tt.modify(&cfg)
			_, err := sweep2d.Run(context.Background(), cfg, p)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestRun_FactoryError verifies that a factory error surfaces with parameter context.
func TestRun_FactoryError(t *testing.T) {
	p := &testutil.StaticProvider{Candles: testutil.MakeAlternatingCandles(10, 120, 80)}
	cfg := baseConfig()
	cfg.StrategyFactory = func(p1, p2 float64) (strategy.Strategy, error) {
		return nil, fmt.Errorf("injected error")
	}
	_, err := sweep2d.Run(context.Background(), cfg, p)
	if err == nil {
		t.Fatal("expected error from factory, got nil")
	}
	if !strings.Contains(err.Error(), "factory") {
		t.Errorf("error %q does not mention factory context", err.Error())
	}
}

// TestWriteCSV verifies that WriteCSV produces a file with the correct structure.
func TestWriteCSV(t *testing.T) {
	t.Parallel()
	candles := testutil.MakeAlternatingCandles(300, 120, 80)
	p := &testutil.StaticProvider{Candles: candles}

	report, err := sweep2d.Run(context.Background(), baseConfig(), p)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sweep2d.csv")
	if err := sweep2d.WriteCSV(path, report); err != nil {
		t.Fatalf("WriteCSV: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")

	// Header + 2 data rows + 1 comment line = 4 lines minimum.
	if len(lines) < 3 {
		t.Fatalf("CSV has %d lines, want >= 3", len(lines))
	}
	// Header must contain both parameter names.
	if !strings.Contains(lines[0], report.Param1Name) {
		t.Errorf("header %q missing Param1Name %q", lines[0], report.Param1Name)
	}
	if !strings.Contains(lines[0], report.Param2Name) {
		t.Errorf("header %q missing Param2Name %q", lines[0], report.Param2Name)
	}
	// Each data row must start with a param1 value.
	for i, v1 := range report.Param1Values {
		wantPrefix := fmt.Sprintf("%g,", v1)
		if !strings.HasPrefix(lines[i+1], wantPrefix) {
			t.Errorf("row %d: got %q, want prefix %q", i+1, lines[i+1], wantPrefix)
		}
	}
}
