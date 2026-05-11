package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/signalaudit"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// TestValidateFlagInputs
// ---------------------------------------------------------------------------

func TestValidateFlagInputs_ValidInput(t *testing.T) {
	t.Parallel()
	flags, err := validateFlagInputs(
		"universes/nifty50-large-cap.yaml",
		"2018-01-01",
		"2024-01-01",
		"",
		100000, 0.10, 0.0005,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !flags.from.Equal(time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("from: got %s", flags.from)
	}
	if !flags.to.Equal(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("to: got %s", flags.to)
	}
	if flags.universeFile != "universes/nifty50-large-cap.yaml" {
		t.Errorf("universeFile: got %q", flags.universeFile)
	}
}

func TestValidateFlagInputs_MissingUniverse(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("", "2018-01-01", "2024-01-01", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error for missing --universe")
	}
	if !strings.Contains(err.Error(), "--universe") {
		t.Errorf("error should mention --universe, got: %v", err)
	}
}

func TestValidateFlagInputs_MissingFrom(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "", "2024-01-01", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error for missing --from")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestValidateFlagInputs_MissingTo(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "2018-01-01", "", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error for missing --to")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestValidateFlagInputs_InvalidFromDate(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "not-a-date", "2024-01-01", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error for invalid --from date")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

func TestValidateFlagInputs_InvalidToDate(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "2018-01-01", "bad", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error for invalid --to date")
	}
	if !strings.Contains(err.Error(), "--to") {
		t.Errorf("error should mention --to, got: %v", err)
	}
}

func TestValidateFlagInputs_ToNotAfterFrom(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "2024-01-01", "2018-01-01", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error when --to is not after --from")
	}
	if !strings.Contains(err.Error(), "--to must be strictly after") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateFlagInputs_ToEqualFrom(t *testing.T) {
	t.Parallel()
	_, err := validateFlagInputs("universes/test.yaml", "2024-01-01", "2024-01-01", "", 100000, 0.10, 0.0005)
	if err == nil {
		t.Fatal("expected error when --to equals --from")
	}
}

func TestValidateFlagInputs_OutPathPreserved(t *testing.T) {
	t.Parallel()
	flags, err := validateFlagInputs("universes/test.yaml", "2018-01-01", "2024-01-01", "runs/audit.csv", 100000, 0.10, 0.0005)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flags.outPath != "runs/audit.csv" {
		t.Errorf("outPath: got %q, want %q", flags.outPath, "runs/audit.csv")
	}
}

// ---------------------------------------------------------------------------
// TestBuildEngineConfig
// ---------------------------------------------------------------------------

func TestBuildEngineConfig_FieldsSet(t *testing.T) {
	t.Parallel()
	from := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	f := &cliFlags{
		from:         from,
		to:           to,
		cash:         250000,
		positionSize: 0.05,
		slippage:     0.001,
	}
	cfg := buildEngineConfig(f)
	if !cfg.From.Equal(from) {
		t.Errorf("From: got %s, want %s", cfg.From, from)
	}
	if !cfg.To.Equal(to) {
		t.Errorf("To: got %s, want %s", cfg.To, to)
	}
	if cfg.InitialCash != 250000 {
		t.Errorf("InitialCash: got %.0f, want 250000", cfg.InitialCash)
	}
	if cfg.PositionSizeFraction != 0.05 {
		t.Errorf("PositionSizeFraction: got %.4f, want 0.05", cfg.PositionSizeFraction)
	}
	if cfg.OrderConfig.SlippagePct != 0.001 {
		t.Errorf("SlippagePct: got %.4f, want 0.001", cfg.OrderConfig.SlippagePct)
	}
	if cfg.OrderConfig.CommissionModel != model.CommissionZerodhaFull {
		t.Errorf("CommissionModel: got %q, want ZerodhaFull", cfg.OrderConfig.CommissionModel)
	}
}

// ---------------------------------------------------------------------------
// TestCountCovidTrades
// ---------------------------------------------------------------------------

func TestCountCovidTrades_TradesInWindow(t *testing.T) {
	t.Parallel()
	cell := signalaudit.Cell{
		Trades: []model.Trade{
			{ExitTime: time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)},  // inside
			{ExitTime: time.Date(2020, 6, 30, 0, 0, 0, 0, time.UTC)},  // inside (< exclusive upper)
			{ExitTime: time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)}, // before window
			{ExitTime: time.Date(2020, 7, 1, 0, 0, 0, 0, time.UTC)},   // at exclusive upper — outside
		},
	}
	got := countCovidTrades(cell)
	if got != 2 {
		t.Errorf("countCovidTrades: got %d, want 2", got)
	}
}

