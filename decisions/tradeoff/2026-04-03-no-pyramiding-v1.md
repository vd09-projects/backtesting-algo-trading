# No pyramiding — single position per instrument enforced in v1

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-03       |
| Status   | accepted         |
| Category | tradeoff         |
| Tags     | portfolio, position-sizing, pyramiding, v1-scope, engine |

## Context

When implementing `Portfolio.openLong`, we needed to decide what happens when a Buy signal arrives while a long position in the same instrument is already open. Some strategies intentionally scale into winning positions (pyramid). Others should only hold one position at a time. The engine has no way to distinguish intent from the signal alone.

## Options considered

### Option A: Silently skip (chosen for v1)
If a position is already open for the instrument, ignore the Buy signal. One position per instrument at all times.

- **Pros**: Simple. Safe default — prevents unintended over-exposure. Easy to reason about in tests.
- **Cons**: Strategies designed to scale in (add to a winner) will silently produce wrong results. The skip is invisible in the trade log.

### Option B: Allow pyramiding — add to the existing position
Merge the new order into the open position, updating quantity and average entry price.

- **Pros**: Supports a wider class of strategies correctly.
- **Cons**: Averaging-in logic adds complexity to P&L accounting (weighted average entry price). Deferred to when there's a concrete strategy that requires it.

### Option C: Reject with an error
Return an error when a Buy arrives on an open position, forcing the strategy to emit Hold when already long.

- **Pros**: Explicit — bugs surface immediately.
- **Cons**: Breaks any strategy that emits Buy continuously (e.g., "hold while indicator is bullish"). Too strict for a general-purpose engine.

## Decision

**Option A** for v1 — silent skip.

The primary goal right now is a working engine for single-entry, single-exit strategies (e.g., simple moving average crossovers). Pyramiding can be added when there's a strategy that requires it. The silent-skip behaviour is documented here so it doesn't get confused with a bug.

## Consequences

- Strategies that emit Buy on every bar while bullish work correctly — they open once and hold.
- Strategies that explicitly try to scale in will produce incorrect results silently. **Anyone writing such a strategy must know this constraint exists.**
- The trade log does not record skipped buy signals — they leave no trace.

## Revisit trigger

Add pyramiding support when a concrete strategy requires scale-in behaviour, or when position sizing becomes more sophisticated than a single `sizeFraction` parameter.
