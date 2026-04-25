package engine

// NSE CNC (delivery) statutory charge rates.
// Sources: Zerodha support → "Charges and taxes for equity delivery"
// and SEBI circular SEBI/HO/MRD2/DCAP/CIR/P/2019/67 for exchange charges.
// All rates are fractions (not percentages) unless noted.
const (
	// nseZerodhaBrokerageRate is 0.03% of trade value.
	// Zerodha equity delivery brokerage: min(rate×notional, ₹20) per order.
	nseZerodhaBrokerageRate = 0.0003

	// nseZerodhaBrokerageCap is the maximum per-order brokerage Zerodha charges.
	nseZerodhaBrokerageCap = 20.0

	// nseSTTRate is Securities Transaction Tax for equity delivery.
	// 0.10% on both buy and sell legs.
	nseSTTRate = 0.001

	// nseExchangeChargesRate is the NSE transaction charge for equity delivery.
	// 0.00345% of trade value.
	nseExchangeChargesRate = 0.0000345

	// nseSEBIChargesRate is the SEBI turnover fee.
	// 0.0001% of trade value.
	nseSEBIChargesRate = 0.000001

	// nseStampDutyRate is stamp duty on buy-side equity delivery orders.
	// 0.015% of trade value. Applies to buy leg only.
	nseStampDutyRate = 0.00015

	// nseGSTRate is Goods and Services Tax on brokerage and exchange charges.
	// 18%. Applies to (brokerage + exchange charges) only;
	// STT, SEBI charges, and stamp duty are statutory levies exempt from GST.
	nseGSTRate = 0.18
)

// calcZerodhaFullCommission returns the full NSE CNC delivery cost for a single
// leg at the given trade value (fill price × quantity).
//
// Cost components:
//
//	brokerage      = min(0.03% × notional, ₹20)
//	STT            = 0.10% × notional                      (both sides)
//	exchange       = 0.00345% × notional
//	SEBI           = 0.0001% × notional
//	stamp duty     = 0.015% × notional                     (buy leg only)
//	GST            = 18% × (brokerage + exchange charges)
func calcZerodhaFullCommission(tradeValue float64, isBuy bool) float64 {
	brokerage := tradeValue * nseZerodhaBrokerageRate
	if brokerage > nseZerodhaBrokerageCap {
		brokerage = nseZerodhaBrokerageCap
	}

	stt := tradeValue * nseSTTRate
	exchangeCharges := tradeValue * nseExchangeChargesRate
	sebi := tradeValue * nseSEBIChargesRate

	var stamp float64
	if isBuy {
		stamp = tradeValue * nseStampDutyRate
	}

	gst := (brokerage + exchangeCharges) * nseGSTRate

	return brokerage + stt + exchangeCharges + sebi + stamp + gst
}
