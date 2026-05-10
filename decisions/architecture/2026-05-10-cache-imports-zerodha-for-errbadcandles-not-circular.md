# cache package imports zerodha parent package for ErrBadCandles type-check — not circular

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | bad-candle, ErrBadCandles, package-dependency, not-circular, TASK-0100 |

## Context

`pkg/provider/zerodha/cache` is a sub-package of `pkg/provider/zerodha`. Before TASK-0100, `cache` was interface-coupled only — it imported `pkg/model` and `pkg/provider`, not `pkg/provider/zerodha`. After TASK-0100, `CachedProvider.FetchCandles` must type-check the inner provider's error for `*zerodha.ErrBadCandles`. To call `errors.As(err, &badCandles)` where `badCandles *zerodha.ErrBadCandles`, the `cache` package must import `pkg/provider/zerodha`.

## Is this circular?

`pkg/provider/zerodha` imports: `pkg/model`, `pkg/provider` (for the interface), standard library.
`pkg/provider/zerodha/cache` imports: `pkg/model`, `pkg/provider`, `pkg/provider/zerodha` (new).

There is no cycle: `zerodha` does not import `zerodha/cache`. In Go, a sub-package importing its parent is valid and not a circular dependency. Verified with `go build` — no import cycle error.

## Options considered

### Option A: Import zerodha directly (chosen)
- **Pros**: Direct, type-safe, explicit.
- **Cons**: `cache` now has a compile-time dependency on the Zerodha implementation, not just the abstract interface. If a second `DataProvider` is added, `cache` would need a second `errors.As` branch or a shared error type.

### Option B: Define a shared interface in `pkg/provider` that `ErrBadCandles` satisfies
- **Pros**: Keeps `cache` interface-coupled; supports future providers.
- **Cons**: Premature abstraction — there is one provider. The interface would be used by exactly one concrete type. Adds indirection with no current benefit.

### Option C: Move `ErrBadCandles` to `pkg/provider/` (the interface package)
- **Pros**: `cache` imports `pkg/provider` only (already does). Fully interface-coupled.
- **Cons**: `pkg/provider/` currently contains only the abstract `DataProvider` interface and its helpers. Adding an error type that was originally introduced to handle a Zerodha-specific artifact (tick-vs-aggregation) is a design smell — it implies all providers produce this kind of error.

## Decision

Option A for now: import `pkg/provider/zerodha` in `pkg/provider/zerodha/cache`. This is the minimal change. The dependency is documented here.

The decision is revisit-later rather than accepted — the coupling is real and will matter if a second provider implementation is ever added.

## Consequences

- `zerodha/cache` has a new compile-time dependency on `zerodha`. Any structural change to `zerodha.ErrBadCandles` (rename, field change) requires updating `cache.go`.
- The `TestCachedProvider_PropagatesErrBadCandles` and `TestCachedProvider_DoesNotCacheEmptyOnErrBadCandles` tests import `zerodha` directly — consistent with the production code.

## Revisit trigger

If a second `DataProvider` implementation is added that also needs to surface skipped-candle information, reconsider moving `ErrBadCandles` (or a generalized equivalent) to `pkg/provider/`. At that point Option C becomes the correct design.
