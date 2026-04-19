package analytics_test

import (
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
	assertFloatEqual(t, "WinRate", 0, r.WinRate) // gated: 1 < MinTradesForMetrics
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
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
	assertFloatEqual(t, "WinRate", 0, r.WinRate) // gated: 2 < MinTradesForMetrics
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
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
	assertFloatEqual(t, "WinRate", 0, r.WinRate) // gated: 3 < MinTradesForMetrics
	assertFloatEqual(t, "MaxDrawdown", 50, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
}

func TestCompute_BreakevenCountsAsLoss(t *testing.T) {
	trades := []model.Trade{trade(100), trade(0)}
	r := analytics.Compute(trades, nil, "")

	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
	assertFloatEqual(t, "WinRate", 0, r.WinRate) // gated: 2 < MinTradesForMetrics
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
}

// TestComputeSharpe_SmallCurveGate verifies that any curve shorter than
// MinCurvePointsForMetrics produces SharpeRatio = 0 and CurveMetricsInsufficient = true.
// Mathematical correctness of the Sharpe formula is tested in TestComputeSharpeInternal
// (analytics_internal_test.go).
func TestComputeSharpe_SmallCurveGate(t *testing.T) {
	cases := []struct {
		name  string
		curve []model.EquityPoint
		tf    model.Timeframe
	}{
		{"empty curve", nil, model.TimeframeDaily},
		{"single point", makeEquityCurve(100), model.TimeframeDaily},
		{"two points", makeEquityCurve(100, 110), model.TimeframeDaily},
		{"constant equity", makeEquityCurve(100, 100, 100, 100), model.TimeframeDaily},
		{"non-trivial 4-point curve", makeEquityCurve(100, 110, 99, 108.9), model.TimeframeDaily},
		{"unknown timeframe", makeEquityCurve(100, 110, 99, 108.9), model.Timeframe("unknown")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := analytics.Compute(nil, tc.curve, tc.tf)
			assertFloatEqual(t, "SharpeRatio", 0, r.SharpeRatio)
			if !r.CurveMetricsInsufficient {
				t.Errorf("CurveMetricsInsufficient: got false for %d-point curve", len(tc.curve))
			}
		})
	}
}

// --- ProfitFactor, AvgWin, AvgLoss ---

func TestCompute_ProfitFactor_AllWinners(t *testing.T) {
	// No losing trades → ProfitFactor = 0 (guard against division by zero).
	// Trade metrics are gated (2 < MinTradesForMetrics), so all three are 0.
	r := analytics.Compute([]model.Trade{trade(100), trade(50)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 0, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 0, r.AvgLoss)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
}

func TestCompute_ProfitFactor_AllLosers(t *testing.T) {
	// No winning trades → ProfitFactor = 0, AvgWin = 0.
	// Trade metrics are gated (2 < MinTradesForMetrics), so AvgLoss is also 0.
	r := analytics.Compute([]model.Trade{trade(-30), trade(-20)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 0, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 0, r.AvgLoss)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
}

func TestCompute_ProfitFactor_Mixed(t *testing.T) {
	// grossProfit = 100+50 = 150; grossLoss = 30+20 = 50 → PF = 3.0, AvgWin = 75, AvgLoss = 25.
	// Trade metrics are gated (4 < MinTradesForMetrics), so all three are 0.
	r := analytics.Compute([]model.Trade{trade(100), trade(50), trade(-30), trade(-20)}, nil, "")
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 0, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 0, r.AvgLoss)
	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
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

func TestCompute_SortinoRatio_SmallCurveGate(t *testing.T) {
	// 4-point curve fires CurveMetricsInsufficient → SortinoRatio = 0.
	// Internal math tested in TestComputeSortinoInternal.
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "SortinoRatio", 0, r.SortinoRatio)
	if !r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got false, want true")
	}
}

// --- CalmarRatio ---

func TestCompute_CalmarRatio_SmallCurveGate(t *testing.T) {
	// 4-point curve fires CurveMetricsInsufficient → CalmarRatio = 0.
	// Internal math tested in TestComputeCalmarInternal.
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "CalmarRatio", 0, r.CalmarRatio)
	if !r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got false, want true")
	}
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

