package rsimeanrev_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// makeCandles is a local alias so test code stays readable.
func makeCandles(closes []float64) []model.Candle { return testutil.MakeCandles(closes) }

// --- Constructor tests ---

func TestNew_invalidParams(t *testing.T) {
	cases := []struct {
		name                 string
		period               int
		oversold, overbought float64
	}{
		{"period zero", 0, 30, 70},
		{"period negative", -1, 30, 70},
		{"oversold negative", 14, -1, 70},
		{"overbought over 100", 14, 30, 101},
		{"oversold equals overbought", 14, 50, 50},
		{"oversold greater than overbought", 14, 70, 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := rsimeanrev.New(model.TimeframeDaily, tc.period, tc.oversold, tc.overbought)
			require.Error(t, err, "expected error for period=%d oversold=%g overbought=%g",
				tc.period, tc.oversold, tc.overbought)
		})
	}
}

func TestNew_valid(t *testing.T) {
	s, err := rsimeanrev.New(model.TimeframeDaily, 14, 30, 70)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := rsimeanrev.New(model.TimeframeDaily, 14, 30, 70)
	require.NoError(t, err)
	require.Equal(t, "rsi-mean-reversion", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := rsimeanrev.New(tf, 14, 30, 70)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsPeriodPlusOne(t *testing.T) {
	s, err := rsimeanrev.New(model.TimeframeDaily, 14, 30, 70)
	require.NoError(t, err)
	require.Equal(t, 15, s.Lookback())
}

// --- Signal tests ---
//
// All signal tests use period=3 (Lookback=4) so sequences are short enough
// to verify RSI by hand.
//
//   oversold:  closes = [100, 99, 98, 97]
//     changes: -1, -1, -1 → avg_gain=0, avg_loss=1 → RSI=0 < 30 → Buy
//
//   overbought: closes = [100, 101, 102, 103]
//     changes: +1, +1, +1 → avg_gain=1, avg_loss=0 → RSI=100 > 70 → Sell
//
//   neutral:   closes = [100, 101, 100, 101]
//     changes: +1, -1, +1 → avg_gain=2/3, avg_loss=1/3 → RS=2 → RSI≈66.7 → Hold

func TestStrategy_Next_holdDuringLookback(t *testing.T) {
	s, err := rsimeanrev.New(model.TimeframeDaily, 3, 30, 70)
	require.NoError(t, err)
	// Exactly period (3) candles: Lookback=4, guard must return Hold.
	closes := []float64{100, 99, 98}
	got := s.Next(makeCandles(closes))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_oversoldBuy(t *testing.T) {
	// closes = [100,99,98,97]: 3 consecutive losses, no gains → RSI=0 < 30 → Buy
	closes := []float64{100, 99, 98, 97}
	s, err := rsimeanrev.New(model.TimeframeDaily, 3, 30, 70)
	require.NoError(t, err)

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{3, model.SignalHold}, // below Lookback → guard
		{4, model.SignalBuy},  // RSI=0 < 30 → Buy
	}
	for _, tc := range cases {
		got := s.Next(makeCandles(closes[:tc.nCandles]))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}

func TestStrategy_Next_overboughtSell(t *testing.T) {
	// closes = [100,101,102,103]: 3 consecutive gains, no losses → RSI=100 > 70 → Sell
	closes := []float64{100, 101, 102, 103}
	s, err := rsimeanrev.New(model.TimeframeDaily, 3, 30, 70)
	require.NoError(t, err)

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{3, model.SignalHold}, // below Lookback → guard
		{4, model.SignalSell}, // RSI=100 > 70 → Sell
	}
	for _, tc := range cases {
		got := s.Next(makeCandles(closes[:tc.nCandles]))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}

func TestStrategy_Next_neutralHold(t *testing.T) {
	// closes = [100,101,100,101]: alternating ±1
	// changes: +1,-1,+1 → avg_gain=2/3, avg_loss=1/3 → RS=2 → RSI≈66.7 → Hold
	closes := []float64{100, 101, 100, 101}
	s, err := rsimeanrev.New(model.TimeframeDaily, 3, 30, 70)
	require.NoError(t, err)
	got := s.Next(makeCandles(closes))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_continuousSignal(t *testing.T) {
	// Signal fires on every bar where RSI is in the zone, not just on entry.
	// [100,99,98,97,96]: all declining → RSI stays at 0 on every bar → Buy every bar.
	s, err := rsimeanrev.New(model.TimeframeDaily, 3, 30, 70)
	require.NoError(t, err)
	closes := []float64{100, 99, 98, 97, 96}

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{4, model.SignalBuy}, // first valid bar → Buy
		{5, model.SignalBuy}, // still declining → RSI still 0 → Buy again
	}
	for _, tc := range cases {
		got := s.Next(makeCandles(closes[:tc.nCandles]))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}
