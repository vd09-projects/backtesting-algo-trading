package strategy_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// testIST is the IST timezone used in tests, matching the production constant.
var testIST = time.FixedZone("IST", 5*3600+30*60)

// makeIntraday builds a slice of 5-min candles for the given IST date starting at
// startHour:startMin and advancing 5 minutes per bar for count bars.
// Close values are sequential floats starting at 100.0 + float64(i).
func makeIntraday(date time.Time, startHour, startMin, count int) []model.Candle {
	cs := make([]model.Candle, count)
	for i := range cs {
		ts := time.Date(
			date.Year(), date.Month(), date.Day(),
			startHour, startMin+i*5, 0, 0,
			testIST,
		)
		// time.Date normalises overflowing minutes automatically.
		price := 100.0 + float64(i)
		cs[i] = model.Candle{
			Instrument: "NSE:TEST",
			Timeframe:  model.Timeframe5Min,
			Timestamp:  ts,
			Open:       price,
			High:       price,
			Low:        price,
			Close:      price,
			Volume:     1000,
		}
	}
	return cs
}

// concatBars concatenates two candle slices into a new slice.
// Using this helper avoids the appendAssign lint warning.
func concatBars(a, b []model.Candle) []model.Candle {
	out := make([]model.Candle, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// --- IsSessionOpen tests ---

// TestIsSessionOpen_TrueAt0915: a bar stamped exactly 09:15 IST must return true.
func TestIsSessionOpen_TrueAt0915(t *testing.T) {
	ts := time.Date(2024, 1, 15, 9, 15, 0, 0, testIST)
	bar := model.Candle{
		Instrument: "NSE:TEST",
		Timeframe:  model.Timeframe5Min,
		Timestamp:  ts,
		Open:       100, High: 100, Low: 100, Close: 100, Volume: 1000,
	}
	require.True(t, strategy.IsSessionOpen(bar), "09:15 IST bar must return true")
}

// TestIsSessionOpen_FalseAt0920: a bar stamped 09:20 IST must return false.
func TestIsSessionOpen_FalseAt0920(t *testing.T) {
	ts := time.Date(2024, 1, 15, 9, 20, 0, 0, testIST)
	bar := model.Candle{
		Instrument: "NSE:TEST",
		Timeframe:  model.Timeframe5Min,
		Timestamp:  ts,
		Open:       100, High: 100, Low: 100, Close: 100, Volume: 1000,
	}
	require.False(t, strategy.IsSessionOpen(bar), "09:20 IST bar must return false")
}

// TestIsSessionOpen_FalseAt0900: 09:00 IST is before the NSE session open.
func TestIsSessionOpen_FalseAt0900(t *testing.T) {
	ts := time.Date(2024, 1, 15, 9, 0, 0, 0, testIST)
	bar := model.Candle{
		Instrument: "NSE:TEST",
		Timeframe:  model.Timeframe5Min,
		Timestamp:  ts,
		Open:       100, High: 100, Low: 100, Close: 100, Volume: 1000,
	}
	require.False(t, strategy.IsSessionOpen(bar), "09:00 IST bar must return false")
}

// TestIsSessionOpen_FalseAtMidday: a bar in the middle of the session must return false.
func TestIsSessionOpen_FalseAtMidday(t *testing.T) {
	ts := time.Date(2024, 1, 15, 12, 0, 0, 0, testIST)
	bar := model.Candle{
		Instrument: "NSE:TEST",
		Timeframe:  model.Timeframe5Min,
		Timestamp:  ts,
		Open:       100, High: 100, Low: 100, Close: 100, Volume: 1000,
	}
	require.False(t, strategy.IsSessionOpen(bar), "12:00 IST bar must return false")
}

// TestIsSessionOpen_UTCTimestamp_StillDetectedCorrectly: timestamps stored in UTC
// (09:15 IST = 03:45 UTC) must still be detected correctly.
func TestIsSessionOpen_UTCTimestamp_StillDetectedCorrectly(t *testing.T) {
	// 09:15 IST = 03:45 UTC
	tsUTC := time.Date(2024, 1, 15, 3, 45, 0, 0, time.UTC)
	bar := model.Candle{
		Instrument: "NSE:TEST",
		Timeframe:  model.Timeframe5Min,
		Timestamp:  tsUTC,
		Open:       100, High: 100, Low: 100, Close: 100, Volume: 1000,
	}
	require.True(t, strategy.IsSessionOpen(bar), "03:45 UTC (= 09:15 IST) bar must return true")
}

// TestIsSessionOpen_GoldenTest_FiresExactlyOncePerDay: synthetic 5-min slice
// spanning 2 full sessions. IsSessionOpen must fire exactly once per day
// (the 09:15 bar), and the count across both sessions must be exactly 2.
func TestIsSessionOpen_GoldenTest_FiresExactlyOncePerDay(t *testing.T) {
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)
	day2 := time.Date(2024, 1, 16, 0, 0, 0, 0, testIST)

	// NSE session: 09:15 to 15:30 = 75 bars of 5 min
	all := concatBars(makeIntraday(day1, 9, 15, 75), makeIntraday(day2, 9, 15, 75))

	openCounts := map[string]int{}
	for _, bar := range all {
		if strategy.IsSessionOpen(bar) {
			dateStr := bar.Timestamp.In(testIST).Format("2006-01-02")
			openCounts[dateStr]++
		}
	}

	require.Equal(t, 2, len(openCounts), "IsSessionOpen should fire on exactly 2 distinct IST dates")
	for date, count := range openCounts {
		require.Equal(t, 1, count, "IsSessionOpen should fire exactly once on %s", date)
	}
}

// --- PreviousSessionClose tests ---

// TestPreviousSessionClose_NoHistory_ReturnsFalse: only bars from today in slice.
// PreviousSessionClose must return (0, false).
func TestPreviousSessionClose_NoHistory_ReturnsFalse(t *testing.T) {
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)
	bars := makeIntraday(day1, 9, 15, 10) // only day1 bars
	gotClose, ok := strategy.PreviousSessionClose(bars, len(bars)-1)
	require.False(t, ok, "no prior session → ok must be false")
	require.Equal(t, 0.0, gotClose, "no prior session → close must be 0")
}

