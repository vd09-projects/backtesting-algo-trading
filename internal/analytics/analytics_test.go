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

func TestCompute_Empty(t *testing.T) {
	r := analytics.Compute(nil)

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
	r := analytics.Compute([]model.Trade{trade(100)})

	assertEqual(t, "TradeCount", 1, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 100, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
}

func TestCompute_SingleLoser(t *testing.T) {
	r := analytics.Compute([]model.Trade{trade(-50)})

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
	r := analytics.Compute(trades)

	assertEqual(t, "TradeCount", 2, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 300, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 100, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 0, r.LossCount)
}

func TestCompute_AllLosers(t *testing.T) {
	trades := []model.Trade{trade(-50), trade(-50)}
	r := analytics.Compute(trades)

	assertEqual(t, "TradeCount", 2, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", -100, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 0, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
	assertEqual(t, "WinCount", 0, r.WinCount)
	assertEqual(t, "LossCount", 2, r.LossCount)
}

func TestCompute_Mixed(t *testing.T) {
	// Equity curve: +200 → 200, -100 → 100, +50 → 150
	// Peak = 200, trough = 100 → MaxDrawdown = (200-100)/200 * 100 = 50%
	trades := []model.Trade{trade(200), trade(-100), trade(50)}
	r := analytics.Compute(trades)

	assertEqual(t, "TradeCount", 3, r.TradeCount)
	assertFloatEqual(t, "TotalPnL", 150, r.TotalPnL)
	assertFloatEqual(t, "WinRate", 66.6667, r.WinRate)
	assertFloatEqual(t, "MaxDrawdown", 50, r.MaxDrawdown)
	assertEqual(t, "WinCount", 2, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
}

func TestCompute_BreakevenCountsAsLoss(t *testing.T) {
	trades := []model.Trade{trade(100), trade(0)}
	r := analytics.Compute(trades)

	assertEqual(t, "WinCount", 1, r.WinCount)
	assertEqual(t, "LossCount", 1, r.LossCount)
	assertFloatEqual(t, "WinRate", 50, r.WinRate)
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
