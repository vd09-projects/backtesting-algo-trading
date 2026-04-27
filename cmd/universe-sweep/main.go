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
//	    --cash 100000
//
// The output is CSV written to stdout:
//
//	instrument,sharpe,trade_count,total_pnl,max_drawdown,insufficient_data
//
// Rows are sorted descending by Sharpe. Instruments with fewer than the minimum
// trades or candle-points required for reliable metrics are flagged with
// insufficient_data=true (Sharpe is zeroed for those rows).
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
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
)

func main() {
	universeFile := flag.String("universe", "", "Path to YAML universe file (required)")
	stratName := flag.String("strategy", "", "Strategy name: sma-crossover, rsi-mean-reversion, donchian-breakout (required)")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	positionSize := flag.Float64("position-size", 0.10, "Fraction of cash deployed per trade")
	slippage := flag.Float64("slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")

	// Strategy-specific parameters.
	fastPeriod := flag.Int("fast-period", 10, "sma-crossover: fast SMA period")
	slowPeriod := flag.Int("slow-period", 50, "sma-crossover: slow SMA period")
	rsiPeriod := flag.Int("rsi-period", 14, "rsi-mean-reversion: RSI period")
	oversold := flag.Float64("oversold", 30, "rsi-mean-reversion: oversold threshold")
	overbought := flag.Float64("overbought", 70, "rsi-mean-reversion: overbought threshold")
	donchianPeriod := flag.Int("donchian-period", 20, "donchian-breakout: channel lookback period")
	macdFastPeriod := flag.Int("macd-fast-period", 12, "macd-crossover: fast EMA period")
	macdSlowPeriod := flag.Int("macd-slow-period", 26, "macd-crossover: slow EMA period")
	macdSignalPeriod := flag.Int("macd-signal-period", 9, "macd-crossover: signal EMA period")
	bbPeriod := flag.Int("bb-period", 20, "bollinger-mean-reversion: Bollinger Band period")
	bbNumStdDev := flag.Float64("bb-num-std-dev", 2.0, "bollinger-mean-reversion: number of standard deviations")

	flag.Parse()

	if *universeFile == "" {
		cmdutil.Fatalf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if *stratName == "" {
		cmdutil.Fatalf("--strategy is required (sma-crossover | rsi-mean-reversion)")
	}
	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2020-01-01)")
	}
	if *toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2024-12-31)")
	}

	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		cmdutil.Fatalf("--from %q: %v", *fromStr, err)
	}
	to, err := time.Parse("2006-01-02", *toStr)
	if err != nil {
		cmdutil.Fatalf("--to %q: %v", *toStr, err)
	}
	if !to.After(from) {
		cmdutil.Fatalf("--to must be strictly after --from")
	}

	tf := model.Timeframe(*tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		cmdutil.Fatalf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", *tfStr)
	}

	instruments, err := universesweep.ParseUniverseFile(*universeFile)
	if err != nil {
		cmdutil.Fatalf("universe file: %v", err)
	}

	selectedStrategy, err := strategyRegistry(*stratName, tf, &strategyParams{
		fastPeriod:       *fastPeriod,
		slowPeriod:       *slowPeriod,
		rsiPeriod:        *rsiPeriod,
		oversold:         *oversold,
		overbought:       *overbought,
		donchianPeriod:   *donchianPeriod,
		macdFastPeriod:   *macdFastPeriod,
		macdSlowPeriod:   *macdSlowPeriod,
		macdSignalPeriod: *macdSignalPeriod,
		bbPeriod:         *bbPeriod,
		bbNumStdDev:      *bbNumStdDev,
	})
	if err != nil {
		cmdutil.Fatalf("--strategy: %v", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Running %q across %d instruments  %s → %s\n",
		*stratName, len(instruments), from.Format("2006-01-02"), to.Format("2006-01-02"))

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
				CommissionModel: model.CommissionZerodha,
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

type strategyParams struct {
	fastPeriod       int
	slowPeriod       int
	rsiPeriod        int
	oversold         float64
	overbought       float64
	donchianPeriod   int
	macdFastPeriod   int
	macdSlowPeriod   int
	macdSignalPeriod int
	bbPeriod         int
	bbNumStdDev      float64
}

func strategyRegistry(name string, tf model.Timeframe, p *strategyParams) (strategy.Strategy, error) {
	switch name {
	case "sma-crossover":
		return smacrossover.New(tf, p.fastPeriod, p.slowPeriod)
	case "rsi-mean-reversion":
		return rsimeanrev.New(tf, p.rsiPeriod, p.oversold, p.overbought)
	case "donchian-breakout":
		return donchian.New(tf, p.donchianPeriod)
	case "macd-crossover":
		return macd.New(tf, p.macdFastPeriod, p.macdSlowPeriod, p.macdSignalPeriod)
	case "bollinger-mean-reversion":
		return bollinger.New(tf, p.bbPeriod, p.bbNumStdDev)
	default:
		return nil, fmt.Errorf("unknown strategy %q; available: sma-crossover, rsi-mean-reversion, donchian-breakout, macd-crossover, bollinger-mean-reversion", name)
	}
}
