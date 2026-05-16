# Gap-and-Go — Marcus Rules

| Field    | Value                                                                 |
|----------|-----------------------------------------------------------------------|
| Date     | 2026-05-13                                                            |
| Status   | experimental                                                          |
| Category | algorithm                                                             |
| Tags     | gap-and-go, 5min, CNC, intraday, event-driven, behavioral, TASK-0075 |

## Evaluation verdict

GO — edge mechanistically sound (institutional flow lag on NSE midcap catalyst events), structurally distinct from killed Donchian breakout, signal frequency expected to clear audit floor, infrastructure complete. Walk-forward gate identified as primary structural risk.

## Edge thesis

NSE midcap stocks gapping 1%+ above prior close on above-average volume represent institutional order flow that hasn't fully resolved at the open. On large-cap names (RELIANCE, INFY), price discovery is near-instant — every HFT and institutional algorithm watches these continuously. On midcap names, the flow lag can span hours: retail is slower to react, analyst coverage is thinner, and the opening auction doesn't fully clear the overnight position imbalance. A stock that gaps up 1% on an earnings beat or sector upgrade at 09:15 IST is often still seeing net institutional buying through the first session. That persistence is the edge.

Structural distinction from Donchian breakout (killed at universe gate, DSR-corrected Sharpe negative, only 7/15 sufficient instruments on daily bars): Donchian detects multi-week trend continuation on daily bars, competing with every published trend-following system. It fires on the 10th consecutive day of a move that every algorithmic system has already detected and priced. Gap-and-Go detects a single-day catalyst event and trades the first-session propagation — a different competition set, a shorter hold horizon, and a mechanism tied to event-driven order flow, not trend persistence.

**Decision (2026-05.1.0) — algorithm: experimental**
scope: gap-and-go, entry-rules
tags: gap-and-go, entry-bar, 5min, CNC, NSE-midcap, TASK-0075
owner: marcus

Gap-and-Go entry on close of second bar (09:20 IST, bar index 1 from session start), not the open of the 09:15 gap bar. The 09:15 Open is where spread is widest and prints are most adversarial post-gap; the second bar's close confirms the gap is holding and gets a more stable fill. No-entry condition: if price has already moved 1.5x gapPct from PreviousSessionClose by bar index 1, skip the trade — the gap has been chased and you are buying exhaustion, not continuation. One position per instrument per day, no pyramid.

## Signal frequency estimate

- Gap trigger rate (1.0% threshold, 1.3x volume): 8-12% of trading days on NSE midcap names
- Estimated trades per year per instrument: 20-30
- Estimated trades over 2021-2023 window (2.5 years, ~625 trading days): 50-75 per instrument
- Signal audit gate (>= 30 trades): expected to clear at 1.0% threshold
- Risk flag: if more than 20% of instruments fail the signal audit at default parameters (1.0% gap, 1.3x volume), test 0.75% gap threshold before declaring a kill

**Decision (2026-05.1.2) — algorithm: experimental**
scope: gap-and-go, signal-frequency
tags: gap-and-go, volume-threshold, gap-threshold, signal-audit, TASK-0075
owner: marcus

Default parameters: gapThreshold=1.0%, volumeMultiplier=1.3x (20-day average). These are the starting points for the parameter sweep. Expected trigger rate: 8-12% of trading days at 1.0% threshold on NSE midcap names. If more than 20% of instruments fail signal audit at defaults, test 0.75% threshold before declaring a kill — the edge may require a lower bar to generate sufficient trades on lower-vol names.

## Entry rules

**Gap detection:**
`gapPct = (Open09:15 - PreviousSessionClose) / PreviousSessionClose`

where `PreviousSessionClose` is computed via `pkg/strategy/session.go:PreviousSessionClose(bars, i)` at bar index i corresponding to the 09:15 bar.

**Entry conditions (all must hold):**
1. `gapPct >= gapThreshold` (default 1.0%)
2. Volume at the 09:15 bar (bar index 0 from session start) exceeds `volumeMultiplier x 20-day average daily volume` (default 1.3x)
3. No existing position in this instrument
4. `abs(bars[1].Close - PreviousSessionClose) / PreviousSessionClose < 1.5 x gapPct` — the gap has not been fully chased by bar 1's close

