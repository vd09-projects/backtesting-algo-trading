package analytics_test

import (
	"math"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

var baseTime = time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

func trade(pnl float64) model.Trade {
	return model.Trade{
		Instrument:  "NSE:NIFTY50",
		Direction:   model.DirectionLong,
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   100 + pnl,
		EntryTime:   baseTime,
		ExitTime:    baseTime.Add(time.Hour),
		RealizedPnL: pnl,
	}
}

func makeEquityCurve(values ...float64) []model.EquityPoint {
	pts := make([]model.EquityPoint, len(values))
	for i, v := range values {
		pts[i] = model.EquityPoint{Timestamp: baseTime.Add(time.Duration(i) * time.Hour), Value: v}
	}
	return pts
}

func TestCompute_Empty(t *testing.T) {
	r := analytics.Compute(nil, nil, "")

	if r.TradeCount != 0 {
		t.Errorf("TradeCount: got %d, want 0", r.TradeCount)
	}
	if r.TotalPnL != 0 {
		t.Errorf("TotalPnL: got %f, want 0", r.TotalPnL)
	}
	if r.WinRate != 0 {
		t.Errorf("WinRate: got %f, want 0", r.WinRate)
	}
	if r.MaxDrawdown != 0 {
		t.Errorf("MaxDrawdown: got %f, want 0", r.MaxDrawdown)
	}
	if r.WinCount != 0 {
		t.Errorf("WinCount: got %d, want 0", r.WinCount)
	}
	if r.LossCount != 0 {
		t.Errorf("LossCount: got %d, want 0", r.LossCount)
	}
}

func TestCompute_SingleWinner(t *testing.T) {
	r := analytics.Compute([]model.Trade{trade(100)}, nil, "")

	assertEqual(t, "TradeCount", 1, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 100, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
}

func TestCompute_SingleLoser(t *testing.T) {
	r := analytics.Compute([]model.Trade{trade(-50)}, nil, "")

	assertEqual(t, "TradeCount", 1, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", -50, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 0, r.WinRate)
	// No positive peak ever reached, so no measurable drawdown %
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 0, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
}

func TestCompute_AllWinners(t *testing.T) {
	trades := []model.Trade{trade(100), trade(200)}
	r := analytics.Compute(trades, nil, "")

	assertEqual(t, "TradeCount", 2, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 300, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
}

func TestCompute_AllLosers(t *testing.T) {
	trades := []model.Trade{trade(-50), trade(-50)}
	r := analytics.Compute(trades, nil, "")

	assertEqual(t, "TradeCount", 2, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", -100, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 0, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 0, r.WinCount)
	assertEqual(t, "LossCount", 2, r.LossCount)
}

func TestCompute_Mixed(t *testing.T) {
	// Trade-level stats are independent of the equity curve; MaxDrawdown comes
	// from the per-bar curve. Curve [100, 120, 60, 90]:
	//   peak=120@t1, trough=60@t2 → MaxDrawdown = (120-60)/120 * 100 = 50%
	trades := []model.Trade{trade(200), trade(-100), trade(50)}
	curve := makeEquityCurve(100, 120, 60, 90)
	r := analytics.Compute(trades, curve, "")

	assertEqual(t, "TradeCount", 3, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 150, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 66.6667, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 50, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
}

func TestCompute_BreakevenCountsAsLoss(t *testing.T) {
	trades := []model.Trade{trade(100), trade(0)}
	r := analytics.Compute(trades, nil, "")

	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
	assertFloatEqual(t, "WinRate", 50, r.WinRate)
}

// TestComputeSharpe verifies annualized Sharpe ratio computation from equity curve.
//
// Reference sequence: equity [100, 110, 99, 108.9] → per-bar returns [+0.1, -0.1, +0.1]
//
//	mean   = 0.1/3
//	stddev = 2·0.1/√3   (sample, n-1 denominator)
//	Sharpe = (mean/stddev)·√N = (√3/6)·√N
//	       = √(3N)/6
//
// Daily (N=252):    √756/6  = √(36·21)/6 = 6√21/6 = √21   ≈ 4.5826
// 15min (N=6300):   √18900/6 = √(900·21)/6 = 30√21/6 = 5√21 ≈ 22.9129
func TestComputeSharpe(t *testing.T) {
	sqrtOf21 := math.Sqrt(21)

	cases := []struct {
		name       string
		curve      []model.EquityPoint
		tf         model.Timeframe
		wantSharpe float64
	}{
		{
			name:       "empty curve",
			curve:      nil,
			tf:         model.TimeframeDaily,
			wantSharpe: 0,
		},
		{
			name:       "single point",
			curve:      makeEquityCurve(100),
			tf:         model.TimeframeDaily,
			wantSharpe: 0,
		},
		{
			name:       "two points — only one return, sample stddev undefined",
			curve:      makeEquityCurve(100, 110),
			tf:         model.TimeframeDaily,
			wantSharpe: 0,
		},
		{
			name:       "constant equity — zero variance",
			curve:      makeEquityCurve(100, 100, 100, 100),
			tf:         model.TimeframeDaily,
			wantSharpe: 0,
		},
		{
			// Returns [+0.1, -0.1, +0.1]; daily annualization N=252; Sharpe = √21
			name:       "known sequence daily",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.TimeframeDaily,
			wantSharpe: sqrtOf21,
		},
		{
			// Same returns, 15min annualization N=6300; Sharpe = 5·√21
			name:       "known sequence 15min",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.Timeframe15Min,
			wantSharpe: 5 * sqrtOf21,
		},
		{
			// Returns [-0.1, -0.1, +0.1]; mean is negative; Sharpe = -√21
			name:       "negative mean — daily",
			curve:      makeEquityCurve(100, 90, 81, 89.1),
			tf:         model.TimeframeDaily,
			wantSharpe: -sqrtOf21,
		},
		{
			// Returns [+0.1, -0.1, +0.1]; weekly annualization N=52; Sharpe = √(3·52)/6 = √156/6
			name:       "known sequence weekly",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.TimeframeWeekly,
			wantSharpe: math.Sqrt(156) / 6,
		},
		{
			// Returns [+0.1, -0.1, +0.1]; 5min annualization N=18900; Sharpe = √(3·18900)/6 = √56700/6
			name:       "known sequence 5min",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.Timeframe5Min,
			wantSharpe: math.Sqrt(56700) / 6,
		},
		{
			// Returns [+0.1, -0.1, +0.1]; 1min annualization N=94500; Sharpe = √(3·94500)/6 = √283500/6
			name:       "known sequence 1min",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.Timeframe1Min,
			wantSharpe: math.Sqrt(283500) / 6,
		},
		{
			name:       "unknown timeframe — annualization factor 0",
			curve:      makeEquityCurve(100, 110, 99, 108.9),
			tf:         model.Timeframe("unknown"),
			wantSharpe: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := analytics.Compute(nil, tc.curve, tc.tf)
			assertFloatEqual(t, "SharpeRatio", tc.wantSharpe, r.SharpeRatio)
		})
	}
}

// --- ProfitFactor, AvgWin, AvgLoss ---

func TestCompute_ProfitFactor_AllWinners(t *testing.T) {
	// No losing trades → ProfitFactor = 0 (guard against division by zero).
	r := analytics.Compute([]model.Trade{trade(100), trade(50)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 75, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 0, r.AvgLoss)
}

func TestCompute_ProfitFactor_AllLosers(t *testing.T) {
	// No winning trades → ProfitFactor = 0, AvgWin = 0; AvgLoss = (30+20)/2 = 25.
	r := analytics.Compute([]model.Trade{trade(-30), trade(-20)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 0, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 25, r.AvgLoss)
}

func TestCompute_ProfitFactor_Mixed(t *testing.T) {
	// grossProfit = 100+50 = 150; grossLoss = 30+20 = 50
	// PF = 150/50 = 3.0; AvgWin = 150/2 = 75; AvgLoss = 50/2 = 25
	r := analytics.Compute([]model.Trade{trade(100), trade(50), trade(-30), trade(-20)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 3.0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 75, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 25, r.AvgLoss)
}

// TestCompute_MaxDrawdown_NeverExceeds100 is a regression test for the bug where
// losses that exceeded all prior gains caused MaxDrawdown > 100%. The old
// implementation accumulated equity from zero; when cumulative P&L went negative
// the formula (peak - negative_equity) / peak produced values > 100%. The fix
// uses the per-bar equity curve, which starts at initial cash, so the denominator
// is always a real account value.
//
// Scenario: account at 100 000, gains 10 000 (peak=110 000), then loses 12 000
// (trough=98 000). Real MaxDrawdown = (110 000−98 000)/110 000 = 10.91%.
// Old implementation: equity 0→+10 000→−2 000, DD = (10 000−(−2 000))/10 000 = 120%.
func TestCompute_MaxDrawdown_NeverExceeds100(t *testing.T) {
	curve := makeEquityCurve(100_000, 110_000, 98_000)
	r := analytics.Compute(
		[]model.Trade{trade(10_000), trade(-12_000)},
		curve,
		model.TimeframeDaily,
	)

	if r.MaxDrawdown >= 100 {
		t.Errorf("MaxDrawdown must never exceed 100%% — got %.2f%%", r.MaxDrawdown)
	}
	// (110 000 − 98 000) / 110 000 × 100 = 10.909…%
	assertFloatEqual(t, "MaxDrawdown", 10.9091, r.MaxDrawdown)
}

// --- SortinoRatio ---

func TestCompute_SortinoRatio(t *testing.T) {
	// Curve [100, 110, 99, 108.9]; returns [+0.1, -0.1, +0.1]; daily (N=252).
	//
	// σ_d = √(sum(min(r,0)²) / n) = √(0.01/3) = 0.1/√3
	// mean = 0.1/3
	// Sortino = (mean / σ_d) · √N = (0.1/3)/(0.1/√3) · √252 = (√3/3)·√252 = √756/3 = 2√21
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "SortinoRatio", 2*math.Sqrt(21), r.SortinoRatio)
}

func TestCompute_SortinoRatio_ZeroDownside(t *testing.T) {
	// All returns non-negative → downside deviation = 0 → Sortino = 0.
	curve := makeEquityCurve(100, 110, 121, 133.1)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "SortinoRatio", 0, r.SortinoRatio)
}

// --- CalmarRatio ---

func TestCompute_CalmarRatio(t *testing.T) {
	// Curve [100, 110, 99, 108.9] daily; returns [+0.1, -0.1, +0.1].
	// annReturn = (0.1/3) · 252 = 8.4
	// MaxDrawdown from curve: peak=110@t1, trough=99@t2 → (110-99)/110 * 100 = 10%
	// Calmar = 8.4 / (10/100) = 84.0
	trades := []model.Trade{trade(200), trade(-100)}
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(trades, curve, model.TimeframeDaily)
	assertFloatEqual(t, "CalmarRatio", 84.0, r.CalmarRatio)
}

func TestCompute_CalmarRatio_ZeroDrawdown(t *testing.T) {
	// Monotonically increasing curve → MaxDrawdown = 0 → Calmar = 0.
	// (The previous test used curve [100,110,99,108.9] which has a 10% drawdown —
	// that was wrong; this test now uses a curve with no drawdown.)
	trades := []model.Trade{trade(100), trade(200)}
	curve := makeEquityCurve(100, 110, 120, 130)
	r := analytics.Compute(trades, curve, model.TimeframeDaily)
	assertFloatEqual(t, "CalmarRatio", 0, r.CalmarRatio)
}

// --- MaxDrawdownDuration ---

func TestComputeMaxDrawdownDuration(t *testing.T) {
	cases := []struct {
		name    string
		curve   []model.EquityPoint
		wantDur time.Duration
	}{
		{
			name:    "empty curve",
			curve:   nil,
			wantDur: 0,
		},
		{
			name:    "single point",
			curve:   makeEquityCurve(100),
			wantDur: 0,
		},
		{
			name:    "all upward — no drawdown",
			curve:   makeEquityCurve(100, 110, 120),
			wantDur: 0,
		},
		{
			// Peak=100@t0, trough=80@t2, recovery=105@t4 (105 >= 100).
			// Duration = t4 - t0 = 4 hours.
			name:    "peak-trough-recovery",
			curve:   makeEquityCurve(100, 90, 80, 95, 105),
			wantDur: 4 * time.Hour,
		},
		{
			// Peak=100@t0, trough=80@t2, no recovery (95 < 100), last bar=t3.
			// Duration = t3 - t0 = 3 hours.
			name:    "peak-trough-no-recovery",
			curve:   makeEquityCurve(100, 90, 80, 95),
			wantDur: 3 * time.Hour,
		},
		{
			// First drawdown:  peak=100@t0, trough=90@t1, depth=10%, recovered@t2 (100>=100).
			// Second drawdown: peak=110@t3, trough=80@t4, depth≈27.3%, recovered@t5 (115>=110).
			// Max drawdown = second; duration = t5 - t3 = 2 hours.
			name:    "two drawdowns — max is second one",
			curve:   makeEquityCurve(100, 90, 100, 110, 80, 115),
			wantDur: 2 * time.Hour,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := analytics.Compute(nil, tc.curve, model.TimeframeDaily)
			if r.MaxDrawdownDuration != tc.wantDur {
				t.Errorf("MaxDrawdownDuration: got %v, want %v", r.MaxDrawdownDuration, tc.wantDur)
			}
		})
	}
}

// --- TailRatio ---

func TestCompute_TailRatio_Symmetric(t *testing.T) {
	// Curve [100, 110, 99, 108.9]; returns sorted ascending: [-0.1, +0.1, +0.1].
	// n=3; p5=sorted[⌊0.05·3⌋]=sorted[0]=-0.1; p95=sorted[⌊0.95·3⌋]=sorted[2]=+0.1.
	// TailRatio = 0.1 / |-0.1| = 1.0
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "TailRatio", 1.0, r.TailRatio)
}

func TestCompute_TailRatio_AllPositive(t *testing.T) {
	// All returns non-negative → p5 >= 0 → TailRatio = 0.
	curve := makeEquityCurve(100, 110, 121, 133.1)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "TailRatio", 0, r.TailRatio)
}

// --- helpers ---

func assertEqual(t *testing.T, field string, want, got int) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %d, want %d", field, got, want)
	}
}

const floatTolerance = 0.0001

func assertFloatEqual(t *testing.T, field string, want, got float64) {
	t.Helper()
	diff := got - want
	if diff < -floatTolerance || diff > floatTolerance {
		t.Errorf("%s: got %.4f, want %.4f", field, got, want)
	}
}
