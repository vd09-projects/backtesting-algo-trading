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
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/momentum"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
	stubstrategy "github.com/vikrantdhawan/backtesting-algo-trading/strategies/stub"
)

// flags holds all parsed CLI flag values for cmd/backtest.
type flags struct {
	instrument        string
	fromStr           string
	toStr             string
	tfStr             string
	cash              float64
	stratName         string
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
	commissionStr     string
	outPath           string
	curvePath         string
	sizingModel       string
	volTarget         float64
	gateThreshold     float64
	doBootstrap       bool
	bootstrapSeed     int64
	bootstrapN        int
}

func main() {
	var f flags
	flag.StringVar(&f.instrument, "instrument", "NSE:NIFTY 50", "Instrument to backtest (e.g. \"NSE:NIFTY 50\", \"NSE:INFY\")")
	flag.StringVar(&f.fromStr, "from", "", "Start date in YYYY-MM-DD (inclusive)")
	flag.StringVar(&f.toStr, "to", "", "End date in YYYY-MM-DD (exclusive)")
	flag.StringVar(&f.tfStr, "timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	flag.Float64Var(&f.cash, "cash", 100000, "Starting cash in ₹")
	flag.StringVar(&f.stratName, "strategy", "stub", "Strategy name: stub, sma-crossover, rsi-mean-reversion, donchian-breakout, macd-crossover, bollinger-mean-reversion, momentum")
	flag.IntVar(&f.fastPeriod, "fast-period", 10, "sma-crossover: fast SMA period")
	flag.IntVar(&f.slowPeriod, "slow-period", 50, "sma-crossover: slow SMA period")
	flag.IntVar(&f.rsiPeriod, "rsi-period", 14, "rsi-mean-reversion: RSI period")
	flag.Float64Var(&f.oversold, "oversold", 30, "rsi-mean-reversion: oversold threshold (buy below)")
	flag.Float64Var(&f.overbought, "overbought", 70, "rsi-mean-reversion: overbought threshold (sell above)")
	flag.IntVar(&f.donchianPeriod, "donchian-period", 20, "donchian-breakout: channel lookback period")
	flag.IntVar(&f.macdFastPeriod, "macd-fast-period", 12, "macd-crossover: fast EMA period")
	flag.IntVar(&f.macdSlowPeriod, "macd-slow-period", 26, "macd-crossover: slow EMA period")
	flag.IntVar(&f.macdSignalPeriod, "macd-signal-period", 9, "macd-crossover: signal EMA period")
	flag.IntVar(&f.bbPeriod, "bb-period", 20, "bollinger-mean-reversion: Bollinger Band period")
	flag.Float64Var(&f.bbNumStdDev, "bb-num-std-dev", 2.0, "bollinger-mean-reversion: number of standard deviations")
	flag.IntVar(&f.momentumLookback, "momentum-lookback", 231, "momentum: ROC lookback period (default 231 = 252-21, skip-last-month convention)")
	flag.Float64Var(&f.momentumThreshold, "momentum-threshold", 10.0, "momentum: ROC threshold in percent (buy above, sell below negative)")
	flag.StringVar(&f.commissionStr, "commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")
	flag.StringVar(&f.outPath, "out", "", "Path for JSON results export; when omitted a default name is generated from the run params")
	flag.StringVar(&f.curvePath, "output-curve", "", "Path for equity curve CSV export (omit to skip)")
	flag.StringVar(&f.sizingModel, "sizing-model", "fixed", "Position sizing model: fixed | vol-target")
	flag.Float64Var(&f.volTarget, "vol-target", 0.10, "Annualized volatility target when --sizing-model=vol-target (e.g. 0.10 = 10%)")
	flag.Float64Var(&f.gateThreshold, "proliferation-gate-threshold", 0.0, "Sharpe threshold for proliferation gate PASS/FAIL (0 = disabled; 0.5 recommended for NSE daily)")
	flag.BoolVar(&f.doBootstrap, "bootstrap", false, "Run Monte Carlo bootstrap for Sharpe confidence intervals")
	flag.Int64Var(&f.bootstrapSeed, "bootstrap-seed", 42, "RNG seed for bootstrap (logged with results for reproducibility)")
	flag.IntVar(&f.bootstrapN, "bootstrap-n", 0, "Bootstrap simulation count (0 = default 10,000)")
	flag.Parse()

	from, to, tf := parseAndValidateFlags(&f)

	p := collectStrategyParams(&f)
	selectedStrategy, err := strategyRegistry(f.stratName, tf, p)
	if err != nil {
		cmdutil.Fatalf("--strategy: %v", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	provider, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	sm, err := parseSizingConfig(f.sizingModel, f.volTarget)
	if err != nil {
		cmdutil.Fatalf("%v", err)
	}

	commissionModel, err := cmdutil.ParseCommissionModel(f.commissionStr)
	if err != nil {
		cmdutil.Fatalf("--commission: %v", err)
	}

	// Auto-generate default output path when --out is not supplied.
	// The generated name includes timeframe so daily and intraday runs are distinguishable.
	if f.outPath == "" {
		f.outPath = cmdutil.DefaultOutPath(f.stratName, f.instrument, f.tfStr,
			from.Format("2006-01-02"), to.Format("2006-01-02"))
	}

	eng := engine.New(engine.Config{
		Instrument:           f.instrument,
		From:                 from,
		To:                   to,
		InitialCash:          f.cash,
		PositionSizeFraction: 0.1,
		SizingModel:          sm,
		VolatilityTarget:     f.volTarget,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: commissionModel,
		},
	})

	fmt.Printf("Running strategy %q on %s  %s → %s\n",
		selectedStrategy.Name(), f.instrument,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	if err := eng.Run(ctx, provider, selectedStrategy); err != nil {
		cmdutil.Fatalf("engine: %v", err)
	}

	port := eng.Portfolio()
	curve := port.EquityCurve()
	trades := port.ClosedTrades()
	report := analytics.Compute(trades, curve, tf)
	benchmark := analytics.ComputeBenchmark(eng.Candles(), f.cash)

	var regimeSplits []analytics.RegimeReport
	if f.curvePath != "" {
		regimeSplits = analytics.ComputeRegimeSplits(curve, analytics.NSERegimes2018_2024, tf)
	}

	bootstrapResult := runBootstrap(f.doBootstrap, trades, f.bootstrapSeed, f.bootstrapN)

	runCfg := buildRunConfig(f.stratName, f.instrument, f.tfStr, f.fromStr, f.toStr, f.commissionStr, p)

	if err := output.Write(report, output.Config{
		PrintToStdout:  true,
		FilePath:       f.outPath,
		Benchmark:      &benchmark,
		CurvePath:      f.curvePath,
		Curve:          curve,
		GateThreshold:  f.gateThreshold,
		RegimeSplits:   regimeSplits,
		Bootstrap:      bootstrapResult,
		BootstrapSeed:  f.bootstrapSeed,
		BootstrapNSims: f.bootstrapN,
		RunConfig:      runCfg,
	}); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}
}

// parseAndValidateFlags validates required flags, parses dates and timeframe,
// and calls cmdutil.Fatalf on any error.
func parseAndValidateFlags(f *flags) (from, to time.Time, tf model.Timeframe) {
	if f.fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2024-01-01)")
	}
	if f.toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2024-12-31)")
	}

	var err error
	from, err = time.Parse("2006-01-02", f.fromStr)
	if err != nil {
		cmdutil.Fatalf("--from %q: %v", f.fromStr, err)
	}
	to, err = time.Parse("2006-01-02", f.toStr)
	if err != nil {
		cmdutil.Fatalf("--to %q: %v", f.toStr, err)
	}
	if !to.After(from) {
		cmdutil.Fatalf("--to must be strictly after --from")
	}

	tf = model.Timeframe(f.tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		cmdutil.Fatalf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", f.tfStr)
	}
	return from, to, tf
}

