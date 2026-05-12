// cmd/evaluate runs the full end-to-end evaluation pipeline for a strategy:
// (1) universe sweep with DSR gate, (2) walk-forward on survivors,
// (3) bootstrap on walk-forward survivors.
//
// # Usage
//
//	go run ./cmd/evaluate \
//	    --strategy macd-crossover \
//	    --universe universes/nifty50-large-cap.yaml \
//	    --from 2018-01-01 \
//	    --to   2025-01-01 \
//	    --out-dir results/
//
// # Params override
//
// Strategy parameters can be overridden with --params key=value (repeatable).
// Unrecognized keys are passed through; strategies ignore keys they don't use.
// Omitting --params uses strategy defaults.
//
//	--params fast-period=17 --params slow-period=26 --params macd-signal-period=9
//
// # Output layout
//
// All outputs are written to --out-dir/YYYY-MM-DD-{strategy}-{timeframe}/:
//
//	universe-sweep.csv          — per-instrument sweep results
//	wf-{instrument}.json        — walk-forward report per survivor
//	wf-{instrument}-folds.csv   — per-fold CSV per WF survivor
//	bootstrap-{instrument}.json — bootstrap result per WF survivor
//	verdict.json                — pipeline summary
//
// # Pipeline gates (unchanged from existing CLIs)
//
//	Universe gate:      DSRAverageSharpe > 0 AND PassFraction >= 40%
//	Walk-forward gate:  WF passes >= floor(60% x universe_gate_passes) AND >= 6
//	Bootstrap gate:     SharpeP5 > 0 AND P(Sharpe > 0) > 80%
//
// # Credentials
//
// Identical to cmd/backtest: KITE_API_KEY, KITE_ACCESS_TOKEN (or interactive flow).
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/walkforward"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// buildProductionProvider constructs the cached Zerodha provider used in production.
func buildProductionProvider(ctx context.Context) (provider.DataProvider, error) {
	cmdutil.LoadDotEnv(".env")
	return cmdutil.BuildProvider(ctx)
}

// evalFlags holds parsed CLI flags for the evaluate pipeline.
type evalFlags struct {
	stratName     string
	universeFile  string
	fromStr       string
	toStr         string
	tfStr         string
	outDir        string
	commissionStr string
	cash          float64
	positionSize  float64
	slippage      float64
	bootstrapSeed int64
	bootstrapN    int
	isYears       int
	oosYears      int
	stepYears     int
	rawParams     []string
}

// evalPipeline holds validated, parsed state ready for pipeline execution.
type evalPipeline struct {
	flags       evalFlags
	from        time.Time
	to          time.Time
	tf          model.Timeframe
	stratParams map[string]float64
	commModel   model.CommissionModel
	instruments []string
	stageDir    string
	provider    provider.DataProvider
	ctx         context.Context //nolint:containedctx // pipeline carries its context; not a request handler
}

// run is the testable entry point.
//
// **Decision (run() extraction for cmd/evaluate — providerFactory injection for testability) — convention: experimental**
// scope: cmd/evaluate
// tags: testability, run-function, flag-newFlagSet, providerFactory
// owner: priya
//
// All pipeline logic is split across parseEvalFlags, buildPipeline, and runPipeline.
// run() orchestrates them; each function stays within the cyclop limit.
// Pattern follows cmd/walk-forward, cmd/monitor, and cmd/fetch-history precedents.
func run(args []string, stdout, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) error {
	f, err := parseEvalFlags(args, stderr)
	if err != nil {
		return err
	}
	pl, err := buildPipeline(f, stderr, providerFactory)
	if err != nil {
		return err
	}
	return runPipeline(pl, stdout, stderr)
}

