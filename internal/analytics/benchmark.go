package analytics

import (
	"math"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// BenchmarkReport holds buy-and-hold performance metrics over a given instrument and period.
type BenchmarkReport struct {
	TotalReturn      float64 // percentage, e.g. 12.5 means 12.5%
	AnnualizedReturn float64 // CAGR percentage
	SharpeRatio      float64 // annualized, using per-bar close returns
	MaxDrawdown      float64 // peak-to-trough percentage
}

// ComputeBenchmark computes buy-and-hold metrics over the provided candle series.
// Entry is at candles[0].Open; exit is at candles[last].Close. No transaction costs.
// initialCash is the notional capital — it scales the equity curve but does not affect
// return percentages or Sharpe ratio.
// Returns a zero BenchmarkReport if candles has fewer than 2 bars.
func ComputeBenchmark(candles []model.Candle, initialCash float64) BenchmarkReport {
	if len(candles) < 2 {
		return BenchmarkReport{}
	}

	firstOpen := candles[0].Open
	lastClose := candles[len(candles)-1].Close

	totalReturn := (lastClose - firstOpen) / firstOpen * 100

	// Equity curve: position value at each bar's close.
	curve := make([]model.EquityPoint, len(candles))
	for i, c := range candles {
		curve[i] = model.EquityPoint{
			Timestamp: c.Timestamp,
			Value:     initialCash * c.Close / firstOpen,
		}
	}

	// Max drawdown from equity curve.
	var peak, maxDD float64
	for _, pt := range curve {
		if pt.Value > peak {
			peak = pt.Value
		}
		if peak > 0 {
			dd := (peak - pt.Value) / peak * 100
			if dd > maxDD {
				maxDD = dd
			}
		}
	}

	// CAGR: annualize the total return by the actual elapsed time.
	var annualizedReturn float64
	duration := candles[len(candles)-1].Timestamp.Sub(candles[0].Timestamp)
	years := duration.Hours() / (365.25 * 24)
	if years > 0 {
		annualizedReturn = (math.Pow(1+totalReturn/100, 1/years) - 1) * 100
	}

	// Sharpe ratio from the close-based equity curve, using the first candle's timeframe.
	sharpe := computeSharpe(computeReturns(curve), candles[0].Timeframe)

	return BenchmarkReport{
		TotalReturn:      totalReturn,
		AnnualizedReturn: annualizedReturn,
		SharpeRatio:      sharpe,
		MaxDrawdown:      maxDD,
	}
}