**Entry execution:** signal generated at bar index 1 (09:20 IST close), filled at bar index 2's Open (09:25 IST Open), via the standard pendingSignal mechanism. This is consistent with the engine's gap-transparent fill model (convention/2026-05-07-overnight-gap-fill-confirmed-correct.md).

**Commission model:** `CommissionZerodhaFull` (CNC rates). Not MIS — no forced close applies. Overnight gap risk is present and correctly modelled by the engine.

**Long-only:** bidirectional gap-and-go (short-side on gap-down) is deferred. Short-side requires securities lending, NSE short-delivery risk management, and different stop-loss mechanics. The long-only edge is tested first; if it passes the pipeline, the short-side thesis is evaluated separately.

## Exit rules

Three exit conditions evaluated each bar in priority order:

**1. Stop-loss (PriceExit wrapper):**
`stopLossPct = stopLossMultiplier x gapPct`
Default: `stopLossMultiplier = 0.8`. If the gap was 1.0%, SL is 0.8% below entry price. Rationale: a reversal of 0.8x the gap magnitude signals that the institutional flow has flipped and a gap-fill scenario is underway. PriceExit handles this at next bar's Open.

**2. Take-profit (PriceExit wrapper):**
`targetProfitPct = 2.0 x gapPct`
2:1 reward/risk on gap magnitude. If the gap was 1.0%, TP is 2.0% above entry. PriceExit handles this at next bar's Open.

**3. Time-stop (TimedExit wrapper):**
`maxHoldBars = 150` (2 trading sessions x 75 bars/session at 5-min frequency).
Gap thesis is 1-2 session continuation. A gap that has not resolved in 2 sessions has either faded (time-stop correct) or hit TP already. 150 bars is shorter than ORB's 225-bar default because gap events are catalytic and decay faster than range breakouts.

Note: SL and TP percentages are computed at entry from the actual gapPct and passed to `NewPriceExit`. The implementation must compute these dynamically, not use a fixed percentage. `NewTimedExit(NewPriceExit(inner, stopLossPct, targetProfitPct), maxHoldBars)` is the wrapper composition order.

**Decision (2026-05.1.1) — algorithm: experimental**
scope: gap-and-go, exit-rules
tags: gap-and-go, stop-loss, take-profit, time-stop, TASK-0075
owner: marcus

Gap-and-Go exits: SL at 0.8x gapPct below entry (gap-flip signal — if gap was 1%, SL is 0.8% below entry); TP at 2.0x gapPct above entry (2:1 R/R on gap magnitude); time-stop at 150 bars (2 sessions). All implemented via PriceExit and TimedExit wrappers. Composition: `NewTimedExit(NewPriceExit(inner, stopLossPct, targetProfitPct), 150)`. SL and TP percentages computed at entry from actual gapPct. Time-stop is shorter than ORB's 225 bars because gap events decay faster than range breakouts.

## Parameter sweep axes (sensitivity analysis before universe sweep)

Run on NSE:INDHOTEL as primary orientation instrument (see decision below). Use the plateau-midpoint selection procedure (80% Sharpe floor within valid trade-count region per `decisions/algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md`).

| Axis | Values to test | Default (sweep start) |
|---|---|---|
| Gap threshold | 0.75%, 1.0%, 1.5% | 1.0% |
| Volume multiplier | 1.0x, 1.3x, 1.5x | 1.3x |
| SL multiplier | 0.5x, 0.8x, 1.0x gap | 0.8x |
| Hold bars | 75, 150, 225 | 150 (2 sessions) |

Plateau midpoint from this sweep becomes the fixed parameter for universe sweep across all 48 midcap instruments (and optionally 15 large-cap instruments).

**Decision (2026-05.1.4) — algorithm: experimental**
scope: gap-and-go, orientation-instrument
tags: gap-and-go, parameter-sweep, midcap, orientation, TASK-0075
owner: marcus