func TestCountCovidTrades_NoTrades(t *testing.T) {
	t.Parallel()
	got := countCovidTrades(signalaudit.Cell{})
	if got != 0 {
		t.Errorf("countCovidTrades empty cell: got %d, want 0", got)
	}
}

func TestCountCovidTrades_NoTradesInWindow(t *testing.T) {
	t.Parallel()
	cell := signalaudit.Cell{
		Trades: []model.Trade{
			{ExitTime: time.Date(2019, 12, 31, 0, 0, 0, 0, time.UTC)},
			{ExitTime: time.Date(2021, 6, 15, 0, 0, 0, 0, time.UTC)},
		},
	}
	got := countCovidTrades(cell)
	if got != 0 {
		t.Errorf("countCovidTrades: got %d, want 0", got)
	}
}

// ---------------------------------------------------------------------------
// TestSummariseReport
// ---------------------------------------------------------------------------

func TestSummariseReport_KilledAndExcluded(t *testing.T) {
	t.Parallel()
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{
				Strategy:    "sma-crossover",
				TotalTrades: 500,
				Killed:      false,
				Cells: []signalaudit.Cell{
					{Excluded: false},
					{Excluded: true},
				},
			},
			{
				Strategy:    "rsi-mean-reversion",
				TotalTrades: 5,
				Killed:      true,
				Cells: []signalaudit.Cell{
					{Excluded: true},
					{Excluded: true},
				},
			},
		},
	}
	killed, excluded := summariseReport(report)
	if killed != 1 {
		t.Errorf("killed: got %d, want 1", killed)
	}
	if excluded != 3 {
		t.Errorf("excluded: got %d, want 3", excluded)
	}
}

func TestSummariseReport_NoneKilled(t *testing.T) {
	t.Parallel()
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{Strategy: "sma-crossover", TotalTrades: 200, Killed: false},
		},
	}
	killed, excluded := summariseReport(report)
	if killed != 0 {
		t.Errorf("killed: got %d, want 0", killed)
	}
	if excluded != 0 {
		t.Errorf("excluded: got %d, want 0", excluded)
	}
}

// ---------------------------------------------------------------------------
// TestAllStrategyFactories
// ---------------------------------------------------------------------------

func TestAllStrategyFactories_ReturnsOneFactoryPerAuditEntry(t *testing.T) {
	t.Parallel()
	factories := allStrategyFactories(model.TimeframeDaily)
	if len(factories) == 0 {
		t.Fatal("allStrategyFactories returned empty slice")
	}
	// Every factory must have a non-empty name and a non-nil New function.
	for _, f := range factories {
		if f.Name == "" {
			t.Error("factory has empty Name")
		}
		if f.New == nil {
			t.Errorf("factory %q has nil New function", f.Name)
		}
		// Verify the factory actually constructs a strategy without panic.
		s := f.New()
		if s == nil {
			t.Errorf("factory %q returned nil strategy", f.Name)
		}
	}
}

func TestAllStrategyFactories_NamesMatchAuditParamOverrides(t *testing.T) {
	t.Parallel()
	factories := allStrategyFactories(model.TimeframeDaily)
	for _, f := range factories {
		if _, ok := auditParamOverrides[f.Name]; !ok {
			t.Errorf("factory name %q not found in auditParamOverrides", f.Name)
		}
	}
}

// ---------------------------------------------------------------------------
// TestPrintCCICells
// ---------------------------------------------------------------------------

