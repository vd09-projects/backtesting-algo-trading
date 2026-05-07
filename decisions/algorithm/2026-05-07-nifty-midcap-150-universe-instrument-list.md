# Nifty Midcap 150 universe — instrument list and gate thresholds

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | [midcap, universe, instrument-selection, nifty-midcap-150, gate-thresholds, evaluation-methodology, TASK-0072] |

## Context

All prior strategy evaluation ran on the 15 Nifty50 large-cap instruments in `universes/nifty50-large-cap.yaml`. The edge thesis is weaker on large-caps: heavy analyst coverage, high institutional participation, and efficient price discovery reduce the persistence of behavioral inefficiencies. Midcap names (Nifty Midcap 150) have thinner analyst coverage, more retail participation, and less efficient price discovery — conditions more likely to sustain exploitable patterns in trend-following and momentum strategies. TASK-0072 extended the pipeline to a midcap universe using the same daily-bar infrastructure.

## Decision

Marcus approved an initial 25-instrument candidate list for `universes/nifty-midcap-liquid.yaml` across 9 sectors on 2026-05-07. The list was subsequently expanded to **48 instruments across 10 sectors** on 2026-05-07 to improve test coverage and sector variety. **Marcus re-review of the 23 new additions is pending before TASK-0091 runs.**

Full instrument list:

- **Financials (7):** NSE:MUTHOOTFIN, NSE:CHOLAFIN, NSE:ABCAPITAL, NSE:LICHSGFIN, NSE:FEDERALBNK, NSE:SUNDARMFIN, NSE:M&MFIN
- **Consumer/FMCG (6):** NSE:GODREJCP, NSE:COLPAL, NSE:DABUR, NSE:EMAMILTD, NSE:RADICO, NSE:TATACONSUM
- **Industrials/Capital Goods (5):** NSE:BHEL, NSE:CUMMINSIND, NSE:SCHAEFFLER, NSE:ELGIEQUIP, NSE:THERMAX
- **Auto/Auto Ancillary (5):** NSE:BALKRISIND, NSE:MOTHERSON, NSE:EXIDEIND, NSE:SUNDRMFAST, NSE:ENDURANCE
- **IT/Tech (4):** NSE:MPHASIS, NSE:COFORGE, NSE:LTTS, NSE:PERSISTENT
- **Pharma/Healthcare (5):** NSE:TORNTPHARM, NSE:ALKEM, NSE:LALPATHLAB, NSE:AJANTPHARM, NSE:IPCALAB
- **Chemicals (4):** NSE:PIIND, NSE:DEEPAKNTR, NSE:VINATIORGA, NSE:NAVINFLUOR
- **Infrastructure/Real Estate (3):** NSE:OBEROIRLTY, NSE:PRESTIGE, NSE:PHOENIXLTD
- **Metals/Materials (4):** NSE:NATIONALUM, NSE:NMDC, NSE:SAIL, NSE:RATNAMANI
- **Hotels/Hospitality (5):** NSE:LEMONTREE, NSE:EIHOTEL, NSE:INDHOTEL, NSE:MHRIL, NSE:THOMASCOOK

**Gate thresholds are unchanged** from the large-cap pipeline (cross-instrument proliferation gate, 2026-04-25): DSR-corrected average Sharpe > 0, ≥ 40% pass fraction, ≥ 30 trades minimum per instrument. Changing thresholds retroactively for a new universe would undermine their purpose as pre-committed gates.

## Consequences

- **Marcus re-review required** for the 23 new additions before TASK-0091 runs.
- TASK-0091 runs the first universe sweep; gate results will identify which instruments produce sufficient trades and positive DSR-corrected Sharpe.
- ABCAPITAL (IPO 2017) and LEMONTREE (IPO Mar 2018) are flagged for Kite history verification before sweep.
- INDHOTEL (Indian Hotels/Taj) may have graduated to large-cap by 2026 — verify index membership before sweep.
- Marginal ADV names (BHEL, SCHAEFFLER, EXIDEIND, NATIONALUM, SAIL) are first candidates for replacement if `insufficient_data=true` appears in sweep results.

## Related decisions

- [Cross-instrument universe gate supersedes single-instrument proliferation gate](./2026-04-25-cross-instrument-proliferation-gate.md) — the gate being applied here unchanged
- [ADV floor for midcap daily-bar universe](./2026-05-07-adv-floor-midcap-daily-bar-universe.md) — liquidity screening applied to this list

## Revisit trigger

After TASK-0091 sweep results: if more than 40% of instruments show `insufficient_data=true` or the DSR-corrected average Sharpe is severely negative (< -0.3), revisit the instrument list and consider raising the ADV floor or excluding PSU names with compressed liquidity.
