# Project Task Backlog

**Last updated:** 2026-05-17 | **Open tasks:** 38 | **Next up:** TASK-0074

---

## In Progress

<!-- Currently being worked on. Keep at most 2-3 tasks here. -->

<!-- empty -->

## Up Next

<!-- Prioritized queue. The top item here is the answer to "what should I work on next?" -->

### [TASK-0074] Strategy — Opening Range Breakout (5-min, CNC overnight hold)

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** Intraday strategy for 5-min bars with 2-3 day CNC holds. Thesis: the first 30-60 minutes of the NSE session define price discovery; a clean breakout from that range in the first hour tends to persist intraday and sometimes into the next session. TimedExit wrapper provides the N-day time-stop for flat/sideways positions.
- **Acceptance criteria:**
  - [x] Marcus (algo-trading-veteran) rules on whether strategy is long-only or bidirectional — decision recorded in `decisions/algorithm/` before implementation begins
  - [x] Marcus defines: range window duration (30 / 45 / 60 min), breakout confirmation method (close above/below? volume threshold?), time-stop N (days), position sizing rule
  - [ ] `strategies/orb/` package implementing `Strategy` interface: range computed from first N 5-min bars using `IsSessionOpen()` from TASK-0078, long on close above high, exit on time-stop or target
  - [ ] Uses `pkg/strategy/timed_exit.go` wrapper for N-day time-stop and `pkg/strategy/price_exit.go` for SL/TP
  - [ ] CLI registered in all strategy registries (`cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`)
  - [ ] All public functions tested; golden test for range computation and signal generation
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (implementation). Marcus rules written 2026-05-13 to `decisions/algorithm/2026-05-13-orb-marcus-rules.md`. Strategy is ready for Priya to implement. Unblocked. All infra complete: TASK-0059 (WF factory API), TASK-0071 (gap handling), TASK-0078 (session helpers), TASK-0098 (PriceExit), TASK-0099/0101 (5-min cache 15/15 large-cap). Concrete rules: range = first 6 bars (09:15-09:44 IST); entry on first close > RangeHigh × 1.001 after bar 6, no-entry cutoff 11:30 IST; SL = 1.5× range width below entry; TP = 2× range width above entry; TimedExit = 225 bars (3 sessions). Long-only, no pyramid. CommissionZerodhaFull (CNC). Sweep axes: range window (6/9/12 bars), buffer (0.05/0.10/0.20%), SL multiplier (1.0/1.5/2.0×), hold bars (150/225/300). TASK-0113 (signal audit) blocked on this task.

---

### [TASK-0075] Strategy — Gap-and-Go (5-min, CNC overnight hold)

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-04
- **Source:** session
- **Context:** Intraday strategy for 5-min bars with 1-2 day CNC holds. Thesis: NSE large/midcap stocks opening 1-2%+ above/below prior close on above-average volume tend to continue in the gap direction for 1-2 sessions before reversion. Captures institutional order flow from overnight news. TimedExit and PriceExit wrappers provide the time-stop and SL/TP.
- **Acceptance criteria:**
  - [x] Marcus defines: gap threshold % (1.0%), volume threshold (1.3× 20-day average), entry bar (close of 2nd bar, 09:20 IST), SL (0.8× gapPct), TP (2.0× gapPct), time-stop (150 bars = 2 sessions), long-only initially
  - [ ] `strategies/gapandgo/` package implementing `Strategy` interface: gap detection via `PreviousSessionClose`, entry on 2nd bar close (09:20 IST), no-entry if gap already 1.5× chased by bar 1
  - [ ] Uses `pkg/strategy/price_exit.go` (SL/TP) and `pkg/strategy/timed_exit.go` (150-bar time-stop) wrappers; composition: `NewTimedExit(NewPriceExit(inner, stopLossPct, targetProfitPct), 150)`
  - [ ] SL and TP percentages computed dynamically at entry from actual gapPct (not fixed %)
  - [ ] Volume threshold: 20-day average volume tracking via pre-allocated ring buffer (no hot-loop allocation)
  - [ ] CLI registered in all strategy registries (`cmd/backtest`, `cmd/universe-sweep`, `cmd/walk-forward`); name: `"gap-and-go"`
  - [ ] All public functions tested; golden test covering: gap-up enter, no-gap skip, volume-below-threshold skip, gap-already-chased skip, SL exit, TP exit, time-stop exit
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (implementation). Marcus rules written 2026-05-13 to `decisions/algorithm/2026-05-13-gap-and-go-marcus-rules.md`. Unblocked. All infra complete: TASK-0059 (WF factory API), TASK-0071 (gap handling), TASK-0078 (session helpers — `PreviousSessionClose`), TASK-0098 (PriceExit), TASK-0099/0101 (5-min cache 48/48 midcap + 15/15 large-cap). Concrete rules: gap threshold=1.0%; volume=1.3× 20-day avg; entry=close of bar index 1 (09:20 IST); no-entry if abs(bars[1].Close-PrevClose)/PrevClose >= 1.5×gapPct; SL=0.8×gapPct; TP=2.0×gapPct; time-stop=150 bars (2 sessions); long-only, no pyramid; CommissionZerodhaFull (CNC). Sweep axes: gap [0.75,1.0,1.5%], volume [1.0,1.3,1.5×], SL multiplier [0.5,0.8,1.0×], hold bars [75,150,225]. Orientation instrument: NSE:INDHOTEL. TASK-0119 (signal audit) and TASK-0120 (param sensitivity) blocked on this task.

---

### [TASK-0126] Research spike — 5-min strategy canvas (broad, 10+ candidates)

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-17
- **Source:** session
- **Context:** Cast a wide net across strategy types for 5-min NSE trading before committing engineering effort. The current portfolio (MACD, SMA, RSI, Bollinger, CCI, Momentum — all daily-bar) shares the same edge bucket. We need candidates from structurally different buckets: structural session-boundary effects, volume-based, volatility-regime, multi-timeframe, relative strength (stock vs index), and statistical (pairs). The output is a prioritized candidate list, not code.
- **Acceptance criteria:**
  - [ ] Written canvas covering at least 10 candidate strategies, each with: (1) one-sentence edge thesis, (2) behavioral mechanism (who is on the other side), (3) data requirements on current NSE 5-min Kite infrastructure, (4) feasibility verdict (yes/needs-engine-change/no), (5) Marcus pre-screen go/evaluate/skip
  - [ ] Candidates span at least 4 distinct edge buckets: structural, volume-based, volatility-regime, relative-strength/multi-timeframe, statistical-arbitrage (scope-only for pairs)
  - [ ] Candidate list ranked by: feasibility on current infra first, edge strength second
  - [ ] Output recorded as `decisions/algorithm/2026-05-17-5min-strategy-canvas.md`
  - [ ] Top 2-3 candidates each get a Marcus evaluation session ticket created (source: decision)
