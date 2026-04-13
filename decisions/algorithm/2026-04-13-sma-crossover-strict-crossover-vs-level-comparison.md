# SMA crossover: strict crossover detection, not level comparison

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | sma-crossover, signal-semantics, crossover, level-comparison, strategy, Next, BarResult, TASK-0012 |

## Context

When implementing the SMA crossover strategy, two approaches to signal generation were evaluated. The engine already enforces no-pyramiding, so both approaches produce identical trade entries and exits. The question was purely about what `Signal` value to emit on non-event bars.

**Level comparison**: emit `SignalBuy` on every bar where fast SMA > slow SMA; `SignalSell` on every bar where fast SMA < slow SMA.

**Strict crossover detection**: emit `SignalBuy` only on the bar where fast SMA transitions from ≤ slow SMA to > slow SMA; `SignalSell` only on the transition bar in the other direction; `SignalHold` on all other bars.

## Options considered

### Option A: Level comparison
- **Pros**: Simpler — no need to compare previous bar; a single `if fast > slow` is the entire logic.
- **Cons**: Emits `SignalBuy` on every bar of a sustained uptrend (potentially 30–50 consecutive bars). `BarResult.Signal` becomes a regime label, not an action. Anyone reading the bar log sees dozens of consecutive Buy signals and can't distinguish "entered here" from "still in trend." If the engine ever gains pyramiding or partial sizing, level comparison would silently change execution semantics.

### Option B: Strict crossover detection
- **Pros**: `Signal` means "act on this bar" — one Buy, followed by Holds until the crossover reverses. BarResult logs are clean and diagnostic. Semantics are preserved regardless of future engine features.
- **Cons**: Requires comparing the previous bar's SMAs, which in turn requires `len(candles) > slowPeriod` (one bar beyond what Lookback() guarantees). Needs an internal guard.

## Decision

Strict crossover detection. The behavioral equivalence under no-pyramiding is real, but it argues for whichever approach is more semantically correct — and that's strict crossover. A Signal should mean "I have new information that says act," not "the regime is currently bullish." The diagnostic value of clean BarResult logs and the future-safety justification reinforced the choice.

Since `Next` receives the full candle history on every call, the previous bar's SMA is computed from `fastVals[n-2]` / `slowVals[n-2]` — no stored state required.

## Consequences

`Next` needs one extra bar beyond `slowPeriod` before it can detect a crossover (needs `prevSMA` to be a valid computed value, not the zero-fill from talib's lookback period). This is handled by a guard: `if n <= slowPeriod { return SignalHold }`. `Lookback()` still returns `slowPeriod` per the acceptance criterion — the guard handles the first eligible bar internally.

## Related decisions

- [SignalHold filtered at engine level — never passed to portfolio](../convention/2026-04-07-hold-signal-not-passed-to-portfolio.md) — signal semantics convention this builds on
- [Strategy.Lookback() as a first-class interface method](../architecture/2026-04-02-strategy-lookback-as-interface-method.md) — establishes that Lookback governs when Next is first called
