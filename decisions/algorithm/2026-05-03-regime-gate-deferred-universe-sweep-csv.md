# Regime gate deferred from universe sweep — per-period trade log required

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-03       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | regime-gate, universe-sweep, CSV-limitation, deferred, TASK-0052, TASK-0055 |

## Context

TASK-0052 (universe sweep) required applying both the universe gate and the regime gate. The universe gate was applied cleanly from the CSV output. The regime gate — which checks that no single regime (pre-COVID, COVID, post-recovery) accounts for >= 70% of total Sharpe — requires computing per-trade Sharpe within each regime window, which means knowing which trades fell in which regime window.

The `cmd/universe-sweep` CSV output (`runs/universe-sweep-2026-05-03.csv`) contains only aggregate per-instrument metrics: sharpe, trade_count, total_pnl, max_drawdown, insufficient_data. It does not contain per-trade timestamps or per-period equity breakdowns.

## Options considered

### Option A: Run a separate per-regime backtest for each survivor × instrument pair
Re-run `cmd/backtest` for each of the 3 regime windows separately and compute per-window Sharpe. Produces exact regime contribution fractions.
- **Cons**: 2 survivors × 26 instruments × 3 regimes = 156 additional runs. Significant overhead at a stage where the full instrument list hasn't been filtered yet. Adds complexity to TASK-0052 scope.

### Option B: Defer regime gate to portfolio construction (TASK-0055) (chosen)
The regime gate decision (2026-04-27) explicitly states that regime failures are flagged (half-weight) but NOT killed. The gate is informational, not a kill condition. Portfolio construction already requires full equity curves and trade logs — the regime computation can be done there without additional runs.
- **Pros**: No extra runs. The survivors entering walk-forward are defined by the universe gate alone. Regime concentration is a portfolio-sizing concern, not a survival condition.
- **Cons**: Regime information is not available during walk-forward planning. A heavily regime-concentrated strategy could consume walk-forward compute budget before the concentration is discovered.

### Option C: Extend `cmd/universe-sweep` to output per-period Sharpe columns
Add pre-COVID / COVID / post-recovery Sharpe columns to the universe sweep CSV.
- **Cons**: Requires engine changes to support sub-period statistics in a single sweep run. Out of scope for TASK-0052; deferred to a potential future improvement.

## Decision

**Regime gate deferred to TASK-0055 (portfolio construction).** The universe sweep CSV does not contain per-period trade data. Per the regime gate decision (2026-04-27), failures result in half-weight in portfolio, not a kill — so deferral does not risk a strategy proceeding past a hard gate.

Both MACD crossover and SMA crossover are flagged as `RegimeConcentrated: deferred` in their survivor metric records. TASK-0055 must compute per-regime Sharpe contributions from the walk-forward equity curves and apply the 70% concentration threshold before finalizing portfolio weights.

## Consequences

- Walk-forward (TASK-0053) proceeds with both survivors, regime status unknown.
- TASK-0055 (portfolio construction) must include regime gate computation as an explicit acceptance criterion. The survivor metric records for MACD and SMA note `RegimeConcentrated: deferred`.
- If `cmd/universe-sweep` is extended to output per-period statistics (Option C), this decision can be superseded.

## Related decisions

- [Regime gate — no single regime accounts for more than 70% of total Sharpe](./2026-04-27-regime-gate.md) — the gate being deferred; specifies that failures are flagged, not killed

## Revisit trigger

If `cmd/universe-sweep` is extended to output per-regime Sharpe columns (Option C), this deferral is no longer necessary. Revisit at that point and apply the regime gate directly in TASK-0052 for future evaluation runs.
