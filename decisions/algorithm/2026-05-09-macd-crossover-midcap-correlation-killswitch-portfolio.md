# MACD crossover Nifty Midcap 150 — correlation screen, kill-switch thresholds, portfolio construction

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, correlation, kill-switch, portfolio, nifty-midcap-150, midcap, TASK-0097, survivor |

## Context

6 bootstrap survivors from TASK-0096 advance to this stage:
NSE:PERSISTENT (SharpeP5=0.151), NSE:TORNTPHARM (0.146), NSE:COFORGE (0.041), NSE:SUNDARMFIN (0.035), NSE:MUTHOOTFIN (0.025), NSE:INDHOTEL (0.019).

This decision covers three items: (1) pairwise correlation screen, (2) per-instrument kill-switch threshold derivation, (3) portfolio construction recommendation.

Source data: `results/2026-05-09-TASK-0096/` (bootstrap JSONs + bootstrap-results.csv), `results/2026-05-09-TASK-0097/correlate-output.txt`.

---

## 1. Correlation screen

Pipeline convention: |r| > 0.70 full-period is informational (not a kill gate). Secondary stress threshold: |r| > 0.60 in either stress period (2020 COVID crash, 2022 rate-correction) is an early-warning flag.

All 15 pairs computed across three windows (full period 2018–2025, 2020 crash, 2022 rate-correction):

| Pair                         | Full-period | 2020 crash | 2022 corr | Flag |
|------------------------------|-------------|------------|-----------|------|
| PERSISTENT / TORNTPHARM      | 0.1189      | 0.0639     | 0.0666    |      |
| PERSISTENT / COFORGE         | 0.3559      | 0.4083     | 0.4710    |      |
| PERSISTENT / SUNDARMFIN      | 0.0109      | -0.0506    | 0.2074    |      |
| PERSISTENT / MUTHOOTFIN      | 0.0268      | 0.1502     | 0.1633    |      |
| PERSISTENT / INDHOTEL        | 0.1256      | 0.1840     | 0.1549    |      |
| TORNTPHARM / COFORGE         | 0.0598      | -0.0573    | 0.0420    |      |
| TORNTPHARM / SUNDARMFIN      | 0.0325      | -0.0802    | 0.0151    |      |
| TORNTPHARM / MUTHOOTFIN      | -0.0220     | 0.0489     | -0.0182   |      |
| TORNTPHARM / INDHOTEL        | 0.0225      | -0.0924    | 0.0084    |      |
| COFORGE / SUNDARMFIN         | 0.0022      | 0.0852     | 0.0984    |      |
| COFORGE / MUTHOOTFIN         | -0.0029     | 0.2311     | 0.0755    |      |
| COFORGE / INDHOTEL           | 0.0568      | 0.1098     | 0.0757    |      |
| SUNDARMFIN / MUTHOOTFIN      | 0.0379      | 0.0160     | 0.0239    |      |
| SUNDARMFIN / INDHOTEL        | 0.0038      | 0.1860     | 0.0647    |      |
| MUTHOOTFIN / INDHOTEL        | 0.0440      | 0.1026     | 0.1231    |      |

**Result: zero pairs breach 0.70 full-period. Zero pairs breach 0.60 in either stress window.**

The highest correlation is PERSISTENT/COFORGE at 0.3559 (full-period), reaching 0.4710 during the 2022 rate-correction — elevated but well below both thresholds. Both are IT sector; the correlation is sector-driven and expected. See IT-sector cap rule in section 3.

All 6 survivors pass the correlation screen. No kills.

---

## 2. Kill-switch thresholds per instrument

Method: kill-switch triggers at 1.5× the historical MaxDrawdown observed in the full backtest window (2018–2025, zerodha_full commission). Duration kill-switch at 2× the historical MaxDrawdownDuration. These are live-monitoring triggers — breach means halt the instrument immediately and review before re-enabling.

`MaxDrawdownDuration` extracted from per-instrument JSON (nanoseconds → calendar days).

| Instrument     | MaxDD (hist) | Kill-switch MaxDD | MaxDD duration (hist) | Kill-switch duration | Bootstrap WorstDD P95 |
|----------------|-------------|-------------------|----------------------|----------------------|-----------------------|
| NSE:PERSISTENT  | 3.75%       | 5.62%             | 602 days             | 1204 days            | 41.43%                |
| NSE:TORNTPHARM  | 1.78%       | 2.67%             | 508 days             | 1016 days            | 25.75%                |
| NSE:COFORGE     | 8.38%       | 12.57%            | 1252 days            | 2504 days            | 53.36%                |
| NSE:SUNDARMFIN  | 4.64%       | 6.96%             | 539 days             | 1078 days            | 44.15%                |
| NSE:INDHOTEL    | 5.80%       | 8.70%             | 862 days             | 1724 days            | 55.79%                |
| NSE:MUTHOOTFIN  | 4.18%       | 6.27%             | 935 days             | 1870 days            | 42.14%                |

