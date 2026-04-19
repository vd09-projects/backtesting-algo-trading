package analytics

// Mathematical correctness tests for unexported computation functions.
// These use small synthetic curves to verify formulas precisely, intentionally
// bypassing the signal-frequency gate in Compute(). Gate behavior is tested
// separately in analytics_test.go.

import (
	"math"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// TestNormInvCDF covers all branches including edge cases and the lower tail.
func TestNormInvCDF(t *testing.T) {
	t.Parallel()
	tests := []struct {
		p    float64
		want float64
	}{
		{0.5, 0.0},                    // median of standard normal
		{0.975, 1.959964},             // central region, upper end (standard ±1.96)
		{0.025, -1.959964},            // central region, lower end
		{0.01, -2.326348},             // lower tail (p < 0.02425)
		{0.99, 2.326348},              // upper tail (p > 0.97575)
		{0.0, math.Inf(-1)},           // boundary: p ≤ 0 → -Inf
		{1.0, math.Inf(1)},            // boundary: p ≥ 1 → +Inf
		{-0.1, math.Inf(-1)},          // p < 0 → -Inf
		{1.1, math.Inf(1)},            // p > 1 → +Inf
	}
	const tol = 1e-5
	for _, tt := range tests {
		got := normInvCDF(tt.p)
		if math.IsInf(tt.want, 0) {
			if got != tt.want {
				t.Errorf("normInvCDF(%g) = %v, want %v", tt.p, got, tt.want)
			}
			continue
		}
		if math.Abs(got-tt.want) > tol {
			t.Errorf("normInvCDF(%g) = %.6f, want %.6f (tol %.1e)", tt.p, got, tt.want, tol)
		}
	}
}

var internalBase = time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

func makePts(values ...float64) []model.EquityPoint {
	pts := make([]model.EquityPoint, len(values))
	for i, v := range values {
		pts[i] = model.EquityPoint{
			Timestamp: internalBase.Add(time.Duration(i) * time.Hour),
			Value:     v,
		}
	}
	return pts
}

func assertClose(t *testing.T, field string, want, got float64) {
	t.Helper()
	const tol = 0.0001
	d := got - want
	if d < -tol || d > tol {
		t.Errorf("%s: got %.4f, want %.4f", field, got, want)
	}
}

// TestComputeSharpeInternal verifies the Sharpe formula with small synthetic curves.
//
// Reference: curve [100, 110, 99, 108.9] → returns [+0.1, −0.1, +0.1]
//
//	mean   = 0.1/3
//	stddev = 2·0.1/√3  (sample, n−1 denominator)
//	Sharpe = (mean/stddev)·√N = √(3N)/6
//
// Daily (N=252): √21 ≈ 4.5826
func TestComputeSharpeInternal(t *testing.T) {
	sqrt21 := math.Sqrt(21)

	cases := []struct {
		name  string
		curve []model.EquityPoint
		tf    model.Timeframe
		want  float64
	}{
		{"nil returns", nil, model.TimeframeDaily, 0},
		{"single point", makePts(100), model.TimeframeDaily, 0},
		{"two points — one return, no variance", makePts(100, 110), model.TimeframeDaily, 0},
		{"constant — zero variance", makePts(100, 100, 100, 100), model.TimeframeDaily, 0},
		{"known sequence daily", makePts(100, 110, 99, 108.9), model.TimeframeDaily, sqrt21},
		{"known sequence 15min", makePts(100, 110, 99, 108.9), model.Timeframe15Min, 5 * sqrt21},
		{"negative mean daily", makePts(100, 90, 81, 89.1), model.TimeframeDaily, -sqrt21},
		{"known sequence weekly", makePts(100, 110, 99, 108.9), model.TimeframeWeekly, math.Sqrt(156) / 6},
		{"known sequence 5min", makePts(100, 110, 99, 108.9), model.Timeframe5Min, math.Sqrt(56700) / 6},
		{"known sequence 1min", makePts(100, 110, 99, 108.9), model.Timeframe1Min, math.Sqrt(283500) / 6},
		{"unknown timeframe", makePts(100, 110, 99, 108.9), model.Timeframe("unknown"), 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeSharpe(computeReturns(tc.curve), tc.tf)
			assertClose(t, "Sharpe", tc.want, got)
		})
	}
}

// TestComputeSortinoInternal verifies the Sortino formula.
//
// curve [100, 110, 99, 108.9]; returns [+0.1, −0.1, +0.1]; daily (N=252).
//
//	σ_d = √(0.01/3) = 0.1/√3; mean = 0.1/3
//	Sortino = (0.1/3)/(0.1/√3)·√252 = (√3/3)·√252 = √84 = 2√21
func TestComputeSortinoInternal(t *testing.T) {
	curve := makePts(100, 110, 99, 108.9)
	got := computeSortino(computeReturns(curve), model.TimeframeDaily)
	assertClose(t, "Sortino", 2*math.Sqrt(21), got)
}

func TestComputeSortino_ZeroDownsideInternal(t *testing.T) {
	// All returns non-negative → downside deviation = 0 → Sortino = 0.
	curve := makePts(100, 110, 121, 133.1)
	got := computeSortino(computeReturns(curve), model.TimeframeDaily)
	assertClose(t, "Sortino", 0, got)
}

// TestComputeCalmarInternal verifies the Calmar formula.
//
// curve [100, 110, 99, 108.9] daily; returns [+0.1, −0.1, +0.1].
//
//	annReturn = (0.1/3)·252 = 8.4
//	MaxDrawdown = (110−99)/110·100 = 10%
//	Calmar = 8.4 / 0.10 = 84.0
func TestComputeCalmarInternal(t *testing.T) {
	curve := makePts(100, 110, 99, 108.9)
	maxDD := computeMaxDrawdownDepth(curve)
	got := computeCalmar(computeReturns(curve), model.TimeframeDaily, maxDD)
	assertClose(t, "Calmar", 84.0, got)
}

func TestComputeCalmar_ZeroDrawdownInternal(t *testing.T) {
	// Monotonically increasing curve → MaxDrawdown = 0 → Calmar = 0.
	curve := makePts(100, 110, 120, 130)
	maxDD := computeMaxDrawdownDepth(curve)
	got := computeCalmar(computeReturns(curve), model.TimeframeDaily, maxDD)
	assertClose(t, "Calmar", 0, got)
}

// TestComputeTailRatioInternal verifies the tail ratio formula.
//
// returns from [100, 110, 99, 108.9] sorted ascending: [−0.1, +0.1, +0.1]
// n=3; p5=sorted[0]=−0.1; p95=sorted[2]=+0.1; TailRatio = 0.1/0.1 = 1.0
func TestComputeTailRatioInternal(t *testing.T) {
	curve := makePts(100, 110, 99, 108.9)
	got := computeTailRatio(computeReturns(curve))
	assertClose(t, "TailRatio", 1.0, got)
}

func TestComputeTailRatio_AllPositiveInternal(t *testing.T) {
	// All returns non-negative → p5 >= 0 → TailRatio = 0.
	curve := makePts(100, 110, 121, 133.1)
	got := computeTailRatio(computeReturns(curve))
	assertClose(t, "TailRatio", 0, got)
}
