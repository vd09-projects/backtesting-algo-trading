// Package sweep provides a single-parameter sweep over a strategy factory,
// ranking each run by Sharpe ratio and identifying the "plateau" — the
// parameter range where Sharpe stays within 80% of the peak.
package sweep

import (
	"context"
	"fmt"
	"math"
	"sort"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// plateauThreshold is the fraction of peak Sharpe that a parameter value must
// achieve to be considered part of the plateau.
const plateauThreshold = 0.80

// Config defines a single-parameter sweep over a strategy factory.
type Config struct {
	// ParameterName is a human-readable label for the sweep parameter (e.g., "period", "threshold").
	ParameterName string
	// Min, Max, Step define the closed interval [Min, Max] stepped by Step.
	Min, Max, Step float64
	// Timeframe is the bar timeframe used to annualize the Sharpe ratio.
	Timeframe model.Timeframe
	// EngineConfig is the fixed engine configuration applied to every run in the sweep.
	// Instrument, From, To, InitialCash, and OrderConfig are held constant across runs.
	EngineConfig engine.Config
	// StrategyFactory creates a strategy for a given parameter value.
	// The sweep calls it once per step; the caller maps the float64 to the
	// strategy's own configuration (e.g., RSI period, SMA lookback).
	StrategyFactory func(float64) (strategy.Strategy, error)
}

// Result holds the outcome for a single parameter value.
type Result struct {
	ParamValue  float64
	SharpeRatio float64
	TotalPnL    float64
	TradeCount  int
	MaxDrawdown float64
}

// PlateauRange describes the parameter range where Sharpe stays within
// plateauThreshold of the peak Sharpe.
type PlateauRange struct {
	MinParam  float64 // smallest qualifying parameter value
	MaxParam  float64 // largest qualifying parameter value
	Count     int     // number of qualifying parameter values
	MinSharpe float64 // lowest Sharpe among qualifying values
}

// Report is the complete output of a parameter sweep.
type Report struct {
	ParameterName string
	Results       []Result      // sorted descending by SharpeRatio
	Plateau       *PlateauRange // nil if no parameter value produced a positive Sharpe
}

// Run executes a parameter sweep over cfg.StrategyFactory for parameter values
// in [cfg.Min, cfg.Max] stepped by cfg.Step. Each run uses a fresh engine with
// cfg.EngineConfig. Results are returned sorted descending by Sharpe ratio.
func Run(ctx context.Context, cfg Config, p provider.DataProvider) (Report, error) { //nolint:gocritic // Config is a caller-constructed value type; pointer would leak internals
	if err := validateConfig(cfg); err != nil {
		return Report{}, err
	}

	steps := paramSteps(cfg.Min, cfg.Max, cfg.Step)
	results := make([]Result, 0, len(steps))

	for _, v := range steps {
		s, err := cfg.StrategyFactory(v)
		if err != nil {
			return Report{}, fmt.Errorf("sweep: factory error at parameter %g: %w", v, err)
		}

		eng := engine.New(cfg.EngineConfig)
		if err := eng.Run(ctx, p, s); err != nil {
			return Report{}, fmt.Errorf("sweep: engine run failed at parameter %g: %w", v, err)
		}

		closed := eng.Portfolio().ClosedTrades()
		curve := eng.Portfolio().EquityCurve()
		rep := analytics.Compute(closed, curve, cfg.Timeframe)

		results = append(results, Result{
			ParamValue:  v,
			SharpeRatio: rep.SharpeRatio,
			TotalPnL:    rep.TotalPnL,
			TradeCount:  rep.TradeCount,
			MaxDrawdown: rep.MaxDrawdown,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].SharpeRatio > results[j].SharpeRatio
	})

	return Report{
		ParameterName: cfg.ParameterName,
		Results:       results,
		Plateau:       computePlateau(results),
	}, nil
}

// validateConfig returns an error if cfg is not a valid sweep configuration.
func validateConfig(cfg Config) error { //nolint:gocritic // value semantics intentional; cfg is read-only in this function
	if cfg.ParameterName == "" {
		return fmt.Errorf("sweep: ParameterName must not be empty")
	}
	if cfg.Step <= 0 {
		return fmt.Errorf("sweep: Step must be positive, got %g", cfg.Step)
	}
	if cfg.Max < cfg.Min {
		return fmt.Errorf("sweep: Max (%g) must be >= Min (%g)", cfg.Max, cfg.Min)
	}
	if cfg.StrategyFactory == nil {
		return fmt.Errorf("sweep: StrategyFactory must not be nil")
	}
	if cfg.Timeframe == "" {
		return fmt.Errorf("sweep: Timeframe must not be empty")
	}
	return nil
}

// paramSteps returns the sequence of parameter values in [lo, hi] stepped by step.
// Integer step counting avoids floating-point accumulation drift.
func paramSteps(lo, hi, step float64) []float64 {
	n := int(math.Round((hi-lo)/step)) + 1
	values := make([]float64, n)
	for i := range values {
		values[i] = lo + float64(i)*step
	}
	return values
}

// computePlateau identifies the parameter range where Sharpe stays within
// plateauThreshold of the peak. results must be sorted descending by Sharpe.
// Returns nil if results is empty or the peak Sharpe is non-positive.
func computePlateau(results []Result) *PlateauRange {
	if len(results) == 0 {
		return nil
	}

	peak := results[0].SharpeRatio
	if peak <= 0 {
		return nil
	}

	floor := plateauThreshold * peak

	var minParam, maxParam, minSharpe float64
	count := 0

	for _, r := range results {
		if r.SharpeRatio < floor {
			continue
		}
		if count == 0 {
			minParam = r.ParamValue
			maxParam = r.ParamValue
			minSharpe = r.SharpeRatio
		} else {
			if r.ParamValue < minParam {
				minParam = r.ParamValue
			}
			if r.ParamValue > maxParam {
				maxParam = r.ParamValue
			}
			if r.SharpeRatio < minSharpe {
				minSharpe = r.SharpeRatio
			}
		}
		count++
	}

	if count == 0 {
		return nil
	}

	return &PlateauRange{
		MinParam:  minParam,
		MaxParam:  maxParam,
		Count:     count,
		MinSharpe: minSharpe,
	}
}
