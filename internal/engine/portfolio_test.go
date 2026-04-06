package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// signalStrategy emits signals from a fixed sequence, then Hold for any bar beyond the slice.
type signalStrategy struct {
	signals []model.Signal
	i       int
}

func (s *signalStrategy) Name() string               { return "signal-seq" }
func (s *signalStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *signalStrategy) Lookback() int              { return 1 }
func (s *signalStrategy) Next(_ []model.Candle) model.Signal {
	if s.i >= len(s.signals) {
		return model.SignalHold
	}
	sig := s.signals[s.i]
	s.i++
	return sig
}

// ── helpers ───────────────────────────────────────────────────────────────────

// runWithSignals runs the engine with a fixed signal sequence and returns the final portfolio.
// One extra candle is appended beyond len(signals) so the last signal can fill at the next open.
// makeCandles produces: Open = 100+i, High = 105+i, Low = 95+i, Close = 102+i.
func runWithSignals(t *testing.T, initialCash, sizeFraction float64, cfg model.OrderConfig, signals []model.Signal) *engine.Portfolio {
	t.Helper()
	n := len(signals) + 1 // trailing candle so the final signal can fill
	candles := makeCandles(n)

	ecfg := engine.Config{
		Instrument:           "TEST",
		From:                 base,
		To:                   base.AddDate(0, 0, n+1),
		InitialCash:          initialCash,
		OrderConfig:          cfg,
		PositionSizeFraction: sizeFraction,
	}
	s := &signalStrategy{signals: signals}
	p := &stubProvider{candles: candles}

	e := engine.New(ecfg)
	require.NoError(t, e.Run(context.Background(), p, s))
	return e.Portfolio()
}

// ── existing behavior tests ──────────────────────────────────────────────────

func TestPortfolio_BuyOpensPosition(t *testing.T) {
	// Signal=Buy at bar 0 → fills at bar 1's open (101). sizeFraction=1.0 deploys all cash.
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{model.SignalBuy})

	assert.Len(t, port.Positions, 1)
	pos, ok := port.Positions["TEST"]
	require.True(t, ok)
	assert.Equal(t, model.DirectionLong, pos.Direction)
	assert.Greater(t, pos.Quantity, 0.0)
	assert.InDelta(t, 0.0, port.Cash, 1e-6)
}

func TestPortfolio_SellClosesPosition(t *testing.T) {
	// Buy at bar 0 → fills at bar 1 open (101). Sell at bar 1 → fills at bar 2 open (102).
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{model.SignalBuy, model.SignalSell})

	assert.Empty(t, port.Positions, "position should be closed after sell")
	closed := closedTrades(port)
	require.Len(t, closed, 1)

	tr := closed[0]
	assert.Equal(t, "TEST", tr.Instrument)
	assert.Equal(t, model.DirectionLong, tr.Direction)
	assert.False(t, tr.ExitTime.IsZero())
	// Entry at 101, exit at 102 — profitable.
	assert.Greater(t, tr.RealizedPnL, 0.0, "profitable trade expected")
}

func TestPortfolio_PnLCalculation(t *testing.T) {
	// Controlled candles: signal=Buy at bar 0, Buy fills at bar 1 open=100, Sell fills at bar 2 open=110.
	// qty = 1000/100 = 10; PnL = (110-100)*10 = 100.
	candles := []model.Candle{
		{
			Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: base,
			Open: 99, High: 105, Low: 95, Close: 99, Volume: 1000,
		},
		{
			Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: base.Add(24 * time.Hour),
			Open: 100, High: 105, Low: 99, Close: 100, Volume: 1000,
		},
		{
			Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: base.Add(48 * time.Hour),
			Open: 110, High: 115, Low: 109, Close: 110, Volume: 1000,
		},
	}

	cfg := engine.Config{
		Instrument:           "X",
		From:                 base,
		To:                   base.AddDate(0, 0, 5),
		InitialCash:          1000,
		PositionSizeFraction: 1.0,
	}
	s := &signalStrategy{signals: []model.Signal{model.SignalBuy, model.SignalSell}}
	prov := &stubProvider{candles: candles}

	e := engine.New(cfg)
	require.NoError(t, e.Run(context.Background(), prov, s))
	port := e.Portfolio()

	closed := closedTrades(port)
	require.Len(t, closed, 1)

	tr := closed[0]
	assert.InDelta(t, 10.0, tr.Quantity, 1e-6)
	assert.InDelta(t, 100.0, tr.EntryPrice, 1e-6)
	assert.InDelta(t, 110.0, tr.ExitPrice, 1e-6)
	assert.InDelta(t, 100.0, tr.RealizedPnL, 1e-6)
}

