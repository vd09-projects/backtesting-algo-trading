package bollinger_test

import (
	"testing"

	talib "github.com/markcheno/go-talib"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// --- Constructor tests ---

func TestNew_invalidParams(t *testing.T) {
	cases := []struct {
		name      string
		period    int
		numStdDev float64
	}{
		{"period zero", 0, 2.0},
		{"period negative", -1, 2.0},
		{"numStdDev zero", 20, 0.0},
		{"numStdDev negative", 20, -1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := bollinger.New(model.TimeframeDaily, tc.period, tc.numStdDev)
			require.Error(t, err, "period=%d numStdDev=%f should be rejected", tc.period, tc.numStdDev)
		})
	}
}

func TestNew_valid(t *testing.T) {
	s, err := bollinger.New(model.TimeframeDaily, 20, 2.0)
	require.NoError(t, err)
	require.NotNil(t, s)
}

// --- Metadata tests ---

func TestStrategy_Name(t *testing.T) {
	s, err := bollinger.New(model.TimeframeDaily, 20, 2.0)
	require.NoError(t, err)
	require.Equal(t, "bollinger-mean-reversion", s.Name())
}

func TestStrategy_Timeframe(t *testing.T) {
	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := bollinger.New(tf, 20, 2.0)
		require.NoError(t, err)
		require.Equal(t, tf, s.Timeframe())
	}
}

func TestStrategy_Lookback_equalsPeriod(t *testing.T) {
	cases := []struct {
		period int
		want   int
	}{
		{20, 20},
		{10, 10},
		{50, 50},
	}
	for _, tc := range cases {
		s, err := bollinger.New(model.TimeframeDaily, tc.period, 2.0)
		require.NoError(t, err)
		require.Equal(t, tc.want, s.Lookback(), "period=%d", tc.period)
	}
}

// --- Signal helpers ---

// stableDropSeries returns a series of nStable bars at stablePrice followed by
// one bar at dropPrice. Intended to drive close below the lower Bollinger Band.
func stableDropSeries(nStable int, stablePrice, dropPrice float64) []float64 {
	closes := make([]float64, nStable+1)
	for i := range nStable {
		closes[i] = stablePrice
	}
	closes[nStable] = dropPrice
	return closes
}

// stableSpikeSeries returns a series of nStable bars at stablePrice followed by
// one bar at spikePrice. Intended to drive close above the upper Bollinger Band.
func stableSpikeSeries(nStable int, stablePrice, spikePrice float64) []float64 {
	closes := make([]float64, nStable+1)
	for i := range nStable {
		closes[i] = stablePrice
	}
	closes[nStable] = spikePrice
	return closes
}

// lowerBandCrossBar returns the first index i >= period where
// closes[i-1] >= lower[i-1] and closes[i] < lower[i].
// Returns -1 if no such bar exists.
func lowerBandCrossBar(closes []float64, period int, numStdDev float64) int {
	_, _, lower := talib.BBands(closes, period, numStdDev, numStdDev, talib.SMA)
	for i := period; i < len(closes); i++ {
		if closes[i-1] >= lower[i-1] && closes[i] < lower[i] {
			return i
		}
	}
	return -1
}

// upperBandCrossBar returns the first index i >= period where
// closes[i-1] <= upper[i-1] and closes[i] > upper[i].
// Returns -1 if no such bar exists.
func upperBandCrossBar(closes []float64, period int, numStdDev float64) int {
	upper, _, _ := talib.BBands(closes, period, numStdDev, numStdDev, talib.SMA)
	for i := period; i < len(closes); i++ {
		if closes[i-1] <= upper[i-1] && closes[i] > upper[i] {
			return i
		}
	}
	return -1
}

// --- Signal tests ---

func TestStrategy_Next_holdAtLookback(t *testing.T) {
	// With exactly Lookback() candles, only one valid BB value exists.
	// Crossover detection requires two consecutive valid values → Hold.
	s, err := bollinger.New(model.TimeframeDaily, 20, 2.0)
	require.NoError(t, err)

	closes := make([]float64, s.Lookback())
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got, "exactly Lookback() bars must return Hold")
}

func TestStrategy_Next_holdOnFlatPrices(t *testing.T) {
	// Flat prices → zero-width bands (lower = upper = mean = price).
	// close == lower always → close is never strictly < lower → no signal.
	period := 20
	s, err := bollinger.New(model.TimeframeDaily, period, 2.0)
	require.NoError(t, err)

	closes := make([]float64, 60)
	for i := range closes {
		closes[i] = 100
	}
	got := s.Next(testutil.MakeCandles(closes))
	require.Equal(t, model.SignalHold, got)
}

