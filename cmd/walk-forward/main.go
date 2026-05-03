// cmd/walk-forward runs walk-forward validation on a fixed-parameter strategy.
//
// Walk-forward splits historical data into overlapping IS/OOS window pairs (folds),
// runs a full backtest in each window, and checks whether the strategy's OOS Sharpe
// degrades unacceptably versus its IS Sharpe — the primary overfitting signal for
// fixed-parameter daily-bar strategies.
//
// # Usage
//
//	go run ./cmd/walk-forward \
//	    --instrument "NSE:TCS" \
//	    --from 2018-01-01 \
//	    --to   2025-01-01 \
//	    --strategy sma-crossover \
//	    --fast-period 10 --slow-period 50
//
// # Flag defaults
//
//	--is-years   2   (in-sample window length)
//	--oos-years  1   (out-of-sample window length)
//	--step-years 1   (window step size)
//
// These defaults produce 4–5 folds over a 2018–2025 outer window, covering
// pre-COVID, the COVID crash/recovery, the 2022 correction, and 2023. See
// decisions/algorithm/2026-04-22-walk-forward-window-sizing-default.md.
//
// # Output
//
// JSON to stdout: per-fold WindowResults plus aggregate Report (OverfitFlag,
// NegativeFoldFlag, averages). Use --out to additionally write a fold-level CSV.
//
// # Exit codes
//
//	0 — no flags set (strategy passes walk-forward gate)
//	1 — OverfitFlag or NegativeFoldFlag set (strategy fails walk-forward gate)
//
// Exit code 1 enables scripting: a walk-forward runner script can call this binary
// and branch on the exit code.
//
// # --to is exclusive
//
// --to 2025-01-01 covers data through 2024-12-31. Consistent with engine.Config.To
// and provider.FetchCandles([from, to)) semantics throughout this repo.
//
// # Credentials
//
// Read from KITE_API_KEY and KITE_API_SECRET environment variables (or a .env file
// in the working directory). Token handling is identical to cmd/backtest.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/walkforward"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/ccimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/momentum"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var ee *exitCodeError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		cmdutil.Fatalf("%v", err)
	}
}