- **Notes:** Owner: Marcus (research) → task-manager (create evaluation tickets). Do NOT implement anything from this research — the output feeds evaluation sessions only. Existing strategy adaptations (MACD/RSI/SMA on 5-min) are covered separately in TASK-0127 and are NOT part of this canvas.

---

### [TASK-0127] Strategy — adapt existing daily-bar strategies to 5-min (MACD, RSI, SMA, Bollinger, CCI, Momentum)

- **Status:** todo
- **Priority:** high
- **Created:** 2026-05-17
- **Source:** session
- **Context:** Six existing strategies built for daily bars should be evaluated at 5-min resolution. Key challenges: (1) parameters need recalibration — MACD(12,26,9) on 5-min tracks 60-min and 130-min trends, not 2-week/5-week; (2) session-boundary behavior — most indicators tolerate cross-session computation for trend-following, but mean-reversion strategies may need session-reset logic; (3) signal frequency on 5-min is ~75× higher than daily, so per-trade P&L must cover commission at this frequency. Each strategy gets a signal audit before any walk-forward work.
- **Acceptance criteria:**
  - [ ] For each strategy (MACD, RSI mean-reversion, SMA crossover, Bollinger mean-reversion, CCI mean-reversion, Momentum): recalibrated parameter set proposed with rationale (what market duration is the strategy targeting at 5-min?)
  - [ ] Session-boundary decision recorded per strategy: cross-session indicator computation accepted or session-reset required — rationale in one sentence
  - [ ] Signal audit run on each adapted strategy across NSE midcap/large-cap 5-min universe; per-instrument trade counts recorded; strategies with <20 trades/year on >80% of instruments flagged for kill
  - [ ] For strategies passing signal audit: universe gate run (`cmd/universe-sweep`) with 5-min data; results in `runs/`
  - [ ] Marcus analysis: which adapted strategies survive signal audit and deserve full evaluation pipeline? Record in `decisions/algorithm/`
  - [ ] Tests written before any implementation changes (TDD)
- **Notes:** Owner: Marcus (parameter recalibration decision) → Priya (implementation changes) → Marcus (analysis). This is not a new strategy package — it's adapting existing `strategies/` packages to accept 5-min timeframe. The GlobalRegistry already registers these strategies; the engine is timeframe-agnostic. Primary risk: commission drag at 5-min frequency. Zerodha CNC commission is ₹20 flat — at average NSE midcap price of ~₹800, that's 2.5% round-trip on a 1-share position. Position sizing must account for this. TASK-0128 (composite signals) depends on which strategies survive this task. **Multiple parameter variants:** For each strategy, create 2–3 named variants with different parameter sets (e.g., `macd-5min-fast` 9/21/9, `macd-5min-standard` 12/26/9, `macd-5min-slow` 26/52/18) rather than forcing a single calibration choice. Register each variant in `GlobalRegistry`. The evaluation pipeline runs all variants independently — survivors self-select. **Session-boundary decision:** For trend-following strategies (MACD, SMA, Momentum), cross-session indicator computation is acceptable — trends persist across sessions. For intraday mean-reversion strategies (RSI, Bollinger, CCI), document explicitly whether indicator should reset at session open or compute continuously — decision matters for signal quality.

---

### [TASK-0128] Research — composite signal design for 5-min (regime filters, not signal stacking)

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-17
- **Source:** session
- **Context:** Three composite combinations worth evaluating, each structured as a regime filter on an existing strategy (not oscillator conjunction). (1) MACD crossover + price above VWAP: long signal only when price > VWAP at signal time — VWAP is the institutional benchmark, aligning with institutional flow. (2) RSI oversold (<30) + volume >= 1.5× session average: volume confirms exhaustion selling vs. slow bleed. (3) SMA golden cross + first 120 min of session (09:15–11:15 IST): morning crossovers have more follow-through on NSE due to institutional participation peak. Each combination needs Marcus evaluation before implementation.
- **Acceptance criteria:**
  - [ ] Marcus evaluation session run for each combination (3 evaluations); verdict recorded in `decisions/algorithm/`
  - [ ] VWAP computation added as a utility in `pkg/strategy/` if MACD+VWAP combination is approved (VWAP = cumulative (price×volume) / cumulative volume, session-reset at session open)
  - [ ] For each approved combination: signal audit on 5-min universe, result in `runs/`; strategies with fewer total trades than the base strategy by >60% flagged (filter too aggressive)
  - [ ] Implementation only after Marcus go verdict + signal audit pass
  - [ ] Tests written before implementation (TDD)
- **Notes:** Blocked on TASK-0127 AND TASK-0131 (VWAP utility needed for MACD+VWAP combination). Need to know which base strategies survive 5-min adaptation before designing filters on them. The combinations listed are concrete candidates but Marcus may add or remove at evaluation. Key constraint: adding a filter must reduce false positives (improve precision), not just reduce trade count. A filter that cuts trades by 60% while improving Sharpe by 20% is marginal — the confidence interval widens. Composite signal design tickets (one per combination) should be created after TASK-0127 signals which base strategies survive — don't design filters for strategies that failed.

---

### [TASK-0129] Research spike — verify 1-min historical data depth on Kite Connect

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-17
- **Source:** session
- **Context:** The Kite Connect API specifies 60-day windows per call for 1-min bars. Whether the underlying data store goes back further (like 5-min which empirically has 5+ years) is unverified. If 1-min data extends to 2022 or earlier, the chunking infrastructure already handles multi-call pagination and 1-min becomes viable for backtesting. If total available history is only 60 days, 1-min is not useful for strategy evaluation and should be dropped as a focus area.
- **Acceptance criteria:**
  - [ ] Test fetch attempted on NSE:RELIANCE 1-min from 2022-01-01 to 2022-03-01 via `cmd/fetch-history --timeframe 1min --from 2022-01-01 --to 2022-03-01`
  - [ ] Result recorded: actual available start date, bar count returned, any API errors
  - [ ] If data available from 2022: estimate total history depth; update `maxDaysPerInterval` in `pkg/provider/zerodha/chunk.go` if needed; create TASK for adding `Timeframe1Min` to the model and provider
  - [ ] If data not available past 60 days: record as decision, drop 1-min from strategy focus
  - [ ] Result documented in `decisions/infrastructure/2026-05-17-1min-kite-historical-depth.md`
- **Notes:** Quick empirical test — no code changes expected unless 1-min is viable. Requires valid Kite access token. The existing 5-min depth test (TASK-0099) proved Kite stores much more than the per-call window — same hypothesis for 1-min. Owner: anyone with a valid access token.

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