func TestPrintCCICells_OutputAndCounts(t *testing.T) {
	t.Parallel()
	cells := []signalaudit.Cell{
		{
			Instrument: "NSE:RELIANCE",
			TradeCount: 100,
			Trades: []model.Trade{
				{ExitTime: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC)}, // COVID
				{ExitTime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)}, // not COVID
			},
		},
		{
			Instrument: "NSE:INFY",
			TradeCount: 80,
			Trades:     nil, // no trades → 0% COVID
		},
	}

	var buf strings.Builder
	totalTrades, covidViolations := printCCICells(&buf, cells, 30.0)

	if totalTrades != 180 {
		t.Errorf("totalTrades: got %d, want 180", totalTrades)
	}
	// RELIANCE has 1/100 = 1% COVID — not clustered. INFY has 0% — not clustered.
	if covidViolations != 0 {
		t.Errorf("covidViolations: got %d, want 0", covidViolations)
	}
	if !strings.Contains(buf.String(), "NSE:RELIANCE") {
		t.Error("output missing NSE:RELIANCE")
	}
	if !strings.Contains(buf.String(), "NSE:INFY") {
		t.Error("output missing NSE:INFY")
	}
}

func TestPrintCCICells_ClusterFlaggedWhenAboveThreshold(t *testing.T) {
	t.Parallel()
	// All 3 trades are in COVID window → 100% > 30% threshold → clustered.
	cells := []signalaudit.Cell{
		{
			Instrument: "NSE:TCS",
			TradeCount: 3,
			Trades: []model.Trade{
				{ExitTime: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)},
				{ExitTime: time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)},
				{ExitTime: time.Date(2020, 4, 1, 0, 0, 0, 0, time.UTC)},
			},
		},
	}
	var buf strings.Builder
	_, violations := printCCICells(&buf, cells, 30.0)
	if violations != 1 {
		t.Errorf("covidViolations: got %d, want 1", violations)
	}
	if !strings.Contains(buf.String(), "CLUSTERED") {
		t.Error("expected CLUSTERED flag in output")
	}
}

// ---------------------------------------------------------------------------
// TestPrintCCIDistributionReport
// ---------------------------------------------------------------------------

func TestPrintCCIDistributionReport_NoCCIRow(t *testing.T) {
	t.Parallel()
	// Report with no cci-mean-reversion row — function should return without writing.
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{Strategy: "sma-crossover", TotalTrades: 200},
		},
	}
	var buf strings.Builder
	printCCIDistributionReport(&buf, report)
	if buf.Len() != 0 {
		t.Errorf("expected no output when CCI row absent, got: %q", buf.String())
	}
}

func TestPrintCCIDistributionReport_WithCCIRow(t *testing.T) {
	t.Parallel()
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{
				Strategy:    "cci-mean-reversion",
				TotalTrades: 50,
				Cells: []signalaudit.Cell{
					{Instrument: "NSE:RELIANCE", TradeCount: 30},
					{Instrument: "NSE:INFY", TradeCount: 20},
				},
			},
		},
	}
	var buf strings.Builder
	printCCIDistributionReport(&buf, report)
	out := buf.String()
	if !strings.Contains(out, "CCI Mean-Reversion Distribution Report") {
		t.Error("expected report header in output")
	}
	if !strings.Contains(out, "Verdict") {
		t.Error("expected Verdict line in output")
	}
}

// ---------------------------------------------------------------------------
// TestWriteReport
// ---------------------------------------------------------------------------

func TestWriteReport_NoKillsWritesCSVAndSummary(t *testing.T) {
	t.Parallel()
	report := signalaudit.Report{
		Instruments: []string{"NSE:RELIANCE"},
		Rows: []signalaudit.Row{
			{
				Strategy:    "sma-crossover",
				TotalTrades: 200,
				Killed:      false,
				Cells:       []signalaudit.Cell{{Instrument: "NSE:RELIANCE", TradeCount: 200}},
			},
		},
	}

	var csvOut, errOut strings.Builder
	// No file output path — writes to csvOut directly.
	writeReport(&csvOut, &errOut, report, "", 1)

	if !strings.Contains(csvOut.String(), "sma-crossover") {
		t.Error("CSV output missing sma-crossover row")
	}
	if !strings.Contains(errOut.String(), "Summary:") {
		t.Error("stderr output missing Summary line")
	}
}

