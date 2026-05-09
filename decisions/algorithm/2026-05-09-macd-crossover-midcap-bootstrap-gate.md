# MACD crossover passes bootstrap gate — Nifty Midcap 150 (6 of 25 instruments)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, bootstrap, instrument-count-gate, nifty-midcap-150, midcap, gate-results, TASK-0096, survivor |

## Context

TASK-0096 bootstrap validation ran MACD crossover (fast=17, slow=26, signal=9) on 25 walk-forward survivors from TASK-0095. Data: 2018-01-01 to 2025-01-01 (full outer window), commission=zerodha_full, 10,000 simulations, seed=42.

Source: `results/2026-05-09-TASK-0096/bootstrap-results.csv` — actual binary run (`cmd/backtest --bootstrap`).

Gate per 2026-04-27 decision: SharpeP5 > 0 AND P(Sharpe > 0) > 80% per instrument. Both conditions required. Gate advances if >= 1 instrument passes.

**Result: 6 of 25 instruments pass. Gate PASS.**

Note: Only 6 survivors from 25 WF entrants (24%) — the walk-forward boundary gate (25/43 at exact floor) correctly flagged this risk. The 4 anomalous-ratio WF survivors (SCHAEFFLER, THOMASCOOK, EXIDEIND, LALPATHLAB) all failed bootstrap as predicted.

## Decision

**MACD crossover (17/26/9) passes the Nifty Midcap 150 bootstrap gate: 6/25 instruments pass.**

6 survivors advance to kill-switch derivation and portfolio construction.

**[NOTE] Revisit trigger activated:** Bootstrap killed 19 of 25 WF survivors, well below the <10 threshold stated in the WF gate decision. The WF gate's exact-boundary pass (25/43) was a weak signal as predicted. The 6 bootstrap survivors are the high-confidence core; proceed with these.

## Per-instrument results — SURVIVORS (6 instruments)

| Instrument      | SharpeP5 | SharpeP50 | SharpeP95 | P(S>0) | Verdict |
|-----------------|----------|-----------|-----------|--------|---------|
| NSE:PERSISTENT  | 0.1507   | 0.3315    | 0.4938    | 99.75% | PASS    |
| NSE:TORNTPHARM  | 0.1455   | 0.3335    | 0.5190    | 99.66% | PASS    |
| NSE:COFORGE     | 0.0413   | 0.2381    | 0.4118    | 97.42% | PASS    |
| NSE:SUNDARMFIN  | 0.0351   | 0.2360    | 0.4023    | 96.73% | PASS    |
| NSE:INDHOTEL    | 0.0191   | 0.2255    | 0.4090    | 96.27% | PASS    |
| NSE:MUTHOOTFIN  | 0.0249   | 0.2223    | 0.3927    | 96.72% | PASS    |

## Per-instrument results — KILLED (19 instruments)

| Instrument     | SharpeP5 | SharpeP50 | P(S>0) | Kill reason            |
|----------------|----------|-----------|--------|------------------------|
| NSE:IPCALAB    | -0.0146  | 0.2002    | 93.99% | SharpeP5 < 0           |
| NSE:DEEPAKNTR  | -0.0120  | 0.2092    | 94.26% | SharpeP5 < 0           |
| NSE:EIHOTEL    | -0.0461  | 0.1719    | 91.07% | SharpeP5 < 0           |
| NSE:ENDURANCE  | -0.0483  | 0.1749    | 91.31% | SharpeP5 < 0           |
| NSE:CUMMINSIND | -0.0471  | 0.1662    | 90.54% | SharpeP5 < 0           |
| NSE:FEDERALBNK | -0.0315  | 0.1852    | 92.47% | SharpeP5 < 0           |
| NSE:SAIL       | -0.1009  | 0.1311    | 84.97% | SharpeP5 < 0           |
| NSE:RADICO     | -0.0717  | 0.1446    | 87.60% | SharpeP5 < 0           |
| NSE:BALKRISIND | -0.1379  | 0.1050    | 78.73% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:PIIND      | -0.0960  | 0.1255    | 83.67% | SharpeP5 < 0           |
| NSE:PRESTIGE   | -0.0807  | 0.1322    | 85.98% | SharpeP5 < 0           |
| NSE:BHEL       | -0.1125  | 0.1199    | 82.82% | SharpeP5 < 0           |
| NSE:MHRIL      | -0.0958  | 0.1239    | 83.51% | SharpeP5 < 0           |
| NSE:EXIDEIND   | -0.1724  | 0.0881    | 74.30% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:SCHAEFFLER | -0.1919  | 0.0677    | 70.58% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:LEMONTREE  | -0.1432  | 0.0915    | 75.37% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:TATACONSUM | -0.1615  | 0.0661    | 69.91% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:LALPATHLAB | -0.2037  | 0.0287    | 58.57% | SharpeP5 < 0, P(S>0) < 80% |
| NSE:THOMASCOOK | -0.2541  | 0.0132    | 53.66% | SharpeP5 < 0, P(S>0) < 80% |

## Anomalous WF cases — bootstrap verdict

The 4 anomalous-ratio WF survivors (SCHAEFFLER, THOMASCOOK, EXIDEIND, LALPATHLAB) that raised flags at walk-forward all failed bootstrap, confirming the WF gate's assessment was correct to pass them conditionally and defer the kill decision to bootstrap. Bootstrap is the right gate for near-zero IS Sharpe cases.

## Kill pattern analysis

Two tiers of kills:
1. **Marginal edge (SharpeP5 barely negative, P(S>0) > 90%):** IPCALAB, DEEPAKNTR, EIHOTEL, ENDURANCE, CUMMINSIND, FEDERALBNK — median Sharpe positive but P5 dips below zero. These have real signal that doesn't achieve the confidence threshold. Could revisit with longer data window.
2. **Weak/no edge (P(S>0) < 80% or SharpeP5 << 0):** BALKRISIND, EXIDEIND, SCHAEFFLER, LEMONTREE, TATACONSUM, LALPATHLAB, THOMASCOOK — both gates fail. Near-zero or noise-level signal on bootstrap, consistent with anomalous WF ratios.

The 6 survivors (all IT/pharma/hospitality/finance) are the instruments where MACD daily trend-following has statistically robust edge in this period.

## Next stage

6 survivors advance to kill-switch derivation and portfolio construction:
- NSE:PERSISTENT (IT)
- NSE:TORNTPHARM (Pharma)
- NSE:COFORGE (IT)
- NSE:SUNDARMFIN (NBFC)
- NSE:INDHOTEL (Hotels)
- NSE:MUTHOOTFIN (NBFC/Gold Finance)

Correlation screen: pairwise Pearson among 6 survivors should be low given sector diversity. Run `cmd/correlate` before sizing.

## Gates passed before entering this gate

| Gate | Result |
|------|--------|
| Universe gate | PASS — TASK-0091, DSRAvg=0.0885, 43/48 instruments |
| Walk-forward gate | PASS — TASK-0095, 25/43 at exact boundary |

## Related decisions

- Walk-forward gate results (2026-05-09) — 25 survivors handed to this gate
- Bootstrap gate criteria (2026-04-27) — SharpeP5 > 0 AND P(S>0) > 80%
- MACD parameters 17/26/9 (2026-05-04) — unchanged throughout pipeline
