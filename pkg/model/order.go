package model

// CommissionModel determines how trading commission is calculated.
type CommissionModel string

// Supported commission models.
const (
	CommissionFlat       CommissionModel = "flat"       // fixed fee per trade
	CommissionPercentage CommissionModel = "percentage" // percentage of trade value
	CommissionZerodha    CommissionModel = "zerodha"    // min(₹20, 0.03% of trade value) per fill
)

// OrderConfig holds execution cost parameters applied by the engine on every fill.
type OrderConfig struct {
	SlippagePct     float64 // slippage as a fraction, e.g. 0.0005 = 0.05%
	CommissionModel CommissionModel
	CommissionValue float64 // flat fee in currency units, or fraction if percentage model
}
