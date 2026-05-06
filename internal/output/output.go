// Package output formats backtest results for human consumption and JSON export.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// RunConfig holds the metadata describing a specific backtest run.
// It is embedded at the top level of the JSON output alongside the performance metrics.
//
// All fields are plain strings rather than model.Timeframe or model.CommissionModel to
// decouple the JSON shape from pkg/model type changes. The cmd layer converts typed values
// to strings before constructing RunConfig.
//
// **Decision (RunConfig in internal/output — architecture: experimental)**
// scope: internal/output, cmd/backtest, cmd/universe-sweep
// tags: metadata, JSON, run-config, output
// owner: priya
//
// RunConfig is a serialization DTO for the output layer. It belongs here rather than
// pkg/model because it describes how a run is represented in serialized form — not a
// domain primitive that other packages need to import.
type RunConfig struct {
	Instrument      string            `json:"instrument,omitempty"`
	Timeframe       string            `json:"timeframe,omitempty"`
	From            string            `json:"from,omitempty"`
	To              string            `json:"to,omitempty"`
	Strategy        string            `json:"strategy,omitempty"`
	CommissionModel string            `json:"commission_model,omitempty"`
	Parameters      map[string]string `json:"parameters,omitempty"`
}

// Config controls where backtest results are written.
type Config struct {
	FilePath       string                      // destination for JSON export; ignored if empty
	PrintToStdout  bool                        // print human-readable summary to stdout
	Stdout         io.Writer                   // overrides os.Stdout when PrintToStdout is true; nil means os.Stdout
	Benchmark      *analytics.BenchmarkReport  // optional; when non-nil, printed alongside strategy results
	CurvePath      string                      // destination for equity curve CSV export; ignored if empty
	Curve          []model.EquityPoint         // per-bar equity snapshots; written to CurvePath when that field is non-empty
	GateThreshold  float64                     // Sharpe threshold for proliferation gate; 0 disables the check
	RegimeSplits   []analytics.RegimeReport    // optional; when non-empty, printed as a per-regime table
	Bootstrap      *montecarlo.BootstrapResult // optional; when non-nil, bootstrap section printed
	BootstrapSeed  int64                       // seed used in the bootstrap run; printed in header
	BootstrapNSims int                         // simulation count; 0 displayed as 10000 (the montecarlo default)
	RunConfig      RunConfig                   // optional run metadata embedded at top level of JSON output; zero value omits all metadata fields
	RegimeGate     *analytics.RegimeGateReport // optional; when non-nil, prints regime gate section and includes in JSON
}

// Write formats report as a human-readable summary and/or a JSON file.
// Both outputs are optional and controlled by cfg.
func Write(report analytics.Report, cfg Config) error { //nolint:gocritic // Report and Config are caller-constructed value types; pointer would leak internals
	if cfg.PrintToStdout {
		w := cfg.Stdout
		if w == nil {
			w = os.Stdout
		}
		if err := printSummary(w, report, cfg.Benchmark, cfg.GateThreshold, cfg.RegimeSplits); err != nil {
			return err
		}
		if cfg.Bootstrap != nil {
			if err := printBootstrapSection(w, cfg.Bootstrap, cfg.BootstrapSeed, cfg.BootstrapNSims); err != nil {
				return err
			}
		}
		if cfg.RegimeGate != nil {
			if err := printRegimeGateSection(w, cfg.RegimeGate); err != nil {
				return err
			}
		}
	}

	if cfg.FilePath != "" {
		if err := writeJSON(cfg.FilePath, report, cfg.RunConfig, cfg.Bootstrap, cfg.BootstrapSeed, cfg.BootstrapNSims, cfg.RegimeGate); err != nil {
			return err
		}
	}

	if cfg.CurvePath != "" {
		if err := writeCurveCSV(cfg.CurvePath, cfg.Curve); err != nil {
			return err
		}
	}

	return nil
}

