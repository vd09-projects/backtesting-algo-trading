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
	"os"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

func main() {
	universeFile := flag.String("universe", "", "Path to YAML universe file (required)")
	stratName := flag.String("strategy", "", "Strategy name: "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), ", ")+" (required)")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	positionSize := flag.Float64("position-size", 0.10, "Fraction of cash deployed per trade")
	slippage := flag.Float64("slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")
	commissionStr := flag.String("commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")

	// Strategy-specific parameters registered centrally from GlobalRegistry.
	stratParamPtrs := cmdutil.GlobalRegistry.RegisterFlags(flag.CommandLine)

	flag.Parse()

	if *universeFile == "" {
		cmdutil.Fatalf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if *stratName == "" {
		cmdutil.Fatalf("--strategy is required (%s)", strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
	}
	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2020-01-01)")
	}
	if *toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2024-12-31)")
	}

	from, to, tf := parseDateRangeAndTimeframe(*fromStr, *toStr, *tfStr)

	commissionModel, err := cmdutil.ParseCommissionModel(*commissionStr)
	if err != nil {
		cmdutil.Fatalf("--commission: %v", err)
	}

	instruments, err := universesweep.ParseUniverseFile(*universeFile)
	if err != nil {
		cmdutil.Fatalf("universe file: %v", err)
	}

	// Validate strategy name early: MustGet panics with a descriptive message if
	// the name is unknown.
	cmdutil.GlobalRegistry.MustGet(*stratName)

	stratParams := cmdutil.BuildParamMap(stratParamPtrs)

	selectedStrategy, err := cmdutil.GlobalRegistry.Build(*stratName, tf, stratParams)
	if err != nil {
		cmdutil.Fatalf("--strategy: %v", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Running %q across %d instruments  %s → %s  timeframe=%s commission=%s\n",
		*stratName, len(instruments), from.Format("2006-01-02"), to.Format("2006-01-02"), *tfStr, *commissionStr)

	cfg := universesweep.Config{
		Instruments: instruments,
		Strategy:    selectedStrategy,
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
		cmdutil.Fatalf("universe sweep: %v", err)
	}

	if err := universesweep.WriteCSV(os.Stdout, report); err != nil {
		cmdutil.Fatalf("write CSV: %v", err)
	}
}

func parseDateRangeAndTimeframe(fromStr, toStr, tfStr string) (from, to time.Time, tf model.Timeframe) {
	var err error
	from, err = time.Parse("2006-01-02", fromStr)
	if err != nil {
		cmdutil.Fatalf("--from %q: %v", fromStr, err)
	}
	to, err = time.Parse("2006-01-02", toStr)
	if err != nil {
		cmdutil.Fatalf("--to %q: %v", toStr, err)
	}
	if !to.After(from) {
		cmdutil.Fatalf("--to must be strictly after --from")
	}
	tf = model.Timeframe(tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		cmdutil.Fatalf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", tfStr)
	}
	return
}