### [TASK-0109] Tech debt — `cmd/fetch-history`: wire `HandleIncompleteDataError` for `*ErrIncompleteData` in `fetchOne`

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-13
- **Source:** session
- **Context:** TASK-0083 added `cmdutil.HandleIncompleteDataError` and wired it into `cmd/backtest`, `cmd/walk-forward`, and `internal/universesweep`. `cmd/fetch-history` (TASK-0070, done 2026-05-09) was explicitly excluded from TASK-0083 scope. TASK-0070 notes state "*ErrIncompleteData handling applies to this CLI but is tracked separately." This task closes that gap: `fetchOne` should detect `*ErrIncompleteData`, log the per-instrument diagnostic to stderr, and treat it as a fetch failure (not a hard process exit) — consistent with how `internal/universesweep` handles it for universe-sweep.
- **Acceptance criteria:**
  - [ ] `fetchOne` in `cmd/fetch-history/main.go`: after `FetchCandles` returns an error, call `cmdutil.HandleIncompleteDataError(err, stderr)`; if non-nil, log the diagnostic (already printed by the helper) and return the error — treated as a per-instrument fetch failure, partial-failure manifest path applies
  - [ ] Existing partial-failure behavior preserved: other instruments continue fetching; `fetch-progress.json` updated for successful instruments only
  - [ ] `TestFetchOne_IncompleteData` added: mock provider returns `*ErrIncompleteData` for one instrument; assert (a) stderr contains the `incomplete data:` diagnostic, (b) error is returned, (c) successful instruments still fetched
  - [ ] `go1.25.0 test -race ./cmd/fetch-history/...` passes
  - [ ] `golangci-lint run ./cmd/fetch-history/...` passes
- **Notes:** Owner: Priya (dev). `cmd/fetch-history` is a bulk fetcher — unlike `cmd/backtest` (single instrument, hard exit 2) or `cmd/walk-forward` (single instrument, hard exit 2), fetch-history accumulates errors and continues. Per-instrument failure (not process exit) is the correct semantic, matching `internal/universesweep.runInstrument`. The `HandleIncompleteDataError` helper prints the diagnostic — `fetchOne` just needs to treat the returned `*ExitCodeError{Code:2}` as a regular per-instrument error.

---

## Blocked

<!-- Waiting on something. Each task must state what it's blocked by. -->

### [TASK-0113] Evaluation — signal frequency audit — ORB on 15 large-cap instruments

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0074 (ORB implementation must complete first)
- **Context:** Run ORB signal audit across 15 NSE large-cap instruments (5-min bars, 2018-2023). Gate: >= 30 trades per instrument on >= 40% of the 15-instrument universe. Any instrument with < 30 trades is EXCLUDED from further analysis. If fewer than 30 trades across ALL 15 instruments combined, kill before full backtest. Expected: 40-80 trades/year per instrument (240-480 over 6 years) — should clear comfortably.
- **Acceptance criteria:**
  - [ ] Signal audit run across all 15 large-cap instruments using the plateau-default parameters from Marcus rules (range window=6 bars, buffer=0.10%, SL=1.5×, hold=225 bars)
  - [ ] Per-instrument trade count recorded; instruments with < 30 trades marked EXCLUDED
  - [ ] Gate evaluated: >= 30 trades on >= 6/15 instruments (40%); kill if not met
  - [ ] Results recorded in `runs/` signal audit CSV consistent with existing audit format
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13-evaluate-orb.json, Marcus GO verdict. TASK-0114 (sensitivity analysis) blocked on this.

---

### [TASK-0114] Evaluation — in-sample baseline and parameter sensitivity — ORB on RELIANCE

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0074 (ORB implementation must complete first)
- **Context:** Orientation run + parameter sweep on RELIANCE 2018-2023. Sweep axes: range window (6/9/12 bars), buffer (0.05%/0.10%/0.20%), SL multiplier (1.0×/1.5×/2.0× range width), hold bars (150/225/300). Identify plateau range within 80% of peak Sharpe in the valid region (>= 30 trades). Select plateau-midpoint parameter set for universe sweep. Donchian failed DSR-corrected Sharpe at universe gate — orientation run will tell us if the same pattern is emerging.
- **Acceptance criteria:**
  - [ ] Orientation backtest on RELIANCE with default parameters (range=6 bars, buffer=0.10%, SL=1.5×, hold=225 bars) — baseline Sharpe recorded
  - [ ] Parameter sweep completed across 4 axes; heatmap/CSV output in `runs/`
  - [ ] Valid region (>= 30 trades) identified; plateau (80% of peak Sharpe) identified within valid region
  - [ ] Plateau-midpoint parameter set selected and documented; if no valid plateau, fallback-to-defaults procedure applies per `decisions/algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md`
  - [ ] Sensitivity concern flag recorded if valid region is narrow (< 20% of grid)
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0115 (universe sweep) blocked on this. Plateau selection per accepted procedure in `decisions/algorithm/2026-04-29-plateau-procedure-trade-count-constrained.md`.

---

### [TASK-0115] Evaluation — universe sweep — ORB across Nifty50 large-cap (15 instruments)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0114 (sensitivity analysis and plateau-midpoint parameter selection)
- **Context:** Run ORB across all 15 large-cap instruments using plateau-midpoint parameter from sensitivity analysis. Apply universe gate: DSR-corrected avg Sharpe > 0 AND >= 40% instruments positive with >= 30 trades (nTrials=15). Correlation risk flag: if ORB survivors concentrate on SBIN and TITAN (current MACD large-cap survivors), flag for correlation gate — stress-period r may exceed 0.6. Donchian failed this gate; monitor DSR correction behaviour closely.
- **Acceptance criteria:**
  - [ ] Universe sweep run across all 15 large-cap instruments with plateau-midpoint parameters
  - [ ] DSR-corrected average Sharpe computed (nTrials=15); gate: DSR avg > 0
  - [ ] Pass fraction computed; gate: >= 40% instruments (>= 6/15) with positive Sharpe and >= 30 trades
  - [ ] Instrument-level results in `runs/` CSV consistent with existing sweep format
  - [ ] Correlation risk flag: if SBIN and/or TITAN are among survivors, note in results for correlation gate awareness
  - [ ] Universe gate kill recorded in `decisions/algorithm/` if gate fails
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0116 (walk-forward) blocked on this. Universe gate is the gate that killed Donchian, CCI, Bollinger, and Momentum — DSR correction for 15 instruments is less punishing than 48 (midcap), but ORB must still clear.

---

