package zerodha

import (
	"errors"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

func TestChunkDateRange(t *testing.T) {
	day := 24 * time.Hour

	// Fixed anchor — deterministic, no time.Now()
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		from         time.Time
		to           time.Time
		tf           model.Timeframe
		wantCount    int
		wantErr      error
		wantFirstTo  time.Time // zero means don't check
		wantLastFrom time.Time // zero means don't check
	}{
		{
			name:      "single chunk — range fits within maxDays",
			from:      base,
			to:        base.Add(10 * day),
			tf:        model.TimeframeDaily,
			wantCount: 1,
		},
		{
			name:      "exactly maxDays wide — single chunk",
			from:      base,
			to:        base.Add(time.Duration(maxDaysPerInterval[model.TimeframeDaily]) * day),
			tf:        model.TimeframeDaily,
			wantCount: 1,
		},
		{
			name:        "maxDays+1 wide — two chunks; boundaries are exact",
			from:        base,
			to:          base.Add(time.Duration(maxDaysPerInterval[model.TimeframeDaily]+1) * day),
			tf:          model.TimeframeDaily,
			wantCount:   2,
			wantFirstTo: base.Add(time.Duration(maxDaysPerInterval[model.TimeframeDaily]) * day),
		},
		{
			name:         "multi-chunk with partial last window",
			from:         base,
			to:           base.Add(time.Duration(maxDaysPerInterval[model.Timeframe1Min]*3+7) * day),
			tf:           model.Timeframe1Min,
			wantCount:    4,
			wantLastFrom: base.Add(time.Duration(maxDaysPerInterval[model.Timeframe1Min]*3) * day),
		},
		{
			name:      "empty range — from == to",
			from:      base,
			to:        base,
			tf:        model.TimeframeDaily,
			wantCount: 0,
		},
		{
			name:      "inverted range — from > to",
			from:      base.Add(10 * day),
			to:        base,
			tf:        model.TimeframeDaily,
			wantCount: 0,
		},
		{
			name:    "unsupported timeframe — weekly",
			from:    base,
			to:      base.Add(10 * day),
			tf:      model.TimeframeWeekly,
			wantErr: ErrUnsupportedTimeframe,
		},
		// Verify maxDays constants for each supported timeframe.
		{
			name:      "1min — single chunk fits in 55 days",
			from:      base,
			to:        base.Add(55 * day),
			tf:        model.Timeframe1Min,
			wantCount: 1,
		},
		{
			name:      "5min — single chunk fits in 90 days",
			from:      base,
			to:        base.Add(90 * day),
			tf:        model.Timeframe5Min,
			wantCount: 1,
		},
		{
			name:      "15min — single chunk fits in 180 days",
			from:      base,
			to:        base.Add(180 * day),
			tf:        model.Timeframe15Min,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			windows, err := chunkDateRange(tc.from, tc.to, tc.tf)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("want error %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(windows) != tc.wantCount {
				t.Fatalf("want %d windows, got %d", tc.wantCount, len(windows))
			}

			if tc.wantCount == 0 {
				return
			}

			// First window starts at from.
			if !windows[0].from.Equal(tc.from) {
				t.Errorf("first window from: want %v, got %v", tc.from, windows[0].from)
			}
			// Last window ends at to.
			if !windows[len(windows)-1].to.Equal(tc.to) {
				t.Errorf("last window to: want %v, got %v", tc.to, windows[len(windows)-1].to)
			}
			// Windows are contiguous — no gaps.
			for i := 1; i < len(windows); i++ {
				if !windows[i].from.Equal(windows[i-1].to) {
					t.Errorf("gap between window %d and %d: %v to %v",
						i-1, i, windows[i-1].to, windows[i].from)
				}
			}

			if !tc.wantFirstTo.IsZero() && !windows[0].to.Equal(tc.wantFirstTo) {
				t.Errorf("first window to: want %v, got %v", tc.wantFirstTo, windows[0].to)
			}
			if !tc.wantLastFrom.IsZero() && !windows[len(windows)-1].from.Equal(tc.wantLastFrom) {
				t.Errorf("last window from: want %v, got %v", tc.wantLastFrom, windows[len(windows)-1].from)
			}
		})
	}
}