func printSummary(w io.Writer, r analytics.Report, b *analytics.BenchmarkReport, gateThreshold float64, regimes []analytics.RegimeReport) error { //nolint:gocritic // value semantics intentional; r is read-only
	_, err := fmt.Fprintf(w,
		"=== Backtest Results ===\nTrades:         %d\nWin Rate:       %.2f%%\nTotal P&L:      %.2f\nAvg Win:        %.2f\nAvg Loss:       %.2f\nProfit Factor:  %.4f\nMax Drawdown:   %.2f%%\nMax DD Duration:%v\nSharpe Ratio:   %.4f\nSortino Ratio:  %.4f\nCalmar Ratio:   %.4f\nTail Ratio:     %.4f\n",
		r.TradeCount, r.WinRate, r.TotalPnL, r.AvgWin, r.AvgLoss, r.ProfitFactor,
		r.MaxDrawdown, r.MaxDrawdownDuration, r.SharpeRatio, r.SortinoRatio, r.CalmarRatio, r.TailRatio,
	)
	if err != nil {
		return fmt.Errorf("output: write summary: %w", err)
	}
	if r.TradeMetricsInsufficient {
		if _, err := fmt.Fprintf(w, "WARNING: trade count (%d) below minimum (%d) -- WinRate, ProfitFactor, AvgWin, AvgLoss not reported\n",
			r.TradeCount, analytics.MinTradesForMetrics); err != nil {
			return fmt.Errorf("output: write trade warning: %w", err)
		}
	}
	if r.CurveMetricsInsufficient {
		if _, err := fmt.Fprintf(w, "WARNING: equity curve below minimum (%d bars) -- Sharpe, Sortino, Calmar, TailRatio not reported\n",
			analytics.MinCurvePointsForMetrics); err != nil {
			return fmt.Errorf("output: write curve warning: %w", err)
		}
	}
	if gateThreshold > 0 && !r.TradeMetricsInsufficient && !r.CurveMetricsInsufficient {
		status := "PASS"
		if r.SharpeRatio < gateThreshold {
			status = "FAIL"
		}
		if _, err := fmt.Fprintf(w, "Proliferation gate (≥%.2f): %s (Sharpe %.4f)\n",
			gateThreshold, status, r.SharpeRatio); err != nil {
			return fmt.Errorf("output: write gate result: %w", err)
		}
	}
	if b != nil {
		_, err = fmt.Fprintf(w,
			"\n--- Buy-and-Hold Benchmark ---\nTotal Return:      %.2f%%\nAnnualized Return: %.2f%%\nMax Drawdown:      %.2f%%\nSharpe Ratio:      %.4f\n",
			b.TotalReturn, b.AnnualizedReturn, b.MaxDrawdown, b.SharpeRatio,
		)
		if err != nil {
			return fmt.Errorf("output: write benchmark summary: %w", err)
		}
	}
	if len(regimes) > 0 {
		if err := printRegimeTable(w, regimes); err != nil {
			return err
		}
	}
	return nil
}

func printBootstrapSection(w io.Writer, result *montecarlo.BootstrapResult, seed int64, nSims int) error {
	if nSims <= 0 {
		nSims = 10_000
	}
	if _, err := fmt.Fprintf(w, "\n--- Bootstrap (%d sims, seed=%d) ---\n", nSims, seed); err != nil {
		return fmt.Errorf("output: write bootstrap header: %w", err)
	}
	if _, err := fmt.Fprintf(w,
		"Per-trade Sharpe  p5=%8.4f  p50=%8.4f  p95=%8.4f  mean=%8.4f\n",
		result.SharpeP5, result.SharpeP50, result.SharpeP95, result.MeanSharpe,
	); err != nil {
		return fmt.Errorf("output: write bootstrap sharpe: %w", err)
	}
	if _, err := fmt.Fprintf(w,
		"Worst DD%%         p5=%8.2f  p50=%8.2f  p95=%8.2f\n",
		result.WorstDrawdownP5, result.WorstDrawdownP50, result.WorstDrawdownP95,
	); err != nil {
		return fmt.Errorf("output: write bootstrap drawdown: %w", err)
	}
	if _, err := fmt.Fprintf(w,
		"Prob(Sharpe > 0): %.1f%%\n",
		result.ProbPositiveSharpe*100,
	); err != nil {
		return fmt.Errorf("output: write bootstrap prob: %w", err)
	}
	if _, err := fmt.Fprintf(w,
		"Kill-switch threshold (p5 Sharpe): %.4f\n",
		result.SharpeP5,
	); err != nil {
		return fmt.Errorf("output: write bootstrap kill-switch: %w", err)
	}
	return nil
}

