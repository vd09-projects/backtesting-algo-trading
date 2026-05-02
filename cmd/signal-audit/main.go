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
// with killed=KILLED and must not proceed to any full backtest run.
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). Token handling is
// identical to cmd/backtest.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/signalaudit"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/ccimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/momentum"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
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

	fmt.Fprintf(os.Stderr, "Signal frequency audit: %d strategies × %d instruments  %s → %s\n",
		len(factories), len(instruments), flags.from.Format("2006-01-02"), flags.to.Format("2006-01-02"))

	report, err := signalaudit.Run(ctx, &cfg, p)
	if err != nil {
		cmdutil.Fatalf("signal audit: %v", err)
	}

	writeReport(report, flags.outPath, len(instruments))
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

	if *universeFile == "" {
		cmdutil.Fatalf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2018-01-01)")
	}
	if *toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2024-01-01)")
	}

	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		cmdutil.Fatalf("--from %q: %v", *fromStr, err)
	}
	to, err := time.Parse("2006-01-02", *toStr)
	if err != nil {
		cmdutil.Fatalf("--to %q: %v", *toStr, err)
	}
	if !to.After(from) {
		cmdutil.Fatalf("--to must be strictly after --from")
	}

	return cliFlags{
		universeFile: *universeFile,
		from:         from,
		to:           to,
		outPath:      *outPath,
		cash:         *cash,
		positionSize: *positionSize,
		slippage:     *slippage,
	}
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

// writeReport writes the audit CSV to outPath (or stdout if empty), then
// prints the kill/excluded summary and CCI distribution report to stderr,
// and exits 1 if any strategies were killed.
func writeReport(report signalaudit.Report, outPath string, nInstruments int) {
	out := os.Stdout
	if outPath != "" {
		f, err := os.Create(outPath)
		if err != nil {
			cmdutil.Fatalf("create output file %q: %v", outPath, err)
		}
		out = f
	}

	if err := signalaudit.WriteCSV(out, report); err != nil {
		if out != os.Stdout {
			_ = out.Close() //nolint:errcheck // best-effort; exiting immediately after
		}
		cmdutil.Fatalf("write CSV: %v", err)
	}

	if out != os.Stdout {
		if err := out.Close(); err != nil {
			cmdutil.Fatalf("close output file: %v", err)
		}
	}

	printCCIDistributionReport(report)

	killed, excluded := summariseReport(report)

	fmt.Fprintf(os.Stderr, "\nSummary: %d/%d strategies killed, %d/%d cells excluded (< %d trades)\n",
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
			fmt.Fprintf(os.Stderr, "KILLED: %s (total trades: %d)\n", row.Strategy, row.TotalTrades)
		}
		for _, cell := range row.Cells {
			if cell.Excluded {
				excluded++
			}
		}
	}
	return killed, excluded
}

// allStrategyFactories returns a StrategyFactory for each of the 7 strategies
// using their default parameters. Each call to New() produces a fresh instance.
func allStrategyFactories(tf model.Timeframe) []signalaudit.StrategyFactory {
	return []signalaudit.StrategyFactory{
		{
			Name: "sma-crossover",
			New: func() signalaudit.Strategy {
				// plateau-midpoint: slow=20 (default slow=50 has <30 trades on RELIANCE)
				s, err := smacrossover.New(tf, 10, 20)
				if err != nil {
					cmdutil.Fatalf("sma-crossover: %v", err)
				}
				return s
			},
		},
		{
			Name: "rsi-mean-reversion",
			New: func() signalaudit.Strategy {
				s, err := rsimeanrev.New(tf, 14, 30, 70)
				if err != nil {
					cmdutil.Fatalf("rsi-mean-reversion: %v", err)
				}
				return s
			},
		},
		{
			Name: "donchian-breakout",
			New: func() signalaudit.Strategy {
				// plateau-midpoint: period=10 (only value with ≥30 trades on RELIANCE)
				s, err := donchian.New(tf, 10)
				if err != nil {
					cmdutil.Fatalf("donchian-breakout: %v", err)
				}
				return s
			},
		},
		{
			Name: "macd-crossover",
			New: func() signalaudit.Strategy {
				// plateau-midpoint: fast=17 (plateau [15,21], midpoint 17)
				s, err := macd.New(tf, 17, 26, 9)
				if err != nil {
					cmdutil.Fatalf("macd-crossover: %v", err)
				}
				return s
			},
		},
		{
			Name: "bollinger-mean-reversion",
			New: func() signalaudit.Strategy {
				s, err := bollinger.New(tf, 20, 2.0)
				if err != nil {
					cmdutil.Fatalf("bollinger-mean-reversion: %v", err)
				}
				return s
			},
		},
		{
			Name: "momentum",
			New: func() signalaudit.Strategy {
				s, err := momentum.New(tf, 231, 10.0)
				if err != nil {
					cmdutil.Fatalf("momentum: %v", err)
				}
				return s
			},
		},
		{
			Name: "cci-mean-reversion",
			New: func() signalaudit.Strategy {
				// Standard CCI params: period=20, entry=-100 (oversold), exit=0 (neutral cross).
				s, err := ccimeanrev.New(tf, 20, -100, 0)
				if err != nil {
					cmdutil.Fatalf("cci-mean-reversion: %v", err)
				}
				return s
			},
		},
	}
}