### [TASK-0116] Evaluation — walk-forward validation — ORB

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0115 (universe sweep and universe gate must pass first)
- **Context:** 2yr IS / 1yr OOS / 1yr step on universe-gate survivors, 2018-2024. Gate: OverfitFlag = false (AvgOOSSharpe >= 50% of AvgISSharpe) AND NegativeFoldFlag = false. THIS IS THE HIGHEST-RISK GATE for ORB per Marcus evaluation: IS-to-OOS Sharpe degradation is expected in choppy regimes. The 2018-2024 window includes COVID (2020) and rate shock (2022) — both are wide-range, false-breakout environments that will stress the OOS folds.
- **Acceptance criteria:**
  - [ ] Walk-forward run on all universe-gate survivors with 2yr IS / 1yr OOS / 1yr step
  - [ ] OverfitFlag evaluated per instrument: false if AvgOOSSharpe >= 50% AvgISSharpe
  - [ ] NegativeFoldFlag evaluated per instrument: false if no OOS fold has negative Sharpe
  - [ ] Walk-forward instrument-count gate applied: >= 60% of universe-gate survivors must pass both flags
  - [ ] Results in `runs/` WF CSV consistent with existing walk-forward format
  - [ ] Walk-forward gate kill recorded in `decisions/algorithm/` if gate fails
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0117 (bootstrap) blocked on this. Walk-forward methodology per `decisions/algorithm/2026-04-22-walk-forward-oos-is-sharpe-threshold.md` and `decisions/algorithm/2026-04-22-walk-forward-window-sizing-default.md`.

---

