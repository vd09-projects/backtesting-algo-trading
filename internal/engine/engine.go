package engine

import (
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// EngineConfig holds all parameters needed to run a backtest.
type EngineConfig struct {
	Instrument           string
	From                 time.Time
	To                   time.Time
	InitialCash          float64
	OrderConfig          model.OrderConfig
	PositionSizeFraction float64 // fraction of available cash to deploy per trade, e.g. 0.1 = 10%
}

// BarResult is the engine's record for a single processed bar.
// Signal is collected here; order execution is added in TASK-0004.
type BarResult struct {
	Candle model.Candle
	Signal model.Signal
}

// Engine runs a backtest by feeding candles from a DataProvider to a Strategy
// one bar at a time and collecting the resulting signals.
type Engine struct {
	config     EngineConfig
	barResults []BarResult
}

// New creates an Engine with the given config.
func New(cfg EngineConfig) *Engine {
	return &Engine{config: cfg}
}

// Results returns the per-bar results after Run completes.
// The slice is nil until Run is called.
func (e *Engine) Results() []BarResult {
	return e.barResults
}

// Run fetches candles from provider, then feeds them to strategy one bar at a time.
// It enforces:
//   - No-lookahead: strategy receives candles[:i+1] at bar i, never future bars.
//   - Lookback: strategy.Next is not called until at least strategy.Lookback() candles
//     have been seen. Bars before the lookback threshold are skipped silently.
func (e *Engine) Run(p provider.DataProvider, s strategy.Strategy) error {
	if e.config.Instrument == "" {
		return fmt.Errorf("engine: instrument must not be empty")
	}
	if e.config.From.IsZero() || e.config.To.IsZero() {
		return fmt.Errorf("engine: From and To must be set")
	}
	if !e.config.To.After(e.config.From) {
		return fmt.Errorf("engine: To (%s) must be after From (%s)", e.config.To, e.config.From)
	}

	candles, err := p.FetchCandles(e.config.Instrument, s.Timeframe(), e.config.From, e.config.To)
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

	e.barResults = make([]BarResult, 0, len(candles)-lookback+1)

	for i := range candles {
		// Enforce lookback: skip bars until we have enough history.
		if i+1 < lookback {
			continue
		}

		// No-lookahead: strategy only sees candles up to and including the current bar.
		signal := s.Next(candles[:i+1])

		e.barResults = append(e.barResults, BarResult{
			Candle: candles[i],
			Signal: signal,
		})
	}

	return nil
}
