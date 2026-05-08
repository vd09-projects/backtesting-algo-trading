# Sweep-only strategy registration pattern — rejected in favour of full cmdutil centralization

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-08       |
| Status   | rejected         |
| Category | architecture     |
| Tags     | strategy, registry, cmd/sweep, registration, DRY, rejected |

## Context

After noticing that `cci-mean-reversion` was missing from `cmd/sweep`'s `factoryRegistry` switch despite being registered in `GlobalRegistry`, a plan was drafted to consolidate `cmd/sweep`'s strategy construction into a `strategyRegistration` slice with `paramDef` entries in a dedicated `strategies.go` file. The scope was initially cmd/sweep only.

## Decision

Rejected mid-session. The user observed that the same duplication exists in backtest, universe-sweep, and walk-forward — adding a strategy to cmd/sweep alone still required N separate changes across all cmd mains. The narrow fix would have solved sweep but left the broader problem intact.

Superseded by full centralization: `StrategyEntry.Build` and `StrategyEntry.SweepParams` in `internal/cmdutil/strategies.go` eliminates the per-cmd duplication in every affected binary simultaneously. See `2026-05-08-strategy-wiring-fully-centralized-in-cmdutil`.

## Related decisions

- [Strategy wiring fully centralized in internal/cmdutil](../architecture/2026-05-08-strategy-wiring-fully-centralized-in-cmdutil.md) — the approach that superseded this one
