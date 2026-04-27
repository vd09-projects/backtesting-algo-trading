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
// Supported strategy + sweep-param combinations:
//
//	sma-crossover     + fast-period   (--slow-period sets the fixed slow period)
//	sma-crossover     + slow-period   (--fast-period sets the fixed fast period)
//	rsi-mean-reversion + rsi-period   (--oversold / --overbought set fixed thresholds)
//	rsi-mean-reversion + oversold     (overbought = 100 − oversold; --rsi-period sets fixed period)
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). Token handling is
// identical to cmd/backtest.
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
)

func main() {
	instrument := flag.String("instrument", "NSE:NIFTY 50", "Instrument to sweep (e.g. \"NSE:NIFTY 50\")")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	stratName := flag.String("strategy", "", "Strategy to sweep: sma-crossover | rsi-mean-reversion | donchian-breakout (required)")
	sweepParam := flag.String("sweep-param", "", "Parameter to sweep (required; see supported combinations in usage)")
	minVal := flag.Float64("min", 0, "Sweep range minimum (required)")
	maxVal := flag.Float64("max", 0, "Sweep range maximum (required)")
	stepVal := flag.Float64("step", 0, "Sweep step size (required, must be > 0)")

	// Fixed parameters for the non-swept dimensions.
	fastPeriod := flag.Int("fast-period", 10, "sma-crossover: fixed fast SMA period")
	slowPeriod := flag.Int("slow-period", 50, "sma-crossover: fixed slow SMA period")
	rsiPeriod := flag.Int("rsi-period", 14, "rsi-mean-reversion: fixed RSI period")
	oversold := flag.Float64("oversold", 30, "rsi-mean-reversion: fixed oversold threshold")
	overbought := flag.Float64("overbought", 70, "rsi-mean-reversion: fixed overbought threshold")
	donchianPeriod := flag.Int("donchian-period", 20, "donchian-breakout: fixed channel period")
	macdFastPeriod := flag.Int("macd-fast-period", 12, "macd-crossover: fixed fast EMA period")
	macdSlowPeriod := flag.Int("macd-slow-period", 26, "macd-crossover: fixed slow EMA period")
	macdSignalPeriod := flag.Int("macd-signal-period", 9, "macd-crossover: fixed signal EMA period")
	bbPeriod := flag.Int("bb-period", 20, "bollinger-mean-reversion: fixed Bollinger Band period")
	bbNumStdDev := flag.Float64("bb-num-std-dev", 2.0, "bollinger-mean-reversion: fixed number of standard deviations")

	flag.Parse()

	from, to, tf, err := parseAndValidateFlags(*fromStr, *toStr, *tfStr, *stratName, *sweepParam, *stepVal, *minVal, *maxVal)
	if err != nil {
		cmdutil.Fatalf("%v", err)
	}

	factory, err := factoryRegistry(*stratName, *sweepParam, tf, &fixedParams{
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
		cmdutil.Fatalf("--strategy / --sweep-param: %v", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	cfg := sweep.Config{
		ParameterName: *sweepParam,
		Min:           *minVal,
		Max:           *maxVal,
		Step:          *stepVal,
		Timeframe:     tf,
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

	fmt.Printf("Sweeping %s.%s [%.4g … %.4g step %.4g] on %s  %s → %s\n",
		*stratName, *sweepParam, *minVal, *maxVal, *stepVal,
		*instrument, from.Format("2006-01-02"), to.Format("2006-01-02"))

	report, err := sweep.Run(ctx, cfg, p)
	if err != nil {
		cmdutil.Fatalf("sweep: %v", err)
	}

	if err := output.WriteSweep(os.Stdout, report); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}
}

// parseAndValidateFlags validates required flags and parses dates and timeframe.
func parseAndValidateFlags(fromStr, toStr, tfStr, stratName, sweepParam string, stepVal, minVal, maxVal float64) (from, to time.Time, tf model.Timeframe, err error) { //nolint:gocritic // named returns document purpose of each position
	if fromStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--from is required (e.g. 2024-01-01)")
	}
	if toStr == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--to is required (e.g. 2024-12-31)")
	}
	if stratName == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--strategy is required: sma-crossover | rsi-mean-reversion | donchian-breakout")
	}
	if sweepParam == "" {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--sweep-param is required (e.g. rsi-period, fast-period, oversold)")
	}
	if stepVal <= 0 {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--step must be > 0")
	}
	if maxVal < minVal {
		return time.Time{}, time.Time{}, "", fmt.Errorf("--max must be >= --min")
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

type fixedParams struct {
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

// factoryRegistry returns a StrategyFactory for the given strategy and sweep-param combination.
func factoryRegistry(stratName, sweepParam string, tf model.Timeframe, fixed *fixedParams) (func(float64) (strategy.Strategy, error), error) {
	switch stratName {
	case "sma-crossover":
		switch sweepParam {
		case "fast-period":
			return func(v float64) (strategy.Strategy, error) {
				return smacrossover.New(tf, int(math.Round(v)), fixed.slowPeriod)
			}, nil
		case "slow-period":
			return func(v float64) (strategy.Strategy, error) {
				return smacrossover.New(tf, fixed.fastPeriod, int(math.Round(v)))
			}, nil
		default:
			return nil, fmt.Errorf("sma-crossover does not support sweep-param %q; use fast-period or slow-period", sweepParam)
		}

	case "rsi-mean-reversion":
		switch sweepParam {
		case "rsi-period":
			return func(v float64) (strategy.Strategy, error) {
				return rsimeanrev.New(tf, int(math.Round(v)), fixed.oversold, fixed.overbought)
			}, nil
		case "oversold":
			// Symmetric convention: overbought = 100 − oversold.
			return func(v float64) (strategy.Strategy, error) {
				return rsimeanrev.New(tf, fixed.rsiPeriod, v, 100-v)
			}, nil
		default:
			return nil, fmt.Errorf("rsi-mean-reversion does not support sweep-param %q; use rsi-period or oversold", sweepParam)
		}

	case "donchian-breakout":
		switch sweepParam {
		case "donchian-period":
			return func(v float64) (strategy.Strategy, error) {
				return donchian.New(tf, int(math.Round(v)))
			}, nil
		default:
			return nil, fmt.Errorf("donchian-breakout does not support sweep-param %q; use donchian-period", sweepParam)
		}

	case "macd-crossover":
		return macdFactory(sweepParam, tf, fixed)

	case "bollinger-mean-reversion":
		return bollingerFactory(sweepParam, tf, fixed)

	default:
		return nil, fmt.Errorf("unknown strategy %q; available: sma-crossover, rsi-mean-reversion, donchian-breakout, macd-crossover, bollinger-mean-reversion", stratName)
	}
}

func macdFactory(sweepParam string, tf model.Timeframe, fixed *fixedParams) (func(float64) (strategy.Strategy, error), error) {
	switch sweepParam {
	case "macd-fast-period":
		return func(v float64) (strategy.Strategy, error) {
			return macd.New(tf, int(math.Round(v)), fixed.macdSlowPeriod, fixed.macdSignalPeriod)
		}, nil
	case "macd-slow-period":
		return func(v float64) (strategy.Strategy, error) {
			return macd.New(tf, fixed.macdFastPeriod, int(math.Round(v)), fixed.macdSignalPeriod)
		}, nil
	default:
		return nil, fmt.Errorf("macd-crossover does not support sweep-param %q; use macd-fast-period or macd-slow-period", sweepParam)
	}
}

func bollingerFactory(sweepParam string, tf model.Timeframe, fixed *fixedParams) (func(float64) (strategy.Strategy, error), error) {
	switch sweepParam {
	case "bb-period":
		return func(v float64) (strategy.Strategy, error) {
			return bollinger.New(tf, int(math.Round(v)), fixed.bbNumStdDev)
		}, nil
	case "bb-num-std-dev":
		return func(v float64) (strategy.Strategy, error) {
			return bollinger.New(tf, fixed.bbPeriod, v)
		}, nil
	default:
		return nil, fmt.Errorf("bollinger-mean-reversion does not support sweep-param %q; use bb-period or bb-num-std-dev", sweepParam)
	}
}
