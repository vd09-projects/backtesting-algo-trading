package model

import (
	"fmt"
	"time"
)

// Timeframe represents a candle bar duration.
type Timeframe string

const (
	Timeframe1Min   Timeframe = "1min"
	Timeframe5Min   Timeframe = "5min"
	Timeframe15Min  Timeframe = "15min"
	TimeframeDaily  Timeframe = "daily"
	TimeframeWeekly Timeframe = "weekly"
)

func (t Timeframe) String() string {
	return string(t)
}

// Duration returns the time.Duration corresponding to this timeframe.
func (t Timeframe) Duration() (time.Duration, error) {
	switch t {
	case Timeframe1Min:
		return time.Minute, nil
	case Timeframe5Min:
		return 5 * time.Minute, nil
	case Timeframe15Min:
		return 15 * time.Minute, nil
	case TimeframeDaily:
		return 24 * time.Hour, nil
	case TimeframeWeekly:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unknown timeframe: %q", t)
	}
}
