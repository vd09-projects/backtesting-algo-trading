# Project Task Backlog

**Last updated:** 2026-05-07 | **Open tasks:** 24 | **Next up:** TASK-0072

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

<!-- empty -->

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0072] Data — Nifty Midcap 150 universe YAML

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** All strategies tested so far run only on Nifty50 large-caps — the most analyst-covered, institutionally-traded names in India. Edge thesis is weaker there. Midcap names have thinner analyst coverage, more retail participation, and more persistent behavioral inefficiencies. Same daily-bar pipeline applies, zero new infrastructure needed.
- **Acceptance criteria:**
  - [ ] `universes/nifty-midcap-liquid.yaml` created in same format as `universes/nifty50-large-cap.yaml`
  - [ ] 20–30 Nifty Midcap 150 names selected: continuous Kite daily bar history from 2018-01-01, reasonably liquid (ADV > ₹50 crore), no pending delistings or recent IPOs
  - [ ] Marcus (algo-trading-veteran) reviews final list for liquidity and data quality before first sweep run
  - [ ] `cmd/universe-sweep --universe universes/nifty-midcap-liquid.yaml` runs without errors
- **Notes:** Owner: Priya builds YAML, Marcus reviews instrument list. Zero code — purely a data/config task. Unblocks midcap daily sweep immediately after review. Same 2018-01-01 to 2024-12-31 evaluation window applies.

---

### [TASK-0059] Engine — walk-forward `Run()` factory API for stateful strategy wrappers

- **Status:** todo
- **Priority:** high
- **Created:** 2026-04-27
- **Source:** session
- **Context:** `TimedExit` (added in TASK-0039) is stateful — it tracks `entryBar` and `inPosition` between `Next()` calls. The walk-forward harness currently accepts a single `strategy.Strategy` instance, safe only when stateless. Using `TimedExit` with walk-forward today silently produces corrupted results across fold boundaries. This is a blocker for any intraday strategy that uses the TimedExit wrapper (all planned 2-3 day hold strategies).
- **Acceptance criteria:**
  - [ ] `internal/walkforward.Run()` signature changed to accept `factory func() strategy.Strategy` instead of a single `strategy.Strategy` instance
  - [ ] Each fold constructs a fresh strategy instance via `factory()` — no shared state across folds
  - [ ] Existing callers updated: stateless strategies pass `func() strategy.Strategy { return myStrategy }` closures
  - [ ] All 17 existing walk-forward tests still pass with race detector
  - [ ] New test: `TimedExit`-wrapped strategy used in walk-forward — verify fold 2 starts with clean position state
  - [ ] Godoc on `Run()` updated to remove the concurrent-safety caveat
  - [ ] Tests written before implementation (TDD)
- **Notes:** Priority bumped from medium → high on 2026-05-04: all planned intraday strategies (ORB, gap-and-go) will use TimedExit wrapper, making this a blocker for the intraday pipeline. Breaking API change — scan all callers in `cmd/` before implementing. Owner: Priya (dev).

---

### [TASK-0071] Engine — verify overnight gap handling for intraday CNC backtests

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** The engine processes bars sequentially. On 5-min bars with CNC overnight holds, the bar at 3:25 PM is immediately followed by 9:15 AM next day — a real-world 17-hour gap invisible to the event loop. Three concerns: (1) P&L must capture the overnight gap correctly (next open vs prior close); (2) stop-loss fills must use the gap-down open price, not the stop level; (3) any trade log metrics that assume uniform bar spacing may be wrong. Must be verified with a golden test before any intraday backtest result is trusted.
- **Acceptance criteria:**
  - [ ] Read `internal/engine/` event loop and portfolio accounting: document exactly where P&L is computed and how fill price is determined for the bar following an overnight gap
  - [ ] Write golden test: synthetic 5-min candle series spanning 2 sessions with a 3% gap-down open on day 2; position entered on day 1 close; verify P&L in trade log equals gap-adjusted loss, not zero
  - [ ] Write golden test: synthetic strategy that signals position-close at bar N where bar N+1 opens with a 3% gap below the signal exit price — verify engine fills at bar N+1 open price, not at the signal price from bar N
  - [ ] If bugs found: fix before any intraday strategy is evaluated; record fix in `decisions/`
  - [ ] If engine already handles this correctly: record confirmed-correct note in `decisions/`
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). This is a blocker for any intraday backtest producing valid results. CNC strategies (2-3 day holds) don't need forced session close (that's TASK-0046 for MIS), but they do need correct gap accounting. Run this before TASK-0070 (fetch-history) to avoid running large fetches before we know backtests are valid.

