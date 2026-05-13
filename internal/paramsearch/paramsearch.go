// Package paramsearch runs an N-dimensional parameter grid search over a strategy
// factory, ranking each parameter combination by DSR-corrected average Sharpe ratio.
//
// # Methodology
//
// Grid search runs exclusively on the caller-supplied date window — no out-of-sample
// data is ever accessed. The DSR correction uses nTrials = total grid size (the
// number of parameter combinations tested), matching the multiple-testing penalty
// established in Bailey & López de Prado (2014).
//
// Per-variant DSR is computed as mean(analytics.DSR(instrumentSharpe, gridSize,
// instrumentTradeCount)) across sufficient instruments, matching the methodology
// in internal/universesweep.ApplyUniverseGate exactly. Instruments with
// TradeMetricsInsufficient || CurveMetricsInsufficient are excluded from the
// per-variant aggregate; a variant where all instruments are insufficient is flagged
// InsufficientData=true.
//
// # Concurrency model
//
// Outer loop: variants run sequentially. This avoids gridSize×instruments concurrent
// goroutines and keeps memory bounded. Inner loop: per-variant instrument runs fan out
// in parallel via errgroup with a GOMAXPROCS ceiling — matching universesweep.Run.
//
// **Decision (param-search outer-sequential inner-parallel) — tradeoff: experimental**
// scope: internal/paramsearch
// tags: concurrency, errgroup, grid-search, GOMAXPROCS
// owner: priya
//
// Outer parallelism would require a 2D pre-allocated result matrix and adds complexity
// without a runtime win: network fetch per instrument is the dominant cost, and the
// inner errgroup already saturates cores. Outer sequencing keeps the code simple and
// memory bounded.
//
// **Decision (internal/paramsearch as new package) — architecture: experimental**
// scope: internal/paramsearch, cmd/param-search
// tags: package-boundary, DSR, grid-search, testability
// owner: priya
//
// Grid iteration, per-variant DSR aggregation, and InsufficientData filtering have
// testable invariants (DSR monotonicity, Cartesian product size = nTrials, ranking
// inversion, insufficient-instrument exclusion) that belong in a package with isolated
// tests. Matches the internal/sweep and internal/sweep2d precedents. Single caller
// today; promote to a shared utility if a second cmd needs grid search.
package paramsearch

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sort"

	"golang.org/x/sync/errgroup"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// GridAxis defines one sweep axis: a closed interval [Min, Max] stepped by Step.
type GridAxis struct {
	Name     string
	Min, Max float64
	Step     float64
}

// GridSpec defines the N-dimensional parameter grid.
// The Cartesian product of all axes produces the full variant list.
type GridSpec struct {
	Axes []GridAxis
}

// VariantResult holds the aggregate outcome for one parameter combination.
//
// **Decision (DSR aggregation matches ApplyUniverseGate) — convention: experimental**
// scope: internal/paramsearch
// tags: DSR, aggregation, methodology, universesweep
// owner: priya
//
// Per-variant DSR = mean(analytics.DSR(instrumentSharpe, gridSize, instrumentTradeCount))
// across sufficient instruments. This matches universesweep.ApplyUniverseGate exactly.
// Alternative (DSR of avgSharpe with totalTradeCount) is rejected: it inflates nObs and
// DSR(avg) ≠ avg(DSR), producing rankings inconsistent with the universe gate methodology.
type VariantResult struct {
	// Params is the parameter combination for this variant.
	Params map[string]float64
	// DSRSharpe is the mean DSR-corrected Sharpe across sufficient instruments.
	// Zero when InsufficientData is true.
	DSRSharpe float64
	// RawSharpe is the mean raw Sharpe across sufficient instruments.
	// Zero when InsufficientData is true.
	RawSharpe float64
	// TradeCount is the total trade count across all sufficient instruments.
	TradeCount int
	// InsufficientData is true when all instruments for this variant had
	// TradeMetricsInsufficient || CurveMetricsInsufficient.
	//
	// **Decision (insufficient instruments excluded per variant) — convention: experimental**
	// scope: internal/paramsearch
	// tags: InsufficientData, aggregation, filtering
	// owner: priya
	//
	// Mirrors universesweep.runInstrument: instruments where TradeMetricsInsufficient ||
	// CurveMetricsInsufficient are excluded from per-variant Sharpe and DSR aggregation.
	// VariantResult.InsufficientData=true when zero sufficient instruments exist — sorted
	// last in output, present in CSV so the caller can see the full grid.
	InsufficientData bool
}