// run is the testable entry point. It parses args, validates flags, builds the
// strategy factory, connects to the provider, runs walk-forward, and writes
// output. It returns an error rather than calling os.Exit, so tests can invoke
// it directly without spawning a subprocess.
//
// **Decision (run() extraction for testability) — convention: experimental**
// scope: cmd/walk-forward
// tags: testability, flag-parse, coverage, run-function
// owner: priya
//
// main() previously held all wiring. Extracting to run(args, stdout, stderr)
// allows unit tests to cover flag-parse failures, unknown strategy, and invalid
// commission paths without spawning a subprocess or requiring live credentials.
// Tests that reach walkforward.Run still need a live provider — those paths are
// integration-only and are not exercised in unit tests.
func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("walk-forward", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Core flags.
	instrument := fs.String("instrument", "", "Instrument identifier, e.g. \"NSE:TCS\" (required)")
	fromStr := fs.String("from", "", "Outer window start date YYYY-MM-DD (inclusive, required)")
	toStr := fs.String("to", "", "Outer window end date YYYY-MM-DD (exclusive upper bound, required) — e.g. 2025-01-01 covers data through 2024-12-31")
	stratName := fs.String("strategy", "", "Strategy name: sma-crossover | rsi-mean-reversion | donchian-breakout | macd-crossover | bollinger-mean-reversion | momentum | cci-mean-reversion (required)")

	// Walk-forward window flags — defaults per 2026-04-22 decision.
	isYears := fs.Int("is-years", 2, "In-sample window length in years (default 2)")
	oosYears := fs.Int("oos-years", 1, "Out-of-sample window length in years (default 1)")
	stepYears := fs.Int("step-years", 1, "Window step size in years (default 1)")

	// Engine cost flags.
	cash := fs.Float64("cash", 100_000, "Starting cash in ₹ per fold")
	positionSize := fs.Float64("position-size", 0.10, "Fraction of cash deployed per trade")
	slippage := fs.Float64("slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")
	commissionStr := fs.String("commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")

	// Output flags.
	outPath := fs.String("out", "", "Optional path for fold-level CSV output (default: no CSV)")

	// Strategy-specific parameters — same set as cmd/sweep and cmd/universe-sweep.
	fastPeriod := fs.Int("fast-period", 10, "sma-crossover: fast SMA period")
	slowPeriod := fs.Int("slow-period", 50, "sma-crossover: slow SMA period")
	rsiPeriod := fs.Int("rsi-period", 14, "rsi-mean-reversion: RSI period")
	oversold := fs.Float64("oversold", 30, "rsi-mean-reversion: oversold threshold")
	overbought := fs.Float64("overbought", 70, "rsi-mean-reversion: overbought threshold")
	donchianPeriod := fs.Int("donchian-period", 20, "donchian-breakout: channel lookback period")
	macdFastPeriod := fs.Int("macd-fast-period", 12, "macd-crossover: fast EMA period")
	macdSlowPeriod := fs.Int("macd-slow-period", 26, "macd-crossover: slow EMA period")
	macdSignalPeriod := fs.Int("macd-signal-period", 9, "macd-crossover: signal EMA period")
	bbPeriod := fs.Int("bb-period", 20, "bollinger-mean-reversion: Bollinger Band period")
	bbNumStdDev := fs.Float64("bb-num-std-dev", 2.0, "bollinger-mean-reversion: number of standard deviations")
	momentumLookback := fs.Int("momentum-lookback", 231, "momentum: ROC lookback period (default 231 = 252-21, skip-last-month convention)")
	momentumThreshold := fs.Float64("momentum-threshold", 10.0, "momentum: ROC threshold in percent")
	cciPeriod := fs.Int("cci-period", 20, "cci-mean-reversion: CCI period")
	cciEntry := fs.Int("cci-entry", -100, "cci-mean-reversion: entry threshold (buy when CCI < this)")
	cciExit := fs.Int("cci-exit", 0, "cci-mean-reversion: exit threshold (sell when CCI crosses above this)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	from, to, err := parseAndValidateFlags(*instrument, *fromStr, *toStr, *stratName)
	if err != nil {
		return err
	}

	commissionModel, err := cmdutil.ParseCommissionModel(*commissionStr)
	if err != nil {
		return fmt.Errorf("--commission: %w", err)
	}

	params := &strategyParams{
		fastPeriod:        *fastPeriod,
		slowPeriod:        *slowPeriod,
		rsiPeriod:         *rsiPeriod,
		oversold:          *oversold,
		overbought:        *overbought,
		donchianPeriod:    *donchianPeriod,
		macdFastPeriod:    *macdFastPeriod,
		macdSlowPeriod:    *macdSlowPeriod,
		macdSignalPeriod:  *macdSignalPeriod,
		bbPeriod:          *bbPeriod,
		bbNumStdDev:       *bbNumStdDev,
		momentumLookback:  *momentumLookback,
		momentumThreshold: *momentumThreshold,
		cciPeriod:         *cciPeriod,
		cciEntry:          *cciEntry,
		cciExit:           *cciExit,
	}

	factory, err := strategyFactory(*stratName, model.TimeframeDaily, params)
	if err != nil {
		return fmt.Errorf("--strategy: %w", err)
	}

	ctx := context.Background()
	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		return fmt.Errorf("provider: %w", err)
	}

	wfCfg := buildWalkForwardConfig(*instrument, from, to, *isYears, *oosYears, *stepYears)
	baseCfg := walkforward.EngineConfigTemplate{
		InitialCash:          *cash,
		PositionSizeFraction: *positionSize,
		OrderConfig: model.OrderConfig{
			SlippagePct:     *slippage,
			CommissionModel: commissionModel,
		},
	}

	fmt.Fprintf(stderr, "Walk-forward: strategy=%s instrument=%s from=%s to=%s is=%dy oos=%dy step=%dy commission=%s\n", //nolint:errcheck // progress banner to stderr; non-fatal
		*stratName, *instrument,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
		*isYears, *oosYears, *stepYears, *commissionStr,
	)

	report, err := walkforward.Run(ctx, wfCfg, baseCfg, p, factory)
	if err != nil {
		return fmt.Errorf("walk-forward: %w", err)
	}

	// JSON output to stdout.
	if err := writeReportJSON(stdout, report); err != nil {
		return fmt.Errorf("write JSON: %w", err)
	}

	// Optional CSV output.
	if *outPath != "" {
		if err := writeFoldsCSVFile(*outPath, report.Windows); err != nil {
			return fmt.Errorf("write CSV: %w", err)
		}
		fmt.Fprintf(stderr, "Fold CSV written to %s\n", *outPath) //nolint:errcheck // progress banner to stderr; non-fatal
	}

	if determineExitCode(report) != 0 {
		return &exitCodeError{code: 1}
	}
	return nil
}

// exitCodeError is returned by run() when the walk-forward report has flags set.
// main() translates it to os.Exit(1).
type exitCodeError struct{ code int }

func (e *exitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}

