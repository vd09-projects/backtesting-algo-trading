package analytics_test

import (
	"math"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// corrBase is the epoch for correlation test curves. Chosen inside the 2020 crash
// window so stress-period tests can use a fixed anchor without date arithmetic.
var corrBase = time.Date(2020, 1, 14, 18, 30, 0, 0, time.UTC)

// makeCorrCurve builds a NamedCurve with daily bars starting at corrBase.
// Each value in vals becomes one EquityPoint; timestamps advance by 24h.
func makeCorrCurve(name string, vals ...float64) analytics.NamedCurve {
	pts := make([]model.EquityPoint, len(vals))
	for i, v := range vals {
		pts[i] = model.EquityPoint{
			Timestamp: corrBase.Add(time.Duration(i) * 24 * time.Hour),
			Value:     v,
		}
	}
	return analytics.NamedCurve{Name: name, Curve: pts}
}

// makeWarmupCurve builds a NamedCurve where the first `warmup` bars are flat at
// initialVal, followed by the active equity values in active.
func makeWarmupCurve(name string, warmup int, initialVal float64, active ...float64) analytics.NamedCurve {
	var vals []float64
	for range warmup {
		vals = append(vals, initialVal)
	}
	vals = append(vals, active...)
	return makeCorrCurve(name, vals...)
}

// assertNaN fails if v is not NaN.
func assertNaN(t *testing.T, field string, v float64) {
	t.Helper()
	if !math.IsNaN(v) {
		t.Errorf("%s: got %.4f, want NaN", field, v)
	}
}

// assertApprox fails if got differs from want by more than tol.
func assertApprox(t *testing.T, field string, want, got, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s: got %.6f, want %.6f (tol %.1e)", field, got, want, tol)
	}
}

// ── ComputeCorrelation — Pearson values ──────────────────────────────────────

func TestComputeCorrelation_PerfectPositive(t *testing.T) {
	// equity [100,101,103,106,110] vs [200,202,206,212,220]: returns are identical → r=1
	a := makeCorrCurve("A", 100, 101, 103, 106, 110)
	b := makeCorrCurve("B", 200, 202, 206, 212, 220)
	p := analytics.ComputeCorrelation(a, b)
	assertApprox(t, "FullPeriod", 1.0, p.FullPeriod, 1e-9)
}

func TestComputeCorrelation_PerfectNegative(t *testing.T) {
	// returns of A are negated in B → r=-1
	a := makeCorrCurve("A", 100, 110, 99, 108.9)
	b := makeCorrCurve("B", 100, 90, 101, 91.9)
	p := analytics.ComputeCorrelation(a, b)
	assertApprox(t, "FullPeriod", -1.0, p.FullPeriod, 1e-9)
}

func TestComputeCorrelation_ZeroCorrelation(t *testing.T) {
	// Construct equity curves that produce orthogonal return series after warmup trimming.
	// alignAndTrim skips index 0 (always flat = initial value), so the first two bars
	// in each curve are flat, and the active sequence begins at index 2.
	//
	// A active returns = [+0.05, −0.05, +0.05, −0.05]  (mean = 0)
	// B active returns = [+0.03, +0.03, −0.03, −0.03]  (mean = 0)
	// Inner product = (+0.05)(+0.03)+(−0.05)(+0.03)+(+0.05)(−0.03)+(−0.05)(−0.03) = 0
	// → Pearson = 0 exactly.
	//
	// Equity values derived by compounding: each e_n = e_{n-1} × (1 + r_n).
	a := makeCorrCurve("A", 100, 100, 95, 99.75, 94.7625, 99.500625, 94.52559375)
	b := makeCorrCurve("B", 200, 200, 206, 212.18, 218.5454, 211.989038, 205.629367)
	p := analytics.ComputeCorrelation(a, b)
	if math.IsNaN(p.FullPeriod) {
		t.Fatal("FullPeriod: got NaN, want ~0")
	}
	// Float64 precision in the hardcoded equity values means cov ≈ 0 but not bit-exact;
	// 1e-6 is well within any practical zero-correlation definition.
	assertApprox(t, "FullPeriod", 0, p.FullPeriod, 1e-6)
}