---

### [TASK-0078] Infrastructure — session-boundary utilities for intraday strategies

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-05
- **Source:** session
- **Context:** Both ORB (TASK-0074) and gap-and-go (TASK-0075) need to detect "is this the first bar of today's NSE session?" and "what was the close of the last bar of the previous session?" No utility exists for this. Without it, both strategies will independently hardcode IST timestamp parsing — duplicated, untestable, fragile on NSE holidays and half-days.
- **Acceptance criteria:**
  - [ ] `pkg/strategy/session.go`: `IsSessionOpen(bar model.Candle) bool` — returns true if bar timestamp is 09:15 IST (first bar of NSE session)
  - [ ] `pkg/strategy/session.go`: `PreviousSessionClose(bars []model.Candle, i int) (float64, bool)` — scans backward from index i to find last bar before 09:15 IST on a prior trading day; returns (close, true) or (0, false) if no prior session in slice
  - [ ] IST timezone uses `time.FixedZone("IST", 5*3600+30*60)` — no tzdata dependency, consistent with engine convention
  - [ ] Golden test: synthetic 5-min slice spanning 2 sessions — `IsSessionOpen` fires exactly once per day (09:15 bar only); `PreviousSessionClose` returns correct prior-session close from any intraday bar on day 2
  - [ ] Weekend/holiday gap test: gap of >1 calendar day between sessions — `PreviousSessionClose` returns most recent prior-session close correctly
  - [ ] `golangci-lint run ./pkg/strategy/...` passes
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Unblocks TASK-0074 and TASK-0075 — add as their shared dependency. Single file, no engine changes, pkg/strategy only.

---

### [TASK-0070] Tooling — `cmd/fetch-history` CLI for bulk intraday historical data

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** Zerodha Kite serves 5-min data back to ~2015 (confirmed from official developer forum). Per-request limit is 100 days per call, but total depth is ~10 years. Existing `FetchCandles` + `chunkDateRange` already handle multi-request fetching automatically. A one-shot CLI to drain full history for all instruments in a universe YAML — writing to the existing `CachedProvider` disk cache — enables all future intraday backtests to run from local files without a Zerodha token.
- **Acceptance criteria:**
  - [ ] `cmd/fetch-history/main.go` CLI: flags `--universe`, `--timeframe` (repeatable), `--from`, `--cache-dir`, `--api-key`, `--access-token` (or KITE_API_KEY / KITE_ACCESS_TOKEN env vars matching existing CLI convention)
  - [ ] Reads universe YAML, iterates instruments × timeframes, calls `FetchCandles` for `[--from, today)`
  - [ ] Writes results via `CachedProvider` (existing `pkg/provider/zerodha/cache/`) so cache keys match what `cmd/backtest` and `cmd/universe-sweep` expect
  - [ ] Incremental: uses `CachedProvider.LastCachedTime()` (from TASK-0080) to skip already-fetched ranges; fetches only delta from last cached date to today
  - [ ] Partial-failure recovery: on fetch error, writes `fetch-manifest.json` recording last successfully fetched instrument+timeframe+date; subsequent runs resume from manifest rather than restarting from `--from`
  - [ ] Progress logging: prints `instrument × timeframe: fetched N candles [from → to]` per chunk so long runs are observable
  - [ ] Dry-run flag `--dry-run`: prints what would be fetched without hitting the API
  - [ ] Auth flags and env var fallback covered by tests (mock provider in tests)
  - [ ] Tests written before implementation (TDD); at minimum: dry-run output, partial-failure manifest write, resume-from-manifest
- **Notes:** Owner: Priya (dev). Blocked at runtime on Zerodha access token — no code blocker. Incremental delta fetch depends on TASK-0080 (CachedProvider manifest). Until TASK-0080 is complete, fetch-history CLI fetches full range from --from on every run (no incremental mode).

---