// parseEvalFlags parses and validates CLI flags for the evaluate pipeline.
// Returns an error for any missing or invalid flag.
func parseEvalFlags(args []string, stderr io.Writer) (evalFlags, error) {
	fs := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var f evalFlags
	fs.StringVar(&f.stratName, "strategy", "", "Strategy name (required): "+strings.Join(cmdutil.GlobalRegistry.ListStrategies(), ", "))
	fs.StringVar(&f.universeFile, "universe", "", "Path to universe YAML file (required)")
	fs.StringVar(&f.fromStr, "from", "", "Start date YYYY-MM-DD, inclusive (required)")
	fs.StringVar(&f.toStr, "to", "", "End date YYYY-MM-DD, exclusive upper bound (required)")
	fs.StringVar(&f.tfStr, "timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	fs.StringVar(&f.outDir, "out-dir", "", "Root output directory (required)")
	fs.StringVar(&f.commissionStr, "commission", "zerodha", "Commission model: zerodha | zerodha_full | zerodha_full_mis | flat | percentage")
	fs.Float64Var(&f.cash, "cash", 100_000, "Starting cash in Rs per engine run")
	fs.Float64Var(&f.positionSize, "position-size", 0.10, "Fraction of cash deployed per trade")
	fs.Float64Var(&f.slippage, "slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")
	fs.Int64Var(&f.bootstrapSeed, "bootstrap-seed", 42, "RNG seed for bootstrap")
	fs.IntVar(&f.bootstrapN, "bootstrap-n", 0, "Bootstrap simulation count (0 = default 10,000)")
	fs.IntVar(&f.isYears, "is-years", 2, "Walk-forward in-sample window length in years")
	fs.IntVar(&f.oosYears, "oos-years", 1, "Walk-forward out-of-sample window length in years")
	fs.IntVar(&f.stepYears, "step-years", 1, "Walk-forward window step size in years")
	fs.Func("params", "Strategy parameter override: key=value (repeatable)", func(s string) error {
		f.rawParams = append(f.rawParams, s)
		return nil
	})

	if err := fs.Parse(args); err != nil {
		return evalFlags{}, err
	}
	if err := validateRequiredFlags(f); err != nil {
		return evalFlags{}, err
	}
	return f, nil
}

// validateRequiredFlags checks that all required flags are present.
func validateRequiredFlags(f evalFlags) error { //nolint:gocritic // evalFlags is a config struct; by-value is idiomatic for one-shot validation
	switch {
	case f.stratName == "":
		return fmt.Errorf("--strategy is required (%s)", strings.Join(cmdutil.GlobalRegistry.ListStrategies(), " | "))
	case f.universeFile == "":
		return fmt.Errorf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	case f.fromStr == "":
		return fmt.Errorf("--from is required (e.g. 2020-01-01)")
	case f.toStr == "":
		return fmt.Errorf("--to is required (e.g. 2025-01-01)")
	case f.outDir == "":
		return fmt.Errorf("--out-dir is required (e.g. results/)")
	}
	return nil
}

// buildPipeline parses dates, builds the strategy param map, creates the stage
// directory, loads the universe file, and constructs the provider.
func buildPipeline(f evalFlags, stderr io.Writer, providerFactory func(context.Context) (provider.DataProvider, error)) (evalPipeline, error) { //nolint:gocritic // evalFlags is a config struct; by-value is idiomatic for one-shot pipeline construction
	from, err := time.Parse("2006-01-02", f.fromStr)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("--from %q: %w", f.fromStr, err)
	}
	to, err := time.Parse("2006-01-02", f.toStr)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("--to %q: %w", f.toStr, err)
	}
	if !to.After(from) {
		return evalPipeline{}, fmt.Errorf("--to (%s) must be strictly after --from (%s)", f.toStr, f.fromStr)
	}

	tf := model.Timeframe(f.tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		return evalPipeline{}, fmt.Errorf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", f.tfStr)
	}

	// Validate strategy name early — MustGet panics on unknown names (fail-fast).
	cmdutil.GlobalRegistry.MustGet(f.stratName)

	commModel, err := cmdutil.ParseCommissionModel(f.commissionStr)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("--commission: %w", err)
	}

	overrides, err := parseParams(f.rawParams)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("--params: %w", err)
	}
	stratParams := cmdutil.GlobalRegistry.DefaultParams(f.stratName)
	if stratParams == nil {
		stratParams = make(map[string]float64)
	}
	for k, v := range overrides {
		stratParams[k] = v
	}

	instruments, err := universesweep.ParseUniverseFile(f.universeFile)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("universe file: %w", err)
	}

	stageDir, err := makeStageDir(f.outDir, f.stratName, f.tfStr)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("create stage directory: %w", err)
	}
	fmt.Fprintf(stderr, "evaluate: stage directory: %s\n", stageDir) //nolint:errcheck // progress to stderr

	ctx := context.Background()
	p, err := providerFactory(ctx)
	if err != nil {
		return evalPipeline{}, fmt.Errorf("provider: %w", err)
	}

	return evalPipeline{
		flags:       f,
		from:        from,
		to:          to,
		tf:          tf,
		stratParams: stratParams,
		commModel:   commModel,
		instruments: instruments,
		stageDir:    stageDir,
		provider:    p,
		ctx:         ctx,
	}, nil
}

