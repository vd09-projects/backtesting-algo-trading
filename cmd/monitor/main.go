// cmd/monitor is the weekly kill-switch monitoring CLI.
//
// It reads a live trade log (JSON array of model.Trade) and a pre-committed
// kill-switch thresholds file, evaluates the three halt conditions defined in
// decisions/algorithm/2026-04-21-kill-switch-derivation-methodology.md, and
// prints the alert status to stdout.
//
// Usage:
//
//	go run ./cmd/monitor \
//	    --trades live-trades.json \
//	    --thresholds decisions/algorithm/kill-switch-thresholds.json
//
// With explicit starting equity for the synthetic curve:
//
//	go run ./cmd/monitor \
//	    --trades live-trades.json \
//	    --thresholds decisions/algorithm/kill-switch-thresholds.json \
//	    --initial-equity 150000
//
// # Trade log format
//
// --trades must point to a JSON file containing an array of model.Trade values
// (as produced by cmd/backtest --out or manually appended live records).
// Direction is a string ("long" or "short"); timestamps are RFC 3339.
//
// # Thresholds file format
//
// --thresholds must point to a JSON file with the shape:
//
//	{"sharpe_p5": -0.05, "max_drawdown_pct": 4.10, "max_dd_duration_ns": 38649600000000000}
//
// max_dd_duration_ns is the duration in nanoseconds (time.Duration's JSON encoding).
// Use DeriveKillSwitchThresholds from internal/analytics to derive these from a
// bootstrap result and in-sample report, then serialize the fields manually.
//
// # Output
//
// OK — no threshold breached.
// HALT (Sharpe breached) — rolling per-trade Sharpe below p5 threshold.
// HALT (drawdown breached) — current drawdown exceeds 1.5× in-sample max DD.
// HALT (duration breached) — drawdown duration exceeds 2× in-sample max DD duration.
//
// Multiple HALT lines are printed when multiple thresholds are breached simultaneously.
//
// # Exit codes
//
//	0 — all thresholds OK
//	1 — one or more thresholds breached (enables cron / shell scripting)
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		var ee *exitCodeError
		if errors.As(err, &ee) {
			os.Exit(ee.code)
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// run is the testable entry point for cmd/monitor. It parses args, loads files,
// evaluates kill-switch thresholds, and writes alert output to stdout.
//
// **Decision (run() extraction for testability in cmd/monitor) — convention: experimental**
// scope: cmd/monitor
// tags: testability, flag-parse, coverage, run-function, TASK-0048
// owner: priya
//
// Follows the pattern established in cmd/walk-forward: main() is a one-liner;
// run() uses flag.NewFlagSet with ContinueOnError so tests can invoke it
// directly without spawning subprocesses. cmd/monitor has no DataProvider
// dependency — it is a pure file-reading CLI.
func run(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("monitor", flag.ContinueOnError)
	fs.SetOutput(stderr)

	tradesPath := fs.String("trades", "", "Path to live trade log JSON file (required)")
	thresholdsPath := fs.String("thresholds", "", "Path to kill-switch thresholds JSON file (required)")
	initialEquity := fs.Float64("initial-equity", 150_000, "Starting equity for synthetic curve construction (₹)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *tradesPath == "" {
		return fmt.Errorf("--trades is required")
	}
	if *thresholdsPath == "" {
		return fmt.Errorf("--thresholds is required")
	}

	trades, err := loadTrades(*tradesPath)
	if err != nil {
		return fmt.Errorf("load trades: %w", err)
	}

	thresholds, err := loadThresholds(*thresholdsPath)
	if err != nil {
		return fmt.Errorf("load thresholds: %w", err)
	}

	curve := buildSyntheticCurve(trades, *initialEquity)

	alert := analytics.CheckKillSwitch(trades, curve, thresholds)

	return writeAlertOutput(stdout, alert)
}

// thresholdsFile is the JSON serialization DTO for kill-switch thresholds.
//
// **Decision (thresholdsFile DTO in cmd/monitor, not JSON tags on KillSwitchThresholds) — architecture: experimental**
// scope: cmd/monitor, internal/analytics
// tags: serialization-dto, JSON-tags, package-boundary, pure-computation, TASK-0048
// owner: priya
//
// analytics.KillSwitchThresholds is a pure computation type with no JSON tags.
// Adding JSON tags there would couple the computation layer to a serialization
// concern. A local DTO in cmd/monitor is the correct boundary — the DTO
// handles the JSON layer, and a one-line conversion populates the analytics struct.
// MaxDDDuration is serialized as int64 nanoseconds (standard time.Duration JSON encoding).
type thresholdsFile struct {
	SharpeP5        float64 `json:"sharpe_p5"`
	MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
	MaxDDDurationNs int64   `json:"max_dd_duration_ns"`
}

// loadTrades reads and JSON-decodes a live trade log from path.
//
// **Decision (live trade log format: JSON array of model.Trade) — convention: experimental**
// scope: cmd/monitor
// tags: live-trade-log, JSON, file-format, kill-switch, TASK-0048
// owner: priya
//
// JSON array of model.Trade chosen over CSV. Direction is a string type;
// time.Time marshals as RFC 3339; all fields are plain types. The format
// matches what the engine already produces, so backtest output can be used
// directly as a live log seed. No custom schema or parser needed.
func loadTrades(path string) ([]model.Trade, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", path, err)
	}
	var trades []model.Trade
	if err := json.Unmarshal(data, &trades); err != nil {
		return nil, fmt.Errorf("parse %q: %w", path, err)
	}
	return trades, nil
}

