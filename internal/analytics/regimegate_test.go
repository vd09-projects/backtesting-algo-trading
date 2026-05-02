package analytics_test

import (
	"math"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeTrade creates a Trade with ExitTime t and ReturnOnNotional approximately equal to r.
// EntryPrice=100, Quantity=1 → RealizedPnL = r * 100.
func makeTrade(exitTime time.Time, r float64) model.Trade {
	return model.Trade{
		Instrument:  "NSE:TEST",
		Direction:   model.DirectionLong,
		Quantity:    1,
		EntryPrice:  100,
		ExitPrice:   100 * (1 + r),
		EntryTime:   exitTime.Add(-24 * time.Hour),
		ExitTime:    exitTime,
		RealizedPnL: r * 100,
	}
}

// threeRegimes returns three named windows for testing (matches decision-file windows).
var threeRegimes = []analytics.Regime{
	{
		Name: "Pre-COVID",
		From: time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "COVID",
		From: time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
	},
	{
		Name: "Post-recovery",
		From: time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	},
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_EvenlySpread: trades spread evenly across regimes.
// Each regime gets identical returns → per-trade Sharpe identical → no concentration.
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_EvenlySpread(t *testing.T) {
	t.Parallel()

	// 6 trades per regime, identical returns → Sharpe identical across regimes.
	// Contributions ≈ 0.33 each — well under 0.70.
	var trades []model.Trade
	preCovidDates := []time.Time{
		time.Date(2018, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	covidDates := []time.Time{
		time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	postDates := []time.Time{
		time.Date(2021, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
	}

	for _, d := range preCovidDates {
		trades = append(trades, makeTrade(d, 0.05))
	}
	for _, d := range covidDates {
		trades = append(trades, makeTrade(d, 0.05))
	}
	for _, d := range postDates {
		trades = append(trades, makeTrade(d, 0.05))
	}

	report := analytics.ComputeRegimeGate(trades, threeRegimes)

	if report.RegimeConcentrated {
		t.Error("RegimeConcentrated: want false for evenly spread trades, got true")
	}
	if len(report.Regimes) != 3 {
		t.Fatalf("expected 3 regime contributions, got %d", len(report.Regimes))
	}
	for i, rc := range report.Regimes {
		if rc.TradeCount != 6 {
			t.Errorf("regime[%d] TradeCount: want 6, got %d", i, rc.TradeCount)
		}
		if rc.Contribution >= 0.70 {
			t.Errorf("regime[%d] Contribution: want < 0.70, got %.4f", i, rc.Contribution)
		}
	}
	// Contributions should sum to approximately 1.0.
	var total float64
	for _, rc := range report.Regimes {
		total += rc.Contribution
	}
	if math.Abs(total-1.0) > 1e-6 {
		t.Errorf("contributions sum: want 1.0, got %.6f", total)
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_ConcentratedInOnePeriod: all trades in one regime.
// RegimeConcentrated must be true.
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_ConcentratedInOnePeriod(t *testing.T) {
	t.Parallel()

	// All trades in Pre-COVID, nothing in COVID or Post-recovery.
	// Zero-trade regimes trigger RegimeConcentrated per decision file.
	dates := []time.Time{
		time.Date(2018, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2018, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2019, 10, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	var trades []model.Trade
	for _, d := range dates {
		trades = append(trades, makeTrade(d, 0.05))
	}

	report := analytics.ComputeRegimeGate(trades, threeRegimes)

	if !report.RegimeConcentrated {
		t.Error("RegimeConcentrated: want true when empty regimes exist, got false")
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_HighSharpeConcentration: Sharpe mass dominated by one regime.
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_HighSharpeConcentration(t *testing.T) {
	t.Parallel()

	// Pre-COVID: high consistent returns (large Sharpe).
	// COVID and Post-recovery: near-zero returns (tiny Sharpe).
	// → Pre-COVID should dominate abs-contribution sum.
	var trades []model.Trade

	// Pre-COVID: 10 trades with high return (large Sharpe).
	for i := 0; i < 10; i++ {
		d := time.Date(2018, time.Month(1+i), 15, 0, 0, 0, 0, time.UTC)
		if d.After(time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)) {
			break
		}
		trades = append(trades, makeTrade(d, 0.10))
	}

	// COVID: 10 trades with tiny variance around 0 (near-zero Sharpe).
	covidReturns := []float64{0.0001, -0.0001, 0.0001, -0.0001, 0.0001, -0.0001, 0.0001, -0.0001, 0.0001, -0.0001}
	for i, r := range covidReturns {
		d := time.Date(2020, time.Month(2+(i%6)), 15, 0, 0, 0, 0, time.UTC)
		trades = append(trades, makeTrade(d, r))
	}

	// Post-recovery: same near-zero.
	for i, r := range covidReturns {
		d := time.Date(2021, time.Month(7+(i%6)), 15, 0, 0, 0, 0, time.UTC)
		trades = append(trades, makeTrade(d, r))
	}

	report := analytics.ComputeRegimeGate(trades, threeRegimes)

	// Pre-COVID has a large per-trade Sharpe (high consistent returns).
	// COVID and Post-recovery have near-zero Sharpe.
	// Pre-COVID's abs(S) should dominate.
	if len(report.Regimes) != 3 {
		t.Fatalf("expected 3 regimes, got %d", len(report.Regimes))
	}
	preCOVID := report.Regimes[0]
	if preCOVID.Name != "Pre-COVID" {
		t.Errorf("expected first regime to be Pre-COVID, got %q", preCOVID.Name)
	}
	if preCOVID.Contribution < 0.70 {
		t.Errorf("Pre-COVID contribution: want >= 0.70 for concentrated strategy, got %.4f", preCOVID.Contribution)
	}
	if !report.RegimeConcentrated {
		t.Error("RegimeConcentrated: want true for Sharpe-concentrated strategy, got false")
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_EmptyTrades: no trades → RegimeConcentrated=true.
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_EmptyTrades(t *testing.T) {
	t.Parallel()

	report := analytics.ComputeRegimeGate(nil, threeRegimes)

	if !report.RegimeConcentrated {
		t.Error("RegimeConcentrated: want true for empty trades, got false")
	}
	if len(report.Regimes) != 3 {
		t.Fatalf("expected 3 regime contributions, got %d", len(report.Regimes))
	}
	for i, rc := range report.Regimes {
		if rc.TradeCount != 0 {
			t.Errorf("regime[%d] TradeCount: want 0 for empty trades, got %d", i, rc.TradeCount)
		}
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_AllSharpeZero: returns alternate +/- same magnitude.
// All abs contributions equal (sum zero per-regime sharpe in edge case).
// The gate should still produce sensible output without divide-by-zero.
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_AllSharpeZero(t *testing.T) {
	t.Parallel()

	// All regimes have trades that cancel to zero Sharpe (alternating equal returns).
	// sum(abs(S[j])) == 0 → contributions undefined. Must not panic.
	var trades []model.Trade

	for i := 0; i < 4; i++ {
		d := time.Date(2018, time.Month(3+i*3), 1, 0, 0, 0, 0, time.UTC)
		r := 0.05
		if i%2 == 1 {
			r = -0.05
		}
		trades = append(trades, makeTrade(d, r))
	}
	for i := 0; i < 4; i++ {
		d := time.Date(2020, time.Month(3+i*3), 1, 0, 0, 0, 0, time.UTC)
		r := 0.05
		if i%2 == 1 {
			r = -0.05
		}
		trades = append(trades, makeTrade(d, r))
	}
	for i := 0; i < 4; i++ {
		d := time.Date(2021, time.Month(9+i%4), 1, 0, 0, 0, 0, time.UTC)
		r := 0.05
		if i%2 == 1 {
			r = -0.05
		}
		trades = append(trades, makeTrade(d, r))
	}

	// Must not panic.
	report := analytics.ComputeRegimeGate(trades, threeRegimes)

	if len(report.Regimes) != 3 {
		t.Fatalf("expected 3 regime contributions, got %d", len(report.Regimes))
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_NSERegimesGate: verify the exported constant has the
// correct boundaries from the decision file (2026-04-27).
// ---------------------------------------------------------------------------

func TestNSERegimesGate_Boundaries(t *testing.T) {
	t.Parallel()

	regimes := analytics.NSERegimesGate
	if len(regimes) != 3 {
		t.Fatalf("expected 3 NSERegimesGate entries, got %d", len(regimes))
	}

	wantFrom := []time.Time{
		time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
	}
	wantTo := []time.Time{
		time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	for i, r := range regimes {
		if !r.From.Equal(wantFrom[i]) {
			t.Errorf("NSERegimesGate[%d].From: want %v, got %v", i, wantFrom[i], r.From)
		}
		if !r.To.Equal(wantTo[i]) {
			t.Errorf("NSERegimesGate[%d].To: want %v, got %v", i, wantTo[i], r.To)
		}
	}
}

// ---------------------------------------------------------------------------
// TestComputeRegimeGate_TradeAssignedByExitTime: trade on boundary date goes
// to the regime that starts on that date (inclusive lower bound).
// ---------------------------------------------------------------------------

func TestComputeRegimeGate_TradeAssignedByExitTime(t *testing.T) {
	t.Parallel()

	// Trade exactly at the COVID regime start (2020-02-01) → assigned to COVID.
	boundaryTrade := makeTrade(time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC), 0.05)

	// Pre-COVID: 5 trades.
	var trades []model.Trade
	for i := 0; i < 5; i++ {
		trades = append(trades, makeTrade(time.Date(2018, time.Month(2+i), 1, 0, 0, 0, 0, time.UTC), 0.05))
	}
	// Boundary trade: should go to COVID.
	trades = append(trades, boundaryTrade)
	// Post-recovery: 5 trades.
	for i := 0; i < 5; i++ {
		trades = append(trades, makeTrade(time.Date(2021, time.Month(8+i), 1, 0, 0, 0, 0, time.UTC), 0.05))
	}

	report := analytics.ComputeRegimeGate(trades, threeRegimes)

	if len(report.Regimes) != 3 {
		t.Fatalf("expected 3 regimes, got %d", len(report.Regimes))
	}

	covidRegime := report.Regimes[1]
	if covidRegime.TradeCount != 1 {
		t.Errorf("COVID regime TradeCount: want 1 (boundary trade), got %d", covidRegime.TradeCount)
	}
}
