package analytics

import (
	"math"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// NSE stress periods for correlation analysis (Marcus, 2026-04-21).
// Hardcoded as named values — these are reference-regime definitions, not parameters.
var (
	crash2020Start      = time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)
	crash2020End        = time.Date(2020, 6, 30, 23, 59, 59, 0, time.UTC)
	correction2022Start = time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	correction2022End   = time.Date(2022, 12, 31, 23, 59, 59, 0, time.UTC)
)

// Correlation thresholds (Marcus, 2026-04-21).
const (
	fullPeriodCorrThreshold   = 0.7
	stressPeriodCorrThreshold = 0.6
)

// NamedCurve pairs a strategy name with its per-bar equity curve.
type NamedCurve struct {
	Name  string
	Curve []model.EquityPoint
}

// PairCorrelation holds Pearson correlation metrics for a strategy pair.
// FullPeriod, Crash2020, and Correction2022 are NaN when the underlying
// return series has zero variance or no data points in the window.
// TooCorrelated is true when FullPeriod > 0.7 or either stress period > 0.6;
// NaN values do not trigger the flag.
type PairCorrelation struct {
	NameA, NameB   string
	FullPeriod     float64
	Crash2020      float64
	Correction2022 float64
	TooCorrelated  bool
}

// CorrelationMatrix holds all pairwise correlations for a set of strategies.
type CorrelationMatrix struct {
	Pairs []PairCorrelation
}

// ComputeCorrelation computes pairwise Pearson correlation between two named equity curves.
// Warmup bars (leading bars where equity equals the initial value) are trimmed before
// any computation. Returns and stress-period slices are derived from the trimmed, aligned curves.
func ComputeCorrelation(a, b NamedCurve) PairCorrelation {
	aReturns, bReturns, aPts, bPts := alignAndTrim(a.Curve, b.Curve)

	full := pearson(aReturns, bReturns)
	crash := stressPearson(aPts, bPts, crash2020Start, crash2020End)
	corr := stressPearson(aPts, bPts, correction2022Start, correction2022End)

	tooCorrFull := !math.IsNaN(full) && full > fullPeriodCorrThreshold
	tooCorrCrash := !math.IsNaN(crash) && crash > stressPeriodCorrThreshold
	tooCorrCorr := !math.IsNaN(corr) && corr > stressPeriodCorrThreshold

	return PairCorrelation{
		NameA:          a.Name,
		NameB:          b.Name,
		FullPeriod:     full,
		Crash2020:      crash,
		Correction2022: corr,
		TooCorrelated:  tooCorrFull || tooCorrCrash || tooCorrCorr,
	}
}

// ComputeMatrix computes the full pairwise correlation matrix for a set of named curves.
// Pairs are in lexicographic (i,j) order with i < j.
func ComputeMatrix(curves []NamedCurve) CorrelationMatrix {
	var pairs []PairCorrelation
	for i := range len(curves) {
		for j := i + 1; j < len(curves); j++ {
			pairs = append(pairs, ComputeCorrelation(curves[i], curves[j]))
		}
	}
	return CorrelationMatrix{Pairs: pairs}
}

// alignAndTrim trims warmup bars and returns parallel return slices and equity point
// slices for stress-window computation. Warmup is detected as the leading run of bars
// where equity equals the first bar's value in each curve. The trim point is the
// maximum warmup end across both curves.
//
// **Decision (2026-04.1.0) — architecture: experimental**
// scope: internal/analytics, correlation
// tags: warmup-detection, correlation, TASK-0027
//
// Warmup is detected by first-change heuristic (first bar where equity ≠ curve[0].Value)
// rather than an explicit lookback int parameter. Keeps ComputeCorrelation decoupled from
// strategy configuration. A strategy that never trades produces a constant series → NaN.
// Rejected: explicit warmup int — would require threading strategy config into every caller.
func alignAndTrim(a, b []model.EquityPoint) (aRet, bRet []float64, aPts, bPts []model.EquityPoint) {
	aStart := firstActiveIndex(a)
	bStart := firstActiveIndex(b)
	start := max(aStart, bStart)

	if start >= len(a) || start >= len(b) {
		return nil, nil, nil, nil
	}

	aPts = a[start:]
	bPts = b[start:]

	// Align lengths — curves may differ by a bar if one ended earlier.
	n := min(len(aPts), len(bPts))
	aPts = aPts[:n]
	bPts = bPts[:n]

	aRet = computeReturns(aPts)
	bRet = computeReturns(bPts)
	return aRet, bRet, aPts, bPts
}

// firstActiveIndex returns the index of the first bar where equity differs from
// the initial value. Returns len(curve) if the curve is nil, empty, or all-flat.
func firstActiveIndex(curve []model.EquityPoint) int {
	if len(curve) == 0 {
		return 0
	}
	initial := curve[0].Value
	for i, pt := range curve {
		if pt.Value != initial {
			return i
		}
	}
	return len(curve)
}

// stressPearson computes the Pearson correlation of returns within [from, to] inclusive.
// Returns NaN if the window contains fewer than 2 equity points in either curve.
func stressPearson(aPts, bPts []model.EquityPoint, from, to time.Time) float64 {
	aWindow := filterWindow(aPts, from, to)
	bWindow := filterWindow(bPts, from, to)

	n := min(len(aWindow), len(bWindow))
	if n < 2 {
		return math.NaN()
	}
	aWindow = aWindow[:n]
	bWindow = bWindow[:n]

	return pearson(computeReturns(aWindow), computeReturns(bWindow))
}

// filterWindow returns the subset of pts whose timestamps fall within [from, to].
func filterWindow(pts []model.EquityPoint, from, to time.Time) []model.EquityPoint {
	var out []model.EquityPoint
	for _, pt := range pts {
		if !pt.Timestamp.Before(from) && !pt.Timestamp.After(to) {
			out = append(out, pt)
		}
	}
	return out
}

// pearson computes the Pearson correlation coefficient for two equal-length float slices.
// Returns NaN if either series has zero variance or fewer than 2 elements.
//
// **Decision (2026-04.1.1) — convention: experimental**
// scope: internal/analytics, correlation
// tags: NaN, sentinel, pearson, TASK-0027
//
// math.NaN() is the sentinel for undefined correlation (constant series, empty window),
// not 0. Zero is a valid correlation value; NaN propagates through float math and is
// detectable with math.IsNaN. All callers guard with math.IsNaN before interpreting.
func pearson(x, y []float64) float64 {
	n := len(x)
	if n < 2 || len(y) < 2 || n != len(y) {
		return math.NaN()
	}

	fn := float64(n)
	var sumX, sumY float64
	for i := range x {
		sumX += x[i]
		sumY += y[i]
	}
	meanX := sumX / fn
	meanY := sumY / fn

	var cov, varX, varY float64
	for i := range x {
		dx := x[i] - meanX
		dy := y[i] - meanY
		cov += dx * dy
		varX += dx * dx
		varY += dy * dy
	}

	if varX == 0 || varY == 0 {
		return math.NaN()
	}

	return cov / math.Sqrt(varX*varY)
}
