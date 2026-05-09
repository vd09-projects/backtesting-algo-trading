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
// Strategy-specific flags (all optional, defaulting to registered defaults):
//
//	--fast-period / --slow-period          (sma-crossover)
//	--rsi-period / --oversold / --overbought (rsi-mean-reversion)
//	--donchian-period                       (donchian-breakout)
//	--macd-fast-period / --macd-slow-period / --macd-signal-period (macd-crossover)
//	--bb-period / --bb-num-std-dev          (bollinger-mean-reversion)
//	--momentum-lookback / --momentum-threshold (momentum)
//	--cci-period / --cci-entry-threshold / --cci-exit-threshold (cci-mean-reversion)
//
// Sizing models:
//
//	fixed      — deploy a fixed fraction of cash per trade (default; controlled by --position-size)
//	vol-target — size each trade so annualized dollar vol = cash × --vol-target (default 10%)
//
// Credentials are read from environment variables (or a .env file in the working
// directory). Token resolution priority:
//
//  1. KITE_ACCESS_TOKEN — use directly; KITE_API_SECRET not required.
//  2. Saved token file at ~/.config/backtest/token.json (or BACKTEST_TOKEN_PATH) — reused when not expired.
//  3. Interactive Kite Connect login flow — requires KITE_API_KEY and KITE_API_SECRET; saves token for reuse.
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
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

type flags struct {
	instrument    string
	fromStr       string
	toStr         string
	tfStr         string
	cash          float64
	stratName     string
	commissionStr string
	outPath       string
	curvePath     string
	sizingModel   string
	volTarget     float64
	gateThreshold float64
	doBootstrap   bool
	bootstrapSeed int64
	bootstrapN    int
	doRegimeGate  bool
}

func main() {
	var f flags
	flag.StringVar(&f.instrument, "instrument", "NSE:NIFTY 50", "Instrument to backtest (e.g. \"NSE:NIFTY 50\", \"NSE:INFY\")")
	flag.StringVar(&f.fromStr, "from", "", "Start date in YYYY-MM-DD (inclusive)")
	flag.StringVar(&f.toStr, "to", "", "End date in YYYY-MM-DD (exclusive)")
	flag.StringVar(&f.tfStr, "timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	flag.Float64Var(&f.cash, "cash", 100000, "Starting cash in ₹")
	flag.StringVar(&f.stratName, "strategy", "stub", "Strategy name: "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), ", "))
	flag.StringVar(&f.commissionStr, "commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")
	flag.StringVar(&f.outPath, "out", "", "Path for JSON results export; when omitted a default name is generated from the run params")
	flag.StringVar(&f.curvePath, "output-curve", "", "Path for equity curve CSV export (omit to skip)")
	flag.StringVar(&f.sizingModel, "sizing-model", "fixed", "Position sizing model: fixed | vol-target")
	flag.Float64Var(&f.volTarget, "vol-target", 0.10, "Annualized volatility target when --sizing-model=vol-target (e.g. 0.10 = 10%)")
	flag.Float64Var(&f.gateThreshold, "proliferation-gate-threshold", 0.0, "Sharpe threshold for proliferation gate PASS/FAIL (0 = disabled; 0.5 recommended for NSE daily)")
	flag.BoolVar(&f.doBootstrap, "bootstrap", false, "Run Monte Carlo bootstrap for Sharpe confidence intervals")
	flag.Int64Var(&f.bootstrapSeed, "bootstrap-seed", 42, "RNG seed for bootstrap (logged with results for reproducibility)")
	flag.IntVar(&f.bootstrapN, "bootstrap-n", 0, "Bootstrap simulation count (0 = default 10,000)")
	flag.BoolVar(&f.doRegimeGate, "regime-gate", false, "Compute per-regime per-trade Sharpe gate using NSE regime windows (2018-2024)")

	// Strategy-specific parameters registered centrally from GlobalRegistry.
	stratParamPtrs := cmdutil.GlobalRegistry.RegisterFlags(flag.CommandLine)

	flag.Parse()

	from, to, tf := parseAndValidateFlags(&f)

	// Validate strategy name early: MustGet panics with a descriptive message if
	// the name is unknown.
	cmdutil.GlobalRegistry.MustGet(f.stratName)

	stratParams := cmdutil.BuildParamMap(stratParamPtrs)

	selectedStrategy, err := cmdutil.GlobalRegistry.Build(f.stratName, tf, stratParams)
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

	var regimeGateReport *analytics.RegimeGateReport
	if f.doRegimeGate {
		r := analytics.ComputeRegimeGate(trades, analytics.NSERegimesGate)
		regimeGateReport = &r
	}

	runCfg := output.RunConfig{
		Instrument:      f.instrument,
		Timeframe:       f.tfStr,
		From:            f.fromStr,
		To:              f.toStr,
		Strategy:        f.stratName,
		CommissionModel: f.commissionStr,
		Parameters:      cmdutil.GlobalRegistry.ParamsMap(f.stratName, stratParams),
	}

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
		RegimeGate:     regimeGateReport,
	}); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}
}

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
