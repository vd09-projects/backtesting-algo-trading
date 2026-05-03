# MACD parameters unchanged at 17/26/9 for this evaluation cycle

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-04       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | macd-crossover, parameter-selection, macd-fast-period, post-hoc-risk, instrument-count-gate, TASK-0067 |

## Context

MACD crossover ran at fast=17, slow=26, signal=9 (plateau-midpoint from TASK-0051 parameter sensitivity sweep on NSE:RELIANCE). It passed the universe gate strongly (DSRAvg=0.2715, 14/15 instruments). It then failed the walk-forward instrument-count gate at 9/14 instruments — 64% retention, needs 100%.

The question arose whether to switch from fast=17 to standard fast=12 for this re-evaluation cycle. The rationale for changing: fast=17 came from a single-instrument sweep on RELIANCE, and RELIANCE subsequently failed the walk-forward gate (OverfitFlag, OOSISRatio=0.48).

## Decision

**MACD parameters remain at fast=17, slow=26, signal=9 for this evaluation cycle.** Do not change to 12/26/9.

**Rationale:**

1. **fast=17 was not RELIANCE-specific.** A RELIANCE-fitted parameter would have predicted RELIANCE outperformance. Instead, fast=17 produced DSRAvg=0.2715 across 14 of 15 instruments, with 9 of 14 passing walk-forward — strong cross-instrument evidence that the parameter is capturing real market structure, not RELIANCE-specific noise.

2. **RELIANCE's WF failure is structural, not parameter-related.** RELIANCE (large-cap conglomerate, regime-insensitive trend behavior in 2018-2024) doesn't reward MACD regardless of the fast period. The OverfitFlag at OOSISRatio=0.48 reflects regime concentration in RELIANCE's specific price structure.

3. **Changing to 12/26/9 post-WF failure is the real cherry-picking risk.** After seeing that RELIANCE failed at fast=17, choosing fast=12 on the basis that "standard MACD doesn't use RELIANCE's plateau" constitutes post-hoc parameter selection against the WF fold structure of the 5 failing instruments. This is methodologically unsound even if framed as "going to defaults."

4. **The failure is a gate-design question, not a parameter question.** MACD's 64% retention rate (9/14) and the quality of its passing instruments (OOS Sharpe 0.062–0.472) suggests real edge on a subset of the universe. The 100% retention requirement is the constraint, not the parameter choice. See TASK-0069 for gate-design escalation.

## Consequences

- MACD evaluation does not re-run with different parameters in this cycle
- TASK-0069 escalates the instrument-count gate threshold question to Marcus
- If the gate threshold is relaxed (e.g., to 60-70% retention), MACD at 17/26/9 should be re-evaluated against the same WF results from TASK-0053 — no re-run needed
- If the gate is not relaxed, MACD at 17/26/9 remains killed under the current methodology

## Related decisions

- [MACD crossover fails walk-forward instrument-count gate](./2026-05-04-macd-crossover-walk-forward-instrument-count-gate.md) — the kill this decision responds to
- [MACD crossover passes universe gate (fast=17, slow=26, signal=9)](./2026-05-03-macd-crossover-universe-gate-passed.md) — universe gate evidence

## Revisit trigger

If TASK-0069 results in a relaxed instrument-count gate, re-evaluate MACD at 17/26/9 using existing TASK-0053 results. If the gate threshold is set at ≤9 required passes, MACD advances with 9/14 passing instruments without any further runs.
