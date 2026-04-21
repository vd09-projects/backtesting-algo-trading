package analytics

import (
	"math"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// KillSwitchThresholds holds pre-committed halt conditions for a live strategy.
// All fields must be set before a strategy goes live and must not be changed
// while the strategy is in a drawdown — doing so is live overfitting.
type KillSwitchThresholds struct {
	// SharpeP5 is the 5th-percentile per-trade Sharpe from the bootstrap run.
	// Halt when the rolling-window per-trade Sharpe drops below this value.
	SharpeP5 float64
	// MaxDrawdownPct is 1.5× the worst in-sample drawdown (percent, 0–100).
	// Halt when the current drawdown from the all-time equity peak exceeds this value.
	MaxDrawdownPct float64
	// MaxDDDuration is 2× the worst in-sample drawdown recovery duration.
	// Halt when the strategy remains in drawdown longer than this without recovery.
	MaxDDDuration time.Duration
}

// KillSwitchAlert reports whether live metrics have approached or breached
// kill-switch thresholds.
type KillSwitchAlert struct {
	// SharpeBreached is true when rolling per-trade Sharpe is below KillSwitchThresholds.SharpeP5.
	SharpeBreached bool
	// DrawdownBreached is true when current drawdown exceeds KillSwitchThresholds.MaxDrawdownPct.
	DrawdownBreached bool
	// DurationBreached is true when current drawdown duration exceeds KillSwitchThresholds.MaxDDDuration.
	DurationBreached bool
	// RollingPerTradeSharpe is the per-trade Sharpe from the window trades.
	// Zero when fewer than 2 window trades are provided.
	RollingPerTradeSharpe float64
	// CurrentDrawdownPct is the current drawdown from the all-time equity peak to the last bar (0–100).
	CurrentDrawdownPct float64
	// CurrentDDDuration is the time since the most recent all-time equity high.
	// Zero when equity is at or above its all-time high.
	CurrentDDDuration time.Duration
}

// DeriveKillSwitchThresholds builds KillSwitchThresholds from a bootstrap p5 Sharpe
// and the completed in-sample Report. Pass montecarlo.BootstrapResult.SharpeP5 directly —
// this function is agnostic of the montecarlo package to keep analytics free of simulation
// dependencies.
func DeriveKillSwitchThresholds(sharpeP5 float64, inSample Report) KillSwitchThresholds { //nolint:gocritic // Report is a caller-constructed value type; pointer would leak internals
	return KillSwitchThresholds{
		SharpeP5:       sharpeP5,
		MaxDrawdownPct: inSample.MaxDrawdown * 1.5,
		MaxDDDuration:  time.Duration(float64(inSample.MaxDrawdownDuration) * 2),
	}
}

// CheckKillSwitch evaluates whether rolling live metrics have approached or breached
// kill-switch thresholds. windowTrades is the caller's rolling window of recent closed
// trades (the caller controls the window size — typically the last 6 months of trades).
// curve is the current equity curve.
//
// Per-trade Sharpe uses mean(ReturnOnNotional) / std(ReturnOnNotional) with sample
// variance (n-1), no annualization — identical to the montecarlo bootstrap computation
// per the algorithm decision recorded during TASK-0024.
func CheckKillSwitch(windowTrades []model.Trade, curve []model.EquityPoint, thresholds KillSwitchThresholds) KillSwitchAlert {
	var alert KillSwitchAlert

	if len(windowTrades) >= 2 {
		returns := make([]float64, len(windowTrades))
		for i, t := range windowTrades {
			returns[i] = t.ReturnOnNotional()
		}
		alert.RollingPerTradeSharpe = computePerTradeSharpe(returns)
		alert.SharpeBreached = alert.RollingPerTradeSharpe < thresholds.SharpeP5
	}

	alert.CurrentDrawdownPct = computeCurrentDrawdownDepth(curve)
	alert.DrawdownBreached = alert.CurrentDrawdownPct > thresholds.MaxDrawdownPct

	alert.CurrentDDDuration = computeCurrentDDDuration(curve)
	alert.DurationBreached = thresholds.MaxDDDuration > 0 && alert.CurrentDDDuration > thresholds.MaxDDDuration

	return alert
}

// computePerTradeSharpe returns mean(r) / std(r) with sample variance (n-1).
// Returns 0 for fewer than 2 values or zero variance.
// Identical formula to montecarlo.sampleSharpe (TASK-0024 algorithm decision).
func computePerTradeSharpe(returns []float64) float64 {
	n := float64(len(returns))
	if n < 2 {
		return 0
	}
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
	variance := sumSqDev / (n - 1)
	if variance == 0 {
		return 0
	}
	return mean / math.Sqrt(variance)
}

// computeCurrentDrawdownDepth returns the current drawdown depth as a percentage (0–100):
// the drop from the all-time equity peak to the last bar.
// Returns 0 for fewer than 2 points or when equity is at or above its all-time high.
func computeCurrentDrawdownDepth(curve []model.EquityPoint) float64 {
	if len(curve) < 2 {
		return 0
	}
	peak := curve[0].Value
	for _, pt := range curve {
		if pt.Value > peak {
			peak = pt.Value
		}
	}
	last := curve[len(curve)-1].Value
	if peak == 0 || last >= peak {
		return 0
	}
	return (peak - last) / peak * 100
}

// computeCurrentDDDuration returns the time elapsed since the most recent all-time
// equity high. Returns 0 when equity is at or above its all-time high.
func computeCurrentDDDuration(curve []model.EquityPoint) time.Duration {
	if len(curve) < 2 {
		return 0
	}
	peak := curve[0].Value
	peakTime := curve[0].Timestamp
	for _, pt := range curve {
		if pt.Value > peak {
			peak = pt.Value
			peakTime = pt.Timestamp
		}
	}
	last := curve[len(curve)-1]
	if last.Value >= peak {
		return 0
	}
	return last.Timestamp.Sub(peakTime)
}