### [TASK-0073] Tooling — end-to-end automated evaluation pipeline (`cmd/evaluate`)

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-04
- **Source:** session
- **Context:** Running the full evaluation pipeline (universe sweep → walk-forward → bootstrap) currently requires manual handoff between three CLIs. With multiple strategies and timeframes in play, this is slow and error-prone. A single CLI that runs the full sequence, writes structured outputs to a dated results folder, and produces a summary verdict removes all manual steps. Gates and methodology remain unchanged.
- **Acceptance criteria:**
  - [ ] `cmd/evaluate/main.go` CLI: flags `--strategy`, `--params` (key=value pairs), `--universe`, `--timeframe`, `--from`, `--to`, `--out-dir`
  - [ ] Runs full sequence: (1) universe sweep with DSR gate, (2) walk-forward on survivors, (3) bootstrap on walk-forward survivors
  - [ ] If universe sweep produces zero survivors, pipeline halts immediately and writes `verdict.json` with `"result": "killed_at_universe_gate"` — does not proceed to walk-forward
  - [ ] Each stage writes outputs to `--out-dir/YYYY-MM-DD-{strategy}-{timeframe}/` in same format as existing CLIs
  - [ ] Summary `verdict.json` written at end: lists survivors with gate results, kills with stage and reason
  - [ ] Existing gate thresholds unchanged — no new methodology; parameter search is a separate CLI (TASK-0077)
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). The parameter-sweep mode must enforce DSR-corrected ranking — not raw Sharpe maximization. Marcus's standing order: parameter search on training window only, DSR-corrected rank, OOS untouched during search.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

### [TASK-0046] Engine — session-boundary support for intraday backtesting

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** MIS strategies only — forced session close is not needed for CNC 2-3 day holds (the current intraday focus). TASK-0046 becomes relevant only if MIS (same-day close) strategies are built. Methodology questions resolved: Marcus answered both in session 2026-04-25 (Decision 2026-04.3.0: forced-close at 3:15 PM bar Close; Decision 2026-04.3.1: session detection via IST timestamp ≥ 15:15). Ready to build when MIS strategy work begins.
- **Context:** The engine event loop has no concept of a trading session. For intraday (MIS) strategies, any open position must be closed by 3:15 PM IST or Zerodha auto-squares it at a random market price. Without this logic, intraday backtests are invalid.
- **Acceptance criteria:**
  - [ ] `SessionConfig` struct added: `Exchange string`, `Timezone *time.Location`, `SessionEndTime time.Time` (local time-of-day)
  - [ ] `engine.Config` gains optional `Session *SessionConfig` (nil = no session boundary, current behavior preserved)
  - [ ] `isLastBarOfSession(bar model.Candle, cfg *SessionConfig) bool` helper in `internal/engine/`
  - [ ] Event loop: after applying pending signal, if `isLastBarOfSession` returns true and a position is open, force-close at the configured fill price
  - [ ] Golden test: 2-day intraday candle series with position open at session end → forced close on day 1, correct equity and trade log
  - [ ] Timezone-aware tests covering IST session boundaries
  - [ ] Tests written before implementation (TDD)
- **Notes:** Significant engine change. Golden tests mandatory for any event loop modification. `Session *SessionConfig` being optional (nil pointer) preserves all existing daily-bar tests without modification.

---


### [TASK-0074] Strategy — Opening Range Breakout (5-min, CNC overnight hold)

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-05-04
- **Source:** session
- **Blocked by:** Marcus (algo-trading-veteran) must define entry/exit rules before implementation
- **Context:** Intraday strategy for 5-min bars with 2-3 day CNC holds. Thesis: the first 30-60 minutes of the NSE session define price discovery; a clean breakout from that range in the first hour tends to persist intraday and sometimes into the next session. TimedExit wrapper provides the N-day time-stop for flat/sideways positions.
- **Acceptance criteria:**
  - [ ] Marcus (algo-trading-veteran) rules on whether strategy is long-only or bidirectional — decision recorded in `decisions/algorithm/` before implementation begins
  - [ ] Marcus defines: range window duration (30 / 45 / 60 min), breakout confirmation method (close above/below? volume threshold?), time-stop N (days), position sizing rule
  - [ ] `strategies/orb/` package implementing `Strategy` interface: range computed from first N 5-min bars using `IsSessionOpen()` from TASK-0078, long on close above high, exit on time-stop or target
  - [ ] Uses `pkg/strategy/timed_exit.go` wrapper for N-day time-stop
  - [ ] CLI registered in all strategy registries (`cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`)
  - [ ] All public functions tested; golden test for range computation and signal generation
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Marcus (edge definition) → Priya (implementation). Depends on TASK-0071 (gap handling verified), TASK-0059 (walk-forward factory API), and TASK-0078 (session-boundary utilities) before implementation begins.

---

