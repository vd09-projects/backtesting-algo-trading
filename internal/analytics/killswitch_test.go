package analytics_test

import (
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// tradeReturn builds a trade where ReturnOnNotional == r exactly.
// EntryPrice=1, Quantity=1 → ReturnOnNotional = RealizedPnL / 1 = r.
func tradeReturn(r float64) model.Trade {
	return model.Trade{
		Instrument:  "NSE:TEST",
		Direction:   model.DirectionLong,
		Quantity:    1,
		EntryPrice:  1,
		ExitPrice:   1 + r,
		EntryTime:   baseTime,
		ExitTime:    baseTime.Add(time.Hour),
		RealizedPnL: r,
	}
}

func tradesFromReturns(rs ...float64) []model.Trade {
	trades := make([]model.Trade, len(rs))
	for i, r := range rs {
		trades[i] = tradeReturn(r)
	}
	return trades
}

// --- DeriveKillSwitchThresholds ---

func TestDeriveKillSwitchThresholds_Basic(t *testing.T) {
	t.Parallel()
	inSample := analytics.Report{
		MaxDrawdown:         10.0,
		MaxDrawdownDuration: 100 * 24 * time.Hour,
	}
	got := analytics.DeriveKillSwitchThresholds(-0.2, inSample)

	assertFloatEqual(t, "SharpeP5", -0.2, got.SharpeP5)
	assertFloatEqual(t, "MaxDrawdownPct", 15.0, got.MaxDrawdownPct)
	if got.MaxDDDuration != 200*24*time.Hour {
		t.Errorf("MaxDDDuration: got %v, want %v", got.MaxDDDuration, 200*24*time.Hour)
	}
}

func TestDeriveKillSwitchThresholds_ZeroDrawdown(t *testing.T) {
	t.Parallel()
	got := analytics.DeriveKillSwitchThresholds(0, analytics.Report{})
	assertFloatEqual(t, "MaxDrawdownPct", 0, got.MaxDrawdownPct)
	if got.MaxDDDuration != 0 {
		t.Errorf("MaxDDDuration: got %v, want 0", got.MaxDDDuration)
	}
}

func TestDeriveKillSwitchThresholds_LargeDuration(t *testing.T) {
	t.Parallel()
	// ~1137 days — close to the SMA crossover in-sample worst recovery (3.11 years).
	dur := time.Duration(1137 * 24 * int64(time.Hour))
	got := analytics.DeriveKillSwitchThresholds(0, analytics.Report{MaxDrawdownDuration: dur})
	want := time.Duration(float64(dur) * 2)
	if got.MaxDDDuration != want {
		t.Errorf("MaxDDDuration: got %v, want %v", got.MaxDDDuration, want)
	}
}

// --- CheckKillSwitch ---

func TestCheckKillSwitch_AllGreen(t *testing.T) {
	t.Parallel()
	// returns [0.2, 0.3, 0.2, 0.25] → positive per-trade Sharpe ≈ 4.96
	window := tradesFromReturns(0.2, 0.3, 0.2, 0.25)
	curve := makeEquityCurve(100, 110, 120)
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       -1.0,
		MaxDrawdownPct: 50.0,
		MaxDDDuration:  100 * time.Hour,
	}
	got := analytics.CheckKillSwitch(window, curve, thresholds)

	if got.SharpeBreached {
		t.Errorf("SharpeBreached: got true, want false (Sharpe %.4f)", got.RollingPerTradeSharpe)
	}
	if got.DrawdownBreached {
		t.Errorf("DrawdownBreached: got true, want false (DD %.4f%%)", got.CurrentDrawdownPct)
	}
	if got.DurationBreached {
		t.Errorf("DurationBreached: got true, want false (dur %v)", got.CurrentDDDuration)
	}
	assertFloatEqual(t, "CurrentDrawdownPct", 0, got.CurrentDrawdownPct)
	if got.CurrentDDDuration != 0 {
		t.Errorf("CurrentDDDuration: got %v, want 0", got.CurrentDDDuration)
	}
	if got.RollingPerTradeSharpe <= 0 {
		t.Errorf("RollingPerTradeSharpe: got %.4f, want positive", got.RollingPerTradeSharpe)
	}
}

