package strategy_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// makeVariableCandles returns n daily candles with the given Close values.
// len(closes) must equal n. All other OHLC fields are set to the Close value.
// Timestamps start at 2024-06-01 and advance one day per bar.
func makeVariableCandles(closes []float64) []model.Candle {
	base := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	cs := make([]model.Candle, len(closes))
	for i, c := range closes {
		cs[i] = model.Candle{
			Instrument: "TEST:X",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       c,
			High:       c,
			Low:        c,
			Close:      c,
			Volume:     1000,
		}
	}
	return cs
}

// --- Metadata tests ---

func TestPriceExit_Name(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewPriceExit(inner, 0.05, 0.10)
	require.Equal(t, "price-exit(scripted)", w.Name())
}

func TestPriceExit_Lookback_delegatesToInner(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewPriceExit(inner, 0.05, 0.10)
	require.Equal(t, inner.Lookback(), w.Lookback())
}

func TestPriceExit_Timeframe_delegatesToInner(t *testing.T) {
	inner := &scriptedStrategy{}
	w := strategy.NewPriceExit(inner, 0.05, 0.10)
	require.Equal(t, inner.Timeframe(), w.Timeframe())
}

// --- Golden tests ---

// TestPriceExit_StopLossFires: inner buys at bar0 (Close=100), bar1 price drops to SL
// threshold exactly (100*(1-0.05)=95.0). SL must fire at exact boundary (<=).
// bar2 price drops further to confirm SL already reset state.
func TestPriceExit_StopLossFires(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter
			model.SignalHold, // bar 1: SL fires (inner not called)
			model.SignalHold, // bar 2: out of position
		},
	}
	// SL=5%, TP=disabled; entry at Close=100; SL threshold = 100*(1-0.05) = 95.0
	w := strategy.NewPriceExit(inner, 0.05, 0)
	candles := makeVariableCandles([]float64{100, 95.0, 90})

	got0 := w.Next(candles[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: inner Buy → wrapper Buy, entryPrice=100")

	// bar 1: Close=95.0 == entryPrice*(1-0.05) = 95.0 → SL fires (<=)
	callsBefore := inner.calls
	got1 := w.Next(candles[:2])
	require.Equal(t, model.SignalSell, got1, "bar 1: Close==SL threshold (95.0) → SL fires at exact boundary")
	// inner.Next() must NOT have been called when SL fires
	require.Equal(t, callsBefore, inner.calls, "bar 1: inner must not be called when SL fires")

	// bar 2: out of position after SL reset — inner hold passes through
	got2 := w.Next(candles[:3])
	require.Equal(t, model.SignalHold, got2, "bar 2: out of position after SL, inner hold passes through")
}

// TestPriceExit_TargetProfitFires: inner buys at bar0 (Close=100), bar1 price rises to
// TP threshold exactly (100*(1+0.20)=120.0). TP must fire at exact boundary (>=).
// 20% is used because 100*1.20 is exactly representable in float64; 100*1.10 is not.
func TestPriceExit_TargetProfitFires(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter
			model.SignalHold, // bar 1: TP fires (inner not called)
			model.SignalHold, // bar 2: out of position
		},
	}
	// SL=disabled, TP=20%; entry at Close=100; TP threshold = 100*(1+0.20) = 120.0 (exact float64)
	w := strategy.NewPriceExit(inner, 0, 0.20)
	candles := makeVariableCandles([]float64{100, 120.0, 125})

	got0 := w.Next(candles[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: inner Buy → wrapper Buy, entryPrice=100")

	// bar 1: Close=120.0 == entryPrice*(1+0.20) = 120.0 → TP fires (>=)
	callsBefore := inner.calls
	got1 := w.Next(candles[:2])
	require.Equal(t, model.SignalSell, got1, "bar 1: Close==TP threshold (120.0) → TP fires at exact boundary")
	require.Equal(t, callsBefore, inner.calls, "bar 1: inner must not be called when TP fires")

	// bar 2: out of position after TP reset
	got2 := w.Next(candles[:3])
	require.Equal(t, model.SignalHold, got2, "bar 2: out of position after TP, inner hold passes through")
}

// TestPriceExit_InnerExitsBeforeSLTP: price never hits SL or TP; inner sells first.
// Inner sell must pass through and reset state.
func TestPriceExit_InnerExitsBeforeSLTP(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter at 100
			model.SignalHold, // bar 1: price 102, no trigger
			model.SignalSell, // bar 2: inner exits (price 101, SL would be at 95, TP at 110)
			model.SignalHold, // bar 3: out of position
		},
	}
	// SL=5%, TP=10%: neither fires given price 100→102→101
	w := strategy.NewPriceExit(inner, 0.05, 0.10)
	candles := makeVariableCandles([]float64{100, 102, 101, 100})

	got0 := w.Next(candles[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: enter")

	got1 := w.Next(candles[:2])
	require.Equal(t, model.SignalHold, got1, "bar 1: price 102, no SL/TP, inner Hold passes through")

	got2 := w.Next(candles[:3])
	require.Equal(t, model.SignalSell, got2, "bar 2: inner Sell passes through (price 101, SL/TP not hit)")

	got3 := w.Next(candles[:4])
	require.Equal(t, model.SignalHold, got3, "bar 3: out of position after inner sell")
}

// TestPriceExit_ReEntryAfterSLReset: after SL fires, state resets; inner can buy again.
// Verifies entryPrice is updated to the new entry bar's Close, not the old one.
//
// Important: when SL fires, inner.Next is NOT called. So the inner script's index
// does not advance on SL bars. Script[0]=Buy (bar0), script[1]=Buy (bar2 re-entry,
// since inner was skipped on bar1), script[2]=Hold (bar3), bar4 SL fires → inner skipped.
func TestPriceExit_ReEntryAfterSLReset(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // call 0 → bar 0: first entry at 100
			model.SignalBuy,  // call 1 → bar 2: re-enter (bar 1 inner was skipped — SL fired)
			model.SignalHold, // call 2 → bar 3: hold
			// bar 4: SL fires → inner skipped
		},
	}
	// SL=5%, TP=disabled
	w := strategy.NewPriceExit(inner, 0.05, 0)
	// bar0=100, bar1=93 (SL fires: 93 < 100*0.95=95), bar2=90 (re-entry, entryPrice=90),
	// bar3=86 (86 > 90*0.95=85.5 → no SL), bar4=85 (85 <= 85.5 → SL fires)
	candles := makeVariableCandles([]float64{100, 93, 90, 86, 85})

	got0 := w.Next(candles[:1])
	require.Equal(t, model.SignalBuy, got0, "bar 0: first entry at 100")

	got1 := w.Next(candles[:2])
	require.Equal(t, model.SignalSell, got1, "bar 1: SL fires (93 < 95)")

	// bar 2: not in position after SL reset; inner.Next() returns Buy → re-enter
	got2 := w.Next(candles[:3])
	require.Equal(t, model.SignalBuy, got2, "bar 2: re-entry after SL reset, entryPrice=90")

	// bar 3: price=86 > 90*0.95=85.5 → no SL
	got3 := w.Next(candles[:4])
	require.Equal(t, model.SignalHold, got3, "bar 3: price 86 > new SL threshold 85.5 → Hold")

	// bar 4: price=85 <= 90*0.95=85.5 → SL fires on re-entry position
	got4 := w.Next(candles[:5])
	require.Equal(t, model.SignalSell, got4, "bar 4: SL fires on re-entry position (85 <= 85.5)")
}

