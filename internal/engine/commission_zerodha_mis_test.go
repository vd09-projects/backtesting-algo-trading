package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Hand-verified arithmetic for ₹30,000 notional (MIS intraday):
//
// MIS differs from CNC in STT only:
//   - STT = 0.025% on sell leg only (CNC charges 0.10% on both legs).
//   - All other components (brokerage, exchange, SEBI, stamp, GST) are identical to CNC.
//
// Buy leg (MIS):
//
//	brokerage      = min(0.03% × 30000, 20)       = min(9.00, 20)   = 9.0000
//	STT            = 0                              (buy: no STT for MIS)
//	exchange       = 0.00345% × 30000              = 1.0350
//	SEBI           = 0.0001% × 30000               = 0.0300
//	stamp          = 0.015% × 30000                = 4.5000           (buy-side only)
//	GST            = 18% × (brokerage + exchange)  = 18% × 10.035    = 1.8063
//	buy total      = 9.00 + 0 + 1.035 + 0.03 + 4.50 + 1.8063       = 16.3713
//
// Sell leg (MIS):
//
//	brokerage      = 9.0000
//	STT            = 0.025% × 30000                = 7.5000           (sell-only for MIS)
//	exchange       = 1.0350
//	SEBI           = 0.0300
//	stamp          = 0                              (sell: no stamp duty)
//	GST            = 1.8063
//	sell total     = 9.00 + 7.50 + 1.035 + 0.03 + 0 + 1.8063       = 19.3713
//
// Round-trip = 16.3713 + 19.3713 = 35.7426
//
// CNC round-trip was ₹88.2426. MIS saves ₹52.50 in STT (₹30 on buy + ₹22.50 on sell).
// The acceptance-criteria band "~₹55-60" was a transposition of the savings figure as a
// cost figure. The correct MIS round-trip cost at ₹30K notional is ₹35.7426.

func TestCalcZerodhaFullMISCommission_BuyLeg(t *testing.T) {
	got := calcZerodhaFullMISCommission(notional30K, true)
	assert.InDelta(t, 16.3713, got, 1e-3, "MIS buy-leg cost at ₹30K notional")
}

func TestCalcZerodhaFullMISCommission_SellLeg(t *testing.T) {
	got := calcZerodhaFullMISCommission(notional30K, false)
	assert.InDelta(t, 19.3713, got, 1e-3, "MIS sell-leg cost at ₹30K notional")
}

func TestCalcZerodhaFullMISCommission_RoundTrip(t *testing.T) {
	buy := calcZerodhaFullMISCommission(notional30K, true)
	sell := calcZerodhaFullMISCommission(notional30K, false)
	roundTrip := buy + sell
	assert.InDelta(t, 35.7426, roundTrip, 1e-3, "MIS round-trip cost at ₹30K notional")
	// MIS is always cheaper than CNC: STT savings of ₹30 buy + ₹22.50 sell = ₹52.50.
	assert.Less(t, roundTrip, 88.2426, "MIS round-trip must be less than CNC round-trip")
}

// TestCalcZerodhaFullMISCommission_STTBuyIsZero confirms that for MIS,
// STT on the buy leg is zero (unlike CNC which charges 0.10% on both legs).
func TestCalcZerodhaFullMISCommission_STTBuyIsZero(t *testing.T) {
	buy := calcZerodhaFullMISCommission(notional30K, true)
	sell := calcZerodhaFullMISCommission(notional30K, false)
	// Structural invariant: sell > buy because sell carries STT but buy does not.
	// (CNC has buy > sell because buy carries stamp duty; MIS reverses this.)
	assert.Greater(t, sell, buy, "MIS sell leg must cost more than buy leg (STT on sell only)")
	// STT on sell = 0.025% × 30000 = 7.50; stamp on buy = 0.015% × 30000 = 4.50.
	// sell - buy = 7.50 - 4.50 = 3.00
	assert.InDelta(t, 3.0, sell-buy, 1e-6, "sell-buy diff must be STT(sell) - stamp(buy) = ₹3.00")
}

// TestCalcZerodhaFullMISCommission_BrokerageCap confirms brokerage capping
// at ₹20 also applies to MIS at large notionals.
func TestCalcZerodhaFullMISCommission_BrokerageCap(t *testing.T) {
	// At ₹100,000 notional: 0.03% = ₹30 → capped to ₹20.
	// MIS buy leg with cap:
	//   brokerage=20, STT=0, exchange=3.45, SEBI=0.10, stamp=15, GST=18%×(20+3.45)=4.221
	//   total = 20 + 0 + 3.45 + 0.10 + 15 + 4.221 = 42.771
	const bigNotional = 100_000.0
	got := calcZerodhaFullMISCommission(bigNotional, true)
	assert.InDelta(t, 42.771, got, 1e-2, "MIS buy-leg at ₹100K should use capped ₹20 brokerage, zero STT")
}