### [TASK-0117] Evaluation — Monte Carlo bootstrap — ORB

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0116 (walk-forward gate must pass first)
- **Context:** 10,000 simulations on walk-forward survivors. Gate: SharpeP5 > 0 AND P(Sharpe > 0) > 80% (per-trade non-annualized Sharpe, sample variance, no annualization). SharpeP5 becomes the live kill-switch rolling Sharpe threshold per accepted kill-switch derivation methodology.
- **Acceptance criteria:**
  - [ ] Bootstrap run (10,000 simulations) on each walk-forward survivor using `internal/montecarlo.Bootstrap`
  - [ ] Gate evaluated per instrument: SharpeP5 > 0 AND P(Sharpe > 0) > 80%
  - [ ] SharpeP5 recorded per surviving instrument — this value is the kill-switch Sharpe threshold
  - [ ] Bootstrap results in `runs/` JSON consistent with existing bootstrap format
  - [ ] Bootstrap gate kill recorded in `decisions/algorithm/` if gate fails
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0118 (pre-live brief) blocked on this. Per-trade Sharpe formula: `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance (n-1), no annualization — must match `CheckKillSwitch` formula exactly per accepted kill-switch methodology.

---

### [TASK-0118] Evaluation — pre-live brief — ORB kill-switch thresholds and go/no-go sign-off

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0117 (bootstrap must complete first)
- **Context:** Final checkpoint before live deployment. Document kill-switch thresholds (SharpeP5 from bootstrap, MaxDrawdownPct 1.5× in-sample worst, MaxDDDuration 2× in-sample worst), capital allocation, and explicit APPROVED/NOT APPROVED verdict per surviving instrument. Per accepted kill-switch derivation methodology — thresholds must be committed in writing before any live deployment, not derived after observing live results.
- **Acceptance criteria:**
  - [ ] Kill-switch thresholds documented per surviving instrument: SharpeP5 (from TASK-0117), MaxDD threshold (1.5× in-sample MaxDD), MaxDDDuration threshold (2× in-sample MaxDDDuration)
  - [ ] Capital allocation per instrument specified (vol-target 10% annualized, Rs 3 lakh base)
  - [ ] Correlation check against MACD survivors (SBIN, TITAN large-cap; PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN midcap): full-period r < 0.7 AND stress-period r < 0.6
  - [ ] Explicit APPROVED/NOT APPROVED verdict per instrument recorded in `decisions/algorithm/`
  - [ ] Decision file written to `decisions/algorithm/YYYY-MM-DD-orb-prelive-brief.md`
- **Notes:** Part of ORB evaluation pipeline. Source: evaluation session 2026-05-13. Per `decisions/algorithm/2026-04-21-kill-switch-derivation-methodology.md` — standing order, must be followed. Correlation gate thresholds per `decisions/algorithm/2026-04-27-correlation-gate.md`.

---


---

### [TASK-0119] Evaluation — signal frequency audit — Gap-and-Go on 48 midcap instruments

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0075 (Gap-and-Go implementation must complete first)
- **Context:** Audit trade count per instrument across 48 Nifty Midcap liquid instruments (5-min bars, 2021-2023 window). Default parameters: gapThreshold=1.0%, volumeMultiplier=1.3×. Expected 50-75 trades per instrument over 2.5 years at 1.0% threshold — marginal but should clear the 30-trade floor. If >20% of instruments excluded, re-run at 0.75% threshold before declaring a kill (per Marcus decision 2026-05.1.2 in `decisions/algorithm/2026-05-13-gap-and-go-marcus-rules.md`).
- **Acceptance criteria:**
  - [ ] `cmd/signal-audit` run with `gap-and-go` strategy on `universes/nifty-midcap-liquid.yaml` (2021-2023 window)
  - [ ] Signal audit CSV reviewed; instruments with < 30 trades marked EXCLUDED
  - [ ] If > 20% instruments excluded at 1.0% threshold: re-run at 0.75% gap threshold before kill decision
  - [ ] Decision recorded in `decisions/algorithm/` with per-instrument trade counts and EXCLUDED list
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13-evaluate-gap-and-go.json, Marcus GO verdict. TASK-0120 (param sensitivity) and TASK-0121 (universe sweep) depend on the threshold surviving this gate. Signal audit risk is higher than ORB — monitor closely.

---

### [TASK-0120] Evaluation — in-sample baseline and parameter sensitivity — Gap-and-Go on NSE:INDHOTEL

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0075 (Gap-and-Go implementation must complete first)
- **Context:** Orientation run + parameter sweep on NSE:INDHOTEL 2021-2023 (5-min bars). Four sweep axes: gap threshold [0.75%, 1.0%, 1.5%], volume multiplier [1.0×, 1.3×, 1.5×], SL multiplier [0.5×, 0.8×, 1.0× gap], hold bars [75, 150, 225]. Apply plateau procedure (80% Sharpe floor within valid trade-count region). Select plateau-midpoint parameter for universe sweep. NSE:INDHOTEL chosen as orientation instrument per Marcus — midcap profile, bootstrap MACD survivor, confirmed 5-min data.
- **Acceptance criteria:**
  - [ ] `cmd/param-search` run on NSE:INDHOTEL with all 4 sweep axes (81 variants = 3×3×3×3)
  - [ ] Plateau-midpoint parameter identified within valid region (>= 30 trades) using 80% Sharpe floor
  - [ ] Plateau-midpoint and sweep CSV recorded in `runs/`
  - [ ] Sensitivity concern documented if valid region is empty or all-negative; fallback to defaults per `decisions/algorithm/2026-04-29-fallback-to-defaults-no-valid-plateau.md`
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0121 (universe sweep) blocked on this task's plateau-midpoint output.

---

### [TASK-0121] Evaluation — universe sweep — Gap-and-Go across Nifty Midcap 150 (48 instruments)

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0120 (parameter sensitivity and plateau-midpoint selection must complete first)
- **Context:** Run Gap-and-Go across all 48 midcap instruments using plateau-midpoint parameters from TASK-0120 (or defaults if no valid plateau). Apply universe gate: DSR-corrected avg Sharpe > 0 (nTrials=48) AND >= 40% of sufficient instruments (>= 30 trades) show positive Sharpe. The DSR correction for 48 instruments is punishing — marginal per-instrument Sharpe will compress to near zero or negative.
- **Acceptance criteria:**
  - [ ] `cmd/universe-sweep` run across `universes/nifty-midcap-liquid.yaml` with plateau-midpoint parameters (or defaults)
  - [ ] DSR-corrected average Sharpe computed (nTrials=48); gate: DSR avg > 0
  - [ ] Pass fraction computed; gate: >= 40% instruments (>= 20/48 sufficient) with positive Sharpe
  - [ ] Universe gate kill recorded in `decisions/algorithm/` if gate fails
  - [ ] Advance decision recorded with surviving instrument list if gate passes
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0122 (walk-forward) blocked on this. Universe gate nTrials=48 per `decisions/algorithm/2026-05-09-ntrials-48-for-midcap-dsr-correction.md`.

---

### [TASK-0122] Evaluation — walk-forward validation — Gap-and-Go across midcap universe gate survivors

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0121 (universe sweep and universe gate must pass first)
- **Context:** 2yr IS / 6mo OOS / anchored walk-forward on instruments that passed the universe gate. Gate: OverfitFlag = false AND NegativeFoldFlag = false. Instrument retention: >= floor(0.60 × universe_gate_passes), minimum 6 instruments. Walk-forward fold structure: 2yr IS, 6mo OOS, anchored — per `decisions/algorithm/2026-05-10-5min-data-coverage-both-universes.md` recommendation. 2022 choppy regime is the primary stress period; watch for OverfitFlag clustering on Consumer/FMCG names.
- **Acceptance criteria:**
  - [ ] `cmd/evaluate` (walk-forward stage) run on universe gate survivors with 2yr IS / 6mo OOS / anchored structure
  - [ ] Instrument retention computed: >= floor(0.60 × universe_gate_passes) AND >= 6 instruments
  - [ ] OverfitFlag and NegativeFoldFlag patterns documented per instrument; clustering by sector noted
  - [ ] Walk-forward gate kill recorded in `decisions/algorithm/` if retention fails
  - [ ] Advance decision with WF-surviving instrument list if gate passes
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0123 (bootstrap) blocked on this. Walk-forward is the primary structural risk per Marcus — 2022 OOS period is the critical stress window.

---

### [TASK-0123] Evaluation — Monte Carlo bootstrap — Gap-and-Go walk-forward survivors

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0122 (walk-forward gate must pass first)
- **Context:** 10,000 Monte Carlo simulations on walk-forward survivors. Gate: SharpeP5 > 0 AND P(Sharpe > 0) > 80% (per-trade non-annualized Sharpe, sample variance, no annualization). SharpeP5 becomes the live kill-switch rolling Sharpe threshold per `decisions/algorithm/2026-04-21-kill-switch-derivation-methodology.md`.
- **Acceptance criteria:**
  - [ ] `cmd/evaluate` (bootstrap stage) run on walk-forward survivors using `internal/montecarlo.Bootstrap` (10,000 sims)
  - [ ] Gate evaluated per instrument: SharpeP5 > 0 AND P(Sharpe > 0) > 80%
  - [ ] SharpeP5 recorded per surviving instrument — becomes kill-switch Sharpe threshold
  - [ ] Bootstrap results in `runs/` JSON consistent with existing bootstrap format
  - [ ] Kill or advance decision recorded in `decisions/algorithm/` per instrument
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0124 (correlation gate) blocked on this. Per-trade Sharpe formula: `mean(ReturnOnNotional) / std(ReturnOnNotional)`, sample variance (n-1), no annualization.

---

### [TASK-0124] Evaluation — correlation gate — Gap-and-Go bootstrap survivors vs MACD midcap and ORB

- **Status:** blocked
- **Priority:** high
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0123 (bootstrap gate must pass first)
- **Context:** Pairwise correlation check between Gap-and-Go equity curve and: (1) MACD midcap portfolio survivors (PERSISTENT, TORNTPHARM, COFORGE, SUNDARMFIN, INDHOTEL, MUTHOOTFIN); (2) ORB if it has reached bootstrap by this stage. Gate: full-period Pearson r < 0.7 AND stress-period r < 0.6 in both COVID crash (2020-02-01 to 2020-06-30) and rate correction (2022-01-01 to 2022-12-31). HIGHEST-RISK PAIR: ORB and Gap-and-Go are both 5-min CNC morning-event long-only strategies — their mutual correlation in crash regimes may exceed 0.6. Tiebreaker: retain higher DSR per `decisions/algorithm/2026-04-27-correlation-gate.md`.
- **Acceptance criteria:**
  - [ ] Pairwise Pearson r computed: Gap-and-Go vs MACD midcap portfolio (full-period + 2 stress windows)
  - [ ] If ORB has reached this stage: Gap-and-Go vs ORB correlation computed (full-period + 2 stress windows)
  - [ ] Correlation gate applied: full-period r < 0.7 AND stress-period r < 0.6
  - [ ] If pair fails: tiebreaker applied — retain higher DSR; kill decision recorded for the dropped strategy
  - [ ] Pass or kill decision recorded in `decisions/algorithm/` per pair
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. TASK-0125 (pre-live brief) blocked on this. ORB/Gap-and-Go mutual correlation is the secondary pipeline risk — both are long-only morning-event strategies.

---

### [TASK-0125] Evaluation — pre-live brief — Gap-and-Go kill-switch thresholds and go/no-go sign-off

- **Status:** blocked
- **Priority:** medium
- **Created:** 2026-05-13
- **Source:** decision
- **Blocked by:** TASK-0124 (correlation gate must pass first)
- **Context:** Final checkpoint before any live deployment. Document kill-switch thresholds (SharpeP5 from TASK-0123, MaxDrawdownPct = 1.5× in-sample worst, MaxDDDuration = 2× in-sample worst), early-warning flag (3 consecutive losses on same instrument triggers manual review), capital allocation (vol-target 10%, hard cap Rs 1.5L per position), and explicit APPROVED/NOT APPROVED verdict per surviving instrument. Per `decisions/algorithm/2026-04-21-kill-switch-derivation-methodology.md` — must be written before any live deployment.
- **Acceptance criteria:**
  - [ ] Kill-switch thresholds computed and recorded in `decisions/algorithm/` per instrument: SharpeP5, MaxDD threshold (1.5×), MaxDDDuration threshold (2×)
  - [ ] Early-warning rule documented: 3 consecutive losses on same instrument triggers manual review (not automatic halt)
  - [ ] Capital allocation confirmed: vol-target 10% annualized, hard cap Rs 1.5 lakh per position (50% of Rs 3L capital)
  - [ ] Explicit APPROVED/NOT APPROVED verdict recorded per instrument in `decisions/algorithm/YYYY-MM-DD-gap-and-go-prelive-brief.md`
- **Notes:** Part of Gap-and-Go evaluation pipeline. Source: evaluation session 2026-05-13. Standing order: thresholds must be written before live deployment, not derived after observing live results.

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
- **Notes:** Significant engine change. Golden tests mandatory for any event loop modification. `Session *SessionConfig` being optional (nil pointer) preserves all existing daily-bar tests without modification.

---


---

## Todo (Backlog)

<!-- Lower-priority items. Ordered by priority within this section. -->

<!-- === QUICK WINS (1–5 line changes) === -->

### [TASK-0103] Fix — `cmd/evaluate/makeStageDir`: use `os.MkdirAll` to handle retry-after-error

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-10
- **Source:** discovery
- **Context:** `makeStageDir` in `cmd/evaluate/main.go` calls `os.Mkdir` (not `os.MkdirAll`). If the pipeline fails after the stage directory is created but before `verdict.json` is written (e.g., provider failure), a re-run on the same day will fail with "directory already exists" because `makeStageDir` creates an identical dated path. The user must manually delete the empty directory before retrying.
- **Acceptance criteria:**
  - [ ] `os.Mkdir` in `makeStageDir` replaced with `os.MkdirAll` — silently succeeds if the directory already exists
  - [ ] `TestMakeStageDir_IdempotentOnRetry` added: call `makeStageDir` twice with the same args, assert second call succeeds (no error)
  - [ ] `go1.25.0 test -race ./cmd/evaluate/...` passes
  - [ ] `golangci-lint run ./cmd/evaluate/...` passes
- **Notes:** Discovered during TASK-0073 multi-perspective review (Error Handling & Resilience Inspector). One-line fix. The `--out-dir` must still exist (it's the user's root directory); only the dated subdirectory is created with MkdirAll.

---

### [TASK-0104] Refactor — `cmd/evaluate/applyWFGate`: use `gateResult.PositiveSharpeInstruments` as the gate input

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-10
- **Source:** discovery
- **Context:** In `runPipeline`, `applyWFGate(len(universeSurvivors), ...)` computes the WF floor from `len(universeSurvivors)` — the instruments with positive Sharpe and sufficient data. The decision document (`decisions/algorithm/2026-05-05-walk-forward-instrument-count-gate-relaxed.md`) defines the gate as `WF_passes >= floor(0.60 × universe_gate_passes)` where `universe_gate_passes` is the authoritative count from `universesweep.GateResult.PositiveSharpeInstruments`. The two values are numerically identical, but using the canonical field makes the code's connection to the decision document direct and self-documenting.
- **Acceptance criteria:**
  - [ ] In `runPipeline` (line ~297): `applyWFGate(len(universeSurvivors), ...)` → `applyWFGate(gateResult.PositiveSharpeInstruments, ...)`
  - [ ] Add inline comment: `// gateResult.PositiveSharpeInstruments is the authoritative universe_gate_passes per decision 2026-05-05`
  - [ ] All existing tests pass: `go1.25.0 test -race ./cmd/evaluate/...`
  - [ ] `golangci-lint run ./cmd/evaluate/...` passes
