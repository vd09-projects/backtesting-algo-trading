// cmd/signal-audit runs the signal-frequency audit across all strategies and
// a universe of instruments. It verifies that each strategy generates at least
// 30 trades on each instrument before committing to a full backtest pipeline.
//
// For the CCI mean-reversion strategy it additionally prints a distribution
// report to stderr: average trades/instrument and COVID-window (Jan–Jun 2020)
// trade concentration per instrument.
//
// Usage:
//
//	go run ./cmd/signal-audit \
//	    --universe universes/nifty50-large-cap.yaml \
//	    --from 2018-01-01 \
//	    --to   2024-01-01 \
//	    --out  runs/signal-frequency-audit-YYYY-MM-DD.csv
//
// Output is a strategy × instrument matrix CSV:
//
//	strategy,total_trades,killed,NSE:RELIANCE,NSE:INFY,...
//	sma-crossover,450,false,32,28,...
//
// Cells with fewer than 30 trades are written as EXCLUDED(<count>).
// Strategies with fewer than 30 total trades across the universe are written
// with killed=KILLED and must not proceed to any full backtest pipeline run.
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). Token handling is
// identical to cmd/backtest.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/signalaudit"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// covidWindowStart and covidWindowEnd define the Q1-Q2 2020 clustering window
// per the Marcus audit specification (Jan 1 – Jun 30 2020 inclusive).
var (
	covidWindowStart = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	covidWindowEnd   = time.Date(2020, 7, 1, 0, 0, 0, 0, time.UTC) // exclusive upper bound
)

// cliFlags holds parsed command-line arguments.
type cliFlags struct {
	universeFile string
	from         time.Time
	to           time.Time
	outPath      string
	cash         float64
	positionSize float64
	slippage     float64
}

func main() {
	flags := parseFlags()

	instruments, err := universesweep.ParseUniverseFile(flags.universeFile)
	if err != nil {
		cmdutil.Fatalf("universe file: %v", err)
	}

	tf := model.TimeframeDaily
	factories := allStrategyFactories(tf)

	cfg := signalaudit.Config{
		StrategyFactories: factories,
		Instruments:       instruments,
		EngineConfig:      buildEngineConfig(&flags),
		Timeframe:         tf,
	}

	ctx := context.Background()
	cmdutil.LoadDotEnv(".env")

	p, err := cmdutil.BuildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	fmt.Fprintf(os.Stderr, "Signal frequency audit: %d strategies × %d instruments  %s → %s\n", //nolint:errcheck // progress to stderr
		len(factories), len(instruments), flags.from.Format("2006-01-02"), flags.to.Format("2006-01-02"))

	report, err := signalaudit.Run(ctx, &cfg, p)
	if err != nil {
		cmdutil.Fatalf("signal audit: %v", err)
	}

	writeReport(os.Stdout, os.Stderr, report, flags.outPath, len(instruments))
}

// parseFlags defines, parses, and validates all command-line flags.
// It calls cmdutil.Fatalf on any validation failure and never returns an error.
func parseFlags() cliFlags {
	universeFile := flag.String("universe", "", "Path to YAML universe file (required)")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive, required)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive, required)")
	outPath := flag.String("out", "", "Output CSV path (default: stdout)")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	positionSize := flag.Float64("position-size", 0.10, "Fraction of cash deployed per trade")
	slippage := flag.Float64("slippage", 0.0005, "Slippage as decimal fraction (e.g. 0.0005 = 0.05%)")

	flag.Parse()

	flags, err := validateFlagInputs(*universeFile, *fromStr, *toStr, *outPath, *cash, *positionSize, *slippage)
	if err != nil {
		cmdutil.Fatalf("%v", err)
	}
	return flags
}

