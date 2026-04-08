// Package output formats backtest results for human consumption and JSON export.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
)

// Config controls where backtest results are written.
type Config struct {
	FilePath      string    // destination for JSON export; ignored if empty
	PrintToStdout bool      // print human-readable summary to stdout
	Stdout        io.Writer // overrides os.Stdout when PrintToStdout is true; nil means os.Stdout
}

// Write formats report as a human-readable summary and/or a JSON file.
// Both outputs are optional and controlled by cfg.
func Write(report analytics.Report, cfg Config) error {
	if cfg.PrintToStdout {
		w := cfg.Stdout
		if w == nil {
			w = os.Stdout
		}
		if err := printSummary(w, report); err != nil {
			return err
		}
	}

	if cfg.FilePath != "" {
		if err := writeJSON(cfg.FilePath, report); err != nil {
			return err
		}
	}

	return nil
}

func printSummary(w io.Writer, r analytics.Report) error {
	_, err := fmt.Fprintf(w,
		"=== Backtest Results ===\nTrades:       %d\nWin Rate:     %.2f%%\nTotal P&L:    %.2f\nMax Drawdown: %.2f%%\n",
		r.TradeCount, r.WinRate, r.TotalPnL, r.MaxDrawdown,
	)
	if err != nil {
		return fmt.Errorf("output: write summary: %w", err)
	}
	return nil
}

func writeJSON(path string, r analytics.Report) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("output: marshal report: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("output: write file %q: %w", path, err)
	}
	return nil
}