### [TASK-0075] Strategy — Gap-and-Go (5-min, CNC overnight hold)

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-05-04
- **Source:** session
- **Blocked by:** Marcus (algo-trading-veteran) must define entry/exit rules before implementation
- **Context:** Intraday strategy for 5-min bars with 1-2 day CNC holds. Thesis: NSE large/midcap stocks opening 1-2%+ above/below prior close on above-average volume tend to continue in the gap direction for 1-2 sessions before reversion. Captures institutional order flow from overnight news. TimedExit provides the time-stop if the move stalls.
- **Acceptance criteria:**
  - [ ] Marcus defines: gap threshold % (e.g. 1.0%), volume threshold (e.g. 1.5× 20-day average), entry bar (open of gap bar? first 5-min close?), time-stop N (days)
  - [ ] `strategies/gapandgo/` package implementing `Strategy` interface: computes prior close from last bar of previous session, detects gap condition on first bar of new session, enters in gap direction
  - [ ] Uses `pkg/strategy/timed_exit.go` wrapper for N-day time-stop
  - [ ] CLI registered in all strategy registries
  - [ ] All public functions tested; golden test covering gap-up enter, gap-down enter, no-gap skip
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Marcus (edge definition) → Priya (implementation). Requires TASK-0071 (gap handling verified) and TASK-0078 (session-boundary utilities — `PreviousSessionClose` is the primary dependency here). Long-only initially.

---

### [TASK-0077] Tooling — parameter optimization with DSR correction (`cmd/param-search`)

- **Status:** blocked
- **Priority:** low
- **Created:** 2026-05-04
- **Source:** session
- **Blocked by:** TASK-0073 (cmd/evaluate pipeline must exist first)
- **Context:** Grid-search tool finding DSR-corrected optimal parameters for a strategy. Extends existing `internal/sweep2d` infrastructure. Critical constraint: search runs on training window only; OOS window never touched during search; ranking by DSR-corrected Sharpe, not raw Sharpe. Without these constraints the tool is a professional overfitting engine.
- **Acceptance criteria:**
  - [ ] `cmd/param-search/main.go`: flags `--strategy`, `--param-grid` (JSON file defining axes and ranges), `--universe`, `--timeframe`, `--train-from`, `--train-to`, `--out-dir`
  - [ ] Grid search runs exclusively on `[--train-from, --train-to]` window
  - [ ] DSR correction applied to all variants (number of trials = grid size); ranking by DSR-corrected Sharpe, not raw Sharpe
  - [ ] OOS date range not accepted as a flag — caller must run `cmd/evaluate` separately on winning params; architectural enforcement, not convention
  - [ ] Top-N results written to `--out-dir/param-search-results.csv` with DSR, raw Sharpe, trade count per variant
  - [ ] `--param-grid` JSON schema documented in cmd/param-search/README.md or flag help text
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Marcus standing order 2026-05-04: "parameter search on training window only, DSR-corrected rank, OOS untouched during search." No OOS flag is the architectural enforcement — not a docs warning.

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

### [TASK-0058] Tooling — fix cyclomatic complexity in `cmd/rsi-diagnostic/main.go`

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-27
- **Source:** discovery
- **Context:** `cmd/rsi-diagnostic/main.go` `main()` function has cyclomatic complexity 17, exceeding the project's golangci-lint cyclop limit of 15. Discovered during TASK-0043 build session — the file was pre-existing, not introduced by TASK-0043. The fix pattern is established: extract strategy-dispatch and parameter-parsing logic into named helper functions, matching the refactor applied to `cmd/sweep/main.go` in TASK-0043 (smaFactory, rsiFactory, donchianFactory extraction).
- **Acceptance criteria:**
  - [ ] `golangci-lint run ./cmd/rsi-diagnostic/...` reports 0 issues
  - [ ] `go1.25.0 test -race ./...` still passes
  - [ ] No behavioral changes — refactor only
- **Notes:** The same cyclop issue does NOT exist in cmd/backtest or cmd/sweep after TASK-0043 refactored sweep's factoryRegistry. rsi-diagnostic is the only remaining offender.

---

### [TASK-0062] Tooling — NIFTY 50 TRI benchmark: download CSV and implement StaticCSVProvider

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-04-28
- **Source:** decision
- **Context:** TASK-0045 (research spike) confirmed NIFTY 50 TRI is not available via Zerodha Kite Connect. Decision `2026-04-28-nifty-tri-benchmark-data-source.md` chose Option A: NSE-published CSV loader. This task implements that decision — download the authoritative TRI CSV from NSE and build a minimal `StaticCSVProvider` so the benchmark computation path is provider-agnostic.
- **Acceptance criteria:**
  - [ ] `data/benchmarks/nifty50-tri.csv` downloaded from NSE (nseindia.com/products/content/equities/indices/historical_total_returns.htm) covering 2015-01-01 to present; committed to repo
  - [ ] `pkg/provider/csv/` package created with `StaticCSVProvider` implementing `provider.DataProvider` for a single instrument (daily timeframe only)
  - [ ] `StaticCSVProvider` returns `ErrUnsupportedTimeframe` for non-daily timeframes and `ErrInstrumentNotFound` for instruments not in the loaded file
  - [ ] `BenchmarkReport` computation wired to use `StaticCSVProvider` for the TRI benchmark when `--benchmark-tri` flag is set (or equivalent)
  - [ ] Tests written before implementation (TDD); `go1.25.0 test -race ./pkg/provider/csv/...` passes
  - [ ] `golangci-lint run ./pkg/provider/csv/...` passes
