# lazyProvider initFn uses context.Background(), not the caller's context

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | tradeoff         |
| Tags     | context, lazy-init, BuildProvider, auth, zerodha, cmdutil |

## Context

`BuildProvider` accepts a `context.Context` parameter (retained for call-site compatibility with existing callers). The `lazyProvider.initFn` closure — which runs token loading, `LoginFlow`, and `zerodha.NewProvider` — originally captured this outer `ctx`. If a caller ever passes a request-scoped or deadline-bearing context to `BuildProvider`, the lazy init would fire using that context on the first cache miss, potentially failing because the context was already cancelled or timed out by the time initialisation runs.

Currently all callers pass `context.Background()`, so the captured-ctx path was safe today but fragile by design.

## Options considered

### Option A: Capture caller ctx (prior behaviour)
- **Pros**: Single context flows through; if the caller cancels, auth also cancels.
- **Cons**: If the caller's context expires between `BuildProvider` and the first cache miss, `initFn` fires against a dead context and returns a spurious cancellation error. Auth and instruments-CSV fetch are one-time startup operations, not request-scoped work; cancelling them mid-flight leaves the provider in a permanently broken state (the `sync.Once` has fired but `inner` is nil and `initErr` is set).

### Option B: Create `context.Background()` inside `initFn`
- **Pros**: Auth and provider init are unconditionally non-cancellable. The one-time startup operations complete regardless of caller context lifecycle. Failure modes are real failures (bad token, network error), not spurious cancellations.
- **Cons**: The caller cannot cancel a long-running `LoginFlow` (interactive browser login) via context. In practice, `LoginFlow` reads from stdin and is human-paced; context cancellation of it would be unusual and surprising anyway.

## Decision

`initFn` creates its own `initCtx := context.Background()` and uses it for both `LoginFlow` and `zerodha.NewProvider`. The outer `ctx` parameter on `BuildProvider` is blanked (`_ context.Context`) to make the non-use explicit and prevent accidental future capture. A comment in the function signature explains the reasoning.

## Consequences

- Auth and instruments-CSV fetch cannot be cancelled via context. Acceptable: these are one-time operations that either succeed or fail with a real error.
- `BuildProvider`'s signature is unchanged; callers pass any context they like without side effects on initialisation. If the interface ever needs cancellable init, that's a larger redesign (probably removing `BuildProvider` in favour of a two-phase construct/connect pattern).
- The `_ context.Context` blank identifier makes the deliberate non-use visible to reviewers rather than silently discarding the parameter.

## Revisit trigger

If `LoginFlow` becomes async or long-running enough that callers need to be able to cancel it (e.g. HTTP-handler-initiated auth with a request deadline), the `initFn` will need to accept the per-call context explicitly rather than using `context.Background()`.
