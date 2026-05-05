---
name: Evaluation pipeline status after TASK-0069
description: Current state of the 6-strategy evaluation pipeline after MACD bootstrap gate passed with 4 survivors
type: project
---

As of 2026-05-06, the evaluation pipeline (TASK-0049–0056) has advanced to portfolio construction:

- TASK-0052 (universe sweep): DONE. Survivors: MACD (DSRAvg=0.2715), SMA (DSRAvg=0.0969 — later killed). 
- TASK-0053 (walk-forward): DONE. Both strategies killed under 100% retention gate; MACD relaxed gate (60%) applied.
- TASK-0054 (bootstrap): DONE. 4 MACD survivors: SBIN (P5=0.0719, Prob=98.0%), BAJFINANCE (P5=0.0467, Prob=97.3%), TITAN (P5=0.0854, Prob=98.7%), ICICIBANK (P5=0.0229, Prob=96.2%).
- TASK-0069: DONE. Bootstrap gate run; 5 killed (LT, INFY, AXISBANK, ITC, KOTAKBANK).
- SMA crossover: definitively killed at fast=20/slow=50 universe gate (0 sufficient instruments). No further SMA remediation.

Current Up Next:
- TASK-0055 (portfolio construction): TOP PRIORITY. Run cmd/correlate for 4 MACD survivors; apply correlation gate (r < 0.7 full-period, r < 0.6 stress); select 2-4 uncorrelated instruments; define vol-targeting sizing at ₹3 lakh. SBIN/ICICIBANK pair flagged as likely correlated (both banking). Owner: Marcus (marcus-design agent).
- TASK-0072 (midcap universe YAML): high priority alongside portfolio construction.

**Why:** Pipeline complete through bootstrap; correlation screening is the final gate before live deployment.

**How to apply:** Next evaluation session should use evaluation-run agent for TASK-0055 cmd/correlate run. No code needed for correlation — data already in runs/bootstrap-macd-2026-05-05/. Results in JSON output now include bootstrap.* fields (TASK-0082 done).
