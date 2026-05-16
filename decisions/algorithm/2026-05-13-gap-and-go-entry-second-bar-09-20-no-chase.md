# Gap-and-Go entry on close of second bar (09:20 IST); no-entry if gap chased by bar 1

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, entry-bar, 5min, CNC, NSE-midcap, TASK-0075 |

## Context

Gap-and-Go detects a gap condition at the 09:15 IST bar (first bar of the session) via `PreviousSessionClose(bars, i)`. The question is which bar to use for the actual entry fill: the open of the 09:15 gap bar itself, or a subsequent bar. This decision was made by Marcus during the 2026-05-13 evaluation session.

## Decision

Entry is on the **close of the second bar (09:20 IST, bar index 1 from session start)**, not the open of the 09:15 gap bar.

Two conditions govern entry at bar index 1:
1. The gap condition still holds: `abs(bars[1].Close - PreviousSessionClose) / PreviousSessionClose >= gapPct` is not required, but the no-chase guard below must not fire.
2. **No-entry (gap-already-chased) guard**: if `abs(bars[1].Close - PreviousSessionClose) / PreviousSessionClose >= 1.5 × gapPct`, skip the trade entirely — the gap has been chased and you are buying exhaustion, not continuation.

Fill execution: signal emitted at bar index 1 close, filled at bar index 2's Open via the standard `pendingSignal` mechanism. One position per instrument per day, no pyramid.

Rationale: the 09:15 Open is where spread is widest and prints are most adversarial post-gap — market makers widen quotes around opening uncertainty, and the first print often reflects retail panic and institutional sweep orders, not a stable price level. The second bar's close confirms the gap is holding under follow-through buying. Waiting one bar costs a few ticks of slippage relative to the open but gets a meaningfully cleaner fill. The no-chase guard is the complementary protection: if price has already moved 1.5× the gap magnitude by bar 1, the institutional flow has largely resolved and entering now is buying a move that is substantially complete.

## Consequences

- Implementation: gap detection logic executes at bar index 0 (09:15), but the entry signal is generated at bar index 1 (09:20 close). The strategy must track both the prior session close and the gap condition across two bars per session.
- The 20-day average volume check (volumeMultiplier threshold) is applied at bar index 0 (09:15 bar volume), not bar index 1, since bar index 0 is the gap bar that contains the opening volume surge.
- The no-chase guard (1.5× gapPct) means that large-gap events (e.g., earnings beats driving 3-4% gaps) are more likely to be filtered out if the first session bar sees rapid follow-through. This is by design — the strategy targets institutional order flow lag, not the first wave of the move.

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document; this is the granular entry-rules decision
- [Overnight gap fill: engine is gap-transparent by construction](../convention/2026-05-07-overnight-gap-fill-confirmed-correct.md) — fill model confirmation; fills at next bar's Open
