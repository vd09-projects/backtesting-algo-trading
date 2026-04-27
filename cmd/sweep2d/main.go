// cmd/sweep2d is the CLI entrypoint for the two-parameter grid sweep runner.
//
// It sweeps two strategy parameters simultaneously, producing a [param1 × param2]
// Sharpe ratio matrix as CSV and printing the DSR-corrected peak Sharpe to stderr.
//
// Usage:
//
//	go run ./cmd/sweep2d \
//	    --instrument "NSE:RELIANCE" \
//	    --from 2018-01-01 \
//	    --to   2024-01-01 \
//	    --timeframe daily \
//	    --cash 100000 \
//	    --strategy sma-crossover \
//	    --p1-name fast-period --p1-min 5  --p1-max 30 --p1-step 5 \
//	    --p2-name slow-period --p2-min 20 --p2-max 80 --p2-step 10 \
//	    --out sweep2d.csv
//
// Supported strategies and their axis mappings:
//
//	sma-crossover        p1=fast-period, p2=slow-period
//	rsi-mean-reversion   p1=rsi-period,  p2=oversold (overbought = 100 − oversold)
//
// When --out is omitted, CSV is written to stdout. DSR-corrected peak Sharpe is
// always written to stderr so it does not contaminate piped CSV output.
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). Token handling is
// identical to cmd/backtest.
package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep2d"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
)

func main() {
	instrument := flag.String("instrument", "NSE:NIFTY 50", "Instrument to sweep (e.g. \"NSE:RELIANCE\")")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	stratName := flag.String("strategy", "", "Strategy to sweep: sma-crossover | rsi-mean-reversion (required)")
	outPath := flag.String("out", "", "CSV output path; writes to stdout if omitted")

	// Param1 axis.
	p1Name := flag.String("p1-name", "param1", "Name for the first sweep axis")
	p1Min := flag.Float64("p1-min", 0, "First axis minimum (required)")
	p1Max := flag.Float64("p1-max", 0, "First axis maximum (required)")
	p1Step := flag.Float64("p1-step", 0, "First axis step size (required, must be > 0)")

	// Param2 axis.
	p2Name := flag.String("p2-name", "param2", "Name for the second sweep axis")
	p2Min := flag.Float64("p2-min", 0, "Second axis minimum (required)")
	p2Max := flag.Float64("p2-max", 0, "Second axis maximum (required)")
	p2Step := flag.Float64("p2-step", 0, "Second axis step size (required, must be > 0)")

	flag.Parse()

	f := flags2D{
		fromStr:   *fromStr,
		toStr:     *toStr,
		tfStr:     *tfStr,
		stratName: *stratName,
		p1Min:     *p1Min,
		p1Max:     *p1Max,
		p1Step:    *p1Step,
		p2Min:     *p2Min,
		p2Max:     *p2Max,
		p2Step:    *p2Step,
	}

	from, to, tf, err := parseAndValidateFlags2D(f)
	if err != nil {
		cmdutil.Fatalf("%v", err)
	}

	factory, err := factoryRegistry2D(*stratName, tf)
	if err != nil {
		cmdutil.Fatalf("--strategy: %v", err)
	}

	ctx := context.Background()
	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	cfg := sweep2d.Config2D{
		Param1:    sweep2d.ParamRange{Name: *p1Name, Min: *p1Min, Max: *p1Max, Step: *p1Step},
		Param2:    sweep2d.ParamRange{Name: *p2Name, Min: *p2Min, Max: *p2Max, Step: *p2Step},
		Timeframe: tf,
		EngineConfig: engine.Config{
			Instrument:           *instrument,
			From:                 from,
			To:                   to,
			InitialCash:          *cash,
			PositionSizeFraction: 0.1,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: model.CommissionZerodha,
			},
		},
		StrategyFactory: factory,
	}

	fmt.Fprintf(os.Stderr, "2D sweep: %s  %s[%g…%g step %g] × %s[%g…%g step %g]  %s  %s → %s\n",
		*stratName,
		*p1Name, *p1Min, *p1Max, *p1Step,
		*p2Name, *p2Min, *p2Max, *p2Step,
		*instrument,
		from.Format("2006-01-02"), to.Format("2006-01-02"),
	)

	report, err := sweep2d.Run(ctx, cfg, p)
	if err != nil {
		cmdutil.Fatalf("sweep2d: %v", err)
	}

	if err := writeOutput(os.Stdout, report, *outPath); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}

	fmt.Fprintf(os.Stderr, "peak Sharpe: %.4f  DSR-corrected: %.4f  (variants=%d)\n",
		report.PeakSharpe, report.DSRCorrectedPeakSharpe, report.VariantCount)
}

