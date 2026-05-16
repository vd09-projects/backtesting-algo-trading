# ORB (Opening Range Breakout) — Marcus Rules

| Field    | Value                                                   |
|----------|---------------------------------------------------------|
| Date     | 2026-05-13                                              |
| Status   | experimental                                            |
| Category | algorithm                                               |
| Tags     | orb, opening-range-breakout, 5min, CNC, intraday, breakout, TASK-0074 |

## Evaluation verdict

GO — edge mechanistically sound, infrastructure complete, trade frequency adequate, structurally distinct from killed daily-bar breakout strategies. Walk-forward gate identified as most likely structural barrier.

## Edge thesis

The first 30 minutes of the NSE session represent price discovery: overnight news is absorbed in the opening auction, large institutional orders work through the open, and a clean break from the established range signals that the market has resolved its directional uncertainty. On NSE large-caps, dominated by institutional and FII flow, this mechanism is legitimate.

The 2-3 day CNC hold captures institutional program follow-through: catalysts (earnings, macro) on NSE large-caps often take more than one session to fully work through. This extends ORB beyond pure intraday into a short-term momentum regime.

Structural distinction from Donchian breakout (which failed DSR-corrected Sharpe at universe gate on daily bars): Donchian operates on multi-week daily-bar trends competing with every published trend system. ORB operates on intraday session structure re-established fresh every day; the competition set is different and the arb pressure is lower.

## Signal frequency estimate

- Trigger rate: 15-25% of trading days (clean breakout condition met)
- Estimated trades/year/instrument: 40-80
- Estimated trades over 2018-2023 (6 years, ~1,450 trading days): 240-480 per instrument
- Signal audit gate (>= 30 trades/instrument): expected to clear comfortably
- Risk: tightening entry conditions too aggressively kills frequency; the sweep will surface this

## Entry rules

**Range window:** first 6 bars of the session (09:15-09:44 IST, inclusive). `IsSessionOpen()` identifies the 09:15 bar. This is the sweep starting point — see parameter sweep axes below.

**Range computation:**
- `RangeHigh = max(High) over bars[sessionStart : sessionStart+6]`
- `RangeLow = min(Low) over bars[sessionStart : sessionStart+6]`
- `RangeWidth = RangeHigh - RangeLow`

**Entry signal:** generated on bar 6 (09:45 bar) or later. Long entry triggered on the *first bar* whose `Close > RangeHigh * (1 + buffer)` where buffer = 0.10%. Long only, no pyramid. One position per session at most.

**No-entry cutoff:** 11:30 IST (bar 27 of the 09:15-origin session). Late-session breakouts have materially worse follow-through. No new entries after this time regardless of signal.

**Commission model:** `CommissionZerodhaFull` (CNC rates). Not MIS — no forced close applies. Overnight gap risk is present and correctly modelled by the engine (gap-transparent fill confirmed per accepted convention).

## Exit rules

Three exit conditions are evaluated each bar in priority order:

**1. Stop-loss (PriceExit wrapper):**
`StopLossPct = 1.5 × (RangeWidth / EntryPrice)`
If price falls this percentage below entry, exit at next bar's Open. This is ATR-proportional by construction — range width acts as a volatility proxy.

**2. Take-profit (PriceExit wrapper):**
`TargetProfitPct = 2.0 × (RangeWidth / EntryPrice)`
2:1 reward/risk ratio on range width. Exit at next bar's Open when triggered.

**3. Time-stop (TimedExit wrapper):**
`MaxHoldBars = 225` (3 trading sessions × 75 bars/session at 5-min frequency).
If neither SL nor TP is hit by bar 225, exit at next Open. This handles flat/sideways positions.

Note: PriceExit SL/TP values are proportional to range width at entry, not fixed percentages. The implementation must compute `stopLossPct` and `targetProfitPct` as `(1.5 × RangeWidth / EntryPrice)` and `(2.0 × RangeWidth / EntryPrice)` at signal time and pass them to `NewPriceExit`.

## Parameter sweep axes (sensitivity analysis before universe sweep)

Run on RELIANCE 2018-2023 as the orientation instrument. Use the plateau-midpoint selection procedure (80% Sharpe floor within valid trade-count region).

