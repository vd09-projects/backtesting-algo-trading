package momentum_test

import (
	"testing"

	talib "github.com/markcheno/go-talib"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/momentum"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// --- Constructor tests ---

func TestNew_invalidParams(t *testing.T) {
	cases := []struct {
		name      string
		lookback  int
		threshold float64
	}{
		{"lookback zero", 0, 10.0},
		{"lookback negative", -1, 10.0},
		{"threshold zero", 231, 0.0},
		{"threshold negative", 231, -5.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := momentum.New(model.TimeframeDaily, tc.lookback, tc.threshold)
			require.Error(t, err, "lookback=%d threshold=%f should be rejected", tc.lookback, tc.threshold)
		})
	}
}

func TestNew_valid(t *testing.T) {
	s, err := momentum.New(model.TimeframeDaily, 231, 10.0)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := momentum.New(model.TimeframeDaily, 231, 10.0)
	require.NoError(t, err)
	require.Equal(t, "momentum", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := momentum.New(tf, 231, 10.0)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsLookbackPlusOne(t *testing.T) {
	cases := []struct {
		lookback int
		want     int
	}{
		{231, 232},
		{10, 11},
		{50, 51},
	}
	for _, tc := range cases {
		s, err := momentum.New(model.TimeframeDaily, tc.lookback, 10.0)
		require.NoError(t, err)
		require.Equal(t, tc.want, s.Lookback(), "lookback=%d", tc.lookback)
	}
}

// --- Signal tests ---
//
// ROC is defined as: (close[n-1] / close[n-1-lookback] - 1) * 100
// For a series of lookback+1 bars:
//   - If final bar's close is 110% of the bar lookback steps ago: ROC = 10.0
//   - If final bar's close is 80% of the bar lookback steps ago:  ROC = -20.0
//
// talib.Roc returns a slice of the same length as input; positions [0, lookback-1]
// are zero-filled (uninitialized). With Lookback() = lookback+1, the engine
// passes n = lookback+1 bars minimum, so roc[n-1] = roc[lookback] is always valid.

// rocValue computes the expected ROC for a flat series where the last bar
// differs. Uses the same talib.Roc call for consistency.
func rocValue(closes []float64, lookback int) float64 {
	roc := talib.Roc(closes, lookback)
	return roc[len(roc)-1]
}

// makeROCSeries builds a closes series of length lookback+1 where the first
// lookback bars are at basePrice and the last bar is at finalPrice.
// ROC = (finalPrice/basePrice - 1) * 100.
func makeROCSeries(lookback int, basePrice, finalPrice float64) []float64 {
	closes := make([]float64, lookback+1)
	for i := range lookback {
		closes[i] = basePrice
	}
	closes[lookback] = finalPrice
	return closes
}

func TestStrategy_Next_holdAtExactLookback(t *testing.T) {
	// With exactly Lookback() = lookback+1 bars all at same price, ROC = 0.
	// threshold=10.0, so ROC=0 is within [-10, 10] → Hold.
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	closes := makeROCSeries(lookback, 100.0, 100.0) // ROC = 0
	require.Equal(t, lookback+1, len(closes), "series length must equal Lookback()")

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "ROC=0 with threshold=10 must return Hold")
}

func TestStrategy_Next_buyWhenROCAboveThreshold(t *testing.T) {
	// basePrice=100, finalPrice=120 → ROC = 20.0 > threshold=10.0 → Buy.
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	closes := makeROCSeries(lookback, 100.0, 120.0)
	roc := rocValue(closes, lookback)
	require.Greater(t, roc, 10.0, "test series must produce ROC > threshold")

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalBuy, got, "ROC=%.2f with threshold=10 must return Buy", roc)
}

func TestStrategy_Next_sellWhenROCBelowNegativeThreshold(t *testing.T) {
	// basePrice=100, finalPrice=75 → ROC = -25.0 < -threshold=-10.0 → Sell.
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	closes := makeROCSeries(lookback, 100.0, 75.0)
	roc := rocValue(closes, lookback)
	require.Less(t, roc, -10.0, "test series must produce ROC < -threshold")

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalSell, got, "ROC=%.2f with threshold=10 must return Sell", roc)
}

func TestStrategy_Next_holdWhenROCWithinThreshold(t *testing.T) {
	// basePrice=100, finalPrice=105 → ROC = 5.0; -10 < 5 < 10 → Hold.
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	closes := makeROCSeries(lookback, 100.0, 105.0)
	roc := rocValue(closes, lookback)
	require.Greater(t, roc, -10.0)
	require.Less(t, roc, 10.0)

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "ROC=%.2f within [-10, 10] must return Hold", roc)
}

func TestStrategy_Next_holdJustBelowPositiveThreshold(t *testing.T) {
	// ROC just below threshold must NOT trigger Buy (strictly greater required).
	// closes = [100, 109.99] → ROC ≈ 9.99 < 10.0 → Hold.
	// talib.Roc floating-point output is verified to be below 10.0 before asserting.
	s, err := momentum.New(model.TimeframeDaily, 1, 10.0)
	require.NoError(t, err)

	closes := []float64{100.0, 109.99}
	roc := rocValue(closes, 1)
	require.Less(t, roc, 10.0, "series must produce ROC strictly less than threshold")

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "ROC=%.6f just below threshold=10 must return Hold", roc)
}

func TestStrategy_Next_holdJustAboveNegativeThreshold(t *testing.T) {
	// ROC just above -threshold must NOT trigger Sell (strictly less required).
	// closes = [100, 90.01] → ROC ≈ -9.99 > -10.0 → Hold.
	// talib.Roc floating-point output is verified to be above -10.0 before asserting.
	s, err := momentum.New(model.TimeframeDaily, 1, 10.0)
	require.NoError(t, err)

	closes := []float64{100.0, 90.01}
	roc := rocValue(closes, 1)
	require.Greater(t, roc, -10.0, "series must produce ROC strictly greater than -threshold")

	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "ROC=%.6f just above -threshold=-10 must return Hold", roc)
}

func TestStrategy_Next_buyPersistsAboveThreshold(t *testing.T) {
	// Under level-comparison semantics, every bar where ROC > threshold → Buy.
	// This is by design (engine's no-pyramiding handles the open-position case).
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	// Extend the series — all bars after the initial high stay high.
	// closes[0..lookback-1] = 100, closes[lookback] = 120, closes[lookback+1] = 122
	closes := make([]float64, lookback+2)
	for i := range lookback {
		closes[i] = 100.0
	}
	closes[lookback] = 120.0
	closes[lookback+1] = 122.0

	// Both bars [lookback] and [lookback+1] should emit Buy.
	got1 := s.Next(testutil.MakeCandles(closes[:lookback+1]))
	require.Equal(t, model.SignalBuy, got1, "first bar above threshold must emit Buy")

	got2 := s.Next(testutil.MakeCandles(closes[:lookback+2]))
	require.Equal(t, model.SignalBuy, got2, "subsequent bar still above threshold must emit Buy (level-comparison)")
}

func TestStrategy_Next_fewerThanLookbackBarsReturnsHold(t *testing.T) {
	// If Next is called with fewer bars than Lookback(), it must return Hold.
	// The engine guarantees this won't happen, but defensive guard must exist.
	lookback := 5
	s, err := momentum.New(model.TimeframeDaily, lookback, 10.0)
	require.NoError(t, err)

	// Pass lookback bars (one short of Lookback() = lookback+1).
	closes := makeROCSeries(lookback-1, 100.0, 120.0)
	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "fewer than Lookback() bars must return Hold")
}
