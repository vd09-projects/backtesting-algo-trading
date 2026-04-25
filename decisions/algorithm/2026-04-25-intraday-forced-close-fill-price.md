# Intraday forced-close fill price: 3:15 PM bar Close

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-25       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | intraday, session-boundary, fill-model, MIS, TASK-0046 |

## Context

The engine event loop has no concept of a trading session boundary. For intraday (MIS) strategies on NSE, Zerodha begins auto-squareoff of open positions between 3:15–3:20 PM IST. When implementing session-boundary logic in the engine (TASK-0046), the fill price used to close positions at session end needed to be defined before any code was written. The choice affects P&L realism: a price that is too favourable makes the backtest look better than live execution will be.

## Options considered

### Option A: 3:15 PM bar's Close (selected)
- **Pros**: Most conservative; Zerodha's auto-squareoff begins at 3:15 PM, so using that bar's close approximates the squareoff price without requiring knowledge of the bid/ask at that moment. Any live execution will be at this price or worse, meaning the backtest does not flatter performance.
- **Cons**: Not the exact squareoff price — Zerodha fills at a market price near the bid, which is typically slightly worse than the 3:15 PM close. However, this is a minor over-estimation of execution quality, not a material one for daily-bar-scale P&L.

### Option B: 3:30 PM bar's Close (last bar of the day)
- **Pros**: Uses the final close price of the session.
- **Cons**: The position was actually closed at 3:15-3:20 PM, not 3:30 PM. Using 3:30 PM can produce P&L differences of 0.2-0.5% on volatile close-of-session moves — this is not a conservative approximation.

### Option C: Model the Zerodha squareoff auction explicitly
- **Pros**: Most accurate.
- **Cons**: Requires modelling the bid/ask spread at 3:15 PM, which is not available in the Zerodha historical candle data. Over-engineering for the current stage.

## Decision

Use the **3:15 PM bar's Close** as the forced-close fill price for MIS intraday positions. This is the most conservative approximation available from the candle data: it understates execution quality rather than overstating it, which is the correct bias for a backtest. Zerodha's actual auto-squareoff fills at or slightly worse than bid, so real performance will be at this price or marginally below it — the backtest does not create an edge that live trading cannot replicate.

## Consequences

- Priya implements `SessionConfig.SessionCutoff` as 15:15 IST; any position open at or after a bar timestamped ≥ 15:15 IST is force-closed at that bar's Close price.
- All intraday backtests will use this convention consistently. No per-strategy override.
- If live fill quality data later shows systematic divergence from 3:15 PM Close, this decision should be revisited.

## Related decisions

- [Intraday session-close detection uses IST timestamp ≥ 15:15](./2026-04-25-intraday-session-detection-ist-cutoff.md) — companion decision specifying how the bar is identified

## Revisit trigger

If live fill quality data shows that 3:15 PM Close systematically over- or under-states actual Zerodha auto-squareoff fills by more than 0.3% on average, revisit to either adjust the cutoff time or add a bid/ask spread penalty.
