package engine_test

// Golden tests for overnight gap handling in the engine.
//
// These tests verify that the engine fills at the actual next-bar Open price,
// including when that Open reflects an overnight gap (day 2 opens materially
// different from day 1 close). The engine is gap-transparent by construction:
// it feeds the pending signal into candles[i].Open with no smoothing or
// clamping. These tests pin that behavior so a future refactor cannot silently
// regress it.

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// TestGapDown_PositionEntered_PnLReflectsGap verifies that when a position is
// entered on day 1 and exited on day 2 via a gap-down open, the engine reports
// P&L based on the actual gap-down fill price, not any pre-gap price.
//
// Candle layout (5-min intraday, IST timestamps):
//
//	bar0 09:15 day1 — Buy signal emitted; no fill this bar.
//	bar1 09:20 day1 — Buy fills at Open=505; Hold signal.
//	bar2 09:25 day1 — Hold; position still open.
//	bar3 15:15 day1 — Sell signal emitted; no fill this bar (fills at next open).
//	bar4 09:15 day2 — Sell fills at Open=494 (gap-down from day1 close 509).
//
// Expected: EntryPrice=505, ExitPrice=494, RealizedPnL=(494-505)*qty < 0.
func TestGapDown_PositionEntered_PnLReflectsGap(t *testing.T) {
	ist, err := time.LoadLocation("Asia/Kolkata")
	require.NoError(t, err, "IST timezone must load")

	day1 := time.Date(2026, 1, 2, 0, 0, 0, 0, ist)
	day2 := time.Date(2026, 1, 3, 0, 0, 0, 0, ist)

	candles := []model.Candle{
		// bar0: signal bar — Buy emitted, fills at bar1's open.
		{
			Instrument: "NIFTY",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  day1.Add(9*time.Hour + 15*time.Minute),
			Open:       500,
			High:       502,
			Low:        499,
			Close:      500,
			Volume:     10000,
		},
		// bar1: Buy fills here at Open=505.
		{
			Instrument: "NIFTY",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  day1.Add(9*time.Hour + 20*time.Minute),
			Open:       505,
			High:       507,
			Low:        504,
			Close:      505,
			Volume:     12000,
		},
		// bar2: Hold; position open.
		{
			Instrument: "NIFTY",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  day1.Add(9*time.Hour + 25*time.Minute),
			Open:       506,
			High:       510,
			Low:        505,
			Close:      507,
			Volume:     11000,
		},
		// bar3: Sell signal emitted; fills at bar4's open.
		{
			Instrument: "NIFTY",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  day1.Add(15*time.Hour + 15*time.Minute),
			Open:       508,
			High:       511,
			Low:        507,
			Close:      509,
			Volume:     15000,
		},
		// bar4: gap-down open on day2; Sell fills at Open=494.
		// 494 is approximately 3% below day1 bar3 close of 509.
		{
			Instrument: "NIFTY",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  day2.Add(9*time.Hour + 15*time.Minute),
			Open:       494,
			High:       498,
			Low:        492,
			Close:      495,
			Volume:     20000,
		},
	}

	// Signals: Buy, Hold, Hold, Sell — then the trailing bar fills the Sell.
	// signalStrategy returns Hold once its slice is exhausted.
	signals := []model.Signal{
		model.SignalBuy,
		model.SignalHold,
		model.SignalHold,
		model.SignalSell,
	}

	cfg := engine.Config{
		Instrument:           "NIFTY",
		From:                 candles[0].Timestamp,
		To:                   candles[len(candles)-1].Timestamp.Add(time.Minute),
		InitialCash:          100_000,
		PositionSizeFraction: 1.0,
	}
	prov := &stubProvider{candles: candles}
	sty := &signalStrategy{signals: signals}

	e := engine.New(cfg)
	require.NoError(t, e.Run(context.Background(), prov, sty))

	port := e.Portfolio()
	closed := port.ClosedTrades()
	require.Len(t, closed, 1, "exactly one round-trip trade expected")

	tr := closed[0]

	// sizeFraction=1.0, no slippage, no commission: qty = cash / entryFill.
	expectedQty := 100_000.0 / 505.0
	expectedPnL := (494.0 - 505.0) * expectedQty

	assert.InDelta(t, 505.0, tr.EntryPrice, 1e-6, "entry price must be bar1 Open (505)")
	assert.InDelta(t, 494.0, tr.ExitPrice, 1e-6, "exit price must be bar4 gap-down Open (494)")
	assert.InDelta(t, expectedQty, tr.Quantity, 1e-6, "quantity")
	assert.InDelta(t, expectedPnL, tr.RealizedPnL, 1e-4, "P&L must reflect the gap-adjusted loss")
	assert.Less(t, tr.RealizedPnL, 0.0, "gap-down exit must produce a loss")
}