func TestWriteReport_WithOutPathWritesToFile(t *testing.T) {
	t.Parallel()
	report := signalaudit.Report{
		Instruments: []string{"NSE:RELIANCE"},
		Rows: []signalaudit.Row{
			{
				Strategy:    "sma-crossover",
				TotalTrades: 150,
				Killed:      false,
				Cells:       []signalaudit.Cell{{Instrument: "NSE:RELIANCE", TradeCount: 150}},
			},
		},
	}

	outPath := filepath.Join(t.TempDir(), "audit.csv")

	var errOut strings.Builder
	// csvOut is ignored when outPath is set — output goes to the file.
	writeReport(&errOut, &errOut, report, outPath, 1)

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(content), "sma-crossover") {
		t.Errorf("file missing sma-crossover row, got: %q", string(content))
	}
}

// TestWriteReport_ExitsOneOnKilledStrategies verifies writeReport calls os.Exit(1)
// when any strategy is killed. Uses the subprocess pattern because os.Exit cannot
// be tested in-process.
func TestWriteReport_ExitsOneOnKilledStrategies(t *testing.T) {
	if os.Getenv("SIGNAL_AUDIT_EXIT_TEST") == "1" {
		report := signalaudit.Report{
			Rows: []signalaudit.Row{
				{Strategy: "dead-strategy", TotalTrades: 0, Killed: true, Cells: []signalaudit.Cell{}},
			},
		}
		writeReport(os.Stdout, os.Stderr, report, "", 1)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestWriteReport_ExitsOneOnKilledStrategies$")
	cmd.Env = append(os.Environ(), "SIGNAL_AUDIT_EXIT_TEST=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1 when strategies killed, got: %v", err)
	}
}

func TestPrintCCIDistributionReport_KillLowTrades(t *testing.T) {
	t.Parallel()
	// avg trades = 5/instrument < 25 threshold → KILL
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{
				Strategy:    "cci-mean-reversion",
				TotalTrades: 5,
				Cells: []signalaudit.Cell{
					{Instrument: "NSE:RELIANCE", TradeCount: 5},
				},
			},
		},
	}
	var buf strings.Builder
	printCCIDistributionReport(&buf, report)
	if !strings.Contains(buf.String(), "KILL") {
		t.Errorf("expected KILL verdict, got: %q", buf.String())
	}
}

func TestPrintCCIDistributionReport_KillCovidViolation(t *testing.T) {
	t.Parallel()
	// avg trades = 30 (passes), but 100% COVID → KILL on clustering
	report := signalaudit.Report{
		Rows: []signalaudit.Row{
			{
				Strategy:    "cci-mean-reversion",
				TotalTrades: 30,
				Cells: []signalaudit.Cell{
					{
						Instrument: "NSE:RELIANCE",
						TradeCount: 30,
						// 25 out of 30 trades in COVID window → 83% > 30% threshold → CLUSTERED.
						// TradeCount matches len(Trades) — both drive the covidPct calculation.
						Trades: []model.Trade{
							// 25 COVID-window trades (March 2020).
							{ExitTime: time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 3, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 4, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 5, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 7, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 8, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 10, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 11, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 12, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 13, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 14, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 15, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 16, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 17, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 18, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 19, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 20, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 21, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 22, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 23, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 24, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2020, 3, 25, 0, 0, 0, 0, time.UTC)},
							// 5 non-COVID trades (2021).
							{ExitTime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2021, 4, 1, 0, 0, 0, 0, time.UTC)},
							{ExitTime: time.Date(2021, 5, 1, 0, 0, 0, 0, time.UTC)},
						},
					},
				},
			},
		},
	}
	var buf strings.Builder
	printCCIDistributionReport(&buf, report)
	out := buf.String()
	if !strings.Contains(out, "KILL") {
		t.Errorf("expected KILL verdict for COVID clustering, got: %q", out)
	}
}
