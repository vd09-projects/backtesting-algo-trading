# Trade.RealizedPnL stored on the struct, not computed on-demand

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-02       |
| Status   | accepted         |
| Category | convention       |
| Tags     | trade, pnl, analytics, engine, commission, slippage |

## Context

When designing the `Trade` type, we needed to decide where P&L computation lives. `RealizedPnL` depends on entry price, exit price, quantity, slippage, and commission — all of which the engine knows at fill time. The analytics package reads completed trades to produce reports.

## Options considered

### Option A: Store RealizedPnL as a field on Trade (chosen)
The engine computes P&L at close time, accounting for slippage and commission, and stores the result in `Trade.RealizedPnL`. Analytics reads it directly.

- **Pros**: Analytics is a pure reader — it never needs to know about commission models or slippage rates. P&L computation lives in exactly one place (the engine). Impossible for analytics and engine to disagree on the number.
- **Cons**: `Trade` carries a derived value alongside its raw inputs. Readers must trust that the stored value was computed correctly.

### Option B: Compute P&L on-demand in analytics
`Trade` stores only raw fields (entry/exit price, quantity). Analytics computes P&L when producing the report by applying commission and slippage logic.

- **Pros**: `Trade` is a pure data bag with no derived values.
- **Cons**: Analytics must accept an `OrderConfig` parameter (or equivalent) to know the commission model — coupling it to execution concerns. If the commission model ever changes, analytics must be updated too. Two places can produce different P&L numbers for the same trade.

## Decision

**Option A** — store `RealizedPnL` on `Trade`.

The engine is the only component that knows the full execution context (slippage, commission model, fill price). It should own P&L calculation. Analytics is a reporting layer — it should read results, not reimplement execution logic. This keeps analytics decoupled from `OrderConfig` entirely.

## Consequences

- Every `Trade` produced by the engine must have `RealizedPnL` populated before being appended to the trade log.
- Analytics can sum, aggregate, and analyze P&L without importing or knowing about `OrderConfig`.
- If commission or slippage logic changes, only the engine needs updating — analytics is unaffected.

## Revisit trigger

If we add tax-adjusted P&L, currency conversion, or other post-execution adjustments that analytics needs to apply independently, reconsider whether a single stored field is sufficient or whether a breakdown (gross P&L, commission paid, slippage cost) is needed.
