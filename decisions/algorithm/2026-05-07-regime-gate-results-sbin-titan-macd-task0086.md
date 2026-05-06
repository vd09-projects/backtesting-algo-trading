# Regime gate results — SBIN and TITAN, MACD 17/26/9, TASK-0086

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | regime-gate, macd-crossover, SBIN, TITAN, RegimeConcentrated, portfolio-construction, TASK-0086 |

## Context

TASK-0085 (correlation gate) produced two survivors for MACD crossover (17/26/9): NSE:SBIN and NSE:TITAN. Before finalizing capital weights in TASK-0087, the regime gate (2026-04-27) must be applied: any instrument where a single regime accounts for ≥ 70% of total absolute Sharpe mass is flagged RegimeConcentrated=true and receives 50% of its base allocation.

The `--regime-gate` flag was wired to `cmd/backtest` in TASK-0086a. `ComputeRegimeGate` in `internal/analytics/regimegate.go` runs the full-period trade log through the three NSE regime windows (NSERegimesGate) and computes per-trade Sharpe and contribution fractions. Full-period runs were used (2018-01-01 to 2024-12-31) — the function internally buckets trades by ExitTime into the three windows.

Output JSON saved to `runs/regime-gate-macd-2026-05-07/`.

## Regime gate results

### NSE:SBIN — MACD 17/26/9

Total trades: 50 | Full-period per-trade Sharpe: 0.8015

| Regime                                      | PerTradeSharpe | Contribution | TradeCount |
|---------------------------------------------|----------------|--------------|------------|
| Pre-COVID (2018–Jan 2020)                   | 0.3040         | 39.01%       | 14         |
| COVID crash + recovery (Feb 2020–Jun 2021)  | 0.3260         | 41.84%       | 11         |
| Post-recovery (Jul 2021–2024)               | 0.1492         | 19.15%       | 25         |

**RegimeConcentrated: false**

Max contribution: 41.84% (COVID crash + recovery) — well below the 70% threshold. All three regime windows have ≥ 11 trades (above the zero-trade undefined-Sharpe trigger). Sharpe is positive across all three regimes.

### NSE:TITAN — MACD 17/26/9

Total trades: 58 | Full-period per-trade Sharpe: 0.8045

| Regime                                      | PerTradeSharpe | Contribution | TradeCount |
|---------------------------------------------|----------------|--------------|------------|
| Pre-COVID (2018–Jan 2020)                   | 0.3976         | 45.97%       | 14         |
| COVID crash + recovery (Feb 2020–Jun 2021)  | 0.3681         | 42.56%       | 12         |
| Post-recovery (Jul 2021–2024)               | 0.0993         | 11.47%       | 32         |

**RegimeConcentrated: false**

Max contribution: 45.97% (Pre-COVID) — below the 70% threshold. All three regime windows have ≥ 12 trades. Sharpe is positive in all three regimes, though post-recovery Sharpe is materially lower (0.0993 vs 0.37–0.40 in the earlier regimes), consistent with the grinding post-2021 market.

## Gate outcome and capital allocation

Both instruments pass the regime gate. Neither receives a half-weight penalty.

| Instrument | RegimeConcentrated | Allocation adjustment | Final base allocation |
|------------|-------------------|-----------------------|----------------------|
| NSE:SBIN   | false             | none                  | ₹1,50,000            |
| NSE:TITAN  | false             | none                  | ₹1,50,000            |

Total portfolio notional: ₹3,00,000. Vol-targeting sizing applies on top of the base allocation (10% annualized vol target per instrument), as pre-committed in `decisions/algorithm/2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md`.

## Comparison with Marcus's prior

Marcus's prior (recorded in `2026-05-06-regime-gate-application-macd-task0055.md`) expected max contribution in the 40–60% range for COVID+recovery, with RegimeConcentrated=false for all instruments. Actual results:

- SBIN max contribution: 41.84% (COVID+recovery) — consistent with prior
- TITAN max contribution: 45.97% (Pre-COVID) — consistent with prior; Pre-COVID slightly edges out COVID+recovery for TITAN, which is plausible given Tata consumer brand momentum in 2018–2019

The prior was directionally correct. The gate passes for both instruments on first run.

## Observations

1. **Post-recovery regime is the weakest for both instruments.** SBIN post-recovery contribution 19.15%, TITAN 11.47%. This is structurally expected: 2022 was a sideways-to-down rate-hike market with choppy MACD signals. The lower post-recovery Sharpe does not trigger the gate but should be monitored if the evaluation window is extended into 2025.

2. **TITAN's pre-COVID Sharpe slightly exceeds its COVID+recovery Sharpe** (0.3976 vs 0.3681). This is unusual for a trend-following strategy — COVID+recovery was the sharpest trending regime. One plausible explanation: TITAN's price action in 2018–2019 was strongly trending (Tanishq brand expansion, earnings growth), giving MACD clean signals. The COVID crash phase itself may have produced some whipsaw before the recovery leg.

3. **Trade count distribution is healthy.** No regime has fewer than 11 trades, keeping per-trade Sharpe estimates meaningful. The regime gate's edge case (zero trades → concentrated by default) does not apply here.

## Consequences

- Both NSE:SBIN and NSE:TITAN proceed to TASK-0087 (portfolio composition file) at full base allocation.
- The freed path: correlation gate → regime gate → TASK-0087 (portfolio composition) is now unblocked.
- TASK-0086 acceptance criteria are met in full.

## Related decisions

- [Regime gate methodology (2026-04-27)](./2026-04-27-regime-gate.md) — gate criteria applied
- [Regime gate application — MACD TASK-0055 (2026-05-06)](./2026-05-06-regime-gate-application-macd-task0055.md) — Marcus's prior and computation requirements
- [MACD correlation gate results — SBIN and TITAN survive (2026-05-06)](./2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md) — source of the two instruments evaluated here
- [MACD portfolio sizing — SBIN + TITAN (2026-05-06)](./2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — base allocations confirmed unmodified by this gate
