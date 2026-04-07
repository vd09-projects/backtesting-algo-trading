# context.Context deferred from Run() and DataProvider interface

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-06       |
| Status   | accepted         |
| Category | tradeoff         |
| Tags     | context, engine, provider, interface, cancellation, zerodha |

## Context

During the pre-merge review, golangci-lint and the behavior dimension both flagged that `Engine.Run()` and `DataProvider.FetchCandles()` have no `context.Context` parameter. Without context, there is no way to cancel a stuck `FetchCandles` call, set a deadline on a backtest run, or propagate cancellation signals.

At the time of the review, no network-backed provider existed — only stub implementations used in tests. The engine loop itself is purely local computation.

## Options considered

### Option A: Add context now (before any real provider exists)
- **Pros**: Interface is correct from day one. No downstream migration needed when Zerodha provider is written. Follows Go best practice (context first on all I/O-capable functions).
- **Cons**: Every test stub and the engine run call needs updating immediately. Adds boilerplate before there's a concrete need.

### Option B: Defer until just before the first real provider is implemented
- **Pros**: Defers change to the moment when the cost of NOT having context becomes concrete and visible. Tests stay simpler now.
- **Cons**: If forgotten, the interface gets locked in without context and becomes expensive to change once multiple strategies and providers exist.

## Decision

Deferred (Option B). The `DataProvider` interface is the critical point — the interface contract must be changed **before** the Zerodha provider is written, while there is still only one implementation (the test stub). At that moment the migration cost is minimal: update the stub and the engine call. After that, the interface is closed.

`Engine.Run()` should also receive context at the same time, even if it only passes it through to `FetchCandles`.

## Consequences

If the Zerodha provider gets written without this change, every existing implementation needs context threading added retroactively. The risk is low now (zero real implementations) but grows fast once real code lands.

## Revisit trigger

N/A — resolved.

## Resolution (2026-04-07)

Context was added to both `Engine.Run(ctx context.Context, ...)` and
`DataProvider.FetchCandles(ctx context.Context, ...)` before any Zerodha provider code was
written. The trigger condition was met and the interface is correct. All test stubs were updated
at the same time. No retroactive migration was needed.