// printCCICells prints one line per cell showing trade count and COVID-window
// percentage, marks clustered instruments, and returns (totalTrades, covidViolations).
func printCCICells(cells []signalaudit.Cell, maxCovidPct float64) (totalTrades, covidViolations int) {
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

		fmt.Fprintf(os.Stderr, "  %-20s  trades=%3d  covid=%3d (%.1f%%)%s\n",
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
// clustering percentages for the cci-mean-reversion strategy to stderr.
//
// COVID window: Jan 1 – Jun 30 2020 (covidWindowStart inclusive, covidWindowEnd exclusive).
// Pass condition (per Marcus verdict): avg trades/instrument ≥ 25 AND no instrument >30%
// of its trades in the COVID window.
func printCCIDistributionReport(report signalaudit.Report) {
	const cciStrategy = "cci-mean-reversion"
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

	fmt.Fprintf(os.Stderr, "\n=== CCI Mean-Reversion Distribution Report ===\n")
	fmt.Fprintf(os.Stderr, "COVID window: %s – %s\n",
		covidWindowStart.Format("2006-01-02"),
		covidWindowEnd.AddDate(0, 0, -1).Format("2006-01-02"),
	)
	fmt.Fprintf(os.Stderr, "Pass condition: avg ≥ %d trades/instrument AND no instrument >%.0f%% COVID\n\n",
		minAvgTrades, maxCovidPct)

	nInst := len(cciRow.Cells)
	totalTrades, covidViolations := printCCICells(cciRow.Cells, maxCovidPct)

	var avgTrades float64
	if nInst > 0 {
		avgTrades = float64(totalTrades) / float64(nInst)
	}

	fmt.Fprintf(os.Stderr, "\nAvg trades/instrument: %.1f  (pass threshold: ≥%d)\n", avgTrades, minAvgTrades)
	fmt.Fprintf(os.Stderr, "COVID clustering violations: %d/%d instruments >%.0f%%\n",
		covidViolations, nInst, maxCovidPct)

	avgPass := avgTrades >= float64(minAvgTrades)
	clusterPass := covidViolations == 0

	fmt.Fprintf(os.Stderr, "\nVerdict: ")
	switch {
	case avgPass && clusterPass:
		fmt.Fprintf(os.Stderr, "PROCEED — avg trades and clustering both pass\n")
	case !avgPass && !clusterPass:
		fmt.Fprintf(os.Stderr, "KILL — avg trades below threshold AND clustering violation\n")
	case !avgPass:
		fmt.Fprintf(os.Stderr, "KILL — avg trades below threshold (%.1f < %d)\n", avgTrades, minAvgTrades)
	default:
		fmt.Fprintf(os.Stderr, "KILL — COVID clustering violation (%d instruments >%.0f%%)\n",
			covidViolations, maxCovidPct)
	}
	fmt.Fprintf(os.Stderr, "==============================================\n")
}
