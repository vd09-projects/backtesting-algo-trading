// cmd/universe-sweep runs a fixed strategy across a list of instruments defined
// in a universe file and produces a CSV report ranked by Sharpe ratio.
//
// Usage:
//
//	go run ./cmd/universe-sweep \
//	    --universe universes/nifty50-large-cap.yaml \
//	    --strategy sma-crossover \
//	    --from 2020-01-01 \
//	    --to   2024-12-31 \
//	    --timeframe daily \
//	    --cash 100000 \
//	    --commission zerodha_full
//
// The output is CSV written to stdout:
//
//	instrument,sharpe,trade_count,total_pnl,max_drawdown,insufficient_data
//
// Rows are sorted descending by Sharpe. Instruments with fewer than the minimum
// trades or candle-points required for reliable metrics are flagged with
// insufficient_data=true (Sharpe is zeroed for those rows).
//
// Commission models:
//
//	zerodha         — simplified Zerodha model (default)
//	zerodha_full    — full Zerodha CNC model with STT, exchange charges, SEBI, stamp duty, GST
//	zerodha_full_mis — full Zerodha MIS intraday model
//	flat            — flat per-trade fee
//	percentage      — percentage of notional
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
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
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

// run is the testable entry point. It parses args, validates flags, builds the
// strategy factory, connects to the provider, runs the sweep, and writes CSV output.
// providerFactory is injected so tests can substitute a fake provider without
// requiring live Zerodha credentials.
// It returns an error rather than calling os.Exit, so tests can invoke it directly
// without spawning a subprocess.
func run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error {
	fs := flag.NewFlagSet("universe-sweep", flag.ContinueOnError)
	fs.SetOutput(stderr)

	universeFile := fs.String("universe", "", "Path to YAML universe file (required)")
	stratName := fs.String("strategy", "", "Strategy name: "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), ", ")+" (required)")
	fromStr := fs.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := fs.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := fs.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := fs.Float64("cash", 100000, "Starting cash in ₹")
	positionSize := fs.Float64("position-size", 0.10, "Fraction of cash deployed per trade")
	slippage := fs.Float64("slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")
	commissionStr := fs.String("commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")

	// Strategy-specific parameters registered centrally from GlobalRegistry.
	stratParamPtrs := cmdutil.GlobalRegistry.RegisterFlags(fs)

	if err := fs.Parse(args); err != nil {
		return err
	}

	from, to, tf, err := parseAndValidateFlags(*universeFile, *stratName, *fromStr, *toStr, *tfStr)
	if err != nil {
		return err
	}

	commissionModel, err := cmdutil.ParseCommissionModel(*commissionStr)
	if err != nil {
		return fmt.Errorf("--commission: %w", err)
	}

	instruments, err := universesweep.ParseUniverseFile(*universeFile)
	if err != nil {
		return fmt.Errorf("universe file: %w", err)
	}

	stratParams := cmdutil.BuildParamMap(stratParamPtrs)

	// WalkForwardFactory validates params eagerly and returns a fresh-instance factory.
	// Each instrument gets its own strategy state — no bleed between runs.
	strategyFactory, err := cmdutil.GlobalRegistry.WalkForwardFactory(*stratName, tf, stratParams)
	if err != nil {
		return fmt.Errorf("--strategy: %w", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := providerFactory(ctx)
	if err != nil {
		return fmt.Errorf("provider: %w", err)
	}

	fmt.Fprintf(stderr, "Running %q across %d instruments  %s → %s  timeframe=%s commission=%s\n", //nolint:errcheck // progress to stderr; non-fatal
		*stratName, len(instruments), from.Format("2006-01-02"), to.Format("2006-01-02"), *tfStr, *commissionStr)

	cfg := universesweep.Config{
		Instruments: instruments,
		NewStrategy: strategyFactory,
		EngineConfig: engine.Config{
			From:                 from,
			To:                   to,
			InitialCash:          *cash,
			PositionSizeFraction: *positionSize,
			OrderConfig: model.OrderConfig{
				SlippagePct:     *slippage,
				CommissionModel: commissionModel,
			},
		},
		Timeframe: tf,
	}

	report, err := universesweep.Run(ctx, &cfg, p)
	if err != nil {
		return fmt.Errorf("universe sweep: %w", err)
	}

	if err := universesweep.WriteCSV(stdout, report); err != nil {
		return fmt.Errorf("write CSV: %w", err)
	}

	return nil
}

// parseAndValidateFlags validates required flags and parses dates and timeframe.
// Returns an error rather than calling os.Exit so callers can be tested without
// subprocess spawning.
func parseAndValidateFlags(universeFile, stratName, fromStr, toStr, tfStr string) (from, to time.Time, tf model.Timeframe, err error) {
	if universeFile == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if stratName == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--strategy is required (%s)",
			strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
	}
	if fromStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from is required (e.g. 2020-01-01)")
	}
	if toStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to is required (e.g. 2024-12-31)")
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
