package analytics

import (
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Regime defines a named calendar window for performance attribution.
type Regime struct {
	Name string
	From time.Time // inclusive
	To   time.Time // exclusive
}

// RegimeReport holds per-regime performance metrics.
type RegimeReport struct {
	Name        string
	From        time.Time
	To          time.Time
	SharpeRatio float64 // annualized; 0 when curve has fewer than 2 points
	MaxDrawdown float64 // peak-to-trough %, 0–100
}

// NSERegimes2018_2024 are the three distinct market regimes for NSE
// over the 2018–2024 backtest window used for baseline strategy evaluation.
var NSERegimes2018_2024 = []Regime{
	{
		Name: "Pre-COVID (2018–2019)",
		From: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "COVID Crash + Recovery (2020–2021)",
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "Grind (2022–2024)",
		From: time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// ComputeRegimeSplits computes per-regime Sharpe and max drawdown by slicing the
// equity curve into each regime's [From, To) window. tf is required to annualize
// Sharpe correctly (the curve carries no bar-frequency information).
// Returns one RegimeReport per input regime, in the same order, with zeroed metrics
// for any regime that has fewer than 2 curve points.
func ComputeRegimeSplits(curve []model.EquityPoint, regimes []Regime, tf model.Timeframe) []RegimeReport {
	reports := make([]RegimeReport, len(regimes))
	for i, regime := range regimes {
		reports[i] = RegimeReport{
			Name: regime.Name,
			From: regime.From,
			To:   regime.To,
		}

		var slice []model.EquityPoint
		for _, pt := range curve {
			if (pt.Timestamp.Equal(regime.From) || pt.Timestamp.After(regime.From)) &&
				pt.Timestamp.Before(regime.To) {
				slice = append(slice, pt)
			}
		}

		returns := computeReturns(slice)
		reports[i].SharpeRatio = computeSharpe(returns, tf)
		reports[i].MaxDrawdown = computeMaxDrawdownDepth(slice)
	}
	return reports
}