// TestPriceExit_BothDisabled: stopLossPct=0 and targetProfitPct=0 — pure passthrough.
// PriceExit must delegate all signals to inner with no SL/TP intervention.
func TestPriceExit_BothDisabled(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,
			model.SignalHold,
			model.SignalHold,
			model.SignalSell,
			model.SignalHold,
		},
	}
	w := strategy.NewPriceExit(inner, 0, 0)
	// Use prices that would trigger SL/TP if they were enabled: drop below 95, rise above 110
	candles := makeVariableCandles([]float64{100, 94, 112, 98, 97})

	require.Equal(t, model.SignalBuy, w.Next(candles[:1]), "bar 0: Buy passes through")
	require.Equal(t, model.SignalHold, w.Next(candles[:2]), "bar 1: price 94 — SL disabled, Hold passes through")
	require.Equal(t, model.SignalHold, w.Next(candles[:3]), "bar 2: price 112 — TP disabled, Hold passes through")
	require.Equal(t, model.SignalSell, w.Next(candles[:4]), "bar 3: inner Sell passes through")
	require.Equal(t, model.SignalHold, w.Next(candles[:5]), "bar 4: out of position, Hold")
}

// TestPriceExit_OnlySLEnabled: stopLossPct>0, targetProfitPct=0.
// TP guard must be inactive; price rising above TP level must NOT trigger a sell.
// SL must still fire.
func TestPriceExit_OnlySLEnabled(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // bar 0: enter at 100
			model.SignalHold, // bar 1: price 115 — TP would fire if enabled (>110), must not
			model.SignalHold, // bar 2: price 120 — still no TP exit
			model.SignalHold, // bar 3: price 93 — SL fires (93 < 95)
			model.SignalHold, // bar 4: out of position
		},
	}
	// SL=5%, TP=disabled (0)
	w := strategy.NewPriceExit(inner, 0.05, 0)
	candles := makeVariableCandles([]float64{100, 115, 120, 93, 90})

	require.Equal(t, model.SignalBuy, w.Next(candles[:1]), "bar 0: enter at 100")
	require.Equal(t, model.SignalHold, w.Next(candles[:2]), "bar 1: price 115 — TP disabled, must NOT sell")
	require.Equal(t, model.SignalHold, w.Next(candles[:3]), "bar 2: price 120 — TP disabled, must NOT sell")
	require.Equal(t, model.SignalSell, w.Next(candles[:4]), "bar 3: SL fires (93 < 95)")
	require.Equal(t, model.SignalHold, w.Next(candles[:5]), "bar 4: out of position")
}