func TestStrategy_Next_buyAtLowerBandCross(t *testing.T) {
	// 50 stable bars + 1 sharp crash. The crash drives close well below the
	// lower band in one bar, while the previous bar sat exactly on the band.
	period, numStdDev := 20, 2.0
	closes := stableDropSeries(50, 100.0, 20.0)

	buyBar := lowerBandCrossBar(closes, period, numStdDev)
	require.True(t, buyBar > 0, "drop series must produce a lower-band crossover")

	s, err := bollinger.New(model.TimeframeDaily, period, numStdDev)
	require.NoError(t, err)

	// One bar before the crossover: Hold.
	got := s.Next(testutil.MakeCandles(closes[:buyBar]))
	require.Equal(t, model.SignalHold, got, "bar before lower-band cross must be Hold")

	// At the crossover bar: Buy.
	got = s.Next(testutil.MakeCandles(closes[:buyBar+1]))
	require.Equal(t, model.SignalBuy, got, "lower-band cross bar must emit Buy")
}

func TestStrategy_Next_sellAtUpperBandCross(t *testing.T) {
	// 50 stable bars + 1 sharp spike. The spike drives close well above the
	// upper band in one bar, while the previous bar sat exactly on the band.
	period, numStdDev := 20, 2.0
	closes := stableSpikeSeries(50, 100.0, 200.0)

	sellBar := upperBandCrossBar(closes, period, numStdDev)
	require.True(t, sellBar > 0, "spike series must produce an upper-band crossover")

	s, err := bollinger.New(model.TimeframeDaily, period, numStdDev)
	require.NoError(t, err)

	// One bar before the crossover: Hold.
	got := s.Next(testutil.MakeCandles(closes[:sellBar]))
	require.Equal(t, model.SignalHold, got, "bar before upper-band cross must be Hold")

	// At the crossover bar: Sell.
	got = s.Next(testutil.MakeCandles(closes[:sellBar+1]))
	require.Equal(t, model.SignalSell, got, "upper-band cross bar must emit Sell")
}

func TestStrategy_Next_holdAfterLowerCross(t *testing.T) {
	// After a lower-band crossover (Buy), a bar where close was already below
	// the lower band on the previous bar must return Hold (strict crossover).
	period, numStdDev := 20, 2.0
	closes := stableDropSeries(50, 100.0, 20.0)

	buyBar := lowerBandCrossBar(closes, period, numStdDev)
	require.True(t, buyBar > 0)

	afterBar := buyBar + 1
	if afterBar >= len(closes) {
		closes = append(closes, 20.0)
	}

	// Verify this is a pure continuation bar: previous close was already below lower.
	_, _, lower := talib.BBands(closes, period, numStdDev, numStdDev, talib.SMA)
	if closes[afterBar-1] >= lower[afterBar-1] {
		t.Skip("consecutive crossover — not a pure continuation bar")
	}

	s, err := bollinger.New(model.TimeframeDaily, period, numStdDev)
	require.NoError(t, err)

	got := s.Next(testutil.MakeCandles(closes[:afterBar+1]))
	require.Equal(t, model.SignalHold, got, "continuation bar after lower-band cross must be Hold")
}

func TestStrategy_Next_holdAfterUpperCross(t *testing.T) {
	// After an upper-band crossover (Sell), a bar where close was already above
	// the upper band on the previous bar must return Hold (strict crossover).
	period, numStdDev := 20, 2.0
	closes := stableSpikeSeries(50, 100.0, 200.0)

	sellBar := upperBandCrossBar(closes, period, numStdDev)
	require.True(t, sellBar > 0)

	afterBar := sellBar + 1
	if afterBar >= len(closes) {
		closes = append(closes, 200.0)
	}

	// Verify this is a pure continuation bar: previous close was already above upper.
	upper, _, _ := talib.BBands(closes, period, numStdDev, numStdDev, talib.SMA)
	if closes[afterBar-1] <= upper[afterBar-1] {
		t.Skip("consecutive crossover — not a pure continuation bar")
	}

	s, err := bollinger.New(model.TimeframeDaily, period, numStdDev)
	require.NoError(t, err)

	got := s.Next(testutil.MakeCandles(closes[:afterBar+1]))
	require.Equal(t, model.SignalHold, got, "continuation bar after upper-band cross must be Hold")
}
