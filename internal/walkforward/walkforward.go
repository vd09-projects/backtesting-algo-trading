// Package walkforward implements a fixed rolling walk-forward validation harness.
//
// Walk-forward is a regime-stability test for stateless fixed-parameter strategies:
// it checks whether a strategy's OOS Sharpe degrades significantly versus its IS Sharpe
// across multiple folds — a signal of overfitting to a specific regime rather than
// robust edge. It is NOT a parameter-optimisation test.
//
// # Window structure
//
// generateWindows steps from cfg.From by cfg.StepSize. Each fold is:
//
//	IS window:  [stepStart, stepStart + InSampleWindow)
//	OOS window: [stepStart + InSampleWindow, stepStart + InSampleWindow + OutOfSampleWindow)
//
// A fold is excluded when its OOS end exceeds cfg.To.
//
// # Degenerate windows
//
// A fold with zero OOS closed trades is marked Degenerate=true and excluded from
// all scoring (averages, flag checks, DeduplicatedFoldCount). "No trades" ≠ overfitting.
//
// # Scoring
//
// OverfitFlag:     avg OOS Sharpe < 50% of avg IS Sharpe (over non-degenerate folds).
// NegativeFoldFlag: 2+ non-degenerate folds with OOS Sharpe < 0.
//
// # Concurrency model
//
// Sequential inside a run, parallel across folds. Each fold spawns two engine.New()
// calls (IS + OOS) via errgroup. Results are written into a pre-allocated slice at
// fixed indices to preserve fold ordering regardless of goroutine scheduling.
// See go-patterns.md §"Sequential inside, parallel across".
package walkforward

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// WalkForwardConfig holds walk-forward harness parameters.
// It is separate from engine.Config: the harness stamps From/To/Instrument onto
// a base engine.Config template per fold; the caller is not responsible for that stamping.
//
//nolint:revive // acceptance criteria mandates this name; walkforward.Config would be less clear
type WalkForwardConfig struct {
	InSampleWindow    time.Duration // duration of each in-sample period
	OutOfSampleWindow time.Duration // duration of each out-of-sample period
	StepSize          time.Duration // how far to advance the window each fold
	Instrument        string        // instrument identifier passed to provider and engine
	From              time.Time     // start of the overall evaluation range (UTC)
	To                time.Time     // end of the overall evaluation range (UTC)
}

// EngineConfigTemplate holds engine configuration fields that are constant across
// all folds. Run stamps Instrument, From, and To onto a fresh engine.Config per fold.
//
// **Decision (EngineConfigTemplate vs passing engine.Config directly) — tradeoff: experimental**
// scope: internal/walkforward
// tags: API, engine.Config, walk-forward
// Passing a full engine.Config and overwriting From/To/Instrument per fold would work
// but makes it implicit which fields the harness controls vs. the caller. A separate
// template type makes the boundary explicit: the harness owns From/To/Instrument;
// the caller owns cost model and sizing. Prevents accidental bugs where a caller sets
// cfg.From/To thinking they're the outer window bounds, not fold bounds.
type EngineConfigTemplate struct {
	InitialCash          float64
	OrderConfig          model.OrderConfig
	PositionSizeFraction float64
	SizingModel          model.SizingModel
	VolatilityTarget     float64
}

// WindowResult holds the in-sample and out-of-sample results for a single fold.
type WindowResult struct {
	// Period boundaries.
	InSampleStart    time.Time
	InSampleEnd      time.Time
	OutOfSampleStart time.Time
	OutOfSampleEnd   time.Time

	// Sharpe metrics (per-trade non-annualized, sample variance n-1).
	InSampleSharpe    float64
	OutOfSampleSharpe float64

	// TradeCount is the number of OOS closed trades.
	TradeCount int

	// Degenerate is true when TradeCount == 0. Degenerate folds are excluded
	// from scoring and do not affect any flag or average.
	Degenerate bool
}

// Report is the aggregate result of a walk-forward run.
type Report struct {
	// Windows contains all folds, including degenerate ones.
	Windows []WindowResult

	// Averages and counts computed over non-degenerate folds only.
	AvgInSampleSharpe     float64
	AvgOutOfSampleSharpe  float64
	DeduplicatedFoldCount int // non-degenerate folds that contributed to averages

	// Overfit flag: avg OOS Sharpe < 50% of avg IS Sharpe.
	OverfitFlag bool

	// NegativeFoldFlag: 2+ non-degenerate folds with OOS Sharpe < 0.
	NegativeFoldFlag  bool
	NegativeFoldCount int // count of non-degenerate folds with OOS Sharpe < 0
}