// Report is the complete output of a parameter grid search.
type Report struct {
	// Results is sorted descending by DSRSharpe; insufficient variants sorted last.
	Results []VariantResult
	// GridSize is the total number of parameter combinations tested.
	// It equals the nTrials used in all DSR computations.
	GridSize int
}

// Config defines a parameter grid search run.
type Config struct {
	// GridSpec defines the N-dimensional parameter grid.
	GridSpec GridSpec
	// Instruments is the list of instruments to evaluate each variant against.
	Instruments []string
	// EngineConfig is the fixed engine configuration applied to every run.
	// The Instrument field is overwritten per instrument; From/To define the training window.
	EngineConfig engine.Config
	// Timeframe is the bar timeframe used to annualize the Sharpe ratio.
	Timeframe model.Timeframe
	// StrategyFactory constructs a fresh strategy for the given parameter map.
	// Called once per (variant, instrument) pair.
	StrategyFactory func(params map[string]float64) (strategy.Strategy, error)
}

// Run executes the parameter grid search defined by cfg.
// It iterates variants sequentially; for each variant it fans out instrument
// runs in parallel via errgroup (GOMAXPROCS ceiling).
// Results are sorted descending by DSRSharpe; insufficient variants are sorted last.
func Run(ctx context.Context, cfg Config, p provider.DataProvider) (Report, error) { //nolint:gocritic // Config is a caller-constructed value type; pointer would leak internals
	if err := validateConfig(cfg); err != nil {
		return Report{}, err
	}

	variants := cartesianProduct(cfg.GridSpec.Axes)
	gridSize := len(variants)
	results := make([]VariantResult, gridSize)

	for i, params := range variants {
		result, err := runVariant(ctx, &cfg, p, params, gridSize)
		if err != nil {
			return Report{}, err
		}
		result.Params = params
		results[i] = result
	}

	// Sort: sufficient variants descending by DSRSharpe; insufficient variants last.
	sort.SliceStable(results, func(i, j int) bool {
		ri, rj := results[i], results[j]
		if ri.InsufficientData != rj.InsufficientData {
			return !ri.InsufficientData // sufficient before insufficient
		}
		return ri.DSRSharpe > rj.DSRSharpe
	})

	return Report{Results: results, GridSize: gridSize}, nil
}