- **Notes:** `StaticCSVProvider` should satisfy `provider.DataProvider` at compile time via a `var _ provider.DataProvider = (*StaticCSVProvider)(nil)` check. NSE CSV columns: Date, Open, High, Low, Close (or just Index Value for TRI — inspect the actual download first). TRI values will be in the 9,000–28,000 range for 2015–2024. No chunking, no auth, no rate limits needed.

---

### [TASK-0079] Tech debt — centralized strategy registry

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-05
- **Source:** discovery
- **Context:** Every new strategy requires manual registration in 4+ CLI files: `cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`, `cmd/sweep` (plus `cmd/sweep2d` per TASK-0061). Six strategies already, two more incoming (TASK-0074, TASK-0075). Forgetting any one registration produces silent wrong behaviour — strategy silently unavailable — not a compile error. Maintenance tax compounds with every addition.
- **Acceptance criteria:**
  - [ ] `internal/cmdutil/registry.go`: `StrategyRegistry` map type with `Register(name string, factory func() strategy.Strategy)` and `MustGet(name string) func() strategy.Strategy` (panics on unknown name at startup, not silently at runtime)
  - [ ] `internal/cmdutil/strategies.go`: single authoritative list of all strategy registrations — one entry per strategy, one file to update when adding a new strategy
  - [ ] `cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`, `cmd/sweep` all consume the central registry; local maps removed
  - [ ] Adding a new strategy requires exactly one file change in one location; no `init()` auto-registration (violates CLAUDE.md no-global-state rule)
  - [ ] `TestStrategyRegistry` covers: known strategies return non-nil factory, unknown strategy panics with descriptive message, `ListStrategies()` returns sorted names
  - [ ] `golangci-lint run ./...` and `go1.25.0 test -race ./...` pass
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Tech debt. No new dependencies. `init()` pattern explicitly rejected per repo rules — use explicit registration in `internal/cmdutil/strategies.go`. Related: TASK-0061 extends sweep2d; that extension should also consume the central registry when done.

---

### [TASK-0090] Tech debt — CachedProvider test: add corrupt-superset fallback coverage

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-07
- **Source:** discovery
- **Context:** TASK-0089 added range-aware superset lookup to `CachedProvider.FetchCandles`. The fallback path — where `findSupersetFile` returns a match but `readCache` fails (corrupt or TTL-expired superset file) — is not covered by any test. The existing `TestCorruptCacheFile` only exercises the exact-match corrupt path. The fallback behavior (fall through to network) is correct but untested.
- **Acceptance criteria:**
  - [ ] `TestSupersetHit_CorruptSupersetFallback` added to `pkg/provider/zerodha/cache/cache_test.go`: write a corrupt wide cache file (non-JSON bytes), request a narrow range, assert inner provider is called exactly once (fallback to network), and correct candles are returned
  - [ ] `go1.25.0 test -race ./pkg/provider/zerodha/cache/...` passes
  - [ ] `golangci-lint run ./pkg/provider/zerodha/cache/...` still passes
- **Notes:** Discovered during go-quality-review standard gate on TASK-0089. Low priority: fallback behavior is identical to a regular cache miss; this is a coverage gap on a defensive path, not a behavioral gap. Single test in `cache_test.go` — no production code changes needed. **Session context (2026-05-07):** The same session also implemented lazy auth in `internal/cmdutil/cmdutil.go` — `lazyProvider` wraps the Zerodha client init behind `sync.Once` so token load is deferred to first cache miss; full cache hits now bypass auth entirely. Decisions recorded: `decisions/architecture/2026-05-07-lazy-provider-pattern-defer-auth-to-cache-miss.md` and `decisions/tradeoff/2026-05-07-init-fn-uses-background-context-not-caller-ctx.md`.

---

