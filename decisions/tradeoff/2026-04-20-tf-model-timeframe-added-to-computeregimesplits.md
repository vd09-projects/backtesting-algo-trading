# `tf model.Timeframe` added to `ComputeRegimeSplits` signature for correct Sharpe annualization

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-20       |
| Status   | accepted         |
| Category | tradeoff         |
| Tags     | analytics, regime-split, timeframe, annualization, Sharpe, ComputeRegimeSplits, internal/analytics, TASK-0034 |

## Context

TASK-0034 acceptance criteria specified the signature as `ComputeRegimeSplits(curve []model.EquityPoint, regimes []Regime) []RegimeReport` — no timeframe parameter. During planning it emerged that `computeSharpe` (the unexported helper in the same package) requires a `model.Timeframe` to apply the correct annualization factor (252 for daily, 6300 for 15min, etc.). The equity curve carries no bar-frequency information of its own — timestamps tell you *when* bars occurred but not their nominal duration (a daily bar and a 15min bar spanning market close can have the same timestamp distance).

## Options considered

### Option A: Match AC exactly — hardcode `model.TimeframeDaily`

The NSE 2018–2024 context is daily bars. Hardcoding `model.TimeframeDaily` inside `ComputeRegimeSplits` would match the AC signature and produce correct results for the only current caller.

- **Pros**: Matches the spec literally; no interface change needed.
- **Cons**: Silently wrong for any non-daily caller. The function name doesn't signal the assumption. A 15min strategy caller would get Sharpe annualized by 252 instead of 6300 — a factor of ~25× error — with no warning.

### Option B: Add `tf model.Timeframe` as a third parameter (chosen)

Deviate from the AC signature by one argument. Every caller must pass the timeframe explicitly; the compiler enforces it.

- **Pros**: Correct for any timeframe. Mismatch is impossible — the caller must think about timeframe at the call site. Consistent with `computeSharpe`'s own interface.
- **Cons**: One-argument deviation from the AC as written. All future callers carry the `tf` obligation.

### Option C: Return non-annualized Sharpe

Omit annualization in regime reports entirely, returning raw mean/stddev ratio.

- **Pros**: No timeframe needed; matches the AC signature.
- **Cons**: Regime Sharpe would be incomparable to the full-period Sharpe printed in the main backtest report (which is annualized). A regime showing Sharpe 0.05 (raw) vs 0.73 (annualized daily) would be misleading — the numbers look smaller without explanation.

## Decision

Added `tf model.Timeframe` as a third parameter to `ComputeRegimeSplits`. Option A was rejected because it hides a timeframe assumption that will produce silently wrong numbers for any future non-daily strategy. Option C was rejected because non-annualized regime Sharpe is not comparable to the annualized full-period Sharpe — mixing the two in the same output is confusing.

The AC deviation is intentional and narrow: one additional parameter, fully motivated. The function signature is `ComputeRegimeSplits(curve []model.EquityPoint, regimes []Regime, tf model.Timeframe) []RegimeReport`.

## Consequences

All callers of `ComputeRegimeSplits` must supply the bar timeframe. The only current caller is `cmd/backtest/main.go`, which already has `tf` in scope from flag parsing. Any future caller that doesn't know the timeframe (e.g., a post-processing tool reading a CSV) must infer it from context or pass `model.TimeframeDaily` explicitly — and accept the semantic responsibility for that choice.

## Related decisions

- [NSE annualization factors](../convention/2026-04-10-nse-annualization-factors.md) — defines the per-timeframe constants that make the `tf` parameter necessary
- [Sharpe returns 0 for degenerate inputs](../tradeoff/2026-04-10-sharpe-zero-for-degenerate-inputs.md) — same degenerate-input policy applies to per-regime Sharpe