// TestPriceExit_OnlyTPEnabled: stopLossPct=0, targetProfitPct>0.
// SL guard must be inactive; price dropping below SL level must NOT trigger a sell.
// TP must still fire.
// Uses 20% TP (threshold 120.0) because 100*1.20 is exactly representable in float64.
func TestPriceExit_OnlyTPEnabled(t *testing.T) {
	inner := &scriptedStrategy{
		script: []model.Signal{
			model.SignalBuy,  // call 0 → bar 0: enter at 100
			model.SignalHold, // call 1 → bar 1: price 80 — SL would fire if enabled, must not
			model.SignalHold, // call 2 → bar 2: price 85 — still no SL exit
			// bar 3: TP fires at 120 → inner skipped
			model.SignalHold, // call 3 → bar 4: out of position
		},
	}
	// SL=disabled (0), TP=20%; TP threshold = 100*1.20 = 120.0 (exact float64)
	w := strategy.NewPriceExit(inner, 0, 0.20)
	candles := makeVariableCandles([]float64{100, 80, 85, 120, 105})

	require.Equal(t, model.SignalBuy, w.Next(candles[:1]), "bar 0: enter at 100")
	require.Equal(t, model.SignalHold, w.Next(candles[:2]), "bar 1: price 80 — SL disabled, must NOT sell")
	require.Equal(t, model.SignalHold, w.Next(candles[:3]), "bar 2: price 85 — SL disabled, must NOT sell")
	require.Equal(t, model.SignalSell, w.Next(candles[:4]), "bar 3: TP fires (120 >= 100*1.20=120.0 exact)")
	require.Equal(t, model.SignalHold, w.Next(candles[:5]), "bar 4: out of position")
}