// runPipeline executes the three pipeline stages sequentially.
func runPipeline(pl evalPipeline, stdout, stderr io.Writer) error { //nolint:gocritic // evalPipeline is a pipeline config struct; by-value is idiomatic
	_ = stdout // not used; progress goes to stderr

	// Stage 1: Universe sweep.
	gateResult, universeSurvivors, kills, instrumentSharpe, err := runUniverseSweep(pl, stderr)
	if err != nil {
		return err
	}
	if !gateResult.GatePass || len(universeSurvivors) == 0 {
		v := verdictJSON{
			Result:    "killed_at_universe_gate",
			Strategy:  pl.flags.stratName,
			From:      pl.flags.fromStr,
			To:        pl.flags.toStr,
			Timeframe: pl.flags.tfStr,
			GateStats: newGateStats(gateResult),
			Kills:     kills,
		}
		if werr := writeVerdictJSON(pl.stageDir, &v); werr != nil {
			return fmt.Errorf("write verdict.json: %w", werr)
		}
		fmt.Fprintf(stderr, "evaluate: KILLED at universe gate — verdict.json written\n") //nolint:errcheck // progress
		return nil
	}
	fmt.Fprintf(stderr, "evaluate: universe gate PASS — %d/%d instruments advance to walk-forward\n", //nolint:errcheck // progress
		len(universeSurvivors), len(pl.instruments))

	// Stage 2: Walk-forward.
	wfPassed, newKills, err := runWalkForward(pl, universeSurvivors, stderr)
	kills = append(kills, newKills...)
	if err != nil {
		return err
	}

	wfGatePassed, wfGateFailed := applyWFGate(len(universeSurvivors), wfPassed, wfGateFailed2Kills(universeSurvivors, wfPassed))
	if len(wfGatePassed) == 0 {
		for _, inst := range wfGateFailed {
			kills = append(kills, killRecord{
				Instrument: inst,
				Stage:      "walk_forward_gate",
				Reason:     fmt.Sprintf("WFGateFail(passed=%d,required=%d)", len(wfPassed), int(math.Floor(0.60*float64(len(universeSurvivors))))),
			})
		}
		v := verdictJSON{
			Result:    "killed_at_walk_forward_gate",
			Strategy:  pl.flags.stratName,
			From:      pl.flags.fromStr,
			To:        pl.flags.toStr,
			Timeframe: pl.flags.tfStr,
			GateStats: newGateStats(gateResult),
			Kills:     kills,
		}
		if werr := writeVerdictJSON(pl.stageDir, &v); werr != nil {
			return fmt.Errorf("write verdict.json: %w", werr)
		}
		fmt.Fprintf(stderr, "evaluate: KILLED at walk-forward gate — %d/%d passed WF\n", //nolint:errcheck // progress
			len(wfPassed), len(universeSurvivors))
		return nil
	}
	fmt.Fprintf(stderr, "evaluate: WF gate PASS — %d instruments advance to bootstrap\n", len(wfGatePassed)) //nolint:errcheck // progress

	// Stage 3: Bootstrap.
	survivors, bootstrapKills, err := runBootstrap(pl, wfGatePassed, instrumentSharpe, stderr)
	kills = append(kills, bootstrapKills...)
	if err != nil {
		return err
	}

	result := "pass"
	if len(survivors) == 0 {
		result = "killed_at_bootstrap"
	}
	v := verdictJSON{
		Result:    result,
		Strategy:  pl.flags.stratName,
		From:      pl.flags.fromStr,
		To:        pl.flags.toStr,
		Timeframe: pl.flags.tfStr,
		GateStats: newGateStats(gateResult),
		Survivors: survivors,
		Kills:     kills,
	}
	if werr := writeVerdictJSON(pl.stageDir, &v); werr != nil {
		return fmt.Errorf("write verdict.json: %w", werr)
	}
	fmt.Fprintf(stderr, "evaluate: pipeline complete — result=%s survivors=%d kills=%d\n", //nolint:errcheck // progress
		result, len(survivors), len(kills))
	return nil
}

