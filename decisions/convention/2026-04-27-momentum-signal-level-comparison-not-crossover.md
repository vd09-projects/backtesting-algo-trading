# Momentum strategy signal semantics: level-comparison, not crossover

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | momentum, signal-semantics, level-comparison, crossover, roc, TASK-0043 |

## Context

The `strategies/momentum` package uses `talib.Roc` to compute a 12-month rate-of-change on closes. The signal logic needed to decide between two approaches: (1) strict crossover detection — emit Buy only on the bar where ROC transitions from ≤ threshold to > threshold, and (2) level-comparison — emit Buy on every bar where ROC > threshold.

The SMA crossover and MACD crossover strategies both use strict crossover detection (see related decision). The question was whether momentum should follow the same pattern.

## Options considered

### Option A: Strict crossover (matches SMA/MACD pattern)
- **Pros**: Consistent with other strategies. Signal means "something changed", not "regime is still active". BarResult logs are diagnostic.
- **Cons**: Not semantically appropriate here. ROC threshold comparison is not a transition between two indicator lines — it's a periodic state assertion: "is the 12-month return still above 10%?" Each bar is an independent reading; there is no natural "previous ROC line" being crossed.

### Option B: Level-comparison (chosen)
- **Pros**: Semantically correct for a threshold test on a single indicator value. Each bar's ROC independently asserts whether the momentum regime holds. Under the engine's no-pyramiding constraint, this produces one entry per regime transition in practice — behaviourally equivalent to crossover detection for practical purposes, but without the semantic mismatch of treating "ROC still above threshold" as the same kind of event as "ROC just crossed threshold".
- **Cons**: Emits Buy on consecutive bars while above the threshold. Slightly noisier BarResult logs.

## Decision

Level-comparison is used: `ROC(lookback) > threshold → Buy`, `ROC(lookback) < -threshold → Sell`, `Hold` otherwise. This is the correct semantics for a single-indicator threshold test.

The distinction matters for a specific reason: the SMA and MACD strategies emit signals when two time series cross — the event has a natural "before" and "after" state, and the crossover is the transition. ROC against a static threshold has no such pair — each bar's ROC is an independent assessment of whether momentum exceeds the configured magnitude. Forcing strict crossover semantics onto this case would require tracking "was ROC above threshold last bar?" solely to produce a Hold on bars 2..N of a bullish momentum regime. That tracking state adds complexity without adding meaning.

Under the engine's no-pyramiding rule, the practical result is identical: one long position is opened when ROC first exceeds threshold and held until ROC falls below -threshold. The BarResult log will show consecutive Buy signals during the hold period, but this is a documentation difference, not a behavioural one.

## Consequences

- The momentum strategy differs from SMA/MACD/Bollinger in signal semantics. Code reviewers reading `Next()` should expect level-comparison and not interpret it as a bug.
- BarResult logs for momentum will show consecutive Buy signals during a momentum regime. This is correct and expected — not a sign of the no-pyramiding guard misfiring.
- If a future strategy wraps momentum with `TimedExit` (TASK-0039), the consecutive Buy signals are harmless because the wrapper tracks entry bar index from the first Buy; subsequent Buys on the same position are silently skipped by the engine.

## Related decisions

- [SMA crossover: strict crossover detection, not level comparison](../algorithm/2026-04-13-sma-crossover-strict-crossover-vs-level-comparison.md) — the counterpart decision: why SMA uses strict crossover, and why the same semantics don't apply to all strategies.

## Revisit trigger

If a future strategy uses ROC or a similar single-indicator threshold and there is a strong reason to prefer strict crossover detection (e.g., the threshold test is applied to a derived series where transitions carry more information than the level), revisit whether this convention should apply broadly or only to the momentum strategy.