// printRegimeGateSection prints the per-regime PerTradeSharpe, Contribution%, TradeCount,
// and RegimeConcentrated flag for the regime gate evaluation (TASK-0086).
func printRegimeGateSection(w io.Writer, r *analytics.RegimeGateReport) error {
	if _, err := fmt.Fprintf(w, "\n--- Regime Gate ---\n%-42s  %-14s  %-12s  %s\n",
		"Regime", "PerTradeSharpe", "Contribution", "TradeCount"); err != nil {
		return fmt.Errorf("output: write regime gate header: %w", err)
	}
	for _, rc := range r.Regimes {
		if _, err := fmt.Fprintf(w, "%-42s  %-14.4f  %-12s  %d\n",
			rc.Name, rc.PerTradeSharpe, fmt.Sprintf("%.2f%%", rc.Contribution*100), rc.TradeCount); err != nil {
			return fmt.Errorf("output: write regime gate row %q: %w", rc.Name, err)
		}
	}
	if _, err := fmt.Fprintf(w, "RegimeConcentrated: %v\n", r.RegimeConcentrated); err != nil {
		return fmt.Errorf("output: write regime concentrated flag: %w", err)
	}
	return nil
}

func printRegimeTable(w io.Writer, regimes []analytics.RegimeReport) error {
	if _, err := fmt.Fprintf(w, "\n--- Regime Split ---\n%-38s  %-10s  %-10s  %s\n",
		"Regime", "Sharpe", "MaxDD%", "Period"); err != nil {
		return fmt.Errorf("output: write regime header: %w", err)
	}
	for _, reg := range regimes {
		period := reg.From.Format("2006-01") + " – " + reg.To.Format("2006-01")
		if _, err := fmt.Fprintf(w, "%-38s  %-10.4f  %-10.2f  %s\n",
			reg.Name, reg.SharpeRatio, reg.MaxDrawdown, period); err != nil {
			return fmt.Errorf("output: write regime row %q: %w", reg.Name, err)
		}
	}
	return nil
}

// WriteSweep prints a ranked table of sweep results and, if present, the plateau
// and DSR-corrected peak Sharpe to w. Results are expected to arrive pre-sorted
// descending by Sharpe ratio.
func WriteSweep(w io.Writer, report sweep.Report) error { //nolint:gocritic // Report is a caller-constructed value type; pointer would leak internals
	if _, err := fmt.Fprintf(w,
		"=== Parameter Sweep: %s ===\n%-4s  %-10s  %-8s  %-12s  %-6s  %-8s\n",
		report.ParameterName,
		"Rank", report.ParameterName, "Sharpe", "P&L", "Trades", "MaxDD%",
	); err != nil {
		return fmt.Errorf("output: write sweep header: %w", err)
	}

	for i, r := range report.Results {
		if _, err := fmt.Fprintf(w,
			"%-4d  %-10.2f  %-8.4f  %-12.2f  %-6d  %-8.2f\n",
			i+1, r.ParamValue, r.SharpeRatio, r.TotalPnL, r.TradeCount, r.MaxDrawdown,
		); err != nil {
			return fmt.Errorf("output: write sweep row %d: %w", i+1, err)
		}
	}

	if report.Plateau != nil {
		p := report.Plateau
		if _, err := fmt.Fprintf(w,
			"\nPlateau: %s in [%.2f, %.2f] (%d values, min Sharpe %.4f)\n",
			report.ParameterName, p.MinParam, p.MaxParam, p.Count, p.MinSharpe,
		); err != nil {
			return fmt.Errorf("output: write sweep plateau: %w", err)
		}
	}

	if len(report.Results) > 0 && report.VariantCount > 1 && report.NObservations > 1 {
		peakSharpe := report.Results[0].SharpeRatio
		dsr := analytics.DSR(peakSharpe, float64(report.VariantCount), float64(report.NObservations))
		if _, err := fmt.Fprintf(w,
			"Peak Sharpe: %.4f  DSR-corrected: %.4f  (variants: %d, obs: %d)\n",
			peakSharpe, dsr, report.VariantCount, report.NObservations,
		); err != nil {
			return fmt.Errorf("output: write sweep DSR: %w", err)
		}
	}

	return nil
}

// writeCurveCSV writes the equity curve to path in CSV format.
//
// CSV format:
//
//	timestamp,equity_value
//	2018-01-02T09:15:00Z,100000.00
//
// Timestamps are RFC 3339 UTC. equity_value is rounded to two decimal places.
// The file is created or truncated at path before writing. An empty curve
// writes only the header row.
func writeCurveCSV(path string, curve []model.EquityPoint) error {
	var buf bytes.Buffer
	buf.WriteString("timestamp,equity_value\n")
	for _, pt := range curve {
		fmt.Fprintf(&buf, "%s,%.2f\n", pt.Timestamp.UTC().Format(time.RFC3339), pt.Value)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("output: write curve file %q: %w", path, err)
	}
	return nil
}

