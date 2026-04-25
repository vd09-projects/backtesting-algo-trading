package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Hand-verified arithmetic for ₹30,000 notional (CNC delivery):
//
// Buy leg:
//   brokerage      = min(0.03% × 30000, 20)       = min(9.00, 20)   = 9.0000
//   STT            = 0.10% × 30000                 = 30.0000          (both sides)
//   exchange       = 0.00345% × 30000              = 1.0350
//   SEBI           = 0.0001% × 30000               = 0.0300
//   stamp          = 0.015% × 30000                = 4.5000           (buy-side only)
//   GST            = 18% × (brokerage + exchange)  = 18% × 10.035    = 1.8063
//   buy total      = 9.00 + 30.00 + 1.035 + 0.03 + 4.50 + 1.8063   = 46.3713
//
// Sell leg:
//   brokerage      = 9.0000
//   STT            = 30.0000
//   exchange       = 1.0350
//   SEBI           = 0.0300
//   stamp          = 0  (sell-side: no stamp duty)
//   GST            = 1.8063
//   sell total     = 9.00 + 30.00 + 1.035 + 0.03 + 0 + 1.8063      = 41.8713
//
// Round-trip      = 46.3713 + 41.8713 = 88.2426  ← within ₹88-95 band

const notional30K = 30_000.0

func TestCalcZerodhaFullCommission_BuyLeg(t *testing.T) {
	got := calcZerodhaFullCommission(notional30K, true)
	assert.InDelta(t, 46.3713, got, 1e-3, "buy-leg cost at ₹30K notional")
}

func TestCalcZerodhaFullCommission_SellLeg(t *testing.T) {
	got := calcZerodhaFullCommission(notional30K, false)
	assert.InDelta(t, 41.8713, got, 1e-3, "sell-leg cost at ₹30K notional")
}

func TestCalcZerodhaFullCommission_RoundTrip(t *testing.T) {
	buy := calcZerodhaFullCommission(notional30K, true)
	sell := calcZerodhaFullCommission(notional30K, false)
	roundTrip := buy + sell
	assert.InDelta(t, 88.2426, roundTrip, 1e-3, "round-trip cost at ₹30K notional")
	// Guard: must stay within the acceptance-criteria band ₹88–95.
	assert.GreaterOrEqual(t, roundTrip, 88.0, "round-trip must be ≥ ₹88")
	assert.LessOrEqual(t, roundTrip, 95.0, "round-trip must be ≤ ₹95")
}

// TestCalcZerodhaFullCommission_BrokerageCap confirms that for large notionals
// brokerage is capped at ₹20 per leg.
func TestCalcZerodhaFullCommission_BrokerageCap(t *testing.T) {
	// At ₹100,000 notional: 0.03% = ₹30 → capped to ₹20.
	const bigNotional = 100_000.0
	got := calcZerodhaFullCommission(bigNotional, true)
	// Verify it does not include ₹30 brokerage (uncapped).
	// With cap: brokerage=20, STT=100, exchange=3.45, SEBI=0.10, stamp=15, GST=18%×(20+3.45)
	// GST = 0.18 × 23.45 = 4.221; total = 20+100+3.45+0.10+15+4.221 = 142.771
	// Without cap: brokerage=30, STT=100, exchange=3.45, SEBI=0.10, stamp=15, GST=18%×(30+3.45)
	// GST = 0.18 × 33.45 = 6.021; total = 30+100+3.45+0.10+15+6.021 = 154.571
	assert.InDelta(t, 142.771, got, 1e-2, "buy-leg at ₹100K should use capped ₹20 brokerage")
}

// TestCalcZerodhaFullCommission_StampDutyBuyOnly confirms stamp duty applies on
// buy but not sell, and the difference equals exactly 0.015% of notional.
func TestCalcZerodhaFullCommission_StampDutyBuyOnly(t *testing.T) {
	buy := calcZerodhaFullCommission(notional30K, true)
	sell := calcZerodhaFullCommission(notional30K, false)
	diff := buy - sell
	expectedStamp := notional30K * 0.00015 // 0.015%
	assert.InDelta(t, expectedStamp, diff, 1e-6, "buy-sell diff must equal stamp duty (0.015%)")
}