// wfGateFailed2Kills returns instruments in universeSurvivors that are not in wfPassed.
func wfGateFailed2Kills(universeSurvivors, wfPassed []string) []string {
	passed := make(map[string]struct{}, len(wfPassed))
	for _, inst := range wfPassed {
		passed[inst] = struct{}{}
	}
	var failed []string
	for _, inst := range universeSurvivors {
		if _, ok := passed[inst]; !ok {
			failed = append(failed, inst)
		}
	}
	return failed
}

// runUniverseSweep executes the universe sweep stage and applies the gate.
// Returns the gate result, surviving instrument names, kills list, and a
// per-instrument Sharpe map (used later for survivor records).
func runUniverseSweep(pl evalPipeline, stderr io.Writer) ( //nolint:gocritic // evalPipeline is a pipeline config struct; by-value is idiomatic
	universesweep.GateResult,
	[]string,
	[]killRecord,
	map[string]float64,
	error,
) {
	fmt.Fprintf(stderr, "evaluate: [1/3] universe sweep — strategy=%s instruments=%d from=%s to=%s\n", //nolint:errcheck // progress
		pl.flags.stratName, len(pl.instruments), pl.flags.fromStr, pl.flags.toStr)

	// WalkForwardFactory validates params eagerly and returns a fresh-instance factory.
	// Each instrument gets its own strategy state — no bleed between runs.
	sweepStratFactory, err := cmdutil.GlobalRegistry.WalkForwardFactory(pl.flags.stratName, pl.tf, pl.stratParams)
	if err != nil {
		return universesweep.GateResult{}, nil, nil, nil, fmt.Errorf("--strategy: %w", err)
	}

	sweepCfg := universesweep.Config{
		Instruments: pl.instruments,
		NewStrategy: sweepStratFactory,
		EngineConfig: engine.Config{
			From:                 pl.from,
			To:                   pl.to,
			InitialCash:          pl.flags.cash,
			PositionSizeFraction: pl.flags.positionSize,
			OrderConfig: model.OrderConfig{
				SlippagePct:     pl.flags.slippage,
				CommissionModel: pl.commModel,
			},
		},
		Timeframe: pl.tf,
	}

	sweepReport, err := universesweep.Run(pl.ctx, &sweepCfg, pl.provider, stderr)
	if err != nil {
		return universesweep.GateResult{}, nil, nil, nil, fmt.Errorf("universe sweep: %w", err)
	}

	if werr := writeUniverseSweepCSV(pl.stageDir, sweepReport); werr != nil {
		return universesweep.GateResult{}, nil, nil, nil, fmt.Errorf("write universe-sweep.csv: %w", werr)
	}

	gateResult := universesweep.ApplyUniverseGate(sweepReport, len(pl.instruments))
	fmt.Fprintf(stderr, "evaluate: universe gate — DSRAvg=%.4f PassFraction=%.1f%% GatePass=%v\n", //nolint:errcheck // progress
		gateResult.DSRAverageSharpe, gateResult.PassFraction*100, gateResult.GatePass)

	var kills []killRecord
	var survivors []string
	sharpeMap := make(map[string]float64, len(sweepReport.Results))

	for _, r := range sweepReport.Results {
		sharpeMap[r.Instrument] = r.Sharpe
		switch {
		case r.InsufficientData:
			kills = append(kills, killRecord{Instrument: r.Instrument, Stage: "universe_gate", Reason: "InsufficientData"})
		case r.Sharpe <= 0:
			kills = append(kills, killRecord{Instrument: r.Instrument, Stage: "universe_gate", Reason: fmt.Sprintf("NegativeSharpe(%.4f)", r.Sharpe)})
		default:
			survivors = append(survivors, r.Instrument)
		}
	}
	return gateResult, survivors, kills, sharpeMap, nil
}

