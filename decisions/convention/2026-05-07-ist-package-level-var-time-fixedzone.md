# IST timezone as package-level var via time.FixedZone — no tzdata dependency

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | IST, timezone, pkg/strategy, session-boundary, time.FixedZone, TASK-0078 |

## Context

TASK-0078 added `pkg/strategy/session.go` with `IsSessionOpen` and `PreviousSessionClose`, both of which require converting `Candle.Timestamp` to IST (UTC+5:30) before comparing against session boundary times. The IST timezone value (`time.FixedZone("IST", 5*3600+30*60)`) needs to be available inside these functions without introducing a `tzdata` dependency or per-call allocation.

Two options were considered: construct the `*time.Location` once at package load as a `var`, or construct it on each function call.

## Options considered

### Option A: Package-level var (selected)

```go
var ist = time.FixedZone("IST", 5*3600+30*60)
```

- **Pros**: Constructed once; zero per-call allocation; consistent with how other packages handle fixed-offset zones (e.g., the auth code in `pkg/provider/zerodha` uses the same pattern for token expiry). No `tzdata` dependency — `time.FixedZone` uses a hardcoded offset, not a named zone requiring OS timezone database.
- **Cons**: Technically a package-level variable, which the CLAUDE.md rule "no global state" cautions against. However, the rule targets mutable state with side effects; `ist` is an immutable value set once at package init from a pure function — categorically different from a mutable global or an `init()` with I/O.

### Option B: Construct per call

```go
func IsSessionOpen(bar model.Candle) bool {
    ist := time.FixedZone("IST", 5*3600+30*60)
    // ...
}
```

- **Pros**: Avoids any package-level state.
- **Cons**: Allocates a `*time.Location` on every call to `IsSessionOpen` and `PreviousSessionClose`. Both functions are called on every bar in the event loop's hot path (inside strategy `Next()` implementations). Unnecessary allocation with no behavioral difference.

## Decision

Use a **package-level var**: `var ist = time.FixedZone("IST", 5*3600+30*60)` in `pkg/strategy/session.go`.

This is consistent with the accepted interpretation of the no-global-state rule: the rule prohibits mutable state and `init()` functions with side effects. A package-level var initialized from a pure, side-effect-free function (`time.FixedZone`) that is never mutated after initialization is not global mutable state — it is a package constant that happens to be a reference type. The `time.Location` value returned by `time.FixedZone` is itself immutable once created.

Using `time.FixedZone` with a numeric offset (not a named zone string like "Asia/Kolkata") means no OS timezone database is required. This avoids the `tzdata` import that would otherwise be needed for Kolkata-based lookups, and matches the engine convention documented in the TASK-0078 acceptance criteria.

## Consequences

- `IsSessionOpen` and `PreviousSessionClose` are zero-allocation in the hot path.
- No `tzdata` import in `go.mod` — the repo stays dependency-minimal.
- Any future `pkg/strategy/` function that needs IST can reference the same `ist` var rather than constructing its own.
- The var is unexported (`ist`, not `IST`) — callers cannot depend on it directly; timezone handling stays internal to the package.

## Related decisions

- [Value semantics for domain types](./2026-04-06-value-semantics-for-domain-types.md) — same principle: avoid allocations in the hot path by choosing the representation that eliminates per-call overhead.

## Revisit trigger

If the engine is extended to support markets outside IST (US equities, crypto), the `ist` var should be generalised to a configurable timezone injected via session config rather than hardcoded at package level.
