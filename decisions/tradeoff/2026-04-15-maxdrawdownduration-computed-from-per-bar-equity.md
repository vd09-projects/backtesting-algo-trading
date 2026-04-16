# MaxDrawdownDuration computed from per-bar equity curve, not trade P&L accumulation

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-15       |
| Status   | superseded       |
| Category | tradeoff         |
| Tags     | drawdown, duration, equity-curve, analytics, MaxDrawdownDuration, TASK-0017 |

## Context

TASK-0017 added `MaxDrawdownDuration time.Duration` to `analytics.Report`, tracking wall-clock
time from the max-drawdown peak to first equity recovery (or last bar if never recovered). The
implementation had to choose which equity time series to compute duration against.

Two series exist in the analytics package:

1. The **closed-trade P&L accumulation** — one sample per closed trade, used by the existing
   `MaxDrawdown` (depth) computation. No timestamps; only equity values.
2. The **per-bar `[]model.EquityPoint` curve** — one sample per bar, passed into `Compute()`.
   Each point carries both a value and a `time.Time` timestamp.

## Options considered

### Option A: Per-bar EquityPoint curve (chosen)
- **Pros**: Has timestamps — mandatory for any duration calculation. One sample per bar means
  intra-drawdown recovery can be identified at bar granularity.
- **Cons**: May identify a different max-drawdown event than `MaxDrawdown` (depth) on
  low-turnover strategies where bars move equity between closed trades.

### Option B: Change MaxDrawdown depth to also use EquityPoint, then share the peak
- **Pros**: Both fields describe the same drawdown event — fully consistent.
- **Cons**: Breaking behavior change to `MaxDrawdown` outside TASK-0017's scope. Would invalidate
  existing tests and the accepted decision at
  `decisions/algorithm/2026-04-07-max-drawdown-from-equity-curve.md`. Deferred to a separate task.

## Decision

`MaxDrawdownDuration` is computed from the per-bar `[]model.EquityPoint` curve: one pass to find
the peak index preceding the maximum drawdown (by percentage on the equity curve), then a forward
scan for the first bar where value >= peak value. If no recovery occurs, duration runs to the last
bar.

`MaxDrawdown` (depth %) is left unchanged — it still uses the closed-trade P&L accumulation per
the 2026-04-07 decision. Changing it was out of scope and would break the existing test suite
without a clear benefit limited to this task.

## Consequences

On strategies with frequent trades the two series track closely and both metrics will describe the
same event. On low-turnover strategies (daily signals, few trades per year), the per-bar curve
captures mark-to-market swings during open positions that the trade-close curve does not. In that
case `MaxDrawdown` and `MaxDrawdownDuration` may not describe the same drawdown event — the depth
and duration numbers are internally consistent within each metric but not cross-consistent with
each other.

This is a known limitation, documented inline in `computeMaxDrawdownDuration`. A follow-up task
(unscheduled) would unify both metrics on the EquityPoint curve and remove the inconsistency.

## Related decisions

- [MaxDrawdown computed from equity curve, not per-trade losses](../algorithm/2026-04-07-max-drawdown-from-equity-curve.md) — established that MaxDrawdown depth uses trade accumulation; this decision is directly downstream of that one
- [Equity curve records every bar, including warmup](../convention/2026-04-10-equity-curve-covers-all-bars.md) — the EquityPoint curve has consistent bar-level coverage, which is why it's usable here

## Superseded

The revisit trigger fired: `MaxDrawdown` depth was moved to `computeMaxDrawdownDepth(curve []model.EquityPoint)`
(2026-04-16). Both `MaxDrawdown` and `MaxDrawdownDuration` now walk the same per-bar equity curve,
so they always describe the same drawdown event. The inconsistency this decision documented no longer exists.
The inline decision mark in `computeMaxDrawdownDuration` was removed at the same time.