// loadThresholds reads a thresholds JSON file and converts it to analytics.KillSwitchThresholds.
func loadThresholds(path string) (analytics.KillSwitchThresholds, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return analytics.KillSwitchThresholds{}, fmt.Errorf("read %q: %w", path, err)
	}
	var tf thresholdsFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return analytics.KillSwitchThresholds{}, fmt.Errorf("parse %q: %w", path, err)
	}
	return analytics.KillSwitchThresholds{
		SharpeP5:       tf.SharpeP5,
		MaxDrawdownPct: tf.MaxDrawdownPct,
		MaxDDDuration:  time.Duration(tf.MaxDDDurationNs),
	}, nil
}

// buildSyntheticCurve constructs a []model.EquityPoint from a trade list.
//
// **Decision (synthetic equity curve built from trades, not separate curve file) — tradeoff: experimental**
// scope: cmd/monitor
// tags: equity-curve, synthetic, trades, no-separate-input, TASK-0048
// owner: priya
//
// Curve built by sorting trades by ExitTime, then walking forward: equity[i] =
// initialEquity + sum(RealizedPnL[0..i]). No separate --curve input required.
// This covers the drawdown and duration checks from CheckKillSwitch with no
// additional user burden. A live mark-to-market curve would be more precise, but
// the trade-based curve is sufficient for weekly monitoring cadence and eliminates
// a required input that users rarely have at hand.
func buildSyntheticCurve(trades []model.Trade, initialEquity float64) []model.EquityPoint {
	if len(trades) == 0 {
		return nil
	}

	// Sort by ExitTime so the equity curve is time-ordered regardless of input order.
	sorted := make([]model.Trade, len(trades))
	copy(sorted, trades)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ExitTime.Before(sorted[j].ExitTime)
	})

	curve := make([]model.EquityPoint, len(sorted))
	equity := initialEquity
	for i, t := range sorted {
		equity += t.RealizedPnL
		curve[i] = model.EquityPoint{
			Timestamp: t.ExitTime,
			Value:     equity,
		}
	}
	return curve
}

// writeAlertOutput prints the alert status to w and returns an exitCodeError
// if any threshold is breached.
func writeAlertOutput(w io.Writer, alert analytics.KillSwitchAlert) error {
	breached := alert.SharpeBreached || alert.DrawdownBreached || alert.DurationBreached

	if !breached {
		fmt.Fprintln(w, "OK") //nolint:errcheck // stdout write; non-fatal
		return nil
	}

	if alert.SharpeBreached {
		fmt.Fprintln(w, "HALT (Sharpe breached)") //nolint:errcheck // stdout write; non-fatal
	}
	if alert.DrawdownBreached {
		fmt.Fprintln(w, "HALT (drawdown breached)") //nolint:errcheck // stdout write; non-fatal
	}
	if alert.DurationBreached {
		fmt.Fprintln(w, "HALT (duration breached)") //nolint:errcheck // stdout write; non-fatal
	}

	return &exitCodeError{code: 1}
}

// exitCodeError is returned by run() when a kill-switch threshold is breached.
// main() translates it to os.Exit(1).
type exitCodeError struct{ code int }

func (e *exitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.code)
}
