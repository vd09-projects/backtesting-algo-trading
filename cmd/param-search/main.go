// cmd/param-search runs an N-dimensional parameter grid search for a strategy
// across a universe of instruments and ranks results by DSR-corrected Sharpe ratio.
//
// # Usage
//
//	go run ./cmd/param-search \
//	    --strategy macd-crossover \
//	    --param-grid param-grids/macd-grid.json \
//	    --universe universes/nifty50-large-cap.yaml \
//	    --train-from 2018-01-01 \
//	    --train-to   2022-12-31 \
//	    --out-dir    results/param-search/ \
//	    --top-n      10
//
// # Critical constraint
//
// The grid search runs exclusively on the [--train-from, --train-to] window.
// There is no --oos-from or --oos-to flag — this is architectural enforcement, not
// a documentation warning. Out-of-sample evaluation must be done separately via
// cmd/evaluate after selecting the winning parameters from this tool's output.
//
// # Param-grid JSON schema
//
// See cmd/param-search/README.md for the full schema with examples.
// Brief form:
//
//	{
//	  "axes": [
//	    {"name": "fast-period", "min": 5,  "max": 30, "step": 5},
//	    {"name": "slow-period", "min": 20, "max": 60, "step": 10}
//	  ]
//	}
//
// Each axis corresponds to a strategy parameter name as registered in
// internal/cmdutil/strategies.go. Unrecognized axis names are silently ignored
// by strategy constructors (consistent with --params key=value convention).
//
// # DSR correction
//
// All variants are ranked by mean(DSR per sufficient instrument), where nTrials =
// total grid size (Cartesian product of all axes). DSR = Deflated Sharpe Ratio
// (Bailey & López de Prado, 2014). Instruments with insufficient trade or curve
// metrics are excluded from each variant's aggregate.
//
// # Output
//
// --out-dir/param-search-results.csv with columns:
//
//	params             — JSON object: {"fast-period":17,"slow-period":26}
//	dsr_sharpe         — DSR-corrected mean Sharpe across sufficient instruments
//	raw_sharpe         — raw mean Sharpe across sufficient instruments
//	trade_count        — total trades across all sufficient instruments
//	insufficient_data  — true when all instruments had insufficient data for this variant
//
// Rows are sorted descending by dsr_sharpe; insufficient_data=true rows are last.
//
// # Credentials
//
// Identical to cmd/backtest: KITE_API_KEY and KITE_ACCESS_TOKEN (or interactive flow).
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/paramsearch"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

func main() {
	factory := func(ctx context.Context) (provider.DataProvider, error) {
		cmdutil.LoadDotEnv(".env")
		return cmdutil.BuildProvider(ctx)
	}
	if err := run(os.Args[1:], os.Stdout, os.Stderr, factory); err != nil {
		cmdutil.Fatalf("%v", err)
	}
}

// paramSearchFlags holds the parsed CLI flags.
type paramSearchFlags struct {
	stratName     string
	paramGridFile string
	universeFile  string
	trainFromStr  string
	trainToStr    string
	tfStr         string
	outDir        string
	commissionStr string
	topN          int
	cash          float64
	positionSize  float64
	slippage      float64
}

// gridSpecJSON is the JSON schema for --param-grid files.
// See cmd/param-search/README.md for full documentation.
type gridSpecJSON struct {
	Axes []gridAxisJSON `json:"axes"`
}

