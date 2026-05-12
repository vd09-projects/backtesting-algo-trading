package strategy

import (
	"fmt"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// PriceExit wraps an inner Strategy and forces a sell when the current bar's
// Close hits a stop-loss or target-profit threshold relative to the entry price.
// It is useful for intraday strategies (e.g. ORB, Gap-and-Go) where price-based
// exits should take priority over time-based or signal-based exits from the inner
// strategy.
//
// Signal priority on each call to Next:
//  1. If not in position: delegate to inner. On SignalBuy, record
//     entryPrice = bar.Close and enter position.
//  2. If in position and stopLossPct > 0: if bar.Close <= entryPrice*(1-stopLossPct),
//     emit SignalSell and reset state. inner.Next is NOT called.
//  3. If in position and targetProfitPct > 0: if bar.Close >= entryPrice*(1+targetProfitPct),
//     emit SignalSell and reset state. inner.Next is NOT called.
//  4. If in position and neither guard fires: delegate to inner. An inner SignalSell
//     passes through and resets state.
//
// Note: SL and TP checks do NOT apply on the entry bar itself — inPosition is
// set after the BUY is returned, so the first SL/TP evaluation happens on the
// following bar.
//
// # Guard semantics
//
// A zero value for stopLossPct or targetProfitPct disables that guard entirely.
// Percentages are decimal fractions: 0.05 = 5%, not 5.
//
// # Inner-state desync
//
// When PriceExit triggers a SL or TP exit, inner.Next is not called on that bar.
// If inner is itself stateful (e.g. a nested TimedExit), inner's internal position
// tracking will be one bar behind after a PriceExit-triggered exit. This is harmless
// under the engine's no-pyramiding rule: PriceExit gates all BUY/SELL decisions,
// and inner's state resets on the next factory construction per fold.
//
// # Statefulness
//
// PriceExit maintains mutable state (entryPrice, inPosition) between calls.
// It is NOT safe for concurrent use across walk-forward folds or goroutines.
// Callers that run folds in parallel must construct a fresh PriceExit per fold
// using a factory: func() strategy.Strategy { return NewPriceExit(inner, sl, tp) }.
type PriceExit struct {
	inner           Strategy
	stopLossPct     float64
	targetProfitPct float64
	entryPrice      float64
	inPosition      bool
}

// NewPriceExit constructs a PriceExit that wraps inner and exits on a fixed
// stop-loss or target-profit threshold.
//
// stopLossPct is a decimal fraction: 0.05 exits when Close falls 5% below the
// entry price (i.e. Close <= entryPrice * 0.95). Set to 0 to disable.
//
// targetProfitPct is a decimal fraction: 0.10 exits when Close rises 10% above
// the entry price (i.e. Close >= entryPrice * 1.10). Set to 0 to disable.
//
// Compose with TimedExit for combined price and time exits:
//
//	NewPriceExit(NewTimedExit(inner, maxHoldBars), slPct, tpPct)
//
// Price exits fire first; the time-stop in TimedExit is the fallback.
func NewPriceExit(inner Strategy, stopLossPct, targetProfitPct float64) Strategy {
	return &PriceExit{
		inner:           inner,
		stopLossPct:     stopLossPct,
		targetProfitPct: targetProfitPct,
	}
}

// Name returns "price-exit(<inner.Name()>)".
func (p *PriceExit) Name() string {
	return fmt.Sprintf("price-exit(%s)", p.inner.Name())
}

// Lookback delegates to the inner strategy's Lookback.
func (p *PriceExit) Lookback() int { return p.inner.Lookback() }

// Timeframe delegates to the inner strategy's Timeframe.
func (p *PriceExit) Timeframe() model.Timeframe { return p.inner.Timeframe() }

// Next returns the signal for the current bar.
//
// When not in position, delegates to inner. A SignalBuy from inner enters the
// position and records entryPrice = bar.Close.
//
// When in position, SL and TP thresholds are evaluated first (price-based exits
// take priority). If either fires, SignalSell is returned and inner.Next is not
// called. If neither fires, inner.Next is called; an inner SignalSell passes
// through and resets state.
func (p *PriceExit) Next(candles []model.Candle) model.Signal {
	bar := candles[len(candles)-1]

	if !p.inPosition {
		innerSig := p.inner.Next(candles)
		if innerSig == model.SignalBuy {
			p.inPosition = true
			p.entryPrice = bar.Close
		}
		return innerSig
	}

	// In position: check SL before delegating to inner.
	if p.stopLossPct > 0 && bar.Close <= p.entryPrice*(1-p.stopLossPct) {
		p.inPosition = false
		p.entryPrice = 0
		return model.SignalSell
	}

	// In position: check TP before delegating to inner.
	if p.targetProfitPct > 0 && bar.Close >= p.entryPrice*(1+p.targetProfitPct) {
		p.inPosition = false
		p.entryPrice = 0
		return model.SignalSell
	}

	// Neither SL nor TP fired: delegate to inner.
	innerSig := p.inner.Next(candles)
	if innerSig == model.SignalSell {
		p.inPosition = false
		p.entryPrice = 0
	}
	return innerSig
}
