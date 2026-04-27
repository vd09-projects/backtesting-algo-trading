package model

// CommissionModel determines how trading commission is calculated.
type CommissionModel string

// Supported commission models.
const (
	CommissionFlat           CommissionModel = "flat"             // fixed fee per trade
	CommissionPercentage     CommissionModel = "percentage"       // percentage of trade value
	CommissionZerodha        CommissionModel = "zerodha"          // min(₹20, 0.03% of trade value) per fill
	CommissionZerodhaFull    CommissionModel = "zerodha_full"     // full NSE CNC delivery cost stack (brokerage + STT + exchange + SEBI + stamp + GST)
	CommissionZerodhaFullMIS CommissionModel = "zerodha_full_mis" // full NSE MIS intraday cost stack: STT 0.025% on sell leg only; all other charges same as CommissionZerodhaFull
)

// OrderConfig holds execution cost parameters applied by the engine on every fill.
type OrderConfig struct {
	SlippagePct     float64 // slippage as a fraction, e.g. 0.0005 = 0.05%
	CommissionModel CommissionModel
	CommissionValue float64 // flat fee in currency units, or fraction if percentage model
}