- **Notes:** Discovered during TASK-0073 multi-perspective review (Domain Logic Reviewer). One-line change. The numerical outcome is unchanged; this is pure clarity/traceability. `len(universeSurvivors)` counts instruments with `Sharpe > 0 && !InsufficientData` from the sweep loop, which is the same count as `gateResult.PositiveSharpeInstruments` — they're equivalent by construction. The canonical field just makes it obvious.

---

### [TASK-0105] Tech debt — `pkg/strategy/price_exit.go`: document negative-pct behavior in `NewPriceExit` godoc

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-11
- **Source:** discovery
- **Context:** `NewPriceExit` godoc says "Set to 0 to disable" for stopLossPct and targetProfitPct. Negative values are silently treated as disabled (the `> 0` guard means negative inputs behave identically to zero). A caller passing `-0.05` expecting a 5% stop-loss would get no stop-loss at all without any error. Low risk for current internal usage, but the behavior should be named.
- **Acceptance criteria:**
  - [ ] One sentence added to `NewPriceExit` godoc: "Negative values are treated as 0 (disabled)."
  - [ ] `golangci-lint run ./pkg/strategy/...` still passes
  - [ ] `go1.25.0 test -race ./pkg/strategy/...` still passes
- **Notes:** Discovered during TASK-0098 multi-perspective review (Tech Debt Sentinel). One-line doc change, no production logic change. Low priority — PriceExit is an internal composable wrapper used by strategy authors, not a public API receiving untrusted input.

---

### [TASK-0112] Tech debt — `internal/paramsearch`: add read-only contract to `Config.StrategyFactory` godoc

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-13
- **Source:** discovery
- **Context:** `Config.StrategyFactory` receives the `params` map as a shared reference across concurrent goroutine calls for the same variant. The factory must not mutate or retain the map. No doc comment states this contract. The single internal caller (`cmd/param-search`) honors the contract, but the API is exported and the invariant is invisible.
- **Acceptance criteria:**
  - [ ] One sentence added to `StrategyFactory` field godoc in `internal/paramsearch/paramsearch.go` (~line 134): "params is shared read-only across concurrent goroutine calls for the same variant; the factory must not mutate or retain it."
  - [ ] `golangci-lint run ./internal/paramsearch/...` passes
  - [ ] `go1.25.0 test -race ./internal/paramsearch/...` passes
- **Notes:** Discovered during TASK-0077 multi-perspective review (Concurrency & State Safety Reviewer). One-sentence doc addition, no production logic change. Low priority — current caller is safe; this is a clarity/contract improvement.

---

<!-- === SMALL TEST ADDITIONS === -->

### [TASK-0108] Tech debt — `cmd/universe-sweep`: add `insufficient_data=true` CSV assertion to incomplete-data sweep test

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-13
- **Source:** discovery
- **Context:** `TestRun_IncompleteData_SweepContinues` in `cmd/universe-sweep/main_test.go` (added in TASK-0083) verifies that the sweep doesn't abort and that the diagnostic appears on stderr. It does not verify that the incomplete instrument's CSV row has `insufficient_data=true`. The behavior is guaranteed by the `internal/universesweep` layer, but the cmd-layer test should round-trip the full signal.
- **Acceptance criteria:**
  - [ ] `TestRun_IncompleteData_SweepContinues` updated to parse the CSV stdout and assert the NSE:RELIANCE row contains `insufficient_data=true` (6th column)
  - [ ] `go1.25.0 test -race ./cmd/universe-sweep/...` passes
  - [ ] `golangci-lint run ./cmd/universe-sweep/...` passes
- **Notes:** Discovered during TASK-0083 multi-perspective review (Test Coverage Auditor). One assertion added to an existing test — no new test function, no production code changes. The behavior is already guaranteed; this closes the cmd-layer assertion gap.

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