// Run executes the walk-forward harness.
//
// It generates fold windows via generateWindows, runs each fold's IS and OOS
// engine calls in parallel (via errgroup), assembles WindowResults, and scores
// the full set via scoreFolds.
//
// The strategy and provider are called concurrently across folds; callers must
// ensure that the provided implementations are safe for concurrent use, or pass
// distinct instances per fold. The fakes in walkforward_test.go are stateless
// (staticProvider, neverTradeStrategy) or have only fold-local state. If a
// strategy carries mutable state between Next() calls (e.g., a history buffer)
// the caller is responsible for providing a fresh instance — or wrapping in a
// factory. For now the API takes a single instance; if stateful strategies become
// common, the signature should change to accept a factory func.
//
// **Decision (strategy concurrency — single instance vs factory) — tradeoff: experimental**
// scope: internal/walkforward
// tags: strategy, concurrency, API
// Taking a single strategy.Strategy instance keeps the API simple for the current
// all-stateless-strategy set. The approved plan specified a single instance. Revisit
// if mutable-state strategies are added; at that point the signature changes to
// func() strategy.Strategy (factory) so each fold gets a fresh copy.
func Run( //nolint:gocritic // WalkForwardConfig and EngineConfigTemplate are config structs; pointer would complicate the call site for no hot-loop benefit
	ctx context.Context,
	cfg WalkForwardConfig, //nolint:gocritic // config struct; pointer API would be awkward for a one-shot call
	baseCfg EngineConfigTemplate,
	p provider.DataProvider,
	s strategy.Strategy,
) (Report, error) {
	if cfg.Instrument == "" {
		return Report{}, fmt.Errorf("walkforward: instrument must not be empty")
	}
	if cfg.From.IsZero() || cfg.To.IsZero() {
		return Report{}, fmt.Errorf("walkforward: From and To must be set")
	}
	if !cfg.To.After(cfg.From) {
		return Report{}, fmt.Errorf("walkforward: To (%s) must be after From (%s)", cfg.To, cfg.From)
	}
	if cfg.InSampleWindow <= 0 {
		return Report{}, fmt.Errorf("walkforward: InSampleWindow must be positive")
	}
	if cfg.OutOfSampleWindow <= 0 {
		return Report{}, fmt.Errorf("walkforward: OutOfSampleWindow must be positive")
	}
	if cfg.StepSize <= 0 {
		return Report{}, fmt.Errorf("walkforward: StepSize must be positive")
	}

	windows := generateWindows(&cfg)
	if len(windows) == 0 {
		return Report{Windows: []WindowResult{}}, nil
	}

	results := make([]WindowResult, len(windows))

	// **Decision (errgroup parallelism ceiling = GOMAXPROCS) — convention: experimental**
	// scope: internal/walkforward
	// tags: concurrency, errgroup, parallelism
	// Cap to GOMAXPROCS to bound memory (each fold runs two full engine passes, each
	// holding a full candle series in memory). For the default 2y IS + 1y OOS on daily
	// bars this is ~730+365 candles per fold, trivially small. The ceiling is a guard
	// against future high-frequency timeframes where candle series could be large.
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	for i := range windows {
		i := i // capture loop variable
		g.Go(func() error {
			wr, err := runFold(gctx, windows[i], cfg.Instrument, baseCfg, p, s)
			if err != nil {
				return fmt.Errorf("walkforward: fold %d: %w", i, err)
			}
			results[i] = wr
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Report{}, err
	}

	report := scoreFolds(results)
	return report, nil
}

// generateWindows produces the ordered list of WindowResults (time fields only populated)
// for cfg. A fold is included only when its OOS end does not exceed cfg.To.
// Windows step by cfg.StepSize from cfg.From.
func generateWindows(cfg *WalkForwardConfig) []WindowResult { //nolint:gocritic // pointer avoids hugeParam copy on internal call
	var windows []WindowResult
	for start := cfg.From; ; start = start.Add(cfg.StepSize) {
		isEnd := start.Add(cfg.InSampleWindow)
		oosEnd := isEnd.Add(cfg.OutOfSampleWindow)
		if oosEnd.After(cfg.To) {
			break
		}
		windows = append(windows, WindowResult{
			InSampleStart:    start,
			InSampleEnd:      isEnd,
			OutOfSampleStart: isEnd,
			OutOfSampleEnd:   oosEnd,
		})
	}
	return windows
}

// runFold executes the IS and OOS engine runs for a single fold.
// IS and OOS runs execute sequentially (not in parallel with each other) because
// "sequential inside, parallel across" — within a fold the two runs are ordered
// work units, not independent sub-runs that need separate goroutines.
//
// **Decision (IS and OOS within a fold: sequential not parallel) — convention: experimental**
// scope: internal/walkforward/runFold
// tags: concurrency, fold-internal-sequencing
// IS and OOS within a fold could be parallelised (they're independent). But each
// fold goroutine is already one unit of parallelism in the outer errgroup; adding
// a second goroutine inside each fold goroutine would double the concurrency without
// halving the runtime (Go folds are CPU-bound on candle processing, not I/O-bound).
// Sequential inside the fold keeps the code simple and avoids nested errgroups.
func runFold( //nolint:gocritic // WindowResult and EngineConfigTemplate pass by value intentionally; not in a hot loop
	ctx context.Context,
	w WindowResult, //nolint:gocritic // pre-populated time fields; not in a hot loop — copy is fine
	instrument string,
	baseCfg EngineConfigTemplate,
	p provider.DataProvider,
	s strategy.Strategy,
) (WindowResult, error) {
	isCfg := engine.Config{
		Instrument:           instrument,
		From:                 w.InSampleStart,
		To:                   w.InSampleEnd,
		InitialCash:          baseCfg.InitialCash,
		OrderConfig:          baseCfg.OrderConfig,
		PositionSizeFraction: baseCfg.PositionSizeFraction,
		SizingModel:          baseCfg.SizingModel,
		VolatilityTarget:     baseCfg.VolatilityTarget,
	}
	oosCfg := engine.Config{
		Instrument:           instrument,
		From:                 w.OutOfSampleStart,
		To:                   w.OutOfSampleEnd,
		InitialCash:          baseCfg.InitialCash,
		OrderConfig:          baseCfg.OrderConfig,
		PositionSizeFraction: baseCfg.PositionSizeFraction,
		SizingModel:          baseCfg.SizingModel,
		VolatilityTarget:     baseCfg.VolatilityTarget,
	}

	isEngine := engine.New(isCfg)
	if err := isEngine.Run(ctx, p, s); err != nil {
		return WindowResult{}, fmt.Errorf("IS run: %w", err)
	}
	isTrades := isEngine.Portfolio().ClosedTrades()
	isSharpe := perTradeSharpe(isTrades)

	oosEngine := engine.New(oosCfg)
	if err := oosEngine.Run(ctx, p, s); err != nil {
		return WindowResult{}, fmt.Errorf("OOS run: %w", err)
	}
	oosTrades := oosEngine.Portfolio().ClosedTrades()
	oosSharpe := perTradeSharpe(oosTrades)

	w.InSampleSharpe = isSharpe
	w.OutOfSampleSharpe = oosSharpe
	w.TradeCount = len(oosTrades)
	w.Degenerate = len(oosTrades) == 0
	return w, nil
}

// scoreFolds computes aggregate statistics and sets flags on a slice of WindowResults.
// Degenerate windows (zero OOS trades) are excluded from all averages and flag checks.
// scoreFolds is exported at the package level (lower-case) — it is unexported but
// testable because the test file is in the same package.
func scoreFolds(windows []WindowResult) Report {
	report := Report{
		Windows: windows,
	}

	var sumIS, sumOOS float64
	var nonDegenCount int

	for i := range windows {
		if windows[i].Degenerate {
			continue
		}
		sumIS += windows[i].InSampleSharpe
		sumOOS += windows[i].OutOfSampleSharpe
		nonDegenCount++
		if windows[i].OutOfSampleSharpe < 0 {
			report.NegativeFoldCount++
		}
	}

	report.DeduplicatedFoldCount = nonDegenCount
	if nonDegenCount == 0 {
		// All degenerate — no scoring possible.
		return report
	}

	report.AvgInSampleSharpe = sumIS / float64(nonDegenCount)
	report.AvgOutOfSampleSharpe = sumOOS / float64(nonDegenCount)

	// OverfitFlag: avg OOS < 50% of avg IS.
	// Uses absolute IS as reference: if IS is negative we still apply the 50% rule
	// (OOS must be >= 0.5 * IS). For negative IS, 0.5*IS < IS, so the flag fires
	// when OOS is even worse (more negative) than 50% of an already-negative IS.
	report.OverfitFlag = report.AvgOutOfSampleSharpe < 0.5*report.AvgInSampleSharpe

	report.NegativeFoldFlag = report.NegativeFoldCount >= 2

	return report
}

// perTradeSharpe computes the non-annualized per-trade Sharpe ratio from a slice
// of closed trades. Formula: mean(ReturnOnNotional) / stddev(ReturnOnNotional),
// sample variance (n-1 denominator). Returns 0.0 for fewer than 2 trades or zero variance.
// This replicates sampleSharpe from internal/montecarlo for consistency.
func perTradeSharpe(trades []model.Trade) float64 {
	if len(trades) < 2 {
		return 0.0
	}

	returns := make([]float64, len(trades))
	for i, t := range trades {
		returns[i] = t.ReturnOnNotional()
	}

	n := float64(len(returns))
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / n

	var sumSqDev float64
	for _, r := range returns {
		d := r - mean
		sumSqDev += d * d
	}
	variance := sumSqDev / (n - 1)
	if variance == 0 {
		return 0.0
	}
	return mean / math.Sqrt(variance)
}
