// Package analytics computes performance metrics from a completed trade log.
package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Report holds performance metrics computed from a completed trade log and equity curve.
type Report struct {
	TotalPnL            float64
	WinRate             float64 // percentage 0–100
	MaxDrawdown         float64 // peak-to-trough on equity curve, percentage 0–100
	TradeCount          int
	WinCount            int
	LossCount           int           // includes break-even trades (RealizedPnL <= 0)
	SharpeRatio         float64       // annualized Sharpe ratio from per-bar equity curve returns
	ProfitFactor        float64       // grossProfit / grossLoss; 0 if no losing trades
	AvgWin              float64       // average P&L of winning trades; 0 if no winning trades
	AvgLoss             float64       // average absolute P&L of losing trades; 0 if no losing trades
	SortinoRatio        float64       // annualized Sortino ratio (downside deviation only)
	CalmarRatio         float64       // annualized return / max drawdown (decimal); 0 if max drawdown is zero
	TailRatio           float64       // p95 return / |p5 return|; 0 if p5 return >= 0
	MaxDrawdownDuration time.Duration // wall time from the max-drawdown peak to first recovery (or last bar if never recovered)
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
	}

	r.MaxDrawdown = computeMaxDrawdownDepth(curve)
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
	r.MaxDrawdownDuration = computeMaxDrawdownDuration(curve)

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

// computeMaxDrawdownDepth returns the peak-to-trough drawdown as a percentage (0–100)
// from the per-bar equity curve. Returns 0 for fewer than 2 points.
func computeMaxDrawdownDepth(curve []model.EquityPoint) float64 {
	if len(curve) < 2 {
		return 0
	}
	peak := curve[0].Value
	var maxDD float64
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
	return maxDD
}

// computeMaxDrawdownDuration returns the wall-clock duration from the peak that
// precedes the maximum drawdown to the first subsequent bar where equity recovers
// to that peak value, or to the last bar if the equity never recovers.
//
// The peak and drawdown depth are derived from the per-bar equity curve
// ([]model.EquityPoint), not from the trade-based P&L accumulation that drives
// Report.MaxDrawdown. In practice the two will agree closely for strategies with
// frequent trades, but may diverge for low-turnover strategies where the per-bar
// curve captures intra-trade mark-to-market swings that the closed-trade curve misses.
//
// Both MaxDrawdown (depth %) and MaxDrawdownDuration are now computed from the
// same per-bar EquityPoint curve, so they always describe the same drawdown event.
func computeMaxDrawdownDuration(curve []model.EquityPoint) time.Duration {
	if len(curve) < 2 {
		return 0
	}

	var (
		peakIdx   int
		peakValue = curve[0].Value
		maxDDPct  float64
		ddPeakIdx int
	)

	for i, pt := range curve {
		if pt.Value > peakValue {
			peakValue = pt.Value
			peakIdx = i
		}
		if peakValue > 0 {
			dd := (peakValue - pt.Value) / peakValue * 100
			if dd > maxDDPct {
				maxDDPct = dd
				ddPeakIdx = peakIdx
			}
		}
	}

	if maxDDPct == 0 {
		return 0
	}

	ddPeakValue := curve[ddPeakIdx].Value
	ddPeakTime := curve[ddPeakIdx].Timestamp

	for i := ddPeakIdx + 1; i < len(curve); i++ {
		if curve[i].Value >= ddPeakValue {
			return curve[i].Timestamp.Sub(ddPeakTime)
		}
	}

	// Equity never recovered — duration runs to the last bar.
	return curve[len(curve)-1].Timestamp.Sub(ddPeakTime)
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
