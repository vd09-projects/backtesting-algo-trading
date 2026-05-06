---
name: Evaluation pipeline thresholds
description: Gate-by-gate thresholds as currently configured in this project — authoritative reference for Marcus evaluations
type: project
---

All thresholds are established via `decisions/algorithm/`. The source decision file is listed for each gate.

**Gate 1 — Signal frequency (proliferation) gate**
- Threshold: ≥ 30 trades per instrument across the evaluation period (2018-2024)
- Universe: 15 Nifty50 large-cap instruments
- Pass condition: ≥ 40% of 15 instruments must have ≥ 30 trades
- Decision: `2026-04-10-strategy-proliferation-gate.md`, `2026-04-25-cross-instrument-proliferation-gate.md`
- High risk flag: strategies expected to generate < 35 trades/year on daily bars

**Gate 2 — Universe gate**
- Threshold: DSR-corrected average Sharpe > 0 AND ≥ 40% of 15 instruments show positive Sharpe with ≥ 30 trades
- Parameters: plateau-midpoint parameter from 1D sensitivity sweep on RELIANCE
- Decision: referenced in universe sweep tasks; see `decisions/algorithm/2026-05-03-macd-crossover-universe-gate-passed.md` for example

**Gate 3 — Walk-forward gate**
- Structure: 2yr IS / 1yr OOS / 1yr step, 2018-2024
- Pass condition: OverfitFlag = false AND NegativeFoldFlag = false
- Instrument-count gate (revised 2026-05-05): ≥ 60% of instruments that passed universe gate must also pass WF gate
- Decision: `2026-04-22-walk-forward-oos-is-sharpe-threshold.md`, `2026-05-05-walk-forward-instrument-count-gate-relaxed.md`

**Gate 4 — Bootstrap gate**
- Simulations: 10,000, seed=42
- Pass condition: SharpeP5 > 0 AND P(Sharpe > 0) > 80%
- Sharpe formula: per-trade, mean(ReturnOnNotional)/std(ReturnOnNotional), sample variance, no annualization
- Decision: `2026-04-27-bootstrap-gate.md`, `2026-04-20-bootstrap-sharpe-non-annualized-per-trade.md`

**Gate 5 — Correlation gate**
- Full period: Pearson r < 0.7 (2018-2024 daily log-returns of equity curve)
- Stress periods: Pearson r < 0.6 for COVID crash (2020-02-01 to 2020-06-30) AND rate-hike bear (2022-01-01 to 2022-12-31)
- Tiebreaker: retain higher DSR-corrected Sharpe instrument
- Decision: `2026-04-27-correlation-gate.md`

**Gate 6 — Regime gate**
- Windows: pre-COVID (2018-2020-01), COVID+recovery (2020-02 to 2021-06), post-recovery (2021-07 to 2024-12)
- Metric: abs(S[regime]) / sum(abs(S[all regimes])) per instrument
- Gate: max contribution < 0.70 → RegimeConcentrated=false (full weight)
- Penalty if concentrated: half-weight in portfolio (not a kill)
- Decision: `2026-04-27-regime-gate.md`

**Capital and sizing**
- Total capital: ₹3,00,000
- Vol target: 10% annualized
- Sizing: SizingVolatilityTarget — fraction = volTarget/(instrumentVol × sqrt(252)), capped at 1.0, no leverage
- Rolling window for vol estimate: 20 bars
- Decision: `2026-04-13-vol-targeting-algorithm-choices.md`

**Kill-switch methodology**
- Rolling per-trade Sharpe threshold: SharpeP5 from bootstrap
- Max drawdown threshold: 1.5 × in-sample MaxDrawdown
- Max DD duration threshold: 2 × in-sample MaxDrawdownDuration
- Decision: `2026-04-21-kill-switch-derivation-methodology.md`
