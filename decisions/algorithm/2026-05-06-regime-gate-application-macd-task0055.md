# Regime gate application — MACD crossover survivors, TASK-0055

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-06       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | regime-gate, MACD-crossover, concentration, half-weight, TASK-0055, TASK-0086 |

## Context

The regime gate (2026-04-27) was deferred from TASK-0052 (universe sweep) because the universe-sweep CSV contained no per-trade timestamps, making per-regime Sharpe computation impossible from that output. This deferral was recorded in `decisions/algorithm/2026-05-03-regime-gate-deferred-universe-sweep-csv.md`.

TASK-0055 (portfolio construction) now requires the regime gate to be computed for all surviving instruments before finalizing capital weights. This decision records Marcus's position on the expected outcome and the computation requirements for TASK-0086.

## Decision

The regime gate computation is a **blocking step before finalizing capital weights** in the portfolio. It cannot be skipped or treated as advisory. Per the regime gate methodology (2026-04-27), any instrument where a single regime accounts for ≥ 70% of total absolute Sharpe mass is flagged as RegimeConcentrated=true and receives 50% of its base capital allocation.

## Marcus's prior on MACD crossover regime concentration

MACD crossover (17/26/9) on Nifty50 daily bars is unlikely to hit 70% regime concentration for any of the 4 surviving instruments, for the following structural reasons:

1. **COVID+recovery (Feb 2020 – Jun 2021)** was the strongest trending regime in the evaluation window. MACD is a trend-following strategy — this regime contributed materially. However, this window is only ~17 months of the 72-month evaluation period.

2. **Post-recovery (Jul 2021 – Dec 2024)** is 3.5 years including the 2022 rate-hike bear market. The 2022 bear was not a crash-and-recover event — it was a grinding sideways-to-down market that would produce mixed MACD signals. Post-recovery Sharpe is expected to be positive but lower than COVID+recovery, diluting the COVID window's contribution below 70%.

3. **Pre-COVID (2018-2020)** includes the IL&FS crisis aftermath (late 2018) and mixed market conditions. Per-trade Sharpe for MACD in this period is expected to be modest — possibly flat or slightly negative on some instruments — but not so negative as to dominate the absolute Sharpe mass.

**Expected outcome:** max(contribution[i]) in the 40–60% range for COVID+recovery, not reaching the 70% threshold. All 4 survivors expected to have RegimeConcentrated=false, preserving full capital allocation.

**This prior must be validated against actual computation in TASK-0086.** Apply the gate even when it is expected to pass. Pre-commitment means not skipping gates that confirm priors.

## Computation requirements for TASK-0086

Per-regime backtests must be run for all instruments passing the correlation gate in TASK-0085. Use `cmd/backtest` with restricted date ranges per regime window:

| Regime | Date range |
|---|---|
| Pre-COVID | 2018-01-01 to 2020-01-31 |
| COVID+recovery | 2020-02-01 to 2021-06-30 |
| Post-recovery | 2021-07-01 to 2024-12-31 |

Alternatively, if `cmd/backtest` produces per-trade timestamps in the output JSON, extract trades falling within each window from the existing full-period run. Confirm timestamp availability in `runs/bootstrap-macd-2026-05-05/*.json` before choosing approach.

Per-regime Sharpe formula (same as all other per-trade Sharpe in this project):

```
S[regime] = mean(ReturnOnNotional[trades in regime]) / std(ReturnOnNotional[trades in regime], ddof=1)
```

No annualization. Sample variance (ddof=1). If fewer than 5 trades in a regime window, flag as insufficient sample and treat as RegimeConcentrated=true by default (per the regime gate decision consequence clause).

## Integration with portfolio sizing

Results from TASK-0086 feed into TASK-0087 (portfolio composition decision file). The base allocation is:
- NSE:SBIN: ₹1,50,000
- NSE:TITAN: ₹1,50,000

If either instrument returns RegimeConcentrated=true: halve that instrument's allocation. The freed capital stays in cash; do not reallocate to the other instrument.

## Related decisions

- [Regime gate methodology (2026-04-27)](./2026-04-27-regime-gate.md) — the gate criteria applied here
- [Regime gate deferred from TASK-0052 (2026-05-03)](./2026-05-03-regime-gate-deferred-universe-sweep-csv.md) — why this was deferred
- [Portfolio sizing and kill-switch (2026-05-06)](./2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — base allocations before regime gate adjustment
