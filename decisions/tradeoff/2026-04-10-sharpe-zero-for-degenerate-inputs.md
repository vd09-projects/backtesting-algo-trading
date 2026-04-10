# Sharpe returns 0 for degenerate inputs — Compute() stays error-free

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | tradeoff         |
| Tags     | sharpe, analytics, error-handling, pure-function, Compute, degenerate, zero-variance, timeframe, API-design |

## Context

`analytics.Compute()` is a pure function: no I/O, no side effects, no error return. It accepts trades, an equity curve, and a timeframe, and returns a `Report`. This design was inherited from the original trade-only implementation.

During Sharpe implementation, three degenerate input cases arose that have no meaningful Sharpe value:
1. Fewer than 3 equity points (< 2 returns — sample stddev undefined)
2. Zero-variance equity curve (constant equity — division by zero)
3. Unknown timeframe (annualization factor unknown — result would be unitless garbage)

Each case requires a decision: return an error, panic, or return a sentinel value.

## Options considered

### Option A: Add error return to Compute()
- **Pros**: Caller knows explicitly when Sharpe couldn't be computed; no silent failure.
- **Cons**: Breaking change to every existing call site. Forces callers to handle errors that are almost always benign (e.g., a strategy that produced no trades has no meaningful Sharpe anyway). Breaks the pure-function contract that makes Compute easy to use in pipelines and tests. Error handling for "no data" is noise at the CLI level.

### Option B: Panic on invalid inputs
- **Pros**: Loud failure, impossible to miss.
- **Cons**: Crashes the process on inputs that are valid and expected (a new strategy with no trades, a 5-bar warm-up-only run). Violates the project rule: "Errors are returned, not panicked."

### Option C: Return 0 for degenerate inputs (chosen)
- **Pros**: Keeps Compute a pure value function. Consistent with how other metrics behave (WinRate=0 when no trades; MaxDrawdown=0 when no positive peak). A Sharpe of 0 is visually prominent in output — an analyst will notice and investigate. No call-site changes needed.
- **Cons**: Silent failure — if `sharpeAnnualizationFactor` is missing a new timeframe case, the caller gets Sharpe=0 with no indication of why. The switch statement must be kept exhaustive.

## Decision

Return **0** for all degenerate inputs. The `Compute()` signature stays `(trades, curve, tf) → Report` with no error return.

The guard ordering in `computeSharpe`:
1. `len(curve) < 3` → 0
2. Unknown timeframe (`annFactor == 0`) → 0
3. Zero variance → 0

## Consequences

- **Exhaustive switch required**: `sharpeAnnualizationFactor` must be updated whenever a new `Timeframe` constant is added to `pkg/model`. If it isn't, Sharpe silently returns 0 for that timeframe — tests for the new timeframe are the safety net. Added a coverage requirement: `sharpeAnnualizationFactor` must stay at 100% coverage so any new branch omission is caught immediately.
- **Analyst responsibility**: A Sharpe of exactly 0.0000 in output should be treated as "could not compute" rather than "strategy has zero Sharpe." This is consistent with the other zero-value metrics but relies on the analyst understanding the output.
- **Walk-forward windows**: Short out-of-sample windows could legitimately produce < 3 equity points. Walk-forward logic (TASK-0022) should treat Sharpe=0 as "insufficient data" rather than "flat strategy."

## Related decisions

- [Sharpe uses sample variance (n-1)](../algorithm/2026-04-10-sharpe-sample-variance.md) — the n<2 guard exists because sample variance requires ≥2 observations
- [NSE annualization factors](../convention/2026-04-10-nse-annualization-factors.md) — the switch that must stay exhaustive

## Revisit trigger

If a future analytics metric also needs to signal "could not compute" distinctly from "result is zero" (e.g., Sortino with no negative returns), consider adding an optional `Diagnostics` struct to `Report` rather than changing the error contract.
