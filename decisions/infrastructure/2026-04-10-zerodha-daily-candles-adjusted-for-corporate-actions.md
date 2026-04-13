# Zerodha daily candles are adjusted for corporate actions — no adjustment layer needed

| Field    | Value          |
|----------|----------------|
| Date     | 2026-04-10     |
| Status   | accepted       |
| Category | infrastructure |
| Tags     | zerodha, data-quality, corporate-action, adjusted-prices, split, bonus, dividend, demerger, daily-candles, intraday, TASK-0025 |

## Context

TASK-0025 required verifying whether the Kite Connect historical API endpoint
(`GET /instruments/historical/{instrument_token}/day`) returns adjusted or unadjusted prices for
corporate actions before any strategy is executed against daily candle data. An unadjusted split
appears as an 80% single-bar price crash and would cause phantom signals and a corrupted equity
curve with no error at runtime.

## Verification method

Research consulted:
- Kite Connect v3 API documentation (`/instruments/historical`)
- Zerodha official Twitter (April 2026) — explicit statement on adjustment policy
- Kite Connect developer forum threads: split adjustment (thread 6921), dividend adjustment
  (threads 4014 and 14322), demerger adjustment issues (thread 15640)

**⚠ No live API call was made against a known split event.** The finding below is based on
official Zerodha statements and forum cross-references, not a direct price-continuity check.
A live sanity check is a mandatory step in TASK-0012 before any strategy result is trusted:
fetch a NIFTY 50 stock with a known split or bonus (e.g., MRF 1:10 split 2023-10-18 or
Wipro 1:1 bonus 2024-01-11) using the provider, and verify that the price series shows no
gap at the event date. If a gap is present, this decision must be revised and a corporate
action adjustment layer added before running any strategy.

## Finding

**Kite Connect `day` candles are adjusted for the following corporate actions:**

| Event                        | Day candles | Intraday candles |
|------------------------------|-------------|------------------|
| Stock splits                 | Adjusted ✓  | Not adjusted ✗   |
| Bonus issues                 | Adjusted ✓  | Not adjusted ✗   |
| Rights issues                | Adjusted ✓  | Not adjusted ✗   |
| Spin-offs                    | Adjusted ✓  | Not adjusted ✗   |
| Extraordinary dividends      | Adjusted ✓  | Not adjusted ✗   |
| Regular dividends            | Market-adjusted (price drops naturally on ex-date; Zerodha does not explicitly re-state this in price series) | Same |
| Demergers                    | Adjusted — but Zerodha's COA-based methodology may produce gaps vs. TradingView (see Caveats) | Same |

**There is no API parameter to request unadjusted data.** The Kite Connect historical endpoint
has no `adjusted=0/1` flag; `continuous` applies only to futures contracts for expired-contract
stitching. Adjustment is automatic and applied retroactively — if a split occurred in 2020 and
you fetch data today, all pre-2020 prices are already split-adjusted.

**The current provider implementation calls:**
```
GET /instruments/historical/{token}/day?from=...&to=...
```
with no additional parameters, which is the correct call for adjusted daily data.

## Caveats

**1. Retroactive adjustment creates a backtest/live price divergence.**
Adjustment is applied retroactively — prices fetched today for 2020 events are the 2020 prices
as understood from today's corporate action history. These prices *did not exist at the time*
in that form. This is the correct approach for backtesting (it gives a clean, comparable series
across the full window), but it has a critical implication for live and paper trading:

The exchange quotes unadjusted intraday prices. The backtest engine uses retroactively adjusted
historical prices. These diverge at every split/bonus event in the data window. When this engine
is ever used for paper or live trading, a signal computed from the adjusted backtest series cannot
be compared directly to the live exchange quote without applying an adjustment factor (the
cumulative split/bonus ratio from the event date to today). Failure to account for this will
produce nonsensical entries and exits.

For the current backtesting-only scope this is not a problem. It must be solved before any live
or paper trading use.

**2. Intraday candles (1min, 5min, 15min) are NOT adjusted.**
If a strategy is ever run on intraday timeframes, a corporate action adjustment layer will be
required at that point. The current daily-candle engine is unaffected.

**3. Regular dividends.**
NSE-listed stocks adjust their own market price on ex-dividend date — the opening price on
ex-date is approximately `close_prev - dividend`. Zerodha preserves this natural market
adjustment but does not further adjust the historical pre-ex-date prices upward by the dividend
amount. This means the historical series correctly reflects what a trader would have seen in the
market; it is not "total return" adjusted. For typical NSE dividend yields (0.5–3% annually),
the effect on strategy signals is negligible. Total-return calculations would require separate
dividend data.

**4. Demergers use Zerodha's Cost-of-Acquisition (COA) methodology.**
For instruments undergoing demergers (e.g., ABFRL, MARINE), Zerodha adjusts based on NSE's COA
circular, which may produce a visible price discontinuity in the chart compared to ratio-based
adjustment (used by TradingView, Yahoo Finance). This is not a data error but a methodological
difference. Demerger events are rare. If a strategy is run on a stock that demerged in the
backtest window, visually inspect the equity curve for a phantom crash near the demerger date.

**5. Adjustment lag.**
Adjustments are typically applied the weekend after the corporate action. Fetching data within
hours of a split may return the unadjusted series briefly. This is not a concern for backtesting
historical periods.

## Decision

No corporate action adjustment layer is required. The current implementation calling
`/instruments/historical/{token}/day` returns adjusted data for all materially significant
corporate actions (splits, bonuses, rights, spin-offs). TASK-0012 (SMA crossover) and TASK-0015
(RSI mean-reversion) are unblocked.

## Consequences

- TASK-0025 (gate task) is closed. TASK-0012 and TASK-0015 may proceed, with the condition that
  TASK-0012's first run includes a live price-continuity check on a known split/bonus event.
- If intraday strategies are ever introduced, this decision must be revisited — add a corporate
  action adjustment layer before the first intraday strategy run.
- Total-return calculations (including dividends) are out of scope for the current engine. Price
  return is sufficient for momentum and mean-reversion signal strategies.
- Demerger stocks: visually inspect equity curves in backtests that include ABFRL, MARINE, or
  any NSE stock that demerged in the backtest window.
- Any future extension to paper or live trading requires solving the backtest/live price
  divergence: the engine uses retroactively adjusted prices; the exchange quotes unadjusted
  intraday prices. A cumulative adjustment factor must be applied to live quotes before computing
  signals. This is out of scope for v1 but must not be forgotten.

## Related decisions

- [Zerodha corporate action verification gate](2026-04-10-corporate-action-verification-gate.md) — the gate decision that spawned TASK-0025
- [Zerodha pagination strategy](architecture/2026-04-07-zerodha-pagination-strategy.md) — chunking and rate limits for the same endpoint

## Revisit trigger

- If the first live strategy run shows an unexplained single-bar price crash or equity curve gap,
  re-verify by fetching a known split event (e.g., MRF 1:10 split 2023-10-18 or Wipro 1:1
  bonus 2024-01-11) and checking price continuity across the event date.
- If the data provider is changed (e.g., NSE direct feed, Angel One, Upstox), re-verify
  adjustment behaviour for the new source — do not assume it matches Kite Connect.
