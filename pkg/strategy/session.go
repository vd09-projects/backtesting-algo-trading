package strategy

import (
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ist is the IST timezone (UTC+5:30). Defined as a package-level var using
// time.FixedZone — no tzdata dependency, consistent with engine convention.
//
// **Decision (IST package-level var) — convention: experimental**
// Constructed once at package load as a pure value; zero side effects.
// Re-constructing per call would be correct but wasteful; a package var is the
// established pattern for fixed-offset zones with no OS tzdata dependency.
var ist = time.FixedZone("IST", 5*3600+30*60)

// IsSessionOpen returns true if bar's timestamp represents the first bar of an
// NSE trading session — i.e., the timestamp in IST is exactly 09:15.
//
// The NSE regular session opens at 09:15 IST. A 5-min bar stamped 09:15 is the
// first bar of the day. This function is timezone-safe: timestamps stored in UTC
// or any other zone are converted to IST before comparison.
func IsSessionOpen(bar model.Candle) bool { //nolint:gocritic // hugeParam: Candle is always passed by value throughout the codebase; pointer would be an inconsistency
	t := bar.Timestamp.In(ist)
	return t.Hour() == 9 && t.Minute() == 15
}

// PreviousSessionClose scans backward from bars[i] to find the most recent bar
// whose IST calendar date is earlier than bars[i]'s IST date. It returns that
// bar's Close price and true. If no such bar exists (i == 0 or all bars are on
// the same IST calendar day), it returns (0, false).
//
// This is the correct way to find "what did the prior session close at?" for
// intraday strategies like gap-and-go (TASK-0075) and ORB (TASK-0074):
//
//   - It handles weekends and NSE holidays transparently — no knowledge of the
//     trading calendar is required; it simply finds the most recent prior date.
//   - It is correct when called at any point during the session, including the
//     very first bar (09:15) of a new day.
//   - It does not require the bars slice to be dense; any gap between sessions
//     is handled naturally by the calendar-day comparison.
func PreviousSessionClose(bars []model.Candle, i int) (float64, bool) {
	if i <= 0 {
		return 0, false
	}

	currentDate := bars[i].Timestamp.In(ist)
	currentDay := currentDate.YearDay()
	currentYear := currentDate.Year()

	for j := i - 1; j >= 0; j-- {
		t := bars[j].Timestamp.In(ist)
		if t.Year() < currentYear || (t.Year() == currentYear && t.YearDay() < currentDay) {
			return bars[j].Close, true
		}
	}

	return 0, false
}