func TestPortfolio_CashInsufficientNoBuy(t *testing.T) {
	// sizeFraction=0 → cost=0 → buy skipped
	port := runWithSignals(t, 10_000, 0.0, model.OrderConfig{}, []model.Signal{model.SignalBuy})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
}

func TestPortfolio_SellWithNoPositionIsNoop(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{model.SignalSell})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
	assert.Empty(t, closedTrades(port))
}

func TestPortfolio_DoubleBuyDoesNotPyramid(t *testing.T) {
	// Two consecutive buys: first fills at bar 1, second is skipped (position already open).
	port := runWithSignals(t, 10_000, 0.5, model.OrderConfig{}, []model.Signal{model.SignalBuy, model.SignalBuy})

	assert.Len(t, port.Positions, 1)
	assert.Len(t, openTrades(port), 1)
}

func TestPortfolio_BuySellBuyOpensSecondPosition(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{
		model.SignalBuy, model.SignalSell, model.SignalBuy,
	})

	assert.Len(t, port.Positions, 1)
	assert.Len(t, closedTrades(port), 1)
}

func TestPortfolio_CommissionPushesOverCashNoBuy(t *testing.T) {
	// Cash=100, sizeFraction=1.0, flat commission=5.
	// cost=100, qty=1, fillPrice=100, entryCommission=5, totalCost=105 > 100 → buy skipped.
	port := runWithSignals(t, 100, 1.0,
		model.OrderConfig{CommissionModel: model.CommissionFlat, CommissionValue: 5},
		[]model.Signal{model.SignalBuy},
	)

	assert.Empty(t, port.Positions, "buy must be skipped when commission pushes total cost over cash")
	assert.InDelta(t, 100.0, port.Cash, 1e-6)
}

func TestPortfolio_HoldIsNoop(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{
		model.SignalHold, model.SignalHold,
	})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
}

func TestPortfolio_TradeLogTimestamps(t *testing.T) {
	// Entry fills at bar 1's open; exit fills at bar 2's open.
	port := runWithSignals(t, 10_000, 1.0, model.OrderConfig{}, []model.Signal{model.SignalBuy, model.SignalSell})

	closed := closedTrades(port)
	require.Len(t, closed, 1)
	tr := closed[0]

	assert.True(t, tr.EntryTime.Before(tr.ExitTime), "entry must precede exit")
}

// ── slippage and commission table tests ───────────────────────────────────────