Gap-and-Go sensitivity sweep uses NSE:INDHOTEL as the primary orientation instrument, not NSE:RELIANCE. INDHOTEL is a midcap name with the target volatility profile for this strategy; it is a bootstrap survivor in the MACD midcap portfolio confirming its history is representative; it has clean 5-min data (98,906 bars confirmed, 99.5% complete per decisions/algorithm/2026-05-10-5min-data-coverage-both-universes.md). RELIANCE is acceptable as a secondary large-cap cross-check after the primary sweep. The prior convention of using RELIANCE as orientation instrument applied to daily-bar strategies on the large-cap universe; Gap-and-Go targets the midcap universe primarily.

## Sizing

**Decision (2026-05.1.3) — algorithm: experimental**
scope: gap-and-go, sizing
tags: gap-and-go, vol-targeting, position-cap, TASK-0075
owner: marcus

Vol-target 10% annualized. 20-bar rolling standard deviation of daily log returns (sample variance, consistent with Sharpe computation convention). Fraction = (0.10 / annual_vol). Capped at 1.0. Zero vol → fraction = 0 → trade skipped. Per-instrument notional = fraction x available capital.

Hard per-position cap: Rs 1.5 lakh (50% of Rs 3 lakh capital base). Rationale specific to gap-and-go: index-driven news events (budget day, RBI announcements, global macro) can produce simultaneous gap signals across multiple instruments in the same sector. Without a hard cap, 3 correlated gap-up signals on the same budget morning could deploy 100%+ of capital simultaneously. The 50% cap ensures at most 2 simultaneous positions remain within capital limits even if vol-targeting would allocate more.

Capital base: Rs 3 lakh total. Single instrument position at a time within that instrument (no pyramid). Multiple instruments can hold simultaneous positions up to the aggregate cap.

## Kill-switch thresholds

Kill-switch thresholds derived from bootstrap results per the accepted kill-switch derivation methodology (decisions/algorithm/2026-04-21-kill-switch-derivation-methodology.md). Specific values are not available until bootstrap completes. Framework:

1. Rolling per-trade Sharpe threshold: SharpeP5 from `montecarlo.Bootstrap` (mean(ReturnOnNotional) / std(ReturnOnNotional), sample variance, no annualization)
2. Maximum drawdown threshold: 1.5x in-sample MaxDrawdown
3. Drawdown duration threshold: 2x in-sample MaxDrawdownDuration

Early-warning flag (not a hard halt — manual review trigger): 3 consecutive losing trades on the same instrument. If 3 gap-up entries in a row all result in losses on one instrument, review that name independently — may have entered a gap-fill regime on that specific stock.

## Pipeline risk assessment

**Decision (2026-05.1.5) — algorithm: experimental**
scope: gap-and-go, pipeline-risk
tags: gap-and-go, walk-forward, correlation-gate, ORB, TASK-0075
owner: marcus

Pipeline risk flags in priority order:

1. **Walk-forward gate (primary risk):** The 2022 choppy market regime is the structural stress test. Gap-and-go depends on gap-up flows following through; in 2022, many intraday gaps reversed on rate shock uncertainty. If IS-period (2021-2022) shows strong Sharpe but OOS-period (late 2022-2023) shows negative Sharpe, OverfitFlag fires. 60% instrument retention (>=29 of 48 midcap instruments) is the floor at walk-forward. Expect higher attrition on Consumer/FMCG names (GODREJCP, COLPAL, DABUR) where gap persistence is weaker.

2. **ORB/Gap-and-Go mutual correlation (secondary risk, post-bootstrap):** ORB and Gap-and-Go are both 5-min CNC strategies entering on morning opening behavior and holding 2-3 days. Their equity curves may be materially correlated. If both pass bootstrap, correlation gate between them must be checked before portfolio construction. Stress-period r (COVID crash 2020, rate correction 2022) may exceed 0.6 given both are long-only and triggered on similar morning catalysts. If both pass bootstrap and fail mutual correlation gate, retain the one with higher DSR per the accepted tiebreaker rule.

