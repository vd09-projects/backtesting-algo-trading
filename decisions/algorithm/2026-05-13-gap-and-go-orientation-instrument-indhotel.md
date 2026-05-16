# Gap-and-Go parameter sweep uses NSE:INDHOTEL as orientation instrument, not NSE:RELIANCE

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | algorithm        |
| Tags     | gap-and-go, parameter-sweep, midcap, orientation, TASK-0075 |

## Context

All prior strategy evaluations (SMA, RSI, Donchian, MACD, Bollinger, Momentum, ORB) used NSE:RELIANCE as the orientation instrument for parameter sensitivity analysis. RELIANCE is the most liquid Nifty50 large-cap name and provides a stable, high-quality data anchor. Gap-and-Go's primary universe is Nifty Midcap 150 (48 instruments), and its edge mechanism — institutional flow lag on catalyst events — is expected to be weaker on large-caps like RELIANCE (where price discovery is near-instant) and stronger on midcap names with thinner order books.

## Decision

**NSE:INDHOTEL is the primary orientation instrument for Gap-and-Go parameter sensitivity sweeps.**

INDHOTEL qualifies for three reasons: (1) midcap volatility profile matching the target universe; (2) it is a bootstrap survivor in the MACD midcap portfolio (TASK-0097), confirming its 5-min bar history is representative of the midcap universe and its behavior is well-understood in the evaluation pipeline; (3) confirmed 5-min data coverage — 98,906 bars, 99.5% completeness per `decisions/algorithm/2026-05-10-5min-data-coverage-both-universes.md`.

NSE:RELIANCE remains acceptable as a **secondary large-cap cross-check** after the INDHOTEL primary sweep, to confirm whether the plateau-midpoint parameters also produce reasonable behavior on a large-cap instrument. This is optional — it does not change the parameters selected for the universe sweep.

The prior convention (RELIANCE as orientation instrument) applied to daily-bar strategies evaluated against the 15-instrument Nifty50 large-cap universe. Gap-and-Go is a 5-min-bar strategy targeting the 48-instrument midcap universe. Using a midcap orientation instrument is the methodologically correct choice: plateau parameters selected on RELIANCE behavior might be miscalibrated for the midcap universe where gap event frequency and volatility are different.

## Consequences

- TASK-0120 (parameter sensitivity) runs on NSE:INDHOTEL. The plateau-midpoint from this sweep becomes the fixed parameter for TASK-0121 (universe sweep across 48 midcap instruments).
- If INDHOTEL produces a sensitivity concern (no valid plateau, all-negative Sharpe in valid region), fallback-to-defaults applies per `decisions/algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md`.
- The decision to use INDHOTEL does not affect ORB's orientation instrument choice (RELIANCE — ORB targets the large-cap universe).

## Related decisions

- [Gap-and-Go — Marcus Rules](./2026-05-13-gap-and-go-marcus-rules.md) — full rules document
- [5-min bar coverage both universes](../algorithm/2026-05-10-5min-data-coverage-both-universes.md) — data quality confirmation for INDHOTEL
- [Plateau procedure](../algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md) — how plateau selection works on the orientation instrument
