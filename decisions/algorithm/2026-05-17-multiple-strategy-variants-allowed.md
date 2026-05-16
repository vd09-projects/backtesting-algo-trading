---
id: 2026-05-17-multiple-strategy-variants-allowed
title: "Multiple parameter variants per strategy are allowed; evaluation pipeline decides survivors"
date: 2026-05-17
status: accepted
category: algorithm
tags: [parameter-variants, strategy-naming, evaluation, 5min]
---

## Decision

When porting a strategy to a new timeframe (e.g., daily → 5-min), create 2–3 named variants with different parameter calibrations rather than forcing a single choice. Register each as a separate strategy name in `GlobalRegistry`. Run each through the full evaluation pipeline independently. If multiple variants survive all gates, keep them all as distinct strategies.

## Rationale

Parameter calibration for a new timeframe involves genuine uncertainty — we don't know a priori whether a "fast" MACD or a "slow" MACD will have a better edge at 5-min resolution. Rather than making an arbitrary upfront choice, we let the evaluation pipeline (universe gate + walk-forward + bootstrap) resolve the question empirically. The cost of running 3 variants vs 1 is roughly 3× evaluation time, which is acceptable given the mechanical nature of the pipeline.

## Naming convention

`{strategy}-{timeframe}-{variant}` — examples:
- `macd-5min-fast` (9/21/9)
- `macd-5min-standard` (12/26/9)
- `macd-5min-slow` (26/52/18)
- `rsi-5min-14` (period=14, standard)
- `rsi-5min-7` (period=7, more sensitive)

## Constraints

- Maximum 3 variants per strategy per timeframe to avoid proliferation
- Each variant must have a one-sentence parameter rationale (what market duration is being captured?)
- If all 3 variants fail the universe gate, the strategy is killed at 5-min — not retried with more variants
- DSR correction (nTrials) counts each variant as an independent trial — a strategy with 3 variants that all barely survive the universe gate on DSR correction may be marginal
