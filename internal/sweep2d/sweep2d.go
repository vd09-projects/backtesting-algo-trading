// Package sweep2d runs a two-parameter grid sweep over a strategy factory,
// producing a [param1 × param2] matrix of backtest results and a
// DSR-corrected peak Sharpe ratio to quantify multiple-testing inflation.
package sweep2d

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Config2D defines a two-parameter grid sweep over a strategy factory.
type Config2D struct {
	// Param1 and Param2 define the two sweep axes.
	Param1, Param2 ParamRange
	// Timeframe is the bar timeframe used to annualize the Sharpe ratio.
	Timeframe model.Timeframe
	// EngineConfig is the fixed engine configuration applied to every run.
	EngineConfig engine.Config
	// StrategyFactory creates a strategy for (param1, param2). Called once per grid cell.
	StrategyFactory func(p1, p2 float64) (strategy.Strategy, error)
}

// ParamRange defines one sweep axis: a closed interval [Min, Max] stepped by Step.
type ParamRange struct {
	Name     string
	Min, Max float64
	Step     float64
}

// GridCell holds results for one (param1, param2) parameter combination.
type GridCell struct {
	Param1Value float64
	Param2Value float64
	SharpeRatio float64
	TradeCount  int
	MaxDrawdown float64
}

// Report2D is the complete output of a two-parameter grid sweep.
// Grid[i][j] corresponds to Param1Values[i] and Param2Values[j].
type Report2D struct {
	Param1Name   string
	Param2Name   string
	Param1Values []float64    // row axis
	Param2Values []float64    // column axis
	Grid         [][]GridCell // [i][j]: i indexes Param1Values, j indexes Param2Values
	VariantCount int
	// PeakSharpe is the highest Sharpe observed across the grid.
	PeakSharpe float64
	// DSRCorrectedPeakSharpe deflates PeakSharpe for the number of trials tested.
	// A positive value means the peak Sharpe exceeds the multiple-testing benchmark.
	DSRCorrectedPeakSharpe float64
}

// Run executes a grid sweep over all (param1, param2) combinations.
// Each run is independent; runs execute in parallel using errgroup with
// GOMAXPROCS concurrency. Results are written to pre-allocated grid indices so
// output order is deterministic regardless of goroutine scheduling.
func Run(ctx context.Context, cfg Config2D, p provider.DataProvider) (Report2D, error) { //nolint:gocritic // Config2D is a caller-constructed value type; pointer would leak internals
	if err := validateConfig(cfg); err != nil {
		return Report2D{}, err
	}

	p1Steps := paramSteps(cfg.Param1.Min, cfg.Param1.Max, cfg.Param1.Step)
	p2Steps := paramSteps(cfg.Param2.Min, cfg.Param2.Max, cfg.Param2.Step)

	// Pre-allocate grid at fixed indices. Each goroutine owns a unique [i][j] cell;
	// no mutex is needed because there are no shared writes.
	grid := make([][]GridCell, len(p1Steps))
	for i := range grid {
		grid[i] = make([]GridCell, len(p2Steps))
	}

	var nObsAtomic int64 // equity curve length; identical across all runs (same date range)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for i, v1 := range p1Steps {
		for j, v2 := range p2Steps {
			i, j, v1, v2 := i, j, v1, v2
			g.Go(func() error {
				s, err := cfg.StrategyFactory(v1, v2)
				if err != nil {
					return fmt.Errorf("sweep2d: factory error at (%g, %g): %w", v1, v2, err)
				}
				eng := engine.New(cfg.EngineConfig)
				if err := eng.Run(gctx, p, s); err != nil {
					return fmt.Errorf("sweep2d: engine run failed at (%g, %g): %w", v1, v2, err)
				}
				closed := eng.Portfolio().ClosedTrades()
				curve := eng.Portfolio().EquityCurve()
				atomic.StoreInt64(&nObsAtomic, int64(len(curve)))
				rep := analytics.Compute(closed, curve, cfg.Timeframe)
				grid[i][j] = GridCell{
					Param1Value: v1,
					Param2Value: v2,
					SharpeRatio: rep.SharpeRatio,
					TradeCount:  rep.TradeCount,
					MaxDrawdown: rep.MaxDrawdown,
				}
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return Report2D{}, err
	}

	variantCount := len(p1Steps) * len(p2Steps)
	nObs := float64(atomic.LoadInt64(&nObsAtomic))

	peakSharpe := peakSharpeFromGrid(grid)
	dsr := analytics.DSR(peakSharpe, float64(variantCount), nObs)

	return Report2D{
		Param1Name:             cfg.Param1.Name,
		Param2Name:             cfg.Param2.Name,
		Param1Values:           p1Steps,
		Param2Values:           p2Steps,
		Grid:                   grid,
		VariantCount:           variantCount,
		PeakSharpe:             peakSharpe,
		DSRCorrectedPeakSharpe: dsr,
	}, nil
}

// validateConfig returns an error if cfg is not a valid 2D sweep configuration.
func validateConfig(cfg Config2D) error { //nolint:gocritic // value semantics intentional; cfg is read-only
	if cfg.Param1.Name == "" {
		return fmt.Errorf("sweep2d: Param1.Name must not be empty")
	}
	if cfg.Param2.Name == "" {
		return fmt.Errorf("sweep2d: Param2.Name must not be empty")
	}
	if err := validateRange("Param1", cfg.Param1); err != nil {
		return err
	}
	if err := validateRange("Param2", cfg.Param2); err != nil {
		return err
	}
	if cfg.StrategyFactory == nil {
		return fmt.Errorf("sweep2d: StrategyFactory must not be nil")
	}
	if cfg.Timeframe == "" {
		return fmt.Errorf("sweep2d: Timeframe must not be empty")
	}
	return nil
}

func validateRange(label string, r ParamRange) error {
	if r.Step <= 0 {
		return fmt.Errorf("sweep2d: %s.Step must be positive, got %g", label, r.Step)
	}
	if r.Max < r.Min {
		return fmt.Errorf("sweep2d: %s.Max (%g) must be >= Min (%g)", label, r.Max, r.Min)
	}
	return nil
}

// paramSteps returns the sequence of values in [lo, hi] stepped by step.
// Integer step counting avoids floating-point accumulation drift.
func paramSteps(lo, hi, step float64) []float64 {
	n := int(math.Round((hi-lo)/step)) + 1
	values := make([]float64, n)
	for i := range values {
		values[i] = lo + float64(i)*step
	}
	return values
}

// peakSharpeFromGrid returns the maximum SharpeRatio across all grid cells.
func peakSharpeFromGrid(grid [][]GridCell) float64 {
	peak := math.Inf(-1)
	for _, row := range grid {
		for _, cell := range row {
			if cell.SharpeRatio > peak {
				peak = cell.SharpeRatio
			}
		}
	}
	if math.IsInf(peak, -1) {
		return 0
	}
	return peak
}