// flags2D groups the parsed flag strings for parseAndValidateFlags2D.
// **Decision (flags2D value struct for flag parsing) — convention: experimental**
// Passing flags as a value struct rather than individual parameters keeps
// parseAndValidateFlags2D testable without constructing a flag.FlagSet and
// avoids the 8-parameter function smell. Same approach as cmd/sweep's helpers.
type flags2D struct {
	fromStr   string
	toStr     string
	tfStr     string
	stratName string
	p1Min     float64
	p1Max     float64
	p1Step    float64
	p2Min     float64
	p2Max     float64
	p2Step    float64
}

// parseAndValidateFlags2D validates and parses the flag values for a 2D sweep.
func parseAndValidateFlags2D(f flags2D) (from, to time.Time, tf model.Timeframe, err error) { //nolint:gocritic // named returns document purpose
	if f.fromStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from is required (e.g. 2018-01-01)")
	}
	if f.toStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to is required (e.g. 2024-01-01)")
	}
	if f.stratName == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--strategy is required: sma-crossover | rsi-mean-reversion")
	}
	if f.p1Step <= 0 {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--p1-step must be > 0, got %g", f.p1Step)
	}
	if f.p2Step <= 0 {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--p2-step must be > 0, got %g", f.p2Step)
	}
	if f.p1Max < f.p1Min {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--p1-max (%g) must be >= --p1-min (%g)", f.p1Max, f.p1Min)
	}
	if f.p2Max < f.p2Min {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--p2-max (%g) must be >= --p2-min (%g)", f.p2Max, f.p2Min)
	}

	from, err = time.Parse("2006-01-02", f.fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from %q: %w", f.fromStr, err)
	}
	to, err = time.Parse("2006-01-02", f.toStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to %q: %w", f.toStr, err)
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to must be strictly after --from")
	}

	tf = model.Timeframe(f.tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", f.tfStr)
	}

	return from, to, tf, nil
}

// factoryRegistry2D returns a two-parameter StrategyFactory for the named strategy.
//
// Axis conventions (fixed by this registry):
//   - sma-crossover:      p1 = fast-period, p2 = slow-period
//   - rsi-mean-reversion: p1 = rsi-period,  p2 = oversold (overbought = 100 − oversold)
func factoryRegistry2D(stratName string, tf model.Timeframe) (func(float64, float64) (strategy.Strategy, error), error) {
	switch stratName {
	case "sma-crossover":
		return smaFactory2D(tf), nil
	case "rsi-mean-reversion":
		return rsiFactory2D(tf), nil
	default:
		return nil, fmt.Errorf("unsupported strategy %q for 2D sweep; supported: sma-crossover, rsi-mean-reversion", stratName)
	}
}

// smaFactory2D returns a factory where p1=fast-period, p2=slow-period.
//
// **Decision (sweep2d sma-crossover axis mapping: p1=fast, p2=slow) — convention: experimental**
// p1→fast, p2→slow is fixed by convention for the interaction surface (fast × slow grid).
// No flag allows axis swapping — the grid is always fast on rows, slow on columns.
func smaFactory2D(tf model.Timeframe) func(float64, float64) (strategy.Strategy, error) {
	return func(p1, p2 float64) (strategy.Strategy, error) {
		return smacrossover.New(tf, int(math.Round(p1)), int(math.Round(p2)))
	}
}

