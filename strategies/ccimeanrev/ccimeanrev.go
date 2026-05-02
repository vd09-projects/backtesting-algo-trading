// Package ccimeanrev implements a CCI mean-reversion strategy.
//
// Entry: CCI falls below entryThreshold (e.g. -100, the conventional oversold level).
// Exit:  CCI crosses above exitThreshold (e.g. 0, neutral). The exit is a cross
//
//	detection — it fires only on the bar where CCI transitions from ≤ exitThreshold
//	to > exitThreshold. Long-only: no short positions.
//
// CCI is computed via github.com/markcheno/go-talib. The standard period is 20.
package ccimeanrev

import (
	"fmt"
	"math"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// Compile-time assertion: Strategy must satisfy the strategy.Strategy interface.
var _ strategy.Strategy = (*Strategy)(nil)

// Strategy emits signals based on CCI levels and crosses.
type Strategy struct {
	period         int
	entryThreshold float64 // CCI < entryThreshold → Buy
	exitThreshold  float64 // CCI crossing above exitThreshold → Sell
	timeframe      model.Timeframe
	prevCCI        float64 // CCI value from the previous Next() call; NaN on first call
}

// New constructs a CCI mean-reversion Strategy.
//
// Returns an error if:
//   - period < 1
//   - entryThreshold >= exitThreshold (entry must be strictly below exit)
//
// entryThreshold and exitThreshold are specified as ints matching the conventional
// CCI oversold/neutral levels (e.g. -100, 0). They are stored internally as float64
// for comparison with talib's float64 output.
func New(tf model.Timeframe, period, entryThreshold, exitThreshold int) (*Strategy, error) {
	if period < 1 {
		return nil, fmt.Errorf("ccimeanrev: period must be >= 1, got %d", period)
	}
	if entryThreshold >= exitThreshold {
		return nil, fmt.Errorf(
			"ccimeanrev: entryThreshold (%d) must be < exitThreshold (%d)",
			entryThreshold, exitThreshold,
		)
	}
	return &Strategy{
		period:         period,
		entryThreshold: float64(entryThreshold),
		exitThreshold:  float64(exitThreshold),
		timeframe:      tf,
		prevCCI:        math.NaN(),
	}, nil
}

// Name returns the strategy identifier.
func (s *Strategy) Name() string { return "cci-mean-reversion" }

// Timeframe returns the candle timeframe this strategy operates on.
func (s *Strategy) Timeframe() model.Timeframe { return s.timeframe }

// Lookback returns period.
//
// talib.Cci requires period - 1 bars to produce its first valid output
// (lookbackTotal = period - 1, so first valid index = period - 1 in a 0-based slice).
// Returning period here ensures the candle slice always has enough bars for a
// valid CCI computation on the last element.
func (s *Strategy) Lookback() int { return s.period }

// Next returns the signal for the current bar.
//
// Buy is emitted when CCI < entryThreshold.
// Sell is emitted when CCI crosses above exitThreshold (prevCCI <= exitThreshold
// AND currCCI > exitThreshold). The cross detection requires at least two
// calls to Next() before a Sell can fire.
// Hold is returned for all other cases, including before Lookback() bars are available.
func (s *Strategy) Next(candles []model.Candle) model.Signal {
	n := len(candles)
	if n < s.Lookback() {
		return model.SignalHold
	}

	highs := make([]float64, n)
	lows := make([]float64, n)
	closes := make([]float64, n)
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}

	cciVals := talib.Cci(highs, lows, closes, s.period)
	currCCI := cciVals[n-1]

	prevCCI := s.prevCCI
	s.prevCCI = currCCI

	// Entry: CCI below the oversold threshold.
	if currCCI < s.entryThreshold {
		return model.SignalBuy
	}

	// Exit: CCI crossing above the exit threshold.
	// Cross requires prevCCI to be known (not NaN) and ≤ exitThreshold,
	// while currCCI > exitThreshold.
	if !math.IsNaN(prevCCI) && prevCCI <= s.exitThreshold && currCCI > s.exitThreshold {
		return model.SignalSell
	}

	return model.SignalHold
}