// validateFlagInputs validates the raw flag string values and returns a cliFlags
// or an error. Extracted from parseFlags so tests can exercise validation without
// manipulating os.Args or the global flag.CommandLine.
func validateFlagInputs(universeFile, fromStr, toStr, outPath string, cash, positionSize, slippage float64) (cliFlags, error) {
	if universeFile == "" {
		return cliFlags{}, fmt.Errorf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if fromStr == "" {
		return cliFlags{}, fmt.Errorf("--from is required (e.g. 2018-01-01)")
	}
	if toStr == "" {
		return cliFlags{}, fmt.Errorf("--to is required (e.g. 2024-01-01)")
	}

	from, to, err := cmdutil.ParseDateRange(fromStr, toStr)
	if err != nil {
		return cliFlags{}, err
	}

	return cliFlags{
		universeFile: universeFile,
		from:         from,
		to:           to,
		outPath:      outPath,
		cash:         cash,
		positionSize: positionSize,
		slippage:     slippage,
	}, nil
}

// buildEngineConfig constructs the engine.Config template from parsed flags.
func buildEngineConfig(flags *cliFlags) engine.Config {
	return engine.Config{
		From:                 flags.from,
		To:                   flags.to,
		InitialCash:          flags.cash,
		PositionSizeFraction: flags.positionSize,
		OrderConfig: model.OrderConfig{
			SlippagePct:     flags.slippage,
			CommissionModel: model.CommissionZerodhaFull,
		},
	}
}

// writeReport writes the audit CSV to outPath (or csvOut if empty), then
// prints the kill/excluded summary and CCI distribution report to errOut,
// and exits 1 if any strategies were killed.
func writeReport(csvOut, errOut io.Writer, report signalaudit.Report, outPath string, nInstruments int) {
	out := csvOut
	var f *os.File
	if outPath != "" {
		var err error
		f, err = os.Create(outPath)
		if err != nil {
			cmdutil.Fatalf("create output file %q: %v", outPath, err)
		}
		out = f
	}

	if err := signalaudit.WriteCSV(out, report); err != nil {
		if f != nil {
			_ = f.Close() //nolint:errcheck // best-effort; exiting immediately after
		}
		cmdutil.Fatalf("write CSV: %v", err)
	}

	if f != nil {
		if err := f.Close(); err != nil {
			cmdutil.Fatalf("close output file: %v", err)
		}
	}

	printCCIDistributionReport(errOut, report)

	killed, excluded := summariseReport(report)

	fmt.Fprintf(errOut, "\nSummary: %d/%d strategies killed, %d/%d cells excluded (< %d trades)\n", //nolint:errcheck // progress to stderr
		killed, len(report.Rows),
		excluded, len(report.Rows)*nInstruments,
		signalaudit.MinTradesPerCell,
	)

	if killed > 0 {
		os.Exit(1)
	}
}

// summariseReport counts killed strategies and excluded cells, printing each
// killed strategy to stderr. Returns (killed, excluded) counts.
func summariseReport(report signalaudit.Report) (killed, excluded int) {
	for _, row := range report.Rows {
		if row.Killed {
			killed++
			fmt.Fprintf(os.Stderr, "KILLED: %s (total trades: %d)\n", row.Strategy, row.TotalTrades) //nolint:errcheck // progress to stderr
		}
		for _, cell := range row.Cells {
			if cell.Excluded {
				excluded++
			}
		}
	}
	return killed, excluded
}

// auditParamOverrides maps each strategy name to the parameter set used for
// signal-frequency auditing. These differ from registry defaults because the
// audit uses plateau-midpoints chosen per Marcus's signal-audit verdict — params
// that guarantee ≥30 trades per instrument on the Nifty50 large-cap universe.
//
// References:
//   - decisions/algorithm/2026-05-01-signalaudit-strategy-factory-decoupling.md
//   - decisions/algorithm/2026-04-22-walkforward-strategy-factory-per-fold.md
//
// Strategies absent from this map are excluded from signal-audit with a startup
// warning. Add an entry here when registering a new strategy in strategies.go.
//
// **Decision (signal-audit uses registry iteration + local auditParamOverrides) — architecture: experimental**
// scope: cmd/signal-audit
// tags: signal-audit, registry, package-boundary, plateau-midpoint
// owner: priya
//
// The previous implementation imported 7 concrete strategy packages directly,
// violating the "no concrete type across package boundaries" rule. The registry
// already has WalkForwardFactory which validates params once and returns a
// fresh-instance factory — exactly what signal-audit needs per instrument.
// Plateau-midpoint params that differ from registry defaults live in this local
// map, keeping the decision visible without coupling to concrete strategy packages.
// "stub" is intentionally omitted — it is a test tool, not a real strategy.
var auditParamOverrides = map[string]map[string]float64{
	// slow=20 (not registry default 50) — slow=50 has <30 trades on RELIANCE per audit.
	// See decisions/algorithm/2026-05-01-signalaudit-strategy-factory-decoupling.md.
	"sma-crossover": {"fast-period": 10, "slow-period": 20},

	// Standard defaults — audit params match registry defaults.
	"rsi-mean-reversion": {"rsi-period": 14, "oversold": 30, "overbought": 70},

	// period=10 (not registry default 20) — only value with ≥30 trades on RELIANCE.
	"donchian-breakout": {"donchian-period": 10},

	// fast=17 (plateau [15,21], midpoint 17); slow/signal at registry defaults.
	"macd-crossover": {"macd-fast-period": 17, "macd-slow-period": 26, "macd-signal-period": 9},

	// Standard defaults.
	"bollinger-mean-reversion": {"bb-period": 20, "bb-num-std-dev": 2.0},

	// Standard defaults.
	"momentum": {"momentum-lookback": 231, "momentum-threshold": 10.0},

	// Standard CCI params: entry=-100 (oversold), exit=0 (neutral cross).
	"cci-mean-reversion": {"cci-period": 20, "cci-entry-threshold": -100, "cci-exit-threshold": 0},
}

// allStrategyFactories returns a StrategyFactory for each registered strategy
// using audit-specific parameters from auditParamOverrides. Strategies not
// present in auditParamOverrides are skipped with a stderr warning — this
// prevents silent exclusion when a new strategy is added to the registry without
// a corresponding audit param entry.
//
// WalkForwardFactory validates params once at startup and panics inside the
// returned factory only on unexpected construction failure (i.e. params passed
// validation but Build failed — a programming error, not a runtime condition).
// This is acceptable for a diagnostic tool where a mid-audit panic is preferable
// to silently wrong results.
func allStrategyFactories(tf model.Timeframe) []signalaudit.StrategyFactory {
	var factories []signalaudit.StrategyFactory
	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		params, ok := auditParamOverrides[name]
		if !ok {
			// New strategy registered but not in auditParamOverrides — operator must add an entry.
			fmt.Fprintf(os.Stderr, "signal-audit: WARNING: strategy %q has no auditParamOverrides entry; skipping\n", name) //nolint:errcheck // startup warning
			continue
		}
		factory, err := cmdutil.GlobalRegistry.WalkForwardFactory(name, tf, params)
		if err != nil {
			cmdutil.Fatalf("signal-audit: strategy %q: %v", name, err)
		}
		factories = append(factories, signalaudit.StrategyFactory{
			Name: name,
			New:  factory,
		})
	}
	return factories
}