// rsiFactory2D returns a factory where p1=rsi-period, p2=oversold threshold.
// overbought is computed symmetrically as 100 − oversold.
func rsiFactory2D(tf model.Timeframe) func(float64, float64) (strategy.Strategy, error) {
	return func(p1, p2 float64) (strategy.Strategy, error) {
		return rsimeanrev.New(tf, int(math.Round(p1)), p2, 100-p2)
	}
}

// writeOutput writes the sweep report CSV to outPath, or to stdout if outPath is empty.
// DSR-corrected peak Sharpe is written to stderr by the caller after this returns.
//
// **Decision (sweep2d CSV writer via io.Writer helper) — convention: experimental**
// writeCSVToWriter(w io.Writer) serializes the CSV directly to any writer,
// avoiding a temp file for the stdout path and keeping the smoke test free of
// filesystem I/O. sweep2d.WriteCSV (file path API) is used only when --out is set.
// writeOutput writes the sweep report CSV to outPath when set, or to w otherwise.
//
// **Decision (writeOutput accepts io.Writer for the stdout path) — convention: experimental**
// Injecting the writer instead of calling os.Stdout directly makes the stdout path
// testable without redirecting the process's real stdout. The caller passes os.Stdout
// in production and a *bytes.Buffer in tests.
func writeOutput(w io.Writer, report sweep2d.Report2D, outPath string) error { //nolint:gocritic // Report2D is a caller-constructed result; value semantics match the pattern in sweep2d.WriteCSV
	if outPath != "" {
		if err := sweep2d.WriteCSV(outPath, report); err != nil {
			return fmt.Errorf("write CSV to %q: %w", outPath, err)
		}
		fmt.Fprintf(os.Stderr, "CSV written to %s\n", outPath)
		return nil
	}

	var buf bytes.Buffer
	if err := writeCSVToWriter(&buf, report); err != nil {
		return fmt.Errorf("write CSV to stdout: %w", err)
	}
	if _, err := w.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("stdout write: %w", err)
	}
	return nil
}

// writeCSVToWriter serializes the Sharpe ratio matrix from report as CSV to w.
// Format matches sweep2d.WriteCSV: corner label header, one row per Param1 value,
// metadata footer comment. Exported for use in tests.
func writeCSVToWriter(w io.Writer, report sweep2d.Report2D) error { //nolint:gocritic // Report2D is a caller-constructed result; value semantics are consistent with sweep2d package conventions
	corner := report.Param1Name + `\` + report.Param2Name
	if _, err := fmt.Fprint(w, corner); err != nil {
		return fmt.Errorf("write header corner: %w", err)
	}
	for _, v2 := range report.Param2Values {
		if _, err := fmt.Fprintf(w, ",%g", v2); err != nil {
			return fmt.Errorf("write header value: %w", err)
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return fmt.Errorf("write header newline: %w", err)
	}

	for i, v1 := range report.Param1Values {
		if _, err := fmt.Fprintf(w, "%g", v1); err != nil {
			return fmt.Errorf("write row label: %w", err)
		}
		for j := range report.Param2Values {
			if _, err := fmt.Fprintf(w, ",%.6f", report.Grid[i][j].SharpeRatio); err != nil {
				return fmt.Errorf("write cell: %w", err)
			}
		}
		if _, err := fmt.Fprintln(w); err != nil {
			return fmt.Errorf("write row newline: %w", err)
		}
	}

	if _, err := fmt.Fprintf(w, "# variants=%d  peak_sharpe=%.6f  dsr_corrected=%.6f\n",
		report.VariantCount, report.PeakSharpe, report.DSRCorrectedPeakSharpe); err != nil {
		return fmt.Errorf("write footer: %w", err)
	}
	return nil
}
