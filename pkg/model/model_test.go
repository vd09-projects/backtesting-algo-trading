package model_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ── Timeframe.Duration ────────────────────────────────────────────────────────

func TestTimeframeDuration(t *testing.T) {
	cases := []struct {
		tf      model.Timeframe
		want    time.Duration
		wantErr bool
	}{
		{model.Timeframe1Min, time.Minute, false},
		{model.Timeframe5Min, 5 * time.Minute, false},
		{model.Timeframe15Min, 15 * time.Minute, false},
		{model.TimeframeDaily, 24 * time.Hour, false},
		{model.TimeframeWeekly, 7 * 24 * time.Hour, false},
		{"bogus", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.tf.String(), func(t *testing.T) {
			got, err := tc.tf.Duration()
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestTimeframeString(t *testing.T) {
	assert.Equal(t, "daily", model.TimeframeDaily.String())
	assert.Equal(t, "1min", model.Timeframe1Min.String())
}

// ── Candle validation ─────────────────────────────────────────────────────────

var baseTime = time.Date(2026, 1, 2, 9, 15, 0, 0, time.UTC)

func TestNewCandleValid(t *testing.T) {
	c, err := model.NewCandle("INFY", model.TimeframeDaily, baseTime, 100, 110, 95, 105, 1000)
	require.NoError(t, err)
	assert.Equal(t, "INFY", c.Instrument)
	assert.Equal(t, 105.0, c.Close)
}

func TestNewCandleValidation(t *testing.T) {
	cases := []struct {
		name                         string
		instrument                   string
		open, high, low, cls, volume float64
		ts                           time.Time
		wantErrContains              string
	}{
		{
			name:       "empty instrument",
			instrument: "",
			open:       100, high: 110, low: 95, cls: 105, volume: 500,
			ts:              baseTime,
			wantErrContains: "instrument",
		},
		{
			name:       "zero timestamp",
			instrument: "RELIANCE",
			open:       100, high: 110, low: 95, cls: 105, volume: 500,
			ts:              time.Time{},
			wantErrContains: "timestamp",
		},
		{
			name:       "negative open",
			instrument: "TCS",
			open:       -1, high: 110, low: 95, cls: 105, volume: 500,
			ts:              baseTime,
			wantErrContains: "OHLC",
		},
		{
			name:       "high below low",
			instrument: "TCS",
			open:       100, high: 90, low: 95, cls: 97, volume: 500,
			ts:              baseTime,
			wantErrContains: "high",
		},
		{
			name:       "open above high",
			instrument: "TCS",
			open:       115, high: 110, low: 95, cls: 105, volume: 500,
			ts:              baseTime,
			wantErrContains: "open",
		},
		{
			name:       "close below low",
			instrument: "TCS",
			open:       100, high: 110, low: 95, cls: 90, volume: 500,
			ts:              baseTime,
			wantErrContains: "close",
		},
		{
			name:       "negative volume",
			instrument: "TCS",
			open:       100, high: 110, low: 95, cls: 105, volume: -1,
			ts:              baseTime,
			wantErrContains: "volume",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := model.NewCandle(tc.instrument, model.TimeframeDaily, tc.ts,
				tc.open, tc.high, tc.low, tc.cls, tc.volume)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrContains)
		})
	}
}

// ── Trade P&L ─────────────────────────────────────────────────────────────────

func TestTradeRealizedPnL(t *testing.T) {
	// RealizedPnL is stored directly on Trade; this test verifies it is preserved
	// faithfully and that the engine is responsible for computing it.
	cases := []struct {
		name        string
		entry, exit float64
		qty         float64
		direction   model.Direction
		commission  float64
		wantPnL     float64
	}{
		{
			name: "long profit",
			// buy 10 @ 100, sell @ 110, commission 5 → P&L = (110-100)*10 - 5 = 95
			entry: 100, exit: 110, qty: 10, direction: model.DirectionLong,
			commission: 5, wantPnL: 95,
		},
		{
			name: "long loss",
			// buy 10 @ 100, sell @ 90, commission 5 → P&L = (90-100)*10 - 5 = -105
			entry: 100, exit: 90, qty: 10, direction: model.DirectionLong,
			commission: 5, wantPnL: -105,
		},
		{
			name: "breakeven after commission",
			// buy 5 @ 200, sell @ 200, commission 10 → P&L = 0 - 10 = -10
			entry: 200, exit: 200, qty: 5, direction: model.DirectionLong,
			commission: 10, wantPnL: -10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// The engine computes P&L; we simulate that here to test the Trade struct
			// stores and returns it correctly.
			var pnl float64
			if tc.direction == model.DirectionLong {
				pnl = (tc.exit-tc.entry)*tc.qty - tc.commission
			} else {
				pnl = (tc.entry-tc.exit)*tc.qty - tc.commission
			}

			trade := model.Trade{
				Instrument:  "TEST",
				Direction:   tc.direction,
				Quantity:    tc.qty,
				EntryPrice:  tc.entry,
				ExitPrice:   tc.exit,
				EntryTime:   baseTime,
				ExitTime:    baseTime.Add(24 * time.Hour),
				RealizedPnL: pnl,
			}

			assert.InDelta(t, tc.wantPnL, trade.RealizedPnL, 1e-9)
		})
	}
}

// ── Trade.ReturnOnNotional ────────────────────────────────────────────────────

func TestReturnOnNotional(t *testing.T) {
	cases := []struct {
		name       string
		pnl        float64
		entryPrice float64
		qty        float64
		want       float64
	}{
		{
			name: "long profit",
			// pnl=50, entryPrice=100, qty=5 → 50/(100*5) = 0.10
			pnl: 50, entryPrice: 100, qty: 5, want: 0.10,
		},
		{
			name: "long loss",
			// pnl=-30, entryPrice=150, qty=2 → -30/(150*2) = -0.10
			pnl: -30, entryPrice: 150, qty: 2, want: -0.10,
		},
		{
			name: "small fractional return",
			// pnl=1, entryPrice=1000, qty=1 → 1/1000 = 0.001
			pnl: 1, entryPrice: 1000, qty: 1, want: 0.001,
		},
		{
			name: "zero entry price guard",
			pnl:  50, entryPrice: 0, qty: 5, want: 0,
		},
		{
			name: "zero quantity guard",
			pnl:  50, entryPrice: 100, qty: 0, want: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := model.Trade{
				Instrument:  "TEST",
				EntryPrice:  tc.entryPrice,
				Quantity:    tc.qty,
				RealizedPnL: tc.pnl,
			}
			assert.InDelta(t, tc.want, tr.ReturnOnNotional(), 1e-12)
		})
	}
}

// ── Signal ────────────────────────────────────────────────────────────────────

func TestSignalString(t *testing.T) {
	assert.Equal(t, "buy", model.SignalBuy.String())
	assert.Equal(t, "sell", model.SignalSell.String())
	assert.Equal(t, "hold", model.SignalHold.String())
}
