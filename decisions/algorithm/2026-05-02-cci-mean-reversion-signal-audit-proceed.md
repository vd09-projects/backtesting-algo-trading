---
date: 2026-05-02
topic: CCI mean-reversion signal frequency audit
category: algorithm
status: accepted
tags: [cci, mean-reversion, signal-audit, pre-pipeline-gate, TASK-0065]
owner: marcus
---

# CCI Mean-Reversion — Signal Frequency Audit: PROCEED

## Verdict: PROCEED to full pipeline

Both gate conditions satisfied. CCI mean-reversion clears the pre-pipeline diagnostic.

## Audit parameters

- Strategy: CCI < -100 entry, cross above 0 exit, long-only, daily bars
- CCI period: 20 (standard)
- Universe: 15 Nifty50 large-caps (`universes/nifty50-large-cap.yaml`)
- Period: 2018-01-01 → 2024-01-01
- COVID window: 2020-01-01 – 2020-06-30 (Q1-Q2 2020)

## Results

| Instrument     | Trades | COVID trades | COVID % |
|----------------|--------|--------------|---------|
| NSE:RELIANCE   | 42     | 3            | 7.1%    |
| NSE:INFY       | 32     | 1            | 3.1%    |
| NSE:TCS        | 34     | 2            | 5.9%    |
| NSE:HDFCBANK   | 46     | 5            | 10.9%   |
| NSE:ICICIBANK  | 36     | 5            | 13.9%   |
| NSE:KOTAKBANK  | 40     | 4            | 10.0%   |
| NSE:SBIN       | 34     | 5            | 14.7%   |
| NSE:AXISBANK   | 35     | 5            | 14.3%   |
| NSE:LT         | 29     | 3            | 10.3%   |
| NSE:HINDUNILVR | 32     | 3            | 9.4%    |
| NSE:ITC        | 39     | 4            | 10.3%   |
| NSE:BAJFINANCE | 33     | 4            | 12.1%   |
| NSE:MARUTI     | 30     | 3            | 10.0%   |
| NSE:TITAN      | 33     | 3            | 9.1%    |
| NSE:WIPRO      | 35     | 2            | 5.7%    |

**Avg trades/instrument: 35.3** (pass threshold: ≥25) ✅  
**COVID clustering violations: 0/15** (pass threshold: no instrument >30%) ✅

## Gate assessment

- Signal frequency gate: PASS. Avg 35.3 trades well above 25-trade floor. Min instrument (NSE:LT) = 29 trades — all 15 above floor.
- Clustering gate: PASS. Max COVID concentration is 14.7% (NSE:SBIN), well below 30% threshold. Edge is not a single-event artifact.
- Note: 55/105 cells across ALL 7 strategies excluded at the 30-trade threshold (full audit). CCI mean-reversion specifically: 0 instruments excluded.

## Next step

Proceed to universe sweep (TASK-0052). CCI mean-reversion joins the pipeline as the 7th strategy candidate alongside the original six. Universe sweep applies DSR-corrected average Sharpe gate across all 15 instruments.

## Prior context

- Marcus iterate verdict: `workflows/sessions/2026-05-02-evaluate-cci-mean-reversion.json`
- Pre-committed tiebreaker at correlation gate: prefer Bollinger over CCI if both survive (adaptive bands vs fixed constant — see Marcus evaluate session 2026-05-02)
