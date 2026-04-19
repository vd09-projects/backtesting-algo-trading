package analytics

import "math"

// DSR computes the Deflated Sharpe Ratio: observedSharpe corrected for the
// expected maximum Sharpe arising from testing nTrials independent strategies.
//
// Formula from Bailey & López de Prado (2014). E[max SR] is the expected
// maximum Sharpe from nTrials iid unit-normal draws, computed via the
// Euler–Mascheroni correction:
//
//	E[max SR] ≈ (1−γ)·Φ⁻¹(1−1/n) + γ·Φ⁻¹(1−1/(n·e))
//
// The returned value is observedSharpe minus E[max SR] scaled by the Sharpe
// standard error SE = 1/√(nObservations−1). A positive result means the
// strategy's Sharpe exceeds what multiple testing alone would predict.
// Higher nTrials → higher E[max SR] → lower (more conservative) DSR.
//
// Returns observedSharpe unchanged when nTrials ≤ 1 or nObservations ≤ 1.
func DSR(observedSharpe, nTrials, nObservations float64) float64 {
	if nTrials <= 1 || nObservations <= 1 {
		return observedSharpe
	}
	const eulerMascheroni = 0.5772156649015328
	eMaxSR := (1-eulerMascheroni)*normInvCDF(1-1/nTrials) +
		eulerMascheroni*normInvCDF(1-1/(nTrials*math.E))
	se := 1 / math.Sqrt(nObservations-1)
	return observedSharpe - eMaxSR*se
}

// normInvCDF returns Φ⁻¹(p), the inverse standard normal CDF.
// Uses the rational approximation from Acklam (2002); max absolute error ~1.15e-9.
// Returns ±Inf for p ≤ 0 or p ≥ 1.
func normInvCDF(p float64) float64 {
	const (
		a1 = -3.969683028665376e+01
		a2 = 2.209460984245205e+02
		a3 = -2.759285104469687e+02
		a4 = 1.383577518672690e+02
		a5 = -3.066479806614716e+01
		a6 = 2.506628277459239e+00

		b1 = -5.447609879822406e+01
		b2 = 1.615858368580409e+02
		b3 = -1.556989798598866e+02
		b4 = 6.680131188771972e+01
		b5 = -1.328068155288572e+01

		c1 = -7.784894002430293e-03
		c2 = -3.223964580411365e-01
		c3 = -2.400758277161838e+00
		c4 = -2.549732539343734e+00
		c5 = 4.374664141464968e+00
		c6 = 2.938163982698783e+00

		d1 = 7.784695709041462e-03
		d2 = 3.224671290700398e-01
		d3 = 2.445134137142996e+00
		d4 = 3.754408661907416e+00

		pLow  = 0.02425
		pHigh = 1 - pLow
	)

	switch {
	case p <= 0:
		return math.Inf(-1)
	case p >= 1:
		return math.Inf(1)
	case p < pLow:
		q := math.Sqrt(-2 * math.Log(p))
		return (((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6) /
			((((d1*q+d2)*q+d3)*q+d4)*q + 1)
	case p <= pHigh:
		q := p - 0.5
		r := q * q
		return (((((a1*r+a2)*r+a3)*r+a4)*r+a5)*r + a6) * q /
			(((((b1*r+b2)*r+b3)*r+b4)*r+b5)*r + 1)
	default:
		q := math.Sqrt(-2 * math.Log(1-p))
		return -(((((c1*q+c2)*q+c3)*q+c4)*q+c5)*q + c6) /
			((((d1*q+d2)*q+d3)*q+d4)*q + 1)
	}
}
