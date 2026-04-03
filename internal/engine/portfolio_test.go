package engine_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// makeSignalStrategy returns a strategy that emits signals from a fixed sequence,
// then Hold for any bar beyond the slice.
type signalStrategy struct {
	signals []model.Signal
	i       int
}

func (s *signalStrategy) Name() string                  { return "signal-seq" }
func (s *signalStrategy) Timeframe() model.Timeframe    { return model.TimeframeDaily }
func (s *signalStrategy) Lookback() int                 { return 1 }
func (s *signalStrategy) Next(_ []model.Candle) model.Signal {
	if s.i >= len(s.signals) {
		return model.SignalHold
	}
	sig := s.signals[s.i]
	s.i++
	return sig
}

// ── helpers ───────────────────────────────────────────────────────────────────

func runWithSignals(t *testing.T, initialCash float64, sizeFraction float64, signals []model.Signal) *engine.Portfolio {
	t.Helper()
	n := len(signals)
	candles := makeCandles(n) // reuse helper from engine_test.go

	cfg := engine.EngineConfig{
		Instrument:           "TEST",
		From:                 base,
		To:                   base.AddDate(0, 0, n+1),
		InitialCash:          initialCash,
		PositionSizeFraction: sizeFraction,
	}
	strat := &signalStrategy{signals: signals}
	p := &stubProvider{candles: candles}

	e := engine.New(cfg)
	require.NoError(t, e.Run(p, strat))
	return e.Portfolio()
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestPortfolio_BuyOpensPosition(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{model.SignalBuy})

	assert.Len(t, port.Positions, 1)
	pos, ok := port.Positions["TEST"]
	require.True(t, ok)
	assert.Equal(t, model.DirectionLong, pos.Direction)
	assert.Greater(t, pos.Quantity, 0.0)
	// All cash deployed (sizeFraction=1.0)
	assert.InDelta(t, 0.0, port.Cash, 1e-6)
}

func TestPortfolio_SellClosesPosition(t *testing.T) {
	// Buy bar 0 @ close=102, sell bar 1 @ close=103
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{model.SignalBuy, model.SignalSell})

	assert.Empty(t, port.Positions, "position should be closed after sell")
	closed := closedTrades(port)
	require.Len(t, closed, 1)

	tr := closed[0]
	assert.Equal(t, "TEST", tr.Instrument)
	assert.Equal(t, model.DirectionLong, tr.Direction)
	assert.False(t, tr.ExitTime.IsZero())
	// Bought at close of bar 0 (100+0=100 open → close=102), sold at close of bar 1 (close=103)
	assert.Greater(t, tr.RealizedPnL, 0.0, "profitable trade expected")
}

func TestPortfolio_PnLCalculation(t *testing.T) {
	// Use price-controlled candles: buy @ 100, sell @ 110 → PnL = 10 * qty
	candles := []model.Candle{
		{Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: base,
			Open: 100, High: 105, Low: 99, Close: 100, Volume: 1000},
		{Instrument: "X", Timeframe: model.TimeframeDaily, Timestamp: base.Add(24 * time.Hour),
			Open: 110, High: 115, Low: 109, Close: 110, Volume: 1000},
	}

	cfg := engine.EngineConfig{
		Instrument:           "X",
		From:                 base,
		To:                   base.AddDate(0, 0, 5),
		InitialCash:          1000,
		PositionSizeFraction: 1.0,
	}
	strat := &signalStrategy{signals: []model.Signal{model.SignalBuy, model.SignalSell}}
	prov := &stubProvider{candles: candles}

	e := engine.New(cfg)
	require.NoError(t, e.Run(prov, strat))
	port := e.Portfolio()

	closed := closedTrades(port)
	require.Len(t, closed, 1)

	tr := closed[0]
	// qty = 1000 / 100 = 10; PnL = (110-100) * 10 = 100
	assert.InDelta(t, 100.0, tr.RealizedPnL, 1e-6)
	assert.InDelta(t, 10.0, tr.Quantity, 1e-6)
}

func TestPortfolio_CashInsufficientNoBuy(t *testing.T) {
	// sizeFraction=0 means cost=0, buy should be skipped
	port := runWithSignals(t, 10_000, 0.0, []model.Signal{model.SignalBuy})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
}

func TestPortfolio_SellWithNoPositionIsNoop(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{model.SignalSell})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
	assert.Empty(t, closedTrades(port))
}

func TestPortfolio_DoubleBuyDoesNotPyramid(t *testing.T) {
	// Two consecutive buys should only open one position
	port := runWithSignals(t, 10_000, 0.5, []model.Signal{model.SignalBuy, model.SignalBuy})

	assert.Len(t, port.Positions, 1)
	// Only one open trade sentinel
	open := openTrades(port)
	assert.Len(t, open, 1)
}

func TestPortfolio_BuySellBuyOpensSecondPosition(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{
		model.SignalBuy, model.SignalSell, model.SignalBuy,
	})

	assert.Len(t, port.Positions, 1)
	assert.Len(t, closedTrades(port), 1)
}

func TestPortfolio_HoldIsNoop(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{
		model.SignalHold, model.SignalHold,
	})

	assert.Empty(t, port.Positions)
	assert.InDelta(t, 10_000.0, port.Cash, 1e-6)
}

func TestPortfolio_TradeLogTimestamps(t *testing.T) {
	port := runWithSignals(t, 10_000, 1.0, []model.Signal{model.SignalBuy, model.SignalSell})

	closed := closedTrades(port)
	require.Len(t, closed, 1)
	tr := closed[0]

	assert.True(t, tr.EntryTime.Before(tr.ExitTime), "entry must precede exit")
}

// ── helpers to inspect unexported portfolio internals via public Trades field ──

func closedTrades(p *engine.Portfolio) []model.Trade {
	var out []model.Trade
	for _, tr := range p.Trades {
		if !tr.ExitTime.IsZero() {
			out = append(out, tr)
		}
	}
	return out
}

func openTrades(p *engine.Portfolio) []model.Trade {
	var out []model.Trade
	for _, tr := range p.Trades {
		if tr.ExitTime.IsZero() {
			out = append(out, tr)
		}
	}
	return out
}
