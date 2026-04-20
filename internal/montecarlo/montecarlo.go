// Package montecarlo provides Monte Carlo bootstrap for strategy performance statistics.
// The bootstrap resamples from the per-trade return series to produce confidence intervals
// on Sharpe ratio and max drawdown.
package montecarlo

import (
	"math"
	"math/rand/v2"
	"sort"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

const defaultNSimulations = 10_000

// BootstrapConfig controls the Monte Carlo resampling run.
type BootstrapConfig struct {
	NSimulations int   // number of bootstrap iterations; 0 defaults to 10,000
	Seed         int64 // RNG seed — caller must log this alongside results for reproducibility
}

// BootstrapResult holds statistics derived from the bootstrapped return distributions.
//
// Sharpe values are per-trade Sharpe: mean(ReturnOnNotional) / stddev(ReturnOnNotional),
// sample variance, no annualization factor. The live kill-switch comparison (TASK-0026)
// must use the same computation — rolling window of closed trade ReturnOnNotional values,
// same formula — to preserve the apples-to-apples comparison. See the algorithm decision
// recorded during TASK-0024 for the rationale behind non-annualized per-trade Sharpe.
//
// The p5 Sharpe is the kill-switch threshold: halt live trading when the rolling per-trade
// Sharpe drops below it. The threshold is relative to a sample of len(trades) — the window
// size used in live monitoring should be documented in TASK-0026.
type BootstrapResult struct {
	MeanSharpe         float64
	SharpeP5           float64
	SharpeP50          float64
	SharpeP95          float64
	WorstDrawdownP5    float64 // peak-to-trough %, 0–100
	WorstDrawdownP50   float64
	WorstDrawdownP95   float64
	ProbPositiveSharpe float64 // fraction 0–1: proportion of simulations with Sharpe > 0
}

// Bootstrap resamples the per-trade return series with replacement NSimulations times,
// computing Sharpe ratio and worst drawdown for each resample. Returns a zero BootstrapResult
// for fewer than 2 trades (insufficient to compute sample variance).
//
// cfg.Seed must be logged by the caller — it is the only way to reproduce a specific run.
//
// **Decision (2026-04.2.0) — architecture: experimental**
// scope: internal/montecarlo
// tags: montecarlo, package-boundary, analytics
//
// New package rather than a function in internal/analytics. Analytics is a pure function
// over equity curves ([]model.EquityPoint); bootstrap operates on trade returns
// ([]model.Trade via ReturnOnNotional). Different inputs, different algorithm, different
// dependency (math/rand/v2). Keeping them separate avoids importing a non-deterministic RNG
// into a package that otherwise has none, and prevents coupling two things that evolve
// independently.
//
// **Decision (2026-04.3.0) — tradeoff: experimental**
// scope: internal/montecarlo
// tags: rng, pcg, math/rand/v2
//
// rand.NewPCG(uint64(cfg.Seed), 0) from math/rand/v2 (Go 1.22+). PCG64 is the v2 default:
// high statistical quality, 128-bit state, faster than the old v1 source. Stream fixed at 0;
// only the seed varies. Rejected math/rand v1: deprecated and replaced by v2 in Go 1.22.
func Bootstrap(trades []model.Trade, cfg BootstrapConfig) BootstrapResult {
	if len(trades) < 2 {
		return BootstrapResult{}
	}

	nSim := cfg.NSimulations
	if nSim <= 0 {
		nSim = defaultNSimulations
	}

	returns := make([]float64, len(trades))
	for i, t := range trades {
		returns[i] = t.ReturnOnNotional()
	}

	rng := rand.New(rand.NewPCG(uint64(cfg.Seed), 0))

	sharpes := make([]float64, nSim)
	drawdowns := make([]float64, nSim)
	n := len(returns)
	buf := make([]float64, n) // pre-allocated; reused across iterations to avoid hot-loop allocs

	for i := range nSim {
		for j := range n {
			buf[j] = returns[rng.IntN(n)]
		}
		sharpes[i] = sampleSharpe(buf)
		drawdowns[i] = worstDrawdown(buf)
	}

	sort.Float64s(sharpes)
	sort.Float64s(drawdowns)

	var posCount int
	var sumSharpe float64
	for _, s := range sharpes {
		sumSharpe += s
		if s > 0 {
			posCount++
		}
	}

	return BootstrapResult{
		MeanSharpe:         sumSharpe / float64(nSim),
		SharpeP5:           percentile(sharpes, 0.05),
		SharpeP50:          percentile(sharpes, 0.50),
		SharpeP95:          percentile(sharpes, 0.95),
		WorstDrawdownP5:    percentile(drawdowns, 0.05),
		WorstDrawdownP50:   percentile(drawdowns, 0.50),
		WorstDrawdownP95:   percentile(drawdowns, 0.95),
		ProbPositiveSharpe: float64(posCount) / float64(nSim),
	}
}

// sampleSharpe computes the non-annualized per-trade Sharpe from a return series.
// Uses sample variance (n-1 denominator), consistent with computeSharpe in internal/analytics.
// Returns 0 for zero variance (all returns identical after resampling).
func sampleSharpe(returns []float64) float64 {
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
		return 0
	}
	return mean / math.Sqrt(variance)
}

// worstDrawdown computes the maximum peak-to-trough drawdown as a percentage (0–100)
// from a sequence of per-trade returns using geometric compounding.
//
// **Decision (2026-04.4.0) — convention: experimental**
// scope: internal/montecarlo
// tags: drawdown, geometric-compounding
//
// Geometric compounding (equity *= 1+r) rather than arithmetic accumulation (sum of returns).
// Arithmetic drift accumulates error as trade count grows; geometric is the correct model for
// compound returns. For small per-trade returns the difference is negligible, but establishing
// the convention now avoids an inconsistency if return magnitudes grow in future strategies.
// Equity is floored at 0 to handle returns < -1 (e.g., leveraged losses exceeding notional).
func worstDrawdown(returns []float64) float64 {
	equity := 1.0
	peak := 1.0
	var maxDD float64
	for _, r := range returns {
		equity *= 1 + r
		if equity < 0 {
			equity = 0
		}
		if equity > peak {
			peak = equity
		}
		if peak > 0 {
			dd := (peak - equity) / peak * 100
			if dd > maxDD {
				maxDD = dd
			}
		}
	}
	return maxDD
}

// percentile returns the p-th quantile (p in [0,1]) from a sorted slice using nearest-rank.
//
// **Decision (2026-04.5.0) — tradeoff: experimental**
// scope: internal/montecarlo
// tags: percentile, nearest-rank
//
// Nearest-rank (floor index) rather than linear interpolation. The kill-switch p5 Sharpe is
// an observed simulation value, not an interpolated one — this makes the threshold more
// interpretable and the computation bit-identical regardless of nSim. Linear interpolation
// would produce a smooth estimate but at the cost of a value that never actually occurred
// in any simulation, which is harder to explain and verify.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(math.Floor(p * float64(len(sorted))))
	if idx >= len(sorted) {
		idx = len(sorted) - 1
	}
	return sorted[idx]
}
