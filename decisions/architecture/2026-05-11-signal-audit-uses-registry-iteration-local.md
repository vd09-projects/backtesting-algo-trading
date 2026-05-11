# signal-audit uses GlobalRegistry iteration + local auditParamOverrides

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-11       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | signal-audit, registry, package-boundary, plateau-midpoint, cmd/signal-audit |

## Context

`cmd/signal-audit` previously imported 7 concrete strategy packages directly (`strategies/bollinger`, `strategies/ccimeanrev`, `strategies/donchian`, `strategies/macd`, `strategies/momentum`, `strategies/rsimeanrev`, `strategies/smacrossover`) and constructed strategy instances manually in `allStrategyFactories`. This violated the project rule "Never reference a concrete strategy type across package boundaries." It also created a silent pipeline gap: when a new strategy was registered in `cmdutil.GlobalRegistry`, signal-audit would not automatically include it in the audit, breaking the "new strategy must complete full pipeline before evaluation runs" requirement.

The audit also uses plateau-midpoint params (e.g. sma-crossover slow=20, not the registry default 50) chosen per Marcus's signal-audit verdict to guarantee ≥30 trades per instrument on the Nifty50 large-cap universe. These can't come from the registry defaults, so a local override mechanism was needed.

## Decision

`allStrategyFactories` now iterates `cmdutil.GlobalRegistry.ListStrategies()` and constructs each factory via `GlobalRegistry.WalkForwardFactory(name, tf, auditParamOverrides[name])`. A local `auditParamOverrides map[string]map[string]float64` holds the plateau-midpoint params for each strategy, with decision-file references in its doc comment. A startup check warns and skips any registered strategy missing from `auditParamOverrides`, preventing silent exclusion when a new strategy is added. `signalaudit.Strategy` is a type alias (`type Strategy = strategy.Strategy`), so `WalkForwardFactory`'s return type is directly compatible with `signalaudit.StrategyFactory.New` — no adapter required. `stub` is intentionally omitted from `auditParamOverrides` as it is a test tool, not a real strategy.

## Consequences

The 7 concrete strategy imports were removed from `cmd/signal-audit/main.go`. New strategies registered in `strategies.go` will now automatically appear in signal-audit, provided the developer also adds an entry to `auditParamOverrides`. The startup warning surfaces this requirement at runtime if an entry is missing. The plateau-midpoint params are now in one place (`auditParamOverrides`) with a doc comment pointing to the relevant decision files — previously they were scattered across seven constructor calls with inline comments. `WalkForwardFactory` panics inside the returned factory on unexpected Build failure; this is acceptable for a diagnostic tool where a mid-audit panic is preferable to silently wrong results (params are validated at startup).

## Related decisions

- [Strategy wiring fully centralized in cmdutil](convention/2026-05-07-strategy-wiring-fully-centralized-in-cmdutil.md) — the GlobalRegistry pattern this decision leverages
- [signal-audit strategy factory decoupling](convention/2026-05-01-signalaudit-strategy-factory-decoupling.md) — prior decision on the factory interface shape

## Revisit trigger

If the registry gains a mechanism to store audit-specific param overrides (e.g. an `AuditParams` field on `StrategyEntry`), the local `auditParamOverrides` map could move there. Not currently worth the complexity — the local map is small and visible.