func TestCheckKillSwitch_SharpeBreached(t *testing.T) {
	t.Parallel()
	// returns [-0.1, -0.15, -0.12, -0.08] → negative per-trade Sharpe ≈ -3.64
	window := tradesFromReturns(-0.1, -0.15, -0.12, -0.08)
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       0.0,
		MaxDrawdownPct: 50.0,
		MaxDDDuration:  100 * time.Hour,
	}
	got := analytics.CheckKillSwitch(window, makeEquityCurve(100, 110, 120), thresholds)

	if !got.SharpeBreached {
		t.Errorf("SharpeBreached: got false, want true (Sharpe %.4f < threshold 0.0)", got.RollingPerTradeSharpe)
	}
	if got.DrawdownBreached {
		t.Error("DrawdownBreached: got true, want false")
	}
	if got.DurationBreached {
		t.Error("DurationBreached: got true, want false")
	}
}

func TestCheckKillSwitch_DrawdownBreached(t *testing.T) {
	t.Parallel()
	window := tradesFromReturns(0.2, 0.3, 0.2, 0.25)
	// Peak=100@t0 drops to 40@t1 → current drawdown = 60%.
	curve := makeEquityCurve(100, 40)
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       -1.0,
		MaxDrawdownPct: 20.0,
		MaxDDDuration:  100 * time.Hour,
	}
	got := analytics.CheckKillSwitch(window, curve, thresholds)

	if got.SharpeBreached {
		t.Error("SharpeBreached: got true, want false")
	}
	if !got.DrawdownBreached {
		t.Errorf("DrawdownBreached: got false, want true (DD %.4f%% > threshold 20%%)", got.CurrentDrawdownPct)
	}
	if got.DurationBreached {
		t.Error("DurationBreached: got true, want false")
	}
}

func TestCheckKillSwitch_DurationBreached(t *testing.T) {
	t.Parallel()
	window := tradesFromReturns(0.2, 0.3, 0.2, 0.25)
	// Equity peaks at t1, still in drawdown at t2 → duration = 1 hour.
	curve := makeEquityCurve(100, 110, 99)
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       -1.0,
		MaxDrawdownPct: 50.0,
		MaxDDDuration:  30 * time.Minute,
	}
	got := analytics.CheckKillSwitch(window, curve, thresholds)

	if got.SharpeBreached {
		t.Error("SharpeBreached: got true, want false")
	}
	if got.DrawdownBreached {
		t.Error("DrawdownBreached: got true, want false")
	}
	if !got.DurationBreached {
		t.Errorf("DurationBreached: got false, want true (dur %v > threshold 30m)", got.CurrentDDDuration)
	}
}

func TestCheckKillSwitch_FewerThanTwoTrades(t *testing.T) {
	t.Parallel()
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       1.0, // high — would breach if Sharpe were computed
		MaxDrawdownPct: 50.0,
		MaxDDDuration:  100 * time.Hour,
	}
	got := analytics.CheckKillSwitch(tradesFromReturns(0.3), makeEquityCurve(100, 110), thresholds)

	if got.SharpeBreached {
		t.Error("SharpeBreached: got true, want false — cannot compute Sharpe with 1 trade")
	}
	assertFloatEqual(t, "RollingPerTradeSharpe", 0, got.RollingPerTradeSharpe)
}

func TestCheckKillSwitch_EmptyTrades(t *testing.T) {
	t.Parallel()
	thresholds := analytics.KillSwitchThresholds{SharpeP5: 1.0}
	got := analytics.CheckKillSwitch(nil, makeEquityCurve(100, 110), thresholds)

	if got.SharpeBreached {
		t.Error("SharpeBreached: got true, want false — no trades")
	}
	assertFloatEqual(t, "RollingPerTradeSharpe", 0, got.RollingPerTradeSharpe)
}

func TestCheckKillSwitch_ZeroDurationThreshold_NeverBreaches(t *testing.T) {
	t.Parallel()
	// MaxDDDuration=0 disables the duration threshold.
	window := tradesFromReturns(0.2, 0.3)
	curve := makeEquityCurve(100, 90) // in drawdown for 1 hour
	thresholds := analytics.KillSwitchThresholds{
		SharpeP5:       -1.0,
		MaxDrawdownPct: 50.0,
		MaxDDDuration:  0, // disabled
	}
	got := analytics.CheckKillSwitch(window, curve, thresholds)

	if got.DurationBreached {
		t.Error("DurationBreached: got true, want false — threshold 0 should disable check")
	}
}
