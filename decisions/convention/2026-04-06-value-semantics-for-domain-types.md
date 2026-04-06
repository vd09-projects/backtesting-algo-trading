# Value semantics for domain model types (Candle, Config)

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-06       |
| Status   | accepted         |
| Category | convention       |
| Tags     | candle, model, value-semantics, pointer, gocritic, performance, convention |

## Context

During the pre-merge quality review, golangci-lint's `gocritic` linter flagged two types as `hugeParam` — structs passed by value that exceed a size threshold:

- `Candle` (96 bytes) — `Validate()` uses a value receiver
- `engine.Config` (112 bytes) — `New(cfg Config)` passes by value

The linter's suggestion was to pass/receive by pointer instead.

## Options considered

### Option A: Switch to pointer receivers/parameters
- **Pros**: Eliminates copying large structs on every call. gocritic is satisfied.
- **Cons**: `Candle` is stored in `[]model.Candle` slices and iterated constantly. Switching to `*Candle` would require either a `[]*Candle` slice (heap allocation per candle, cache-unfriendly) or calling methods on addressable slice elements (Go handles this automatically, but it changes the mental model). `Config` passed by pointer to `New()` would require callers to write `engine.New(&cfg)` which is slightly more awkward for a one-shot constructor. Risk of accidental mutation through the pointer.

### Option B: Suppress with nolint and establish a convention
- **Pros**: `Candle` stays value-typed throughout — consistent with how Go slices work best for numeric data. The engine's hot loop iterates `[]model.Candle` and value semantics are cache-friendly. `Config` is a constructor argument called once per backtest run, not in any hot path.
- **Cons**: gocritic warning is suppressed rather than heeded.

## Decision

Value semantics (Option B) for both, with `//nolint:gocritic` directives and inline comments explaining the intent.

**Convention established**: Domain model types in this codebase (`Candle`, `Trade`, `Position`, `Signal`, etc.) use value semantics. These types live in slices, are passed to strategies, and are recorded in trade logs. Pointer semantics would scatter heap allocations through the hot loop and make the code harder to reason about. Size is not a sufficient reason to switch — alignment and cache locality matter more for slice-heavy numerical code.

Future types should default to value semantics unless they hold resources (file handles, connections) or have identity (need reference equality).

## Consequences

- gocritic `hugeParam` will fire on any domain type over ~80 bytes if passed by value. Suppress with `//nolint:gocritic` and a comment explaining the intent.
- Benchmark results should be used to challenge this decision if the engine's candle processing loop shows unexpected GC pressure or copy overhead.

## Revisit trigger

If profiling shows that `Candle` copying is a measurable bottleneck in the engine hot loop (target: < 1ms for 10 years of daily candles). Run `go test -bench=. -benchmem ./internal/engine/` and check allocs/op.