type gridAxisJSON struct {
	Name string  `json:"name"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Step float64 `json:"step"`
}

// paramSearchPipeline holds validated, parsed state ready for the search run.
type paramSearchPipeline struct {
	flags       paramSearchFlags
	trainFrom   time.Time
	trainTo     time.Time
	tf          model.Timeframe
	commModel   model.CommissionModel
	gridSpec    paramsearch.GridSpec
	instruments []string
	provider    provider.DataProvider
	ctx         context.Context //nolint:containedctx // pipeline carries its context; not a request handler
}

// run is the testable entry point for cmd/param-search.
//
// **Decision (run() extraction for cmd/param-search) — convention: experimental**
// scope: cmd/param-search
// tags: testability, run-function, flag-newFlagSet, providerFactory
// owner: priya
//
// All logic in run(); main() is a one-liner. Follows cmd/evaluate, cmd/walk-forward,
// cmd/monitor, cmd/fetch-history precedents. providerFactory injection enables tests
// to substitute a mock provider without live credentials.
func run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error {
	fs := flag.NewFlagSet("param-search", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var f paramSearchFlags
	registerFlagsInto(fs, &f)

	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := validateFlags(f); err != nil {
		return err
	}

	pl, err := buildSearchPipeline(f, providerFactory)
	if err != nil {
		return err
	}
	return runSearchPipeline(pl, stdout, stderr)
}

// buildSearchPipeline parses dates, validates flags, loads grid and universe,
// and constructs the provider.
func buildSearchPipeline(f paramSearchFlags, providerFactory func(context.Context) (provider.DataProvider, error)) (paramSearchPipeline, error) { //nolint:gocritic // value semantics for one-shot pipeline construction
	trainFrom, trainTo, err := cmdutil.ParseDateRange(f.trainFromStr, f.trainToStr)
	if err != nil {
		return paramSearchPipeline{}, err
	}

	tf, err := cmdutil.ParseTimeframe(f.tfStr)
	if err != nil {
		return paramSearchPipeline{}, err
	}

	commModel, err := cmdutil.ParseCommissionModel(f.commissionStr)
	if err != nil {
		return paramSearchPipeline{}, fmt.Errorf("--commission: %w", err)
	}

	// Validate strategy name — MustGet panics on unknown names (fail-fast at startup).
	cmdutil.GlobalRegistry.MustGet(f.stratName)

	// Validate --out-dir exists before any engine runs.
	if _, err := os.Stat(f.outDir); err != nil {
		return paramSearchPipeline{}, fmt.Errorf("--out-dir %q: %w", f.outDir, err)
	}

	gridSpec, err := loadGridSpec(f.paramGridFile)
	if err != nil {
		return paramSearchPipeline{}, fmt.Errorf("--param-grid: %w", err)
	}

	instruments, err := universesweep.ParseUniverseFile(f.universeFile)
	if err != nil {
		return paramSearchPipeline{}, fmt.Errorf("--universe: %w", err)
	}

	ctx := context.Background()
	p, err := providerFactory(ctx)
	if err != nil {
		return paramSearchPipeline{}, fmt.Errorf("provider: %w", err)
	}

	return paramSearchPipeline{
		flags:       f,
		trainFrom:   trainFrom,
		trainTo:     trainTo,
		tf:          tf,
		commModel:   commModel,
		gridSpec:    gridSpec,
		instruments: instruments,
		provider:    p,
		ctx:         ctx,
	}, nil
}

// runSearchPipeline executes the parameter search and writes CSV output.
func runSearchPipeline(pl paramSearchPipeline, stdout, stderr io.Writer) error { //nolint:gocritic // paramSearchPipeline is a pipeline struct; value semantics acceptable
	f := pl.flags
	engCfg := engine.Config{
		From:                 pl.trainFrom,
		To:                   pl.trainTo,
		InitialCash:          f.cash,
		PositionSizeFraction: f.positionSize,
		OrderConfig: model.OrderConfig{
			SlippagePct:     f.slippage,
			CommissionModel: pl.commModel,
		},
	}

	stratName := f.stratName
	tf := pl.tf
	cfg := paramsearch.Config{
		GridSpec:     pl.gridSpec,
		Instruments:  pl.instruments,
		EngineConfig: engCfg,
		Timeframe:    tf,
		StrategyFactory: func(params map[string]float64) (strategy.Strategy, error) {
			// Merge grid params onto strategy defaults.
			base := cmdutil.GlobalRegistry.DefaultParams(stratName)
			for k, v := range params {
				base[k] = v
			}
			return cmdutil.GlobalRegistry.Build(stratName, tf, base)
		},
	}

	fmt.Fprintf(stderr, "param-search: strategy=%s universe=%d instruments grid=%v train=[%s, %s] timeframe=%s\n", //nolint:errcheck // progress; non-fatal write
		f.stratName, len(pl.instruments), pl.gridSpec.Axes, f.trainFromStr, f.trainToStr, f.tfStr)

	report, err := paramsearch.Run(pl.ctx, cfg, pl.provider)
	if err != nil {
		return fmt.Errorf("param-search: %w", err)
	}

	fmt.Fprintf(stderr, "param-search: completed %d variants (gridSize=%d)\n", len(report.Results), report.GridSize) //nolint:errcheck // progress; non-fatal write

	// Write CSV output.
	outPath := filepath.Join(f.outDir, "param-search-results.csv")
	outFile, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("create results file: %w", err)
	}
	defer outFile.Close() //nolint:errcheck // read-only after write; close error non-fatal

	if err := writeResultsCSV(outFile, report, f.topN); err != nil {
		return fmt.Errorf("write results CSV: %w", err)
	}

	fmt.Fprintf(stdout, "Results written to %s\n", outPath) //nolint:errcheck // final output; non-fatal write
	return nil
}

// registerFlags registers all param-search flags on fs.
// Exported for use in TestRun_NoOOSFlag to inspect the FlagSet.
func registerFlags(fs *flag.FlagSet) {
	var f paramSearchFlags
	registerFlagsInto(fs, &f)
}

// registerFlagsInto binds all flags to the provided paramSearchFlags struct.
func registerFlagsInto(fs *flag.FlagSet, f *paramSearchFlags) {
	fs.StringVar(&f.stratName, "strategy", "", "Strategy name (required): "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), ", "))
	fs.StringVar(&f.paramGridFile, "param-grid", "", "Path to param-grid JSON file defining axes and ranges (required; see README.md for schema)")
	fs.StringVar(&f.universeFile, "universe", "", "Path to universe YAML file (required)")
	fs.StringVar(&f.trainFromStr, "train-from", "", "Training window start date YYYY-MM-DD, inclusive (required)")
	fs.StringVar(&f.trainToStr, "train-to", "", "Training window end date YYYY-MM-DD, exclusive (required)")
	fs.StringVar(&f.tfStr, "timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	fs.StringVar(&f.outDir, "out-dir", "", "Output directory for param-search-results.csv (required; must exist)")
	fs.StringVar(&f.commissionStr, "commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")
	fs.IntVar(&f.topN, "top-n", 10, "Number of top-ranked variants to write to CSV (0 = all)")
	fs.Float64Var(&f.cash, "cash", 100_000, "Starting cash in Rs per engine run")
	fs.Float64Var(&f.positionSize, "position-size", 0.10, "Fraction of cash deployed per trade")
	fs.Float64Var(&f.slippage, "slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")
	// NOTE: There are intentionally no --oos-from or --oos-to flags.
	// Per Marcus standing order (2026-05-04): parameter search runs on training window only;
	// OOS evaluation is the caller's responsibility via cmd/evaluate.
	// The absence of these flags is architectural enforcement, not convention.
}

// validateFlags returns an error for any missing required flag.
func validateFlags(f paramSearchFlags) error { //nolint:gocritic // value semantics for one-shot validation
	switch {
	case f.stratName == "":
		return fmt.Errorf("--strategy is required (%s)", strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
	case f.paramGridFile == "":
		return fmt.Errorf("--param-grid is required (path to JSON grid spec; see cmd/param-search/README.md)")
	case f.universeFile == "":
		return fmt.Errorf("--universe is required (path to universe YAML file)")
	case f.trainFromStr == "":
		return fmt.Errorf("--train-from is required (e.g. 2020-01-01)")
	case f.trainToStr == "":
		return fmt.Errorf("--train-to is required (e.g. 2022-12-31)")
	case f.outDir == "":
		return fmt.Errorf("--out-dir is required (must be an existing directory)")
	}
	return nil
}

// loadGridSpec reads and validates a param-grid JSON file.
// Returns an error for malformed JSON or invalid axis definitions.
func loadGridSpec(path string) (paramsearch.GridSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return paramsearch.GridSpec{}, fmt.Errorf("read %q: %w", path, err)
	}

	var raw gridSpecJSON
	if err := json.Unmarshal(data, &raw); err != nil {
		return paramsearch.GridSpec{}, fmt.Errorf("parse %q: %w", path, err)
	}

	if len(raw.Axes) == 0 {
		return paramsearch.GridSpec{}, fmt.Errorf("%q: axes must not be empty", path)
	}

	axes := make([]paramsearch.GridAxis, len(raw.Axes))
	for i, a := range raw.Axes {
		if a.Name == "" {
			return paramsearch.GridSpec{}, fmt.Errorf("%q: axis[%d].name must not be empty", path, i)
		}
		if a.Step <= 0 {
			return paramsearch.GridSpec{}, fmt.Errorf("%q: axis[%d] (%q): step must be > 0, got %g", path, i, a.Name, a.Step)
		}
		if a.Max < a.Min {
			return paramsearch.GridSpec{}, fmt.Errorf("%q: axis[%d] (%q): max (%g) must be >= min (%g)", path, i, a.Name, a.Max, a.Min)
		}
		axes[i] = paramsearch.GridAxis{Name: a.Name, Min: a.Min, Max: a.Max, Step: a.Step}
	}

	return paramsearch.GridSpec{Axes: axes}, nil
}

// writeResultsCSV writes the param-search report to w as CSV.
// topN limits the number of rows written (0 = all).
//
// **Decision (params column as JSON in CSV output) — convention: experimental**
// scope: cmd/param-search
// tags: CSV, params, JSON, serialization
// owner: priya
//
// The params map for each variant is serialized as a compact JSON object in a single
// 'params' CSV column. A fixed column-per-axis schema would require different headers
// for each grid, breaking downstream tooling. JSON params is stable across any
// grid dimensionality; pandas can parse it with df['params'].apply(json.loads).
//
// **Decision (--out-dir existence validated at startup) — convention: experimental**
// scope: cmd/param-search
// tags: validation, out-dir, startup
// owner: priya
//
// --out-dir is checked with os.Stat before any engine runs. A missing directory on
// a 200-variant run would otherwise complete all expensive engine work before failing
// on CSV write.
func writeResultsCSV(w io.Writer, report paramsearch.Report, topN int) error {
	cw := csv.NewWriter(w)
	header := []string{"params", "dsr_sharpe", "raw_sharpe", "trade_count", "insufficient_data"}
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

	limit := len(report.Results)
	if topN > 0 && topN < limit {
		limit = topN
	}

	for _, res := range report.Results[:limit] {
		paramsJSON, err := json.Marshal(res.Params)
		if err != nil {
			return fmt.Errorf("marshal params: %w", err)
		}
		row := []string{
			string(paramsJSON),
			strconv.FormatFloat(res.DSRSharpe, 'f', 6, 64),
			strconv.FormatFloat(res.RawSharpe, 'f', 6, 64),
			strconv.Itoa(res.TradeCount),
			strconv.FormatBool(res.InsufficientData),
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write CSV row: %w", err)
		}
	}

	cw.Flush()
	return cw.Error()
}
