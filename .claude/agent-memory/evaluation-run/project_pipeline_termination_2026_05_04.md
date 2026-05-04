---
name: Pipeline termination at walk-forward gate 2026-05-04
description: TASK-0053 completed; both strategies killed; evaluation pipeline from TASK-0052 has no survivors
type: project
---

As of 2026-05-04, the Phase 1 daily-bar evaluation pipeline (TASK-0052 → TASK-0053 → TASK-0054 → TASK-0055 → TASK-0056) is stalled at TASK-0053.

Walk-forward results (runs/walk-forward-2026-05-04.csv):
- macd-crossover: 9 of 14 eligible instruments passed per-instrument walk-forward gate (64%)
- sma-crossover: 4 of 12 eligible instruments passed per-instrument walk-forward gate (33%)

Both killed at the instrument-count gate (requires WF passes >= universe gate passes, i.e., 100% retention). 0 survivors.

**Why:** The 100% retention gate is very strict. MACD's 64% pass rate may still represent real edge on those 9 instruments. The failures are instrument-specific (RELIANCE, HINDUNILVR, WIPRO via OverfitFlag; TCS, HDFCBANK via NegFoldFlag) rather than a universal strategy failure.

**User decision pending (3 options):**
A. Relax instrument-count gate threshold (e.g., to 60-70% retention) — MACD would advance
B. Revisit strategy parameters (slower SMA periods, different MACD settings) — re-run from TASK-0052
C. Accept both kills and start fresh with new strategy candidates

**How to apply:** When the user next asks about the evaluation pipeline or "what's next," surface this decision point immediately. Do not run TASK-0054 until the user resolves this.
