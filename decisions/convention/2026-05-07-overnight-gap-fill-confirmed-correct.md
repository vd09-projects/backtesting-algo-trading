# Overnight gap fill: engine is gap-transparent by construction

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | accepted         |
| Category | convention       |
| Tags     | overnight-gap, fill-model, intraday, CNC, pending-signal, TASK-0071 |

## Context

TASK-0071 required verifying whether the backtesting engine correctly handles
overnight gaps for CNC intraday positions — specifically, whether a position
entered on day 1 and exited on day 2 will fill at the actual gap-open price,
or whether the engine inadvertently smooths or clamps the fill to a pre-gap
price.

This was a verification task, not a bug fix. No gap-specific logic existed in
the engine, and the question was whether the existing fill model was already
correct or needed an explicit guard.

## Finding

The engine is **gap-transparent by construction**. The fill model uses a
`pendingSignal` variable that carries a signal from bar N to bar N+1. When
bar N+1 is processed, the fill is applied at `candles[i].Open` with no
clamping, smoothing, interpolation, or session-boundary check. The engine has
no concept of "the previous bar's session" — it simply applies the pending fill
to whatever Open the provider returns.

Consequence: if day 2 opens 3% below day 1 close due to an overnight gap, the
exit fill is at day 2's Open (the gapped price). This is correct behaviour for
CNC strategies, which hold positions overnight and are exposed to gap risk.

Two golden tests were written and confirmed this:

- **TestGapDown_PositionEntered_PnLReflectsGap** — 5-min IST candle series
  spanning two sessions with a 3% gap-down open on day 2; position entered at
  day 1 bar 1 Open (505), exited at day 2 bar 4 Open (494); P&L confirmed as
  (494-505) * qty < 0.

- **TestGapDown_ExitFillsAtNextBarOpen_NotSignalBarClose** — Sell signal at
  bar N whose close is 600; bar N+1 opens at 582 (gap-down); fill confirmed at
  582, not 600.

Both tests pass without any engine change, confirming the existing
implementation is correct.

## Decision

No code changes required. The existing `pendingSignal → candles[i].Open` fill
path is the correct and canonical behaviour for gap handling. This decision is
recorded to prevent future "optimisations" from accidentally smoothing or
clamping the next-bar Open, which would break CNC gap exposure.

## Consequences

- CNC strategies correctly reflect overnight gap risk in their P&L.
- MIS strategies are covered by the separate forced-close convention
  (2026-04-25-intraday-forced-close-fill-price.md) — they are closed at the
  3:15 PM bar's Close, so they do not carry overnight exposure and are not
  affected by day 2 gap opens.
- Any future fill model changes must preserve the invariant: fill price equals
  `candles[i].Open` after slippage, with no session-aware clamping for CNC.
- The two golden tests in `internal/engine/engine_gap_test.go` serve as
  regression guards for this behaviour.

## Related decisions

- [Intraday forced-close fill price: 3:15 PM bar Close](../algorithm/2026-04-25-intraday-forced-close-fill-price.md) — MIS path; does not affect CNC.
- [RealizedPnL stored at close time by engine](./2026-04-02-trade-pnl-stored-not-computed.md) — P&L accounting convention.
