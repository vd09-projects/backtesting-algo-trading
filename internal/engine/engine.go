// Package engine implements the backtesting execution loop.
package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Config holds all parameters needed to run a backtest.
type Config struct {
	Instrument           string
	From                 time.Time
	To                   time.Time
	InitialCash          float64
	OrderConfig          model.OrderConfig
	PositionSizeFraction float64 // fraction of available cash to deploy per trade, e.g. 0.1 = 10%
}

// BarResult is the engine's record for a single processed bar.
type BarResult struct {
	Candle model.Candle
	Signal model.Signal
}

// Engine runs a backtest by feeding candles from a DataProvider to a Strategy
// one bar at a time, collecting signals and updating portfolio state.
type Engine struct {
	config     Config
	barResults []BarResult
	portfolio  *Portfolio
}

// New creates an Engine with the given config.
func New(cfg Config) *Engine { //nolint:gocritic // Config is a constructor arg; value semantics are intentional
	return &Engine{config: cfg}
}

// Results returns the per-bar results after Run completes.
func (e *Engine) Results() []BarResult {
	return e.barResults
}

// Portfolio returns the portfolio state after Run completes.
// Returns nil until Run is called.
func (e *Engine) Portfolio() *Portfolio {
	return e.portfolio
}

// Run fetches candles from provider, feeds them to strategy one bar at a time,
// and applies each signal to the portfolio. It enforces:
//   - No-lookahead: strategy receives candles[:i+1] at bar i, never future bars.
//   - Lookback: strategy.Next is not called until at least strategy.Lookback() candles
//     have been seen.
func (e *Engine) Run(ctx context.Context, p provider.DataProvider, s strategy.Strategy) error {
	if e.config.Instrument == "" {
		return fmt.Errorf("engine: instrument must not be empty")
	}
	if e.config.From.IsZero() || e.config.To.IsZero() {
		return fmt.Errorf("engine: From and To must be set")
	}
	if !e.config.To.After(e.config.From) {
		return fmt.Errorf("engine: To (%s) must be after From (%s)", e.config.To, e.config.From)
	}

	candles, err := p.FetchCandles(ctx, e.config.Instrument, s.Timeframe(), e.config.From, e.config.To)
	if err != nil {
		return fmt.Errorf("engine: fetching candles: %w", err)
	}
	if len(candles) == 0 {
		return fmt.Errorf("engine: provider returned no candles for %s", e.config.Instrument)
	}

	lookback := s.Lookback()
	if lookback < 1 {
		return fmt.Errorf("engine: strategy %q declared lookback %d, must be >= 1", s.Name(), lookback)
	}

	e.portfolio = newPortfolio(e.config.InitialCash, e.config.OrderConfig, len(candles))
	e.barResults = make([]BarResult, 0, len(candles)-lookback+1)

	// pendingSignal holds the signal from the previous bar, to be filled at
	// the current bar's open. This enforces the rule: market orders fill at
	// the next candle's open, not the current bar's close.
	pendingSignal := model.SignalHold

	for i := range candles {
		// Apply the previous bar's signal at this bar's open price.
		if pendingSignal != model.SignalHold {
			if err := e.portfolio.applySignal(
				pendingSignal,
				e.config.Instrument,
				candles[i].Open,
				candles[i].Timestamp,
				e.config.PositionSizeFraction,
			); err != nil {
				return fmt.Errorf("engine: fill at bar %d: %w", i, err)
			}
			pendingSignal = model.SignalHold
		}

		// Snapshot equity at this bar's close (after fill, before next signal).
		e.portfolio.RecordEquity(candles[i])

		if i+1 < lookback {
			continue
		}

		signal := s.Next(candles[:i+1])
		pendingSignal = signal

		e.barResults = append(e.barResults, BarResult{
			Candle: candles[i],
			Signal: signal,
		})
	}

	return nil
}
