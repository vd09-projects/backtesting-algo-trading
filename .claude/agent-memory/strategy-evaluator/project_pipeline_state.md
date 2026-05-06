---
name: Project pipeline state and gate history
description: Full evaluation pipeline status as of 2026-05-06 — 7 strategies killed, MACD crossover surviving on 4 instruments, portfolio construction in progress
type: project
---

As of 2026-05-06, the MACD crossover strategy (fast=17, slow=26, signal=9) is the sole survivor of the evaluation pipeline on NSE Nifty50 large-cap daily bars (2018-2024).

**Why:** 7 other strategies were killed at various gates. MACD was nearly killed at walk-forward until the instrument-count gate was relaxed from 100% to 60% retention.

**How to apply:** When evaluating new strategies, the most common kill point has been the universe gate (5 of 7 kills). Signal frequency on daily bars is the first thing to stress-test. The walk-forward relaxation to 60% was a one-time methodology revision — do not assume future strategies get the same treatment.

**Pipeline stage:** Portfolio construction (TASK-0055). Currently in the correlation gate and regime gate computation phase. TASK-0085 (pairwise Pearson r) and TASK-0086 (regime gate) are the next steps before TASK-0087 (final portfolio composition decision).

**Surviving instruments:** NSE:SBIN, NSE:BAJFINANCE, NSE:TITAN, NSE:ICICIBANK at bootstrap gate. Expected to narrow to SBIN + TITAN after correlation gate (banking cluster correlation).

**True holdout:** 2025 onward — never touched during evaluation.
