# Gap-and-Go sizing: vol-target 10% annualized with Rs 1.5 lakh hard per-position cap

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, vol-targeting, position-cap, TASK-0075 |

## Context

Gap-and-Go is a CNC strategy that holds for 1-2 sessions. Standard vol-targeting at 10% annualized applies, matching the rest of the portfolio. However, gap strategies have a specific capital deployment risk not present in trend-following strategies: index-driven events (budget day, RBI announcements, global macro) can trigger simultaneous gap signals across multiple instruments in the same sector. Without a hard position cap, 3 correlated gap-up signals on the same morning could deploy 100%+ of capital simultaneously. Marcus added a hard cap during the 2026-05-13 evaluation session.

## Decision

**Sizing rule:** vol-target 10% annualized. 20-bar rolling standard deviation of daily log returns (sample variance, consistent with Sharpe computation convention). Fraction = (0.10 / annual_vol). Capped at 1.0. Zero vol → fraction = 0 → trade skipped. Per-instrument notional = fraction × available capital.

**Hard per-position cap: Rs 1.5 lakh (50% of Rs 3 lakh capital base).** This cap applies regardless of the vol-targeting fraction output.

Rationale: the vol-targeting fraction is computed on each instrument's own realized volatility. In low-volatility conditions, a single instrument might receive a fraction > 0.50, pushing position size above Rs 1.5L. For trend-following strategies (MACD), this is fine — MACD holds for weeks and positions are rarely simultaneous across instruments. For Gap-and-Go, entry is triggered by a single-day event, and those events can be correlated: budget announcements gap up financials, pharma results gap up the sector, FII buying can gap up 4-5 midcap names on the same morning. A Rs 1.5L cap ensures that even if 2 gap signals trigger simultaneously, total exposure stays within the Rs 3L base.

This is a Gap-and-Go-specific rule. The MACD midcap portfolio does not need an equivalent cap because MACD entries are temporally dispersed. Gap-and-Go entries can cluster in time.

## Consequences

- Capital base: Rs 3 lakh total. The Rs 1.5L cap means a maximum of 2 simultaneous full-size positions before vol-targeting fraction would need to push individual positions below the cap.
- If vol-targeting assigns a fraction below 0.50, the position is sized by vol-targeting (not the cap). The cap only binds when vol-targeting would assign > 50% of capital to a single position.
- No pyramid within an instrument. Multiple instruments can hold simultaneous positions.

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document
- [MACD portfolio sizing — vol-targeting methodology](../algorithm/2026-05-06-macd-portfolio-sizing-sbin-titan-vol-targeting.md) — the vol-targeting approach this follows
