---
name: Strategy registry architecture (TASK-0079)
description: Central strategy registry now lives in internal/cmdutil; all cmd binaries validate via GlobalRegistry.MustGet; one file to edit when adding a new strategy
type: project
---

TASK-0079 (2026-05-07) built the centralized strategy registry. Key facts:

- `internal/cmdutil/registry.go`: `StrategyRegistry` type with `Register`, `MustGet` (panics on unknown name), `ListStrategies`
- `internal/cmdutil/strategies.go`: `GlobalRegistry = buildRegistry()` — the single file to edit when adding a new strategy
- 8 strategies registered: stub, bollinger-mean-reversion, cci-mean-reversion, donchian-breakout, macd-crossover, momentum, rsi-mean-reversion, sma-crossover
- All 4 cmd binaries (backtest, universe-sweep, walk-forward, sweep) now call `GlobalRegistry.MustGet(name)` for name validation
- `cmd/walk-forward` retains `localBuilders` map for per-fold zero-arg factory construction — must be kept in sync with GlobalRegistry manually
- `cmd/sweep` has latent inconsistency: MustGet accepts cci-mean-reversion but local switch has no case — TASK-0061 resolves this
- `cmd/backtest` now includes cci-mean-reversion (was previously missing)

**Why:** Before this, adding a strategy required updating 4+ files; forgetting any produced silent unavailability.

**How to apply:** Step 1.5c preflight now skips strategy registration check for TASK-0079 itself. All future tasks adding new strategies only need to update `internal/cmdutil/strategies.go` + the local `localBuilders` in `cmd/walk-forward/main.go` (until TASK-0061 resolves the walk-forward gap too).