**COFORGE early-warning note:** COFORGE's mechanical kill-switch is 12.57% MaxDD. However, given its marginal SharpeP5 (0.041 — the lowest survivor), apply a qualitative early-warning review at 8% drawdown (roughly 1.0× MaxDD). At 8%, do not automatically kill — manually assess whether recent price action reflects regime change. The 12.57% mechanical threshold remains the hard kill.

**Bootstrap WorstDD P95 interpretation:** These represent the 95th-percentile worst drawdown across 10,000 simulated trade orderings. They are stress-test benchmarks, not live kill-switch levels — they tell us what magnitude of drawdown is possible under unfavorable sequencing. INDHOTEL (55.8%) and COFORGE (53.4%) have the widest uncertainty; position sizing (section 3) accounts for this via vol-targeting.

---

## 3. Portfolio construction

### Vol-targeting

Target: 10% annualized volatility per instrument. ₹50,000 notional per instrument as baseline.

Each instrument is sized independently. Position size adjusts based on 60-day rolling realized volatility of the instrument to maintain the 10% annualized vol target. This means higher-vol instruments (COFORGE, INDHOTEL) will receive smaller nominal positions; lower-vol instruments (TORNTPHARM) will receive larger ones — but each will contribute equally to portfolio vol if uncorrelated.

### IT sector cap

PERSISTENT and COFORGE are both Nifty Midcap IT. When both are deployed simultaneously:
- Combined IT sector exposure capped at 33% of total portfolio, or ₹40,000 each (not ₹50,000 each).
- If only one IT instrument is active: standard ₹50,000 notional applies.
- Rationale: PERSISTENT/COFORGE full-period correlation of 0.36 rising to 0.47 in stress makes co-deployment concentration risk non-negligible at equal weight.

### Sizing summary

| Instrument     | Sector          | Base notional | IT cap applies? | Notes                              |
|----------------|----------------|---------------|-----------------|-------------------------------------|
| NSE:PERSISTENT  | IT             | ₹50,000       | Yes (cap ₹40k)  | Reduce if COFORGE also active       |
| NSE:TORNTPHARM  | Pharma         | ₹50,000       | No              | Lowest MaxDD; standard sizing       |
| NSE:COFORGE     | IT             | ₹50,000       | Yes (cap ₹40k)  | Review at 8% DD; marginal P5        |
| NSE:SUNDARMFIN  | NBFC           | ₹50,000       | No              | Standard sizing                     |
| NSE:INDHOTEL    | Hotels         | ₹50,000       | No              | Wide bootstrap DD P95; vol-target   |
| NSE:MUTHOOTFIN  | NBFC/Gold Fin  | ₹50,000       | No              | Longest duration history; standard  |

Max concurrent positions: 6. Total capital deployed at max: ₹300,000 (or ₹280,000 if both IT instruments active). This aligns with the ₹3L capital assumption used throughout the pipeline.

### Equal-weight vs. vol-target

Vol-targeting is preferred over pure equal-weight for this set. The bootstrap WorstDD P95 spread (TORNTPHARM at 25.75% vs. INDHOTEL at 55.79%) shows materially different tail risk profiles. Pure equal-weight would allocate the same capital to instruments with 2× the drawdown uncertainty. Vol-targeting equalizes the risk contribution; equal-weight equalizes the notional.

Decision: vol-target at 10% annualized per instrument. Equal-weight is acceptable as a fallback if vol data is unavailable for a given instrument at entry time.

---

## Decision summary

All 6 bootstrap survivors pass the correlation screen. No portfolio-level correlation kills. Kill-switch thresholds derived from 1.5× historical MaxDD per instrument. COFORGE carries an additional qualitative early-warning trigger at 8% drawdown. Portfolio uses vol-targeting at 10% annualized per instrument, ₹50k base notional, with IT sector cap (PERSISTENT + COFORGE combined ≤ 33% / ₹40k each when co-deployed).

**TASK-0097 acceptance criteria: all met. Pipeline complete for MACD crossover Nifty Midcap 150.**

## Gates passed before this stage

| Gate              | Result |
|-------------------|--------|
| Universe gate     | PASS — TASK-0091, DSRAvg=0.0885, 43/48 instruments |
| Walk-forward gate | PASS — TASK-0095, 25/43 at exact boundary |
| Bootstrap gate    | PASS — TASK-0096, 6/25 instruments pass SharpeP5>0 AND P(S>0)>80% |
| Correlation screen| PASS — TASK-0097, 0/15 pairs exceed 0.70 full-period |
