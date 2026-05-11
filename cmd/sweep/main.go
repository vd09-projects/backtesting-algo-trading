// cmd/sweep is the CLI entrypoint for the parameter sweep runner.
//
// Usage:
//
//	go run ./cmd/sweep \
//	    --instrument "NSE:NIFTY 50" \
//	    --from 2024-01-01 \
//	    --to   2024-12-31 \
//	    --timeframe daily \
//	    --cash 100000 \
//	    --strategy rsi-mean-reversion \
//	    --sweep-param rsi-period \
//	    --min 7 --max 21 --step 1
//
// Sweepable strategy+param combinations are registered in internal/cmdutil/strategies.go.
// Run with --help to see all available flags and their defaults.
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). Token handling is
// identical to cmd/backtest.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

func main() {
	factory := func(ctx context.Context) (provider.DataProvider, error) {
		return cmdutil.BuildProvider(ctx)
	}
	if err := run(os.Args[1:], os.Stdout, os.Stderr, factory); err != nil {
		cmdutil.Fatalf("%v", err)
	}
}

// run is the testable entry point. providerFactory is injected so tests can
// substitute a fake provider without requiring live Zerodha credentials.
func run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error {
	fs := flag.NewFlagSet("sweep", flag.ContinueOnError)
	fs.SetOutput(stderr)

	instrument := fs.String("instrument", "NSE:NIFTY 50", "Instrument to sweep (e.g. \"NSE:NIFTY 50\")")
	fromStr := fs.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := fs.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := fs.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := fs.Float64("cash", 100000, "Starting cash in ₹")
	stratName := fs.String("strategy", "", "Strategy to sweep: "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | ")+" (required)")
	sweepParam := fs.String("sweep-param", "", "Parameter to sweep (required; use --help to see available params per strategy)")
	minVal := fs.Float64("min", 0, "Sweep range minimum (required)")
	maxVal := fs.Float64("max", 0, "Sweep range maximum (required)")
	stepVal := fs.Float64("step", 0, "Sweep step size (required, must be > 0)")
	commissionStr := fs.String("commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")

	// Fixed parameters for the non-swept dimensions — registered centrally from
	// GlobalRegistry so adding a new strategy only requires a change to
	// internal/cmdutil/strategies.go.
	stratParamPtrs := cmdutil.GlobalRegistry.RegisterFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	from, to, tf, err := parseAndValidateFlags(*fromStr, *toStr, *tfStr, *stratName, *sweepParam, *stepVal, *minVal, *maxVal)
	if err != nil {
		return err
	}

	commissionModel, err := cmdutil.ParseCommissionModel(*commissionStr)
	if err != nil {
		return fmt.Errorf("--commission: %w", err)
	}

	// Name validation: MustGet panics at startup with descriptive message if unknown.
	cmdutil.GlobalRegistry.MustGet(*stratName)

	fixedParams := cmdutil.BuildParamMap(stratParamPtrs)

	factory, err := cmdutil.GlobalRegistry.SweepFactory(*stratName, *sweepParam, tf, fixedParams)
	if err != nil {
		return fmt.Errorf("--strategy / --sweep-param: %w", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := providerFactory(ctx)
	if err != nil {
		return fmt.Errorf("provider: %w", err)
	}

	cfg := sweep.Config{
		ParameterName:   *sweepParam,
		Min:             *minVal,
		Max:             *maxVal,
		Step:            *stepVal,
		StrategyFactory: factory,
		EngineConfig: engine.Config{
			Instrument:           *instrument,
			From:                 from,
			To:                   to,
			InitialCash:          *cash,
			PositionSizeFraction: 0.10,
			OrderConfig: model.OrderConfig{
				SlippagePct:     0.0005,
				CommissionModel: commissionModel,
			},
		},
		Timeframe: tf,
	}

	fmt.Fprintf(stderr, "Sweeping %s.%s in [%.4g, %.4g] step=%.4g  %s → %s  timeframe=%s commission=%s\n", //nolint:errcheck // progress to stderr; non-fatal
		*stratName, *sweepParam, *minVal, *maxVal, *stepVal,
		from.Format("2006-01-02"), to.Format("2006-01-02"), *tfStr, *commissionStr)

	report, err := sweep.Run(ctx, cfg, p)
	if err != nil {
		return fmt.Errorf("sweep: %w", err)
	}

	if err := output.WriteSweep(stdout, report); err != nil {
		return fmt.Errorf("write results: %w", err)
	}

	return nil
}

func parseAndValidateFlags(fromStr, toStr, tfStr, stratName, sweepParam string, stepVal, minVal, maxVal float64) (from, to time.Time, tf model.Timeframe, err error) {
	if fromStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from is required (e.g. 2024-01-01)")
	}
	if toStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to is required (e.g. 2024-12-31)")
	}
	if stratName == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--strategy is required (%s)",
			strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
	}
	if sweepParam == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--sweep-param is required (e.g. rsi-period, fast-period, oversold)")
	}
	if stepVal <= 0 {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--step must be > 0 (got %.4g)", stepVal)
	}
	if minVal >= maxVal {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--min (%.4g) must be < --max (%.4g)", minVal, maxVal)
	}

	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from %q: %w", fromStr, err)
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to %q: %w", toStr, err)
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to must be strictly after --from")
	}

	tf = model.Timeframe(tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		return time.Time{}, time.Time{}, "", fmt.Errorf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", tfStr)
	}

	return from, to, tf, nil
}
