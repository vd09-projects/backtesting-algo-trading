// cmd/correlate computes pairwise Pearson correlation across multiple strategy equity curves.
//
// Usage:
//
//	go run ./cmd/correlate \
//	    --curve "sma-crossover:runs/sma-crossover-curve.csv" \
//	    --curve "rsi-mean-rev:runs/rsi-mean-rev-curve.csv"
//
// Each --curve flag takes the form "name:path". At least two curves are required.
// The tool prints a correlation matrix with full-period and stress-period (NSE 2020 crash,
// 2022 correction) Pearson coefficients, plus a sizing note for correlated pairs.
//
// **Decision (2026-04.1.0) — architecture: experimental**
// scope: cmd/correlate
// tags: CLI, correlation, TASK-0027
//
// New binary rather than extending cmd/backtest. Correlation requires multiple strategy
// results; adding multi-strategy input to backtest would complicate its single-strategy model.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
)

type curveFlag []string

func (c *curveFlag) String() string     { return strings.Join(*c, ", ") }
func (c *curveFlag) Set(v string) error { *c = append(*c, v); return nil }

func main() {
	var curves curveFlag
	flag.Var(&curves, "curve", "name:path pair for a strategy equity curve CSV (repeatable; min 2)")
	flag.Parse()

	if len(curves) < 2 {
		fmt.Fprintln(os.Stderr, "correlate: at least two --curve flags required")
		flag.Usage()
		os.Exit(1)
	}

	named, err := loadCurves(curves)
	if err != nil {
		fmt.Fprintf(os.Stderr, "correlate: %v\n", err)
		os.Exit(1)
	}

	matrix := analytics.ComputeMatrix(named)
	if err := output.WriteCorrelationMatrix(os.Stdout, matrix); err != nil {
		fmt.Fprintf(os.Stderr, "correlate: %v\n", err)
		os.Exit(1)
	}
}

func loadCurves(flags []string) ([]analytics.NamedCurve, error) {
	out := make([]analytics.NamedCurve, 0, len(flags))
	for _, f := range flags {
		name, path, ok := strings.Cut(f, ":")
		if !ok || name == "" || path == "" {
			return nil, fmt.Errorf("invalid --curve %q: expected name:path", f)
		}
		pts, err := output.LoadCurveCSV(path)
		if err != nil {
			return nil, fmt.Errorf("load curve %q: %w", name, err)
		}
		out = append(out, analytics.NamedCurve{Name: name, Curve: pts})
	}
	return out, nil
}