// collectStrategyParams builds a strategyParams from the flags struct.
func collectStrategyParams(f *flags) *strategyParams {
	return &strategyParams{
		fastPeriod:        f.fastPeriod,
		slowPeriod:        f.slowPeriod,
		rsiPeriod:         f.rsiPeriod,
		oversold:          f.oversold,
		overbought:        f.overbought,
		donchianPeriod:    f.donchianPeriod,
		macdFastPeriod:    f.macdFastPeriod,
		macdSlowPeriod:    f.macdSlowPeriod,
		macdSignalPeriod:  f.macdSignalPeriod,
		bbPeriod:          f.bbPeriod,
		bbNumStdDev:       f.bbNumStdDev,
		momentumLookback:  f.momentumLookback,
		momentumThreshold: f.momentumThreshold,
	}
}

func runBootstrap(enabled bool, trades []model.Trade, seed int64, nSims int) *montecarlo.BootstrapResult {
	if !enabled {
		return nil
	}
	if len(trades) < 2 {
		fmt.Println("NOTE: --bootstrap skipped — fewer than 2 closed trades")
		return nil
	}
	r := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: nSims, Seed: seed})
	return &r
}

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
}

func strategyRegistry(name string, tf model.Timeframe, p *strategyParams) (strategy.Strategy, error) {
	switch name {
	case "stub":
		return stubstrategy.New(tf), nil
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
	case "momentum":
		return momentum.New(tf, p.momentumLookback, p.momentumThreshold)
	default:
		return nil, fmt.Errorf("unknown strategy %q; available: stub, sma-crossover, rsi-mean-reversion, donchian-breakout, macd-crossover, bollinger-mean-reversion, momentum", name)
	}
}