// parseAndValidateFlags validates the four required flags and parses dates.
// Returns (from, to time.Time, error).
func parseAndValidateFlags(instrument, fromStr, toStr, stratName string) (from, to time.Time, err error) {
	if instrument == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("--instrument is required (e.g. \"NSE:TCS\")")
	}
	if fromStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("--from is required (e.g. 2018-01-01)")
	}
	if toStr == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("--to is required (e.g. 2025-01-01, exclusive upper bound)")
	}
	if stratName == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("--strategy is required: sma-crossover | rsi-mean-reversion | donchian-breakout | macd-crossover | bollinger-mean-reversion | momentum | cci-mean-reversion")
	}

	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("--from %q: %w", fromStr, err)
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("--to %q: %w", toStr, err)
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, fmt.Errorf("--to (%s) must be strictly after --from (%s)", toStr, fromStr)
	}
	return from, to, nil
}

// strategyParams holds all strategy-specific parameters resolved from CLI flags.
type strategyParams struct {
	fastPeriod        int
	slowPeriod        int
	rsiPeriod         int
	oversold          float64
	overbought        float64
	donchianPeriod    int
	macdFastPeriod    int
	macdSlowPeriod    int
	macdSignalPeriod  int
	bbPeriod          int
	bbNumStdDev       float64
	momentumLookback  int
	momentumThreshold float64
	cciPeriod         int
	cciEntry          int
	cciExit           int
}

// strategyBuilder is a function that validates params eagerly and returns a
// closure that constructs a fresh strategy instance on each call.
type strategyBuilder func(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error)

// strategyRegistry maps strategy names to their builder functions.
// Each builder validates params once at startup and returns a panic-free closure.
//
// **Decision (strategyFactory table dispatch replaces flat switch) — architecture: experimental**
// scope: cmd/walk-forward
// tags: factory, panic-free, cyclop, error-handling
// owner: priya
//
// The flat switch over 7 strategies hit the cyclop complexity limit (max=15).
// A dispatch table reduces strategyFactory complexity to O(1) map lookup while
// keeping each builder self-contained. Builders validate params eagerly — errors
// surface before any fold runs, never inside a goroutine closure.
var strategyRegistry = map[string]strategyBuilder{
	"sma-crossover":            buildSMAFactory,
	"rsi-mean-reversion":       buildRSIFactory,
	"donchian-breakout":        buildDonchianFactory,
	"macd-crossover":           buildMACDFactory,
	"bollinger-mean-reversion": buildBollingerFactory,
	"momentum":                 buildMomentumFactory,
	"cci-mean-reversion":       buildCCIFactory,
}

// strategyFactory looks up the named strategy in strategyRegistry, validates
// params eagerly, and returns a closure that constructs a fresh strategy instance
// on each call. Returns an error if the name is unknown or params are invalid.
func strategyFactory(name string, tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	builder, ok := strategyRegistry[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q; available: sma-crossover, rsi-mean-reversion, donchian-breakout, macd-crossover, bollinger-mean-reversion, momentum, cci-mean-reversion", name)
	}
	return builder(tf, params)
}

func buildSMAFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := smacrossover.New(tf, params.fastPeriod, params.slowPeriod); err != nil {
		return nil, fmt.Errorf("sma-crossover params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := smacrossover.New(tf, params.fastPeriod, params.slowPeriod)
		if err != nil {
			panic(fmt.Sprintf("sma-crossover: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildRSIFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := rsimeanrev.New(tf, params.rsiPeriod, params.oversold, params.overbought); err != nil {
		return nil, fmt.Errorf("rsi-mean-reversion params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := rsimeanrev.New(tf, params.rsiPeriod, params.oversold, params.overbought)
		if err != nil {
			panic(fmt.Sprintf("rsi-mean-reversion: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildDonchianFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := donchian.New(tf, params.donchianPeriod); err != nil {
		return nil, fmt.Errorf("donchian-breakout params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := donchian.New(tf, params.donchianPeriod)
		if err != nil {
			panic(fmt.Sprintf("donchian-breakout: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildMACDFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := macd.New(tf, params.macdFastPeriod, params.macdSlowPeriod, params.macdSignalPeriod); err != nil {
		return nil, fmt.Errorf("macd-crossover params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := macd.New(tf, params.macdFastPeriod, params.macdSlowPeriod, params.macdSignalPeriod)
		if err != nil {
			panic(fmt.Sprintf("macd-crossover: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildBollingerFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := bollinger.New(tf, params.bbPeriod, params.bbNumStdDev); err != nil {
		return nil, fmt.Errorf("bollinger-mean-reversion params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := bollinger.New(tf, params.bbPeriod, params.bbNumStdDev)
		if err != nil {
			panic(fmt.Sprintf("bollinger-mean-reversion: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildMomentumFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := momentum.New(tf, params.momentumLookback, params.momentumThreshold); err != nil {
		return nil, fmt.Errorf("momentum params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := momentum.New(tf, params.momentumLookback, params.momentumThreshold)
		if err != nil {
			panic(fmt.Sprintf("momentum: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

func buildCCIFactory(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error) {
	if _, err := ccimeanrev.New(tf, params.cciPeriod, params.cciEntry, params.cciExit); err != nil {
		return nil, fmt.Errorf("cci-mean-reversion params: %w", err)
	}
	return func() strategy.Strategy {
		s, err := ccimeanrev.New(tf, params.cciPeriod, params.cciEntry, params.cciExit)
		if err != nil {
			panic(fmt.Sprintf("cci-mean-reversion: params validated at startup, unexpected error: %v", err))
		}
		return s
	}, nil
}

// buildWalkForwardConfig constructs a WalkForwardConfig from the outer window
// boundaries and year-based window sizes. Years are converted to durations using
// the 365-day convention consistent with the existing walkforward test suite and
// the 2026-04-22 window sizing decision.
//
// **Decision (year-to-duration: 365 * 24h not calendar year arithmetic) — convention: experimental**
// scope: cmd/walk-forward
// tags: time, year, duration, walk-forward
// owner: priya
//
// time.AddDate(n, 0, 0) would give calendar-exact year boundaries but would cause
// the fold count to depend on how many leap years fall in the window — mildly
// surprising behavior for a flag documented as "in years". The existing
// walkforward_test.go uses 365*24h throughout and documents the leap-year
// arithmetic in comments. Staying consistent with that convention keeps fold
// arithmetic predictable.
func buildWalkForwardConfig(instrument string, from, to time.Time, isYears, oosYears, stepYears int) walkforward.WalkForwardConfig {
	year := 365 * 24 * time.Hour
	return walkforward.WalkForwardConfig{
		Instrument:        instrument,
		From:              from,
		To:                to,
		InSampleWindow:    time.Duration(isYears) * year,
		OutOfSampleWindow: time.Duration(oosYears) * year,
		StepSize:          time.Duration(stepYears) * year,
	}
}

// determineExitCode returns 1 if the report has any flag set, 0 otherwise.
// This enables shell scripting: callers can check $? to branch on walk-forward result.
func determineExitCode(report walkforward.Report) int {
	if report.OverfitFlag || report.NegativeFoldFlag {
		return 1
	}
	return 0
}

// writeReportJSON serializes the full walkforward.Report as indented JSON to w.
func writeReportJSON(w io.Writer, report walkforward.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write report JSON: %w", err)
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write trailing newline: %w", err)
	}
	return nil
}

// writeFoldsCSVFile opens path, writes a fold-level CSV, and closes the file.
// Separate from writeFoldsCSV so that main() does not hold a deferred close
// across an os.Exit call.
func writeFoldsCSVFile(path string, windows []walkforward.WindowResult) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %q: %w", path, err)
	}
	writeErr := writeFoldsCSV(f, windows)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fmt.Errorf("close %q: %w", path, closeErr)
	}
	return nil
}

// writeFoldsCSV writes a fold-level CSV to w. Columns:
//
//	fold_index, is_start, is_end, oos_start, oos_end, is_sharpe, oos_sharpe, trade_count, degenerate
//
// Dates are formatted as YYYY-MM-DD (UTC). Degenerate is "true" or "false".
// An empty windows slice writes only the header row.
func writeFoldsCSV(w io.Writer, windows []walkforward.WindowResult) error {
	var buf bytes.Buffer
	buf.WriteString("fold_index,is_start,is_end,oos_start,oos_end,is_sharpe,oos_sharpe,trade_count,degenerate\n")
	for i := range windows {
		win := &windows[i]
		fmt.Fprintf(&buf, "%d,%s,%s,%s,%s,%.6f,%.6f,%d,%t\n",
			i,
			win.InSampleStart.UTC().Format("2006-01-02"),
			win.InSampleEnd.UTC().Format("2006-01-02"),
			win.OutOfSampleStart.UTC().Format("2006-01-02"),
			win.OutOfSampleEnd.UTC().Format("2006-01-02"),
			win.InSampleSharpe,
			win.OutOfSampleSharpe,
			win.TradeCount,
			win.Degenerate,
		)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writeFoldsCSV: %w", err)
	}
	return nil
}
