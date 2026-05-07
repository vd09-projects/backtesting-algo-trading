---
name: Engine gap handling — confirmed correct (TASK-0071)
description: The engine fills at candles[i].Open via pendingSignal with no gap clamping — CNC overnight gaps are correctly reflected in P&L without any engine change
type: project
---

The engine is gap-transparent by construction (confirmed 2026-05-07, TASK-0071). The `pendingSignal → candles[i].Open` fill path applies the pending fill to whatever Open the provider returns — no clamping, smoothing, or session-boundary check exists or is needed for CNC strategies.

Two golden tests confirm this:
- `TestGapDown_PositionEntered_PnLReflectsGap` (5-min IST, 3% gap-down, multi-session)
- `TestGapDown_ExitFillsAtNextBarOpen_NotSignalBarClose` (daily, gap-down from signal bar close)

Both passed without any engine change. Decision recorded: `decisions/convention/2026-05-07-overnight-gap-fill-confirmed-correct.md`.

**Why:** MIS session-boundary handling (TASK-0046) is a separate concern — forced close at 3:15 PM bar's Close. CNC holds overnight and is correctly exposed to gap risk via the existing fill model.

**How to apply:** Gap-verification tasks for CNC strategies do not need Marcus or engine changes. Golden tests are the right tool. Any future fill model refactor must preserve the invariant: fill price = `candles[i].Open` after slippage, no session-aware clamping for CNC.
