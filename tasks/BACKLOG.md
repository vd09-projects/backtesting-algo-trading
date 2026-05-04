# Project Task Backlog

**Last updated:** 2026-05-05 | **Open tasks:** 25 | **Next up:** TASK-0069

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

---

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0069] Evaluation — reconsider instrument-count gate threshold for MACD at 17/26/9

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** MACD crossover (fast=17, slow=26, signal=9) passed 9/14 instruments at walk-forward — 64% retention. The instrument-count gate requires 100% retention (same count as universe gate pass). Marcus ruled that this is a gate-design question, not a parameter question: the 9 passing instruments show solid OOS Sharpe (0.062–0.472 range), and the failures cluster in two structural patterns (OverfitFlag on large-cap defensives: RELIANCE, HINDUNILVR, WIPRO; NegFoldFlag on higher-vol names: TCS, HDFCBANK). The 100% retention requirement may be too strict for a useful portfolio strategy.
- **Acceptance criteria:**
  - [ ] Marcus (algo-trading-veteran) reviews MACD's 9/14 WF pass pattern and rules on whether 60–70% instrument retention is a defensible threshold for this strategy
  - [ ] If gate is relaxed: document new threshold in `decisions/algorithm/` with explicit rationale for why 9/14 constitutes sufficient cross-instrument evidence
  - [ ] If gate is relaxed and MACD advances: run bootstrap (TASK-0054 logic) for the 9 passing instrument pairs
  - [ ] Record decision with revisit trigger: if relaxed threshold allows strategies with fewer passing instruments, apply consistent standard to future strategies
  - [ ] If gate is NOT relaxed: record decision and mark MACD as killed permanently under current methodology
- **Notes:** MACD WF results: passes = SBIN, BAJFINANCE, TITAN, LT, ICICIBANK, INFY, AXISBANK, ITC, KOTAKBANK (9 instruments). Failures = TCS (NegFoldFlag), RELIANCE (OverfitFlag 0.48), HINDUNILVR (OverfitFlag 0.34), WIPRO (OverfitFlag 0.43), HDFCBANK (NegFoldFlag). Marcus standing order (2026-05-04): "the parameter is not the issue; the gate threshold is the question." Owner: Marcus (algo-trading-veteran). Unblocked 2026-05-04: TASK-0068 complete — SMA crossover killed at universe gate (zero sufficient instruments), no SMA survivors to affect gate-design precedent.

---

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
- **Notes:** Owner: Priya (dev). Blocked at runtime on Zerodha access token — no code blocker. Incremental delta fetch depends on TASK-0080 (CachedProvider manifest). Until TASK-0080 is complete, fetch-history CLI fetches full range from --from on every run (no incremental mode). First run for Nifty50 × 5-min × 2015→today ≈ 15 instruments × 30 chunks × 350ms = ~2.5 min.

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
- **Notes:** Owner: Priya (dev). The parameter-sweep mode must enforce DSR-corrected ranking — not raw Sharpe maximization — to avoid being a professional overfitting engine. Marcus's standing order: parameter search on training window only, DSR-corrected rank, OOS untouched during search.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

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
- **Notes:** Owner: Marcus (edge definition) → Priya (implementation). Depends on TASK-0071 (gap handling verified), TASK-0059 (walk-forward factory API), and TASK-0078 (session-boundary utilities) before implementation begins. Long-only vs bidirectional must be resolved by Marcus before build.

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

