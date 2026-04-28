package signalaudit_test

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/signalaudit"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// Test fakes
// ---------------------------------------------------------------------------

// staticProvider returns a flat candle series (price 100) for any instrument.
type staticProvider struct{}

func (p *staticProvider) FetchCandles(
	_ context.Context,
	instrument string,
	_ model.Timeframe,
	from, to time.Time,
) ([]model.Candle, error) {
	var candles []model.Candle
	for ts := from; ts.Before(to); ts = ts.Add(24 * time.Hour) {
		candles = append(candles, model.Candle{
			Instrument: instrument,
			Timeframe:  model.TimeframeDaily,
			Timestamp:  ts,
			Open:       100,
			High:       101,
			Low:        99,
			Close:      100,
			Volume:     1000,
		})
	}
	return candles, nil
}

func (p *staticProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.TimeframeDaily}
}

// toggleStrategy alternates Buy/Sell every other candle, guaranteeing closed trades.
// Lookback is 2 so the engine needs at least 2 candles before producing signals.
type toggleStrategy struct{ name string }

func (t *toggleStrategy) Name() string               { return t.name }
func (t *toggleStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (t *toggleStrategy) Lookback() int              { return 2 }
func (t *toggleStrategy) Next(candles []model.Candle) model.Signal {
	if len(candles)%2 == 0 {
		return model.SignalBuy
	}
	return model.SignalSell
}

// holdStrategy never trades.
type holdStrategy struct{ name string }

func (h *holdStrategy) Name() string               { return h.name }
func (h *holdStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (h *holdStrategy) Lookback() int              { return 1 }
func (h *holdStrategy) Next(_ []model.Candle) model.Signal {
	return model.SignalHold
}

func baseEngineConfig(from, to time.Time) engine.Config {
	return engine.Config{
		From:                 from,
		To:                   to,
		InitialCash:          100_000,
		PositionSizeFraction: 0.10,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	}
}

// ---------------------------------------------------------------------------
// Config validation
// ---------------------------------------------------------------------------

func TestRun_ReturnsErrorOnEmptyStrategies(t *testing.T) {
	t.Parallel()

	from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: nil,
		Instruments:       []string{"NSE:RELIANCE"},
		EngineConfig:      baseEngineConfig(from, to),
		Timeframe:         model.TimeframeDaily,
	}
	_, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err == nil {
		t.Fatal("expected error for empty strategy factories, got nil")
	}
}

func TestRun_ReturnsErrorOnEmptyInstruments(t *testing.T) {
	t.Parallel()

	from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  nil,
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}
	_, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err == nil {
		t.Fatal("expected error for empty instruments, got nil")
	}
}

// ---------------------------------------------------------------------------
// Matrix shape
// ---------------------------------------------------------------------------

func TestRun_MatrixHasCorrectDimensions(t *testing.T) {
	t.Parallel()

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	strategies := []signalaudit.StrategyFactory{
		{Name: "toggle-a", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle-a"} }},
		{Name: "toggle-b", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle-b"} }},
	}
	instruments := []string{"NSE:RELIANCE", "NSE:INFY", "NSE:TCS"}

	cfg := signalaudit.Config{
		StrategyFactories: strategies,
		Instruments:       instruments,
		EngineConfig:      baseEngineConfig(from, to),
		Timeframe:         model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	if len(report.Rows) != len(strategies) {
		t.Fatalf("expected %d rows (one per strategy), got %d", len(strategies), len(report.Rows))
	}
	for _, row := range report.Rows {
		if len(row.Cells) != len(instruments) {
			t.Fatalf("strategy %q: expected %d cells, got %d", row.Strategy, len(instruments), len(row.Cells))
		}
	}
}

// ---------------------------------------------------------------------------
// EXCLUDED flag (< 30 trades per cell)
// ---------------------------------------------------------------------------

func TestRun_CellExcludedWhenTradeCountBelowThreshold(t *testing.T) {
	t.Parallel()

	// 35 days → at most ~17 trades from toggleStrategy; below MinTradesForAudit (30).
	from := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 2, 5, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  []string{"NSE:RELIANCE"},
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if len(report.Rows) != 1 || len(report.Rows[0].Cells) != 1 {
		t.Fatalf("unexpected report shape: %d rows", len(report.Rows))
	}
	cell := report.Rows[0].Cells[0]
	if !cell.Excluded {
		t.Errorf("expected cell to be Excluded=true for short window (trade count %d < 30)", cell.TradeCount)
	}
}

func TestRun_CellNotExcludedWhenTradeCountAtThreshold(t *testing.T) {
	t.Parallel()

	// 2 years → toggleStrategy generates ~365 trades; well above threshold.
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  []string{"NSE:RELIANCE"},
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	cell := report.Rows[0].Cells[0]
	if cell.Excluded {
		t.Errorf("expected cell to NOT be Excluded; trade count = %d", cell.TradeCount)
	}
}

// ---------------------------------------------------------------------------
// Strategy kill flag (< 30 trades across entire universe combined)
// ---------------------------------------------------------------------------

func TestRun_StrategyKilledWhenTotalTradesBelowThreshold(t *testing.T) {
	t.Parallel()

	// holdStrategy never fires any trades → total = 0 across all instruments.
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "hold", New: func() signalaudit.Strategy { return &holdStrategy{name: "hold"} }},
		},
		Instruments:  []string{"NSE:RELIANCE", "NSE:INFY"},
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	if len(report.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(report.Rows))
	}
	row := report.Rows[0]
	if !row.Killed {
		t.Errorf("expected holdStrategy to be Killed=true (0 total trades), got Killed=false; TotalTrades=%d", row.TotalTrades)
	}
}

