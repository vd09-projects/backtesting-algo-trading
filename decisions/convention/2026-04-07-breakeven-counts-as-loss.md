# Break-even trades classified as losses in analytics

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | convention |
| Tags     | analytics, win-rate, pnl, trade, classification, convention |

## Context

When implementing `analytics.Compute`, every closed trade must be classified as either a win or a loss to compute `WinRate`, `WinCount`, and `LossCount`. A break-even trade has `RealizedPnL == 0` — it recovered exactly what it put in, including commission and slippage.

The question was: should a break-even trade count as a win, a loss, or a third category?

## Options considered

### Option A: Break-even is a loss (RealizedPnL <= 0 → loss)
- **Pros**: Simple binary classification. A trade that made no money is a failed trade — it consumed time, commission, and margin for zero gain. Aligns with most trading analytics conventions. Win rate stays conservative and doesn't inflate.
- **Cons**: May feel counterintuitive to a user who "got their money back".

### Option B: Break-even is a win (RealizedPnL >= 0 → win)
- **Pros**: Could argue the strategy at least didn't lose money.
- **Cons**: Inflates win rate. A strategy with many break-evens looks better than it is. Misleading for strategy evaluation.

### Option C: Third category (break-even as its own bucket)
- **Pros**: Most precise.
- **Cons**: Complicates the Report struct and every downstream consumer. Win rate becomes ambiguous — do you include or exclude break-evens from the denominator? Added complexity for an edge case that rarely occurs in practice (commission makes exact break-even nearly impossible on real fills).

## Decision

Break-even trades (`RealizedPnL <= 0`) count as losses. `WinCount` increments only when `RealizedPnL > 0`. This keeps the classification binary, avoids inflating win rate, and treats "made no money" the same as "lost money" from a strategy evaluation standpoint.

## Consequences

- `LossCount` includes both losing and break-even trades — callers cannot distinguish them without re-reading the trade log directly.
- Win rate is conservative by convention. A strategy must show positive PnL on a trade to call it a win.
- If a future consumer needs to distinguish break-even trades, they must iterate `[]Trade` directly — `Report` does not expose this breakdown.

## Related decisions

- [Trade.RealizedPnL stored on the struct, not computed on-demand](../convention/2026-04-02-trade-pnl-stored-not-computed.md) — analytics relies on pre-computed PnL; classification here reads `RealizedPnL` directly

## Revisit trigger

If a strategy evaluation use case requires distinguishing break-even trades (e.g. a "scratch trade" metric), add a `BreakevenCount` field to `Report` and adjust `LossCount` to exclude break-evens.