// runVariant executes one parameter combination across all instruments in cfg,
// aggregating per-instrument DSR-corrected Sharpe.
func runVariant(ctx context.Context, cfg *Config, p provider.DataProvider, params map[string]float64, gridSize int) (VariantResult, error) {
	instruments := cfg.Instruments
	type instrumentResult struct {
		sharpe           float64
		tradeCount       int
		insufficientData bool
	}
	instResults := make([]instrumentResult, len(instruments))

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for i, inst := range instruments {
		i, inst := i, inst
		g.Go(func() error {
			s, err := cfg.StrategyFactory(params)
			if err != nil {
				return fmt.Errorf("paramsearch: strategy factory error for instrument %q: %w", inst, err)
			}

			engCfg := cfg.EngineConfig
			engCfg.Instrument = inst
			eng := engine.New(engCfg)

			if err := eng.Run(gctx, p, s); err != nil {
				return fmt.Errorf("paramsearch: engine run failed for instrument %q: %w", inst, err)
			}

			port := eng.Portfolio()
			trades := port.ClosedTrades()
			curve := port.EquityCurve()
			rep := analytics.Compute(trades, curve, cfg.Timeframe)

			instResults[i] = instrumentResult{
				sharpe:           rep.SharpeRatio,
				tradeCount:       rep.TradeCount,
				insufficientData: rep.TradeMetricsInsufficient || rep.CurveMetricsInsufficient,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return VariantResult{}, err
	}

	// Aggregate: compute mean(DSR per sufficient instrument).
	var dsrSum, sharpeSum float64
	var totalTrades, sufficientCount int

	for _, ir := range instResults {
		if ir.insufficientData {
			continue
		}
		dsr := analytics.DSR(ir.sharpe, float64(gridSize), float64(ir.tradeCount))
		dsrSum += dsr
		sharpeSum += ir.sharpe
		totalTrades += ir.tradeCount
		sufficientCount++
	}

	if sufficientCount == 0 {
		return VariantResult{InsufficientData: true}, nil
	}

	return VariantResult{
		DSRSharpe:  dsrSum / float64(sufficientCount),
		RawSharpe:  sharpeSum / float64(sufficientCount),
		TradeCount: totalTrades,
	}, nil
}

// validateConfig returns an error if cfg is not a valid parameter search configuration.
func validateConfig(cfg Config) error { //nolint:gocritic // value semantics intentional; cfg is read-only
	if len(cfg.GridSpec.Axes) == 0 {
		return fmt.Errorf("paramsearch: GridSpec must have at least one axis")
	}
	for i, ax := range cfg.GridSpec.Axes {
		if ax.Name == "" {
			return fmt.Errorf("paramsearch: Axis[%d].Name must not be empty", i)
		}
		if ax.Step <= 0 {
			return fmt.Errorf("paramsearch: Axis[%d] (%q): Step must be positive, got %g", i, ax.Name, ax.Step)
		}
		if ax.Max < ax.Min {
			return fmt.Errorf("paramsearch: Axis[%d] (%q): Max (%g) must be >= Min (%g)", i, ax.Name, ax.Max, ax.Min)
		}
	}
	if len(cfg.Instruments) == 0 {
		return fmt.Errorf("paramsearch: Instruments must not be empty")
	}
	if cfg.StrategyFactory == nil {
		return fmt.Errorf("paramsearch: StrategyFactory must not be nil")
	}
	if cfg.Timeframe == "" {
		return fmt.Errorf("paramsearch: Timeframe must not be empty")
	}
	return nil
}

// cartesianProduct returns all parameter combinations from the given axes.
// Each combination is a map from axis name to value. The result is deterministic:
// outer axes vary slowest (first axis is the outermost loop).
func cartesianProduct(axes []GridAxis) []map[string]float64 {
	if len(axes) == 0 {
		return nil
	}

	// Build the value sequence for each axis.
	axisValues := make([][]float64, len(axes))
	for i, ax := range axes {
		axisValues[i] = axisSteps(ax.Min, ax.Max, ax.Step)
	}

	// Compute total number of combinations.
	total := 1
	for _, vals := range axisValues {
		total *= len(vals)
	}

	result := make([]map[string]float64, 0, total)
	indices := make([]int, len(axes))

	for {
		combo := make(map[string]float64, len(axes))
		for i, ax := range axes {
			combo[ax.Name] = axisValues[i][indices[i]]
		}
		result = append(result, combo)

		// Increment indices (last axis varies fastest).
		carry := true
		for i := len(indices) - 1; i >= 0 && carry; i-- {
			indices[i]++
			if indices[i] < len(axisValues[i]) {
				carry = false
			} else {
				indices[i] = 0
			}
		}
		if carry {
			break // all combinations exhausted
		}
	}

	return result
}

// axisSteps returns the sequence of values in [lo, hi] stepped by step.
// Integer step counting avoids floating-point accumulation drift.
func axisSteps(lo, hi, step float64) []float64 {
	n := int(math.Round((hi-lo)/step)) + 1
	values := make([]float64, n)
	for i := range values {
		values[i] = lo + float64(i)*step
	}
	return values
}