// TestPreviousSessionClose_ReturnsLastBarOfPriorSession: bars from day1 and day2.
// Calling from any bar on day2 must return the close of the last bar of day1.
func TestPreviousSessionClose_ReturnsLastBarOfPriorSession(t *testing.T) {
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)
	day2 := time.Date(2024, 1, 16, 0, 0, 0, 0, testIST)

	// day1: 09:15–15:30 (75 bars), close values 100..174
	session1 := makeIntraday(day1, 9, 15, 75)
	// day2: 09:15–15:30 (75 bars)
	session2 := makeIntraday(day2, 9, 15, 75)
	all := concatBars(session1, session2)

	// last bar of day1 is session1[74], close = 100 + 74 = 174.0
	wantClose := session1[74].Close

	// Call from several different bars in session2 — all must return same prior close.
	for _, idx := range []int{75, 80, 90, 149} {
		got, ok := strategy.PreviousSessionClose(all, idx)
		require.True(t, ok, "idx=%d: prior session exists → ok must be true", idx)
		require.Equal(t, wantClose, got, "idx=%d: close must equal last bar of day1", idx)
	}
}

// TestPreviousSessionClose_CalledAtFirstBarOfDay: calling at i=75 (the 09:15 bar of day2)
// must still find and return the prior session close correctly.
func TestPreviousSessionClose_CalledAtFirstBarOfDay(t *testing.T) {
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)
	day2 := time.Date(2024, 1, 16, 0, 0, 0, 0, testIST)

	session1 := makeIntraday(day1, 9, 15, 75)
	session2 := makeIntraday(day2, 9, 15, 75)
	all := concatBars(session1, session2)

	wantClose := session1[74].Close // last bar of day1

	got, ok := strategy.PreviousSessionClose(all, 75) // first bar of day2
	require.True(t, ok)
	require.Equal(t, wantClose, got, "first bar of day2 → must still find day1 close")
}

// TestPreviousSessionClose_WeekendHolidayGap: gap of >1 calendar day between sessions.
// Monday's data starts at index after Friday — PreviousSessionClose must return Friday's last bar close.
func TestPreviousSessionClose_WeekendHolidayGap(t *testing.T) {
	// Friday 2024-01-12, Monday 2024-01-15
	friday := time.Date(2024, 1, 12, 0, 0, 0, 0, testIST)
	monday := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)

	sessionFri := makeIntraday(friday, 9, 15, 75)
	sessionMon := makeIntraday(monday, 9, 15, 75)
	all := concatBars(sessionFri, sessionMon)

	wantClose := sessionFri[74].Close // last bar of Friday

	// Call from any bar on Monday
	for _, idx := range []int{75, 100, 149} {
		got, ok := strategy.PreviousSessionClose(all, idx)
		require.True(t, ok, "idx=%d: Friday data exists → ok must be true", idx)
		require.Equal(t, wantClose, got, "idx=%d: must return Friday's last close despite weekend gap", idx)
	}
}

// TestPreviousSessionClose_IndexZero_ReturnsFalse: called at i=0 (no prior bars at all).
func TestPreviousSessionClose_IndexZero_ReturnsFalse(t *testing.T) {
	day1 := time.Date(2024, 1, 15, 0, 0, 0, 0, testIST)
	bars := makeIntraday(day1, 9, 15, 10)
	gotClose, ok := strategy.PreviousSessionClose(bars, 0)
	require.False(t, ok)
	require.Equal(t, 0.0, gotClose)
}
