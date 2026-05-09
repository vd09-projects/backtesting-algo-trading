# MACD crossover passes walk-forward instrument-count gate — Nifty Midcap 150 (25 of 43 instruments)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, walk-forward, instrument-count-gate, nifty-midcap-150, midcap, gate-results, bootstrap, TASK-0095, survivor |

## Context

TASK-0095 walk-forward validation ran MACD crossover (fast=17, slow=26, signal=9 — parameters unchanged per 2026-05-04 decision) on 43 eligible instruments from the TASK-0091 midcap universe gate handoff. Data: 2018-01-01 to 2024-12-31, standard walk-forward configuration (2yr IS / 1yr OOS / 1yr step), commission=zerodha_full.

Source: `results/2026-05-09-TASK-0095/wf-results.csv` — actual binary run (`cmd/walk-forward`), all 43 instruments.

The instrument-count gate (from the 2026-05-05 revision): >= floor(0.60 * 43) = 25 instruments must pass the per-instrument walk-forward gate (OverfitFlag=false AND NegativeFoldFlag=false).

**Result: 25 of 43 instruments pass. Gate PASS at exact boundary (25 = floor(0.60 * 43) = 25).**

Note: 25/43 = 58.1%, which equals the integer floor threshold (25 instruments). The gate criteria is `WF_passes >= floor(0.60 * N)`, which evaluates to `25 >= 25` — a strict pass. Exact-boundary gate verdicts are logged as [FLAGGED] but continue as PASS.

## Decision

**MACD crossover (17/26/9) passes the Nifty Midcap 150 walk-forward instrument-count gate: 25/43 instruments pass.**

Advances to bootstrap gate on 25 surviving instruments.

## Per-instrument results — SURVIVORS (25 instruments)

| Instrument       | Avg IS Sharpe | Avg OOS Sharpe | OOSISRatio | NegFolds/5 | Verdict |
|------------------|---------------|----------------|------------|------------|---------|
| NSE:INDHOTEL     | 0.2276        | 0.3834         | 1.6845     | 0          | PASS    |
| NSE:SUNDARMFIN   | 0.0983        | 0.3813         | 3.8805     | 1          | PASS    |
| NSE:PRESTIGE     | 0.1710        | 0.3763         | 2.2001     | 1          | PASS    |
| NSE:PERSISTENT   | 0.2748        | 0.3414         | 1.2424     | 1          | PASS    |
| NSE:BHEL         | 0.2664        | 0.3343         | 1.2548     | 1          | PASS    |
| NSE:CUMMINSIND   | 0.1813        | 0.3221         | 1.7764     | 1          | PASS    |
| NSE:RADICO       | 0.0948        | 0.2682         | 2.8291     | 0          | PASS    |
| NSE:THOMASCOOK   | -0.0092       | 0.2496         | -27.0342   | 0          | PASS*   |
| NSE:ENDURANCE    | 0.1575        | 0.2648         | 1.6816     | 0          | PASS    |
| NSE:TORNTPHARM   | 0.2806        | 0.2810         | 1.0014     | 1          | PASS    |
| NSE:EIHOTEL      | 0.1657        | 0.2260         | 1.3641     | 1          | PASS    |
| NSE:MHRIL        | 0.2304        | 0.2164         | 0.9391     | 1          | PASS    |
| NSE:FEDERALBNK   | 0.2315        | 0.2133         | 0.9214     | 1          | PASS    |
| NSE:MUTHOOTFIN   | 0.2011        | 0.2069         | 1.0290     | 1          | PASS    |
| NSE:LALPATHLAB   | 0.0226        | 0.1918         | 8.4766     | 1          | PASS*   |
| NSE:TATACONSUM   | 0.1117        | 0.1729         | 1.5480     | 1          | PASS    |
| NSE:PIIND        | 0.1495        | 0.1716         | 1.1479     | 1          | PASS    |
| NSE:DEEPAKNTR    | 0.1686        | 0.1652         | 0.9797     | 1          | PASS    |
| NSE:COFORGE      | 0.1496        | 0.1909         | 1.2760     | 1          | PASS    |
| NSE:IPCALAB      | 0.1466        | 0.1841         | 1.2558     | 1          | PASS    |
| NSE:BALKRISIND   | 0.1092        | 0.2196         | 2.0112     | 1          | PASS    |
| NSE:SAIL         | 0.1176        | 0.1798         | 1.5289     | 1          | PASS    |
| NSE:EXIDEIND     | 0.0034        | 0.1457         | 42.3317    | 1          | PASS*   |
| NSE:SCHAEFFLER   | -0.0956       | 0.1243         | -1.3004    | 0          | PASS*   |
| NSE:LEMONTREE    | 0.1153        | 0.0698         | 0.6055     | 1          | PASS    |

*Anomalous OOSISRatio due to near-zero or negative IS Sharpe denominator — see anomalous cases note below.

OOS Sharpe range: +0.070 (LEMONTREE) to +0.383 (INDHOTEL).

## Per-instrument results — KILLED (18 instruments)

### Killed: NegativeFoldFlag only (10 instruments)

