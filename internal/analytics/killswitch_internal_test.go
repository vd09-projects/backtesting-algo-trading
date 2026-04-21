package analytics

import (
	"testing"
	"time"
)

// --- computePerTradeSharpe ---

func TestComputePerTradeSharpe_Degenerate(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		returns []float64
	}{
		{"nil", nil},
		{"single value", []float64{0.1}},
		{"zero variance — all identical", []float64{0.5, 0.5, 0.5}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := computePerTradeSharpe(tc.returns); got != 0 {
				t.Errorf("computePerTradeSharpe(%v) = %.4f, want 0", tc.returns, got)
			}
		})
	}
}

// TestComputePerTradeSharpe_KnownValue verifies the formula with a hand-computed case.
//
// returns [0.2, 0.3, 0.25]; n=3; mean=0.25
// deviations: [-0.05, 0.05, 0]; sumSqDev = 0.0025+0.0025+0 = 0.005
// variance(n-1=2) = 0.0025; std = 0.05
// Sharpe = 0.25/0.05 = 5.0
func TestComputePerTradeSharpe_KnownValue(t *testing.T) {
	t.Parallel()
	assertClose(t, "Sharpe", 5.0, computePerTradeSharpe([]float64{0.2, 0.3, 0.25}))
}

func TestComputePerTradeSharpe_NegativeMean(t *testing.T) {
	t.Parallel()
	// Symmetric to the positive case: same magnitude, negative Sharpe.
	assertClose(t, "Sharpe", -5.0, computePerTradeSharpe([]float64{-0.2, -0.3, -0.25}))
}

// --- computeCurrentDrawdownDepth ---

func TestComputeCurrentDrawdownDepth(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values []float64
		want   float64
	}{
		{"empty", nil, 0},
		{"single point", []float64{100}, 0},
		{"monotone up — at all-time high", []float64{100, 110, 120}, 0},
		{"recovered to initial peak", []float64{100, 90, 100}, 0},
		// peak=110@t1, last=99: (110-99)/110*100 = 10.0%
		{"in drawdown — peak in middle", []float64{100, 110, 99}, (110 - 99.0) / 110 * 100},
		// peak=110@t0, last=99@t2: same depth regardless of peak position
		{"in drawdown — peak at first bar", []float64{110, 90, 99}, (110 - 99.0) / 110 * 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertClose(t, "drawdownDepth", tc.want, computeCurrentDrawdownDepth(makePts(tc.values...)))
		})
	}
}

func TestComputeCurrentDrawdownDepth_FullLoss(t *testing.T) {
	t.Parallel()
	// Equity goes to zero: 100% drawdown.
	assertClose(t, "drawdownDepth", 100.0, computeCurrentDrawdownDepth(makePts(100, 0)))
}

// --- computeCurrentDDDuration ---

func TestComputeCurrentDDDuration(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		values []float64
		want   time.Duration
	}{
		{"empty", nil, 0},
		{"single point", []float64{100}, 0},
		{"monotone up — not in drawdown", []float64{100, 110, 120}, 0},
		// Recovery back to initial peak: last=100 >= peak=100 → 0.
		{"recovered to peak", []float64{100, 90, 100}, 0},
		// Peak at t1=110, last=99@t2: duration = t2-t1 = 1 hour.
		{"in drawdown — peak in middle", []float64{100, 110, 99}, time.Hour},
		// Peak at t0=110, last=99@t2: duration = t2-t0 = 2 hours.
		{"in drawdown — peak at first bar", []float64{110, 90, 99}, 2 * time.Hour},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeCurrentDDDuration(makePts(tc.values...))
			if got != tc.want {
				t.Errorf("computeCurrentDDDuration(%v) = %v, want %v", tc.values, got, tc.want)
			}
		})
	}
}