// printCCICells prints one line per cell showing trade count and COVID-window
// percentage, marks clustered instruments, and returns (totalTrades, covidViolations).
func printCCICells(w io.Writer, cells []signalaudit.Cell, maxCovidPct float64) (totalTrades, covidViolations int) {
	for _, cell := range cells {
		covidCount := countCovidTrades(cell)

		var covidPct float64
		if cell.TradeCount > 0 {
			covidPct = float64(covidCount) / float64(cell.TradeCount) * 100
		}

		clusterFlag := ""
		if covidPct > maxCovidPct {
			clusterFlag = " [CLUSTERED]"
			covidViolations++
		}

		fmt.Fprintf(w, "  %-20s  trades=%3d  covid=%3d (%.1f%%)%s\n", //nolint:errcheck // progress to stderr
			cell.Instrument, cell.TradeCount, covidCount, covidPct, clusterFlag)
		totalTrades += cell.TradeCount
	}
	return totalTrades, covidViolations
}

// countCovidTrades returns the number of trades in cell whose ExitTime falls
// within the COVID window [covidWindowStart, covidWindowEnd).
func countCovidTrades(cell signalaudit.Cell) int {
	n := 0
	for _, tr := range cell.Trades {
		if !tr.ExitTime.Before(covidWindowStart) && tr.ExitTime.Before(covidWindowEnd) {
			n++
		}
	}
	return n
}

