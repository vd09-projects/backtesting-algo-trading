---
name: cyclop factory table pattern
description: When a factory switch over 7+ strategies hits the cyclop complexity limit, refactor to a dispatch table (map[string]builderFunc) — keeps each builder self-contained and reduces strategyFactory to O(1) lookup
type: feedback
---

When cmd/*  strategy factory functions grow to 7+ cases, the flat switch trips the cyclop linter (max=15). The fix is a dispatch table: `map[string]strategyBuilder` where each entry is a named per-strategy builder function. The builder validates params eagerly and returns the closure. `strategyFactory` becomes a single map lookup.

**Why:** Flat switch with eager validation + closure return per case is verbose and hits cyclop. Table dispatch is cleaner, easier to extend, and cyclomatic complexity collapses to ~2.

**How to apply:** Any time a strategy factory switch in cmd/ approaches 7+ cases, pre-empt the cyclop finding by using a registry map from the start. Pattern: `type strategyBuilder func(tf model.Timeframe, params *strategyParams) (func() strategy.Strategy, error)`.