| Axis            | Values to test          | Default (sweep start) |
|-----------------|-------------------------|-----------------------|
| Range window    | 6, 9, 12 bars           | 6 bars (30 min)       |
| Buffer          | 0.05%, 0.10%, 0.20%     | 0.10%                 |
| SL multiplier   | 1.0x, 1.5x, 2.0x range  | 1.5x                  |
| Hold bars       | 150, 225, 300 bars       | 225 bars (3 sessions) |

Plateau midpoint from this sweep becomes the fixed parameter for universe sweep across all 15 large-cap instruments.

## Sizing

Vol-target 10% annualized. 20-bar rolling standard deviation of daily log returns (sample variance, consistent with Sharpe computation). Fraction = (0.10 / annual_vol). Capped at 1.0. Zero vol → fraction = 0 → trade skipped. Per-instrument notional = fraction × available capital, refined after range-width-derived SL is known from sweep results.

Capital base: Rs 3 lakh total. Single instrument position at a time (no pyramid).

## Kill-switch thresholds

Kill-switch thresholds are derived from bootstrap results per the accepted kill-switch derivation methodology and must be committed in a companion decision file before any live deployment. Specific values are not available until bootstrap completes.

Framework (from accepted methodology):
1. Rolling per-trade Sharpe threshold: SharpeP5 from `montecarlo.Bootstrap` (same formula: `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance, no annualization)
2. Maximum drawdown threshold: 1.5 × in-sample MaxDrawdown
3. Drawdown duration threshold: 2 × in-sample MaxDrawdownDuration

## Pipeline risk flags

1. **Walk-forward gate (highest risk):** IS-to-OOS Sharpe degradation is expected, particularly in choppy regimes. The 2018-2024 window includes COVID (2020) and rate shock (2022) — both are wide-range, false-breakout environments. OverfitFlag (OOS < 50% IS) could trigger if the strategy has strong IS performance in trending periods that doesn't survive OOS.

2. **Correlation gate (moderate risk):** MACD survivors on large-caps are SBIN and TITAN. If ORB also concentrates on SBIN and TITAN in the universe sweep, stress-period correlation could breach r < 0.6. Monitor instrument overlap at universe gate stage and prefer instruments outside the MACD survivor set.

3. **Universe gate (low-moderate risk):** DSR correction for 15 instruments is less punishing than for 48 midcap instruments. Given the frequency estimate, this gate should clear — but if SL triggers dominate and per-trade Sharpe is negative, DSR won't save it.

## Implementation notes for Priya

- Timeframe: `model.Timeframe5Min`
- Use `IsSessionOpen(bar)` to detect the 09:15 bar and begin range accumulation
- Range is a stateful struct computed over the first `rangeWindowBars` bars of each session; reset on each new `IsSessionOpen()` trigger
- Entry signal emitted on the bar after the range window closes, if `Close > RangeHigh × (1 + buffer)` and IST time < 11:30
- `NewPriceExit(inner, stopLossPct, targetProfitPct)` wraps the inner ORB strategy; SL and TP percentages are computed dynamically at entry from range width
- `NewTimedExit(pricExitWrapped, maxHoldBars)` wraps the PriceExit layer for the time-stop
- Strategy must be stateful (range accumulation across bars within a session); `Next()` receives the full candle slice — session tracking via index i
- Per the no-allocation-in-hot-loop constraint: pre-allocate the range buffer at construction, not per-call
- Register in all strategy registries: `cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`
- Name: `"orb"` for registry consistency

## Related decisions

- [Overnight gap fill: engine is gap-transparent by construction](../convention/2026-05-07-overnight-gap-fill-confirmed-correct.md) — CNC fill model confirmed correct; no engine change needed
- [Intraday forced-close fill price](./2026-04-25-intraday-forced-close-fill-price.md) — MIS only, does not apply to CNC ORB
- [Kill-switch derivation methodology](./2026-04-21-kill-switch-derivation-methodology.md) — framework for post-bootstrap thresholds
- [Plateau procedure: trade-count constrained](./2026-04-29-plateau-procedure-trade-count-constrained.md) — governs parameter sweep plateau selection
