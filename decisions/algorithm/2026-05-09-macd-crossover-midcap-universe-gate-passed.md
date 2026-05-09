# MACD crossover passes universe gate — Nifty Midcap 150 (48 instruments)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | macd-crossover, universe-gate, DSR, nifty-midcap-150, midcap, walk-forward, TASK-0091, survivor |

## Context

TASK-0091 universe sweep for macd-crossover (fast=17, slow=26, signal=9 — unchanged from large-cap evaluation per 2026-05-04 decision) across 48 Nifty Midcap 150 instruments in `universes/nifty-midcap-liquid.yaml`, 2018-01-01 to 2024-12-31, `--commission zerodha_full`. This is the first universe sweep on the midcap universe. Prior large-cap result (TASK-0052): DSR avg Sharpe 0.2715, 14/15 instruments, 93.3% pass fraction.

28 instruments were missing from cache and pre-fetched via `cmd/fetch-history` before the sweep. LEMONTREE (IPO March 2018) was fetched from 2018-03-01; the engine handled ~2 months of missing early data gracefully. All 48 instruments returned `insufficient_data=false` with minimum 52 trades.

## Decision

**MACD crossover passes the Nifty Midcap 150 universe gate.** Advances to walk-forward on 43 positive-Sharpe instruments.

**Gate results (nTrials=48, DSR formula from analytics.DSR):**

- Sufficient instruments: 48/48 (all instruments have trade_count >= 30, insufficient_data=false)
- Positive raw Sharpe: 43/48 = 89.6% — passes >= 40%
- DSR-corrected average Sharpe (nTrials=48): **0.0885** — passes DSRAvg > 0
- Min trade count across all instruments: 52 (NSE:SUNDARMFIN, NSE:DEEPAKNTR)

**Per-instrument results (sorted by Sharpe):**

| Instrument | Raw Sharpe | Trades | DSR |
|---|---|---|---|
| NSE:PERSISTENT | 1.153449 | 60 | 0.8591 |
| NSE:TORNTPHARM | 0.785464 | 57 | 0.4834 |
| NSE:COFORGE | 0.770817 | 61 | 0.4790 |
| NSE:SUNDARMFIN | 0.744234 | 52 | 0.4277 |
| NSE:INDHOTEL | 0.719756 | 54 | 0.4092 |
| NSE:IPCALAB | 0.676686 | 54 | 0.3662 |
| NSE:ABCAPITAL | 0.655508 | 53 | 0.3420 |
| NSE:LTTS | 0.633745 | 56 | 0.3289 |
| NSE:MUTHOOTFIN | 0.619670 | 58 | 0.3202 |
| NSE:DEEPAKNTR | 0.612131 | 52 | 0.2956 |
| NSE:EIHOTEL | 0.592019 | 54 | 0.2815 |
| NSE:MOTHERSON | 0.578460 | 54 | 0.2679 |
| NSE:ENDURANCE | 0.557754 | 54 | 0.2472 |
| NSE:CHOLAFIN | 0.532381 | 58 | 0.2330 |
| NSE:CUMMINSIND | 0.513847 | 60 | 0.2195 |
| NSE:FEDERALBNK | 0.473595 | 54 | 0.1631 |
| NSE:SAIL | 0.472711 | 59 | 0.1759 |
| NSE:RADICO | 0.470788 | 60 | 0.1765 |
| NSE:VINATIORGA | 0.440203 | 55 | 0.1326 |
| NSE:ALKEM | 0.437729 | 57 | 0.1356 |
| NSE:RATNAMANI | 0.437593 | 54 | 0.1271 |
| NSE:BALKRISIND | 0.434133 | 55 | 0.1265 |
| NSE:THERMAX | 0.433025 | 58 | 0.1336 |
| NSE:MPHASIS | 0.427289 | 55 | 0.1197 |
| NSE:M&MFIN | 0.389628 | 58 | 0.0902 |
| NSE:PIIND | 0.386439 | 58 | 0.0870 |
| NSE:NMDC | 0.381383 | 59 | 0.0845 |
| NSE:PRESTIGE | 0.377607 | 62 | 0.0882 |
| NSE:SUNDRMFAST | 0.361767 | 64 | 0.0770 |
| NSE:BHEL | 0.337579 | 56 | 0.0328 |
| NSE:MHRIL | 0.316685 | 58 | 0.0173 |
| NSE:ELGIEQUIP | 0.301446 | 56 | -0.0034 |
| NSE:EXIDEIND | 0.292280 | 55 | -0.0154 |
| NSE:SCHAEFFLER | 0.284692 | 63 | -0.0024 |
| NSE:LEMONTREE | 0.263189 | 57 | -0.0389 |
| NSE:COLPAL | 0.260762 | 61 | -0.0311 |
| NSE:TATACONSUM | 0.228341 | 61 | -0.0635 |
| NSE:LICHSGFIN | 0.223241 | 56 | -0.0816 |
| NSE:NATIONALUM | 0.213610 | 58 | -0.0858 |
| NSE:AJANTPHARM | 0.199202 | 69 | -0.0749 |
| NSE:GODREJCP | 0.135882 | 53 | -0.1776 |
| NSE:LALPATHLAB | 0.081605 | 57 | -0.2205 |
| NSE:THOMASCOOK | 0.049690 | 53 | -0.2638 |
| NSE:OBEROIRLTY | -0.039293 | 63 | -0.3264 |
| NSE:NAVINFLUOR | -0.081951 | 69 | -0.3561 |
| NSE:EMAMILTD | -0.135195 | 55 | -0.4428 |
| NSE:DABUR | -0.152634 | 56 | -0.4575 |
| NSE:PHOENIXLTD | -0.167338 | 70 | -0.4395 |

