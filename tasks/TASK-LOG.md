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
| 2026-05-07 | TASK-0086 | status → done | All 6 acceptance criteria met. SBIN: RegimeConcentrated=false (max 41.84% COVID+recovery). TITAN: RegimeConcentrated=false (max 45.97% pre-COVID). No allocation adjustment. Results in decisions/algorithm/2026-05-07-regime-gate-results-sbin-titan-macd-task0086.md | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0087 | status → todo (unblocked) | TASK-0085 and TASK-0086 both done; moved from Blocked to Up Next | |
| 2026-05-07 | TASK-0055 | notes updated | TASK-0085 done (SBIN+TITAN survive correlation gate), TASK-0086 done (both RegimeConcentrated=false); blocked on TASK-0087 only | |
| 2026-05-07 | TASK-0048 | status → done | cmd/monitor built: JSON trade log + thresholds JSON + synthetic equity curve + CheckKillSwitch wiring; 13 tests pass; 85.9% coverage; 0 lint issues; decision record decisions/convention/2026-05-07-live-trade-log-json-array-format.md | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0056 | blocker updated | TASK-0048 resolved 2026-05-07; still blocked by TASK-0087 only | |
| 2026-05-07 | TASK-0088 | created | priority: low, source: discovery | cmd/monitor test cleanup: TestRun_InvalidThresholdsJSON + Trade.Instrument in TestBuildSyntheticCurve_Order fixtures |
| 2026-05-07 | TASK-0087 | status → done | portfolio composition file written: decisions/algorithm/2026-05-07-macd-portfolio-composition.md; both instruments RegimeConcentrated=false | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0055 | status → done | MACD portfolio construction complete: NSE:SBIN + NSE:TITAN survivors; correlation gate, regime gate, composition file all done | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0056 | status → done | pre-live brief complete; NSE:SBIN APPROVED (p5=0.0719, 98.0%), NSE:TITAN APPROVED (p5=0.0854, 98.7%); threshold JSONs written; portfolio cleared for live | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0089 | created | priority: medium, source: user | CachedProvider range-aware lookup: serve subset from disk, skip network call when superset cached |
| 2026-05-07 | TASK-0089 | status → done | All 9 acceptance criteria met; 90.8% coverage, 0 lint issues, TDD; archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0090 | created | priority: low, source: discovery | CachedProvider: add TestSupersetHit_CorruptSupersetFallback to cover corrupt-superset-file fallback path |
| 2026-05-07 | TASK-0090 | updated | notes: added lazy auth session context (lazyProvider + context.Background() fix in internal/cmdutil/cmdutil.go; decisions recorded) | no status change |
| 2026-05-07 | TASK-0076 | updated | AC: added lazyProvider.SupportedTimeframes() update requirement in internal/cmdutil/cmdutil.go — maintenance trap from lazy auth fix | no status change |
| 2026-05-07 | TASK-0072 | status → done | universes/nifty-midcap-liquid.yaml created; 25 instruments across 9 sectors; Marcus approved; CLI parse verified; archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0091 | created | priority: high, source: session | Eval run: Nifty Midcap 150 universe sweep (MACD crossover + other survivors) — unblocked by TASK-0072 |
| 2026-05-07 | TASK-0059 | status → done | All 7 acceptance criteria met. Run() accepts factory func() strategy.Strategy; runFold() calls factory() twice per fold (IS+OOS); TestRun_TimedExitFoldStateIsolation added; pre-existing gofumpt in output_test.go fixed. 18 tests, 87.9% coverage, 0 lint issues, race clean. | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0071 | status → done | Engine gap handling verified correct. Both golden tests passed without any engine change — pendingSignal fills at candles[i].Open with no clamping; overnight gaps reflected in CNC P&L. Decision: decisions/convention/2026-05-07-overnight-gap-fill-confirmed-correct.md. New test file: internal/engine/engine_gap_test.go. Quality gate PASS: 0 lint issues, 98.1% engine coverage. | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0074 | notes updated | TASK-0071 (gap handling verified) marked done 2026-05-07; TASK-0059 (walk-forward factory API) also done; remaining blocker is Marcus design + TASK-0078 only | no status change |
| 2026-05-07 | TASK-0075 | notes updated | TASK-0071 (gap handling verified) done 2026-05-07; engine confirmed gap-transparent — gap-and-go strategy will see realistic gap P&L; remaining blocker is Marcus design + TASK-0078 | no status change |
| 2026-05-07 | TASK-0078 | status → done | All 7 AC met. IsSessionOpen + PreviousSessionClose implemented in pkg/strategy/session.go. 11 tests, 100% coverage, 0 lint issues, race-clean. TDD confirmed. Quality gate PASS. | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0074 | notes updated | TASK-0078 (session-boundary utilities) done 2026-05-07 — all infra deps resolved; blocked solely on Marcus rules now | no status change |
| 2026-05-07 | TASK-0075 | notes updated | TASK-0078 (session-boundary utilities) done 2026-05-07 — all infra deps resolved; blocked solely on Marcus rules now | no status change |
| 2026-05-07 | TASK-0079 | status → done | All 7 AC met. StrategyRegistry + GlobalRegistry in internal/cmdutil; 4 cmd binaries migrated; local strategy dispatchers removed; cci-mean-reversion added to cmd/backtest. 5 new tests, 100% registry coverage, 0 lint issues, race-clean. TDD confirmed. Quality gate PASS. | archived to tasks/archive/2026-05.md |
| 2026-05-07 | TASK-0061 | notes updated | Post-TASK-0079 scope: add cci-mean-reversion to cmd/sweep factoryRegistry; cmd/sweep2d should consume GlobalRegistry; cosmetic ValidateStrategyName() helper noted | no status change |
| 2026-05-08 | TASK-0092 | created | priority: medium, source: session — TestSignalAuditCoversAllStrategies; enforces signal-audit manually-maintained allStrategyFactories() covers all GlobalRegistry strategies | harvested from strategy wiring centralization session |
| 2026-05-08 | TASK-0061 | notes updated | Trimmed resolved items (fixedParams duplication gone, MustGet discard documented as intentional); remaining scope is sweep2d factoryRegistry2D extension + GlobalRegistry name validation hookup | strategy wiring centralization 2026-05-08 |
| 2026-05-08 | TASK-0063 | status → done | cmd/backtest doc comment now lists all 7 strategies with flag groups; AC met as part of strategy wiring centralization refactor | archived to tasks/archive/2026-05.md |
| 2026-05-09 | TASK-0070 | status → done | All 9 acceptance criteria met. cmd/fetch-history/main.go + main_test.go; 12 TDD tests; quality gate PASS round 2 (79.5% coverage, 0 lint). providerFactory func(fetchFlags) pattern; atomic fetch-progress.json manifest; dry-run mode. TASK-0083 ErrIncompleteData handling tracked separately. | archived to tasks/archive/2026-05.md |
| 2026-05-09 | TASK-0093 | created | priority: low, source: discovery | cmd/fetch-history: os.MkdirAll guard in fetchAll + 2 missing parseFlags tests (invalid timeframe, missing access-token). Surfaced by multi-perspective review. |
| 2026-05-09 | TASK-0094 | created | priority: low, source: discovery | cmd/fetch-history/fetchOne: refactor 11-parameter signature to fetchState struct. Naming & Clarity review suggestion from TASK-0070. |
| 2026-05-09 | TASK-0095 | created | priority: high, source: session | Nifty Midcap 150 walk-forward for macd-crossover on 43 eligible instruments; direct pipeline next step after TASK-0091 universe sweep. Gate: >= 60% retention (revised 2026-05-05). |
| 2026-05-09 | TASK-0091 | status → done | All 6 acceptance criteria met. Universe sweep: macd-crossover (17/26/9) across 48 nifty-midcap instruments 2018-2024. DSR avg Sharpe=0.0885 (nTrials=48), pass fraction=89.6% (43/48), all instruments sufficient. ABCAPITAL: no replacement needed (Sharpe=0.655, 53 trades). Marginal-ADV (BHEL/SCHAEFFLER/EXIDEIND/NATIONALUM/SAIL): all positive, all sufficient. Gate PASS. Advances to walk-forward on 43 instruments. Decisions recorded: macd-crossover-midcap-universe-gate-passed.md, ntrials-48-for-midcap-dsr-correction.md, marginal-adv-instruments-included-midcap-sweep.md. No task unblocked (no task was blocked by TASK-0091). Survivor annotation stored in archive entry. | archived to tasks/archive/2026-05.md |
| 2026-05-09 | TASK-0097 | status → done | All 4 acceptance criteria met. Correlation screen: 0/15 pairs exceed 0.70 full-period; 0/15 exceed 0.60 in stress windows. Kill-switch thresholds derived at 1.5x historical MaxDD per instrument (PERSISTENT 5.62%, TORNTPHARM 2.67%, COFORGE 12.57% hard/8% early-warning, SUNDARMFIN 6.96%, INDHOTEL 8.70%, MUTHOOTFIN 6.27%). Portfolio: vol-target 10% annualized, Rs50k base notional, IT cap (PERSISTENT+COFORGE <=Rs40k each when co-deployed). Decision: decisions/algorithm/2026-05-09-macd-crossover-midcap-correlation-killswitch-portfolio.md. Pipeline complete for MACD crossover Nifty Midcap 150. | archived to tasks/archive/2026-05.md |
| 2026-05-10 | TASK-0098 | created | priority: high, source: session | Strategy PriceExit wrapper: fixed stop-loss and target-profit for intraday strategies. Marcus ruling 2026-05-10: SL/TP opt-in for ORB/Gap-and-Go only, not MACD. Follows TimedExit pattern in pkg/strategy/. Unblocks TASK-0074 build, TASK-0075 build. |
| 2026-05-10 | TASK-0099 | created | priority: high, source: session | Data: fetch and validate 5-min bar history for Nifty50 large-cap universe. cmd/fetch-history supports 5min but cache unpopulated. Available window ~3yr vs 6yr daily — affects fold count and gate thresholds. Operational task, no code changes. Unblocks TASK-0074 eval, TASK-0075 eval. |
| 2026-05-10 | TASK-0074 | updated | notes | Added TASK-0098 (PriceExit) and TASK-0099 (5-min data) as pre-build requirements alongside existing Marcus rules blocker. |
| 2026-05-10 | TASK-0075 | updated | notes | Added TASK-0098 (PriceExit) and TASK-0099 (5-min data) as pre-build requirements alongside existing Marcus rules blocker. |
| 2026-05-10 | TASK-0099 | updated | scope expanded | Now covers both nifty50-large-cap and nifty-midcap-liquid universes in parallel. Both fetches run simultaneously; single decision file with two sections. Midcap fetch ~3x longer (48 vs 15 instruments). |
| 2026-05-10 | TASK-0099 | status → done | All AC met. Large-cap: 13/15 succeeded (HDFCBANK, ICICIBANK failed — OHLC validation error, flagged for Marcus). Midcap: 40/48 succeeded (8 failed same error). Bar count sanity PASS (99.5%). Session boundaries PASS. Decision file written: decisions/algorithm/2026-05-10-5min-data-coverage-both-universes.md. Kite 5-min window is 5+ years (longer than ~3yr estimate). Archived to tasks/archive/2026-05.md. |
| 2026-05-10 | TASK-0074 | updated | notes | Removed TASK-0099 from blocker list (done). Updated to reflect 13/15 large-cap 5-min instruments available. Blocked solely on Marcus rules + TASK-0098. |
| 2026-05-10 | TASK-0075 | updated | notes | Removed TASK-0099 from blocker list (done). Updated to reflect 40/48 midcap 5-min instruments available. Blocked solely on Marcus rules + TASK-0098. |
| 2026-05-10 | TASK-0100 | created | priority: high, source: bug | OHLC bad-candle skip in cmd/fetch-history — 10 instruments failing on candle[450] Zerodha artifact |
| 2026-05-10 | TASK-0100 | status → done | All 10 acceptance criteria met. ErrBadCandles typed warning; parseKiteCandles skips OHLC-invalid candles; CachedProvider caches non-empty results only; fetchOne all-bad path returns error. 9 TDD tests. Quality gate PASS. Perspective review: 2 blockers found and fixed (all-bad treated as success + empty cache write). | archived to tasks/archive/2026-05.md |
| 2026-05-10 | TASK-0101 | created | priority: high, source: session | Re-run cmd/fetch-history against 10 previously failing 5-min instruments; TASK-0100 fix should allow all to succeed |
| 2026-05-10 | TASK-0101 | status → done | All 5 acceptance criteria met. HDFCBANK 98,905 bars (1 skipped), ICICIBANK 98,905 bars (1 skipped), 8 midcap instruments 98,892–98,920 bars (1 skipped each). All at 99.5% completeness, all pass 89,438 threshold. Coverage decision file updated: large-cap 15/15, midcap 48/48. | archived to tasks/archive/2026-05.md |
| 2026-05-10 | TASK-0074 | updated | notes | 5-min cache now fully complete (TASK-0101 done): 15/15 large-cap available including HDFCBANK and ICICIBANK. |
| 2026-05-10 | TASK-0075 | updated | notes | 5-min cache now fully complete (TASK-0101 done): 48/48 midcap available including all 8 previously failing instruments. |
| 2026-05-10 | TASK-0073 | status → done | All 7 acceptance criteria met. cmd/evaluate/main.go (838 lines) + main_test.go (992 lines). 28 tests. Quality gate PASS (65% coverage, integration-only gap). Perspective review: APPROVE (0 blocking). | archived to tasks/archive/2026-05.md |
| 2026-05-10 | TASK-0077 | status: blocked → todo | Blocker TASK-0073 is now done. TASK-0077 moved from Blocked to Todo (Backlog). | unblocked |
| 2026-05-10 | TASK-0102 | created | priority: low, source: discovery | cmd/evaluate: wire universesweep.Result.Trades to skip bootstrap engine re-run (Tech Debt Sentinel finding from TASK-0073 review) |
| 2026-05-10 | TASK-0103 | created | priority: low, source: discovery | cmd/evaluate/makeStageDir: use os.MkdirAll for retry-after-error (Error Handling Inspector finding from TASK-0073 review) |
| 2026-05-10 | TASK-0104 | created | priority: low, source: discovery | cmd/evaluate/applyWFGate: use gateResult.PositiveSharpeInstruments for decision-document traceability (Domain Logic Reviewer finding from TASK-0073 review) |
| 2026-05-11 | TASK-0098 | status → done | All 10 acceptance criteria met. pkg/strategy/price_exit.go + price_exit_test.go. 7 golden tests + 3 metadata tests. 100% coverage, 0 lint issues, race clean. TDD confirmed (compile-fail on test file verified). Perspective review: APPROVE (0 blocking). | archived to tasks/archive/2026-05.md |
| 2026-05-11 | TASK-0074 | notes updated | TASK-0098 (PriceExit) done 2026-05-11; removed from blocker list; blocked solely on Marcus rules now | no status change |
| 2026-05-11 | TASK-0075 | notes updated | TASK-0098 (PriceExit) done 2026-05-11; removed from blocker list; blocked solely on Marcus rules now | no status change |
| 2026-05-11 | TASK-0105 | created | priority: low, source: discovery | pkg/strategy/price_exit.go NewPriceExit godoc: add note that negative pct values are treated as disabled (same as 0); Tech Debt Sentinel finding from TASK-0098 perspective review |
| 2026-05-11 | TASK-0106 | created | priority: low, source: discovery | internal/walkforward: add TestRun_PriceExitFoldStateIsolation analogous to TestRun_TimedExitFoldStateIsolation; Concurrency & State Safety finding from TASK-0098 perspective review |
| 2026-05-13 | TASK-0092 | status → done | all 4 AC met; single test added to cmd/signal-audit/main_test.go; quality gate PASS | archived to tasks/archive/2026-05.md |
| 2026-05-13 | TASK-0083 | status → done | all 5 AC met; Option B: per-instrument warning for universe-sweep, exit code 2 for backtest/walk-forward; HandleIncompleteDataError + ExitCodeError extracted to internal/cmdutil; providerFactory injection added; quality gate PASS | archived to tasks/archive/2026-05.md |
| 2026-05-13 | TASK-0107 | created | priority: low, source: discovery | internal/universesweep.Run: stderr writer shared across goroutines not synchronized; bytes.Buffer not goroutine-safe; syncWriter wrapper or contract doc needed; Concurrency reviewer finding from TASK-0083 |
| 2026-05-13 | TASK-0108 | created | priority: low, source: discovery | cmd/universe-sweep TestRun_IncompleteData_SweepContinues missing insufficient_data=true CSV assertion; Test Coverage Auditor finding from TASK-0083 |
| 2026-05-13 | TASK-0109 | created | priority: medium, source: session | cmd/fetch-history: wire HandleIncompleteDataError in fetchOne; explicitly deferred from TASK-0083; TASK-0070 done without it |
| 2026-05-13 | TASK-0058 | status → done | completed in commit d1d543f (2026-04-27) as part of quality-gate fix batch; all AC met; countRSISignals extraction reduced complexity 17 → clean; task was not marked done in backlog until this session | archived to tasks/archive/2026-04.md |
| 2026-05-13 | TASK-0077 | status → done | all 7 AC met; internal/paramsearch + cmd/param-search built; DSR aggregation matches ApplyUniverseGate; no OOS flags (architectural enforcement); 15 tests; quality gate PASS; perspective review APPROVE (0 blockers) | archived to tasks/archive/2026-05.md |
| 2026-05-13 | TASK-0110 | created | priority: low, source: discovery | cmd/param-search buildSearchPipeline: replace inline time.Parse + After-check with cmdutil.ParseDateRange; 9th copy of pattern flagged in patterns.md; Tech Debt Sentinel finding from TASK-0077 perspective review |
| 2026-05-13 | TASK-0111 | created | priority: low, source: discovery | internal/paramsearch TestDSRRanking_InvertsRawOrder vacuously true: stubStrategy → 0 trades → InsufficientData=true → sort assertion body never fires; switch to tradingStub; Test Coverage Auditor finding from TASK-0077 perspective review |
| 2026-05-13 | TASK-0112 | created | priority: low, source: discovery | internal/paramsearch Config.StrategyFactory godoc: add read-only params contract ("params is shared read-only across concurrent goroutine calls; factory must not mutate or retain it"); Concurrency reviewer finding from TASK-0077 perspective review |
| 2026-05-13 | TASK-0110 | status → done | all 4 AC met; replaced 11-line inline time.Parse block with cmdutil.ParseDateRange in buildSearchPipeline; 15 tests pass, race clean, 82.4% coverage, 0 lint issues; behavior confirmed identical by Domain Logic Reviewer | archived to tasks/archive/2026-05.md |
| 2026-05-13 | TASK-0111 | status → done | all 3 AC met; TestDSRRanking_InvertsRawOrder renamed TestDSRRanking_SortedDescending, switched to tradingStub, 43-line comment block removed, fired boolean added with correct placement (inside outer !InsufficientData guard before inner comparison); go test -race PASS, golangci-lint clean, 92.9% coverage maintained | archived to tasks/archive/2026-05.md |
| 2026-05-13 | TASK-0074 | reprioritized | medium → high; all infrastructure complete (TASK-0059/0071/0078/0098/0099/0101 done); blocked solely on Marcus rules; ready to build immediately once Marcus unblocked | |
| 2026-05-13 | TASK-0075 | reprioritized | medium → high; all infrastructure complete (same infra as TASK-0074 done); blocked solely on Marcus rules; ready to build immediately once Marcus unblocked | |
| 2026-05-13 | BACKLOG | reordered | Up Next populated: TASK-0080 (medium, incremental manifest), TASK-0062 (medium, TRI benchmark), TASK-0109 (medium, fetch-history ErrIncompleteData). Todo ordered: quick wins (TASK-0103/0104/0105/0112) → small tests (TASK-0108/0088/0090/0093/0107/0106) → refactors (TASK-0094/0061/0102) → larger items (TASK-0076/0084/0036/0037/0057). TASK-0061 note: do after TASK-0074+0075 built. | |
| 2026-05-13 | TASK-0074 | status → todo | Marcus rules written to decisions/algorithm/2026-05-13-orb-marcus-rules.md; unblocked from Blocked section; moved to Up Next (high priority, ahead of TASK-0080). First two ACs checked off. | GO verdict from evaluation session 2026-05-13-evaluate-orb |
| 2026-05-13 | TASK-0113 | created | priority: high, source: decision; blocked by TASK-0074 | ORB signal frequency audit — 15 large-cap instruments, 5-min bars 2018-2023 |
| 2026-05-13 | TASK-0114 | created | priority: high, source: decision; blocked by TASK-0074 | ORB in-sample baseline + parameter sensitivity on RELIANCE |
| 2026-05-13 | TASK-0115 | created | priority: high, source: decision; blocked by TASK-0114 | ORB universe sweep — 15 large-cap instruments, DSR gate, nTrials=15 |
| 2026-05-13 | TASK-0116 | created | priority: high, source: decision; blocked by TASK-0115 | ORB walk-forward validation — 2yr IS / 1yr OOS / 1yr step, 2018-2024 |
| 2026-05-13 | TASK-0117 | created | priority: high, source: decision; blocked by TASK-0116 | ORB Monte Carlo bootstrap — 10,000 simulations, SharpeP5 kill-switch threshold |
| 2026-05-13 | TASK-0118 | created | priority: medium, source: decision; blocked by TASK-0117 | ORB pre-live brief — kill-switch thresholds, correlation gate, go/no-go sign-off |
| 2026-05-13 | TASK-0075 | status → todo | unblocked; Marcus rules written to decisions/algorithm/2026-05-13-gap-and-go-marcus-rules.md; moved from Blocked to Up Next | Entry: 2nd bar close 09:20 IST; gap=1.0%; vol=1.3×; SL=0.8×gapPct; TP=2.0×gapPct; time-stop=150 bars; long-only |
| 2026-05-13 | TASK-0119 | created | priority: high, source: decision; blocked by TASK-0075 | Gap-and-Go signal frequency audit — 48 midcap instruments, 1.0% gap threshold |
| 2026-05-13 | TASK-0120 | created | priority: high, source: decision; blocked by TASK-0075 | Gap-and-Go parameter sensitivity — NSE:INDHOTEL orientation, 81-variant sweep |
| 2026-05-13 | TASK-0121 | created | priority: high, source: decision; blocked by TASK-0120 | Gap-and-Go universe sweep — 48 midcap instruments, DSR gate nTrials=48 |
| 2026-05-13 | TASK-0122 | created | priority: high, source: decision; blocked by TASK-0121 | Gap-and-Go walk-forward — 2yr IS / 6mo OOS / anchored, 60% instrument retention floor |
| 2026-05-13 | TASK-0123 | created | priority: high, source: decision; blocked by TASK-0122 | Gap-and-Go Monte Carlo bootstrap — 10,000 simulations, SharpeP5 kill-switch threshold |
| 2026-05-13 | TASK-0124 | created | priority: high, source: decision; blocked by TASK-0123 | Gap-and-Go correlation gate — vs MACD midcap portfolio and ORB (highest-risk pair) |
| 2026-05-13 | TASK-0125 | created | priority: medium, source: decision; blocked by TASK-0124 | Gap-and-Go pre-live brief — kill-switch thresholds, Rs 1.5L position cap, go/no-go sign-off |
| 2026-05-17 | TASK-0126 | created | priority: high, source: session | Research spike: 5-min strategy canvas, 10+ candidates across edge buckets; output is ranked candidate list with Marcus pre-screen verdicts; feeds evaluation session tickets |
| 2026-05-17 | TASK-0127 | created | priority: high, source: session | Adapt existing daily-bar strategies (MACD, RSI, SMA, Bollinger, CCI, Momentum) to 5-min; recalibrate parameters; session-boundary decision per strategy; signal audit + universe gate |
| 2026-05-17 | TASK-0128 | created | priority: medium, source: session | Composite signal design on 5-min: MACD+VWAP, RSI+volume, SMA+session-timing; regime filters only, not oscillator stacking; blocked on TASK-0127 |
| 2026-05-17 | TASK-0129 | created | priority: medium, source: session | Empirical 1-min Kite data depth verification: test fetch from 2022, record actual available window, update chunk constants or drop 1-min if only 60 days available |
| 2026-05-17 | TASK-0076 | status → cancelled | strategy focus shifted to 1-min/5-min; 30/60-min bars move wrong direction; archived to tasks/archive/2026-05.md |
| 2026-05-17 | TASK-0084 | status → cancelled | stdout parsing still works; fragility reduction not worth effort at current priorities; archived |
| 2026-05-17 | TASK-0036 | status → cancelled | no results to visualize yet; revisit after ORB/Gap-and-Go complete evaluation pipeline; archived |
| 2026-05-17 | TASK-0037 | status → cancelled | SMA/RSI are dead strategies; bootstrap on dead strategy is graveyard maintenance; archived |
| 2026-05-17 | TASK-0057 | status → cancelled | live-trading concern, not backtesting; re-open at live deployment stage with explicit dep approval; archived |
| 2026-05-17 | BACKLOG | reordered | Added TASK-0126/0127/0128/0129 to Up Next; cancelled 5 stale tasks (TASK-0036/0037/0057/0076/0084); open count 37→36 |
| 2026-05-17 | TASK-0127 | updated | Notes: added multiple-variants instruction (2-3 named variants per strategy, e.g. macd-5min-fast/standard/slow); session-boundary decision requirement per strategy (trend-following=cross-session OK, mean-reversion=document explicitly) |
| 2026-05-17 | TASK-0128 | updated | Notes: added blocked on TASK-0131 (VWAP utility); added note to defer composite ticket creation until TASK-0127 survivors known |
| 2026-05-17 | TASK-0130 | created | priority: medium, source: discovery | internal/analytics: MinCurvePointsForMetrics→timeframe-aware function + NSERegimes5Min2021_2024; daily-specific issues discovered in 5-min audit |
| 2026-05-17 | TASK-0131 | created | priority: medium, source: session | pkg/strategy/vwap.go: session-aware VWAP utility; prerequisite for TASK-0128 composite signal work |
| 2026-05-17 | BACKLOG | updated | open count 36→38 |