func TestPortfolio_SlippageAndCommission(t *testing.T) {
	// Fixed candles: signal issued at bar 0. Buy fills at bar 1 (open=100). Sell fills at bar 2 (open=110).
	makeTestCandles := func() []model.Candle {
		t0 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		return []model.Candle{
			{
				Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: t0,
				Open: 99, High: 105, Low: 95, Close: 99, Volume: 1000,
			},
			{
				Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: t0.Add(24 * time.Hour),
				Open: 100, High: 105, Low: 99, Close: 100, Volume: 1000,
			},
			{
				Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: t0.Add(48 * time.Hour),
				Open: 110, High: 115, Low: 109, Close: 110, Volume: 1000,
			},
		}
	}

	cases := []struct {
		name      string
		cash      float64
		sizeFrac  float64
		orderCfg  model.OrderConfig
		wantQty   float64
		wantEntry float64
		wantExit  float64
		wantComm  float64
		wantPnL   float64
		delta     float64
	}{
		{
			name: "no slippage no commission",
			cash: 1000, sizeFrac: 1.0,
			orderCfg: model.OrderConfig{},
			// qty = 1000/100 = 10; PnL = (110-100)*10 = 100
			wantQty: 10.0, wantEntry: 100.0, wantExit: 110.0,
			wantComm: 0.0, wantPnL: 100.0, delta: 1e-6,
		},
		{
			name: "slippage 0.5% only",
			cash: 1000, sizeFrac: 1.0,
			orderCfg: model.OrderConfig{SlippagePct: 0.005},
			// entryFill = 100 * 1.005 = 100.5; qty = 1000/100.5; exitFill = 110 * 0.995 = 109.45
			// PnL = (109.45 - 100.5) * (1000/100.5)
			wantQty:   1000.0 / 100.5,
			wantEntry: 100.5, wantExit: 109.45,
			wantComm: 0.0,
			wantPnL:  (109.45 - 100.5) * (1000.0 / 100.5),
			delta:    1e-6,
		},
		{
			name: "flat commission ₹10 per fill",
			cash: 10_000, sizeFrac: 0.1,
			orderCfg: model.OrderConfig{CommissionModel: model.CommissionFlat, CommissionValue: 10},
			// cost = 1000; qty = 1000/100 = 10; entryComm = 10; exitComm = 10; total comm = 20
			// PnL = (110-100)*10 - 20 = 80
			wantQty: 10.0, wantEntry: 100.0, wantExit: 110.0,
			wantComm: 20.0, wantPnL: 80.0, delta: 1e-6,
		},
		{
			name: "percentage commission 0.1% per fill",
			cash: 10_000, sizeFrac: 0.1,
			orderCfg: model.OrderConfig{CommissionModel: model.CommissionPercentage, CommissionValue: 0.001},
			// cost = 1000; qty = 10; entryComm = 0.001*1000 = 1.0; exitComm = 0.001*1100 = 1.1
			// PnL = 100 - 2.1 = 97.9
			wantQty: 10.0, wantEntry: 100.0, wantExit: 110.0,
			wantComm: 2.1, wantPnL: 97.9, delta: 1e-6,
		},
		{
			name: "zerodha commission — small trade (0.03% < ₹20)",
			cash: 10_000, sizeFrac: 0.1,
			orderCfg: model.OrderConfig{CommissionModel: model.CommissionZerodha},
			// cost = 1000; qty = 10; entryComm = 0.03% of 1000 = 0.3; exitComm = 0.03% of 1100 = 0.33
			// total comm = 0.63; PnL = 100 - 0.63 = 99.37
			wantQty: 10.0, wantEntry: 100.0, wantExit: 110.0,
			wantComm: 0.63, wantPnL: 99.37, delta: 1e-6,
		},
		{
			name: "zerodha commission — large trade (0.03% > ₹20, capped)",
			cash: 1_000_000, sizeFrac: 0.1,
			orderCfg: model.OrderConfig{CommissionModel: model.CommissionZerodha},
			// cost = 100_000; qty = 1000; 0.03% of 100_000 = 30 > 20 → entryComm = 20
			// 0.03% of 110_000 = 33 > 20 → exitComm = 20; total = 40
			// PnL = (110-100)*1000 - 40 = 9960
			wantQty: 1000.0, wantEntry: 100.0, wantExit: 110.0,
			wantComm: 40.0, wantPnL: 9960.0, delta: 1e-6,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candles := makeTestCandles()
			t0 := candles[0].Timestamp

			cfg := engine.Config{
				Instrument:           "X",
				From:                 t0,
				To:                   t0.AddDate(0, 0, 5),
				InitialCash:          tc.cash,
				OrderConfig:          tc.orderCfg,
				PositionSizeFraction: tc.sizeFrac,
			}
			s := &signalStrategy{signals: []model.Signal{model.SignalBuy, model.SignalSell}}
			prov := &stubProvider{candles: candles}

			e := engine.New(cfg)
			require.NoError(t, e.Run(context.Background(), prov, s))
			port := e.Portfolio()

			closed := closedTrades(port)
			require.Len(t, closed, 1, "expected exactly one closed trade")
			tr := closed[0]

			assert.InDelta(t, tc.wantQty, tr.Quantity, tc.delta, "quantity")
			assert.InDelta(t, tc.wantEntry, tr.EntryPrice, tc.delta, "entry fill price")
			assert.InDelta(t, tc.wantExit, tr.ExitPrice, tc.delta, "exit fill price")
			assert.InDelta(t, tc.wantComm, tr.Commission, tc.delta, "commission")
			assert.InDelta(t, tc.wantPnL, tr.RealizedPnL, tc.delta, "realized PnL")
		})
	}
}

// ── helpers to inspect portfolio state ───────────────────────────────────────

func closedTrades(p *engine.Portfolio) []model.Trade { return p.ClosedTrades() }
func openTrades(p *engine.Portfolio) []model.Trade   { return p.OpenTrades() }
