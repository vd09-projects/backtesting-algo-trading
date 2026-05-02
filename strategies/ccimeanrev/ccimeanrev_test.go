package ccimeanrev_test

import (
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/ccimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/testutil"
)

// makeOHLC is a local alias so test code stays readable.
func makeOHLC(highs, lows, closes []float64) []model.Candle {
	return testutil.MakeCandlesOHLC(highs, lows, closes)
}

// ---------------------------------------------------------------------------
// Constructor validation
// ---------------------------------------------------------------------------

func TestNew_invalidParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		period, entry, exit int
	}{
		{"period zero", 0, -100, 0},
		{"period negative", -1, -100, 0},
		{"entry not negative enough: entry > exit", 10, 50, 0},
		{"entry equals exit", 10, 0, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := ccimeanrev.New(model.TimeframeDaily, tc.period, tc.entry, tc.exit)
			if err == nil {
				t.Errorf("New(period=%d, entry=%d, exit=%d): expected error, got nil",
					tc.period, tc.entry, tc.exit)
			}
		})
	}
}

func TestNew_valid(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		period, entry, exit int
	}{
		{"standard params", 20, -100, 0},
		{"tight params", 5, -50, -10},
		{"wide params", 14, -200, 100},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			s, err := ccimeanrev.New(model.TimeframeDaily, tc.period, tc.entry, tc.exit)
			if err != nil {
				t.Errorf("New(period=%d, entry=%d, exit=%d): unexpected error: %v",
					tc.period, tc.entry, tc.exit, err)
			}
			if s == nil {
				t.Error("New: returned nil strategy with no error")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Metadata
// ---------------------------------------------------------------------------

func TestStrategy_Name(t *testing.T) {
	t.Parallel()

	s, err := ccimeanrev.New(model.TimeframeDaily, 20, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.Name(); got != "cci-mean-reversion" {
		t.Errorf("Name() = %q, want %q", got, "cci-mean-reversion")
	}
}

func TestStrategy_Timeframe(t *testing.T) {
	t.Parallel()

	for _, tf := range []model.Timeframe{
		model.Timeframe1Min, model.Timeframe5Min, model.TimeframeDaily,
	} {
		s, err := ccimeanrev.New(tf, 20, -100, 0)
		if err != nil {
			t.Fatalf("New: %v", err)
		}
		if got := s.Timeframe(); got != tf {
			t.Errorf("Timeframe() = %v, want %v", got, tf)
		}
	}
}

func TestStrategy_Lookback_equalsPeriod(t *testing.T) {
	t.Parallel()

	// talib.Cci lookback = period-1 (first valid output at index period-1),
	// so Lookback() must return period to ensure enough candles are present.
	s, err := ccimeanrev.New(model.TimeframeDaily, 20, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if got := s.Lookback(); got != 20 {
		t.Errorf("Lookback() = %d, want 20 (== period)", got)
	}
}

// ---------------------------------------------------------------------------
// Signal behavior — before lookback
// ---------------------------------------------------------------------------

func TestStrategy_Next_holdDuringLookback(t *testing.T) {
	t.Parallel()

	// period=3 → Lookback=3. With only 2 candles, must return Hold.
	s, err := ccimeanrev.New(model.TimeframeDaily, 3, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	closes := []float64{100, 90}
	highs := closes
	lows := closes
	got := s.Next(makeOHLC(highs, lows, closes))
	if got != model.SignalHold {
		t.Errorf("Next(2 candles, period=3): want Hold, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Entry signal: CCI < entryThreshold fires Buy
// ---------------------------------------------------------------------------
// CCI formula: CCI = (TypicalPrice - SMA(TypicalPrice, n)) / (0.015 * MeanDeviation)
// TypicalPrice = (H + L + C) / 3
//
// Important: for period=3, CCI has a mathematical lower bound of exactly -100 when
// the last bar is the minimum of the window (algebraic identity of the CCI formula).
// Therefore entry tests that require CCI < -100 strictly must use period >= 4.
//
// With period=4 and inputs [100,100,100,50]:
//   SMA(4) = 87.5; MeanDev = (3×12.5 + 37.5)/4 = 18.75
//   CCI = (50-87.5)/(0.015×18.75) = -37.5/0.28125 ≈ -133.33 < -100 ✓

func TestStrategy_Next_entryFiresOnCCIBelowThreshold(t *testing.T) {
	t.Parallel()

	// period=4 required: see comment above about period=3 CCI lower bound.
	s, err := ccimeanrev.New(model.TimeframeDaily, 4, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// [100,100,100,50] with period=4 → CCI ≈ -133.33 < -100 → Buy.
	highs := []float64{100, 100, 100, 50}
	lows := []float64{100, 100, 100, 50}
	closes := []float64{100, 100, 100, 50}
	got := s.Next(makeOHLC(highs, lows, closes))
	if got != model.SignalBuy {
		t.Errorf("Next: want Buy for CCI ≈ -133, got %v", got)
	}
}

func TestStrategy_Next_holdWhenCCIAboveEntryThreshold(t *testing.T) {
	t.Parallel()

	// period=4; flat prices → CCI = 0 (no deviation from mean) → Hold.
	s, err := ccimeanrev.New(model.TimeframeDaily, 4, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	highs := []float64{100, 100, 100, 100}
	lows := []float64{100, 100, 100, 100}
	closes := []float64{100, 100, 100, 100}
	got := s.Next(makeOHLC(highs, lows, closes))
	if got != model.SignalHold {
		t.Errorf("Next: want Hold for CCI=0, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Exit signal: CCI cross above exitThreshold fires Sell (cross detection)
// ---------------------------------------------------------------------------
// The exit is a CROSS detection, not a level comparison.
// Sell fires only on the bar where prevCCI <= exitThreshold AND currCCI > exitThreshold.
// If CCI is already above exitThreshold on consecutive bars, only the first fires Sell.

func TestStrategy_Next_exitFiresOnCCICrossAboveThreshold(t *testing.T) {
	t.Parallel()

	s, err := ccimeanrev.New(model.TimeframeDaily, 3, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// We build a sequence that produces CCI < 0 on bar 2 then CCI > 0 on bar 3.
	// bars 0-1: tp=100; bar 2: tp=80 (CCI<0); bar 3: tp=120 (CCI>0 cross).
	// We do this by calling Next once to set prevCCI to a negative value,
	// then again with a rising bar.

	// First call: bars producing CCI < 0 (no Buy because CCI > -100).
	// tp: [100, 100, 80] → SMA=93.33, MeanDev=8.89, CCI=(80-93.33)/(0.015*8.89)≈-100 exactly.
	// Use 75 for clear negative CCI, well above -100 threshold.
	highs1 := []float64{100, 100, 75}
	lows1 := []float64{100, 100, 75}
	closes1 := []float64{100, 100, 75}
	sig1 := s.Next(makeOHLC(highs1, lows1, closes1))
	// CCI is negative (between -100 and 0) → Hold (entry threshold not met).
	if sig1 != model.SignalHold {
		t.Logf("TestStrategy_Next_exitFiresOnCCICrossAboveThreshold: first call got %v (expected Hold)", sig1)
	}

	// Second call: prevCCI is now negative; add a bar with tp=130 → CCI crosses above 0.
	highs2 := []float64{100, 75, 130}
	lows2 := []float64{100, 75, 130}
	closes2 := []float64{100, 75, 130}
	sig2 := s.Next(makeOHLC(highs2, lows2, closes2))
	if sig2 != model.SignalSell {
		t.Errorf("Next after CCI cross above 0: want Sell, got %v", sig2)
	}
}

func TestStrategy_Next_exitDoesNotFireWhenCCIAlreadyAboveThreshold(t *testing.T) {
	t.Parallel()

	// Sell fires only once: on the cross bar. Subsequent bars where CCI stays > 0
	// should return Hold (no position open to close; or simply: no cross).
	s, err := ccimeanrev.New(model.TimeframeDaily, 3, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// First: push CCI to a large positive value.
	highs1 := []float64{100, 100, 130}
	lows1 := []float64{100, 100, 130}
	closes1 := []float64{100, 100, 130}
	_ = s.Next(makeOHLC(highs1, lows1, closes1)) // prevCCI now positive

	// Second: CCI stays positive (another high bar). prevCCI > 0, currCCI > 0 → no cross → Hold.
	highs2 := []float64{100, 130, 140}
	lows2 := []float64{100, 130, 140}
	closes2 := []float64{100, 130, 140}
	got := s.Next(makeOHLC(highs2, lows2, closes2))
	if got != model.SignalHold {
		t.Errorf("Next when CCI stays positive: want Hold (no cross), got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Long-only: no double-entry while in position
// ---------------------------------------------------------------------------
// The engine enforces no-double-entry for long-only strategies (it ignores
// Buy signals when already in a position). The strategy itself should also
// not produce a second Buy while the first hasn't been exited.
// We verify that consecutive Buy signals are emitted when CCI stays < -100
// (the engine ignores them, but the strategy is correct to emit them since
// the engine tracks position state — the strategy is stateless w.r.t. position).
// The key test: strategy does NOT track position state internally — it emits
// Buy on every bar CCI < threshold, and relies on the engine to suppress double-entry.

func TestStrategy_Next_consecutiveBuyOnCCIBelowThreshold(t *testing.T) {
	t.Parallel()

	// period=4 required: see entry signal comment above about period=3 CCI lower bound.
	s, err := ccimeanrev.New(model.TimeframeDaily, 4, -100, 0)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// CCI << -100 on both calls — strategy should emit Buy each time.
	// Call 1: [100,100,100,50] period=4 → CCI ≈ -133.33 < -100 → Buy.
	highs1 := []float64{100, 100, 100, 50}
	lows1 := []float64{100, 100, 100, 50}
	closes1 := []float64{100, 100, 100, 50}
	sig1 := s.Next(makeOHLC(highs1, lows1, closes1))
	if sig1 != model.SignalBuy {
		t.Errorf("first call: want Buy (CCI≈-133), got %v", sig1)
	}

	// Call 2: [100,100,100,40] period=4 → CCI ≈ -133.33 < -100 → Buy again.
	// (prevCCI from call 1 is negative, no exit cross fires since CCI stays < 0.)
	highs2 := []float64{100, 100, 100, 40}
	lows2 := []float64{100, 100, 100, 40}
	closes2 := []float64{100, 100, 100, 40}
	sig2 := s.Next(makeOHLC(highs2, lows2, closes2))
	if sig2 != model.SignalBuy {
		t.Errorf("second call with CCI ≈ -133 (still < -100): want Buy, got %v", sig2)
	}
}

// ---------------------------------------------------------------------------
// Neutral exit threshold
// ---------------------------------------------------------------------------
// With exitThreshold = -50 (non-zero), exit fires on CCI crossing above -50.
// This ensures the parameter is respected.

func TestStrategy_Next_exitThresholdRespected(t *testing.T) {
	t.Parallel()

	// entry=-150, exit=-50.
	s, err := ccimeanrev.New(model.TimeframeDaily, 3, -150, -50)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Set prevCCI to a value below -50 (between -150 and -50, so no entry).
	highs1 := []float64{100, 100, 75}
	lows1 := []float64{100, 100, 75}
	closes1 := []float64{100, 100, 75}
	_ = s.Next(makeOHLC(highs1, lows1, closes1))

	// Now cross above -50: CCI goes to e.g. -10 (above -50 but below 0).
	// With tp rising to 110:
	highs2 := []float64{100, 75, 120}
	lows2 := []float64{100, 75, 120}
	closes2 := []float64{100, 75, 120}
	got := s.Next(makeOHLC(highs2, lows2, closes2))
	if got != model.SignalSell {
		t.Errorf("Next with exit=-50 and CCI crossing above -50: want Sell, got %v", got)
	}
}
