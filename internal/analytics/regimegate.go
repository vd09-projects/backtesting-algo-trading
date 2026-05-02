package analytics

import (
	"math"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// NSERegimesGate are the three NSE market regimes used for the universe sweep
// regime gate (TASK-0052). Windows match the 2026-04-27 decision file exactly.
//
// These differ from NSERegimes2018_2024 (which uses round-year boundaries for
// general attribution) — do not conflate the two.
//
// **Decision (NSERegimesGate separate from NSERegimes2018_2024) — convention: experimental**
// scope: internal/analytics
// tags: regime, universe-gate, TASK-0052
// owner: priya
//
// The existing NSERegimes2018_2024 var uses 2020-01-01 and 2022-01-01 as boundaries.
// The 2026-04-27 decision file specifies 2020-02-01 and 2021-07-01 (matching the
// actual COVID crash and recovery transition dates). A separate var avoids a breaking
// change to tests that reference NSERegimes2018_2024 and keeps the two distinct uses
// (general attribution vs. evaluation gate) independently evolvable.
var NSERegimesGate = []Regime{
	{
		Name: "Pre-COVID (2018–Jan 2020)",
		From: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "COVID crash + recovery (Feb 2020–Jun 2021)",
		From: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "Post-recovery (Jul 2021–2024)",
		From: time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// RegimeGateReport is the output of ComputeRegimeGate.
type RegimeGateReport struct {
	// RegimeConcentrated is true when any single regime accounts for >= 70% of
	// total Sharpe mass (abs-weighted), or when any regime window has zero trades.
	RegimeConcentrated bool
	// Regimes holds the per-regime breakdown in the same order as the input regimes.
	Regimes []RegimeContribution
}

// RegimeContribution holds per-regime metrics from the regime gate computation.
type RegimeContribution struct {
	// Name is copied from the input Regime.Name.
	Name string
	// PerTradeSharpe is mean(ReturnOnNotional) / std(ReturnOnNotional) with sample
	// variance (n-1), no annualization. 0 when fewer than 2 trades are in this regime.
	PerTradeSharpe float64
	// Contribution is abs(PerTradeSharpe) / sum(abs(PerTradeSharpe)) across all regimes.
	// 0 when sum of all abs(PerTradeSharpe) == 0 (degenerate case).
	Contribution float64
	// TradeCount is the number of trades whose ExitTime falls in this regime's window.
	TradeCount int
}

// ComputeRegimeGate evaluates whether a strategy's Sharpe is concentrated in a single
// market regime.
//
// For each regime window, it computes the per-trade Sharpe on trades whose ExitTime
// falls within [regime.From, regime.To). The contribution fraction of each regime is
// abs(S[i]) / sum(abs(S[j])) — this avoids sign-cancellation when regimes have mixed
// Sharpe signs (see 2026-04-27 decision for the full rationale).
//
// RegimeConcentrated is set when:
//   - Any contribution[i] >= 0.70, OR
//   - Any regime window has zero trades (undefined Sharpe treated as concentration).
//
// Per the 2026-04-27 decision, a concentrated strategy is not killed outright —
// it receives half-weight during portfolio construction (TASK-0055). This function
// only computes the flag; weighting is the caller's responsibility.
func ComputeRegimeGate(trades []model.Trade, regimes []Regime) RegimeGateReport {
	contribs := make([]RegimeContribution, len(regimes))
	for i, r := range regimes {
		contribs[i].Name = r.Name
		var regimeTrades []model.Trade
		for _, t := range trades {
			if (t.ExitTime.Equal(r.From) || t.ExitTime.After(r.From)) &&
				t.ExitTime.Before(r.To) {
				regimeTrades = append(regimeTrades, t)
			}
		}
		contribs[i].TradeCount = len(regimeTrades)
		if len(regimeTrades) >= 2 {
			returns := make([]float64, len(regimeTrades))
			for j, t := range regimeTrades {
				returns[j] = t.ReturnOnNotional()
			}
			contribs[i].PerTradeSharpe = computePerTradeSharpe(returns)
		}
		// PerTradeSharpe stays 0 for 0 or 1 trades — computePerTradeSharpe returns 0 for n < 2.
	}

	// Compute abs-sum denominator for contribution fractions.
	var absSum float64
	for _, c := range contribs {
		absSum += math.Abs(c.PerTradeSharpe)
	}

	// Assign contributions when the denominator is non-zero.
	if absSum > 0 {
		for i := range contribs {
			contribs[i].Contribution = math.Abs(contribs[i].PerTradeSharpe) / absSum
		}
	}
	// If absSum == 0: all contributions remain 0 (degenerate — all regimes have
	// zero-variance or near-zero Sharpe). The concentration flag is still evaluated
	// correctly via the zero-trade check below.

	// A strategy is concentrated if any regime has zero trades (undefined Sharpe)
	// or if any regime's abs-contribution fraction reaches the 70% threshold.
	concentrated := false
	for _, c := range contribs {
		if c.TradeCount == 0 || c.Contribution >= 0.70 {
			concentrated = true
			break
		}
	}

	return RegimeGateReport{
		RegimeConcentrated: concentrated,
		Regimes:            contribs,
	}
}