// runWalkForward runs walk-forward validation per survivor and writes stage outputs.
// Returns the list of WF-passed instruments and a kills list for WF-failed ones.
//
// **Decision (WF dispatch sequential per instrument) — tradeoff: experimental**
// scope: cmd/evaluate
// tags: concurrency, walk-forward, orchestrator
// owner: priya
//
// Each walkforward.Run fans out folds internally via errgroup. Adding outer
// parallelism doubles goroutines without halving runtime on CPU-bound folds
// and interleaves progress logs. Sequential loop matches cmd/walk-forward.
func runWalkForward(pl evalPipeline, instruments []string, stderr io.Writer) ([]string, []killRecord, error) { //nolint:gocritic // evalPipeline is a pipeline config struct; by-value is idiomatic
	fmt.Fprintf(stderr, "evaluate: [2/3] walk-forward — %d instruments\n", len(instruments)) //nolint:errcheck // progress

	year := 365 * 24 * time.Hour
	baseCfg := walkforward.EngineConfigTemplate{
		InitialCash:          pl.flags.cash,
		PositionSizeFraction: pl.flags.positionSize,
		OrderConfig: model.OrderConfig{
			SlippagePct:     pl.flags.slippage,
			CommissionModel: pl.commModel,
		},
	}

	var passed []string
	var kills []killRecord

	for _, inst := range instruments {
		factory, err := cmdutil.GlobalRegistry.WalkForwardFactory(pl.flags.stratName, pl.tf, pl.stratParams)
		if err != nil {
			return nil, nil, fmt.Errorf("walk-forward factory for %s: %w", inst, err)
		}
		wfCfg := walkforward.WalkForwardConfig{
			Instrument:        inst,
			From:              pl.from,
			To:                pl.to,
			InSampleWindow:    time.Duration(pl.flags.isYears) * year,
			OutOfSampleWindow: time.Duration(pl.flags.oosYears) * year,
			StepSize:          time.Duration(pl.flags.stepYears) * year,
		}
		report, err := walkforward.Run(pl.ctx, wfCfg, baseCfg, pl.provider, factory)
		if err != nil {
			return nil, nil, fmt.Errorf("walk-forward %s: %w", inst, err)
		}
		if werr := writeWFJSON(pl.stageDir, inst, report); werr != nil {
			return nil, nil, fmt.Errorf("write wf JSON for %s: %w", inst, werr)
		}
		if werr := writeWFFoldsCSV(pl.stageDir, inst, report.Windows); werr != nil {
			return nil, nil, fmt.Errorf("write wf folds CSV for %s: %w", inst, werr)
		}
		fmt.Fprintf(stderr, "evaluate: WF %s — OverfitFlag=%v NegativeFoldFlag=%v\n", //nolint:errcheck // progress
			inst, report.OverfitFlag, report.NegativeFoldFlag)
		if !report.OverfitFlag && !report.NegativeFoldFlag {
			passed = append(passed, inst)
		} else {
			kills = append(kills, killRecord{Instrument: inst, Stage: "walk_forward", Reason: wfKillReason(report)})
		}
	}
	return passed, kills, nil
}

// wfKillReason derives a reason string from a walk-forward report.
func wfKillReason(r walkforward.Report) string {
	switch {
	case r.OverfitFlag && r.NegativeFoldFlag:
		return "OverfitFlag+NegativeFoldFlag"
	case r.OverfitFlag:
		return "OverfitFlag"
	default:
		return "NegativeFoldFlag"
	}
}