### [TASK-0080] Tech debt — CachedProvider incremental time-series manifest

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-05
- **Source:** discovery
- **Context:** `CachedProvider` stores whole-range files keyed on (instrument, timeframe, from, to) tuples — correct for "backtest same range twice" but unable to support incremental accumulation. `cmd/fetch-history` (TASK-0070) needs "last cached candle timestamp per instrument+timeframe" to fetch only the delta. No API exists for this today. Without it, every fetch-history run re-fetches the full date range.
- **Acceptance criteria:**
  - [ ] `pkg/provider/zerodha/cache/manifest.go`: `Manifest` struct with `LastCandleTime time.Time`; serialised as `fetch-manifest.json` in the instrument's cache subdirectory (e.g. `cache/nse_infy/5min/fetch-manifest.json`)
  - [ ] `CachedProvider.RecordFetch(instrument string, tf model.Timeframe, lastCandleTime time.Time) error`: writes/updates manifest after successful fetch
  - [ ] `CachedProvider.LastCachedTime(instrument string, tf model.Timeframe) (time.Time, bool)`: reads manifest; returns (zero, false) if manifest absent
  - [ ] Manifest writes are atomic: write to `.tmp` file then `os.Rename` — partial write cannot corrupt existing manifest
  - [ ] Existing `FetchCandles` cache behaviour unchanged — manifest is additive
  - [ ] Concurrent-access test with race detector: two goroutines calling `RecordFetch` simultaneously — no corruption
  - [ ] `golangci-lint run ./pkg/provider/zerodha/cache/...` passes
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Tech debt unblocking TASK-0070 incremental mode. Atomic rename pattern: `os.WriteFile` to `path+".tmp"`, then `os.Rename(tmp, path)` — POSIX-atomic on Linux/macOS. TASK-0070 incremental AC is explicitly gated on this task.

---

### [TASK-0057] Engine — migrate accounting layer from float64 to shopspring/decimal

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-25
- **Source:** decision
- **Context:** The commission arithmetic in `commission.go` and the broader accounting layer (Portfolio.cash, Trade.RealizedPnL, Trade.Commission, EquityPoint.Value) all use float64. Accumulated rounding errors are negligible for backtesting but not acceptable for live execution accounting. This migration must be coordinated — partial decimal adoption creates a worse inconsistency than uniform float64.
- **Acceptance criteria:**
  - [ ] `shopspring/decimal` added to `go.mod` (requires explicit approval per CLAUDE.md no-new-deps rule — confirm before implementation)
  - [ ] `commission.go` migrated: all intermediate calculations use `decimal.Decimal`; final return values converted to float64 only at the portfolio accounting boundary
  - [ ] `portfolio.go`: `cash` field migrated to `decimal.Decimal`
  - [ ] `pkg/model/trade.go`: `RealizedPnL`, `Commission` fields migrated to `decimal.Decimal`
  - [ ] `pkg/model/equity.go`: `EquityPoint.Value` migrated to `decimal.Decimal`
  - [ ] All existing tests pass with race detector after migration
  - [ ] Golden tests in `commission_zerodha_full_test.go` updated to use exact decimal comparisons
  - [ ] Benchmark (`BenchmarkEngineRun`) remains within 1ms/op budget after migration
- **Notes:** Coordinated migration — do not migrate commission.go alone. Deferred from TASK-0038 per decision `2026-04-25-float64-for-commission-arithmetic`. `shopspring/decimal` dependency must be discussed with the user before implementation per the no-new-dependencies rule in CLAUDE.md.

---

### [TASK-0037] Rigor — bootstrap re-run to fill kill-switch p5 Sharpe thresholds

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-21
- **Source:** session
- **Context:** TASK-0026 documented drawdown and duration kill-switch thresholds for SMA crossover and RSI mean-reversion, but the bootstrap p5 Sharpe threshold is PENDING for both. The CLI commands are ready; the Zerodha token needs to be refreshed to run them.
- **Acceptance criteria:**
  - [ ] Run `go run ./cmd/backtest --strategy sma-crossover ... --bootstrap` (full command in `decisions/algorithm/2026-04-21-kill-switch-sma-crossover.md`)
  - [ ] Run `go run ./cmd/backtest --strategy rsi-mean-reversion ... --bootstrap` (full command in `decisions/algorithm/2026-04-21-kill-switch-rsi-mean-reversion.md`)
  - [ ] Paste the `Per-trade Sharpe p5` value from each run into the respective decision file, replacing `PENDING`
  - [ ] Update decision file status from `accepted` (PENDING) to reflect actual values
- **Notes:** Both strategies failed the proliferation gate — these thresholds are reference values, not live deployment approval. With only 7 and 22 trades respectively, the p5 Sharpe will have wide confidence intervals. Document that caveat alongside the values.

---