// printCCIDistributionReport writes per-instrument trade counts and COVID-window
// clustering percentages for the cci-mean-reversion strategy to w.
//
// COVID window: Jan 1 – Jun 30 2020 (covidWindowStart inclusive, covidWindowEnd exclusive).
// Pass condition (per Marcus verdict): avg trades/instrument ≥ 25 AND no instrument >30%
// of its trades in the COVID window.
func printCCIDistributionReport(w io.Writer, report signalaudit.Report) {
	const cciStrategy = "cci-mean-reversion"
	// minAvgTrades and maxCovidPct are Marcus-specified thresholds from the signal-audit verdict.
	// See decisions/algorithm/2026-05-02-cci-mean-reversion-signal-audit-proceed.md.
	const minAvgTrades = 25
	const maxCovidPct = 30.0

	var cciRow *signalaudit.Row
	for i := range report.Rows {
		if report.Rows[i].Strategy == cciStrategy {
			cciRow = &report.Rows[i]
			break
		}
	}
	if cciRow == nil {
		return
	}

	fmt.Fprintf(w, "\n=== CCI Mean-Reversion Distribution Report ===\n") //nolint:errcheck // progress to stderr
	fmt.Fprintf(w, "COVID window: %s – %s\n",                            //nolint:errcheck // progress to stderr
		covidWindowStart.Format("2006-01-02"),
		covidWindowEnd.AddDate(0, 0, -1).Format("2006-01-02"),
	)
	fmt.Fprintf(w, "Pass condition: avg ≥ %d trades/instrument AND no instrument >%.0f%% COVID\n\n", //nolint:errcheck // progress to stderr
		minAvgTrades, maxCovidPct)

	nInst := len(cciRow.Cells)
	totalTrades, covidViolations := printCCICells(w, cciRow.Cells, maxCovidPct)

	var avgTrades float64
	if nInst > 0 {
		avgTrades = float64(totalTrades) / float64(nInst)
	}

	fmt.Fprintf(w, "\nAvg trades/instrument: %.1f  (pass threshold: ≥%d)\n", avgTrades, minAvgTrades) //nolint:errcheck // progress to stderr
	fmt.Fprintf(w, "COVID clustering violations: %d/%d instruments >%.0f%%\n",                        //nolint:errcheck // progress to stderr
		covidViolations, nInst, maxCovidPct)

	avgPass := avgTrades >= float64(minAvgTrades)
	clusterPass := covidViolations == 0

	fmt.Fprintf(w, "\nVerdict: ") //nolint:errcheck // progress to stderr
	switch {
	case avgPass && clusterPass:
		fmt.Fprintf(w, "PROCEED — avg trades and clustering both pass\n") //nolint:errcheck // progress to stderr
	case !avgPass && !clusterPass:
		fmt.Fprintf(w, "KILL — avg trades below threshold AND clustering violation\n") //nolint:errcheck // progress to stderr
	case !avgPass:
		fmt.Fprintf(w, "KILL — avg trades below threshold (%.1f < %d)\n", avgTrades, minAvgTrades) //nolint:errcheck // progress to stderr
	default:
		fmt.Fprintf(w, "KILL — COVID clustering violation (%d instruments >%.0f%%)\n", //nolint:errcheck // progress to stderr
			covidViolations, maxCovidPct)
	}
	fmt.Fprintf(w, "==============================================\n") //nolint:errcheck // progress to stderr
}