func TestCompute_TailRatio_SmallCurveGate(t *testing.T) {
	// 4-point curve fires CurveMetricsInsufficient → TailRatio = 0.
	// Internal math tested in TestComputeTailRatioInternal.
	curve := makeEquityCurve(100, 110, 99, 108.9)
	r := analytics.Compute(nil, curve, model.TimeframeDaily)
	assertFloatEqual(t, "TailRatio", 0, r.TailRatio)
	if !r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got false, want true")
	}
}

// --- Signal frequency gate ---

// makeNTrades builds n identical trades with the given pnl.
func makeNTrades(n int, pnl float64) []model.Trade {
	trades := make([]model.Trade, n)
	for i := range trades {
		trades[i] = trade(pnl)
	}
	return trades
}

// makeAltCurve builds n equity points alternating +1%/-0.5% returns, producing
// non-zero variance (so curve metrics are non-trivially computable when the gate passes).
func makeAltCurve(n int) []model.EquityPoint {
	pts := make([]model.EquityPoint, n)
	eq := 100_000.0
	for i := range pts {
		pts[i] = model.EquityPoint{
			Timestamp: baseTime.Add(time.Duration(i) * 24 * time.Hour),
			Value:     eq,
		}
		if i%2 == 0 {
			eq *= 1.01
		} else {
			eq *= 0.995
		}
	}
	return pts
}

func TestCompute_TradeMetricsInsufficient(t *testing.T) {
	// 7 trades (< 30) with a sufficient curve (252 pts):
	// trade metrics zeroed, curve metrics computable.
	r := analytics.Compute(makeNTrades(7, 100), makeAltCurve(252), model.TimeframeDaily)

	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
	if r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got true, want false")
	}
	assertFloatEqual(t, "WinRate", 0, r.WinRate)
	assertFloatEqual(t, "ProfitFactor", 0, r.ProfitFactor)
	assertFloatEqual(t, "AvgWin", 0, r.AvgWin)
	assertFloatEqual(t, "AvgLoss", 0, r.AvgLoss)
	// Non-trade fields are unaffected.
	assertEqual(t, "TradeCount", 7, r.TradeCount)
}

func TestCompute_CurveMetricsInsufficient(t *testing.T) {
	// 30 trades (>= 30) with a short curve (10 pts):
	// curve metrics zeroed, trade metrics computable.
	r := analytics.Compute(makeNTrades(30, 100), makeEquityCurve(100, 110, 99, 108.9, 120, 115, 130, 125, 140, 135), model.TimeframeDaily)

	if r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got true, want false")
	}
	if !r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got false, want true")
	}
	assertFloatEqual(t, "SharpeRatio", 0, r.SharpeRatio)
	assertFloatEqual(t, "SortinoRatio", 0, r.SortinoRatio)
	assertFloatEqual(t, "CalmarRatio", 0, r.CalmarRatio)
	assertFloatEqual(t, "TailRatio", 0, r.TailRatio)
	// Trade metrics are reported normally with 30 trades.
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
}

func TestCompute_BothMetricsInsufficient(t *testing.T) {
	// 7 trades + 4-point curve: both gates fire.
	r := analytics.Compute(makeNTrades(7, 100), makeEquityCurve(100, 110, 99, 108.9), model.TimeframeDaily)

	if !r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got false, want true")
	}
	if !r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got false, want true")
	}
}

func TestCompute_SufficientData_NoFlags(t *testing.T) {
	// 30 trades + 252-point curve: neither gate fires.
	r := analytics.Compute(makeNTrades(30, 100), makeAltCurve(252), model.TimeframeDaily)

	if r.TradeMetricsInsufficient {
		t.Error("TradeMetricsInsufficient: got true, want false")
	}
	if r.CurveMetricsInsufficient {
		t.Error("CurveMetricsInsufficient: got true, want false")
	}
	// Both groups are populated.
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
	if r.SharpeRatio == 0 {
		t.Error("SharpeRatio: got 0, want non-zero for sufficient alternating curve")
	}
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
