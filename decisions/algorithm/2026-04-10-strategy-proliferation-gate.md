# Strategy proliferation gate — Sharpe ≥ 0.5 vs buy-and-hold before building variation strategies

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | strategy, sharpe, gate, research-methodology, MACD, bollinger-bands, SMA, RSI, buy-and-hold, overfitting |

## Context

The backlog includes four strategy families: SMA crossover, RSI mean-reversion, MACD trend-following, and Bollinger Band mean-reversion. MACD is a variation of the same trend-following thesis as SMA crossover; Bollinger Bands are a variation of the same mean-reversion thesis as RSI. Building all four before seeing any results would waste significant implementation time if the underlying thesis has no pulse in the target market (NSE Indian equities, daily bars).

The question: under what conditions should the variation strategies (MACD, Bollinger Bands) be built at all?

## Options considered

### Option A: Build all four strategies unconditionally
- **Pros**: Complete baseline set, easier to compare all approaches at once.
- **Cons**: High implementation cost if the thesis is dead; creates attachment to strategies regardless of data signal; encourages building before evaluating.

### Option B: Build variation strategies only if baseline Sharpe exceeds a threshold vs buy-and-hold
- **Pros**: Data-driven go/no-go; prevents reflexive strategy proliferation; consistent with "no edge, no code" research discipline.
- **Cons**: Threshold is somewhat arbitrary; could miss a case where SMA fails but MACD succeeds for structural reasons.

### Option C: Decide after seeing results with no pre-committed threshold
- **Pros**: Flexible.
- **Cons**: Human bias will likely lead to "trying one more thing" regardless of results; no pre-commitment means the decision gets made under the influence of sunk cost.

## Decision

Option B, with a threshold of **Sharpe ≥ 0.5 vs buy-and-hold after Zerodha costs** for the baseline strategy in each thesis category.

Specific gates:
- **MACD** (TASK-0019): Do not start until SMA crossover AND RSI mean-reversion are both evaluated against buy-and-hold. If neither achieves Sharpe ≥ 0.5, cancel MACD.
- **Bollinger Bands** (TASK-0020): Do not start until RSI mean-reversion is evaluated against buy-and-hold. If RSI does not achieve Sharpe ≥ 0.5, cancel Bollinger Bands.

Sharpe ≥ 0.5 is a deliberately low bar — it filters out strategies with no signal at all while not demanding that a dirty test strategy be production-ready. A real edge on daily NSE bars would likely sit between 0.3 and 1.2 after costs.

The pre-commitment matters as much as the number. The threshold is set before seeing results so the decision isn't made retrospectively under confirmation bias.

## Consequences

- If both SMA crossover and RSI underperform buy-and-hold after costs, TASK-0019 and TASK-0020 are cancelled and the project moves to researching a different edge thesis entirely.
- If one passes and one fails, only the passing variation gets built (e.g., SMA passes → MACD proceeds; RSI fails → no Bollinger Bands).
- The gate applies to the initial dirty test, not a formal walk-forward. A strategy that barely clears 0.5 in-sample is still treated with skepticism — this is a filter for "alive or dead," not a promotion to production.

## Revisit trigger

If the target universe changes (e.g., switching from daily NSE equities to 15-min futures), reset the gate — different microstructure and cost structure require re-evaluation. The 0.5 threshold was calibrated for daily bar strategies on liquid NSE large-caps with Zerodha commission structure.
