package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

var wbBase = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

func TestApplySignal_UnknownSignalReturnsError(t *testing.T) {
	p := newPortfolio(10_000, model.OrderConfig{}, 0)
	err := p.applySignal(model.Signal("BOGUS"), "TEST", 100.0, time.Now(), 1.0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown signal")
	assert.Contains(t, err.Error(), "BOGUS")
}

func TestApplySignal_HoldIsNoOp(t *testing.T) {
	p := newPortfolio(10_000, model.OrderConfig{}, 0)
	err := p.applySignal(model.SignalHold, "TEST", 100.0, time.Now(), 1.0)
	require.NoError(t, err)
	assert.Empty(t, p.Trades)
	assert.Empty(t, p.Positions)
	assert.Equal(t, 10_000.0, p.Cash)
}

func TestPortfolio_RecordEquity_CashOnly(t *testing.T) {
	p := newPortfolio(1_000, model.OrderConfig{}, 4)
	candle := model.Candle{Instrument: "X", Timestamp: wbBase, Close: 100}
	p.RecordEquity(candle)

	curve := p.EquityCurve()
	require.Len(t, curve, 1)
	assert.Equal(t, wbBase, curve[0].Timestamp)
	assert.InDelta(t, 1_000.0, curve[0].Value, 1e-6)
}

func TestPortfolio_RecordEquity_WithOpenPosition(t *testing.T) {
	// cash=200, open position qty=8 at close=110 → value = 200 + 8*110 = 1080
	p := newPortfolio(200, model.OrderConfig{}, 4)
	p.Positions["X"] = model.Position{
		Instrument: "X",
		Direction:  model.DirectionLong,
		Quantity:   8,
		EntryPrice: 100,
	}
	candle := model.Candle{Instrument: "X", Timestamp: wbBase, Close: 110}
	p.RecordEquity(candle)

	curve := p.EquityCurve()
	require.Len(t, curve, 1)
	assert.InDelta(t, 1_080.0, curve[0].Value, 1e-6)
}

func TestPortfolio_RecordEquity_MultipleSnapshots(t *testing.T) {
	// No position — equity equals cash at every bar.
	p := newPortfolio(1_000, model.OrderConfig{}, 3)
	times := []time.Time{wbBase, wbBase.Add(24 * time.Hour), wbBase.Add(48 * time.Hour)}

	for i, ts := range times {
		p.RecordEquity(model.Candle{Instrument: "X", Timestamp: ts, Close: float64(100 + i*10)})
	}

	curve := p.EquityCurve()
	require.Len(t, curve, 3)
	for i, pt := range curve {
		assert.Equal(t, times[i], pt.Timestamp, "timestamp mismatch at index %d", i)
		assert.InDelta(t, 1_000.0, pt.Value, 1e-6, "value mismatch at index %d", i)
	}
}

func TestPortfolio_RecordEquity_UnrelatedInstrumentNotCounted(t *testing.T) {
	// Position is in "Y"; candle is for "X" — the position must NOT be counted.
	p := newPortfolio(500, model.OrderConfig{}, 1)
	p.Positions["Y"] = model.Position{Instrument: "Y", Direction: model.DirectionLong, Quantity: 5, EntryPrice: 50}
	candle := model.Candle{Instrument: "X", Timestamp: wbBase, Close: 200}
	p.RecordEquity(candle)

	curve := p.EquityCurve()
	require.Len(t, curve, 1)
	assert.InDelta(t, 500.0, curve[0].Value, 1e-6)
}
