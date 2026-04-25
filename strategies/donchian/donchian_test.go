package donchian_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

func makeOHLC(highs, lows, closes []float64) []model.Candle {
	return testutil.MakeCandlesOHLC(highs, lows, closes)
}

// --- Constructor tests ---

func TestNew_invalidPeriod(t *testing.T) {
	cases := []struct {
		name   string
		period int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := donchian.New(model.TimeframeDaily, tc.period)
			require.Error(t, err, "period=%d should be rejected", tc.period)
		})
	}
}

func TestNew_valid(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 20)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 20)
	require.NoError(t, err)
	require.Equal(t, "donchian-breakout", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := donchian.New(tf, 20)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsPeriodPlusOne(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 20)
	require.NoError(t, err)
	require.Equal(t, 21, s.Lookback())
}

// --- Signal tests ---
//
// Golden sequence used below (period=3, Lookback=4):
//
//   n  Bar  High  Low  Close  window (prior 3)      channelHigh  channelLow  signal
//   ─────────────────────────────────────────────────────────────────────────────────
//   3   0-2  10    5    7     (below Lookback)       —            —           Hold
//   4   3    10    5    7     bars 0-2: H=10, L=5    10.0         5.0         Hold
//   5   4    11    7   11     bars 1-3: H=10, L=5    10.0         5.0         Buy  ← breakout
//   6   5    11    5    7     bars 2-4: H=11, L=5    11.0         5.0         Hold
//   7   6     5    3    3     bars 3-5: H=11, L=5    11.0         5.0         Sell ← break below
//
// Note: at n=7, channelLow = min(Low[3], Low[4], Low[5]) = min(5, 7, 5) = 5; Close=3 < 5 → Sell.

func TestStrategy_Next_holdDuringLookback(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 3)
	require.NoError(t, err)

	// n=3 is below Lookback=4; strategy must return Hold.
	got := s.Next(makeOHLC(
		[]float64{10, 10, 10},
		[]float64{5, 5, 5},
		[]float64{7, 7, 7},
	))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_goldenSequence(t *testing.T) {
	allHighs := []float64{10, 10, 10, 10, 11, 11, 5}
	allLows := []float64{5, 5, 5, 5, 7, 5, 3}
	allCloses := []float64{7, 7, 7, 7, 11, 7, 3}

	s, err := donchian.New(model.TimeframeDaily, 3)
	require.NoError(t, err)

	cases := []struct {
		nCandles int
		want     model.Signal
	}{
		{3, model.SignalHold}, // below Lookback
		{4, model.SignalHold}, // Close=7 within channel [5, 10]
		{5, model.SignalBuy},  // Close=11 > channelHigh=10 → Buy
		{6, model.SignalHold}, // Close=7 within channel [5, 11]
		{7, model.SignalSell}, // Close=3 < channelLow=5 → Sell
	}
	for _, tc := range cases {
		got := s.Next(makeOHLC(
			allHighs[:tc.nCandles],
			allLows[:tc.nCandles],
			allCloses[:tc.nCandles],
		))
		require.Equalf(t, tc.want, got, "Next(%d candles)", tc.nCandles)
	}
}

// TestStrategy_Next_noLookaheadBias confirms that the current bar's own High
// is excluded from the channel window. Bar 3 has High=20 and Close=11. The
// prior three bars all have High=10, so channelHigh=10 and Close=11 > 10 → Buy.
//
// A buggy implementation that includes the current bar's High in the window
// would compute channelHigh=20 and return Hold instead.
func TestStrategy_Next_noLookaheadBias(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 3)
	require.NoError(t, err)

	highs := []float64{10, 10, 10, 20}
	lows := []float64{5, 5, 5, 7}
	closes := []float64{7, 7, 7, 11}

	got := s.Next(makeOHLC(highs, lows, closes))
	require.Equal(t, model.SignalBuy, got,
		"Close=11 must break above prior channel max=10; current bar High=20 must be excluded from window")
}

func TestStrategy_Next_holdWhenWithinChannel(t *testing.T) {
	s, err := donchian.New(model.TimeframeDaily, 3)
	require.NoError(t, err)

	// Channel is [40, 60]; Close=50 is inside → Hold.
	highs := []float64{60, 60, 60, 60}
	lows := []float64{40, 40, 40, 40}
	closes := []float64{50, 50, 50, 50}

	got := s.Next(makeOHLC(highs, lows, closes))
	require.Equal(t, model.SignalHold, got)
}
