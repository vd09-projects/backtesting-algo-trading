# Config.NewStrategy as factory per instrument instead of shared singleton

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-11       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | strategy, factory, state, correctness, universesweep, cmd/universe-sweep |

## Context

`universesweep.Config` originally held a single `strategy.Strategy` instance that was passed to every instrument's engine run. The bug surfaced during a multi-perspective code review of the cmd/ refactor: any strategy that carries mutable state (RSI running accumulator, SMA crossover buffer, MACD signal line) will bleed state from instrument N into instrument N+1 when a shared instance is used. Results are wrong and silent — the sweep CSV shows plausible numbers but they're contaminated across instruments. `cmd/evaluate` had the same bug independently in its universe-sweep stage.

## Decision

`Config.Strategy strategy.Strategy` was renamed to `Config.NewStrategy func() strategy.Strategy`. `runInstrument` calls `cfg.NewStrategy()` once per instrument, giving each engine run a fresh zero-state instance. Callers use `cmdutil.GlobalRegistry.WalkForwardFactory` to construct the factory — it validates params once at startup then returns a `func() strategy.Strategy` that constructs fresh instances on demand. This matches the existing walk-forward fold pattern, which already required per-fold strategy isolation for the same reason.

## Consequences

All three callers — `internal/universesweep/universesweep_test.go` (5 sites), `cmd/universe-sweep/main.go`, and `cmd/evaluate/main.go` — were updated in the same commit. A regression test `TestRun_FreshStrategyInstancePerInstrument` was added to `universesweep_test.go`; it uses `sync/atomic` to count factory calls under the concurrent errgroup fan-out and asserts the factory is called exactly once per instrument. The existing `toggleStrategy` test fake was confirmed stateless (uses `len(candles)%2`, not struct fields) and was kept as-is; the new regression test uses a factory counter instead to prove the call pattern.

## Related decisions

- [errgroup for universe fan-out, GOMAXPROCS ceiling](tradeoff/2026-04-22-errgroup-universe-fan-out-gomaxprocs-ceiling.md) — the fan-out pattern that makes shared-instance contamination possible
- [Walk-forward strategy factory per fold](convention/2026-05-07-walkforward-strategy-factory-per-fold.md) — the prior art that established the factory-per-run pattern

## Revisit trigger

If `strategy.Strategy` implementations are audited and confirmed stateless (read-only candle slice, no struct field mutation), the factory approach could be relaxed. Unlikely — RSI, SMA, MACD all carry running state by design.