// runBootstrap runs bootstrap on each WF-passed instrument and writes stage outputs.
// Returns survivors that pass the bootstrap gate and kills for those that do not.
func runBootstrap(pl evalPipeline, instruments []string, sharpeMap map[string]float64, stderr io.Writer) ([]survivorRecord, []killRecord, error) { //nolint:gocritic // evalPipeline is a pipeline config struct; by-value is idiomatic
	fmt.Fprintf(stderr, "evaluate: [3/3] bootstrap — %d instruments\n", len(instruments)) //nolint:errcheck // progress

	var survivors []survivorRecord
	var kills []killRecord

	for _, inst := range instruments {
		trades, err := collectTrades(pl, inst)
		if err != nil {
			return nil, nil, err
		}
		bsResult := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{
			NSimulations: pl.flags.bootstrapN,
			Seed:         pl.flags.bootstrapSeed,
		})
		if werr := writeBootstrapJSON(pl.stageDir, inst, bsResult); werr != nil {
			return nil, nil, fmt.Errorf("write bootstrap JSON for %s: %w", inst, werr)
		}
		gatePassed := applyBootstrapGate(bsResult.SharpeP5, bsResult.ProbPositiveSharpe)
		fmt.Fprintf(stderr, "evaluate: bootstrap %s — SharpeP5=%.4f ProbPositive=%.3f pass=%v\n", //nolint:errcheck // progress
			inst, bsResult.SharpeP5, bsResult.ProbPositiveSharpe, gatePassed)
		if gatePassed {
			survivors = append(survivors, survivorRecord{
				Instrument:            inst,
				UniverseSharpe:        sharpeMap[inst],
				WFPassed:              true,
				BootstrapP5:           bsResult.SharpeP5,
				BootstrapProbPositive: bsResult.ProbPositiveSharpe,
			})
		} else {
			kills = append(kills, killRecord{
				Instrument: inst,
				Stage:      "bootstrap",
				Reason:     fmt.Sprintf("BootstrapFail(P5=%.4f,ProbPos=%.3f)", bsResult.SharpeP5, bsResult.ProbPositiveSharpe),
			})
		}
	}
	return survivors, kills, nil
}

// collectTrades runs the engine over the full date range for a single instrument
// and returns its closed trades for use in bootstrap resampling.
func collectTrades(pl evalPipeline, instrument string) ([]model.Trade, error) { //nolint:gocritic // evalPipeline is a pipeline config struct; by-value is idiomatic
	s, err := cmdutil.GlobalRegistry.Build(pl.flags.stratName, pl.tf, pl.stratParams)
	if err != nil {
		return nil, fmt.Errorf("build strategy for bootstrap %s: %w", instrument, err)
	}
	eng := engine.New(engine.Config{
		Instrument:           instrument,
		From:                 pl.from,
		To:                   pl.to,
		InitialCash:          pl.flags.cash,
		PositionSizeFraction: pl.flags.positionSize,
		OrderConfig: model.OrderConfig{
			SlippagePct:     pl.flags.slippage,
			CommissionModel: pl.commModel,
		},
	})
	if err := eng.Run(pl.ctx, pl.provider, s); err != nil {
		return nil, fmt.Errorf("engine run for bootstrap %s: %w", instrument, err)
	}
	return eng.Portfolio().ClosedTrades(), nil
}

// ---------------------------------------------------------------------------
// verdict.json DTO types
//
// **Decision (verdict.json as local DTO in cmd/evaluate — not internal/output) — architecture: experimental**
// scope: cmd/evaluate
// tags: serialization-dto, cmd-layer, JSON
// owner: priya
//
// The verdict.json is a pipeline-level aggregation of gate results across stages.
// It is not a per-backtest result (internal/output's domain). Keeping it local
// to cmd/evaluate follows the thresholdsFile DTO precedent in cmd/monitor.
// ---------------------------------------------------------------------------

