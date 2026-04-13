package smacrossover_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// makeCandles is a local alias so test code stays readable.
func makeCandles(closes []float64) []model.Candle { return testutil.MakeCandles(closes) }

// --- Constructor tests ---

func TestNew_invalidPeriods(t *testing.T) {
	cases := []struct {
		name       string
		fast, slow int
	}{
		{"fast zero", 0, 5},
		{"slow zero", 3, 0},
		{"fast equals slow", 5, 5},
		{"fast greater than slow", 10, 5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := smacrossover.New(model.TimeframeDaily, tc.fast, tc.slow)
			require.Error(t, err, "expected error for fast=%d slow=%d", tc.fast, tc.slow)
		})
	}
}

func TestNew_validPeriods(t *testing.T) {
	s, err := smacrossover.New(model.TimeframeDaily, 1, 2)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := smacrossover.New(model.TimeframeDaily, 10, 50)
	require.NoError(t, err)
	require.Equal(t, "sma-crossover", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := smacrossover.New(tf, 10, 50)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsSlowPeriod(t *testing.T) {
	s, err := smacrossover.New(model.TimeframeDaily, 10, 50)
	require.NoError(t, err)
	require.Equal(t, 50, s.Lookback())
}

// --- Signal tests ---
//
// All signal tests use fast=3, slow=5. The sequences below have SMA values
// that can be verified by hand:
//
//   bullish: closes = [10×7, 20×3]
//     n=8 → SMA3 curr=13.33 > SMA5 curr=12;  SMA3 prev=10 == SMA5 prev=10  → Buy
//     n=9 → SMA3 curr=16.67 > SMA5 curr=14;  SMA3 prev=13.33 > SMA5 prev=12 → Hold
//
//   bearish: closes = [20×7, 10×3]
//     n=8 → SMA3 curr=16.67 < SMA5 curr=18;  SMA3 prev=20 == SMA5 prev=20  → Sell
//     n=9 → Hold (already below)

func TestStrategy_Next_holdDuringLookback(t *testing.T) {
	s, err := smacrossover.New(model.TimeframeDaily, 3, 5)
	require.NoError(t, err)
	// Exactly slowPeriod (5) candles: guard must return Hold.
	closes := []float64{10, 10, 10, 10, 10}
	got := s.Next(makeCandles(closes))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_bullishCrossover(t *testing.T) {
	// closes = [10,10,10,10,10,10,10,20,20,20]
	closes := append(make([]float64, 0, 10),
		10, 10, 10, 10, 10, 10, 10, 20, 20, 20)

	s, err := smacrossover.New(model.TimeframeDaily, 3, 5)
	require.NoError(t, err)

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{5, model.SignalHold},  // exactly slowPeriod — guard
		{6, model.SignalHold},  // fast==slow on both bars
		{7, model.SignalHold},  // fast==slow on both bars
		{8, model.SignalBuy},   // crossover: fast(13.33) > slow(12), prev fast==slow
		{9, model.SignalHold},  // fast above slow but no new crossover
		{10, model.SignalHold}, // same
	}
	for _, tc := range cases {
		got := s.Next(makeCandles(closes[:tc.nCandles]))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}

func TestStrategy_Next_bearishCrossover(t *testing.T) {
	// closes = [20,20,20,20,20,20,20,10,10,10]
	closes := append(make([]float64, 0, 10),
		20, 20, 20, 20, 20, 20, 20, 10, 10, 10)

	s, err := smacrossover.New(model.TimeframeDaily, 3, 5)
	require.NoError(t, err)

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{5, model.SignalHold},  // guard
		{6, model.SignalHold},  // fast==slow on both bars
		{7, model.SignalHold},  // same
		{8, model.SignalSell},  // crossover: fast(16.67) < slow(18), prev fast==slow
		{9, model.SignalHold},  // fast below slow but no new crossover
		{10, model.SignalHold}, // same
	}
	for _, tc := range cases {
		got := s.Next(makeCandles(closes[:tc.nCandles]))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}

func TestStrategy_Next_noSignalWhenAlreadyCrossed(t *testing.T) {
	// Start fast already above slow for all bars — no crossover ever fires.
	// Prices are monotonically increasing so SMA3 > SMA5 from the start.
	closes := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	s, err := smacrossover.New(model.TimeframeDaily, 3, 5)
	require.NoError(t, err)

	// n=6: both prev and curr have fast > slow — no crossover → Hold
	got := s.Next(makeCandles(closes[:6]))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_equalSMAs_noSignal(t *testing.T) {
	// All closes identical: SMAs are always equal, no crossover ever.
	closes := make([]float64, 10)
	for i := range closes {
		closes[i] = 100
	}
	s, err := smacrossover.New(model.TimeframeDaily, 3, 5)
	require.NoError(t, err)

	for n := 6; n <= 10; n++ {
		got := s.Next(makeCandles(closes[:n]))
		require.Equalf(t, model.SignalHold, got, "Next(%d candles)", n)
	}
}
