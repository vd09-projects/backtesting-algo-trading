# Task Log

Append-only record of all task operations. Newest entries at the bottom.

| Date | Task | Action | Details | Notes |
|------|------|--------|---------|-------|
| 2026-04-10 | TASK-0011 | status → done | All acceptance criteria met | cmd/backtest/main.go + strategies/stub/ |
| 2026-04-10 | TASK-0025 | status → in-progress | starting corporate action verification for Zerodha data | |
| 2026-04-10 | TASK-0025 | status → done | Kite day candles adjusted for splits/bonuses; decision recorded in decisions/infrastructure/ | TASK-0012 and TASK-0015 unblocked |
| 2026-04-01 00:00 | TASK-0001 | created | priority: critical, source: project | Initialize Go module and folder structure (original) |
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