// TestGapDown_ExitFillsAtNextBarOpen_NotSignalBarClose verifies that a Sell
// signal emitted at bar N fills at bar N+1's Open, even when bar N+1 opens
// materially below bar N's Close. The fill price must be the actual gapped
// Open, not the signal bar's Close.
//
// Candle layout:
//
//	bar0 — Buy signal; fills at bar1's open.
//	bar1 — Buy fills at Open=500; close=600; Sell signal emitted.
//	bar2 — Sell fills at Open=582 (gap-down from close 600).
//
// Expected: ExitPrice=582, RealizedPnL=(582-500)*qty.
func TestGapDown_ExitFillsAtNextBarOpen_NotSignalBarClose(t *testing.T) {
	t0 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	candles := []model.Candle{
		// bar0: Buy signal; no fill here.
		{
			Instrument: "TEST",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  t0,
			Open:       500,
			High:       505,
			Low:        498,
			Close:      500,
			Volume:     1000,
		},
		// bar1: Buy fills at Open=500; Sell signal emitted (close=600).
		{
			Instrument: "TEST",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  t0.Add(24 * time.Hour),
			Open:       500,
			High:       610,
			Low:        498,
			Close:      600,
			Volume:     2000,
		},
		// bar2: Sell fills at Open=582 (gap-down from 600 close).
		{
			Instrument: "TEST",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  t0.Add(48 * time.Hour),
			Open:       582,
			High:       590,
			Low:        580,
			Close:      583,
			Volume:     1500,
		},
	}

	signals := []model.Signal{
		model.SignalBuy,
		model.SignalSell,
	}

	cfg := engine.Config{
		Instrument:           "TEST",
		From:                 t0,
		To:                   t0.AddDate(0, 0, 5),
		InitialCash:          10_000,
		PositionSizeFraction: 1.0,
	}
	prov := &stubProvider{candles: candles}
	sty := &signalStrategy{signals: signals}

	e := engine.New(cfg)
	require.NoError(t, e.Run(context.Background(), prov, sty))

	port := e.Portfolio()
	closed := port.ClosedTrades()
	require.Len(t, closed, 1, "exactly one round-trip trade expected")

	tr := closed[0]

	// sizeFraction=1.0, no slippage, no commission: qty = cash / entryFill.
	expectedQty := 10_000.0 / 500.0
	expectedPnL := (582.0 - 500.0) * expectedQty // NOT based on bar1 Close of 600.

	assert.InDelta(t, 500.0, tr.EntryPrice, 1e-6, "entry must be bar1 Open (500)")
	assert.InDelta(t, 582.0, tr.ExitPrice, 1e-6, "exit must be bar2 gapped Open (582), not bar1 Close (600)")
	assert.InDelta(t, expectedQty, tr.Quantity, 1e-6, "quantity")
	assert.InDelta(t, expectedPnL, tr.RealizedPnL, 1e-4, "P&L must reflect fill at gapped open, not signal bar close")
	assert.Greater(t, tr.RealizedPnL, 0.0, "trade is profitable despite gap-down exit")
}
