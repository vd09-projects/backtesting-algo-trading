# Gap-and-Go exits: SL=0.8× gapPct, TP=2.0× gapPct, time-stop=150 bars (2 sessions)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, stop-loss, take-profit, time-stop, TASK-0075 |

## Context

Gap-and-Go enters on a gap-up event and holds for 1-2 sessions. Three exit conditions are needed: a stop-loss for gap reversals, a take-profit to capture clean continuation moves, and a time-stop to exit positions where the catalyst has faded without strong follow-through. The sizing of SL and TP must be tied to the gap magnitude rather than fixed percentages, since gap sizes vary across instruments and market conditions.

## Decision

Three exit conditions, evaluated each bar in priority order via `PriceExit` and `TimedExit` wrappers:

**Stop-loss:** `stopLossPct = stopLossMultiplier × gapPct` (default multiplier: 0.8). If gap was 1.0%, SL is 0.8% below entry price. Fill at next bar's Open via `PriceExit`.

**Take-profit:** `targetProfitPct = 2.0 × gapPct`. 2:1 reward/risk ratio on gap magnitude. If gap was 1.0%, TP is 2.0% above entry. Fill at next bar's Open via `PriceExit`.

**Time-stop:** `maxHoldBars = 150` (2 trading sessions × 75 bars/session at 5-min frequency). If neither SL nor TP is hit by bar 150, exit at next Open via `TimedExit`.

Wrapper composition: `NewTimedExit(NewPriceExit(inner, stopLossPct, targetProfitPct), 150)`.

**Critical implementation note:** `stopLossPct` and `targetProfitPct` are computed dynamically at entry from the actual `gapPct` detected at signal time, then passed to `NewPriceExit`. They are not fixed percentages — they are fractions proportional to the gap size at the time of the trade.

Rationale for SL at 0.8× gapPct: a reversal of 0.8× the gap magnitude signals that the institutional flow has flipped and a gap-fill scenario is underway. At that point, the thesis is broken and the position should be exited, not held waiting for recovery.

Rationale for TP at 2.0× gapPct: this sets a 2:1 R/R ratio on gap magnitude, consistent with the thesis that institutional follow-through typically runs 1.5-2× the initial gap before exhaustion. Captures clean continuation moves without overstaying.

Rationale for 150-bar time-stop (shorter than ORB's 225 bars): gap events are catalytic and time-decay faster than range breakouts. ORB has a structural range that constrains price, giving the position a defined reference point for 3 sessions. A gap event has no such structure — after 2 sessions, the catalyst is stale and any position that hasn't hit SL or TP is marking time in noise.

## Consequences

- The parameter sweep includes SL multiplier as an axis: [0.5×, 0.8×, 1.0× gapPct]. If 0.8× proves too tight (high churn on false signals), the sweep will surface 1.0× as the better plateau parameter.
- The time-stop sweep includes hold bars: [75, 150, 225]. 75 bars (1 session) is tight but tests whether same-day resolution is better. 225 bars (3 sessions) matches ORB and tests whether longer holds improve results at the cost of per-trade Sharpe.
- `PriceExit` and `TimedExit` wrapper statefulness: both are stateful wrappers. Per `decisions/convention/2026-04-27-timed-exit-statefulness-pkg-strategy.md`, they must not be shared across walk-forward folds — the factory API in `internal/walkforward` enforces this.

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document
- [Gap-and-Go default parameters and sweep axes](./2026-05-13-gap-and-go-default-params-1pct-1-3x-volume-sweep.md) — sweep axes including SL multiplier
- [TimedExit statefulness — not concurrent-safe](../convention/2026-04-27-timed-exit-statefulness-pkg-strategy.md) — wrapper statefulness constraint
