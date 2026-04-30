# signalaudit package uses StrategyFactory — no import of concrete strategy packages

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | signalaudit, strategy-factory, decoupling, cmd-layer, dependency-direction, TASK-0050 |

## Context

`internal/signalaudit` needed to instantiate all 6 strategy implementations to run the trade-count audit. The naive approach would be to import concrete strategy packages (`strategies/smacrossover`, `strategies/rsimeanrev`, etc.) directly. However, the dependency rules in this project flow inward toward `pkg/model/` — `internal/` packages must not import `strategies/`, which itself imports `pkg/`.

## Decision

`internal/signalaudit` does not import concrete strategy packages. It defines a `StrategyFactory func(name string, tf model.Timeframe) (strategy.Strategy, error)` type. The cmd layer (`cmd/signal-audit/main.go`) owns the factory closure: it imports the concrete strategy packages and wires each name to its constructor. `internal/signalaudit.Run()` receives the factory and calls it per (strategy, instrument) cell.

This keeps `internal/signalaudit` free of the `strategies/` dependency direction and matches the established pattern in `internal/sweep` and `internal/walkforward`.

## Consequences

- Every new strategy requires updating the factory closure in `cmd/signal-audit/main.go` — but this is the right place for that coupling (the cmd layer is the composition root).
- The factory pattern means each (strategy, instrument) pair gets a fresh instance, which is correct for stateful strategies.
- Slightly more indirection than a direct import, but it enforces the package boundary that prevents circular imports.

## Related decisions

- [walk-forward factory API for stateful strategy wrappers](../algorithm/2026-04-27-walkforward-factory-api-stateful-strategies.md) — same factory pattern motivation; TASK-0059 will apply it to `internal/walkforward`

## Revisit trigger

If `strategies/` is ever restructured under `internal/` or `pkg/`, re-evaluate whether the factory indirection is still necessary.
