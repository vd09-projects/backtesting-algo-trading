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
type BarResult struct {
	Candle model.Candle
	Signal model.Signal
}

// Engine runs a backtest by feeding candles from a DataProvider to a Strategy
// one bar at a time, collecting signals and updating portfolio state.
type Engine struct {
	config    EngineConfig
	barResults []BarResult
	portfolio  *Portfolio
}

// New creates an Engine with the given config.
func New(cfg EngineConfig) *Engine {
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

	e.portfolio = newPortfolio(e.config.InitialCash)
	e.barResults = make([]BarResult, 0, len(candles)-lookback+1)

	for i := range candles {
		if i+1 < lookback {
			continue
		}

		signal := s.Next(candles[:i+1])

		if err := e.portfolio.applySignal(
			signal,
			e.config.Instrument,
			candles[i].Close,
			candles[i].Timestamp,
			e.config.PositionSizeFraction,
		); err != nil {
			return fmt.Errorf("engine: bar %d: %w", i, err)
		}

		e.barResults = append(e.barResults, BarResult{
			Candle: candles[i],
			Signal: signal,
		})
	}

	return nil
}