| Instrument      | Avg OOS Sharpe | Avg IS Sharpe | OOSISRatio | NegFolds/5 |
|-----------------|----------------|---------------|------------|------------|
| NSE:ABCAPITAL   | 0.1654         | 0.3020        | 0.5478     | 2          |
| NSE:LTTS        | 0.0705         | 0.1390        | 0.5070     | 2          |
| NSE:MOTHERSON   | 0.1474         | 0.0754        | 1.9551     | 2          |
| NSE:CHOLAFIN    | 0.1880         | 0.1159        | 1.6226     | 2          |
| NSE:ALKEM       | 0.1675         | 0.0673        | 2.4890     | 2          |
| NSE:THERMAX     | 0.0598         | 0.0882        | 0.6786     | 2          |
| NSE:MPHASIS     | 0.1186         | 0.0614        | 1.9300     | 3          |
| NSE:M&MFIN      | 0.1364         | 0.1499        | 0.9103     | 2          |
| NSE:SUNDRMFAST  | 0.0781         | 0.1364        | 0.5725     | 2          |
| NSE:ELGIEQUIP   | 0.1353         | 0.2105        | 0.6429     | 2          |

### Killed: OverfitFlag only (1 instrument)

| Instrument     | Avg OOS Sharpe | Avg IS Sharpe | OOSISRatio | NegFolds/5 |
|----------------|----------------|---------------|------------|------------|
| NSE:LICHSGFIN  | 0.0715         | 0.1617        | 0.4424     | 1          |

### Killed: Both NegativeFoldFlag AND OverfitFlag (7 instruments)

| Instrument      | Avg OOS Sharpe | Avg IS Sharpe | OOSISRatio | NegFolds/5 |
|-----------------|----------------|---------------|------------|------------|
| NSE:VINATIORGA  | 0.0518         | 0.2218        | 0.2336     | 3          |
| NSE:RATNAMANI   | -0.0623        | 0.1339        | -0.4652    | 3          |
| NSE:NMDC        | 0.0925         | 0.2291        | 0.4039     | 2          |
| NSE:COLPAL      | -0.0569        | -0.0206       | 2.7648     | 3          |
| NSE:NATIONALUM  | -0.0300        | -0.0234       | 1.2819     | 3          |
| NSE:AJANTPHARM  | -0.1251        | -0.1430       | 0.8749     | 2          |
| NSE:GODREJCP    | -0.1555        | 0.1257        | -1.2373    | 3          |

## Anomalous cases note

Four survivors (THOMASCOOK, LALPATHLAB, EXIDEIND, SCHAEFFLER) show near-zero or negative IS Sharpe, producing extreme or undefined OOSISRatio values. Gate logic is applied as-coded: OverfitFlag=false AND NegativeFoldFlag=false = PASS. OOS Sharpe is positive in all four cases. Bootstrap is the appropriate gate for assessing whether this OOS signal is statistically robust — if SharpeP5 < 0, that is the kill signal.

## Kill clustering analysis

NegativeFoldFlag kills concentrate in NBFCs/financials (ABCAPITAL, CHOLAFIN, M&MFIN), IT services (LTTS, MPHASIS), auto ancillaries (MOTHERSON, ELGIEQUIP, SUNDRMFAST), and specialty chemicals (ALKEM, THERMAX). OverfitFlag kills concentrate in metals/mining (NMDC, NATIONALUM, RATNAMANI), FMCG (COLPAL, GODREJCP), and consumer (VINATIORGA, AJANTPHARM). Pattern consistent with MACD daily-bar trend-following capturing edge in capital goods, pharma, and hotels but not in commodity/metals, defensives, or high-churn NBFCs.

## Survivor sector distribution

- IT/Tech: PERSISTENT, COFORGE
- Pharma: TORNTPHARM, IPCALAB, LALPATHLAB
- Hotels/Hospitality: INDHOTEL, EIHOTEL, MHRIL, THOMASCOOK
- Financials: MUTHOOTFIN, FEDERALBNK
- Chemicals/Diversified: DEEPAKNTR
- Industrials/Capital Goods: CUMMINSIND, ENDURANCE, BHEL, SAIL, BALKRISIND, EXIDEIND, SCHAEFFLER
- Real Estate: PRESTIGE
- Consumer: RADICO, TATACONSUM, LEMONTREE
- Financials (NBFC): SUNDARMFIN
- Pharma/Agrochem: PIIND

Healthy sector diversity. No single sector dominates — reduces sector-cluster correlation risk at bootstrap.

## Bootstrap gate criteria (next stage)

25 surviving instruments advance to bootstrap (TASK-0096). Bootstrap gate per 2026-04-27 decision:
- SharpeP5 > 0 (5th percentile of bootstrap per-trade Sharpe distribution > 0)
- P(Sharpe > 0) > 80% (at least 80% of 10,000 simulations show positive per-trade Sharpe)
- Both conditions must hold per instrument
- Bootstrap parameters: 10,000 simulations, seed=42

## Gates passed before entering this gate

| Gate | Result |
|------|--------|
| Universe gate (DSRAvg > 0, PassFraction >= 40%) | PASS — TASK-0091, DSRAvg=0.0885, 43/48=89.6% |

## Related decisions

- Walk-forward instrument-count gate revised to 60% (2026-05-05) — gate threshold applied: floor(0.60 * 43) = 25
- Walk-forward OOS/IS Sharpe threshold (2026-04-22) — per-instrument pass/fail criteria
- MACD parameters unchanged at 17/26/9 (2026-05-04) — parameters applied in this run
- Bootstrap gate (2026-04-27) — next gate criteria

## Revisit trigger

If bootstrap kills a substantial fraction of the 25 survivors (< 10 pass), revisit whether 25 at the exact gate boundary was meaningful signal. The four anomalous-ratio survivors (THOMASCOOK, LALPATHLAB, EXIDEIND, SCHAEFFLER) are the highest-risk candidates at bootstrap.
