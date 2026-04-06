# Project Task Backlog

**Last updated:** 2026-04-04 | **Open tasks:** 4 | **Next up:** TASK-0006

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

_No tasks in progress._

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0006] Analytics — P&L, win rate, and max drawdown

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-01
- **Source:** project
- **Context:** Standalone package that takes a trade log and computes performance metrics. No engine reference — pure function over []Trade.
- **Acceptance criteria:**
  - [ ] `internal/analytics/` package with `Compute(trades []Trade) Report` function
  - [ ] `Report` struct: TotalPnL, WinRate (%), MaxDrawdown (%), TradeCount, WinCount, LossCount
  - [ ] MaxDrawdown computed correctly from equity curve (not just single-trade losses)
  - [ ] Tests with known trade sequences and hand-verified expected metric values
  - [ ] Edge cases tested: empty trade log, all winners, all losers, single trade
- **Notes:** Start simple. Sharpe, Sortino, equity curves are future work — don't add them here.

---

### [TASK-0007] [ANALYSIS] Zerodha API — auth flow, rate limits, and historical data constraints

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-01
- **Source:** discovery
- **Context:** Before writing a line of Zerodha provider code, we need to know what we're actually building against. The auth flow, rate limits, max candles per request, and historical data depth all affect the implementation design. This is an analysis-only task — the output is a decision doc, not production code.
- **Acceptance criteria:**
  - [ ] Document written in `decisions/` covering: historical data depth per timeframe, rate limits (requests/sec and daily), max candles per API request, timeframes actually supported
  - [ ] Auth flow documented: how the daily access token refresh works end-to-end (login URL, request token, exchange for access token)
  - [ ] Pagination/chunking strategy decided: how to split large date ranges into valid API requests
  - [ ] Caching strategy decided: file-based vs in-memory, cache key structure, invalidation approach
  - [ ] Scratch prototype of auth flow verified to work (not production-quality — just proves the mechanism)
- **Notes:** This unlocks TASK-0008 and TASK-0009. Don't start those without finishing this.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

### [TASK-0008] Zerodha provider — auth, token refresh, and FetchCandles

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-04-01
- **Source:** project
- **Blocked by:** [TASK-0007] — need auth flow and API constraints documented before implementing
- **Context:** First concrete DataProvider implementation. Auth and basic fetch are the hard part — get these right before building the cache layer on top.
- **Acceptance criteria:**
  - [ ] `pkg/provider/zerodha/` package implementing `DataProvider` interface (compile-time check)
  - [ ] Auth: daily token refresh handled internally — not exposed through the interface
  - [ ] `FetchCandles` implemented with pagination/chunking (strategy from TASK-0007 analysis)
  - [ ] Respects rate limits (strategy from TASK-0007 analysis)
  - [ ] Mock-based test suite: no real Zerodha API calls in any test
  - [ ] Integration test with a recorded API response fixture in `testdata/`

---

### [TASK-0009] Zerodha provider — local caching layer

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-04-01
- **Source:** project
- **Blocked by:** [TASK-0007] — cache strategy (file format, key structure, invalidation) must be decided before implementing
- **Context:** Iterative backtesting means the same candle data gets fetched repeatedly. A local cache avoids API hammering and speeds up development cycles significantly.
- **Acceptance criteria:**
  - [ ] Cache layer wraps the Zerodha fetcher transparently — DataProvider interface unchanged
  - [ ] Cache hit returns stored candles without an API call
  - [ ] Cache miss fetches from API, stores result, returns candles
  - [ ] Cache key includes: instrument, timeframe, date range
  - [ ] Cache invalidation: at minimum, manual clear; ideally, TTL-based for recent data
  - [ ] Tests: verify cache hit prevents API call (mock verifies zero API calls on second fetch)
- **Notes:** Cache storage format TBD from TASK-0007 — likely JSON or CSV files in a local `.cache/` directory.

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

_No backlog items._

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