type verdictJSON struct {
	Result    string           `json:"result"`
	Strategy  string           `json:"strategy"`
	From      string           `json:"from"`
	To        string           `json:"to"`
	Timeframe string           `json:"timeframe,omitempty"`
	GateStats *gateStatsJSON   `json:"gate_stats,omitempty"`
	Survivors []survivorRecord `json:"survivors,omitempty"`
	Kills     []killRecord     `json:"kills,omitempty"`
}

type gateStatsJSON struct {
	DSRAverageSharpe          float64 `json:"dsr_average_sharpe"`
	SufficientInstruments     int     `json:"sufficient_instruments"`
	PositiveSharpeInstruments int     `json:"positive_sharpe_instruments"`
	PassFraction              float64 `json:"pass_fraction"`
}

type survivorRecord struct {
	Instrument            string  `json:"instrument"`
	UniverseSharpe        float64 `json:"universe_sharpe"`
	WFPassed              bool    `json:"wf_passed"`
	BootstrapP5           float64 `json:"bootstrap_sharpe_p5"`
	BootstrapProbPositive float64 `json:"bootstrap_prob_positive_sharpe"`
}

type killRecord struct {
	Instrument string `json:"instrument"`
	Stage      string `json:"stage"`
	Reason     string `json:"reason"`
}

// newGateStats converts a universesweep.GateResult to the local DTO.
func newGateStats(gr universesweep.GateResult) *gateStatsJSON {
	return &gateStatsJSON{
		DSRAverageSharpe:          gr.DSRAverageSharpe,
		SufficientInstruments:     gr.SufficientInstruments,
		PositiveSharpeInstruments: gr.PositiveSharpeInstruments,
		PassFraction:              gr.PassFraction,
	}
}

// ---------------------------------------------------------------------------
// parseParams parses --params key=value flag values into a float64 map.
// ---------------------------------------------------------------------------

// parseParams converts "key=value" strings to a float64 param map.
// Returns an error if any entry is malformed or the value is non-numeric.
func parseParams(raw []string) (map[string]float64, error) {
	m := make(map[string]float64, len(raw))
	for _, s := range raw {
		key, valStr, ok := strings.Cut(s, "=")
		if !ok {
			return nil, fmt.Errorf("invalid param %q — expected key=value format", s)
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64)
		if err != nil {
			return nil, fmt.Errorf("param %q: value %q is not a number: %w", key, valStr, err)
		}
		m[strings.TrimSpace(key)] = v
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// makeStageDir creates and returns the dated stage output directory.
// ---------------------------------------------------------------------------

// makeStageDir creates outDir/YYYY-MM-DD-{strategy}-{timeframe}/ and returns its path.
// Returns an error if the parent outDir does not exist.
//
// **Decision (all stage outputs in one dated subdirectory) — architecture: experimental**
// scope: cmd/evaluate
// tags: output-layout, stage-dir
// owner: priya
//
// Writing everything under a dated subdirectory keeps the root outDir clean
// across multiple runs of the same strategy, enabling side-by-side comparison.
// YYYY-MM-DD prefix gives natural sort order.
func makeStageDir(outDir, stratName, tfStr string) (string, error) {
	today := time.Now().UTC().Format("2006-01-02")
	safeName := strings.NewReplacer(":", "_", " ", "_", "/", "_").Replace(stratName)
	dirName := fmt.Sprintf("%s-%s-%s", today, safeName, tfStr)
	stageDir := filepath.Join(outDir, dirName)
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %q: %w", stageDir, err)
	}
	return stageDir, nil
}

// ---------------------------------------------------------------------------
// Gate helpers
// ---------------------------------------------------------------------------

// applyWFGate applies the walk-forward instrument-count gate:
// WF_passes >= floor(0.60 * universeGatePasses) AND WF_passes >= 6.
//
// Returns (wfPassed, allFailed):
//   - wfPassed contains passed if the gate succeeds, nil if it fails.
//   - allFailed contains all instruments (passed + failed) if the gate fails.
func applyWFGate(universeGatePasses int, passed, failed []string) (wfPassed, allFailed []string) {
	required := int(math.Floor(0.60 * float64(universeGatePasses)))
	if required < 6 {
		required = 6
	}
	if len(passed) >= required {
		return passed, failed
	}
	all := make([]string, 0, len(passed)+len(failed))
	all = append(all, passed...)
	all = append(all, failed...)
	return nil, all
}

// applyBootstrapGate returns true when the bootstrap gate passes:
// SharpeP5 > 0 AND ProbPositiveSharpe > 0.80.
func applyBootstrapGate(sharpeP5, probPositiveSharpe float64) bool {
	return sharpeP5 > 0 && probPositiveSharpe > 0.80
}

// ---------------------------------------------------------------------------
// Write helpers
// ---------------------------------------------------------------------------

// writeVerdictJSON serializes v to {dir}/verdict.json.
func writeVerdictJSON(dir string, v *verdictJSON) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal verdict: %w", err)
	}
	path := filepath.Join(dir, "verdict.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// writeUniverseSweepCSV writes the sweep report as universe-sweep.csv in dir.
func writeUniverseSweepCSV(dir string, report universesweep.Report) error {
	path := filepath.Join(dir, "universe-sweep.csv")
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	writeErr := universesweep.WriteCSV(f, report)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", path, closeErr)
	}
	return nil
}

