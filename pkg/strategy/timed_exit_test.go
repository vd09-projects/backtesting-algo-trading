package strategy_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// scriptedStrategy is an unexported test fake that emits a pre-defined sequence
// of signals. Index i is returned on the i-th call to Next. If calls exceed the
// script length, SignalHold is returned.
type scriptedStrategy struct {
	script []model.Signal
	calls  int
}

func (s *scriptedStrategy) Name() string               { return "scripted" }
func (s *scriptedStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *scriptedStrategy) Lookback() int              { return 1 }
func (s *scriptedStrategy) Next(_ []model.Candle) model.Signal {
	if s.calls >= len(s.script) {
		s.calls++
		return model.SignalHold
	}
	sig := s.script[s.calls]
	s.calls++
	return sig
}

// compile-time check: scriptedStrategy satisfies strategy.Strategy.
var _ strategy.Strategy = (*scriptedStrategy)(nil)

// makeCandles returns n daily candles on instrument TEST:X all at close=100.
// Timestamps start at 2024-01-01 and advance one day per bar.
func makeCandles(n int) []model.Candle {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	cs := make([]model.Candle, n)
	for i := range cs {
		cs[i] = model.Candle{
			Instrument: "TEST:X",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       100,
			High:       100,
			Low:        100,
			Close:      100,
			Volume:     1000,
		}
	}
	return cs
}

// --- Metadata tests ---

func TestTimedExit_Name(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewTimedExit(inner, 3)
	require.Equal(t, "timed-exit(scripted)", w.Name())
}

func TestTimedExit_Lookback_delegatesToInner(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewTimedExit(inner, 5)
	require.Equal(t, inner.Lookback(), w.Lookback())
}

func TestTimedExit_Timeframe_delegatesToInner(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewTimedExit(inner, 5)
	require.Equal(t, inner.Timeframe(), w.Timeframe())
}

// --- Acceptance criteria tests (TDD — written before implementation) ---

// TestTimedExit_InnerSellBeforeTimer: inner sell fires at bar 2 (maxHoldBars=5).
// Timer must not fire — inner sell is honored first and resets state.
// After the sell, the wrapper is free to enter again on the next Buy.
func TestTimedExit_InnerSellBeforeTimer(t *testing.T) {
	// Script: Buy(bar0), Hold(bar1), Sell(bar2), Hold(bar3)...
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter
			model.SignalHold, // bar 1: still in
			model.SignalSell, // bar 2: inner exits (timer would fire at bar 5)
			model.SignalHold, // bar 3: out of position
		},
	}
	w := strategy.NewTimedExit(inner, 5)
	all := makeCandles(4)

	got0 := w.Next(all[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: inner Buy → wrapper Buy")

	got1 := w.Next(all[:2])
	require.Equal(t, model.SignalHold, got1, "bar 1: inner Hold → wrapper Hold (timer not yet)")

	got2 := w.Next(all[:3])
	require.Equal(t, model.SignalSell, got2, "bar 2: inner Sell fires before timer → wrapper Sell")

	// After sell, state is reset — bar 3 should Hold (not in position)
	got3 := w.Next(all[:4])
	require.Equal(t, model.SignalHold, got3, "bar 3: out of position after inner sell")
}

// TestTimedExit_TimerFires: inner never sells; timer must emit Sell at bar maxHoldBars.
func TestTimedExit_TimerFires(t *testing.T) {
	maxHoldBars := 3
	// Script: Buy(bar0), Hold, Hold, Hold...
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter
			model.SignalHold, // bar 1
			model.SignalHold, // bar 2
			model.SignalHold, // bar 3 — timer fires here
		},
	}
	w := strategy.NewTimedExit(inner, maxHoldBars)
	all := makeCandles(4)

	got0 := w.Next(all[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: enter")

	got1 := w.Next(all[:2])
	require.Equal(t, model.SignalHold, got1, "bar 1: holding (1 bar since entry)")

	got2 := w.Next(all[:3])
	require.Equal(t, model.SignalHold, got2, "bar 2: holding (2 bars since entry)")

	got3 := w.Next(all[:4])
	require.Equal(t, model.SignalSell, got3, "bar 3: timer fires (3 bars since entry == maxHoldBars)")
}

// TestTimedExit_ExactBoundary: sell fires at exactly maxHoldBars bars after entry,
// not at maxHoldBars-1 or maxHoldBars+1.
// Uses maxHoldBars=2 and checks bar1 (should Hold) and bar2 (should Sell).
func TestTimedExit_ExactBoundary(t *testing.T) {
	maxHoldBars := 2
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter
			model.SignalHold, // bar 1: 1 bar since entry < maxHoldBars=2 → Hold
			model.SignalHold, // bar 2: 2 bars since entry == maxHoldBars=2 → Sell
		},
	}
	w := strategy.NewTimedExit(inner, maxHoldBars)
	all := makeCandles(3)

	got0 := w.Next(all[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: enter")

	// bar 1: barsSinceEntry=1 < maxHoldBars=2 → Hold (NOT early Sell)
	got1 := w.Next(all[:2])
	require.Equal(t, model.SignalHold, got1, "bar 1: barsSinceEntry=1 < 2, must NOT sell yet")

	// bar 2: barsSinceEntry=2 >= maxHoldBars=2 → Sell (NOT deferred to bar 3)
	got2 := w.Next(all[:3])
	require.Equal(t, model.SignalSell, got2, "bar 2: barsSinceEntry=2 >= 2, timer fires exactly here")
}

// TestTimedExit_NoPositionDuringWarmup: if inner never buys, wrapper must return
// whatever inner returns (Hold) and never emit a Sell.
func TestTimedExit_NoPositionDuringWarmup(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalHold,
			model.SignalHold,
			model.SignalHold,
			model.SignalHold,
			model.SignalHold,
		},
	}
	w := strategy.NewTimedExit(inner, 3)
	all := makeCandles(5)

	for i := 1; i <= 5; i++ {
		got := w.Next(all[:i])
		require.Equal(t, model.SignalHold, got, "bar %d: no position — must Hold", i)
	}
}

// TestTimedExit_ReEnterAfterTimerSell: after timer fires and resets state,
// wrapper can enter again on next Buy from inner.
func TestTimedExit_ReEnterAfterTimerSell(t *testing.T) {
	maxHoldBars := 2
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: first entry
			model.SignalHold, // bar 1
			model.SignalHold, // bar 2: timer fires → Sell
			model.SignalBuy,  // bar 3: re-enter
			model.SignalHold, // bar 4
			model.SignalHold, // bar 5: timer fires again → Sell
		},
	}
	w := strategy.NewTimedExit(inner, maxHoldBars)
	all := makeCandles(6)

	require.Equal(t, model.SignalBuy, w.Next(all[:1]), "bar 0: first entry")
	require.Equal(t, model.SignalHold, w.Next(all[:2]), "bar 1: holding")
	require.Equal(t, model.SignalSell, w.Next(all[:3]), "bar 2: timer fires")
	require.Equal(t, model.SignalBuy, w.Next(all[:4]), "bar 3: re-enter")
	require.Equal(t, model.SignalHold, w.Next(all[:5]), "bar 4: holding again")
	require.Equal(t, model.SignalSell, w.Next(all[:6]), "bar 5: timer fires again")
}
