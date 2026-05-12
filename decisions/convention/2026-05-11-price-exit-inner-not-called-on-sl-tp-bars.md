# PriceExit does not call inner.Next() on bars where SL or TP fires

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-11       |
| Status   | experimental     |
| Category | convention       |
| Tags     | pkg/strategy, PriceExit, wrapper, stop-loss, target-profit, inner-delegation, stateful, TASK-0098 |

## Context

`PriceExit.Next()` must decide: when SL or TP fires, should it still call `inner.Next()` before returning `SignalSell`? The acceptance criteria state "SL and TP checked before delegating to inner — price-based exits take priority."

Two behaviours are possible when a price exit triggers:

1. **Call inner, then override**: call `inner.Next()` to get its signal, then ignore it and return `SignalSell` because the price exit fired.
2. **Skip inner entirely**: return `SignalSell` directly without calling `inner.Next()`.

## Options considered

### Option A: Call inner.Next(), then override its result (rejected)

- **Pros**: Inner's internal state always advances one call per bar, regardless of whether PriceExit's exit fires. Stateful inner strategies (e.g. a nested `TimedExit`) remain perfectly in sync with the bar index.
- **Cons**: Calling `inner.Next()` when we're about to discard its result is misleading. If inner is itself stateful, a BUY signal from inner on the same bar as a PriceExit SL would silently update inner's `entryBar` and `inPosition` — corrupting inner's state in a way that's invisible to PriceExit. This is worse than the desync described in Option B.

### Option B: Skip inner.Next() entirely on SL/TP bars (chosen)

- **Pros**: Clean separation — price exits take unambiguous priority. No spurious state mutations in inner on exit bars. The AC's "price-based exits take priority" is expressed directly in the code structure.
- **Cons**: For stateful inner strategies, inner's call counter is one behind after each SL/TP-triggered exit. This is the "inner-state desync" described in the type godoc.

## Decision

Skip `inner.Next()` on bars where SL or TP fires. Return `SignalSell` and reset PriceExit state directly, without calling into inner.

The inner-state desync is real but harmless in practice for two reasons:

1. **Engine no-pyramiding rule**: PriceExit gates all BUY/SELL decisions. Once PriceExit exits, `inPosition` is false. On the next bar, PriceExit calls `inner.Next()` again — any signal inner returns at that point is treated as a fresh out-of-position signal. Inner's internal "I'm still in position" state may be stale, but the engine won't re-enter unless inner emits BUY and PriceExit is also not in position.

2. **Factory-per-fold guarantee**: The walk-forward harness constructs a fresh strategy factory per fold. Any state desync accumulated within a sequential backtest run is scoped to that run and doesn't bleed across folds.

The desync only becomes harmful if: (a) PriceExit exits, (b) inner thinks it's still in position, (c) inner returns Hold on the next bar when it should return a new entry signal. The practical consequence is a missed entry signal at most — not a phantom position or corrupted P&L.

This behavior is documented in the `PriceExit` type godoc under "Inner-state desync."

## Consequences

- Inner.Next() is not called on SL/TP-exit bars. Inner's call counter lags by one per PriceExit-triggered exit.
- For stateful inner strategies, internal position state may be one bar behind after a PriceExit exit. This is harmless under the no-pyramiding rule.
- Test design consequence: in `price_exit_test.go`, the `scriptedStrategy` call counter does not advance on SL/TP bars. Test scripts must account for this — see `TestPriceExit_ReEntryAfterSLReset` where script[1] is Buy (not Hold) because bar 1 was a skipped SL bar.

## Related decisions

- [TimedExit statefulness — first stateful wrapper in pkg/strategy](./2026-04-27-timed-exit-statefulness-pkg-strategy.md) — PriceExit follows the same stateful wrapper pattern; inner-state desync is a similar concern.
- [Walk-forward Run() accepts a factory, not a single strategy instance](../architecture/2026-05-07-walkforward-strategy-factory-per-fold.md) — factory-per-fold is the structural guarantee that bounds the inner-state desync to a single run.

## Revisit trigger

If a composed strategy (e.g. `NewPriceExit(NewTimedExit(inner, N))`) is observed to miss entry signals after SL/TP exits in a production backtest, re-evaluate Option A. The fix would be to call `inner.Next()` on SL/TP bars but not return its result — advancing state without using the signal.
