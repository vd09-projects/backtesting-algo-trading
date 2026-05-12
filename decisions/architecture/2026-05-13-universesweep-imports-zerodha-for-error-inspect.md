# internal/universesweep imports pkg/provider/zerodha for ErrIncompleteData type assertion

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | ErrIncompleteData, zerodha, package-boundary, TASK-0083 |

## Context

CLAUDE.md contains the rule: "All data access must go through the `DataProvider` interface. No package outside `pkg/provider/` should know about Zerodha." The rule exists to prevent packages from bypassing the DataProvider abstraction and calling Zerodha APIs directly.

TASK-0083 Option B requires `internal/universesweep.runInstrument` to catch `*zerodha.ErrIncompleteData` when `eng.Run` returns it (wrapped as `"engine: fetching candles: %w"`). To call `errors.As(err, &ie)` where `ie` is `*zerodha.ErrIncompleteData`, `internal/universesweep` must import `pkg/provider/zerodha`.

## Options considered

### Option A: Import zerodha directly for type assertion (chosen)
- **Pros**: Direct, explicit, type-safe. The only Zerodha-specific symbol imported is the error type, not any API client or auth logic.
- **Cons**: `internal/universesweep` gains a compile-time dependency on a concrete provider package.

### Option B: Move ErrIncompleteData to pkg/provider (shared interface package)
- **Pros**: universesweep stays interface-coupled.
- **Cons**: `pkg/provider` currently defines only the abstract DataProvider interface. Adding a Zerodha-specific error concept ("candles incomplete vs weekday estimate") to the abstract interface package is a category error — it implies all future providers must surface this kind of error.

### Option C: Handle at cmd/ layer only (original plan, rejected)
- **Pros**: universesweep stays clean.
- **Cons**: The sweep would abort on the first incomplete instrument rather than continuing. User explicitly chose Option B (per-instrument warning, sweep continues).

## Decision

Option A: `internal/universesweep` imports `pkg/provider/zerodha` for `errors.As` type assertion only. The CLAUDE.md rule targets **data access** — reading candles, calling FetchCandles. Type-asserting on a typed error returned through the DataProvider interface is **error inspection**, not data access. The error already propagated through the abstract interface; universesweep is just reading its type. This was explicitly approved by the user as the correct semantics for Option B.

The precedent for this pattern is `pkg/provider/zerodha/cache` importing its parent `pkg/provider/zerodha` for `*ErrBadCandles` inspection (see `decisions/architecture/2026-05-10-cache-imports-zerodha-for-errbadcandles-not-circular.md`).

## Consequences

- `internal/universesweep` has a new compile-time dependency on `pkg/provider/zerodha`.
- If a second DataProvider is added that returns a different incomplete-data error type, `runInstrument` would need a second `errors.As` branch or a shared interface in `pkg/provider` (same revisit trigger as the ErrBadCandles decision).

## Revisit trigger

If a second DataProvider implementation is added, reconsider whether `ErrIncompleteData` (or a generalized equivalent) should be promoted to `pkg/provider/` so `internal/universesweep` can remain interface-coupled.

## Related decisions

- [cache imports zerodha for ErrBadCandles inspection](2026-05-10-cache-imports-zerodha-for-errbadcandles-not-circular.md) — same pattern applied to cache sub-package