### [TASK-0054] Evaluation — Monte Carlo bootstrap on walk-forward survivors

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0053 (pipeline terminated — see Notes)
- **Context:** Bootstrap produces the confidence interval on the Sharpe estimate and the kill-switch Sharpe threshold. A strategy with a high point-estimate Sharpe but wide bootstrap distribution (low p5) has too much sampling variance to trust with real capital. The p5 Sharpe from this run becomes the live kill-switch threshold, not a round number.
- **Acceptance criteria:**
  - [ ] Run `cmd/backtest --bootstrap` for each surviving strategy × instrument pair from TASK-0053, 10,000 simulations
  - [ ] Apply bootstrap gate (from TASK-0049): SharpeP5 > 0 AND P(Sharpe > 0) > 80%
  - [ ] Record for each survivor: SharpeP5, SharpeP50, SharpeP95, ProbPositiveSharpe, WorstDrawdownP95
  - [ ] SharpeP5 value recorded as the kill-switch Sharpe threshold for that strategy × instrument pair — this feeds directly into TASK-0056
  - [ ] Kill strategies failing the bootstrap gate; record kill decision in `decisions/algorithm/`
  - [ ] Bootstrap seed logged with every result for reproducibility
- **Notes:** Bootstrap Sharpe is per-trade non-annualized per 2026-04-20 decision. Kill-switch comparison must use the identical formula. Owner: Marcus (algo-trading-veteran).

  PIPELINE TERMINATED: TASK-0053 produced 0 survivors. User chose Option B (parameter re-run) and Option A (gate-design review). Active remediation: TASK-0068 runs SMA at fast=20/slow=50 (new params, pre-committed revisit trigger). TASK-0069 escalates MACD instrument-count gate to Marcus. TASK-0054 unblocks if either remediation produces survivors. Kill records: `decisions/algorithm/2026-05-04-macd-crossover-walk-forward-instrument-count-gate.md`, `decisions/algorithm/2026-05-04-sma-crossover-walk-forward-instrument-count-gate.md`.

```json
{
  "survivor_input_from": "TASK-0053",
  "results_file": "runs/walk-forward-2026-05-04.csv",
  "survivors": []
}
```

---

### [TASK-0055] Evaluation — cross-strategy correlation and portfolio construction

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0054
- **Context:** Selects the final portfolio from bootstrap survivors. Two correlated strategies do not provide diversification — they add correlated risk, especially in the stress periods (2020, 2022) when diversification is most needed. The portfolio at ₹3 lakh targets ~10% annualized vol using vol-targeting sizing per strategy.
- **Acceptance criteria:**
  - [ ] Run `cmd/correlate` for all surviving strategy × instrument pairs from TASK-0054 (full equity curves)
  - [ ] Apply correlation gate (from TASK-0049): full-period r < 0.7 AND stress-period r < 0.6 for every pair in the final portfolio
  - [ ] If two strategies are correlated and from the same edge bucket, keep only the higher-DSR-Sharpe one
  - [ ] Select 2-4 uncorrelated survivors for the portfolio
  - [ ] Define capital allocation per strategy: combined portfolio targets ~10% annualized vol using `SizingVolatilityTarget`
  - [ ] Record portfolio composition and sizing rule in `decisions/algorithm/`
  - [ ] Record excluded strategies with reasons (correlation, gate failure, or sizing constraint)
- **Notes:** At ₹3 lakh total capital with vol-targeting, each strategy typically receives 20-50% of capital depending on realized volatility. No leverage at this stage. Owner: Marcus (algo-trading-veteran).

---

### [TASK-0056] Evaluation — pre-live brief: kill-switch thresholds and go/no-go sign-off

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** TASK-0055, TASK-0048 (cmd/monitor must exist for weekly monitoring cadence)
- **Context:** The final checkpoint before any live capital is allocated. Documents specific kill-switch thresholds, capital allocation, and the explicit go/no-go verdict for each portfolio strategy. No strategy goes live without this document existing and dated before the first trade. This is what the algo reviewer signs.
- **Acceptance criteria:**
  - [ ] For each portfolio strategy: kill-switch thresholds recorded — SharpeP5 threshold (from TASK-0054), MaxDrawdownPct (1.5× in-sample worst), MaxDDDuration (2× in-sample worst)
  - [ ] Thresholds written to `decisions/algorithm/YYYY-MM-DD-kill-switch-{strategy}.md` before first trade
  - [ ] Capital allocation per strategy documented in ₹ and % of ₹3 lakh total
  - [ ] Monitoring cadence documented: weekly kill-switch check via `cmd/monitor`
  - [ ] Explicit go/no-go verdict per strategy: APPROVED FOR LIVE or NOT APPROVED with specific reason
  - [ ] Algo reviewer acknowledgement noted in the brief
