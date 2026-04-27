package strategy

import (
	"fmt"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// TimedExit wraps an inner Strategy and forces a sell after maxHoldBars bars
// regardless of the inner strategy's signal. It is useful for 1–7 day hold
// strategies where indefinite hold-until-reversal is undesirable.
//
// Signal priority on each call to Next:
//  1. If not in position: delegate to inner. On SignalBuy, record entry bar
//     and enter position.
//  2. If in position: honor inner SignalSell first (resets state).
//  3. If in position and barsSinceEntry >= maxHoldBars: emit SignalSell (resets
//     state) regardless of what inner returns.
//  4. Otherwise: delegate to inner.
//
// # Statefulness
//
// TimedExit maintains mutable state (entryBar, inPosition) between calls.
// It is NOT safe for concurrent use across walk-forward folds or goroutines.
// Callers that run folds in parallel must construct a fresh TimedExit per fold.
type TimedExit struct {
	inner       Strategy
	maxHoldBars int
	inPosition  bool
	entryBar    int // bar index (len(candles)-1) when position was entered
}

// NewTimedExit constructs a TimedExit that wraps inner and forces a sell
// after maxHoldBars bars since entry.
func NewTimedExit(inner Strategy, maxHoldBars int) Strategy {
	return &TimedExit{
		inner:       inner,
		maxHoldBars: maxHoldBars,
	}
}

// Name returns "timed-exit(<inner.Name()>)".
func (t *TimedExit) Name() string {
	return fmt.Sprintf("timed-exit(%s)", t.inner.Name())
}

// Lookback delegates to the inner strategy's Lookback.
func (t *TimedExit) Lookback() int { return t.inner.Lookback() }

// Timeframe delegates to the inner strategy's Timeframe.
func (t *TimedExit) Timeframe() model.Timeframe { return t.inner.Timeframe() }

// Next returns the signal for the current bar.
//
// If not in position, it delegates to the inner strategy. A SignalBuy from
// inner opens the position and records the entry bar index.
//
// If in position, inner's SignalSell is honored first (exits before timer).
// If barsSinceEntry >= maxHoldBars, a forced SignalSell is emitted regardless
// of inner's signal. Otherwise inner's signal is returned (Hold in practice,
// since the engine skips redundant Buy signals under no-pyramiding).
func (t *TimedExit) Next(candles []model.Candle) model.Signal {
	innerSig := t.inner.Next(candles)
	currentBar := len(candles) - 1

	if !t.inPosition {
		if innerSig == model.SignalBuy {
			t.inPosition = true
			t.entryBar = currentBar
		}
		return innerSig
	}

	// In position: honor inner sell first.
	if innerSig == model.SignalSell {
		t.inPosition = false
		t.entryBar = 0
		return model.SignalSell
	}

	// Timer check: bars held since entry (not including entry bar itself).
	barsSinceEntry := currentBar - t.entryBar
	if barsSinceEntry >= t.maxHoldBars {
		t.inPosition = false
		t.entryBar = 0
		return model.SignalSell
	}

	return innerSig
}
