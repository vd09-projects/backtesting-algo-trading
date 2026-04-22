// Package universesweep runs a fixed strategy across a list of instruments
// (a "universe") and produces a CSV report ranked by Sharpe ratio.
//
// # Concurrency model
//
// One engine run per instrument, fanned out via errgroup (GOMAXPROCS ceiling).
// Results are written into a pre-allocated slice at fixed indices so output
// order matches input order before the final Sharpe sort — identical to the
// walk-forward harness pattern.
//
// # Signal frequency gate
//
// Rather than adding new gate logic, universesweep reads
// analytics.Report.TradeMetricsInsufficient and CurveMetricsInsufficient
// directly. Either flag set means the result is marked InsufficientData=true
// in the CSV output. No thresholds are re-implemented here.
package universesweep

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Config holds universe-sweep run parameters.
type Config struct {
	Instruments  []string          // validated non-empty list from ParseUniverseFile
	Strategy     strategy.Strategy // fixed strategy instance, used for all instruments
	EngineConfig engine.Config     // template; Instrument field is overwritten per run
	Timeframe    model.Timeframe   // used for analytics.Compute annualization
}

// Result holds per-instrument sweep output.
type Result struct {
	Instrument       string
	Sharpe           float64
	TradeCount       int
	TotalPnL         float64
	MaxDrawdown      float64
	InsufficientData bool // true if TradeMetricsInsufficient || CurveMetricsInsufficient
}

// Report is the aggregate output sorted descending by Sharpe.
type Report struct {
	Results []Result
}

// universeFile is the YAML schema for universe files.
type universeFile struct {
	Instruments []string `yaml:"instruments"`
}

// ParseUniverseFile decodes a YAML file with a top-level `instruments:` key.
// It deduplicates entries and returns an error if the resulting list is empty
// or the file cannot be read.
//
// **Decision (universe file format — YAML with instruments: key) — architecture: experimental**
// scope: internal/universesweep, universes/
// tags: universe, YAML, file-format
// owner: priya
//
// Plain-text line-per-instrument would work, but YAML with a named key leaves
// room to add metadata (exchange, asset class, lot size) per entry later without
// breaking existing files. The `instruments:` key mirrors the terminology used
// throughout the codebase. gopkg.in/yaml.v3 is already in go.mod.
func ParseUniverseFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("universesweep: open universe file %q: %w", path, err)
	}
	defer f.Close() //nolint:errcheck // read-only file; close error is non-fatal

	var uf universeFile
	if err := yaml.NewDecoder(f).Decode(&uf); err != nil {
		return nil, fmt.Errorf("universesweep: decode universe file %q: %w", path, err)
	}

	if len(uf.Instruments) == 0 {
		return nil, fmt.Errorf("universesweep: universe file %q has no instruments", path)
	}

	// Deduplicate while preserving order.
	seen := make(map[string]struct{}, len(uf.Instruments))
	deduped := make([]string, 0, len(uf.Instruments))
	for _, inst := range uf.Instruments {
		if _, ok := seen[inst]; ok {
			continue
		}
		seen[inst] = struct{}{}
		deduped = append(deduped, inst)
	}

	return deduped, nil
}

// Run fans out one engine run per instrument via errgroup (GOMAXPROCS ceiling).
// Results are written into a pre-allocated slice at fixed indices so that the
// pre-sort ordering is deterministic regardless of goroutine scheduling.
// The returned Report has Results sorted descending by Sharpe ratio.
//
// **Decision (errgroup for universe fan-out, GOMAXPROCS ceiling) — tradeoff: experimental**
// scope: internal/universesweep
// tags: concurrency, errgroup, parallelism, GOMAXPROCS
// owner: priya
//
// Unlike the single-instrument parameter sweep (internal/sweep), a universe
// sweep fans out across independent instruments — each run has no ordering
// dependency on any other. errgroup with a GOMAXPROCS ceiling is exactly the
// "parallel across runs" pattern from go-patterns.md. golang.org/x/sync is
// already in go.mod. The ceiling avoids spawning N goroutines for a 500-stock
// universe on a 4-core machine; each goroutine holds a full candle series.
func Run(ctx context.Context, cfg *Config, p provider.DataProvider) (Report, error) {
	if len(cfg.Instruments) == 0 {
		return Report{}, fmt.Errorf("universesweep: instruments list must not be empty")
	}

	results := make([]Result, len(cfg.Instruments))

	// **Decision (pre-allocated fixed-index writes for determinism) — convention: experimental**
	// scope: internal/universesweep
	// tags: determinism, goroutine, slice
	// owner: priya
	//
	// Each goroutine writes to results[i] at a fixed index determined before launch.
	// No mutex needed — each goroutine owns its own index. Output order after Sharpe
	// sort is therefore deterministic: same instruments → same sort → same CSV.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for i := range cfg.Instruments {
		i := i // capture loop variable
		g.Go(func() error {
			result, err := runInstrument(gctx, cfg, p, cfg.Instruments[i])
			if err != nil {
				return fmt.Errorf("universesweep: instrument %q: %w", cfg.Instruments[i], err)
			}
			results[i] = result
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Report{}, err
	}

	// Sort descending by Sharpe.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Sharpe > results[j].Sharpe
	})

	return Report{Results: results}, nil
}

// runInstrument executes a single engine run for the given instrument and
// returns the corresponding Result.
func runInstrument(ctx context.Context, cfg *Config, p provider.DataProvider, instrument string) (Result, error) {
	engCfg := cfg.EngineConfig
	engCfg.Instrument = instrument

	eng := engine.New(engCfg)
	if err := eng.Run(ctx, p, cfg.Strategy); err != nil {
		return Result{}, fmt.Errorf("engine run: %w", err)
	}

	port := eng.Portfolio()
	curve := port.EquityCurve()
	trades := port.ClosedTrades()
	report := analytics.Compute(trades, curve, cfg.Timeframe)

	return Result{
		Instrument:       instrument,
		Sharpe:           report.SharpeRatio,
		TradeCount:       report.TradeCount,
		TotalPnL:         report.TotalPnL,
		MaxDrawdown:      report.MaxDrawdown,
		InsufficientData: report.TradeMetricsInsufficient || report.CurveMetricsInsufficient,
	}, nil
}

// WriteCSV writes a header row followed by one row per result to w.
// The insufficient_data column is written as "true" or "false".
// Column order: instrument, sharpe, trade_count, total_pnl, max_drawdown, insufficient_data.
func WriteCSV(w io.Writer, r Report) error {
	cw := csv.NewWriter(w)

	if err := cw.Write([]string{
		"instrument", "sharpe", "trade_count", "total_pnl", "max_drawdown", "insufficient_data",
	}); err != nil {
		return fmt.Errorf("universesweep: write CSV header: %w", err)
	}

	for _, res := range r.Results {
		row := []string{
			res.Instrument,
			strconv.FormatFloat(res.Sharpe, 'f', 6, 64),
			strconv.Itoa(res.TradeCount),
			strconv.FormatFloat(res.TotalPnL, 'f', 2, 64),
			strconv.FormatFloat(res.MaxDrawdown, 'f', 4, 64),
			strconv.FormatBool(res.InsufficientData),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("universesweep: write CSV row for %q: %w", res.Instrument, err)
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("universesweep: flush CSV: %w", err)
	}
	return nil
}
