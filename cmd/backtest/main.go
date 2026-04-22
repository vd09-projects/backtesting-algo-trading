// cmd/backtest is the CLI entrypoint for the backtesting engine.
//
// Usage:
//
//	go run ./cmd/backtest \
//	    --instrument "NSE:NIFTY 50" \
//	    --from 2024-01-01 \
//	    --to   2024-12-31 \
//	    --timeframe daily \
//	    --cash 100000 \
//	    --strategy sma-crossover \
//	    --out results.json
//
// With volatility-targeting sizing:
//
//	go run ./cmd/backtest \
//	    --instrument "NSE:INFY" \
//	    --from 2018-01-01 \
//	    --to   2025-01-01 \
//	    --timeframe daily \
//	    --strategy rsi-mean-reversion \
//	    --sizing-model vol-target \
//	    --vol-target 0.10 \
//	    --out results.json
//
// With equity curve export:
//
//	go run ./cmd/backtest \
//	    --instrument "NSE:RELIANCE" \
//	    --from 2018-01-01 \
//	    --to   2025-01-01 \
//	    --timeframe daily \
//	    --strategy sma-crossover \
//	    --out runs/sma-crossover.json \
//	    --output-curve runs/sma-crossover-curve.csv
//
// Available strategies:
//
//	stub              — always holds; useful for smoke-testing the pipeline
//	sma-crossover     — SMA crossover; --fast-period / --slow-period
//	rsi-mean-reversion — RSI mean-reversion; --rsi-period / --oversold / --overbought
//
// Sizing models:
//
//	fixed      — deploy a fixed fraction of cash per trade (default; controlled by --position-size)
//	vol-target — size each trade so annualized dollar vol = cash × --vol-target (default 10%)
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). A saved access token is
// reused when present at ~/.config/backtest/token.json and not expired;
// otherwise the Kite Connect login flow is triggered.
//
// Optional overrides:
//
//	BACKTEST_TOKEN_PATH — override the token file location
//	BACKTEST_CACHE_DIR  — override the candle cache directory (default: .cache/zerodha)
package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
	stubstrategy "github.com/vikrantdhawan/backtesting-algo-trading/strategies/stub"
)

