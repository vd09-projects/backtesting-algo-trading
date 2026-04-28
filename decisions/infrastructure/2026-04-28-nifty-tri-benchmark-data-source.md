# NIFTY 50 TRI benchmark — data source decision

| Field    | Value          |
|----------|----------------|
| Date     | 2026-04-28     |
| Status   | accepted       |
| Category | infrastructure |
| Tags     | benchmark, NIFTY-TRI, total-return, data-source, zerodha, NSE, kite-connect, TASK-0045 |

## Context

Marcus established (prior to TASK-0045) that the correct benchmark for multi-year strategy
comparison is the NIFTY 50 Total Return Index (TRI), not the price return index. TRI includes
dividend reinvestment and accumulates ~1.3–1.5% per year above the price return index on
typical NSE dividend yields. Over a 6-year backtest window (2018–2024), this compounds to
roughly 8–10% higher cumulative return — a material gap that would make a strategy look better
or worse than buy-and-hold depending on direction.

Before any benchmark code is written, TASK-0045 required determining whether Zerodha Kite
Connect provides NIFTY 50 TRI data, and if not, documenting the two viable implementation
options.

## Instrument search methodology

The Kite Connect instruments master CSV was fetched directly from the public endpoint
(no authentication required) on 2026-04-28:

```
GET https://api.kite.trade/instruments
```

The file contained **150,306 instrument rows** across all exchanges and segments. The following
patterns were searched exhaustively against the `tradingsymbol` column:

| Pattern         | Scope                    | Matches                                |
|-----------------|--------------------------|----------------------------------------|
| `NIFTY.*TOTAL`  | All segments             | `NIFTY TOTAL MKT` (broad market cap index — unrelated) |
| `NIFTY.*TRI`    | All segments             | None                                   |
| `NIFTY.*RETURN` | All segments             | None                                   |
| `NIFTY50 TR`    | All segments             | `NIFTY50 TR 2X LEV`, `NIFTY50 TR 1X INV` |
| All INDICES segment, NIFTY | INDICES only | See full list below — no plain TRI entry |

### Complete NIFTY-related instruments in INDICES segment (abridged)

The INDICES segment was inspected in full. NIFTY 50-family instruments present:

| Token  | Tradingsymbol       | Name                    | Notes                                |
|--------|---------------------|-------------------------|--------------------------------------|
| 256265 | `NIFTY 50`          | NIFTY 50                | Price return index — available       |
| 258825 | `NIFTY50 PR 2X LEV` | NIFTY50 PR 2X LEV       | 2× leveraged price return — strategy index |
| 259081 | `NIFTY50 PR 1X INV` | NIFTY50 PR 1X INV       | Inverse price return — strategy index |
| 259337 | `NIFTY50 TR 2X LEV` | NIFTY50 TR 2X LEV       | 2× leveraged total return — strategy index |
| 259593 | `NIFTY50 TR 1X INV` | NIFTY50 TR 1X INV       | Inverse total return — strategy index |
| 265225 | `NIFTY50 DIV POINT` | NIFTY50 DIV POINT       | Cumulative dividend points index — not TRI |

## Finding

**NIFTY 50 TRI (Total Return Index) is NOT available as a direct instrument in Kite Connect.**

The `NIFTY50 TR 2X LEV` and `NIFTY50 TR 1X INV` instruments are strategy indices that reference
a total return base but are themselves 2× leveraged and inverse products respectively — they
cannot serve as the benchmark. `NIFTY50 DIV POINT` tracks cumulative dividends only (not the
compounded TRI series). Neither is the plain NIFTY 50 TRI.

This finding is consistent with the known Zerodha data model: per the existing infrastructure
decision `2026-04-10-zerodha-daily-candles-adjusted-for-corporate-actions.md`, Zerodha's day
candles for equity instruments are adjusted for corporate actions but **not total-return
adjusted** — regular dividends are market-adjusted only (price drops on ex-date but historical
prices are not restated upward by the dividend amount). The same limitation applies to index
data: `"NIFTY 50"` (token 256265) in Kite is the price return index, not TRI.

A live candle fetch to verify the `NIFTY 50` price range was not performed (access token
expired 2026-04-21). However, cached data confirms the range: the price return NIFTY 50 was
~8,284 at end of December 2014 and ~18,105 at end of December 2022 (from `.cache/zerodha/`).
The NSE-published TRI values for those dates are approximately 9,850 and 26,700 respectively
(NSE website) — a meaningful difference, confirming TRI is not equivalent to the price index
Kite provides.

## Implementation options