### [TASK-0061] Tooling — extend `cmd/sweep2d` factoryRegistry to all 6 strategies

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-27
- **Source:** session
- **Context:** `cmd/sweep2d/main.go` was built in TASK-0044 with `sma-crossover` and `rsi-mean-reversion` only ("extend as new strategies land"). The remaining four strategies (donchian-breakout, macd-crossover, bollinger-mean-reversion, momentum) need 2D axis mappings added to `factoryRegistry2D`. The `fixedParams` struct is also duplicated between `cmd/sweep` and `cmd/sweep2d` — each new strategy requires updating both files. Consider extracting to `internal/cmdutil` or a shared cmd-layer type at this point.
- **Acceptance criteria:**
  - [ ] `factoryRegistry2D` in `cmd/sweep2d/main.go` handles all 6 strategies
  - [ ] Axis mappings documented in code comments: donchian (p1=period, p2=tbd), macd (p1=fast, p2=slow), bollinger (p1=period, p2=num-std-dev), momentum (p1=lookback, p2=threshold)
  - [ ] `fixedParams` struct duplication between `cmd/sweep` and `cmd/sweep2d` resolved — either extracted to shared location or duplication accepted with a comment
  - [ ] All new factory paths covered by `TestFactoryRegistry2D_KnownStrategies`
  - [ ] `golangci-lint run ./cmd/sweep2d/...` still passes
- **Notes:** Donchian has only one meaningful sweep parameter (period) — its p2 axis is less obvious; defer the axis mapping decision until this task is picked up.

---

### [TASK-0076] Model — add Timeframe30Min and Timeframe60Min

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-04
- **Source:** session
- **Context:** Kite Connect serves 30-min and 60-min bars. Neither is currently in `pkg/model/timeframe.go`. Adding them unblocks hourly-bar strategy testing — useful for strategies that need more resolution than daily but less noise than 5-min.
- **Acceptance criteria:**
  - [ ] `Timeframe30Min` and `Timeframe60Min` constants added to `pkg/model/timeframe.go` with correct `Duration()` implementations
  - [ ] `maxDaysPerInterval` in `pkg/provider/zerodha/chunk.go` updated (Kite limits: 30-min ≈ 200 days, 60-min ≈ 400 days — verify against Kite docs before committing)
  - [ ] `timeframeToInterval` and `SupportedTimeframes` in `pkg/provider/zerodha/provider.go` updated
  - [ ] `provider_test.go` updated: supported timeframe count increases from 4 to 6
  - [ ] `pkg/provider/zerodha/chunk_test.go` updated to include 30-min and 60-min chunk-window cases
  - [ ] `lazyProvider.SupportedTimeframes()` in `internal/cmdutil/cmdutil.go` updated to include `Timeframe30Min` and `Timeframe60Min` — this hardcoded list does not auto-update from the Zerodha provider; missing entries here means cached runs will not advertise the new timeframes
  - [ ] `golangci-lint run ./...` and `go1.25.0 test -race ./...` pass
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Small change — 3 files, ~20 lines total. Verify exact Kite API limits for 30-min and 60-min before setting chunk sizes. The `lazyProvider` AC above is a maintenance trap introduced in 2026-05-07 (lazy auth fix) — the hardcoded list in `cmdutil.go` is the only place that doesn't derive from `provider.go`.

---

### [TASK-0036] Research tooling — Python notebooks layer + file contract

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-16
- **Source:** session
- **Context:** The 2D heatmap, equity curve plots, and regime visualizations have nowhere to live.
  A `notebooks/` directory with a documented file contract is the prerequisite for any
  visualization work and establishes the Go-writes/Python-reads boundary explicitly.
- **Acceptance criteria:**
  - [ ] `notebooks/` directory at project root, version-controlled
  - [ ] `notebooks/README.md` documents file contract: equity curve CSV schema, sweep CSV schema, analytics JSON schema, column names, timestamp format
  - [ ] `notebooks/requirements.txt` with pyarrow, pandas, matplotlib pinned
  - [ ] At least one working notebook: `notebooks/equity-curve.ipynb` reads `runs/<name>-curve.csv` and plots equity curve with regime shading
- **Notes:** Depends on TASK-0029 (equity curve CSV output) for the first working notebook. The file contract in README.md is the formal boundary — Python never feeds back into Go inputs.

---

