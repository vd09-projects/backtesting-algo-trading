# NSE annualization factors for Sharpe and volatility calculations

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-10       |
| Status   | accepted         |
| Category | convention       |
| Tags     | NSE, annualization, sharpe, volatility, timeframe, 15min, daily, bars-per-year, analytics, convention |

## Context

Annualizing per-bar returns to compute Sharpe ratio, Sortino, and volatility requires knowing how many bars occur per trading year for each timeframe. This differs between markets. NSE (National Stock Exchange of India) has a different session length than US equity markets — the naive assumption of the US session (390 minutes/day = 26 fifteen-minute bars) produces incorrect annualization factors for NSE strategies. An initial planning document used 252×26 for 15-min NSE bars; this was incorrect and was caught before implementation.

## Decision

**NSE trading session**: 9:15 AM to 3:30 PM IST = 375 minutes/day.

**Canonical annualization factors for NSE:**

| Timeframe | Bars/day | Bars/year (252 trading days) |
|-----------|----------|------------------------------|
| Daily     | 1        | 252                          |
| 15-min    | 25       | 6,300                        |
| 5-min     | 75       | 18,900                       |
| 1-min     | 375      | 94,500                       |

**Formula**: `annualized_sharpe = mean(r) / stddev(r) * sqrt(bars_per_year)`

where `r` is the series of per-bar returns from the equity curve.

**US comparison** (for reference): 9:30 AM to 4:00 PM = 390 minutes/day = 26 fifteen-minute bars/day. Do not use US factors for NSE strategies.

## Consequences

- All implementations of Sharpe, Sortino, and volatility targeting must use the NSE-specific factors.
- The `Timeframe` enum's `Duration()` method encodes bar length; bars-per-day can be derived as `375min / Duration()` for intraday, 1 for daily. The annualization factor is `252 * bars_per_day`.
- Volatility-targeting position sizing also uses these factors: `instrumentVol_daily * sqrt(252)` annualizes daily vol. For intraday vol targeting, use the appropriate per-bar vol times `sqrt(bars_per_year)`.

## Revisit trigger

If the engine is extended to support international markets (US equities, crypto 24/7), the annualization factor must be parameterized per-market, not hardcoded to NSE values. At that point, replace the constants with a `Market` enum or a per-provider `BarsPerYear(timeframe)` method.
