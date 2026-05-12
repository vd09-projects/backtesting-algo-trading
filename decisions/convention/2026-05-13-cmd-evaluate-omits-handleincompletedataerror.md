# cmd/evaluate omits HandleIncompleteDataError at run() boundary — intentional

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | ErrIncompleteData, cmd/evaluate, WF, bootstrap, TASK-0083 |

## Context

TASK-0083 added `cmdutil.HandleIncompleteDataError` to `cmd/backtest` and `cmd/walk-forward`. The question arose whether `cmd/evaluate` — a three-stage pipeline orchestrator (universe sweep → walk-forward → bootstrap) — should also call `HandleIncompleteDataError` at its `run()` error boundary.

## Decision

`cmd/evaluate` does **not** call `HandleIncompleteDataError` at `run()`. The three pipeline stages have different semantics:

1. **Universe sweep stage** — `runUniverseSweep` calls `universesweep.Run(pl.ctx, &sweepCfg, pl.provider, stderr)`. The `internal/universesweep` package handles `*ErrIncompleteData` per-instrument internally: the diagnostic is logged, the result is flagged `InsufficientData=true`, and the sweep continues. `cmd/evaluate` gets this behavior for free without any code change.

2. **Walk-forward stage** — `walkforward.Run` calls the engine per fold. If a fold's `FetchCandles` returns `*ErrIncompleteData`, the error propagates as a hard failure for that instrument. This is correct: walk-forward requires a complete candle series to compute IS/OOS Sharpe ratios. Partial data on a WF fold is not a recoverable situation.

3. **Bootstrap stage** — `collectTrades` runs the engine over the full date range. Same argument as WF: incomplete data means unreliable trade statistics, which means unreliable bootstrap distribution. Hard failure is correct.

The task AC explicitly covers `cmd/universe-sweep`, `cmd/backtest`, `cmd/walk-forward` only. The omission for `cmd/evaluate` WF/bootstrap stages is intentional, not an oversight.

## Consequences

- `cmd/evaluate` has no special handling for `*ErrIncompleteData` at its outer `run()` boundary.
- If WF or bootstrap hits incomplete data, `cmd/evaluate` returns a wrapped error with the full chain: `"walk-forward NSE:TCS: walk-forward: engine: fetching candles: zerodha: incomplete data…"`. The message is verbose but accurate.
- Universe sweep instruments with incomplete data appear as `insufficient_data=true` in `universe-sweep.csv` and are excluded from the gate calculation — consistent with existing behavior.