func TestComputeCorrelation_ConstantSeries_ReturnsNaN(t *testing.T) {
	// A never trades — all equity flat → zero-variance returns → NaN
	a := makeCorrCurve("A", 100, 100, 100, 100, 100)
	b := makeCorrCurve("B", 100, 101, 103, 106, 110)
	p := analytics.ComputeCorrelation(a, b)
	assertNaN(t, "FullPeriod", p.FullPeriod)
	if p.TooCorrelated {
		t.Error("TooCorrelated: got true for NaN correlation, want false")
	}
}

func TestComputeCorrelation_NamesPreserved(t *testing.T) {
	a := makeCorrCurve("SMA", 100, 101, 103)
	b := makeCorrCurve("RSI", 100, 102, 105)
	p := analytics.ComputeCorrelation(a, b)
	if p.NameA != "SMA" {
		t.Errorf("NameA: got %q, want %q", p.NameA, "SMA")
	}
	if p.NameB != "RSI" {
		t.Errorf("NameB: got %q, want %q", p.NameB, "RSI")
	}
}

// ── Warmup trimming ───────────────────────────────────────────────────────────

func TestComputeCorrelation_WarmupTrimmed(t *testing.T) {
	// A: warmup=4 (flat at 100), then [110, 104.5, 114.95, 109.2025, 120.1228, 114.1167]
	//   firstActiveIndex = 4; A[4:] has 6 elements.
	//
	// B: warmup=2 (flat at 100), then [110, 104.5, 114.95, 109.2025, 120.1228, 114.1167]
	//   firstActiveIndex = 2; B[4:] has 4 elements.
	//
	// After trim to max(4,2)=4, aligned window n=min(6,4)=4:
	//   A[4:8] = [110, 104.5, 114.95, 109.2025]
	//   B[4:8] = [114.95, 109.2025, 120.1228, 114.1167]
	//   Both return sequences = [−0.05, +0.10, −0.05] → Pearson = 1.0.
	a := makeWarmupCurve("A", 4, 100, 110, 104.5, 114.95, 109.2025, 120.1228, 114.1167)
	b := makeWarmupCurve("B", 2, 100, 110, 104.5, 114.95, 109.2025, 120.1228, 114.1167)
	p := analytics.ComputeCorrelation(a, b)
	if math.IsNaN(p.FullPeriod) {
		t.Fatal("FullPeriod: got NaN — warmup trimming may have removed all active bars")
	}
	assertApprox(t, "FullPeriod", 1.0, p.FullPeriod, 1e-9)
}

func TestComputeCorrelation_BothAllWarmup_ReturnsNaN(t *testing.T) {
	// Neither curve ever trades — both constant → NaN
	a := makeCorrCurve("A", 100, 100, 100)
	b := makeCorrCurve("B", 100, 100, 100)
	p := analytics.ComputeCorrelation(a, b)
	assertNaN(t, "FullPeriod", p.FullPeriod)
}

// ── Stress-period windows ─────────────────────────────────────────────────────

func TestComputeCorrelation_Crash2020Window(t *testing.T) {
	// Build a curve that starts exactly at crash2020Start (2020-01-14).
	// We need enough active points inside the window to compute correlation.
	// makeCorrCurve already anchors to corrBase = 2020-01-14.
	a := makeCorrCurve("A", 100, 110, 99, 108.9, 120, 110, 121)
	b := makeCorrCurve("B", 100, 110, 99, 108.9, 120, 110, 121) // identical → r=1
	p := analytics.ComputeCorrelation(a, b)
	if math.IsNaN(p.Crash2020) {
		t.Fatal("Crash2020: got NaN — expected data points inside 2020 crash window")
	}
	assertApprox(t, "Crash2020", 1.0, p.Crash2020, 1e-9)
}

