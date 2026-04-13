package engine

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ── computeInstrumentVol ─────────────────────────────────────────────────────

func TestComputeInstrumentVol_NilSlice(t *testing.T) {
	assert.Equal(t, 0.0, computeInstrumentVol(nil))
}

func TestComputeInstrumentVol_OneCandle(t *testing.T) {
	c := []model.Candle{{Close: 100}}
	assert.Equal(t, 0.0, computeInstrumentVol(c))
}

func TestComputeInstrumentVol_TwoCandles_ReturnsZero(t *testing.T) {
	// Two candles produce only one return — sample std dev requires at least 2 data points.
	c := []model.Candle{{Close: 100}, {Close: 110}}
	assert.Equal(t, 0.0, computeInstrumentVol(c))
}

func TestComputeInstrumentVol_ThreeCandles_KnownVol(t *testing.T) {
	// closes: [100, 110, 100]
	// returns: [ln(110/100), ln(100/110)] = [+r, -r] where r = ln(1.1)
	// mean = 0; sample variance = 2*r² / 1 = 2*r²; vol = r*sqrt(2)
	r := math.Log(1.1)
	wantVol := r * math.Sqrt(2)

	c := makeSizingCandles([]float64{100, 110, 100})
	got := computeInstrumentVol(c)
	assert.InDelta(t, wantVol, got, 1e-9)
}

func TestComputeInstrumentVol_WindowCappedAt20Bars(t *testing.T) {
	// Build 25 candles. The first 5 have a wildly different close (price=1000) relative to the
	// rest (alternating 100/110). The 20-bar window must exclude the first 5 candles.
	//
	// If the window were not capped, return[5] = ln(100/1000) = -ln(10) would skew the vol.
	// With the correct 20-bar window, only bars 5..24 are used → 19 returns, none touching bar4.

	closes := make([]float64, 25)
	for i := range closes {
		if i < 5 {
			closes[i] = 1000 // outlier range — should be excluded
		} else if i%2 == 0 {
			closes[i] = 100
		} else {
			closes[i] = 110
		}
	}

	c := makeSizingCandles(closes)

	// Build a comparison set of exactly the last 20 candles (bars 5..24).
	last20 := c[5:]
	volFull := computeInstrumentVol(c)
	volLast20 := computeInstrumentVol(last20)

	assert.InDelta(t, volLast20, volFull, 1e-9, "vol must equal the last-20-bar vol, not all 25")
}

func TestComputeInstrumentVol_ZeroCloseIgnored(t *testing.T) {
	// A candle with Close=0 is skipped. Placing the zero-close bar first means only
	// return[i=1] is dropped; returns[i=2] and [i=3] are ln(110/100) and ln(100/110)
	// — identical to the 3-candle known case, so we can reuse the same expected vol.
	r := math.Log(1.1)
	wantVol := r * math.Sqrt(2)
	// makeSizingCandles sets all OHLC fields to the given value; Close=0 is deliberate.
	c := makeSizingCandles([]float64{0, 100, 110, 100})
	got := computeInstrumentVol(c)
	assert.InDelta(t, wantVol, got, 1e-9)
}

func TestComputeInstrumentVol_ConstantPrices_ReturnsZero(t *testing.T) {
	// All returns are 0 → mean=0, variance=0, vol=0.
	closes := make([]float64, 10)
	for i := range closes {
		closes[i] = 100
	}
	assert.Equal(t, 0.0, computeInstrumentVol(makeSizingCandles(closes)))
}

// ── sizeFractionForBar ───────────────────────────────────────────────────────

func TestSizeFractionForBar_FixedModel(t *testing.T) {
	cfg := Config{PositionSizeFraction: 0.25, SizingModel: model.SizingFixed}
	// candles are irrelevant for fixed sizing
	got := sizeFractionForBar(cfg, nil)
	assert.InDelta(t, 0.25, got, 1e-9)
}

func TestSizeFractionForBar_VolTarget_ZeroVol_ReturnsZero(t *testing.T) {
	cfg := Config{SizingModel: model.SizingVolatilityTarget, VolatilityTarget: 0.10}
	// constant-price candles → vol=0 → fraction must be 0 (skip buy)
	closes := make([]float64, 10)
	for i := range closes {
		closes[i] = 100
	}
	got := sizeFractionForBar(cfg, makeSizingCandles(closes))
	assert.Equal(t, 0.0, got)
}

func TestSizeFractionForBar_VolTarget_KnownVol(t *testing.T) {
	// closes: [100, 110, 100]  → vol = ln(1.1)*sqrt(2)  (see TestComputeInstrumentVol_ThreeCandles)
	// volTarget = 0.10
	// expected fraction = 0.10 / (vol * sqrt(252))
	r := math.Log(1.1)
	vol := r * math.Sqrt(2)
	wantFraction := 0.10 / (vol * math.Sqrt(252))

	cfg := Config{SizingModel: model.SizingVolatilityTarget, VolatilityTarget: 0.10}
	candles := makeSizingCandles([]float64{100, 110, 100})
	got := sizeFractionForBar(cfg, candles)
	assert.InDelta(t, wantFraction, got, 1e-9)
}

func TestSizeFractionForBar_VolTarget_CapsAtOne(t *testing.T) {
	// Very low vol → notional would exceed available cash → fraction capped at 1.0.
	// Use a tiny volTarget with moderate vol so fraction < 1 normally, then invert: use a
	// very high volTarget with the same moderate vol to force fraction > 1.
	cfg := Config{SizingModel: model.SizingVolatilityTarget, VolatilityTarget: 100.0} // 100x target → always > 1
	candles := makeSizingCandles([]float64{100, 110, 100})
	got := sizeFractionForBar(cfg, candles)
	assert.InDelta(t, 1.0, got, 1e-9, "fraction must be capped at 1.0")
}

func TestSizeFractionForBar_VolTarget_InsufficientHistory_ReturnsZero(t *testing.T) {
	cfg := Config{SizingModel: model.SizingVolatilityTarget, VolatilityTarget: 0.10}
	// Only 2 candles → computeInstrumentVol returns 0 → fraction 0.
	got := sizeFractionForBar(cfg, makeSizingCandles([]float64{100, 110}))
	assert.Equal(t, 0.0, got)
}

// ── helper ───────────────────────────────────────────────────────────────────

// makeSizingCandles builds candles with only Close set (sufficient for vol computation).
func makeSizingCandles(closes []float64) []model.Candle {
	t0 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	c := make([]model.Candle, len(closes))
	for i, cl := range closes {
		c[i] = model.Candle{
			Instrument: "TEST",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  t0.AddDate(0, 0, i),
			Open:       cl,
			High:       cl,
			Low:        cl,
			Close:      cl,
			Volume:     1000,
		}
	}
	return c
}
