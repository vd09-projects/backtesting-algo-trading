# Synthetic equity curve built from trades, not a separate curve file

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | equity-curve, synthetic, trades, no-separate-input, cmd/monitor, TASK-0048 |

## Context

`analytics.CheckKillSwitch` requires a `[]model.EquityPoint` equity curve to compute the
drawdown depth and duration thresholds. The engine produces per-bar equity snapshots during a
backtest; the question is how `cmd/monitor` should obtain a live equity curve.

Three approaches were considered:

### Option A: Require a separate `--curve` CSV file (rejected for now)

The user provides the same equity curve CSV produced by `cmd/backtest --output-curve`. The monitor
reads it via `internal/output.LoadCurveCSV`.

- **Pros**: Mark-to-market accuracy — reflects open position value between trades, not just
  closed trade P&L.
- **Cons**: A live equity curve CSV does not exist for the live trading use case. The engine
  produces it during backtests; a live trader would need to manually export and maintain this file
  after each trading session. High user burden for weekly monitoring cadence. The weekly check
  does not need intrabar accuracy — closed-trade granularity is sufficient.

### Option B: Derive curve from portfolio cash balance (out of scope)

Track live portfolio cash balance in a separate accounting file, updated after each trade.

- **Pros**: Most accurate representation of live equity.
- **Cons**: Requires additional infrastructure not yet in scope. Out of scope for TASK-0048.

### Option C: Build synthetic curve from sorted trade list (chosen)

Sort the trades in the live trade log by `ExitTime`, then walk forward accumulating equity from
a configurable starting value:

```
equity[i] = initialEquity + sum(RealizedPnL[0..i])
curve[i] = EquityPoint{Timestamp: trade[i].ExitTime, Value: equity[i]}
```

- **Pros**: No additional input required beyond the trade log (already required). Simple,
  deterministic, testable. Sufficient granularity for weekly monitoring — CheckKillSwitch's
  drawdown and duration checks operate on any curve, not just per-bar curves.
- **Cons**: No intrabar visibility. A position that loses heavily mid-trade then recovers will
  not show the intrabar drawdown. This is acceptable at the current monitoring cadence.

## Decision

Build the equity curve synthetically from the trade log in `buildSyntheticCurve`. Trades are
sorted by `ExitTime` (input order is not guaranteed); equity accumulates from `--initial-equity`
(default ₹150,000 matching the MACD portfolio's per-instrument notional).

A `--curve` flag for a separate curve file is not added at this time — it can be added later
if intrabar accuracy is required. The tradeoff is documented here so the choice is visible.

## Consequences

- `cmd/monitor` requires only two inputs: `--trades` and `--thresholds`. Simple to cron-schedule.
- Drawdown depth and duration computed from closed-trade equity only. No intrabar drawdown signal.
- If the strategy carries open positions for extended periods, the equity curve will have gaps
  between the last close and now. `computeCurrentDrawdownDepth` returns 0 for curves with <2
  points — if all positions are open, no drawdown is reported. Acceptable for daily-bar strategies
  where trade frequency is low.

## Revisit trigger

If intraday strategies (5-min bars, multiple open positions) are deployed live, the synthetic
curve will miss intrabar peaks and troughs. At that point, add a `--curve` flag to accept an
external mark-to-market curve and use `internal/output.LoadCurveCSV` for it.