3. **MACD/Gap-and-Go correlation (low risk):** MACD survivors run daily-bar trend-following (17/26/9 crossover). Gap-and-Go is event-driven intraday. Mechanistically different; stress-period correlation should be below 0.6 in most regimes. Long-only exposure creates some co-movement in crash periods, but directional overlap is lower than ORB/Gap-and-Go.

4. **Universe gate (low-moderate risk):** DSR correction for 48 instruments is punishing (nTrials=48 raises E[max SR] vs nTrials=15). The universe gate requires DSR-corrected average Sharpe > 0. Signal frequency being at the margin of the 30-trade floor means thin-sample DSR penalties could push borderline instruments below the threshold. The parameter sweep plateau selection is the mitigant — if defaults don't clear, the sweep should find a higher-frequency parameter region.

## Walk-forward structure recommendation

Per decisions/algorithm/2026-05-10-5min-data-coverage-both-universes.md, the recommended fold structure for 5-min bar strategies is:

- **IS window:** 2 years
- **OOS window:** 6 months
- **Folds:** anchored walk-forward (IS grows with each fold)
- **Coverage:** 2021-01-01 to 2026-05-09 (5.36 years, ~6 OOS periods)

This matches the daily-bar walk-forward structure and is consistent with the walk-forward gate methodology.

## Implementation notes for Priya

- Timeframe: `model.Timeframe5Min`
- Use `PreviousSessionClose(bars, i)` from `pkg/strategy/session.go` to compute the prior session close at the 09:15 bar (identified via `IsSessionOpen(bar)`)
- Gap detection and volume check happen at bar index 0 from session start (09:15 bar); entry signal generated at bar index 1 (09:20 bar close); filled at bar index 2's Open
- Volume threshold requires tracking 20-day average volume. This requires a running average state variable initialized from bar history. Pre-compute before the hot loop or use a rolling window structure initialized in `New()`.
- The no-entry "gap already chased" check (`abs(bars[1].Close - PreviousSessionClose) / PreviousSessionClose < 1.5 x gapPct`) is evaluated at bar index 1, before emitting the entry signal
- `NewPriceExit(inner, stopLossPct, targetProfitPct)` wraps the inner Gap-and-Go strategy; SL and TP percentages are computed dynamically at entry from gapPct
- `NewTimedExit(priceExitWrapped, maxHoldBars)` provides the 150-bar time-stop
- Strategy must be stateful (gap detection across bars within a session, volume average across days); `Next()` receives the full candle slice — session and gap tracking via index i
- Per the no-allocation-in-hot-loop constraint: pre-allocate volume history ring buffer at construction, not per-call
- Register in all strategy registries: `cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`
- Name: `"gap-and-go"` for registry consistency
- Universe: primary evaluation on `universes/nifty-midcap-liquid.yaml` (48 instruments); secondary check on `universes/nifty50-large-cap.yaml` (15 instruments) if midcap pipeline advances

## Related decisions

- [Overnight gap fill: engine is gap-transparent by construction](../convention/2026-05-07-overnight-gap-fill-confirmed-correct.md) — CNC fill model confirmed correct; gap P&L correctly reflected
- [ORB Marcus rules](./2026-05-13-orb-marcus-rules.md) — sibling intraday strategy; ORB/Gap-and-Go correlation gate risk documented above
- [Kill-switch derivation methodology](./2026-04-21-kill-switch-derivation-methodology.md) — framework for post-bootstrap thresholds
- [Plateau procedure: trade-count constrained](./2026-04-29-plateau-procedure-trade-count-constrained.md) — governs parameter sweep plateau selection
- [Walk-forward fold structure: 2yr IS / 6mo OOS](../algorithm/2026-05-10-5min-data-coverage-both-universes.md) — confirmed viable for 5-min bar window
- [nTrials=48 for midcap DSR correction](./2026-05-09-ntrials-48-for-midcap-dsr-correction.md) — DSR correction penalty for 48-instrument midcap universe