### [TASK-0084] Tooling — update evaluation-run agent to read bootstrap stats from JSON output

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-06
- **Source:** session
- **Context:** The evaluation-run pipeline agent (used in TASK-0069) parsed bootstrap distribution stats from stdout because the `--out` JSON did not contain them. TASK-0082 added bootstrap stats to the JSON output under a `"bootstrap"` key. The agent's stdout parsing is now redundant and fragile — it should be updated to read `bootstrap.sharpe_p5`, `bootstrap.prob_positive_sharpe`, etc. directly from the JSON file instead.
- **Acceptance criteria:**
  - [ ] Evaluation-run pipeline agent updated to read bootstrap stats from `--out` JSON (`bootstrap.sharpe_p5`, `bootstrap.sharpe_p50`, `bootstrap.sharpe_p95`, `bootstrap.prob_positive_sharpe`, `bootstrap.worst_drawdown_p95`, `bootstrap.n`, `bootstrap.seed`) instead of parsing stdout
  - [ ] Stdout parsing of bootstrap block removed from agent logic
  - [ ] Agent still works correctly when `bootstrap` key is absent (non-bootstrap runs)
- **Notes:** TASK-0082 is the prerequisite — it added the bootstrap fields to the JSON. The agent file to update is in `.claude/agents/` (evaluation-run agent). Low priority: stdout parsing still works; this is a fragility reduction.

---

### [TASK-0063] Tooling — update `cmd/backtest` package doc comment to list all 6 strategies

- **Status:** todo
- **Priority:** low
- **Created:** 2026-04-29
- **Source:** discovery
- **Context:** The package-level doc comment in `cmd/backtest/main.go` lists only `stub`, `sma-crossover`, and `rsi-mean-reversion` under "Available strategies". The `strategyRegistry` and `--strategy` flag help text now correctly list all 6, but the doc comment at the top of the file is stale and would mislead someone reading the source. Discovered during TASK-0051 quality review.
- **Acceptance criteria:**
  - [ ] `cmd/backtest/main.go` package doc comment "Available strategies" section updated to list all 6 strategies with their flag descriptions
  - [ ] `golangci-lint run ./cmd/backtest/...` still passes
- **Notes:** Pure documentation change — no logic, no tests needed. Low priority; do alongside any other `cmd/backtest` touch.

---

### [TASK-0083] Tech debt — handle `*ErrIncompleteData` typed error at cmd/ layer boundary

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-05
- **Source:** session
- **Context:** TASK-0081 introduced `*ErrIncompleteData` as a typed error from `FetchCandles` when chunk merge returns fewer candles than 90% of the weekday estimate. The cmd/ entrypoints (`cmd/universe-sweep`, `cmd/backtest`, `cmd/walk-forward`, `cmd/fetch-history`) currently propagate this as a generic `error` — no user-facing message distinguishes "no data" from "partial data". Callers should type-assert and print a clear diagnostic before exiting.
- **Acceptance criteria:**
  - [ ] `cmd/universe-sweep`, `cmd/backtest`, `cmd/walk-forward`: any `FetchCandles` error path type-asserts `*zerodha.ErrIncompleteData`; if matched, prints `incomplete data: instrument=%s from=%s to=%s expected≈%d got=%d` and exits with code 2 (distinct from generic error exit code 1)
  - [ ] `cmd/fetch-history` (TASK-0070): same typed-error handling wired in when that CLI is built
  - [ ] `golangci-lint run ./cmd/...` passes
  - [ ] Tests: mock provider returns `*ErrIncompleteData` → CLI prints correct diagnostic and exits with code 2
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Discovered during TASK-0081 harvest — the typed error is defined but not handled at the cmd/ boundary. Exit code 2 for incomplete data follows Unix convention (1 = generic error, 2 = misuse/data problem). `cmd/fetch-history` handling should be added as part of TASK-0070 build, not this task.

---

### [TASK-0088] Tech debt — cmd/monitor test cleanup: missing thresholds JSON test + Trade.Instrument in fixtures

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-07
- **Source:** discovery
- **Context:** Two cosmetic gaps found in cmd/monitor quality review (TASK-0048 post-build gate). Neither is a bug, but both are quick fixes to bring the test suite up to the same standard as the rest of the codebase.
- **Acceptance criteria:**
  - [ ] `TestRun_InvalidThresholdsJSON` added to `cmd/monitor/monitor_test.go`: write a thresholds file with invalid JSON, call `run()`, assert error returned
  - [ ] Three `model.Trade` struct literals in `TestBuildSyntheticCurve_Order` updated to include `Instrument: "NSE:TEST"` — satisfies repo rule "every Trade must carry an instrument identifier"
  - [ ] `go1.25.0 test -race ./cmd/monitor/...` and `golangci-lint run ./cmd/monitor/...` still pass after changes
- **Notes:** Discovered during go-quality-review standard gate on TASK-0048. Both fixes are in `cmd/monitor/monitor_test.go` only — no production code changes.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
