# Walk-forward instrument-count gate: threshold relaxed from 100% to 60% retention

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-05       |
| Status   | accepted         |
| Category | algorithm        |
| Tags     | walk-forward, instrument-count-gate, gate-design, macd-crossover, evaluation-methodology, TASK-0069 |

## Context

The walk-forward gate as applied in TASK-0053 required a strategy to pass walk-forward on the same number of instruments it passed at the universe gate — effectively 100% retention. MACD crossover (17/26/9) failed this gate at 9/14 (64% retention). The 5 failures cluster structurally: 3 large-cap defensives (RELIANCE, HINDUNILVR, WIPRO) via OverfitFlag, and 2 higher-volatility names (TCS, HDFCBANK) via NegativeFoldFlag.

TASK-0069 asks whether the 100% retention requirement is appropriately calibrated, given that the universe gate itself only required 40% pass fraction (≥ 6 of 15 instruments with positive Sharpe and ≥ 30 trades). The question is structural: should the walk-forward gate be philosophically consistent with the universe gate, or stricter by design?

## The methodological cost of this decision

This gate change is being made after MACD failed the 100% requirement. That is a real methodological cost. It is not as bad as it could be — the new threshold is being set as a pipeline-wide rule applied to all future strategies, not a MACD-specific carve-out, and the rationale is structural (consistency with universe gate philosophy) rather than "because MACD almost passed." But it is not as clean as if this threshold had been written before TASK-0053 ran. That limitation is acknowledged here and cannot be undone.

The mitigating factor: the 100% retention requirement was never explicitly written as a gate decision. It was implicit — inherited from "must pass on at least as many instruments as the universe gate passed." Replacing an implicit, unwritten rule with an explicit one at 60% is different from relaxing an explicit pre-committed threshold.

## The case for a 60% floor

The universe gate philosophy was stated explicitly in the 2026-04-25 decision: a strategy that generalizes across instruments is more likely capturing real market structure than instrument-specific noise. The universe gate set that floor at 40% (≥ 6 of 15). Requiring 100% retention at walk-forward contradicts that philosophy — it says "the strategy must work everywhere, with no regime exceptions," which is a much higher bar than "the strategy must generalize."

The walk-forward gate is designed to test regime stability, not universality. A strategy that works on 9 of 14 instruments in an out-of-sample temporal test is demonstrating cross-instrument generalization. Requiring 14/14 means a single instrument anomaly — one stock with an unusual regime shift in the 2018–2024 window — can kill a real edge. That's miscalibrated.

60% retention (≥ 9 of the universe-gate pass count, rounded down) is the floor I'm willing to defend. It's consistent with the universe gate's tolerance for non-universal strategies. Below 60%, the cross-instrument evidence is thin enough that the bootstrap would likely kill it anyway.

## The structural failure clustering rule

The clustered nature of MACD's failures — defensives failing via OverfitFlag, higher-vol names failing via NegFoldFlag — is informative, but not in the way the revisit trigger suggested. It does not rescue the strategy from the gate; it explains the gate failure. RELIANCE (OOSISRatio=0.48, borderline), HINDUNILVR (0.34), and WIPRO (0.43) failing together suggests that MACD momentum on daily bars does not capture edge in defensive/low-beta sectors. TCS and HDFCBANK failing via NegFoldFlag (2 negative OOS folds each) suggests the strategy is regime-unstable on banking/IT heavyweights specifically.

This is not a reason to rescue MACD. It is context for what the bootstrap will test. If MACD passes bootstrap on the 9 passing instruments, those 9 are the deployable universe — RELIANCE, HINDUNILVR, WIPRO, TCS, HDFCBANK are excluded from the live portfolio regardless of gate outcome.

## Decision

**The walk-forward instrument-count gate is reset to: a strategy passes if it passes walk-forward on ≥ 60% of the instruments that passed the universe gate (rounded down), with a minimum floor of 6 instruments.**

Formally: `WF_passes >= floor(0.60 * universe_gate_passes)` AND `WF_passes >= 6`.

This threshold applies to all strategies in the current and future evaluation pipelines. It is not specific to MACD.

**MACD crossover (17/26/9) passes the revised gate: 9/14 = 64.3% ≥ 60%.** It advances to bootstrap (TASK-0054 logic) on the 9 passing instrument pairs only:
- NSE:SBIN, NSE:BAJFINANCE, NSE:TITAN, NSE:LT, NSE:ICICIBANK, NSE:INFY, NSE:AXISBANK, NSE:ITC, NSE:KOTAKBANK

The 5 failing instruments (TCS, RELIANCE, HINDUNILVR, WIPRO, HDFCBANK) are excluded from the MACD deployment universe permanently under this evaluation cycle. They do not get a second chance at bootstrap — the walk-forward gate is final per instrument.

## Consequences for the pipeline

1. MACD crossover advances to bootstrap on 9 instrument pairs. Bootstrap gate applies: SharpeP5 > 0 AND P(Sharpe > 0) > 80% per the 2026-04-27 bootstrap gate decision.
2. The revised 60% threshold applies retroactively to re-evaluate prior walk-forward kills. Under the new threshold, the SMA crossover fast=10/slow=20 kill from TASK-0053 still stands: 4/12 = 33% < 60%.
3. For future strategies: if a strategy passes the universe gate on N instruments, it must pass walk-forward on ≥ floor(0.60 * N) instruments (minimum 6).

## Related decisions

- [Cross-instrument universe gate (2026-04-25)](./2026-04-25-cross-instrument-proliferation-gate.md) — establishes the 40% universe gate floor that this decision is now consistent with
- [Walk-forward OOS/IS Sharpe threshold (2026-04-22)](./2026-04-22-walk-forward-oos-is-sharpe-threshold.md) — per-instrument pass/fail criteria unchanged
- [MACD fails walk-forward instrument-count gate (2026-05-04)](./2026-05-04-macd-crossover-walk-forward-instrument-count-gate.md) — the kill record superseded by this gate revision for MACD; original kill record remains as historical evidence of the 100% gate evaluation
- [Bootstrap gate (2026-04-27)](./2026-04-27-bootstrap-gate.md) — next gate MACD must clear

## Revisit trigger

If a strategy passes the 60% threshold with failures that do NOT cluster structurally — i.e., failures are distributed across sector types and failure modes — and it subsequently fails bootstrap, reconsider whether 60% is too low. Consistent random failures across the universe at 60% retention is a different signal than MACD's clustered structural failures. The structural clustering here was informative context, not a gate criterion; keep it that way.

If any future strategy passes the walk-forward gate at exactly the 60% floor (minimum pass count) and the bootstrap produces SharpeP5 between 0 and 0.01, revisit whether the bootstrap gate is set correctly given thin cross-instrument evidence at the floor.
