# Strategy.Lookback() as a first-class interface method

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-02       |
| Status   | accepted         |
| Category | architecture     |
| Tags     | strategy, interface, engine, lookback, no-lookahead |

## Context

When designing the `Strategy` interface, we needed a way for the engine to know how many historical candles a strategy requires before it can emit a meaningful signal. Without this, the engine either calls the strategy from bar 0 (wrong — most indicators need a warm-up period) or each strategy silently ignores early bars and emits holds (brittle — the engine can't distinguish "holding" from "not ready yet").

## Options considered

### Option A: Lookback as a first-class interface method
Make `Lookback() int` a required method on the `Strategy` interface. The engine reads this once at startup and begins calling `Next()` only after that many candles have accumulated.

- **Pros**: Engine enforces the constraint universally — no strategy can accidentally receive fewer bars than it needs. Makes the contract explicit and testable.
- **Cons**: Every strategy must implement one more method. Constant-value implementations feel boilerplate-y for simple strategies.

### Option B: Lookback in a config/options struct passed at construction
Strategy takes a config struct at construction time that includes a lookback field. Engine reads it from the struct.

- **Pros**: Groups all strategy parameters in one place.
- **Cons**: Requires a config convention enforced by documentation rather than the compiler. Easy to forget in custom strategies. Not inspectable without knowing the struct layout.

### Option C: Engine starts at bar 0, strategies handle their own warm-up
Engine always calls `Next()` from the first bar. Strategies emit `SignalHold` until they have enough data internally.

- **Pros**: Simpler interface.
- **Cons**: Engine cannot distinguish "holding because the strategy decided to" from "holding because insufficient data". Silent failures. No-lookahead enforcement becomes the strategy's responsibility, not the engine's.

## Decision

**Option A** — `Lookback() int` as a required interface method.

The engine's job is to enforce the no-lookahead constraint. Delegating that responsibility to individual strategies is a footgun: a strategy that gets this wrong produces incorrect backtest results silently. By making lookback part of the interface, the engine can enforce it universally without any per-strategy special-casing, and the contract is verifiable at compile time.

## Consequences

- Every `Strategy` implementation must declare a lookback, even trivial ones that only look at the current bar (they return `1`).
- The engine loop starts at index `Lookback() - 1`, passing `candles[:i+1]` to `Next()` — no future bars ever reach the strategy.
- Testable: a stub strategy with a known lookback can verify the engine starts at the right bar.

## Revisit trigger

If we later need variable lookbacks (e.g., adaptive strategies that change their required history at runtime), this interface becomes insufficient. Revisit if any strategy needs a dynamic lookback value.
