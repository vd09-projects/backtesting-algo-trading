package zerodha

import (
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// maxDaysPerInterval maps each supported Timeframe to the maximum date range
// allowed per Kite Connect historical API request. Values include a safety
// margin below the published limit to avoid timezone boundary edge cases.
var maxDaysPerInterval = map[model.Timeframe]int{
	model.Timeframe1Min:  55,  // API limit: 60 days; 5-day margin
	model.Timeframe5Min:  90,  // API limit: 100 days; 10-day margin
	model.Timeframe15Min: 180, // API limit: 200 days; 20-day margin
	model.TimeframeDaily: 1800,
	// TimeframeWeekly intentionally absent — Kite Connect has no weekly interval.
}

type dateWindow struct {
	from time.Time
	to   time.Time
}

// chunkDateRange splits [from, to) into consecutive windows no wider than
// maxDaysPerInterval[tf] days. Windows are half-open: [w.from, w.to).
// The next window starts exactly at the prior window's w.to.
//
// Returns ErrUnsupportedTimeframe if tf has no entry in maxDaysPerInterval.
// Returns an empty slice if from >= to.
func chunkDateRange(from, to time.Time, tf model.Timeframe) ([]dateWindow, error) {
	maxDays, ok := maxDaysPerInterval[tf]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedTimeframe, tf)
	}

	step := time.Duration(maxDays) * 24 * time.Hour
	var windows []dateWindow
	cur := from
	for cur.Before(to) {
		end := cur.Add(step)
		if end.After(to) {
			end = to
		}
		windows = append(windows, dateWindow{from: cur, to: end})
		cur = end
	}
	return windows, nil
}
