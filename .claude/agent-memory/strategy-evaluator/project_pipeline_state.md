---
name: Project pipeline state and gate history
description: Full evaluation pipeline status as of 2026-05-13 — MACD pipeline complete, ORB and Gap-and-Go both received GO verdicts, intraday 5-min strategy phase in progress
type: project
---

As of 2026-05-13, the project has two distinct pipeline tracks:

**Track 1: Daily-bar strategies (complete)**

MACD crossover (fast=17, slow=26, signal=9) is the sole survivor across two universes:
- Large-cap (Nifty50, 15 instruments): SBIN + TITAN survive all gates including correlation, regime, kill-switch brief. Portfolio complete. Kill-switch thresholds committed.
- Midcap (Nifty Midcap 150, 48 instruments): 6 instruments survive all gates (PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN). Correlation screen passed (zero pairs breach 0.7 full-period or 0.6 stress). Vol-target 10%, Rs 50k base notional, IT-sector cap (PERSISTENT+COFORGE capped at Rs 40k each when co-deployed). Pipeline complete as of 2026-05-09.

**Track 2: Intraday 5-min strategies (in progress)**

Data infrastructure complete: 5-min bar cache 15/15 large-cap and 48/48 midcap instruments (2021-2023 window, 99.5% bar completeness). Session helpers, PriceExit, TimedExit wrappers all done.

- ORB (TASK-0074): Marcus GO verdict 2026-05-13. Rules in `decisions/algorithm/2026-05-13-orb-marcus-rules.md`. Awaiting Priya implementation. Evaluation pipeline tasks created (TASK-0113 through TASK-0118, blocked on implementation).

- Gap-and-Go (TASK-0075): Marcus GO verdict 2026-05-13. Rules in `decisions/algorithm/2026-05-13-gap-and-go-marcus-rules.md`. TASK-0075 unblocked, ready for Priya to implement. Evaluation pipeline tasks created (TASK-0119 through TASK-0125, blocked on implementation).

**Why:** The intraday strategies target event-driven and session-structure edge on NSE midcap names — different competition set from daily-bar trend-following.

**How to apply:** When evaluating new strategies, the pipeline stage is "intraday evaluation". Both ORB and Gap-and-Go are queued for implementation before evaluation pipeline runs can begin. Any new strategy proposal should queue behind these two.

**True holdout:** 2025 onward — never touched during evaluation.

**Walk-forward structure for 5-min strategies:** 2yr IS / 6mo OOS / anchored, 6 OOS periods covering 2023-2026. Different from daily-bar WF (2yr IS / 1yr OOS / 1yr step) because 5-min data window is shorter (2021-2023 only).
