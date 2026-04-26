package macd_test

import (
	"testing"

	talib "github.com/markcheno/go-talib"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// --- Constructor tests ---

func TestNew_invalidPeriods(t *testing.T) {
	cases := []struct {
		name               string
		fast, slow, signal int
	}{
		{"fast zero", 0, 26, 9},
		{"fast negative", -1, 26, 9},
		{"slow zero", 12, 0, 9},
		{"slow negative", 12, -1, 9},
		{"signal zero", 12, 26, 0},
		{"signal negative", 12, 26, -1},
		{"fast equals slow", 26, 26, 9},
		{"fast greater than slow", 30, 26, 9},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := macd.New(model.TimeframeDaily, tc.fast, tc.slow, tc.signal)
			require.Error(t, err, "params fast=%d slow=%d signal=%d should be rejected", tc.fast, tc.slow, tc.signal)
		})
	}
}

func TestNew_valid(t *testing.T) {
	s, err := macd.New(model.TimeframeDaily, 12, 26, 9)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := macd.New(model.TimeframeDaily, 12, 26, 9)
	require.NoError(t, err)
	require.Equal(t, "macd-crossover", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := macd.New(tf, 12, 26, 9)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsSlowPlusSignalMinusOne(t *testing.T) {
	cases := []struct {
		fast, slow, signal int
		want               int
	}{
		{12, 26, 9, 34}, // 26 + 9 - 1
		{5, 20, 5, 24},  // 20 + 5 - 1
		{3, 10, 4, 13},  // 10 + 4 - 1
	}
	for _, tc := range cases {
		s, err := macd.New(model.TimeframeDaily, tc.fast, tc.slow, tc.signal)
		require.NoError(t, err)
		require.Equal(t, tc.want, s.Lookback(), "fast=%d slow=%d signal=%d", tc.fast, tc.slow, tc.signal)
	}
}

// --- Signal tests ---
//
// MACD crossover requires two consecutive valid signal-line values.
// talib fills uninitialized positions with zeros. The first real (non-zero-fill)
// signal-line value appears at index slow+signal-2 (0-based), so:
//
//   n = slow+signal-1 (= Lookback())  → only index n-1 is valid → Hold
//   n = slow+signal   (= Lookback()+1) → indices n-2 and n-1 are valid → first possible signal
//
// goldenCrossBar / goldenDeadCrossBar search for crossovers starting at index
// slow+signal-1, ensuring both prev and curr values are in the valid region.
// Starting earlier would find the spurious zero→nonzero transition at initialization.

// reversalSeries returns a price series that declines for the first half then
// rises for the second half, producing a genuine MACD crossover well past the
// initialization window.
func reversalSeries(n int) []float64 {
	mid := n / 2
	closes := make([]float64, n)
	for i := 0; i < mid; i++ {
		closes[i] = 200 - float64(i)
	}
	bottom := closes[mid-1]
	for i := mid; i < n; i++ {
		closes[i] = bottom + float64(i-mid+1)
	}
	return closes
}

func goldenCrossBar(closes []float64, fast, slow, signal int) int {
	macdLine, sigLine, _ := talib.Macd(closes, fast, slow, signal)
	// Start at slow+signal-1 so that macdLine[i-1] and sigLine[i-1] are valid
	// (not zero-filled). Starting from index 1 would catch the spurious
	// zero→nonzero initialization transition.
	start := slow + signal - 1
	for i := start; i < len(closes); i++ {
		if macdLine[i-1] <= sigLine[i-1] && macdLine[i] > sigLine[i] {
			return i
		}
	}
	return -1
}

func goldenDeadCrossBar(closes []float64, fast, slow, signal int) int {
	macdLine, sigLine, _ := talib.Macd(closes, fast, slow, signal)
	start := slow + signal - 1
	for i := start; i < len(closes); i++ {
		if macdLine[i-1] >= sigLine[i-1] && macdLine[i] < sigLine[i] {
			return i
		}
	}
	return -1
}

func TestStrategy_Next_holdAtLookback(t *testing.T) {
	// With exactly Lookback() candles, only one valid signal-line value exists.
	// No previous bar to compare against → Hold.
	s, err := macd.New(model.TimeframeDaily, 12, 26, 9)
	require.NoError(t, err)

	closes := make([]float64, s.Lookback())
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "exactly Lookback() bars must return Hold")
}

func TestStrategy_Next_holdOnFlatPrices(t *testing.T) {
	// Flat prices → MACD = 0 and signal = 0 throughout → no crossover.
	s, err := macd.New(model.TimeframeDaily, 12, 26, 9)
	require.NoError(t, err)

	closes := make([]float64, 60)
	for i := range closes {
		closes[i] = 100
	}
	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_buyAtGoldenCross(t *testing.T) {
	// Decline-then-rise series produces a genuine MACD golden cross in the valid region.
	// Verify: Hold at the bar before the crossover, Buy at the crossover bar.
	fast, slow, signal := 12, 26, 9
	closes := reversalSeries(200)

	buyBar := goldenCrossBar(closes, fast, slow, signal)
	require.True(t, buyBar > 0, "reversal series must produce a MACD golden cross in the valid region")

	s, err := macd.New(model.TimeframeDaily, fast, slow, signal)
	require.NoError(t, err)

	// One bar before the crossover: Hold.
	got := s.Next(testutil.MakeCandles(closes[:buyBar]))
	require.Equal(t, model.SignalHold, got, "bar before golden cross must be Hold")

	// At the crossover bar: Buy.
	got = s.Next(testutil.MakeCandles(closes[:buyBar+1]))
	require.Equal(t, model.SignalBuy, got, "golden cross bar must emit Buy")
}

func TestStrategy_Next_sellAtDeadCross(t *testing.T) {
	// Rise-then-fall series produces a genuine MACD dead cross in the valid region.
	// Verify: Sell at the dead cross bar.
	fast, slow, signal := 12, 26, 9
	// Invert reversalSeries: rise then decline.
	raw := reversalSeries(200)
	closes := make([]float64, len(raw))
	for i, v := range raw {
		closes[i] = 400 - v
	}

	sellBar := goldenDeadCrossBar(closes, fast, slow, signal)
	require.True(t, sellBar > 0, "rise-then-fall series must produce a MACD dead cross in the valid region")

	s, err := macd.New(model.TimeframeDaily, fast, slow, signal)
	require.NoError(t, err)

	got := s.Next(testutil.MakeCandles(closes[:sellBar+1]))
	require.Equal(t, model.SignalSell, got, "dead cross bar must emit Sell")
}

func TestStrategy_Next_strictCrossover_noBuyOnContinuation(t *testing.T) {
	// After a golden cross, if MACD stays above signal without a new crossover,
	// the strategy must return Hold (not a repeated Buy).
	fast, slow, signal := 12, 26, 9
	closes := reversalSeries(200)

	buyBar := goldenCrossBar(closes, fast, slow, signal)
	require.True(t, buyBar > 0)

	macdLine, sigLine, _ := talib.Macd(closes, fast, slow, signal)
	afterBar := buyBar + 1
	if afterBar >= len(closes) {
		t.Skip("no bar after crossover in series")
	}
	// Only test pure continuation: both prev and curr have MACD > signal.
	if macdLine[afterBar-1] <= sigLine[afterBar-1] {
		t.Skip("consecutive crossover — not a pure continuation bar")
	}
	if macdLine[afterBar] <= sigLine[afterBar] {
		t.Skip("MACD dropped at or below signal — not a continuation bar")
	}

	s, err := macd.New(model.TimeframeDaily, fast, slow, signal)
	require.NoError(t, err)

	got := s.Next(testutil.MakeCandles(closes[:afterBar+1]))
	require.Equal(t, model.SignalHold, got, "continuation bar (MACD above signal, no new crossover) must be Hold")
}

func TestStrategy_Next_strictCrossover_noSellOnContinuation(t *testing.T) {
	// After a dead cross, if MACD stays below signal without a new crossover, Hold.
	fast, slow, signal := 12, 26, 9
	raw := reversalSeries(200)
	closes := make([]float64, len(raw))
	for i, v := range raw {
		closes[i] = 400 - v
	}

	sellBar := goldenDeadCrossBar(closes, fast, slow, signal)
	require.True(t, sellBar > 0)

	macdLine, sigLine, _ := talib.Macd(closes, fast, slow, signal)
	afterBar := sellBar + 1
	if afterBar >= len(closes) {
		t.Skip("no bar after crossover in series")
	}
	if macdLine[afterBar-1] >= sigLine[afterBar-1] {
		t.Skip("consecutive crossover — skip")
	}
	if macdLine[afterBar] >= sigLine[afterBar] {
		t.Skip("MACD rose at or above signal — not a continuation bar")
	}

	s, err := macd.New(model.TimeframeDaily, fast, slow, signal)
	require.NoError(t, err)

	got := s.Next(testutil.MakeCandles(closes[:afterBar+1]))
	require.Equal(t, model.SignalHold, got, "continuation bar (MACD below signal, no new crossover) must be Hold")
}
