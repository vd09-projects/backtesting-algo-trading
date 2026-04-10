// Package analytics computes performance metrics from a completed trade log.
package analytics

import (
	"math"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Report holds performance metrics computed from a completed trade log and equity curve.
type Report struct {
	TotalPnL    float64
	WinRate     float64 // percentage 0–100
	MaxDrawdown float64 // peak-to-trough on equity curve, percentage 0–100
	TradeCount  int
	WinCount    int
	LossCount   int     // includes break-even trades (RealizedPnL <= 0)
	SharpeRatio float64 // annualized Sharpe ratio from per-bar equity curve returns
}

// Compute derives performance metrics from a slice of closed trades and the equity curve.
// tf is the bar timeframe, used to annualize the Sharpe ratio.
// It is a pure function — it does not modify the input slices.
func Compute(trades []model.Trade, curve []model.EquityPoint, tf model.Timeframe) Report {
	if len(trades) == 0 && len(curve) == 0 {
		return Report{}
	}

	var r Report
	r.TradeCount = len(trades)

	var equity, peak, maxDD float64

	for _, t := range trades {
		r.TotalPnL += t.RealizedPnL

		if t.RealizedPnL > 0 {
			r.WinCount++
		} else {
			r.LossCount++
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

	r.SharpeRatio = computeSharpe(curve, tf)

	return r
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

// computeSharpe computes the annualized Sharpe ratio from the equity curve.
// Returns 0 for fewer than 3 equity points (need at least 2 returns for sample stddev),
// zero-variance equity curves, or unknown timeframes.
func computeSharpe(curve []model.EquityPoint, tf model.Timeframe) float64 {
	if len(curve) < 3 {
		return 0
	}

	annFactor := sharpeAnnualizationFactor(tf)
	if annFactor == 0 {
		return 0
	}

	// Compute per-bar returns: r[i] = (curve[i+1] - curve[i]) / curve[i]
	returns := make([]float64, len(curve)-1)
	for i := range returns {
		prev := curve[i].Value
		if prev == 0 {
			returns[i] = 0
			continue
		}
		returns[i] = (curve[i+1].Value - prev) / prev
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
	// Sample variance (n-1 denominator). len(returns) >= 2 is guaranteed by len(curve) >= 3.
	variance := sumSqDev / (n - 1)
	if variance == 0 {
		return 0
	}

	return mean / math.Sqrt(variance) * math.Sqrt(annFactor)
}
