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
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/walkforward"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

// buildProductionProvider constructs the cached Zerodha provider used in production.
func buildProductionProvider(_ context.Context) (provider.DataProvider, error) {
	cmdutil.LoadDotEnv(".env")
	return cmdutil.BuildProvider(context.Background())
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider); err != nil {
		var ee *cmdutil.ExitCodeError
		if errors.As(err, &ee) {
			os.Exit(ee.Code)
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
// run is the testable entry point.
//
// **Decision (add providerFactory injection to cmd/walk-forward to enable ErrIncompleteData unit tests) — convention: experimental**
// scope: cmd/walk-forward
// tags: testability, providerFactory, ErrIncompleteData, TASK-0083
// owner: priya
//
// Previously run() called cmdutil.BuildProvider directly, making the ErrIncompleteData
// path untestable. providerFactory injection follows cmd/universe-sweep and
// cmd/fetch-history precedents. main() passes buildProductionProvider; tests inject mocks.
func run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error {
	fs := flag.NewFlagSet("walk-forward", flag.ContinueOnError)
	fs.SetOutput(stderr)

	// Core flags.
	instrument := fs.String("instrument", "", "Instrument identifier, e.g. \"NSE:TCS\" (required)")
	fromStr := fs.String("from", "", "Outer window start date YYYY-MM-DD (inclusive, required)")
	toStr := fs.String("to", "", "Outer window end date YYYY-MM-DD (exclusive upper bound, required) — e.g. 2025-01-01 covers data through 2024-12-31")
	stratName := fs.String("strategy", "", "Strategy name: "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | ")+" (required)")

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

	// Strategy-specific parameters — registered centrally from GlobalRegistry so
	// adding a new strategy only requires a change to internal/cmdutil/strategies.go.
	stratParamPtrs := cmdutil.GlobalRegistry.RegisterFlags(fs)

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

	// Validate strategy name early: MustGet panics with a descriptive message if
	// the name is unknown — fail-fast at startup before any provider I/O.
	cmdutil.GlobalRegistry.MustGet(*stratName)

	stratParams := cmdutil.BuildParamMap(stratParamPtrs)

	factory, err := cmdutil.GlobalRegistry.WalkForwardFactory(*stratName, model.TimeframeDaily, stratParams)
	if err != nil {
		return fmt.Errorf("--strategy: %w", err)
	}

	ctx := context.Background()

	p, err := providerFactory(ctx)
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
		if ee := cmdutil.HandleIncompleteDataError(err, stderr); ee != nil {
			return ee
		}
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
		return &cmdutil.ExitCodeError{Code: 1}
	}
	return nil
}

// parseAndValidateFlags validates the four required flags and parses dates.
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
		return time.Time{}, time.Time{}, fmt.Errorf("--strategy is required: %s",
			strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
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