func main() {
	instrument := flag.String("instrument", "NSE:NIFTY 50", "Instrument to backtest (e.g. \"NSE:NIFTY 50\", \"NSE:INFY\")")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	stratName := flag.String("strategy", "stub", "Strategy name: stub, sma-crossover, rsi-mean-reversion")
	fastPeriod := flag.Int("fast-period", 10, "sma-crossover: fast SMA period")
	slowPeriod := flag.Int("slow-period", 50, "sma-crossover: slow SMA period")
	rsiPeriod := flag.Int("rsi-period", 14, "rsi-mean-reversion: RSI period")
	oversold := flag.Float64("oversold", 30, "rsi-mean-reversion: oversold threshold (buy below)")
	overbought := flag.Float64("overbought", 70, "rsi-mean-reversion: overbought threshold (sell above)")
	outPath := flag.String("out", "", "Path for JSON results export (omit to skip)")
	curvePath := flag.String("output-curve", "", "Path for equity curve CSV export (omit to skip)")
	sizingModel := flag.String("sizing-model", "fixed", "Position sizing model: fixed | vol-target")
	volTarget := flag.Float64("vol-target", 0.10, "Annualized volatility target when --sizing-model=vol-target (e.g. 0.10 = 10%)")
	gateThreshold  := flag.Float64("proliferation-gate-threshold", 0.0, "Sharpe threshold for proliferation gate PASS/FAIL (0 = disabled; 0.5 recommended for NSE daily)")
	doBootstrap    := flag.Bool("bootstrap", false, "Run Monte Carlo bootstrap for Sharpe confidence intervals")
	bootstrapSeed  := flag.Int64("bootstrap-seed", 42, "RNG seed for bootstrap (logged with results for reproducibility)")
	bootstrapN     := flag.Int("bootstrap-n", 0, "Bootstrap simulation count (0 = default 10,000)")
	flag.Parse()

	// Validate required flags.
	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2024-01-01)")
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

	selectedStrategy, err := strategyRegistry(*stratName, tf, strategyParams{
		fastPeriod: *fastPeriod,
		slowPeriod: *slowPeriod,
		rsiPeriod:  *rsiPeriod,
		oversold:   *oversold,
		overbought: *overbought,
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

	sm, err := parseSizingConfig(*sizingModel, *volTarget)
	if err != nil {
		cmdutil.Fatalf("%v", err)
	}

	eng := engine.New(engine.Config{
		Instrument:           *instrument,
		From:                 from,
		To:                   to,
		InitialCash:          *cash,
		PositionSizeFraction: 0.1,
		SizingModel:          sm,
		VolatilityTarget:     *volTarget,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	})

	fmt.Printf("Running strategy %q on %s  %s → %s\n",
		selectedStrategy.Name(), *instrument,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	if err := eng.Run(ctx, p, selectedStrategy); err != nil {
		cmdutil.Fatalf("engine: %v", err)
	}

	port := eng.Portfolio()
	curve := port.EquityCurve()
	trades := port.ClosedTrades()
	report := analytics.Compute(trades, curve, tf)
	benchmark := analytics.ComputeBenchmark(eng.Candles(), *cash)

	var regimeSplits []analytics.RegimeReport
	if *curvePath != "" {
		regimeSplits = analytics.ComputeRegimeSplits(curve, analytics.NSERegimes2018_2024, tf)
	}

	var bootstrapResult *montecarlo.BootstrapResult
	if *doBootstrap {
		if len(trades) < 2 {
			fmt.Println("NOTE: --bootstrap skipped — fewer than 2 closed trades")
		} else {
			r := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{
				NSimulations: *bootstrapN,
				Seed:         *bootstrapSeed,
			})
			bootstrapResult = &r
		}
	}

	if err := output.Write(report, output.Config{
		PrintToStdout:  true,
		FilePath:       *outPath,
		Benchmark:      &benchmark,
		CurvePath:      *curvePath,
		Curve:          curve,
		GateThreshold:  *gateThreshold,
		RegimeSplits:   regimeSplits,
		Bootstrap:      bootstrapResult,
		BootstrapSeed:  *bootstrapSeed,
		BootstrapNSims: *bootstrapN,
	}); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}
}

type strategyParams struct {
	fastPeriod int
	slowPeriod int
	rsiPeriod  int
	oversold   float64
	overbought float64
}

func strategyRegistry(name string, tf model.Timeframe, p strategyParams) (strategy.Strategy, error) {
	switch name {
	case "stub":
		return stubstrategy.New(tf), nil
	case "sma-crossover":
		return smacrossover.New(tf, p.fastPeriod, p.slowPeriod)
	case "rsi-mean-reversion":
		return rsimeanrev.New(tf, p.rsiPeriod, p.oversold, p.overbought)
	default:
		return nil, fmt.Errorf("unknown strategy %q; available: stub, sma-crossover, rsi-mean-reversion", name)
	}
}

func parseSizingModel(s string) (model.SizingModel, error) {
	switch s {
	case "fixed":
		return model.SizingFixed, nil
	case "vol-target":
		return model.SizingVolatilityTarget, nil
	default:
		return 0, fmt.Errorf("%q is not valid; choose one of: fixed, vol-target", s)
	}
}

func parseSizingConfig(sizingModel string, volTarget float64) (model.SizingModel, error) {
	sm, err := parseSizingModel(sizingModel)
	if err != nil {
		return 0, fmt.Errorf("--sizing-model: %w", err)
	}
	if sm == model.SizingVolatilityTarget && volTarget <= 0 {
		return 0, fmt.Errorf("--vol-target must be positive when --sizing-model=vol-target (got %.4f)", volTarget)
	}
	return sm, nil
}