- **Notes:** Aggregates outputs from TASK-0049 through TASK-0055. Cannot be done in isolation. No strategy goes live without this document. The live period (2025 onward) is the true holdout — the walk-forward OOS windows are the proxy for out-of-sample evidence. Owner: Marcus (algo-trading-veteran).

---

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
- **Notes:** This is a significant engine change. Golden tests are mandatory for any event loop modification per repo standards. The `Session *SessionConfig` being optional (nil pointer) preserves all existing daily-bar tests without modification.

---

### [TASK-0048] Tooling — weekly kill-switch monitor (`cmd/monitor`)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-04-25
- **Source:** session
- **Blocked by:** Trade log file format decision required. The engine currently outputs `model.Trade` structs to JSON via `cmd/backtest --out`. A weekly monitor needs a file format for live trades that accumulates across sessions. Decision needed: reuse the existing JSON format, or define a separate append-only CSV schema for live trade records.
- **Context:** Marcus specified weekly kill-switch monitoring (not quarterly re-validation). The existing `analytics.CheckKillSwitch()` and `DeriveKillSwitchThresholds()` functions are ready. This task wires them into a runnable binary that reads live trade history and thresholds, outputs alert status, and can be cron-scheduled.
- **Acceptance criteria:**
  - [ ] Trade log file format decided and documented (decision record in `decisions/`)
  - [ ] `cmd/monitor/main.go` reads: (1) live trade log in the agreed format, (2) kill-switch thresholds JSON produced by `DeriveKillSwitchThresholds`
  - [ ] Calls `analytics.CheckKillSwitch(recentTrades, liveCurve, thresholds)`
  - [ ] Prints clear alert status: `OK`, `HALT (Sharpe breached)`, `HALT (drawdown breached)`, `HALT (duration breached)`
  - [ ] Exit code 0 if OK, non-zero if any threshold breached (enables shell scripting / cron alerting)
  - [ ] Tests: known trade sequence crossing each threshold type → correct alert output and exit code

---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

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
- **Notes:** This is a coordinated migration — do not migrate commission.go alone. Deferred from TASK-0038 per decision `2026-04-25-float64-for-commission-arithmetic`. The `shopspring/decimal` dependency must be discussed with the user before implementation per the no-new-dependencies rule in CLAUDE.md. Do not start until that discussion has happened.

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
- **Notes:** Donchian has only one meaningful sweep parameter (period) — its p2 axis is less obvious; defer the axis mapping decision until this task is picked up. The `fixedParams` duplication is a low-friction issue for now (2 files to update per new strategy) but compounds at 6 strategies.

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
- **Notes:** `StaticCSVProvider` should satisfy `provider.DataProvider` at compile time via a `var _ provider.DataProvider = (*StaticCSVProvider)(nil)` check. NSE CSV columns: Date, Open, High, Low, Close (or just Index Value for TRI — inspect the actual download first). TRI values will be in the 9,000–28,000 range for 2015–2024. No chunking, no auth, no rate limits needed. Follow-up to decision `2026-04-28-nifty-tri-benchmark-data-source.md`.

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
  - [ ] `golangci-lint run ./...` and `go1.25.0 test -race ./...` pass
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Small change — 3 files, ~20 lines total. Verify exact Kite API limits for 30-min and 60-min before setting chunk sizes.

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
- **Notes:** Depends on TASK-0029 (equity curve CSV output) for the first working notebook.
  The file contract in README.md is the formal boundary — Python never feeds back into Go inputs.

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

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
