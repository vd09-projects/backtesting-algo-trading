package engine

import (
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Portfolio tracks cash, open positions, and the completed trade log.
// Fills are at stated price with no slippage or commission — execution realism
// is added in TASK-0005.
type Portfolio struct {
	Cash      float64
	Positions map[string]model.Position // keyed by instrument
	Trades    []model.Trade
}

func newPortfolio(initialCash float64) *Portfolio {
	return &Portfolio{
		Cash:      initialCash,
		Positions: make(map[string]model.Position),
	}
}

// applySignal translates a signal into a portfolio action at the given fill price and time.
// Buy  → open a long position (skipped if one is already open, or if cash is insufficient)
// Sell → close the open long position (skipped if no position is open)
// Hold → no-op
func (p *Portfolio) applySignal(
	signal model.Signal,
	instrument string,
	fillPrice float64,
	fillTime time.Time,
	positionSizeFraction float64,
) error {
	switch signal {
	case model.SignalBuy:
		return p.openLong(instrument, fillPrice, fillTime, positionSizeFraction)
	case model.SignalSell:
		return p.closeLong(instrument, fillPrice, fillTime)
	case model.SignalHold:
		return nil
	default:
		return fmt.Errorf("portfolio: unknown signal %q", signal)
	}
}

func (p *Portfolio) openLong(instrument string, price float64, t time.Time, sizeFraction float64) error {
	if _, open := p.Positions[instrument]; open {
		// Already holding — do not pyramid. Silently skip.
		return nil
	}

	cost := p.Cash * sizeFraction
	if cost <= 0 || cost > p.Cash {
		// Insufficient cash — skip.
		return nil
	}

	quantity := cost / price
	p.Cash -= cost
	p.Positions[instrument] = model.Position{
		Instrument: instrument,
		Direction:  model.DirectionLong,
		Quantity:   quantity,
		EntryPrice: price,
	}
	// EntryTime is stored on the Trade when the position closes.
	// We stash it on a side-channel using the trade log convention:
	// keep a sentinel open trade with ExitTime zero until closed.
	p.Trades = append(p.Trades, model.Trade{
		Instrument: instrument,
		Direction:  model.DirectionLong,
		Quantity:   quantity,
		EntryPrice: price,
		EntryTime:  t,
	})
	return nil
}

func (p *Portfolio) closeLong(instrument string, price float64, t time.Time) error {
	pos, open := p.Positions[instrument]
	if !open {
		// No position to close — silently skip.
		return nil
	}

	// Find the open (sentinel) trade and complete it.
	for i := len(p.Trades) - 1; i >= 0; i-- {
		tr := &p.Trades[i]
		if tr.Instrument == instrument && tr.ExitTime.IsZero() {
			tr.ExitPrice = price
			tr.ExitTime = t
			// Clean fill: no slippage or commission. TASK-0005 adds those.
			tr.RealizedPnL = (price - tr.EntryPrice) * tr.Quantity
			break
		}
	}

	p.Cash += pos.Quantity * price
	delete(p.Positions, instrument)
	return nil
}

// openTrades returns trades that have been opened but not yet closed.
func (p *Portfolio) openTrades() []model.Trade {
	var open []model.Trade
	for _, tr := range p.Trades {
		if tr.ExitTime.IsZero() {
			open = append(open, tr)
		}
	}
	return open
}

// closedTrades returns only completed round-trip trades.
func (p *Portfolio) closedTrades() []model.Trade {
	var closed []model.Trade
	for _, tr := range p.Trades {
		if !tr.ExitTime.IsZero() {
			closed = append(closed, tr)
		}
	}
	return closed
}