// writeWFJSON writes the walk-forward report as wf-{instrument}.json in dir.
func writeWFJSON(dir, instrument string, report walkforward.Report) error {
	safe := sanitizeInstrument(instrument)
	path := filepath.Join(dir, fmt.Sprintf("wf-%s.json", safe))
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal WF report for %s: %w", instrument, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// writeWFFoldsCSV writes a fold-level CSV as wf-{instrument}-folds.csv in dir.
func writeWFFoldsCSV(dir, instrument string, windows []walkforward.WindowResult) error {
	safe := sanitizeInstrument(instrument)
	path := filepath.Join(dir, fmt.Sprintf("wf-%s-folds.csv", safe))
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	writeErr := writeFoldsCSV(f, windows)
	closeErr := f.Close()
	if writeErr != nil {
		return writeErr
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", path, closeErr)
	}
	return nil
}

// writeFoldsCSV writes a fold-level CSV matching cmd/walk-forward format.
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

// bootstrapDTO is the local serialization type for bootstrap results.
// montecarlo.BootstrapResult has no JSON tags; this DTO provides them without
// adding a serialization dependency to the montecarlo package.
type bootstrapDTO struct {
	MeanSharpe         float64 `json:"mean_sharpe"`
	SharpeP5           float64 `json:"sharpe_p5"`
	SharpeP50          float64 `json:"sharpe_p50"`
	SharpeP95          float64 `json:"sharpe_p95"`
	WorstDrawdownP5    float64 `json:"worst_drawdown_p5"`
	WorstDrawdownP50   float64 `json:"worst_drawdown_p50"`
	WorstDrawdownP95   float64 `json:"worst_drawdown_p95"`
	ProbPositiveSharpe float64 `json:"prob_positive_sharpe"`
}

// writeBootstrapJSON writes bootstrap-{instrument}.json in dir.
func writeBootstrapJSON(dir, instrument string, result montecarlo.BootstrapResult) error {
	safe := sanitizeInstrument(instrument)
	path := filepath.Join(dir, fmt.Sprintf("bootstrap-%s.json", safe))
	dto := bootstrapDTO{
		MeanSharpe:         result.MeanSharpe,
		SharpeP5:           result.SharpeP5,
		SharpeP50:          result.SharpeP50,
		SharpeP95:          result.SharpeP95,
		WorstDrawdownP5:    result.WorstDrawdownP5,
		WorstDrawdownP50:   result.WorstDrawdownP50,
		WorstDrawdownP95:   result.WorstDrawdownP95,
		ProbPositiveSharpe: result.ProbPositiveSharpe,
	}
	data, err := json.MarshalIndent(dto, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bootstrap for %s: %w", instrument, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

// sanitizeInstrument replaces filesystem-unsafe characters in instrument names.
func sanitizeInstrument(instrument string) string {
	return strings.NewReplacer(":", "_", " ", "_", "/", "_").Replace(instrument)
}