// WriteCorrelationMatrix prints a pairwise correlation matrix table to w.
// NaN values are printed as "  n/a  " to signal an empty or constant-series window.
func WriteCorrelationMatrix(w io.Writer, m analytics.CorrelationMatrix) error {
	if _, err := fmt.Fprintf(w, "\n--- Strategy Correlation Matrix ---\n%-24s  %-24s  %-12s  %-12s  %-12s  %s\n",
		"Strategy A", "Strategy B", "Full-Period", "2020-Crash", "2022-Corr", "Note",
	); err != nil {
		return fmt.Errorf("output: write correlation header: %w", err)
	}
	for _, p := range m.Pairs {
		note := ""
		if p.TooCorrelated {
			note = "WARN: too correlated — halve combined allocation"
		}
		if _, err := fmt.Fprintf(w, "%-24s  %-24s  %-12s  %-12s  %-12s  %s\n",
			p.NameA, p.NameB,
			formatCorr(p.FullPeriod),
			formatCorr(p.Crash2020),
			formatCorr(p.Correction2022),
			note,
		); err != nil {
			return fmt.Errorf("output: write correlation row %q/%q: %w", p.NameA, p.NameB, err)
		}
	}
	return nil
}

func formatCorr(v float64) string {
	if math.IsNaN(v) {
		return "  n/a  "
	}
	return fmt.Sprintf("%.4f", v)
}

// BootstrapStats holds the bootstrap distribution statistics for JSON serialization.
// It is written to the "bootstrap" key in the output JSON when a bootstrap run was
// performed; the key is absent entirely when no bootstrap was run.
//
// **Decision (bootstrap stats JSON placement — convention: experimental)**
// scope: internal/output
// tags: JSON, bootstrap, omitempty, serialization
// owner: priya
//
// Fields are placed under a named "bootstrap" nested key (not promoted to the top level)
// using a *BootstrapStats pointer with omitempty. Top-level promotion with omitempty would
// suppress valid zero results (e.g. SharpeP5 == 0.0 is a legitimate bootstrap outcome).
// A pointer-to-struct means the block is either fully present or fully absent — no
// zero-value ambiguity.
type BootstrapStats struct {
	SharpeP5           float64 `json:"sharpe_p5"`
	SharpeP50          float64 `json:"sharpe_p50"`
	SharpeP95          float64 `json:"sharpe_p95"`
	ProbPositiveSharpe float64 `json:"prob_positive_sharpe"`
	WorstDrawdownP95   float64 `json:"worst_drawdown_p95"`
	N                  int     `json:"n"`
	Seed               int64   `json:"seed"`
}

// jsonResult merges RunConfig metadata fields with analytics.Report fields at the
// JSON top level. Go's encoding/json promotes embedded struct fields to the top level,
// so both RunConfig and analytics.Report fields appear as siblings in the output JSON.
//
// **Decision (jsonResult struct embedding for top-level JSON merge — convention: experimental)**
// scope: internal/output
// tags: JSON, embedding, serialization
// owner: priya
//
// analytics.Report is embedded (anonymous field) so its exported fields are promoted
// to the top level of the JSON object. RunConfig fields use omitempty so that a
// zero-valued RunConfig produces no extra keys — existing callers that pass no RunConfig
// get identical JSON output.
//
// Bootstrap is a named pointer field (not embedded) so it appears under a "bootstrap"
// sub-key. The omitempty tag means the key is absent entirely when no bootstrap was run.
//
// RegimeGate is a named pointer field so it appears under a "regime_gate" sub-key.
// The omitempty tag means the key is absent entirely when no regime gate was run.
type jsonResult struct {
	RunConfig
	analytics.Report
	Bootstrap  *BootstrapStats             `json:"bootstrap,omitempty"`
	RegimeGate *analytics.RegimeGateReport `json:"regime_gate,omitempty"`
}

func writeJSON(path string, r analytics.Report, rc RunConfig, br *montecarlo.BootstrapResult, seed int64, nSims int, rg *analytics.RegimeGateReport) error { //nolint:gocritic // value semantics intentional; r and rc are read-only
	result := jsonResult{RunConfig: rc, Report: r, RegimeGate: rg}
	if br != nil {
		n := nSims
		if n <= 0 {
			n = 10_000
		}
		result.Bootstrap = &BootstrapStats{
			SharpeP5:           br.SharpeP5,
			SharpeP50:          br.SharpeP50,
			SharpeP95:          br.SharpeP95,
			ProbPositiveSharpe: br.ProbPositiveSharpe,
			WorstDrawdownP95:   br.WorstDrawdownP95,
			N:                  n,
			Seed:               seed,
		}
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("output: marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("output: write file %q: %w", path, err)
	}
	return nil
}
