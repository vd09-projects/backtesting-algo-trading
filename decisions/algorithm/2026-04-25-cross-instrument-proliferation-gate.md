# Cross-instrument universe gate supersedes single-instrument proliferation gate

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-25       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | proliferation-gate, cross-instrument, universe-sweep, DSR, evaluation-methodology, TASK-0049, TASK-0052 |

## Context

The original proliferation gate (2026-04-10) required strategies to achieve Sharpe ≥ 0.5 on NSE:RELIANCE 2018-2024 before variation strategies were built. Both SMA crossover (Sharpe 0.447) and RSI mean-reversion (Sharpe 0.469, 7 trades) failed this gate on RELIANCE, causing MACD and Bollinger Bands to be cancelled (TASK-0019 and TASK-0020).

The problem with the single-instrument gate is that a strategy failing on RELIANCE tells you almost nothing about whether it has edge on NSE large-caps generally. RELIANCE's price behaviour (large-cap trending stock, limited mean-reversion windows, liquidity that absorbs signals quickly) is not representative of the full universe. Building a gate around one instrument's results is testing the wrong thing — it tests the match between the strategy and that one instrument, not whether the edge thesis is real.

## Options considered

### Option A: Keep the single-instrument gate, tune the instrument or threshold
- **Pros**: Simple, fast to run.
- **Cons**: Still single-instrument. A different threshold would require seeing results before setting it — post-hoc rationalization. Choosing a friendlier instrument would be survivorship bias by hand.

### Option B: Cross-instrument universe gate — strategy must work on ≥ 40% of 15 Nifty50 large-caps (selected)
- **Pros**: Tests whether the edge thesis generalises across instruments. A strategy that works on 6+ of 15 instruments is more likely capturing real market structure than instrument-specific noise. DSR correction is applied automatically by the universe-sweep CLI.
- **Cons**: Requires building all strategy families before running any gate — more upfront implementation work. The 40% threshold is a pre-committed number, not derived from data, so it is somewhat arbitrary. However, it was declared before any results were seen, which is its primary value.

### Option C: CPCV with cross-instrument bootstrap
- **Pros**: More statistical rigour.
- **Cons**: CPCV infrastructure not built; significantly more work. Not appropriate at this stage.

## Decision

The **2026-04-10 single-instrument proliferation gate is superseded** for the new six-strategy evaluation. The replacement gate is: after universe sweep on 15 Nifty50 large-cap instruments (`universes/nifty50-large-cap.yaml`), a strategy survives if **DSR-corrected average Sharpe > 0 AND ≥ 40% of instruments show positive Sharpe with ≥ 30 trades**. A strategy that only passes on RELIANCE is still killed.

The key changes from the old gate:
- Evidence scope: 15 instruments instead of 1
- Primary metric: DSR-corrected average Sharpe across instruments (penalised for the 15-instrument search) rather than raw Sharpe on one instrument
- Trade count gate: ≥ 30 trades per instrument (same as the existing analytics gate); insufficient-trade instruments are excluded from the 40% count, not counted as failures
- No minimum absolute Sharpe level — a strategy with average DSR-corrected Sharpe of 0.1 across 8 instruments passes this gate (the higher bars are in walk-forward and bootstrap)

The six gate thresholds for the full evaluation pipeline (universe, walk-forward, bootstrap, regime, correlation, kill-switch) are documented together in TASK-0049 before any evaluation run begins.

## Consequences

- TASK-0019 (MACD) and TASK-0020 (Bollinger) are re-opened as TASK-0041 and TASK-0042, superseding the prior cancellations.
- All six strategy families (SMA, RSI, Donchian, MACD, Bollinger, Momentum) are built before any gate is applied.
- The single-instrument proliferation gate decision (2026-04-10) is marked superseded.
- The prior gate failure decisions (2026-04-16 for SMA and RSI on RELIANCE) remain valid as historical record of single-instrument behaviour but do not constitute gate failures under the new methodology.

## Related decisions

- [Strategy proliferation gate — Sharpe ≥ 0.5 (superseded)](./2026-04-10-strategy-proliferation-gate.md) — the gate this decision replaces
- [SMA crossover fails proliferation gate — NSE:RELIANCE](./2026-04-16-sma-crossover-proliferation-gate-failed.md) — prior gate verdict; remains valid as single-instrument evidence but not a gate failure under this methodology
- [RSI mean-reversion fails proliferation gate — NSE:RELIANCE](./2026-04-16-rsi-mean-reversion-proliferation-gate-failed.md) — same

## Revisit trigger

If more than 50% of Nifty50 instruments consistently produce insufficient trade counts (< 30 trades in the 2018-2023 window) for any strategy family, revisit the 40% threshold and the minimum-trades exclusion rule — the gate may be mathematically unachievable for low-frequency strategies.