**Marginal-ADV instruments (BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM, SAIL):** All five show positive raw Sharpe (0.21–0.47) and sufficient trade counts (55–63 trades). None triggered `insufficient_data=true`. All five remain eligible for walk-forward. No replacements needed for universe v1.

**ABCAPITAL data check:** IPO 2017, trade_count=53, Sharpe=0.655, insufficient_data=false. History reaches 2018-01-01 cleanly — no data quality issue.

**LEMONTREE data check:** IPO March 2018, fetched from 2018-03-01. trade_count=57, Sharpe=0.263, insufficient_data=false. Engine handled ~2 months of missing data gracefully; result is valid.

**Eligible for walk-forward (positive raw Sharpe):**
NSE:PERSISTENT, NSE:TORNTPHARM, NSE:COFORGE, NSE:SUNDARMFIN, NSE:INDHOTEL, NSE:IPCALAB, NSE:ABCAPITAL, NSE:LTTS, NSE:MUTHOOTFIN, NSE:DEEPAKNTR, NSE:EIHOTEL, NSE:MOTHERSON, NSE:ENDURANCE, NSE:CHOLAFIN, NSE:CUMMINSIND, NSE:FEDERALBNK, NSE:SAIL, NSE:RADICO, NSE:VINATIORGA, NSE:ALKEM, NSE:RATNAMANI, NSE:BALKRISIND, NSE:THERMAX, NSE:MPHASIS, NSE:M&MFIN, NSE:PIIND, NSE:NMDC, NSE:PRESTIGE, NSE:SUNDRMFAST, NSE:BHEL, NSE:MHRIL, NSE:ELGIEQUIP, NSE:EXIDEIND, NSE:SCHAEFFLER, NSE:LEMONTREE, NSE:COLPAL, NSE:TATACONSUM, NSE:LICHSGFIN, NSE:NATIONALUM, NSE:AJANTPHARM, NSE:GODREJCP, NSE:LALPATHLAB, NSE:THOMASCOOK (43 instruments)

Excluded from walk-forward (negative raw Sharpe): NSE:OBEROIRLTY, NSE:NAVINFLUOR, NSE:EMAMILTD, NSE:DABUR, NSE:PHOENIXLTD (5 instruments — no basis for further validation).

## Consequences

- MACD crossover advances to walk-forward on 43 eligible midcap instruments.
- The DSR avg drops from 0.2715 (large-cap, nTrials=15) to 0.0885 (midcap, nTrials=48). This is expected: the heavier multiple-testing penalty from 48 trials compresses DSR even when raw Sharpe distribution looks strong. The edge generalises across sectors (IT, pharma, hospitality, financials, chemicals, industrials).
- The 5 negative-Sharpe instruments are concentrated in: real estate/infrastructure (OBEROIRLTY, PHOENIXLTD, PRESTIGE barely positive), chemicals (NAVINFLUOR negative while VINATIORGA positive), and FMCG (EMAMILTD, DABUR). Sector-level signal: MACD trend-following works better on growth/momentum sectors than defensive FMCG in midcap.
- Marginal-ADV instruments BHEL and SAIL (PSU, cyclical) show decent Sharpe (0.34, 0.47) — worth watching in walk-forward to see if regime concentration explains their performance.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate applied here, thresholds unchanged
- [MACD parameters unchanged at 17/26/9](./2026-05-04-macd-parameters-unchanged-17-26-9.md) — parameters applied in this sweep
- [Nifty Midcap 150 universe — instrument list and gate thresholds](./2026-05-07-nifty-midcap-150-universe-instrument-list.md) — universe definition
- [MACD crossover passes universe gate — large-cap (TASK-0052)](./2026-05-03-macd-crossover-universe-gate-passed.md) — prior large-cap baseline
- [nTrials=48 for midcap DSR correction](./2026-05-09-ntrials-48-for-midcap-dsr-correction.md) — methodology decision for this run

## Revisit trigger

If walk-forward on the 43 eligible instruments retains fewer than 60% (the gate from the large-cap walk-forward revision), return here to assess whether the universe sweep DSRAvg of 0.0885 was meaningful or whether sector concentration (IT, hospitality) accounts for most of it.