### [TASK-0093] Tech debt — `cmd/fetch-history`: add `os.MkdirAll` guard and two missing parseFlags tests

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-09
- **Source:** discovery
- **Context:** Three small gaps surfaced during the multi-perspective review of TASK-0070. (1) `fetchAll` does not call `os.MkdirAll(cacheDir)` before manifest operations — if `--cache-dir` doesn't exist yet, all manifest saves silently fail with warnings. (2) Two `parseFlags` paths are untested: invalid `--timeframe` value and missing `--access-token` with no env-var fallback.
- **Acceptance criteria:**
  - [ ] `fetchAll` in `cmd/fetch-history/main.go`: add `os.MkdirAll(cacheDir, 0o755)` call before `loadManifest` — ensures the cache root dir exists before any manifest writes
  - [ ] `TestRun_InvalidTimeframe` added: pass `--timeframe invalid-value`, assert error contains the invalid timeframe string
  - [ ] `TestRun_MissingAccessToken` added: no `--access-token` flag and no `KITE_ACCESS_TOKEN` env var set, assert error mentions `--access-token`
  - [ ] `go1.25.0 test -race ./cmd/fetch-history/...` passes
  - [ ] `golangci-lint run ./cmd/fetch-history/...` passes
- **Notes:** Discovered during multi-perspective review (Error Handling Inspector + Test Coverage Auditor). The MkdirAll fix is a correctness improvement for the edge case where `--cache-dir` doesn't exist yet; the two test additions close untested `parseFlags` branches. All three changes are in `cmd/fetch-history/` only.

---

### [TASK-0107] Tech debt — `internal/universesweep.Run`: synchronize stderr writes across goroutines

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-13
- **Source:** discovery
- **Context:** `universesweep.Run` fans out per-instrument engine runs via errgroup and passes a shared `stderr io.Writer` to each goroutine. When multiple instruments return `*ErrIncompleteData` concurrently, each goroutine calls `fmt.Fprintf(stderr, ...)` on the shared writer. `bytes.Buffer` (used in tests) is not goroutine-safe for concurrent writes — concurrent incomplete-data writes are a theoretical data race under `go test -race` on multi-core machines. Production uses `os.Stderr` whose small-write syscalls are atomic on Linux/macOS, so production is safe, but tests are technically racy.
- **Acceptance criteria:**
  - [ ] `Run` wraps `stderr` in a mutex-protected writer before passing to goroutines, OR adds a doc comment to `Run`'s signature stating "stderr must be safe for concurrent writes (os.Stderr is safe; bytes.Buffer is not)"
  - [ ] Preferred fix: introduce `syncWriter` in `internal/universesweep/universesweep.go`: `type syncWriter struct { mu sync.Mutex; w io.Writer }` with a `Write` method; wrap `stderr` in `Run` before goroutine launch
  - [ ] `TestRun_IncompleteData_AllInstrumentsIncomplete` (existing) passes reliably under race detector with `bytes.Buffer` after fix
  - [ ] `go1.25.0 test -race ./internal/universesweep/...` passes
  - [ ] `golangci-lint run ./internal/universesweep/...` passes
- **Notes:** Discovered during TASK-0083 multi-perspective review (Concurrency & State Safety Reviewer). Production is safe — `os.Stderr` write(2) syscalls ≤ PIPE_BUF are atomic on Linux/macOS. The risk is test-only: `bytes.Buffer` concurrent writes are a data race that the race detector may or may not catch depending on goroutine scheduling and GOMAXPROCS. syncWriter wrapper is ~10 lines.

---

### [TASK-0106] Tech debt — add `PriceExit` fold-isolation test to `internal/walkforward`

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-11
- **Source:** discovery
- **Context:** `internal/walkforward` has `TestRun_TimedExitFoldStateIsolation` that verifies a shared `TimedExit` instance causes cross-fold state corruption and a factory-constructed one does not. No equivalent test exists for `PriceExit`. Both are stateful wrappers that must not be shared across folds. The walk-forward factory API already enforces this structurally, but a regression test would document the broken behavior and prevent silent regressions.
- **Acceptance criteria:**
  - [ ] `TestRun_PriceExitFoldStateIsolation` added to `internal/walkforward/` test suite (or equivalent file matching `TestRun_TimedExitFoldStateIsolation`): construct a `PriceExit` wrapping a scripted inner, run two folds, verify second fold gets correct fresh state (entryPrice and inPosition reset)
  - [ ] `go1.25.0 test -race ./internal/walkforward/...` passes
  - [ ] `golangci-lint run ./internal/walkforward/...` passes
- **Notes:** Discovered during TASK-0098 multi-perspective review (Concurrency & State Safety Reviewer). Prerequisite: TASK-0074 or TASK-0075 must be built first so there is a realistic factory usage to model the test after. Low priority — the walk-forward factory API already enforces safety; this is a documentation-via-test improvement.

---

<!-- === MEDIUM REFACTORS === -->

### [TASK-0094] Refactor — `cmd/fetch-history/fetchOne`: reduce 11-parameter signature with `fetchState` struct

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-09
- **Source:** discovery
- **Context:** `fetchOne` currently takes 11 parameters — including `manifestPath`, `manifest *progressManifest`, and `completed map[string]bool` which are all shared loop state. The Naming & Clarity review (TASK-0070 multi-perspective review) flagged the 11-parameter count as a maintenance smell. A `fetchState` struct grouping these three would reduce cognitive load for anyone extending the function.
- **Acceptance criteria:**
  - [ ] `fetchState` struct introduced in `cmd/fetch-history/main.go`: fields `ManifestPath string`, `Manifest *progressManifest`, `Completed map[string]bool`
  - [ ] `fetchOne` signature reduced: `fetchOne(ctx, stdout, stderr, p, inst, tf, from, today, state *fetchState) error`
  - [ ] `fetchAll` updated to construct and pass `fetchState`
  - [ ] All existing tests pass: `go1.25.0 test -race ./cmd/fetch-history/...`
  - [ ] `golangci-lint run ./cmd/fetch-history/...` passes
  - [ ] No behavior change — pure refactor
- **Notes:** Discovered during multi-perspective review (Naming & Clarity Guardian). Low priority: existing 11-parameter signature is comprehensible and lint-clean; this is a readability improvement, not a correctness fix.

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
- **Notes:** Donchian has only one meaningful sweep parameter (period) — its p2 axis is less obvious; defer the axis mapping decision until this task is picked up. **Remaining scope (updated 2026-05-08 after strategy wiring centralization):** `cmd/sweep` and all other cmd mains now use `cmdutil.GlobalRegistry` for construction — `fixedParams` duplication and `MustGet` discarded-return issues are resolved. `cmd/sweep2d` is the only remaining outlier: it still has a local `factoryRegistry2D` switch and doesn't use GlobalRegistry for name validation. The acceptance criteria for this task should focus on: (1) extending `factoryRegistry2D` to cover all strategies including `cci-mean-reversion`, and (2) wiring `GlobalRegistry.MustGet` for name validation in `cmd/sweep2d`, matching the pattern in all other cmd mains. **Note (2026-05-13):** Do after TASK-0074 and TASK-0075 are implemented — sweep2d needs ORB and Gap-and-Go axis mappings.

