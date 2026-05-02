// Package signalaudit audits signal frequency across a matrix of strategies
// and instruments. Before running a full backtest pipeline, this package
// verifies that each strategy generates enough trades to produce statistically
// reliable metrics on each instrument.
//
// # Threshold
//
// MinTradesPerCell = 30. Cells below this threshold are marked Excluded.
// If a strategy's total trades across all instruments combined are also below
// this threshold, the strategy is marked Killed — it must not proceed to any
// full backtest run.
//
// # Concurrency model
//
// The audit fans out across (strategy, instrument) pairs using an errgroup
// with a GOMAXPROCS ceiling, identical to the universe sweep pattern. Each
// (strategy, instrument) pair is independent — no shared state, no ordering
// dependency. Results are written at fixed pre-allocated indices for
// deterministic output order.
package signalaudit

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"runtime"
	"strconv"

	"golang.org/x/sync/errgroup"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// MinTradesPerCell is the minimum number of closed trades a strategy must
// generate on a single instrument to be considered sufficient for reliable
// metrics on that instrument.
const MinTradesPerCell = 30

// Strategy is an alias for strategy.Strategy, re-exported so callers do not
// need to import both packages.
type Strategy = strategy.Strategy

// StrategyFactory describes a named strategy and a constructor function that
// produces a fresh instance. A factory is required (rather than a shared
// instance) to ensure each (strategy, instrument) run starts with clean state.
type StrategyFactory struct {
	Name string
	New  func() Strategy
}

// Config holds signal-audit run parameters.
type Config struct {
	StrategyFactories []StrategyFactory // one factory per strategy to audit
	Instruments       []string          // instruments to audit, in order
	EngineConfig      engine.Config     // template; Instrument field is overwritten per run
	Timeframe         model.Timeframe   // timeframe used for each engine run
}

// Cell holds the trade-count result for a single (strategy, instrument) pair.
type Cell struct {
	Instrument string
	TradeCount int
	Excluded   bool          // true when TradeCount < MinTradesPerCell
	Trades     []model.Trade // all closed trades; populated so callers can compute time-windowed metrics
}

// Row holds the audit result for a single strategy across all instruments.
type Row struct {
	Strategy    string
	TotalTrades int    // sum of TradeCount across all Cells
	Killed      bool   // true when TotalTrades < MinTradesPerCell
	Cells       []Cell // in the same order as Config.Instruments
}

// Report is the audit output: one Row per strategy, Instruments lists them in input order.
type Report struct {
	Instruments []string // column headers, in input order
	Rows        []Row
}

// Run executes the signal-frequency audit for all (strategy, instrument) pairs
// defined in cfg. It fans out via errgroup (GOMAXPROCS ceiling) and returns a
// Report with per-cell trade counts, exclusion flags, and per-strategy kill flags.
//
// A strategy factory is called once per instrument to produce a fresh strategy
// instance, ensuring no state leaks between instruments.
func Run(ctx context.Context, cfg *Config, p provider.DataProvider) (Report, error) {
	if len(cfg.StrategyFactories) == 0 {
		return Report{}, fmt.Errorf("signalaudit: strategy factories list must not be empty")
	}
	if len(cfg.Instruments) == 0 {
		return Report{}, fmt.Errorf("signalaudit: instruments list must not be empty")
	}

	nStrats := len(cfg.StrategyFactories)
	nInst := len(cfg.Instruments)

	// Pre-allocate results: rows[si][ii] = trade count for strategy si × instrument ii.
	// Using a flat slice of ints indexed as [si*nInst + ii] avoids 2D slice allocation.
	counts := make([]int, nStrats*nInst)

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.GOMAXPROCS(0))

	// Pre-allocate trade slices in parallel with counts.
	trades := make([][]model.Trade, nStrats*nInst)

	for si := range cfg.StrategyFactories {
		for ii := range cfg.Instruments {
			si, ii := si, ii // capture loop variables
			g.Go(func() error {
				tc, ts, err := runPair(gctx, cfg, p, si, ii)
				if err != nil {
					return fmt.Errorf("signalaudit: strategy %q instrument %q: %w",
						cfg.StrategyFactories[si].Name, cfg.Instruments[ii], err)
				}
				// Each goroutine writes to a unique (si, ii) index — no mutex needed.
				counts[si*nInst+ii] = tc
				trades[si*nInst+ii] = ts
				return nil
			})
		}
	}

	if err := g.Wait(); err != nil {
		return Report{}, err
	}

	// Build Report from counts and trades.
	rows := make([]Row, nStrats)
	for si, sf := range cfg.StrategyFactories {
		cells := make([]Cell, nInst)
		total := 0
		for ii, inst := range cfg.Instruments {
			tc := counts[si*nInst+ii]
			total += tc
			cells[ii] = Cell{
				Instrument: inst,
				TradeCount: tc,
				Excluded:   tc < MinTradesPerCell,
				Trades:     trades[si*nInst+ii],
			}
		}
		rows[si] = Row{
			Strategy:    sf.Name,
			TotalTrades: total,
			Killed:      total < MinTradesPerCell,
			Cells:       cells,
		}
	}

	return Report{
		Instruments: append([]string(nil), cfg.Instruments...),
		Rows:        rows,
	}, nil
}

// runPair runs a single engine instance for cfg.StrategyFactories[si] on
// cfg.Instruments[ii] and returns the closed trade count and the full trade slice.
// The trade slice lets callers compute time-windowed metrics (e.g. COVID-window
// clustering) without re-running the engine.
func runPair(ctx context.Context, cfg *Config, p provider.DataProvider, si, ii int) (int, []model.Trade, error) {
	engCfg := cfg.EngineConfig
	engCfg.Instrument = cfg.Instruments[ii]

	stgy := cfg.StrategyFactories[si].New()

	eng := engine.New(engCfg)
	if err := eng.Run(ctx, p, stgy); err != nil {
		return 0, nil, fmt.Errorf("engine run: %w", err)
	}

	closed := eng.Portfolio().ClosedTrades()
	return len(closed), append([]model.Trade(nil), closed...), nil
}

// WriteCSV writes the audit report as a CSV matrix to w.
//
// Header row: strategy, total_trades, killed, <instrument1>, <instrument2>, ...
// Each data row corresponds to one strategy. Trade counts are written as
// integers; cells with Excluded=true are written as "EXCLUDED(<count>)";
// killed strategies have their "killed" column written as "KILLED".
func WriteCSV(w io.Writer, r Report) error {
	cw := csv.NewWriter(w)

	// Build header.
	header := make([]string, 0, 3+len(r.Instruments))
	header = append(header, "strategy", "total_trades", "killed")
	header = append(header, r.Instruments...)
	if err := cw.Write(header); err != nil {
		return fmt.Errorf("signalaudit: write CSV header: %w", err)
	}

	for _, row := range r.Rows {
		killedStr := strconv.FormatBool(row.Killed)
		if row.Killed {
			killedStr = "KILLED"
		}
		rec := make([]string, 0, 3+len(row.Cells))
		rec = append(rec, row.Strategy, strconv.Itoa(row.TotalTrades), killedStr)
		for _, cell := range row.Cells {
			if cell.Excluded {
				rec = append(rec, fmt.Sprintf("EXCLUDED(%d)", cell.TradeCount))
			} else {
				rec = append(rec, strconv.Itoa(cell.TradeCount))
			}
		}
		if err := cw.Write(rec); err != nil {
			return fmt.Errorf("signalaudit: write CSV row for strategy %q: %w", row.Strategy, err)
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		return fmt.Errorf("signalaudit: flush CSV: %w", err)
	}
	return nil
}
