# TASK-0051 routing: MACD proceeds on defaults; other 5 require parameter sensitivity pass first

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | TASK-0051, TASK-0052, signal-audit, routing, parameter-sensitivity, macd, 30-trade-floor, kill-gate, Marcus |

## Context

Signal frequency audit (TASK-0050, run 2026-04-29) produced the following trade counts on NSE:RELIANCE 2018–2024 at default parameters:

| Strategy | Trades/instrument | Result |
|---|---|---|
| MACD (12/26/9) | 44–65 | PASSES (all 15 instruments) |
| SMA crossover (10/50) | 18–23 | EXCLUDED (all 15) |
| RSI mean-reversion (14, 30/70) | 3–9 | EXCLUDED (all 15) |
| Donchian breakout (20) | 13–19 | EXCLUDED (all 15) |
| Bollinger mean-reversion (20, 2.0σ) | 12–19 | EXCLUDED (all 15) |
| Momentum (231, 10%) | 1–4 | EXCLUDED (all 15) |

The question was how to route each strategy into TASK-0051 (in-sample baseline and parameter sensitivity).

## Decision

**MACD** advances to TASK-0051 on its default parameters (12, 26, 9). The signal audit confirmed sufficient trade frequency; the orientation run and parameter sweep proceed normally.

**SMA, RSI, Donchian, Bollinger, Momentum** each require a parameter sensitivity pass on RELIANCE (TASK-0051) to find a parameter region that generates ≥ 30 trades per instrument AND shows plausible Sharpe. The sweep ranges are:
- SMA: slow-period 20→60 step 5 (fast fixed at 10)
- RSI: rsi-period 7→21 step 2
- Donchian: donchian-period 10→30 step 2
- Bollinger: bb-period 10→40 step 5
- Momentum: momentum-lookback 30→180 step 15

A strategy that cannot clear 30 trades per instrument at **any** parameter in its reasonable sweep range is killed at this gate — not held for pooled analysis across instruments. This is a pre-gate kill condition distinct from the universe sweep gate.

## Consequences

- MACD is the clear frontrunner: already in pipeline shape. Expect it to advance through TASK-0052 (universe sweep) without parameter changes.
- Momentum (231-day lookback) is the most at risk: 1–4 trades/instrument at default is near-zero signal. The sweep range (30→180 days) is the last viable rescue attempt.
- RSI at 3–9 trades/instrument is the second most at risk. Loosening the bands (e.g. 35/65) is an alternative if the period sweep doesn't help — but that would require a separate sweep run.
- Strategies killed here have their rejection recorded in `decisions/algorithm/`.

## Related decisions

- [plateau procedure — trade-count-constrained strategies](2026-04-29-plateau-procedure-trade-count-constrained.md) — defines how the 80% Sharpe floor is applied within the ≥30-trade valid region
- [fallback to strategy defaults when no valid plateau](2026-04-29-fallback-to-defaults-no-valid-plateau.md) — what happens when no ≥30-trade region exists in the sweep

## Revisit trigger

If the pipeline is extended with pooled-instrument analysis (treating all 15 instruments as one equity curve), the per-instrument 30-trade floor may be relaxed. Revisit routing then.
