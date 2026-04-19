package analytics_test

import (
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// makeDailyCurve builds a synthetic daily equity curve from startDate for n days.
// values cycles through the provided value list. startDate is inclusive.
func makeDailyCurve(startDate time.Time, n int, values []float64) []model.EquityPoint {
	pts := make([]model.EquityPoint, n)
	for i := range pts {
		pts[i] = model.EquityPoint{
			Timestamp: startDate.AddDate(0, 0, i),
			Value:     values[i%len(values)],
		}
	}
	return pts
}

func TestComputeRegimeSplits_ThreeRegimes(t *testing.T) {
	// 730 days starting 2018-01-01 spanning the first two NSE regimes.
	// Values alternate slightly so Sharpe is non-zero.
	start := time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
	curve := makeDailyCurve(start, 2558, []float64{100000, 100100, 100050, 100200}) // ~7 years

	regimes := analytics.NSERegimes2018_2024
	splits := analytics.ComputeRegimeSplits(curve, regimes, model.TimeframeDaily)

	if len(splits) != len(regimes) {
		t.Fatalf("got %d regime reports, want %d", len(splits), len(regimes))
	}

	for i, r := range splits {
		if r.Name != regimes[i].Name {
			t.Errorf("splits[%d].Name: got %q, want %q", i, r.Name, regimes[i].Name)
		}
		if !r.From.Equal(regimes[i].From) {
			t.Errorf("splits[%d].From: got %v, want %v", i, r.From, regimes[i].From)
		}
		if !r.To.Equal(regimes[i].To) {
			t.Errorf("splits[%d].To: got %v, want %v", i, r.To, regimes[i].To)
		}
		if r.MaxDrawdown < 0 {
			t.Errorf("splits[%d].MaxDrawdown: got %v, want >= 0", i, r.MaxDrawdown)
		}
		// With a curve that has some variation, Sharpe should be non-zero.
		if r.SharpeRatio == 0 {
			t.Errorf("splits[%d].SharpeRatio: got 0, want non-zero for regime %q", i, r.Name)
		}
	}
}

func TestComputeRegimeSplits_RegimeBoundaries(t *testing.T) {
	// Boundary: point exactly at regime.To must not appear in that regime's slice.
	// Three in-range points give 2 returns, which is enough for a non-zero Sharpe.
	regime := analytics.Regime{
		Name: "test",
		From: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	curve := []model.EquityPoint{
		{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100000},
		{Timestamp: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC), Value: 105000},
		{Timestamp: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC), Value: 103000},
		{Timestamp: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC), Value: 110000}, // exactly at To — excluded
	}

	splits := analytics.ComputeRegimeSplits(curve, []analytics.Regime{regime}, model.TimeframeDaily)

	if len(splits) != 1 {
		t.Fatalf("got %d splits, want 1", len(splits))
	}
	// 3 in-range points → 2 returns → Sharpe is computable (non-zero).
	if splits[0].SharpeRatio == 0 {
		t.Errorf("expected non-zero Sharpe with 3 in-range points")
	}
}

func TestComputeRegimeSplits_EmptyRegime(t *testing.T) {
	// Regime that has no curve points within it → zeroed metrics, name/period preserved.
	regime := analytics.Regime{
		Name: "no-data",
		From: time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	curve := []model.EquityPoint{
		{Timestamp: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), Value: 100000},
	}

	splits := analytics.ComputeRegimeSplits(curve, []analytics.Regime{regime}, model.TimeframeDaily)

	if len(splits) != 1 {
		t.Fatalf("got %d splits, want 1", len(splits))
	}
	r := splits[0]
	if r.Name != "no-data" {
		t.Errorf("Name: got %q, want %q", r.Name, "no-data")
	}
	if r.SharpeRatio != 0 {
		t.Errorf("SharpeRatio: got %v, want 0", r.SharpeRatio)
	}
	if r.MaxDrawdown != 0 {
		t.Errorf("MaxDrawdown: got %v, want 0", r.MaxDrawdown)
	}
}

func TestComputeRegimeSplits_EmptyCurve(t *testing.T) {
	splits := analytics.ComputeRegimeSplits(nil, analytics.NSERegimes2018_2024, model.TimeframeDaily)
	if len(splits) != len(analytics.NSERegimes2018_2024) {
		t.Fatalf("got %d splits, want %d", len(splits), len(analytics.NSERegimes2018_2024))
	}
	for i, r := range splits {
		if r.SharpeRatio != 0 || r.MaxDrawdown != 0 {
			t.Errorf("splits[%d]: expected zeroed metrics for empty curve", i)
		}
	}
}