---

### [TASK-0102] Tech debt — `cmd/evaluate`: wire `universesweep.Result.Trades` to skip bootstrap engine re-run

- **Status:** todo
- **Priority:** low
- **Created:** 2026-05-10
- **Source:** discovery
- **Context:** `collectTrades` in `cmd/evaluate/main.go` runs a full engine re-run per instrument to collect trades for bootstrap resampling. `universesweep.Result` already has a `Trades []model.Trade` field populated during the universe sweep. The bootstrap stage discards those and re-fetches candles + re-runs the engine. For large universes (40+ instruments), this doubles provider calls and CPU time. The data is already available — it just isn't threaded through.
- **Acceptance criteria:**
  - [ ] `runPipeline` passes `sweepReport.Results` trades through to `runBootstrap` instead of calling `collectTrades` per instrument
  - [ ] `collectTrades` function removed or left as a fallback only
  - [ ] `runBootstrap` signature updated to accept pre-collected trades: `runBootstrap(pl, wfGatePassed, instrumentSharpe, instrumentTrades map[string][]model.Trade, stderr)` — or equivalent
  - [ ] Existing bootstrap tests still pass: `go1.25.0 test -race ./cmd/evaluate/...`
  - [ ] `golangci-lint run ./cmd/evaluate/...` passes
- **Notes:** Discovered during TASK-0073 multi-perspective review (Tech Debt Sentinel). For correctness: the universe sweep uses the same `[from, to)` range, same strategy, same commission model as bootstrap — the trades are identical. The only edge case: if bootstrap re-run is intentionally different (different date range is currently impossible since it uses pl.from/pl.to). Low priority; correctness is unaffected — this is pure efficiency.

---

<!-- === LARGER ITEMS === -->

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

### [TASK-0130] Fix — `internal/analytics`: timeframe-aware `MinCurvePointsForMetrics` + add `NSERegimes5Min` for 5-min strategies

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-17
- **Source:** discovery
- **Context:** Two daily-specific assumptions in `internal/analytics/` break for 5-min strategies. (1) `MinCurvePointsForMetrics = 252` — correct for daily (252 days = 1 year), wrong for 5-min (252 bars = 3.4 trading days). If this threshold is used to gate Sharpe computation, it's trivially easy to pass at 5-min even on a meaningless 4-day backtest. Should be timeframe-aware: `MinCurvePoints(tf) = sharpeAnnualizationFactor(tf)` (i.e., one year of bars for the given timeframe). (2) `NSERegimes2018_2024` includes a Pre-COVID window (2018–Jan 2020) for which there is zero 5-min data. Any 5-min regime analysis using this variable gets empty data for pre-COVID. Need `NSERegimes5Min2021_2024` (three regimes within the 2021–2024 data window: recovery/bull 2021–Q1 2022, rate shock 2022, grind 2022–2024).
- **Acceptance criteria:**
  - [ ] `MinCurvePointsForMetrics` replaced with `MinCurvePoints(tf model.Timeframe) int` function: returns `sharpeAnnualizationFactor(tf)` as int — 252 for daily, 18900 for 5-min, 94500 for 1-min
  - [ ] All callers of `MinCurvePointsForMetrics` updated to `MinCurvePoints(tf)` with appropriate timeframe passed through
  - [ ] `NSERegimes5Min2021_2024` defined in `internal/analytics/regime.go`: three windows within 2021–2024 (recovery/bull, rate-shock, grind); dates to be determined based on actual NSE 5-min data availability
  - [ ] Existing `NSERegimes2018_2024` and `NSERegimesGate` unchanged — daily strategy evaluation not affected
  - [ ] All existing tests pass; new tests for `MinCurvePoints` covering all supported timeframes
  - [ ] `go1.25.0 test -race ./internal/analytics/...` passes
  - [ ] `golangci-lint run ./internal/analytics/...` passes
  - [ ] Tests written before implementation (TDD)
- **Notes:** `MinCurvePointsForMetrics` is a const today — changing it to a function is a breaking change to the public API of `internal/analytics`. All callers must be updated (search for `MinCurvePointsForMetrics` across codebase). The practical impact on existing daily-bar backtests: none (252 → `MinCurvePoints(TimeframeDaily)` = 252, same value). The impact on 5-min backtests: prevents spurious Sharpe computation on tiny evaluation windows.

---

### [TASK-0131] Feature — `pkg/strategy/vwap.go`: session-aware VWAP computation utility

- **Status:** todo
- **Priority:** medium
- **Created:** 2026-05-17
- **Source:** session
- **Context:** VWAP (Volume-Weighted Average Price) is the primary institutional benchmark for intraday execution. Composite strategies that use price-vs-VWAP as a regime filter (e.g., MACD+VWAP: take long signals only when price > VWAP) require a session-reset VWAP computation. VWAP = cumulative(Price × Volume) / cumulative(Volume), reset at session open. NSE 5-min bars include volume data. This utility is a prerequisite for TASK-0128 (composite signal design).
- **Acceptance criteria:**
  - [ ] `VWAP` type in `pkg/strategy/vwap.go`: pre-allocated struct with running `cumulativePV float64`, `cumulativeVolume float64`, `lastSessionDate time.Time`
  - [ ] `VWAP.Update(bar model.Candle) float64`: updates running VWAP, detects session reset via `IsSessionOpen` from `pkg/strategy/session.go`, returns current VWAP; session reset: zero both accumulators before processing the new bar
  - [ ] `VWAP.Current() float64`: returns current VWAP without updating (0 before first bar of session)
  - [ ] No allocations in `Update` (pre-allocated struct, no slice growth)
  - [ ] `NewVWAP() *VWAP` constructor
  - [ ] Golden tests: single-session VWAP computation verified against manual calculation; session reset verified (second session starts fresh); zero-volume bar handling (skip or treat as previous VWAP)
  - [ ] `go1.25.0 test -race ./pkg/strategy/...` passes
  - [ ] `golangci-lint run ./pkg/strategy/...` passes
  - [ ] Tests written before implementation (TDD)
- **Notes:** Owner: Priya (dev). Price for VWAP = `(High + Low + Close) / 3` (typical price), standard convention. Session reset uses `IsSessionOpen` from `pkg/strategy/session.go` (TASK-0078, done) — a bar is the first of a new session when the previous bar was not in a session or has a different date. Zero-volume bars: if `bar.Volume == 0`, carry forward previous VWAP without updating accumulators. TASK-0128 is blocked on this.

---

_Completed and cancelled tasks are moved to `tasks/archive/YYYY-MM.md`_
