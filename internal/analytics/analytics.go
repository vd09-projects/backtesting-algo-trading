// Package analytics computes performance metrics from a completed trade log.
package analytics

import (
	"math"
	"sort"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Report holds performance metrics computed from a completed trade log and equity curve.
type Report struct {
	TotalPnL     float64
	WinRate      float64 // percentage 0–100
	MaxDrawdown  float64 // peak-to-trough on equity curve, percentage 0–100
	TradeCount   int
	WinCount     int
	LossCount    int     // includes break-even trades (RealizedPnL <= 0)
	SharpeRatio  float64 // annualized Sharpe ratio from per-bar equity curve returns
	ProfitFactor float64 // grossProfit / grossLoss; 0 if no losing trades
	AvgWin       float64 // average P&L of winning trades; 0 if no winning trades
	AvgLoss      float64 // average absolute P&L of losing trades; 0 if no losing trades
	SortinoRatio float64 // annualized Sortino ratio (downside deviation only)
	CalmarRatio  float64 // annualized return / max drawdown (decimal); 0 if max drawdown is zero
	TailRatio    float64 // p95 return / |p5 return|; 0 if p5 return >= 0
}

// Compute derives performance metrics from a slice of closed trades and the equity curve.
// tf is the bar timeframe, used to annualize return-based metrics.
// It is a pure function — it does not modify the input slices.
func Compute(trades []model.Trade, curve []model.EquityPoint, tf model.Timeframe) Report {
	if len(trades) == 0 && len(curve) == 0 {
		return Report{}
	}

	var r Report
	r.TradeCount = len(trades)

	var equity, peak, maxDD float64
	var grossProfit, grossLoss float64

	for _, t := range trades {
		r.TotalPnL += t.RealizedPnL

		if t.RealizedPnL > 0 {
			r.WinCount++
			grossProfit += t.RealizedPnL
		} else {
			r.LossCount++
			grossLoss += t.RealizedPnL // negative
		}

		equity += t.RealizedPnL
		if equity > peak {
			peak = equity
		}
		if peak > 0 {
			dd := (peak - equity) / peak * 100
			if dd > maxDD {
				maxDD = dd
			}
		}
	}

	r.MaxDrawdown = maxDD
	if r.TradeCount > 0 {
		r.WinRate = float64(r.WinCount) / float64(r.TradeCount) * 100
	}
	if r.LossCount > 0 {
		r.ProfitFactor = grossProfit / math.Abs(grossLoss)
		r.AvgLoss = math.Abs(grossLoss) / float64(r.LossCount)
	}
	if r.WinCount > 0 {
		r.AvgWin = grossProfit / float64(r.WinCount)
	}

	returns := computeReturns(curve)
	r.SharpeRatio = computeSharpe(returns, tf)
	r.SortinoRatio = computeSortino(returns, tf)
	r.CalmarRatio = computeCalmar(returns, tf, r.MaxDrawdown)
	r.TailRatio = computeTailRatio(returns)

	return r
}

// computeReturns derives per-bar returns from the equity curve.
// Returns nil for fewer than 2 equity points.
func computeReturns(curve []model.EquityPoint) []float64 {
	if len(curve) < 2 {
		return nil
	}
	returns := make([]float64, len(curve)-1)
	for i := range returns {
		prev := curve[i].Value
		if prev == 0 {
			returns[i] = 0
			continue
		}
		returns[i] = (curve[i+1].Value - prev) / prev
	}
	return returns
}

// sharpeAnnualizationFactor returns the number of bars per year for the given timeframe.
// NSE trading session is 9:15 AM – 3:30 PM = 375 minutes per day.
//
// Note: a Sharpe ratio > 2.5 on daily bars for a non-HFT strategy is a red flag —
// it likely indicates overfitting or an insufficient sample size.
func sharpeAnnualizationFactor(tf model.Timeframe) float64 {
	switch tf {
	case model.TimeframeWeekly:
		return 52
	case model.TimeframeDaily:
		return 252
	case model.Timeframe15Min:
		return 252 * 25 // NSE: 375 min/day ÷ 15 min/bar = 25 bars/day
	case model.Timeframe5Min:
		return 252 * 75 // NSE: 375 min/day ÷ 5 min/bar = 75 bars/day
	case model.Timeframe1Min:
		return 252 * 375 // NSE: 375 bars/day
	default:
		return 0
	}
}

// computeSharpe computes the annualized Sharpe ratio from pre-computed per-bar returns.
// Returns 0 for fewer than 2 returns (need at least 2 for sample stddev),
// zero-variance sequences, or unknown timeframes.
func computeSharpe(returns []float64, tf model.Timeframe) float64 {
	if len(returns) < 2 {
		return 0
	}

	annFactor := sharpeAnnualizationFactor(tf)
	if annFactor == 0 {
		return 0
	}

	n := float64(len(returns))
	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / n

	var sumSqDev float64
	for _, r := range returns {
		d := r - mean
		sumSqDev += d * d
	}
	// Sample variance (n-1 denominator).
	variance := sumSqDev / (n - 1)
	if variance == 0 {
		return 0
	}

	return mean / math.Sqrt(variance) * math.Sqrt(annFactor)
}

// computeSortino computes the annualized Sortino ratio from pre-computed per-bar returns.
// Downside deviation uses population-style denominator over all observations (not just negative
// ones), which is the most common convention for Sortino. Target return is 0.
// Returns 0 for fewer than 2 returns, zero downside deviation, or unknown timeframes.
func computeSortino(returns []float64, tf model.Timeframe) float64 {
	if len(returns) < 2 {
		return 0
	}

	annFactor := sharpeAnnualizationFactor(tf)
	if annFactor == 0 {
		return 0
	}

	n := float64(len(returns))
	var sum, sumSqDown float64
	for _, r := range returns {
		sum += r
		d := math.Min(r, 0)
		sumSqDown += d * d
	}
	mean := sum / n

	downsideDev := math.Sqrt(sumSqDown / n)
	if downsideDev == 0 {
		return 0
	}

	return mean / downsideDev * math.Sqrt(annFactor)
}

// computeCalmar computes the Calmar ratio: annualized return (decimal) / max drawdown (decimal).
// maxDrawdownPct is in the 0–100 range as stored in Report.MaxDrawdown.
// Returns 0 for fewer than 2 returns, zero max drawdown, or unknown timeframes.
func computeCalmar(returns []float64, tf model.Timeframe, maxDrawdownPct float64) float64 {
	if len(returns) < 2 || maxDrawdownPct == 0 {
		return 0
	}

	annFactor := sharpeAnnualizationFactor(tf)
	if annFactor == 0 {
		return 0
	}

	var sum float64
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))
	annReturn := mean * annFactor

	return annReturn / (maxDrawdownPct / 100)
}

// computeTailRatio computes the ratio of the 95th to the absolute 5th percentile of
// per-bar returns. A value < 1 indicates the left tail is heavier than the right —
// the strategy is short-vol in disguise.
// Returns 0 for fewer than 2 returns or a non-negative 5th percentile.
func computeTailRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	sort.Float64s(sorted)

	n := len(sorted)
	p5 := sorted[int(math.Floor(0.05*float64(n)))]
	p95 := sorted[int(math.Floor(0.95*float64(n)))]

	if p5 >= 0 {
		return 0
	}

	return p95 / math.Abs(p5)
}
