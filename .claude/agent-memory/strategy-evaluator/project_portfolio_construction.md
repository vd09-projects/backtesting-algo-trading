---
name: Portfolio construction state — MACD complete, ORB and Gap-and-Go evaluation queued
description: Current portfolio state: MACD complete on large-cap and midcap; ORB and Gap-and-Go have GO verdicts, awaiting implementation and evaluation pipeline
type: project
---

As of 2026-05-13:

**MACD crossover — Large-cap (COMPLETE)**
- Survivors: NSE:SBIN + NSE:TITAN
- Sizing: vol-target 10% annualized, Rs 1.5L notional each
- Kill-switch: SBIN SharpeP5=0.0719, MaxDD=4.10%, MaxDDDuration=448 days; TITAN SharpeP5=0.0854, MaxDD=4.72%, MaxDDDuration=1,388 days
- Decisions: `decisions/algorithm/2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md`, `decisions/algorithm/2026-05-07-kill-switch-thresholds-live-brief.md`

**MACD crossover — Midcap (COMPLETE)**
- Survivors: PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN
- Sizing: vol-target 10%, Rs 50k base notional; IT-sector cap (PERSISTENT+COFORGE <= Rs 40k each when co-deployed)
- Kill-switches: per `decisions/algorithm/2026-05-09-macd-crossover-midcap-correlation-killswitch-portfolio.md`
- All 6 pass pairwise correlation (max r=0.36 full-period, 0.47 stress — below all thresholds)

**ORB (Opening Range Breakout) — 5-min, CNC, large-cap (EVALUATION QUEUED)**
- Marcus GO verdict: 2026-05-13
- Rules: `decisions/algorithm/2026-05-13-orb-marcus-rules.md`
- Entry: 1st close > RangeHigh × 1.001 after 6-bar range window; no-entry cutoff 11:30 IST
- Exit: SL = 1.5× range width; TP = 2.0× range width; time-stop = 225 bars (3 sessions)
- Sizing: vol-target 10%, no hard cap (trend-following, not event-clustered)
- Evaluation tasks: TASK-0113 through TASK-0118, all blocked on TASK-0074 implementation
- Primary risk: walk-forward OverfitFlag in 2022 choppy regime
- Secondary risk: ORB/Gap-and-Go mutual correlation if both pass bootstrap (stress-period r may > 0.6)

**Gap-and-Go — 5-min, CNC, midcap primary (EVALUATION QUEUED)**
- Marcus GO verdict: 2026-05-13
- Rules: `decisions/algorithm/2026-05-13-gap-and-go-marcus-rules.md`
- Entry: 2nd bar close (09:20 IST); gapThreshold=1.0%; volumeMultiplier=1.3× 20-day avg; no-entry if gap chased 1.5× by bar 1
- Exit: SL = 0.8× gapPct; TP = 2.0× gapPct; time-stop = 150 bars (2 sessions)
- Sizing: vol-target 10%, hard per-position cap Rs 1.5L (50% of Rs 3L) — prevents over-concentration on event days
- Orientation instrument: NSE:INDHOTEL (not RELIANCE — midcap universe strategy)
- Evaluation tasks: TASK-0119 through TASK-0125, all blocked on TASK-0075 implementation
- Primary risk: walk-forward OverfitFlag in 2022; secondary: ORB/Gap-and-Go mutual correlation

**How to apply:** Any new strategy evaluation must account for: (1) MACD long-only large-cap trend-following and MACD long-only midcap trend-following as existing correlations to check against; (2) ORB and Gap-and-Go as queued long-only intraday strategies that will need correlation checks once they reach bootstrap. The portfolio is building toward MACD (daily) + ORB (intraday) + Gap-and-Go (intraday) — all long-only, which concentrates long exposure in crash regimes.
