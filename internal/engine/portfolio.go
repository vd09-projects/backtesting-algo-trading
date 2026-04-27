package engine

import (
	"fmt"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// Portfolio tracks cash, open positions, and the completed trade log.
type Portfolio struct {
	Cash        float64
	Positions   map[string]model.Position // keyed by instrument
	Trades      []model.Trade
	orderConfig model.OrderConfig
	equityCurve []model.EquityPoint
}

// newPortfolio creates a Portfolio with the given initial cash and order config.
// capacity pre-allocates the equity curve slice; pass 0 if unknown.
func newPortfolio(initialCash float64, cfg model.OrderConfig, capacity int) *Portfolio {
	return &Portfolio{
		Cash:        initialCash,
		Positions:   make(map[string]model.Position),
		orderConfig: cfg,
		equityCurve: make([]model.EquityPoint, 0, capacity),
	}
}

// RecordEquity snapshots total portfolio value at candle.Close.
// Value = cash + mark-to-market value of the open position in candle.Instrument (if any).
// Call once per bar after applying any pending signal.
func (p *Portfolio) RecordEquity(candle model.Candle) { //nolint:gocritic // Candle is a value type at API boundaries; pointer would leak internals
	value := p.Cash
	if pos, ok := p.Positions[candle.Instrument]; ok {
		value += pos.Quantity * candle.Close
	}
	p.equityCurve = append(p.equityCurve, model.EquityPoint{
		Timestamp: candle.Timestamp,
		Value:     value,
	})
}

// EquityCurve returns the equity time series recorded by RecordEquity.
func (p *Portfolio) EquityCurve() []model.EquityPoint {
	return p.equityCurve
}

// calcFillPrice applies slippage to a base price.
// Buy fills cost more (price * (1 + slippagePct)); sells receive less (price * (1 - slippagePct)).
func (p *Portfolio) calcFillPrice(basePrice float64, isBuy bool) float64 {
	if p.orderConfig.SlippagePct == 0 {
		return basePrice
	}
	if isBuy {
		return basePrice * (1 + p.orderConfig.SlippagePct)
	}
	return basePrice * (1 - p.orderConfig.SlippagePct)
}

// calcCommission returns the commission cost for a fill of the given trade value (fillPrice × qty).
// isBuy distinguishes buy fills from sell fills; most models ignore this, but
// CommissionZerodhaFull uses it to gate stamp duty to the buy leg only.
func (p *Portfolio) calcCommission(tradeValue float64, isBuy bool) float64 {
	switch p.orderConfig.CommissionModel {
	case model.CommissionFlat:
		return p.orderConfig.CommissionValue
	case model.CommissionPercentage:
		return tradeValue * p.orderConfig.CommissionValue
	case model.CommissionZerodha:
		c := tradeValue * 0.0003 // 0.03%
		if c > 20 {
			return 20
		}
		return c
	case model.CommissionZerodhaFull:
		return calcZerodhaFullCommission(tradeValue, isBuy)
	case model.CommissionZerodhaFullMIS:
		return calcZerodhaFullMISCommission(tradeValue, isBuy)
	default:
		return 0
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

	fillPrice := p.calcFillPrice(price, true)
	cost := p.Cash * sizeFraction
	if cost <= 0 {
		// Zero or negative budget — skip.
		return nil
	}

	quantity := cost / fillPrice
	entryCommission := p.calcCommission(quantity*fillPrice, true)
	totalCost := quantity*fillPrice + entryCommission
	if totalCost > p.Cash {
		// Commission pushes total over available cash — skip.
		return nil
	}

	p.Cash -= totalCost
	p.Positions[instrument] = model.Position{
		Instrument: instrument,
		Direction:  model.DirectionLong,
		Quantity:   quantity,
		EntryPrice: fillPrice,
	}
	// Sentinel open trade; completed in closeLong.
	p.Trades = append(p.Trades, model.Trade{
		Instrument: instrument,
		Direction:  model.DirectionLong,
		Quantity:   quantity,
		EntryPrice: fillPrice,
		EntryTime:  t,
		Commission: entryCommission,
	})
	return nil
}

func (p *Portfolio) closeLong(instrument string, price float64, t time.Time) error {
	pos, open := p.Positions[instrument]
	if !open {
		// No position to close — silently skip.
		return nil
	}

	fillPrice := p.calcFillPrice(price, false)
	exitCommission := p.calcCommission(fillPrice*pos.Quantity, false)

	// Find the open sentinel trade and complete it.
	for i := len(p.Trades) - 1; i >= 0; i-- {
		tr := &p.Trades[i]
		if tr.Instrument == instrument && tr.ExitTime.IsZero() {
			tr.ExitPrice = fillPrice
			tr.ExitTime = t
			tr.Commission += exitCommission
			tr.RealizedPnL = (fillPrice-tr.EntryPrice)*tr.Quantity - tr.Commission
			break
		}
	}

	p.Cash += fillPrice*pos.Quantity - exitCommission
	delete(p.Positions, instrument)
	return nil
}

// OpenTrades returns trades that have been opened but not yet closed.
func (p *Portfolio) OpenTrades() []model.Trade {
	var open []model.Trade
	for _, tr := range p.Trades {
		if tr.ExitTime.IsZero() {
			open = append(open, tr)
		}
	}
	return open
}

// ClosedTrades returns only completed round-trip trades.
func (p *Portfolio) ClosedTrades() []model.Trade {
	var closed []model.Trade
	for _, tr := range p.Trades {
		if !tr.ExitTime.IsZero() {
			closed = append(closed, tr)
		}
	}
	return closed
}