// buildRunConfig assembles the output.RunConfig metadata for a backtest run from the
// strategy name, instrument, timeframe, date range, commission model, and strategy params.
// Only the parameters relevant to the selected strategy are included in the Parameters map.
func buildRunConfig(stratName, instrument, tf, from, to, commissionStr string, p *strategyParams) output.RunConfig {
	params := strategyParamsMap(stratName, p)
	return output.RunConfig{
		Instrument:      instrument,
		Timeframe:       tf,
		From:            from,
		To:              to,
		Strategy:        stratName,
		CommissionModel: commissionStr,
		Parameters:      params,
	}
}

// strategyParamsMap returns the strategy-specific parameters as a string map for
// embedding in the run metadata. Only the parameters relevant to the named strategy
// are included to avoid cluttering the metadata with irrelevant defaults.
func strategyParamsMap(stratName string, p *strategyParams) map[string]string {
	switch stratName {
	case "sma-crossover":
		return map[string]string{
			"fast_period": fmt.Sprintf("%d", p.fastPeriod),
			"slow_period": fmt.Sprintf("%d", p.slowPeriod),
		}
	case "rsi-mean-reversion":
		return map[string]string{
			"rsi_period": fmt.Sprintf("%d", p.rsiPeriod),
			"oversold":   fmt.Sprintf("%.4g", p.oversold),
			"overbought": fmt.Sprintf("%.4g", p.overbought),
		}
	case "donchian-breakout":
		return map[string]string{
			"donchian_period": fmt.Sprintf("%d", p.donchianPeriod),
		}
	case "macd-crossover":
		return map[string]string{
			"fast_period":   fmt.Sprintf("%d", p.macdFastPeriod),
			"slow_period":   fmt.Sprintf("%d", p.macdSlowPeriod),
			"signal_period": fmt.Sprintf("%d", p.macdSignalPeriod),
		}
	case "bollinger-mean-reversion":
		return map[string]string{
			"bb_period":      fmt.Sprintf("%d", p.bbPeriod),
			"bb_num_std_dev": fmt.Sprintf("%.4g", p.bbNumStdDev),
		}
	case "momentum":
		return map[string]string{
			"lookback":  fmt.Sprintf("%d", p.momentumLookback),
			"threshold": fmt.Sprintf("%.4g", p.momentumThreshold),
		}
	default:
		return nil
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
