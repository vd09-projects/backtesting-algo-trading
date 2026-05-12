# PriceExit zero-guard condition uses `> 0` not `!= 0`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-11       |
| Status   | experimental     |
| Category | convention       |
| Tags     | pkg/strategy, PriceExit, wrapper, stop-loss, target-profit, guard, zero-value, TASK-0098 |

## Context

`NewPriceExit` accepts `stopLossPct` and `targetProfitPct` as `float64` fields. A zero value is the documented way to disable each guard independently. The implementation in `Next()` must check whether each guard is active before evaluating the threshold. The question is: use `> 0` or `!= 0`?

## Options considered

### Option A: Use `!= 0` (rejected)

Check whether the value is non-zero before applying the guard.

- **Pros**: Technically matches "disabled when zero" in the narrowest sense.
- **Cons**: A negative value (e.g. `-0.05`) would pass the `!= 0` check and produce a threshold of `entryPrice * (1 - (-0.05)) = entryPrice * 1.05` — an *upward* stop-loss that fires when price *rises* 5%. This is nonsensical and silently dangerous.

### Option B: Use `> 0` (chosen)

Only activate the guard when the percentage is strictly positive.

- **Pros**: Negative inputs are silently treated as disabled rather than inverting the guard direction. Matches natural semantics: "disabled when zero or absent, active when positive."
- **Cons**: Negative input is silently treated as zero rather than returned as an error. Callers who pass `-0.05` thinking it means "5% stop" get no stop-loss protection — surprising but not dangerous, and unlikely given how the type is used.

## Decision

Use `> 0` for both guards. `p.stopLossPct > 0` and `p.targetProfitPct > 0` in `PriceExit.Next()`.

The reasoning: `PriceExit` is an internal composable wrapper used by strategy authors, not a public API receiving untrusted input. Negative percentages have no sensible interpretation in this domain (a stop-loss should trigger when price falls, not rises). The `> 0` guard makes both "disabled" (zero) and "nonsensical" (negative) inputs do the same thing — nothing — without requiring an error path in a function that has no error return by interface contract.

The `NewPriceExit` godoc says "Set to 0 to disable" but does not document negative-value behavior. TASK-0105 tracks a one-line addition to note that negative values are treated as disabled.

## Consequences

- Negative `stopLossPct` or `targetProfitPct` silently disables the corresponding guard. A caller passing `-0.05` gets no stop-loss with no error.
- This is acceptable for an internal wrapper type. Revisit if `PriceExit` is ever exposed in a public API or wired to user-facing config parsing.

## Related decisions

- [TimedExit statefulness — first stateful wrapper in pkg/strategy](./2026-04-27-timed-exit-statefulness-pkg-strategy.md) — PriceExit is the second stateful Strategy wrapper, following the same design pattern.
- [MACD guard condition: n <= slow+signal-1](./2026-04-26-macd-guard-talib-initialization-boundary.md) — thematic sibling: guard conditions in strategy Next() functions preventing nonsensical trigger states.

## Revisit trigger

If `PriceExit` is ever exposed to user-facing config parsing or a public API boundary — add input validation to `NewPriceExit` returning an error for negative values.
