# SignalHold filtered at engine level — never passed to portfolio

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | convention |
| Tags     | engine, portfolio, signal, hold, applySignal, convention, coverage |

## Context

The engine's hot loop processes a `pendingSignal` from the prior bar. The question arose: should `SignalHold` be passed through to `portfolio.applySignal` (which has an explicit `case model.SignalHold: return nil` branch), or should the engine filter it out before the call?

## Options considered

### Option A: Engine filters Hold before calling applySignal
- **Pros**: Avoids an unnecessary function call in the hot loop. The `if pendingSignal != model.SignalHold` guard makes the intent explicit at the call site — "we only apply signals that require action". The `applySignal` no-op branch still exists as a defensive guard for unexpected values, but it is not the primary path.
- **Cons**: The `SignalHold` branch in `applySignal` becomes unreachable from the engine (only reachable via direct whitebox calls), which shows as a coverage gap unless explicitly tested.

### Option B: Always pass pendingSignal to applySignal (including Hold)
- **Pros**: Uniform code path — one place handles all signals. `applySignal` coverage is complete naturally.
- **Cons**: Unnecessary function call overhead on Hold (the most common signal in any strategy). Minor, but the engine is a hot loop processing thousands of bars.

## Decision

The engine guards with `if pendingSignal != model.SignalHold` before calling `applySignal`. Hold is filtered at the call site — the portfolio layer is not involved. `applySignal` retains its `case model.SignalHold: return nil` branch as a defensive guard (in case it is ever called directly or a new code path bypasses the engine guard), and that branch is covered by a whitebox test `TestApplySignal_HoldIsNoOp`.

## Consequences

- `applySignal`'s Hold branch is only reachable via direct whitebox test, not through the engine. This is intentional and documented.
- The whitebox test `TestApplySignal_HoldIsNoOp` was added specifically to cover this branch, confirming the guard works correctly when called directly.
- Future code paths that call `applySignal` directly (e.g. in tests or alternate engine implementations) will behave correctly for Hold without relying on the caller to pre-filter.

## Related decisions

- [No pyramiding — single position per instrument enforced in v1](../tradeoff/2026-04-03-no-pyramiding-v1.md) — the same engine loop also silently skips redundant Buy signals via portfolio-level guard