### Option A — External CSV loader for NSE-published TRI data (recommended)

NSE publishes historical TRI data for the NIFTY 50 at:
```
https://www.nseindia.com/products/content/equities/indices/historical_total_returns.htm
```

The data is downloadable as a CSV with daily closing TRI values from 1999 onwards. Format is
date + index value. This is the authoritative source — NSE constructs the TRI directly.

Implementation approach:
- A new `NSECSVProvider` (or a simpler `StaticCSVProvider`) reads a locally stored CSV file
  and implements a subset of `DataProvider` (specifically `FetchCandles` for a single
  instrument and daily timeframe).
- The CSV is stored in the repo under `data/benchmarks/nifty50-tri.csv` (version-controlled,
  updated manually as new data is published).
- `BenchmarkReport` consumers use this provider for the TRI benchmark instead of `FetchCandles`
  against Kite.

**Pros:**
- Authoritative data from NSE.
- No API dependency — works without Zerodha token.
- Simple implementation: file read + CSV parse, no chunking, no auth, no rate limits.
- The CSV can be committed to the repo — the TRI series is a public dataset.

**Cons:**
- Requires manual CSV update when extending the backtest window beyond the downloaded range.
- Slightly different data pipeline from the main Kite provider — two providers in use.

### Option B — Second DataProvider implementation for NSE direct data

Build a full `NSEDataProvider` implementing `DataProvider` that reads from NSE's REST API
(or a scraper) to provide TRI and other NSE-published index data on demand.

**Pros:**
- Fully automated — no manual CSV updates.
- Extensible if other NSE-only data series are needed.

**Cons:**
- NSE does not have a documented public REST API for historical index data — scraping would be
  needed. Scraping is fragile, undocumented, and may violate NSE terms of service.
- Adds a second `DataProvider` concrete implementation — significant engineering effort for
  what is a single benchmark series.
- Complexity far exceeds the need. The TRI series changes slowly (one data point per trading
  day) and is stable enough for a committed CSV file.

## Decision

**Option A — external CSV loader using NSE-published TRI data.**

The authoritative NSE TRI CSV is simple to obtain, stable, and the right scope for this use
case. The `DataProvider` interface is clean but not the right abstraction for a static, slowly
updating benchmark series that doesn't need Zerodha auth, chunking, or rate-limit handling.

Implementation next step: download the NIFTY 50 TRI CSV from NSE, store at
`data/benchmarks/nifty50-tri.csv`, and build a minimal `StaticCSVProvider` that reads it.
That work is tracked as a follow-up build task — no code is written in this spike.

**Decision (TRI data source — infrastructure: accepted)** — NSE-published CSV loader chosen
over a second DataProvider implementation. Authoritative source, no API fragility, simple
pipeline, minimal engineering footprint.

## Consequences

- The current `BenchmarkReport` uses `"NSE:NIFTY 50"` (price return). Any benchmark comparison
  made before the TRI loader is built understates the benchmark by ~8–10% over 6 years.
  This is an existing known limitation, not a new one.
- The build task for the TRI loader will create: `data/benchmarks/nifty50-tri.csv` (data file)
  and a `StaticCSVProvider` (likely in `pkg/provider/static/` or `pkg/provider/csv/`).
- `pkg/provider` package boundary: `StaticCSVProvider` must implement `DataProvider` so the
  benchmark computation path is identical regardless of TRI source.
- Manual update cadence for `nifty50-tri.csv`: once per evaluation run (download latest from
  NSE before running the full pipeline). The file should cover 2015-01-01 to present.
- A CLAUDE.md architecture note may be warranted: `pkg/provider/csv/` is a valid second
  provider implementation for static data, distinct from the Zerodha live-feed provider.

## Related decisions

- [Zerodha daily candles are adjusted for corporate actions](2026-04-10-zerodha-daily-candles-adjusted-for-corporate-actions.md) — confirms Kite's price series is not total-return adjusted; this is the root cause for needing an external TRI source
- [BenchmarkReport annualized return uses elapsed calendar time](../algorithm/2026-04-13-benchmark-cagr-from-elapsed-time.md) — TRI CAGR computation will use the same elapsed-time formula

## Revisit trigger

- If Zerodha adds a NIFTY 50 TRI instrument to their index universe (the TR 2X LEV / 1X INV
  products imply they compute TRI internally — they may expose it as a standalone index in a
  future API update), re-run the instruments search and compare candle data against the NSE CSV
  to verify.
- If NSE discontinues or moves the TRI CSV download URL, switch to fetching from NSE's
  indices data service or a reliable mirror.
