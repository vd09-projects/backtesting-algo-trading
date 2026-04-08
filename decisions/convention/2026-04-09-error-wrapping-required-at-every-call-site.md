# Every error return must be wrapped with call-site context

| Field    | Value          |
|----------|----------------|
| Date     | 2026-04-09     |
| Status   | accepted       |
| Category | convention     |
| Tags     | error-handling, wrapping, fmt.Errorf, convention, zerodha, provider |

## Context

During a code review of the `pkg/provider/zerodha` package, three bare error returns were
found in `auth.go` and `instruments.go`:

```go
// auth.go — request construction
if err != nil {
    return "", err  // no context: which request? which call site?
}

// auth.go — file write
return os.WriteFile(path, data, 0o600)  // no context: what was being written?

// instruments.go — request construction
if err != nil {
    return nil, err  // no context
}
```

These were caught by code review, not by the linter. `errcheck` only flags silently
dropped errors — it does not flag bare `return err` without wrapping.

## Options considered

### Option A — Require wrapping at every named-function call site (chosen)

Every `return err` (or `return nil, err`) must add context via `fmt.Errorf("...: %w", err)`.

- **Pros:** Error messages are self-contained. When a caller sees an error, the chain
  pinpoints exactly where it originated. Debuggable without a stack trace.
- **Cons:** Slightly more verbose at each call site.

### Option B — Wrap only at package boundaries

Only wrap when the error exits the package. Internal helpers can return bare errors.

- **Pros:** Less boilerplate.
- **Cons:** Internal chains lose context. When `doHTTP` is called from three different
  functions, a bare `return err` from inside can't be distinguished at the package boundary.
  This is how the three bare returns went unnoticed — they were "internal" call sites.

## Decision

**Wrap every error at every named-function call site.** Use `fmt.Errorf("context: %w", err)`
with `%w` (not `%v`) to preserve the chain for `errors.Is` / `errors.As`.

The wrapping message should describe what was being attempted, not the error itself:

```go
// BAD — describes the error (the error already says this)
return fmt.Errorf("request failed: %w", err)

// GOOD — describes the operation
return fmt.Errorf("build instruments request: %w", err)
return fmt.Errorf("write token file: %w", err)
return fmt.Errorf("session exchange: %w", err)
```

The only accepted exceptions are:
1. **Intentional best-effort operations** — where the error is deliberately discarded with
   `_ = someFunc() //nolint:errcheck // reason`. These are not return paths.
2. **Re-wrapping an already-wrapped sentinel** — e.g., `ErrAuthRequired` is wrapped once
   at the HTTP layer by `doHTTP` with HTTP status code context. Callers that just propagate
   it do not need to add another layer.

## Consequences

- All error messages in the zerodha package now form a self-describing chain, e.g.:
  `zerodha: load instruments: build instruments request: invalid character in URL`
- New call sites that add bare `return err` will be caught in code review (the linter
  does not flag this automatically).
- Consistent with the `doHTTP` centralisation decision — auth errors are wrapped once
  at the HTTP layer; all other errors are wrapped at their own call sites.

## Related decisions

- [`doHTTP` centralizes 401/403 → ErrAuthRequired mapping](../architecture/2026-04-08-dohttp-centralizes-auth-errors.md) — establishes that auth errors are wrapped at the HTTP helper level; this decision extends the wrapping convention to all other error returns.

## Revisit trigger

If a structured logging or tracing library is adopted (e.g., `slog` with attributes),
reconsider whether string-based `fmt.Errorf` wrapping is still the right approach, or
whether error context should be attached as structured fields instead.
