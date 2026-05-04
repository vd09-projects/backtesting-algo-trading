---
name: Evaluation pipeline status after TASK-0053
description: Current state of the 6-strategy evaluation pipeline after walk-forward kills both survivors
type: project
---

As of 2026-05-04, the evaluation pipeline (TASK-0049–0056) is in remediation:

- TASK-0052 (universe sweep): DONE. Survivors: MACD (DSRAvg=0.2715, 14 instruments), SMA (DSRAvg=0.0969, 12 instruments). Killed: donchian, RSI, bollinger, momentum.
- TASK-0053 (walk-forward): DONE. Both survivors killed: MACD 9/14, SMA 4/12 at instrument-count gate (100% retention required).
- TASK-0054 (bootstrap): BLOCKED pending survivors.

Active remediation:
- TASK-0067: DONE. Changed SMA --fast-period default from 10 → 20 in cmd/universe-sweep and cmd/walk-forward.
- TASK-0068: UP NEXT. Run SMA universe sweep at fast=20/slow=50; apply gate; walk-forward per-instrument. Requires Zerodha token. Data may be served from cache (.cache/zerodha/ last modified 2026-04-29).
- TASK-0069: BLOCKED (by TASK-0068). Marcus to reconsider MACD instrument-count gate threshold — 9/14 (64%) passing with solid OOS Sharpe may be sufficient under a relaxed gate.

**Why:** Walk-forward killed both strategies at 100% retention gate. User chose Option B (parameter retune) for SMA and Option A (gate-design review) for MACD.

**How to apply:** When starting next session, TASK-0068 is the top of Up Next. It's purely an evaluation run (no code changes needed) — load .env credentials, run the CLI commands, apply the gate, record decisions.