func TestComputeCorrelation_NoDataInWindow_ReturnsNaN(t *testing.T) {
	// Curve entirely before the 2022 correction window (2022-01-01).
	// Build a curve anchored to 2019 — correction2022 has no data → NaN.
	base2019 := time.Date(2019, 1, 1, 18, 30, 0, 0, time.UTC)
	makePre2022 := func(name string, vals ...float64) analytics.NamedCurve {
		pts := make([]model.EquityPoint, len(vals))
		for i, v := range vals {
			pts[i] = model.EquityPoint{
				Timestamp: base2019.Add(time.Duration(i) * 24 * time.Hour),
				Value:     v,
			}
		}
		return analytics.NamedCurve{Name: name, Curve: pts}
	}
	a := makePre2022("A", 100, 110, 99, 108.9)
	b := makePre2022("B", 100, 110, 99, 108.9)
	p := analytics.ComputeCorrelation(a, b)
	assertNaN(t, "Correction2022", p.Correction2022)
}

// ── TooCorrelated thresholds ─────────────────────────────────────────────────

func TestComputeCorrelation_TooCorrelated_FullPeriodTriggers(t *testing.T) {
	// Perfectly correlated pair → FullPeriod=1.0 > 0.7 → TooCorrelated=true
	a := makeCorrCurve("A", 100, 110, 121, 133.1)
	b := makeCorrCurve("B", 100, 110, 121, 133.1)
	p := analytics.ComputeCorrelation(a, b)
	if !p.TooCorrelated {
		t.Errorf("TooCorrelated: got false, want true (FullPeriod=%.4f)", p.FullPeriod)
	}
}

func TestComputeCorrelation_TooCorrelated_BelowThreshold(t *testing.T) {
	// Anti-correlated: FullPeriod=-1, both stress values unreachable for this data.
	// -1 is not > 0.7, so TooCorrelated must be false.
	a := makeCorrCurve("A", 100, 110, 99, 108.9)
	b := makeCorrCurve("B", 100, 90, 101, 91.9)
	p := analytics.ComputeCorrelation(a, b)
	if p.TooCorrelated {
		t.Errorf("TooCorrelated: got true, want false (FullPeriod=%.4f)", p.FullPeriod)
	}
}

// ── ComputeMatrix ─────────────────────────────────────────────────────────────

func TestComputeMatrix_ThreeCurves_ThreePairs(t *testing.T) {
	a := makeCorrCurve("A", 100, 110, 121)
	b := makeCorrCurve("B", 100, 110, 121)
	c := makeCorrCurve("C", 100, 90, 81)
	m := analytics.ComputeMatrix([]analytics.NamedCurve{a, b, c})
	if len(m.Pairs) != 3 {
		t.Fatalf("Pairs: got %d, want 3", len(m.Pairs))
	}
}

func TestComputeMatrix_TwoCurves_OnePair(t *testing.T) {
	a := makeCorrCurve("X", 100, 110, 121)
	b := makeCorrCurve("Y", 100, 110, 121)
	m := analytics.ComputeMatrix([]analytics.NamedCurve{a, b})
	if len(m.Pairs) != 1 {
		t.Fatalf("Pairs: got %d, want 1", len(m.Pairs))
	}
	if m.Pairs[0].NameA != "X" || m.Pairs[0].NameB != "Y" {
		t.Errorf("names: got (%q,%q), want (X,Y)", m.Pairs[0].NameA, m.Pairs[0].NameB)
	}
}

func TestComputeMatrix_SingleCurve_NoPairs(t *testing.T) {
	a := makeCorrCurve("A", 100, 110, 121)
	m := analytics.ComputeMatrix([]analytics.NamedCurve{a})
	if len(m.Pairs) != 0 {
		t.Fatalf("Pairs: got %d, want 0", len(m.Pairs))
	}
}

func TestComputeMatrix_Empty_NoPairs(t *testing.T) {
	m := analytics.ComputeMatrix(nil)
	if len(m.Pairs) != 0 {
		t.Fatalf("Pairs: got %d, want 0", len(m.Pairs))
	}
}
