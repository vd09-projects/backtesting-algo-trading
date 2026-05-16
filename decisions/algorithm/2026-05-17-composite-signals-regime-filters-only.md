---
id: 2026-05-17-composite-signals-regime-filters-only
title: "Composite signals on 5-min are regime filters only — not oscillator conjunction"
date: 2026-05-17
status: accepted
category: algorithm
tags: [composite-signals, regime-filter, VWAP, overfitting, 5min]
---

## Decision

Composite signal strategies on 5-min bars are acceptable only as regime filters on a base strategy signal, not as oscillator conjunctions. A regime filter reduces a base signal's scope to favorable market conditions with an independent economic rationale. Oscillator conjunction (MACD AND RSI AND WMA all agree) is not acceptable — it reduces trade count without adding independent information.

## What passes

A regime filter has:
1. An economic mechanism independent of the base signal's mechanism
2. A clear directional hypothesis (when filter condition is true, edge improves for a specific reason)
3. A reasonable expectation that it will not cut trade count by more than 50%

Accepted combinations under this rule:
- **MACD crossover + price above VWAP**: VWAP is an institutional benchmark; price above VWAP means institutional buyers are net positive — directionally aligned with a long MACD signal
- **RSI oversold + volume >= 1.5× session average**: volume confirms exhaustion selling (capitulation) vs. slow drift — reduces false entries where RSI is low simply due to thin afternoon selling
- **SMA crossover + first 120 min of session**: NSE institutional participation is highest at open; afternoon crossovers on 5-min have lower follow-through due to reduced liquidity

## What fails

- MACD AND RSI AND WMA must all agree: three momentum-based indicators are highly correlated; requiring all three adds no independent information while cutting trade count dramatically
- Any filter added post-hoc after observing it improves IS Sharpe: must have economic rationale articulated before looking at results

## Implementation constraint

VWAP combinations require TASK-0131 (session-aware VWAP utility) as a prerequisite. Composite strategy tickets are not created until TASK-0127 identifies which base strategies survive 5-min adaptation — don't build filters for dead base strategies.
