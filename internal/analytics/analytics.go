// Package analytics computes performance metrics from a completed trade log.
package analytics

import "github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"

// Report holds performance metrics computed from a completed trade log.
type Report struct {
	TotalPnL    float64
	WinRate     float64 // percentage 0–100
	MaxDrawdown float64 // peak-to-trough on equity curve, percentage 0–100
	TradeCount  int
	WinCount    int
	LossCount   int // includes break-even trades (RealizedPnL <= 0)
}

// Compute derives performance metrics from a slice of closed trades.
// It is a pure function — it does not modify the input slice.
func Compute(trades []model.Trade) Report {
	if len(trades) == 0 {
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
	r.WinRate = float64(r.WinCount) / float64(r.TradeCount) * 100

	return r
}