func TestRun_StrategyNotKilledWhenTotalTradesAtThreshold(t *testing.T) {
	t.Parallel()

	// toggleStrategy generates many trades across 2 years × 2 instruments.
	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  []string{"NSE:RELIANCE", "NSE:INFY"},
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}
	row := report.Rows[0]
	if row.Killed {
		t.Errorf("expected toggleStrategy NOT killed; TotalTrades=%d", row.TotalTrades)
	}
}

// ---------------------------------------------------------------------------
// TotalTrades aggregation
// ---------------------------------------------------------------------------

func TestRun_TotalTradesIsAggregateAcrossAllInstruments(t *testing.T) {
	t.Parallel()

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	instruments := []string{"NSE:RELIANCE", "NSE:INFY", "NSE:TCS"}

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  instruments,
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	row := report.Rows[0]
	var sumFromCells int
	for _, c := range row.Cells {
		sumFromCells += c.TradeCount
	}
	if row.TotalTrades != sumFromCells {
		t.Errorf("TotalTrades=%d, want sum of cells=%d", row.TotalTrades, sumFromCells)
	}
}

// ---------------------------------------------------------------------------
// WriteCSV
// ---------------------------------------------------------------------------

func TestWriteCSV_HeaderMatchesInstruments(t *testing.T) {
	t.Parallel()

	instruments := []string{"NSE:RELIANCE", "NSE:INFY"}
	report := signalaudit.Report{
		Instruments: instruments,
		Rows: []signalaudit.Row{
			{
				Strategy:    "toggle",
				TotalTrades: 100,
				Killed:      false,
				Cells: []signalaudit.Cell{
					{Instrument: "NSE:RELIANCE", TradeCount: 50, Excluded: false},
					{Instrument: "NSE:INFY", TradeCount: 50, Excluded: false},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := signalaudit.WriteCSV(&buf, report); err != nil {
		t.Fatalf("WriteCSV: unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// header + 1 strategy row
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d:\n%s", len(lines), output)
	}

	header := lines[0]
	if !strings.HasPrefix(header, "strategy,total_trades,killed,") {
		t.Errorf("unexpected header prefix: %q", header)
	}
	for _, inst := range instruments {
		if !strings.Contains(header, inst) {
			t.Errorf("header missing instrument %q: %q", inst, header)
		}
	}
}

func TestWriteCSV_ExcludedCellWrittenAsEXCLUDED(t *testing.T) {
	t.Parallel()

	report := signalaudit.Report{
		Instruments: []string{"NSE:RELIANCE"},
		Rows: []signalaudit.Row{
			{
				Strategy:    "rsi",
				TotalTrades: 10,
				Killed:      true,
				Cells: []signalaudit.Cell{
					{Instrument: "NSE:RELIANCE", TradeCount: 10, Excluded: true},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := signalaudit.WriteCSV(&buf, report); err != nil {
		t.Fatalf("WriteCSV: unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "EXCLUDED") {
		t.Errorf("expected EXCLUDED in output for excluded cell, got:\n%s", output)
	}
	if !strings.Contains(output, "KILLED") {
		t.Errorf("expected KILLED in output for killed strategy, got:\n%s", output)
	}
}

func TestWriteCSV_EmptyReport(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	if err := signalaudit.WriteCSV(&buf, signalaudit.Report{}); err != nil {
		t.Fatalf("WriteCSV: unexpected error on empty report: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// Header only.
	if len(lines) != 1 {
		t.Fatalf("expected 1 line for empty report, got %d:\n%s", len(lines), output)
	}
}

// ---------------------------------------------------------------------------
// Instruments column ordering preserved
// ---------------------------------------------------------------------------

func TestRun_InstrumentOrderPreserved(t *testing.T) {
	t.Parallel()

	from := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	instruments := []string{"NSE:TITAN", "NSE:INFY", "NSE:RELIANCE"}

	cfg := signalaudit.Config{
		StrategyFactories: []signalaudit.StrategyFactory{
			{Name: "toggle", New: func() signalaudit.Strategy { return &toggleStrategy{name: "toggle"} }},
		},
		Instruments:  instruments,
		EngineConfig: baseEngineConfig(from, to),
		Timeframe:    model.TimeframeDaily,
	}

	report, err := signalaudit.Run(context.Background(), &cfg, &staticProvider{})
	if err != nil {
		t.Fatalf("Run: unexpected error: %v", err)
	}

	row := report.Rows[0]
	for i, inst := range instruments {
		if row.Cells[i].Instrument != inst {
			t.Errorf("cell[%d].Instrument = %q, want %q", i, row.Cells[i].Instrument, inst)
		}
	}
}
