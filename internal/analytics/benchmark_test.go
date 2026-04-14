package analytics_test

import (
	"math"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// benchCandle builds a minimal valid Candle for benchmark tests.
// High = max(open, close), Low = min(open, close).
func benchCandle(tf model.Timeframe, ts time.Time, open, cls float64) model.Candle {
	high := open
	if cls > high {
		high = cls
	}
	low := open
	if cls < low {
		low = cls
	}
	return model.Candle{
		Instrument: "TEST:INSTR",
		Timeframe:  tf,
		Timestamp:  ts,
		Open:       open,
		High:       high,
		Low:        low,
		Close:      cls,
		Volume:     100,
	}
}

var benchBase = time.Date(2024, 1, 1, 9, 0, 0, 0, time.UTC)

func dailyCandle(day int, open, cls float64) model.Candle {
	return benchCandle(model.TimeframeDaily, benchBase.Add(time.Duration(day)*24*time.Hour), open, cls)
}

func TestComputeBenchmark_Empty(t *testing.T) {
	r := analytics.ComputeBenchmark(nil, 10000)
	if r.TotalReturn != 0 || r.AnnualizedReturn != 0 || r.SharpeRatio != 0 || r.MaxDrawdown != 0 {
		t.Errorf("expected zero report for nil candles, got %+v", r)
	}
}

func TestComputeBenchmark_SingleCandle(t *testing.T) {
	candles := []model.Candle{dailyCandle(0, 100, 110)}
	r := analytics.ComputeBenchmark(candles, 10000)
	if r.TotalReturn != 0 || r.AnnualizedReturn != 0 || r.SharpeRatio != 0 || r.MaxDrawdown != 0 {
		t.Errorf("expected zero report for single candle, got %+v", r)
	}
}

func TestComputeBenchmark_TotalReturn(t *testing.T) {
	// Buy at 100 (first open), sell at 130 (last close) → 30% total return.
	candles := []model.Candle{
		dailyCandle(0, 100, 100),
		dailyCandle(1, 130, 130),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "TotalReturn", 30, r.TotalReturn)
}

func TestComputeBenchmark_MaxDrawdown(t *testing.T) {
	// Closes: 100, 120, 90, 110 with firstOpen=100.
	// Equity (initialCash=10000): 10000, 12000, 9000, 11000.
	// Peak=12000, trough=9000 → (12000-9000)/12000*100 = 25%.
	candles := []model.Candle{
		dailyCandle(0, 100, 100),
		dailyCandle(1, 120, 120),
		dailyCandle(2, 90, 90),
		dailyCandle(3, 110, 110),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "MaxDrawdown", 25, r.MaxDrawdown)
}

func TestComputeBenchmark_AnnualizedReturn(t *testing.T) {
	// 44% total return over exactly 2 years → CAGR = sqrt(1.44)-1 = 20%.
	// 2 years = 2 * 365.25 * 24 hours = 17532 hours.
	twoYearsLater := benchBase.Add(17532 * time.Hour)
	candles := []model.Candle{
		benchCandle(model.TimeframeDaily, benchBase, 100, 100),
		benchCandle(model.TimeframeDaily, twoYearsLater, 144, 144),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "TotalReturn", 44, r.TotalReturn)
	assertFloatEqual(t, "AnnualizedReturn", 20, r.AnnualizedReturn)
}

func TestComputeBenchmark_AnnualizedReturn_UnderOneYear(t *testing.T) {
	// 10% return over 0.5 years → CAGR = (1.10^2 - 1) * 100 = 21%.
	// 0.5 years = 0.5 * 8766 = 4383 hours.
	halfYearLater := benchBase.Add(4383 * time.Hour)
	candles := []model.Candle{
		benchCandle(model.TimeframeDaily, benchBase, 100, 100),
		benchCandle(model.TimeframeDaily, halfYearLater, 110, 110),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "TotalReturn", 10, r.TotalReturn)
	want := (1.10*1.10 - 1) * 100 // ≈ 21%
	assertFloatEqual(t, "AnnualizedReturn", want, r.AnnualizedReturn)
}

func TestComputeBenchmark_Sharpe(t *testing.T) {
	// Close sequence [100, 110, 99, 108.9] with firstOpen=100.
	// Per-bar returns: [+0.1, -0.1, +0.1] → daily Sharpe = √21 ≈ 4.5826.
	candles := []model.Candle{
		dailyCandle(0, 100, 100),
		dailyCandle(1, 110, 110),
		dailyCandle(2, 99, 99),
		dailyCandle(3, 108.9, 108.9),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "SharpeRatio", math.Sqrt(21), r.SharpeRatio)
}

func TestComputeBenchmark_ZeroInitialCash(t *testing.T) {
	// initialCash=0 → all equity points are zero → computeSharpe prev==0 branch fires.
	// Variance is zero, so Sharpe must be 0; other metrics still compute normally.
	candles := []model.Candle{
		dailyCandle(0, 100, 100),
		dailyCandle(1, 110, 110),
		dailyCandle(2, 99, 99),
		dailyCandle(3, 108.9, 108.9),
	}
	r := analytics.ComputeBenchmark(candles, 0)
	assertFloatEqual(t, "TotalReturn", 8.9, r.TotalReturn)
	assertFloatEqual(t, "SharpeRatio", 0, r.SharpeRatio)
}

func TestComputeBenchmark_NoDrawdown(t *testing.T) {
	// Monotonically increasing closes → no drawdown.
	candles := []model.Candle{
		dailyCandle(0, 100, 100),
		dailyCandle(1, 110, 110),
		dailyCandle(2, 120, 120),
	}
	r := analytics.ComputeBenchmark(candles, 10000)
	assertFloatEqual(t, "MaxDrawdown", 0, r.MaxDrawdown)
}
