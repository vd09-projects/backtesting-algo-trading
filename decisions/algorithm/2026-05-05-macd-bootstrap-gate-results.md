# MACD crossover (17/26/9) bootstrap gate results — 4 survivors, 5 killed

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-05       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | bootstrap, macd-crossover, gate-results, TASK-0069, evaluation-pipeline |

## Context

MACD crossover (fast=17, slow=26, signal=9) advanced to bootstrap after the walk-forward instrument-count gate was revised from 100% to 60% retention in a concurrent decision (see related decisions below). Bootstrap was run on the 9 instrument pairs that passed walk-forward: SBIN, BAJFINANCE, TITAN, LT, ICICIBANK, INFY, AXISBANK, ITC, KOTAKBANK. The bootstrap gate (SharpeP5 > 0 AND P(Sharpe > 0) > 80%) was applied per the 2026-04-27 bootstrap gate decision. 10,000 simulations, seed=42.

## Bootstrap results

| Instrument | SharpeP5 | SharpeP50 | SharpeP95 | P(Sharpe > 0) | Gate 1 (P5 > 0) | Gate 2 (Prob > 80%) | Verdict |
|---|---|---|---|---|---|---|---|
| SBIN       | 0.0719   | 0.3195    | 0.5551    | 98.0%         | PASS            | PASS                | GO      |
| BAJFINANCE | 0.0467   | 0.2526    | 0.4171    | 97.3%         | PASS            | PASS                | GO      |
| TITAN      | 0.0854   | 0.3102    | 0.5323    | 98.7%         | PASS            | PASS                | GO      |
| ICICIBANK  | 0.0229   | 0.2489    | 0.4579    | 96.2%         | PASS            | PASS                | GO      |
| LT         | -0.0484  | 0.1870    | 0.3727    | 91.2%         | FAIL            | PASS                | KILL    |
| INFY       | -0.0501  | 0.1839    | 0.3777    | 90.9%         | FAIL            | PASS                | KILL    |
| AXISBANK   | -0.1484  | 0.1079    | 0.3029    | 78.4%         | FAIL            | FAIL                | KILL    |
| ITC        | -0.1806  | 0.0763    | 0.2885    | 70.6%         | FAIL            | FAIL                | KILL    |
| KOTAKBANK  | -0.3146  | -0.0003   | 0.1827    | 49.9%         | FAIL            | FAIL                | KILL    |

## Decision

**4 survivors advance: NSE:SBIN, NSE:BAJFINANCE, NSE:TITAN, NSE:ICICIBANK.**

**5 instruments killed at bootstrap gate: LT, INFY, AXISBANK, ITC, KOTAKBANK.**

The gate is applied per instrument, not per strategy — MACD crossover as a strategy is live on the 4 surviving instrument pairs. The 5 kills are permanent for this evaluation cycle; they do not get a second chance at a later gate.

## Kill classification notes

LT and INFY are gate kills only — not thesis kills. Both show P(Sharpe > 0) > 90% (91.2% and 90.9% respectively), and their P50 bootstrap Sharpes are positive (0.187 and 0.184). The edge thesis is not disproven on these instruments. The kill is on the technical criterion: SharpeP5 < 0 means the left tail of the bootstrap distribution crosses zero, making a negative Sharpe outcome non-negligible. Under the current bootstrap gate, that is sufficient to exclude from live deployment. If a future evaluation cycle revises the SharpeP5 floor or adds a conditional approval for instruments where P5 is borderline negative but P50 is strongly positive, LT and INFY are the candidates to revisit first.

AXISBANK, ITC, and KOTAKBANK fail both gates. These are not borderline — the entire bootstrap distribution is shifted left. KOTAKBANK's P50 is effectively zero (-0.0003); its SharpeP5 of -0.3146 indicates the strategy has no recoverable edge there. These kills stand regardless of any future threshold revision.

## Survivors entering the portfolio stage

The 4 survivors carry the following bootstrap metrics into TASK-0054 (correlation and portfolio construction):

| Instrument | SharpeP5 | SharpeP50 | SharpeP95 | P(Sharpe > 0) |
|---|---|---|---|---|
| SBIN       | 0.0719   | 0.3195    | 0.5551    | 98.0%         |
| BAJFINANCE | 0.0467   | 0.2526    | 0.4171    | 97.3%         |
| TITAN      | 0.0854   | 0.3102    | 0.5323    | 98.7%         |
| ICICIBANK  | 0.0229   | 0.2489    | 0.4579    | 96.2%         |

All four are in the same sector bucket (financials/diversified financials + consumer discretionary). Correlation between SBIN and ICICIBANK, and between SBIN and BAJFINANCE, is likely elevated — both are banking names. TASK-0054 must assess pairwise correlation before sizing; do not assume these are uncorrelated return streams.

## What this does not unlock

MACD on TCS, RELIANCE, HINDUNILVR, WIPRO, HDFCBANK was killed at walk-forward. Those 5 do not re-enter at bootstrap — the walk-forward gate is final per instrument.

MACD on LT, INFY, AXISBANK, ITC, KOTAKBANK is now killed at bootstrap. Those 5 instruments are permanently excluded from MACD deployment in this evaluation cycle.

## Related decisions

- [Walk-forward instrument-count gate revised to 60% (2026-05-05)](./2026-05-05-walk-forward-instrument-count-gate-relaxed.md) — gate change that allowed MACD to advance to bootstrap; without this revision MACD would have been killed at walk-forward
- [Bootstrap gate design (2026-04-27)](./2026-04-27-bootstrap-gate.md) — the gate criteria applied here: SharpeP5 > 0 AND P(Sharpe > 0) > 80%
- [MACD walk-forward instrument-count gate kill (2026-05-04)](./2026-05-04-macd-crossover-walk-forward-instrument-count-gate.md) — original kill record (100% retention threshold), superseded by gate revision for MACD

## Revisit trigger

If LT or INFY are re-evaluated under a future methodology that introduces a conditional approval path for borderline-negative P5 with strongly positive P50, the gate evidence here is the baseline. Do not treat this kill as a thesis rejection for those two instruments.
