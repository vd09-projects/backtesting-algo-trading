# SMA crossover Lookback() returns slowPeriod, crossover guard handles first bar

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | sma-crossover, lookback, guard, slowPeriod, strategy, interface, TASK-0012 |

## Context

Strict crossover detection (see related decision) requires comparing the current bar's SMAs against the previous bar's SMAs. The previous bar's slow SMA is not a valid computed value until `len(candles) >= slowPeriod + 1`. `Lookback()` controls when the engine first calls `Next` — should it return `slowPeriod` or `slowPeriod + 1`?

## Options considered

### Option A: Lookback returns slowPeriod + 1
- **Pros**: Engine guarantees enough history for crossover detection on every `Next` call. No internal guard needed.
- **Cons**: Violates the TASK-0012 acceptance criterion ("Lookback returns `slowPeriod` so the engine starts feeding at the right bar"). Delays the first signal by one extra bar unnecessarily — the strategy can still return `Hold` on the first bar and handle it correctly internally.

### Option B: Lookback returns slowPeriod, internal guard handles first bar
- **Pros**: Matches acceptance criterion. Lookback semantics remain "minimum bars for a valid slow SMA," consistent with how every other strategy in this engine will define lookback. The one-extra-bar delay for crossover detection is an implementation detail, not a contract.
- **Cons**: Requires an explicit guard: `if n <= slowPeriod { return SignalHold }`. One extra conditional in `Next`.

## Decision

`Lookback()` returns `slowPeriod`. The guard `if n <= slowPeriod { return SignalHold }` in `Next` handles the single bar where the previous slow SMA is not yet valid. This keeps the lookback contract consistent across all strategies: "I need at least this many bars for my primary indicator to have a valid value."

## Consequences

On the bar where `len(candles) == slowPeriod` (the first bar the engine calls `Next`), the guard fires and returns `SignalHold`. Crossover detection is active from `len(candles) == slowPeriod + 1` onward. This is a silent, correct behaviour — no Buy or Sell is ever emitted on insufficient history.

Any future strategy using crossover logic should apply the same pattern: declare `Lookback()` at the primary indicator's lookback, add a `+1` guard internally.

## Related decisions

- [SMA crossover: strict crossover detection, not level comparison](2026-04-13-sma-crossover-strict-crossover-vs-level-comparison.md) — the crossover approach that necessitates this guard
- [Strategy.Lookback() as a first-class interface method](../architecture/2026-04-02-strategy-lookback-as-interface-method.md) — the lookback contract this decision respects
