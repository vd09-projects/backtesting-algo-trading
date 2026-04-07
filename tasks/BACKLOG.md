# Project Task Backlog

**Last updated:** 2026-04-07 | **Open tasks:** 5 | **Next up:** TASK-0008

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

_No tasks in progress._

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0008] Zerodha provider — auth, token refresh, and FetchCandles

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-01
- **Source:** project
- **Context:** First concrete DataProvider implementation. Auth and basic fetch are the hard part — get these right before building the cache layer on top.
- **Acceptance criteria:**
  - [ ] `pkg/provider/zerodha/` package implementing `DataProvider` interface (compile-time check)
  - [ ] Auth: daily token refresh handled internally — not exposed through the interface
  - [ ] `FetchCandles` implemented with pagination/chunking (strategy from TASK-0007 analysis)
  - [ ] Respects rate limits (strategy from TASK-0007 analysis)
  - [ ] Mock-based test suite: no real Zerodha API calls in any test
  - [ ] Integration test with a recorded API response fixture in `testdata/`
- **Notes:** Auth strategy and pagination decisions in `decisions/infrastructure/`. Token persisted to `~/.config/backtest/token.json` with 6AM IST expiry. Chunk strategy: per-interval maxDays map, 350ms sleep between chunks.

---

### [TASK-0009] Zerodha provider — local caching layer

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-01
- **Source:** project
- **Context:** Iterative backtesting means the same candle data gets fetched repeatedly. A local cache avoids API hammering and speeds up development cycles significantly.
- **Acceptance criteria:**
  - [ ] Cache layer wraps the Zerodha fetcher transparently — DataProvider interface unchanged
  - [ ] Cache hit returns stored candles without an API call
  - [ ] Cache miss fetches from API, stores result, returns candles
  - [ ] Cache key includes: instrument, timeframe, date range
  - [ ] Cache invalidation: at minimum, manual clear; ideally, TTL-based for recent data
  - [ ] Tests: verify cache hit prevents API call (mock verifies zero API calls on second fetch)
- **Notes:** File-based JSON in `.cache/zerodha/`, keyed on exact (instrument, timeframe, from, to). TTL only for recent data. Decision recorded in `decisions/infrastructure/2026-04-07-zerodha-cache-strategy.md`.

---

### [TASK-0010] Output package — result formatting and JSON export

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-07
- **Source:** project
- **Context:** `internal/analytics/` produces a `Report`; `internal/output/` needs to format it for human consumption and persist it as JSON. Without this, backtest results only exist in memory.
- **Acceptance criteria:**
  - [ ] `internal/output/` package with `Write(report analytics.Report, cfg OutputConfig) error`
  - [ ] Human-readable text summary printed to stdout (P&L, win rate, max drawdown, trade count)
  - [ ] JSON export to a configurable output file path
  - [ ] `OutputConfig` specifies: output file path, whether to print to stdout
  - [ ] Tests: known `Report` → expected JSON output (table-driven)
  - [ ] Edge case: empty report (zero trades) produces valid output, not a crash

---

### [TASK-0011] CLI entrypoint — cmd/backtest wiring

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-07
- **Source:** project
- **Context:** The engine, analytics, and output packages all exist but nothing wires them together into a runnable command. This task produces the first end-to-end `go run cmd/backtest` invocation.
- **Acceptance criteria:**
  - [ ] `cmd/backtest/main.go` parses flags: `--instrument`, `--from`, `--to`, `--timeframe`, `--cash`, `--strategy`, `--out` (output file)
  - [ ] Wires provider → engine → analytics → output in the correct order
  - [ ] Strategy selected by name from a registry of available strategies (at minimum one stub)
  - [ ] Graceful error handling: invalid flags, provider errors, engine errors each produce a useful message and non-zero exit code
  - [ ] `go run cmd/backtest --help` documents all flags
- **Notes:** Depends on TASK-0008 (provider) and TASK-0012 (first strategy). Can be started structurally before those are done, with a stub provider and stub strategy.

---

### [TASK-0012] First concrete strategy — SMA crossover

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-07
- **Source:** project
- **Context:** The engine needs at least one real strategy to prove the full pipeline end-to-end. SMA crossover (fast SMA crosses above/below slow SMA) is the simplest meaningful strategy and a good integration test of the go-talib dependency.
- **Acceptance criteria:**
  - [ ] `strategies/smacrossover/` package implementing the `Strategy` interface
  - [ ] Uses `github.com/markcheno/go-talib` for SMA computation — no hand-rolled math
  - [ ] Configurable fast and slow period (e.g. 10/50); defaults baked in
  - [ ] Lookback returns `slowPeriod` so the engine starts feeding at the right bar
  - [ ] Tests: known OHLCV sequence → expected signal at each bar (table-driven)
  - [ ] No import of any concrete type outside `pkg/` — only interfaces and model types

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

_No blocked tasks._

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

_No backlog items._

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
