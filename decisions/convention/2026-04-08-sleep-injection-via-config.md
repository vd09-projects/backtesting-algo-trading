# Sleep injection via Config for rate-limit throttling

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-08 |
| Status   | accepted   |
| Category | convention |
| Tags     | zerodha, provider, testing, sleep, config, injection, rate-limit, convention |

## Context

`FetchCandles` sleeps 350ms between chunk requests to stay under the Kite Connect 3 req/sec rate
limit (decided in the pagination strategy). During tests, every multi-chunk fetch would incur
350ms × (chunks − 1) of real wall-clock wait. A test with two chunks takes 350ms; a test
with ten chunks takes 3.15 seconds. These times compound quickly across the test suite.

Several patterns exist in Go for making sleeps testable without importing a mocking framework:

1. **Global variable** — `var sleep = time.Sleep` at package level, overridden in tests
2. **Build tag** — compile a no-op sleep in test builds
3. **Interface** — define a `Sleeper` interface, inject an implementation
4. **Config field** — `Sleep func(time.Duration)` in the provider Config struct

## Options considered

### Option A — Global variable (`var sleep = time.Sleep`)

- **Pros:** Zero boilerplate at call sites. No change to constructor signature.
- **Cons:** Mutable package-level state. Violates the repo's "no global state" rule (CLAUDE.md).
  Test isolation breaks if two tests run in parallel and both modify the global.

### Option B — Build tag (`//go:build !test`)

- **Pros:** Zero runtime overhead in production. Transparent to callers.
- **Cons:** Requires two files (one for prod, one for test), adds build complexity, and hides the
  dependency rather than making it explicit. Counter to the repo's "explicit injection" rule.

### Option C — `Sleeper` interface

- **Pros:** Strongly typed. Can be swapped for a token-bucket implementation later.
- **Cons:** A single-method interface for `time.Sleep` is overkill. The interface adds indirection
  (`p.sleeper.Sleep(d)`) with no practical benefit over a function field. Go idiom favours
  function types over single-method interfaces.

### Option D — `Sleep func(time.Duration)` in Config (chosen)

- **Pros:** Explicit dependency injection (consistent with repo convention). Nil-safe default to
  `time.Sleep`. Tests pass a no-op: `Sleep: func(time.Duration) {}`. No global state, no build
  tags, no interface boilerplate. The zero-value Config naturally gets production behaviour.
- **Cons:** One extra field on Config. Callers who build Config manually must be aware of it,
  though the nil-default means they can ignore it.

## Decision

`Config.Sleep func(time.Duration)` — nil defaults to `time.Sleep` at runtime.

```go
if cfg.Sleep == nil {
    cfg.Sleep = time.Sleep
}
```

Tests use:

```go
Sleep: func(time.Duration) {}, // no-op: multi-chunk tests run instantly
```

This pattern applies to **any future injectable behavior** in this package (or the repo at large)
that needs to be overridable for tests without a full mock: timers, clocks, random sources, UUIDs.
Function fields in Config are the preferred mechanism unless a more complex interface is justified.

## Consequences

- The `Sleep` field on Config is a documentation signal: callers know rate-limiting exists and is
  injectable.
- Any future method that adds sleeps (retry backoff, reconnect delay) should follow the same
  pattern rather than adding a new field per behavior — consider a single `Clock` abstraction if
  multiple time-related functions accumulate.
- Tests that forget `Sleep: func(time.Duration){}` in multi-chunk scenarios will be slow but not
  wrong — a common gotcha to watch for in code review.

## Related decisions

- [Zerodha historical data — pagination and rate-limit strategy](../architecture/2026-04-07-zerodha-pagination-strategy.md) — the 350ms constant that motivates this pattern
- [context.Context deferred from Run() and DataProvider interface](../tradeoff/2026-04-06-context-parameter-deferred.md) — same "explicit injection via config" philosophy applied to context

## Revisit trigger

If the number of injectable behavior fields on Config grows beyond 3–4, consolidate into a
`Clock` or `TestHooks` sub-struct rather than polluting the top-level Config.
