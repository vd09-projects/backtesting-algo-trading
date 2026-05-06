# Task Log

Append-only record of all task operations. Newest entries at the bottom.

| Date | Task | Action | Details | Notes |
|------|------|--------|---------|-------|
| 2026-04-10 | TASK-0011 | status → done | All acceptance criteria met | cmd/backtest/main.go + strategies/stub/ |
| 2026-04-10 | TASK-0025 | status → in-progress | starting corporate action verification for Zerodha data | |
| 2026-04-10 | TASK-0025 | status → done | Kite day candles adjusted for splits/bonuses; decision recorded in decisions/infrastructure/ | TASK-0012 and TASK-0015 unblocked |
| 2026-04-01 00:00 | TASK-0001 | created | priority: critical, source: project | Initialize Go module and folder structure (original) |
| 2026-04-26 | TASK-0041 | status → done | All acceptance criteria met; strategies/macd/ + CLI wiring in backtest/sweep/universe-sweep | |
| 2026-04-16 | TASK-0029 | status → done | all 4 acceptance criteria met | `--output-curve` flag, Config.CurvePath/Curve, writeCurveCSV, round-trip test at 100% |
| 2026-04-14 | TASK-0015 | status → done | All acceptance criteria met; commit 3f98308 | strategies/rsimeanrev/ |
| 2026-04-14 | TASK-0018 | status → done | All acceptance criteria met; commit 5e05325 | analytics.ComputeBenchmark + output.Config.Benchmark; 100% analytics coverage |
| 2026-04-01 00:00 | TASK-0002 | created | priority: critical, source: project | Define core domain types in pkg/model |
| 2026-04-01 00:00 | TASK-0003 | created | priority: high, source: project | Define DataProvider interface in pkg/provider (original) |
| 2026-04-01 00:00 | TASK-0004 | created | priority: high, source: project | Define Strategy interface in pkg/strategy (original) |
| 2026-04-01 00:00 | TASK-0005 | created | priority: high, source: project | Build engine with trivial test strategy (original, monolithic) |
| 2026-04-01 00:00 | TASK-0006 | created | priority: medium, source: project | Add analytics for basic performance metrics |
| 2026-04-01 00:00 | TASK-0007 | created | priority: medium, source: project | Implement Zerodha Kite Connect data provider (original, monolithic) |
| 2026-04-01 00:00 | TASK-0001 | decomposed | merged TASK-0001 + TASK-0003 + TASK-0004 → new TASK-0001 | Original 3 were each <1 day; combined into 3-4 day scaffolding + interface design task |
| 2026-04-01 00:00 | TASK-0005 | decomposed | split into TASK-0003, TASK-0004, TASK-0005 | Engine was 2+ week monolith; split into event loop / portfolio state / execution model |
| 2026-04-01 00:00 | TASK-0007 | decomposed | split into TASK-0007 (analysis), TASK-0008, TASK-0009 | Zerodha provider split; analysis ticket added because auth flow and API limits are unknown |
| 2026-04-01 00:00 | TASK-0003 | created | priority: high, source: project | Engine event loop and candle feeding (from decompose) |
| 2026-04-01 00:00 | TASK-0004 | created | priority: high, source: project | Engine portfolio state and trade log (from decompose) |
| 2026-04-01 00:00 | TASK-0005 | created | priority: high, source: project | Engine order execution model and benchmark (from decompose) |
| 2026-04-01 00:00 | TASK-0007 | created | priority: high, source: discovery | [ANALYSIS] Zerodha API auth flow, rate limits, and data constraints |
| 2026-04-01 00:00 | TASK-0008 | created | priority: medium, source: project | Zerodha provider auth and FetchCandles; blocked on TASK-0007 |
| 2026-04-01 00:00 | TASK-0009 | created | priority: medium, source: project | Zerodha provider local caching layer; blocked on TASK-0007 |
| 2026-04-02 00:00 | TASK-0001 | status → in-progress | starting implementation | |
| 2026-04-02 00:00 | TASK-0001 | status → done | all acceptance criteria met | go mod init, folder structure, deps, DataProvider + Strategy interfaces, compile-time check tests, go test ./... passes |
| 2026-04-02 00:00 | TASK-0002 | status → in-progress | starting implementation | |
| 2026-04-02 00:00 | TASK-0002 | status → done | all acceptance criteria met | Candle/Timeframe/Signal/Position/Trade/OrderConfig in separate files; table-driven tests for Duration, Candle validation, Trade P&L; go test ./... passes |
| 2026-04-02 00:00 | TASK-0003 | status → in-progress | starting implementation | |
| 2026-04-02 00:00 | TASK-0003 | status → done | all acceptance criteria met | Engine struct, EngineConfig, Run method, BarResult; no-lookahead + lookback enforced; 8 tests passing |
| 2026-04-03 00:00 | TASK-0004 | status → in-progress | starting implementation | |
| 2026-04-03 00:00 | TASK-0004 | status → done | all acceptance criteria met | Portfolio (cash, positions, trade log), applySignal wired into Run; no-pyramid guard, insufficient-cash guard; 9 tests passing |
| 2026-04-04 00:00 | TASK-0005 | status → in-progress | starting implementation | |
| 2026-04-04 00:00 | TASK-0005 | status → done | all acceptance criteria met | Fill at next open via pending-signal buffer; slippage (pct); CommissionFlat/Percentage/Zerodha (₹20 cap); Trade.Commission field; 22 tests + BenchmarkEngineRun 256µs/op (budget 1ms) |
| 2026-04-06 00:00 | TASK-0006 | status → in-progress | starting implementation | |
| 2026-04-06 00:00 | TASK-0006 | status → done | all acceptance criteria met | Compute([]Trade) Report; TotalPnL, WinRate, MaxDrawdown (equity curve), TradeCount, WinCount, LossCount; 7 tests covering empty/single/all-winners/all-losers/mixed/breakeven; go test -race passes |
| 2026-04-07 00:00 | TASK-0007 | status → done | all acceptance criteria met | Six decisions recorded in decisions/infrastructure/; auth prototype at cmd/authtest/main.go; unblocks TASK-0008 and TASK-0009 |
| 2026-04-07 00:00 | TASK-0008 | status → todo | blocker TASK-0007 resolved | unblocked |
| 2026-04-07 00:00 | TASK-0009 | status → todo | blocker TASK-0007 resolved | unblocked |
| 2026-04-07 00:00 | TASK-0010 | created | priority: medium, source: project | Output package — result formatting and JSON export |
| 2026-04-07 00:00 | TASK-0011 | created | priority: medium, source: project | CLI entrypoint — cmd/backtest wiring |
| 2026-04-07 00:00 | TASK-0012 | created | priority: low, source: project | First concrete strategy — SMA crossover |
| 2026-04-08 00:00 | TASK-0008 | status → in-progress | starting implementation | |
| 2026-04-09 00:00 | TASK-0008 | status → done | all acceptance criteria met; archived | |
| 2026-04-09 00:00 | TASK-0009 | status → in-progress | pkg/provider/zerodha/cache/ implemented; all tests green | |
| 2026-04-09 00:00 | TASK-0009 | status → done | verified end-to-end via cmd/providertest; archived | |
| 2026-04-09 00:00 | TASK-0010 | status → done | all acceptance criteria met; archived to 2026-04.md | |
| 2026-04-10 00:00 | TASK-0013 | created | priority: high, source: session | Equity curve time series in Portfolio — prerequisite for Sharpe |
| 2026-04-10 00:00 | TASK-0014 | created | priority: high, source: session | Annualized Sharpe ratio in analytics; blocked by TASK-0013 |
| 2026-04-10 00:00 | TASK-0011 | reprioritized | medium → high | Critical path for running any end-to-end test |
| 2026-04-10 00:00 | TASK-0012 | reprioritized | low → high | First runnable strategy needed before anything can be validated |
| 2026-04-10 00:00 | TASK-0015 | created | priority: high, source: session | RSI mean-reversion strategy — second baseline |
| 2026-04-10 00:00 | TASK-0016 | created | priority: medium, source: session | Profit factor, avg win/loss, Sortino ratio |
| 2026-04-10 00:00 | TASK-0017 | created | priority: medium, source: session | Drawdown duration tracking |
| 2026-04-10 00:00 | TASK-0018 | created | priority: medium, source: session | Buy-and-hold benchmark comparison |
| 2026-04-10 00:00 | TASK-0019 | created | priority: medium, source: session | MACD trend-following strategy |
| 2026-04-10 00:00 | TASK-0020 | created | priority: medium, source: session | Bollinger Band mean-reversion strategy |
| 2026-04-10 00:00 | TASK-0021 | created | priority: medium, source: session | Volatility-based position sizing |
| 2026-04-10 00:00 | TASK-0022 | created | priority: low, source: session | Walk-forward validation framework |
| 2026-04-10 00:00 | TASK-0023 | created | priority: low, source: session | Parameter sweep runner |
| 2026-04-10 00:00 | TASK-0024 | created | priority: low, source: session | Monte Carlo bootstrap for Sharpe confidence intervals |
| 2026-04-10 01:00 | TASK-0025 | created | priority: high, source: session | Data quality — verify corporate action handling in Zerodha historical data; placed before TASK-0012 |
| 2026-04-10 01:00 | TASK-0018 | reprioritized | moved from Todo (Backlog) to Up Next | Must be available when first strategy results arrive, not after profit factor/Sortino |
| 2026-04-10 01:00 | TASK-0021 | reprioritized | moved from Todo (Backlog) to Up Next | Sizing must be in place before more strategies are added, not retrofitted later |
| 2026-04-10 01:00 | TASK-0019 | updated | added explicit conditional gate: cancel if SMA crossover + RSI both fail Sharpe >= 0.5 vs buy-and-hold | |
| 2026-04-10 01:00 | TASK-0020 | updated | added explicit conditional gate: cancel if RSI mean-reversion fails Sharpe >= 0.5 vs buy-and-hold | |
| 2026-04-10 02:00 | TASK-0016 | updated | added CalmarRatio to acceptance criteria; Calmar = annualized return / max drawdown | |
| 2026-04-10 02:00 | TASK-0014 | updated | fixed NSE 15min annualization factor: 252*26 → 252*25 (NSE session = 375 min = 25 bars/day) | |
| 2026-04-10 03:00 | TASK-0016 | updated | title updated to include Calmar; now matches acceptance criteria scope | |
| 2026-04-10 03:00 | TASK-0021 | updated | vol sizing criterion clarified: formula yields notional (₹), divide by fillPrice to get quantity; instrumentVol specified as non-annualized daily std dev | |
| 2026-04-10 04:00 | TASK-0013 | status → done | implemented EquityPoint model, Portfolio.RecordEquity, Portfolio.EquityCurve, engine wiring, pre-allocated slice | all tests green, lint clean |
| 2026-04-10 05:00 | TASK-0014 | status → done | SharpeRatio field on Report; computeSharpe from equity curve per-bar returns; annualization for all 5 timeframes; 11 table-driven tests; output.printSummary updated; analytics 96.3% coverage | all tests green, 0 lint issues |
| 2026-04-13 00:00 | TASK-0012 | status → done | strategies/smacrossover/ implemented; --fast-period/--slow-period flags wired into cmd/backtest; table-driven tests passing | archived to 2026-04.md |
| 2026-04-13 01:00 | TASK-0021 | reprioritized | moved to top of Up Next; fixed-fraction sizing produces non-comparable results across strategies with different hold durations; must be done before strategy results are interpreted | per Marcus review |
| 2026-04-13 01:00 | TASK-0015 | updated | notes: added edge thesis (retail panic → mispricing absorbed by larger participants) and exit-rule gap (no stop-loss; indefinite hold if RSI never recovers) | per Marcus review |
| 2026-04-13 01:00 | TASK-0023 | reprioritized | low → medium; run immediately after first strategy results, not after four strategies are built; if RSI(14)/30/70 is a local peak and not a plateau, there is no edge | per Marcus review |
| 2026-04-13 01:00 | TASK-0024 | reprioritized | low → medium; must run before TASK-0022 and TASK-0023; bootstrapped distribution is the input to kill-switch definition (TASK-0026) | per Marcus review |
| 2026-04-13 01:00 | TASK-0026 | created | priority: high, source: session; blocked by TASK-0024; kill-switch thresholds per strategy before any live capital | per Marcus review |
| 2026-04-13 02:00 | TASK-0015 | updated | notes: added holdout declaration (2015-2022 train, 2023+ holdout) | per Marcus review |
| 2026-04-13 02:00 | TASK-0021 | updated | notes: added holdout declaration | per Marcus review |
| 2026-04-13 02:00 | TASK-0016 | updated | title and acceptance criteria: added TailRatio (95th/5th percentile return); renamed to include tail ratio | per Marcus review |
| 2026-04-13 02:00 | TASK-0017 | updated | notes: removed stale TASK-0013 blocker reference; TASK-0013 is done, task is ready to implement | per Marcus review |
| 2026-04-13 02:00 | TASK-0027 | created | priority: medium, source: session; strategy correlation analysis before portfolio assembly; do not start until 2+ strategy results exist | per Marcus review |
| 2026-04-13 | TASK-0021 | status → in-progress | implementation complete: pkg/model/sizing.go (SizingModel enum), internal/engine/sizing.go (computeInstrumentVol + sizeFractionForBar), engine.Config extended; all acceptance criteria met; 98.5% coverage | |
| 2026-04-13 | TASK-0021 | status → done | all acceptance criteria met; archived to 2026-04.md | commit 5852d63 |
| 2026-04-13 | TASK-0015 | status → in-progress | strategies/rsimeanrev/ created; all 6 acceptance criteria met; 8 tests passing (go test -race); hand-verified RSI values in test comments | |
| 2026-04-14 | TASK-0023 | status → done | internal/sweep/ (Config/Result/Report/PlateauRange, Run, computePlateau); internal/output/output.go (WriteSweep); cmd/sweep/main.go; all tests green, 0 lint issues | archived to 2026-04.md |
| 2026-04-14 | TASK-0016 | status → done | analytics.Report +6 fields (ProfitFactor, AvgWin, AvgLoss, SortinoRatio, CalmarRatio, TailRatio); computeReturns extracted; 8 new tests; output.printSummary updated; all tests green, 0 lint issues | archived to 2026-04.md |
| 2026-04-15 | TASK-0028 | created | priority: high, source: user; run both baseline strategies on declared Nifty 50 instrument 2018–2024, check proliferation gate (Sharpe ≥ 0.5), record gate decisions in decisions/algorithm/ | per Marcus + Priya review |
| 2026-04-15 | TASK-0017 | moved | Todo (Backlog) → Up Next; run before TASK-0028 for cleaner output | |
| 2026-04-16 | TASK-0017 | status → done | MaxDrawdownDuration added to analytics.Report; computeMaxDrawdownDuration from per-bar equity curve; 5 table-driven tests; printSummary updated; lint clean | archived to 2026-04.md |
| 2026-04-16 | TASK-0028 | status → in-progress | instrument declared: NSE:RELIANCE; both runs complete; gate failed for both strategies (SMA Sharpe=0.447, RSI Sharpe=0.469); MaxDrawdown bug fixed (computeMaxDrawdownDepth from equity curve) | remaining: gate decisions in decisions/algorithm/, regime window review |
| 2026-04-16 | TASK-0028 | criteria update | gate decisions recorded: sma-crossover-proliferation-gate-failed.md + rsi-mean-reversion-proliferation-gate-failed.md | 6/7 criteria done; only regime window review remains |
| 2026-04-16 | TASK-0019 | status → cancelled | SMA crossover failed proliferation gate (Sharpe 0.447); MACD not built per gate rule | archived to 2026-04.md |
| 2026-04-16 | TASK-0020 | status → cancelled | RSI mean-reversion failed proliferation gate (Sharpe 0.469, 7 trades); Bollinger Bands not built per gate rule | archived to 2026-04.md |
| 2026-04-16 | TASK-0029 | created | priority: high, source: session | equity curve CSV output; unblocks TASK-0028 regime review |
| 2026-04-16 | TASK-0030 | created | priority: high, source: session | signal frequency gate N<30 in analytics.Compute |
| 2026-04-16 | TASK-0031 | created | priority: medium, source: session | RSI signal frequency diagnostic on RELIANCE; pre-condition for mean-reversion re-test |
| 2026-04-16 | TASK-0032 | created | priority: medium, source: session | 2D parameter sweep + DSR calculation; internal/sweep2d |
| 2026-04-16 | TASK-0033 | created | priority: medium, source: session | automated proliferation gate PASS/FAIL in CLI output; depends on TASK-0030 |
| 2026-04-16 | TASK-0034 | created | priority: medium, source: session | regime-split report in analytics; depends on TASK-0029 |
| 2026-04-16 | TASK-0035 | created | priority: low, source: session | multi-instrument sweep CLI cmd/universe-sweep; depends on TASK-0030 |
| 2026-04-16 | TASK-0036 | created | priority: low, source: session | Python notebooks layer + file contract |
| 2026-04-16 | TASK-0024 | criteria update | added Trade.ReturnOnNotional() requirement + explicit Seed int64 in BootstrapConfig for determinism | session review surfaced these gaps |
| 2026-04-19 | TASK-0030 | status → done | MinTradesForMetrics=30, MinCurvePointsForMetrics=252 constants + flags in Report; gate zeroes metrics; warnings in printSummary; all tests pass | math tests split into analytics_internal_test.go; sweep golden test updated to 300 candles |
| 2026-04-19 | TASK-0031 | status → done | cmd/rsi-diagnostic built; RELIANCE 2018–2025: 52 oversold bars, 147 overbought, 199 total signal bars — thresholds NOT miscalibrated; root cause: long-only strategy requires RSI<30→RSI>70 cycle; RELIANCE trending behaviour means overbought bars mostly fire with no open long; decision recorded in decisions/algorithm/2026-04-19-rsi-signal-frequency-diagnostic-reliance.md | archived to 2026-04.md |
| 2026-04-19 | TASK-0032 | status → done | internal/sweep2d (Run, Config2D, ParamRange, GridCell, Report2D, WriteCSV); internal/analytics/dsr.go (DSR + normInvCDF); sweep.Report gains VariantCount+NObservations; WriteSweep prints DSR-corrected peak Sharpe; strategies/testutil gains StaticProvider, ThresholdStrategy, MakeAlternatingCandles, TestEngineConfig; golang.org/x/sync v0.20.0 added; all tests pass race detector | archived to 2026-04.md |
| 2026-04-19 | TASK-0033 | status → done | GateThreshold float64 in output.Config; --proliferation-gate-threshold flag in cmd/backtest; gate logic in printSummary (skipped when threshold=0 or insufficient sample); 5 new tests in output_test.go; all tests pass race detector | archived to 2026-04.md |
| 2026-04-20 | TASK-0034 | status → in-progress | starting implementation | directly unblocks TASK-0028 final criterion |
| 2026-04-20 | TASK-0034 | status → done | all 5 criteria met; regime.go + regime_test.go + output.Config.RegimeSplits + printRegimeTable; quality gate passed | archived to 2026-04.md |
| 2026-04-20 | TASK-0028 | status → done | all 7 criteria met; regime review complete (SMA 0.35/0.73/0.37, RSI 1.10/0.21/0.44); gate failure confirmed | archived to 2026-04.md |
| 2026-04-20 | TASK-0024 | status → done | all 6 criteria met; internal/montecarlo + Trade.ReturnOnNotional; per-trade non-annualized Sharpe (Marcus sign-off); quality gate passed 93.3%/100% | archived to 2026-04.md |
| 2026-04-20 | TASK-0026 | status → todo (unblocked) | TASK-0024 complete; moved from Blocked to Up Next | kill-switch now implementable |
| 2026-04-21 | TASK-0026 | status → done | KillSwitchThresholds + CheckKillSwitch in internal/analytics/killswitch.go; 3 decision files in decisions/algorithm/; 61 tests pass; bootstrap p5 Sharpe pending token refresh (TASK-0037) | archived to 2026-04.md |
| 2026-04-21 | TASK-0037 | created | priority: low, source: session | Bootstrap re-run to fill kill-switch p5 Sharpe thresholds for SMA + RSI strategies |
| 2026-04-21 | TASK-0027 | status → done | all acceptance criteria met; correlation.go + load.go + cmd/correlate + WriteCorrelationMatrix; 13 tests pass, lint clean | archived to 2026-04.md |
| 2026-04-22 | TASK-0022 | status → in-progress | picked from Todo (Backlog); Marcus pre-check: walk-forward = regime-stability test; Priya plan: internal/walkforward/ harness | |
| 2026-04-22 | TASK-0022 | status → done | all 5 acceptance criteria met; internal/walkforward/walkforward.go + walkforward_test.go; 17 tests pass; lint clean; 9 decisions harvested | archived to 2026-04.md |
| 2026-04-22 | TASK-0035 | status → done | all 5 acceptance criteria met; internal/universesweep/ + cmd/universe-sweep/ + universes/nifty50-large-cap.yaml; buildProvider extracted to cmdutil; 9 tests pass; lint clean | archived to 2026-04.md |
| 2026-04-25 | TASK-0038 | created | priority: high, source: session | Full NSE cost model CNC delivery — CommissionZerodhaFull with STT, exchange charges, GST, SEBI, stamp duty |
| 2026-04-25 | TASK-0039 | created | priority: high, source: session | TimedExit strategy wrapper — N-bar hold exit in pkg/strategy |
| 2026-04-25 | TASK-0040 | created | priority: high, source: session | Donchian Channel Breakout strategy — strategies/donchian/ |
| 2026-04-25 | TASK-0041 | created | priority: high, source: session | MACD Crossover strategy — strategies/macd/; supersedes cancelled TASK-0019 under new cross-instrument evaluation methodology |
| 2026-04-25 | TASK-0042 | created | priority: high, source: session | Bollinger Band Mean Reversion strategy — strategies/bollinger/; supersedes cancelled TASK-0020 under new cross-instrument evaluation methodology |
| 2026-04-25 | TASK-0043 | created | priority: high, source: session | 12-Month Rate-of-Change Momentum strategy — strategies/momentum/; 231-bar skip-last-month convention (Marcus) |
| 2026-04-25 | TASK-0044 | created | priority: high, source: session | cmd/sweep2d CLI entrypoint — wires existing internal/sweep2d package |
| 2026-04-25 | TASK-0045 | created | priority: high, source: session | NIFTY TRI benchmark research spike — 2hr timebox; Zerodha or NSE CSV |
| 2026-04-25 | TASK-0046 | created | priority: high, source: session | Session-boundary engine support for intraday — BLOCKED on Marcus fill-price + bar-granularity decisions |
| 2026-04-25 | TASK-0047 | created | priority: high, source: session | MIS commission model (intraday STT 0.025% sell-only) — BLOCKED on TASK-0038 |
| 2026-04-25 | TASK-0048 | created | priority: high, source: session | Weekly kill-switch monitor cmd/monitor — BLOCKED on trade log file format decision |
| 2026-04-25 | TASK-0049 | created | priority: high, source: session | Evaluation pre-commit gate definitions — BLOCKED on TASK-0038; owner: Marcus |
| 2026-04-25 | TASK-0050 | created | priority: high, source: session | Signal frequency audit 6 strategies × 15 instruments — BLOCKED on TASK-0038/0040-0043; owner: Marcus |
| 2026-04-25 | TASK-0051 | created | priority: high, source: session | In-sample baseline + parameter sensitivity, RELIANCE 2018-2024 — BLOCKED on TASK-0049/0050; owner: Marcus |
| 2026-04-25 | TASK-0052 | created | priority: high, source: session | Universe sweep cross-instrument primary gate — BLOCKED on TASK-0051/0044; owner: Marcus |
| 2026-04-25 | TASK-0053 | created | priority: high, source: session | Walk-forward validation on universe survivors — BLOCKED on TASK-0052; owner: Marcus |
| 2026-04-25 | TASK-0054 | created | priority: high, source: session | Monte Carlo bootstrap on walk-forward survivors — BLOCKED on TASK-0053; owner: Marcus |
| 2026-04-25 | TASK-0055 | created | priority: high, source: session | Cross-strategy correlation and portfolio construction — BLOCKED on TASK-0054; owner: Marcus |
| 2026-04-25 | TASK-0056 | created | priority: high, source: session | Pre-live brief kill-switch thresholds and go/no-go sign-off — BLOCKED on TASK-0055/0048; owner: Marcus |
| 2026-04-25 | TASK-0038 | status → done | all criteria met; commission.go (new), commission_zerodha_full_test.go (new), portfolio.go (modified), pkg/model/order.go (modified); ₹88.24 round-trip on ₹30K hand-verified; quality gate PASS; 5 decisions harvested | archived to 2026-04.md |
| 2026-04-25 | TASK-0047 | status → todo (unblocked) | TASK-0038 complete; moved from Blocked to Up Next | side-aware architecture in place for MIS extension |
| 2026-04-25 | TASK-0057 | created | priority: low, source: decision | Migrate engine accounting layer from float64 to shopspring/decimal; deferred from TASK-0038 decision 2026-04-25-float64-for-commission-arithmetic |
| 2026-04-25 | TASK-0046 | blocker updated | methodology questions answered by Marcus (Decision 2026-04.3.0 + 2026-04.3.1); now blocked on phase sequencing only |
| 2026-04-25 | TASK-0051 | title corrected | "2018-2024" → "2018-2023" to match acceptance criteria (to date 2024-01-01 is exclusive) |
| 2026-04-25 | TASK-0052 | blocker corrected | removed spurious TASK-0044 dependency; cmd/universe-sweep exists from TASK-0035, sweep2d not required for universe gate |
| 2026-04-25 | TASK-0039 | reprioritized | moved from position 2 to position 6 in Up Next (after TASK-0043); not on critical path for evaluation pipeline — strategies 0040-0043 unblock TASK-0050 and must be picked up first |
| 2026-04-25 12:28 | TASK-0040 | status → done | all criteria met, quality gate PASS | |
| 2026-04-27 | TASK-0042 | status → done | all 9 criteria met; strategies/bollinger/ + CLI wiring in all three CLIs; tests first (TDD) | |
| 2026-04-27 | TASK-0043 | status → done | all 8 criteria met; strategies/momentum/ + CLI wiring in all three CLIs; cmd/sweep factoryRegistry refactored into per-strategy helpers to satisfy cyclop limit | archived to 2026-04.md |
| 2026-04-27 | TASK-0050 | status → todo (unblocked) | TASK-0043 complete — all 6 strategies implemented; moved from Blocked to Up Next | |
| 2026-04-27 | TASK-0058 | created | priority: medium, source: discovery | cmd/rsi-diagnostic/main.go cyclop complexity 17 > 15; pre-existing, surfaced during TASK-0043 build |
| 2026-04-27 | TASK-0039 | status → done | all 7 criteria met; pkg/strategy/timed_exit.go + timed_exit_test.go; 8 tests pass, quality gate PASS | archived to 2026-04.md |
| 2026-04-27 | TASK-0059 | created | priority: medium, source: session | walk-forward Run() factory API for stateful wrappers; triggered by TimedExit statefulness (TASK-0039) |
| 2026-04-27 | TASK-0047 | status → done | all 5 criteria met; CommissionZerodhaFullMIS + calcZerodhaFullMISCommission + portfolio switch case; 5 golden tests pass, quality gate PASS | archived to 2026-04.md |
| 2026-04-27 | TASK-0049 | status → todo (unblocked) | TASK-0047 done — MIS commission model complete; moved from Blocked to Up Next | |
| 2026-04-27 | TASK-0060 | created | priority: medium, source: discovery | --commission CLI flag for cmd/backtest, cmd/sweep, cmd/universe-sweep; discovered during TASK-0047 harvest (CLIs hardcode CommissionZerodha) |
| 2026-04-27 | TASK-0044 | status → done | all 6 criteria met; cmd/sweep2d/main.go + main_test.go; 5 tests (TDD), quality gate PASS | archived to 2026-04.md |
| 2026-04-27 | TASK-0061 | created | priority: low, source: session | extend cmd/sweep2d factoryRegistry to all 6 strategies + resolve fixedParams duplication with cmd/sweep |
| 2026-04-28 | TASK-0045 | status → done | research spike complete; NIFTY 50 TRI not in Kite; decision recorded in decisions/infrastructure/2026-04-28-nifty-tri-benchmark-data-source.md | archived to 2026-04.md |
| 2026-04-28 | TASK-0062 | created | priority: medium, source: decision | NIFTY 50 TRI benchmark: download NSE CSV + implement StaticCSVProvider in pkg/provider/csv/ | spawned from TASK-0045 decision |
| 2026-04-29 | TASK-0050 | status → done | internal/signalaudit + cmd/signal-audit implemented; 11 tests (TDD), quality gate PASS (89.8% coverage, 0 lint issues, race clean) | archived to 2026-04.md |
| 2026-04-29 | TASK-0051 | status → in-progress | tooling gate complete: --commission flag added to cmd/backtest + cmd/sweep; ParseCommissionModel in internal/cmdutil; sweep.computePlateau updated to valid-region (≥30 trades) logic with SensitivityConcern field; quality gate PASS (92.4% coverage); remaining: CLI runs requiring live Zerodha token | |
| 2026-04-29 | TASK-0060 | scope updated | cmd/backtest + cmd/sweep --commission done in TASK-0051; scope narrowed to cmd/universe-sweep only; ParseCommissionModel already in internal/cmdutil | |
| 2026-04-29 | TASK-0063 | created | priority: low, source: discovery | cmd/backtest package doc comment Available strategies lists only 3 strategies; cosmetic fix alongside next cmd/backtest touch |
| 2026-05-01 10:00 | TASK-0064 | created | priority: medium, source: discovery | runs output missing timeframe/metadata in filename and JSON |
| 2026-05-01 | TASK-0051 | status → done | All acceptance criteria met: 6 baseline runs (runs/baseline-2026-04-30/), 6 sweeps, plateau-params.json produced; Step 4 signal audit with plateau-midpoint params across 15 instruments → runs/baseline-2026-05-01/signal-audit-plateau-params.csv; cmd/signal-audit updated to plateau params (MACD fast=17, SMA slow=20, Donchian period=10); sensitivity concerns confirmed for RSI/Bollinger/Momentum | archived to 2026-05.md |
| 2026-05-01 | TASK-0052 | status → todo | unblocked by TASK-0051; moved from Blocked to Up Next; plateau-midpoint params available in runs/baseline-2026-04-30/plateau-params.json | |
| 2026-05-01 | TASK-0060 | status → done | --commission flag wired into cmd/universe-sweep; ParseCommissionModel called at startup with Fatalf on invalid value; parseDateRangeAndTimeframe extracted to fix cyclop limit; golangci-lint clean, all tests pass | archived to 2026-05.md |
| 2026-05-01 | TASK-0064 | status → done | RunConfig struct in internal/output; jsonResult embedding for top-level JSON merge; DefaultOutPath in internal/cmdutil; cmd/backtest wired + auto-out; cmd/sweep + cmd/universe-sweep log run config at startup; quality gate PASS, all tests pass, 0 lint issues | archived to 2026-05.md |
| 2026-05-02 16:25 | TASK-0065 | status → done | audit run: avg=35.3 trades, 0/15 COVID violations, PROCEED recorded | |
| 2026-05-03 | TASK-0052 | status → done | Universe sweep complete: runs/universe-sweep-2026-05-03.csv (90 rows). Survivors: macd-crossover (DSRAvg=0.2715, 14 eligible instruments), sma-crossover (DSRAvg=0.0969, 12 eligible instruments). Killed: donchian-breakout (DSRAvg=-0.1194), rsi-mean-reversion (0 sufficient), bollinger-mean-reversion (0 sufficient), momentum (0 sufficient). Kill decisions + survivor metrics recorded in decisions/algorithm/. Regime gate deferred. | archived to 2026-05.md |
| 2026-05-03 | TASK-0053 | status → todo (unblocked) | TASK-0052 complete; moved from Blocked to Up Next; survivor handoff JSON written to Notes; 14 instruments eligible for MACD walk-forward, 12 for SMA | |
| 2026-05-03 | TASK-0052 | notes updated | CCI mean-reversion (7th candidate, post-hoc) evaluated and killed at universe gate: DSRAvg=-0.0960 (fails >0), PassFraction=0.750, SufficientInstrumentCount=12. Kill decision: decisions/algorithm/2026-05-03-cci-mean-reversion-universe-gate-failed.md. CCI rows appended to runs/universe-sweep-2026-05-03.csv. TASK-0052 now fully complete with all 7 strategies. TASK-0053 survivor list unchanged. | archived 2026-05.md updated |
| 2026-05-03 | TASK-0066 | created | priority: high, source: session | Build cmd/walk-forward CLI entrypoint — wires internal/walkforward to a runnable binary; unblocks TASK-0053 |
| 2026-05-03 | TASK-0053 | status → blocked | blocked by TASK-0066 (cmd/walk-forward CLI does not exist); moved from Up Next to Blocked | |
| 2026-05-03 | TASK-0066 | status → done | cmd/walk-forward CLI complete. Factory dispatch table, run() extraction, 73.4% cmd coverage, 88% walkforward coverage. Quality gate PASS. | archived to 2026-05.md |
| 2026-05-03 | TASK-0053 | status → todo (unblocked) | TASK-0066 complete; moved from Blocked to Up Next |
| 2026-05-04 | TASK-0053 | status → done | Walk-forward ran on 26 instrument×strategy pairs (14 MACD, 12 SMA). Both strategies killed at instrument-count gate: MACD 9/14, SMA 4/12. 0 survivors. Kill records in decisions/algorithm/. | archived to 2026-05.md |
| 2026-05-04 | TASK-0054 | notes updated | Pipeline terminated — 0 survivors from TASK-0053. Remains blocked pending user decision: (A) relax instrument-count gate, (B) revisit parameters, or (C) start fresh. Handoff JSON appended to Notes. | |
| 2026-05-04 | TASK-0067 | created | priority: high, source: session | Update SMA --fast-period default 10→20 in cmd/universe-sweep + cmd/walk-forward; prepare evaluation re-run pipeline. Quality gate PASS. |
| 2026-05-04 | TASK-0068 | created | priority: high, source: session | Run SMA universe sweep + walk-forward at fast=20/slow=50; blocked only on Zerodha token |
| 2026-05-04 | TASK-0069 | created | priority: high, source: session | Reconsider MACD instrument-count gate threshold (100% retention too strict?); blocked by TASK-0068 |
| 2026-05-04 | TASK-0054 | notes updated | Blocker note updated: TASK-0068 (SMA re-run) and TASK-0069 (MACD gate review) are active remediation paths | |
| 2026-05-04 | TASK-0068 | status → done | Universe gate failed: all 15 instruments InsufficientData=true (trade_count 12–20 < 30 minimum). fast=20/slow=50 generates ~2–3 trades/year — statistically infeasible. SMA crossover killed definitively. Kill decision: decisions/algorithm/2026-05-04-sma-crossover-fast20-slow50-universe-gate-failed.md. Results: runs/universe-sweep-sma-20-50-2026-05-04.csv. | archived to 2026-05.md |
| 2026-05-04 | TASK-0069 | status → todo (unblocked) | TASK-0068 complete; blocker removed. Moved from Blocked to Up Next. No SMA survivors to affect gate-design precedent — MACD gate-design review proceeds independently. | |
| 2026-05-04 | TASK-0070 | created | priority: high, source: session | cmd/fetch-history CLI — bulk intraday historical data fetcher; leverages existing chunking; incremental delta fetch; writes to CachedProvider disk cache | owner: Priya |
| 2026-05-04 | TASK-0071 | created | priority: high, source: session | Verify overnight gap handling for intraday CNC backtests — golden tests for P&L and stop-loss fill across 17hr session gap | owner: Priya; blocker for all intraday backtest validity |
| 2026-05-04 | TASK-0072 | created | priority: high, source: session | Nifty Midcap 150 universe YAML — 20-30 instruments, continuous history from 2018, Marcus reviews list | owner: Priya builds, Marcus reviews |
| 2026-05-04 | TASK-0073 | created | priority: medium, source: session | cmd/evaluate — end-to-end automated evaluation pipeline (sweep → walk-forward → bootstrap); DSR-enforced param search on training window only | owner: Priya |
| 2026-05-04 | TASK-0074 | created | priority: medium, source: session | Opening Range Breakout strategy (5-min, CNC overnight hold) | blocked: Marcus must define entry/exit rules |
| 2026-05-04 | TASK-0075 | created | priority: medium, source: session | Gap-and-go strategy (5-min, CNC overnight hold) | blocked: Marcus must define entry/exit rules |
| 2026-05-04 | TASK-0076 | created | priority: low, source: session | Add Timeframe30Min + Timeframe60Min to model and Zerodha provider | owner: Priya |
| 2026-05-04 | TASK-0077 | created | priority: low, source: session | cmd/param-search — parameter optimization with DSR correction; training window only, OOS untouched | blocked: TASK-0073; owner: Priya |
| 2026-05-04 | TASK-0059 | reprioritized | medium → high | All planned intraday strategies use TimedExit wrapper — factory API is now a blocker for intraday walk-forward | |
| 2026-05-04 | TASK-0046 | blocker updated | Phase 2 sequencing → MIS-only blocker; CNC 2-3 day intraday focus does not require forced session close; TASK-0046 deferred until MIS strategy work begins | |
| 2026-05-04 | TASK-0062 | duplicate removed | First TASK-0062 entry ("TRI CSV loader") removed from backlog — superseded by the second, more detailed TASK-0062 entry ("TRI benchmark + StaticCSVProvider") | |
| 2026-05-05 | TASK-0059 | duplicate removed | Stale medium-priority copy in Todo (Backlog) removed; canonical high-priority entry remains in Up Next | |
| 2026-05-05 | TASK-0077 | moved | Todo (Backlog) → Blocked section; status was blocked but placed in wrong section | |
| 2026-05-05 | TASK-0070 | AC patched | Added: auth flags / env vars; partial-failure manifest recovery; incremental mode depends on TASK-0080 | Priya review |
| 2026-05-05 | TASK-0071 | AC patched | Stop-loss golden test replaced with correct criterion: engine fill at gap-down open, not signal price | Priya review |
| 2026-05-05 | TASK-0073 | AC patched | Removed --param-sweep flag (belongs in TASK-0077); added zero-survivor halt criterion | Priya review |
| 2026-05-05 | TASK-0074 | AC patched | Added Marcus long-only/bidirectional ruling as pre-implementation gate; ORB session detection now references TASK-0078 | Priya review |
| 2026-05-05 | TASK-0076 | AC patched | Added chunk_test.go update criterion for 30-min and 60-min cases | Priya review |
| 2026-05-05 | TASK-0078 | created | priority: high, source: session | Session-boundary utilities (IsSessionOpen, PreviousSessionClose) in pkg/strategy/; unblocks TASK-0074 + TASK-0075 | owner: Priya |
| 2026-05-05 | TASK-0079 | created | priority: medium, source: discovery | Tech debt: centralized strategy registry in internal/cmdutil; eliminates 4+ CLI manual registrations per new strategy | owner: Priya |
| 2026-05-05 | TASK-0080 | created | priority: medium, source: discovery | Tech debt: CachedProvider incremental manifest (LastCachedTime, RecordFetch, atomic writes); unblocks TASK-0070 incremental mode | owner: Priya |
| 2026-05-05 | TASK-0069 | status → done | Bootstrap gate complete: 4 survivors (SBIN, BAJFINANCE, TITAN, ICICIBANK), 5 killed. Decision: decisions/algorithm/2026-05-05-macd-bootstrap-gate-results.md. TASK-0054 unblocked. | evaluation-run |
| 2026-05-05 | TASK-0054 | status → done | Bootstrap completed under TASK-0069 remediation. 4 MACD survivors: SBIN, BAJFINANCE, TITAN, ICICIBANK. Decision: decisions/algorithm/2026-05-05-macd-bootstrap-gate-results.md |
| 2026-05-05 | TASK-0055 | status → todo (unblocked) | TASK-0054 complete; MACD has 4 bootstrap survivors. Correlation and portfolio construction can proceed. |
| 2026-05-05 | TASK-0081 | created | priority: high, source: session | zerodha.NewProvider token required even when all data cached; blocks automated eval runs. Includes chunk-completeness validation with ErrIncompleteData typed error |
| 2026-05-05 | TASK-0082 | created | priority: medium, source: session | cmd/backtest --bootstrap missing distribution stats in JSON output; evaluation pipeline had to parse stdout |
| 2026-05-05 | TASK-0069 | archived | moved to tasks/archive/2026-05.md |
| 2026-05-05 | TASK-0054 | archived | moved to tasks/archive/2026-05.md |
| 2026-05-05 | TASK-0055 | moved to Up Next | unblocked; TASK-0054 complete |
| 2026-05-05 | BACKLOG | reordered | done tasks archived; sections priority-sorted; TASK-0082 moved to correct medium-priority slot |
| 2026-05-05 | TASK-0081 | status → done | All 6 acceptance criteria met: instruments CSV cache, skip-network on cache hit, stale-cache fetch, ErrIncompleteData typed error (90% threshold), CachedProvider unchanged, lint+race PASS. Coverage 89.7%. | archived to tasks/archive/2026-05.md |
| 2026-05-05 | TASK-0083 | created | priority: medium, source: session | Tech debt: handle *ErrIncompleteData at cmd/ layer boundary (universe-sweep, backtest, walk-forward); typed error propagated as generic today; exit code 2 convention |
| 2026-05-06 | TASK-0082 | status → done | All 6 acceptance criteria met: BootstrapStats DTO added to internal/output, *BootstrapStats field under "bootstrap" key in jsonResult, 5 TDD tests, lint+race PASS (85.8% coverage). Fields absent when bootstrap not run (omitempty pointer). | archived to tasks/archive/2026-05.md |
| 2026-05-06 | TASK-0084 | created | priority: low, source: session | Tooling: update evaluation-run agent to read bootstrap stats from JSON "bootstrap" key instead of parsing stdout; fragility reduction after TASK-0082 |
| 2026-05-06 | TASK-0055 | status → in-progress | Marcus GO verdict from strategy-evaluator evaluate session; portfolio construction proceeding with correlation and regime gate analysis |
| 2026-05-06 | TASK-0085 | created | priority: high, source: decision | Correlation gate: run pairwise Pearson r on all 6 MACD survivor pairs (SBIN, BAJFINANCE, TITAN, ICICIBANK); full-period + stress-period; Marcus expects SBIN/ICICIBANK to fail | Marcus evaluate session 2026-05-06 |
| 2026-05-06 | TASK-0086 | created | priority: high, source: decision | Regime gate: compute per-regime Sharpe contributions for MACD survivors (deferred from TASK-0052); three regime windows; not a kill condition — half-weight on flag | Marcus evaluate session 2026-05-06 |
| 2026-05-06 | TASK-0087 | created | priority: high, source: decision | Portfolio composition: record final portfolio, sizing, kill-switch thresholds in decisions/algorithm/; blocked by TASK-0085 and TASK-0086 | Marcus evaluate session 2026-05-06 |
| 2026-05-06 | TASK-0056 | blocker updated | Blocked by TASK-0087 (portfolio composition must be written first) + TASK-0048; was blocked by TASK-0055 directly |
| 2026-05-07 | TASK-0085 | status → done | Correlation gate completed — SBIN + TITAN survivors | Results in decisions/algorithm/2026-05-06-macd-correlation-gate-results-sbin-titan-survivors.md |
