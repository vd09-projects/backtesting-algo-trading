// Package output formats backtest results for human consumption and JSON export.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/sweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Config controls where backtest results are written.
type Config struct {
	FilePath      string                     // destination for JSON export; ignored if empty
	PrintToStdout bool                       // print human-readable summary to stdout
	Stdout        io.Writer                  // overrides os.Stdout when PrintToStdout is true; nil means os.Stdout
	Benchmark     *analytics.BenchmarkReport // optional; when non-nil, printed alongside strategy results
	CurvePath     string                     // destination for equity curve CSV export; ignored if empty
	Curve         []model.EquityPoint        // per-bar equity snapshots; written to CurvePath when that field is non-empty
	GateThreshold float64                    // Sharpe threshold for proliferation gate; 0 disables the check
}

// Write formats report as a human-readable summary and/or a JSON file.
// Both outputs are optional and controlled by cfg.
func Write(report analytics.Report, cfg Config) error { //nolint:gocritic // Report and Config are caller-constructed value types; pointer would leak internals
	if cfg.PrintToStdout {
		w := cfg.Stdout
		if w == nil {
			w = os.Stdout
		}
		if err := printSummary(w, report, cfg.Benchmark, cfg.GateThreshold); err != nil {
			return err
		}
	}

	if cfg.FilePath != "" {
		if err := writeJSON(cfg.FilePath, report); err != nil {
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

func printSummary(w io.Writer, r analytics.Report, b *analytics.BenchmarkReport, gateThreshold float64) error { //nolint:gocritic // value semantics intentional; r is read-only
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
	return nil
}

// WriteSweep prints a ranked table of sweep results and, if present, the plateau
// and DSR-corrected peak Sharpe to w. Results are expected to arrive pre-sorted
// descending by Sharpe ratio.
func WriteSweep(w io.Writer, report sweep.Report) error {
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

func writeJSON(path string, r analytics.Report) error { //nolint:gocritic // value semantics intentional; r is read-only
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("output: marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("output: write file %q: %w", path, err)
	}
	return nil
}
