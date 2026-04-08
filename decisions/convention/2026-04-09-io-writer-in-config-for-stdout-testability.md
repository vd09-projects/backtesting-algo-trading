# io.Writer field in Config for stdout testability

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-09       |
| Status   | accepted         |
| Category | convention       |
| Tags     | output, testing, io.Writer, Config, testability, stdout, convention |

## Context

`internal/output.Write` needs to print a human-readable summary to stdout. Unit tests must not pollute test output, so they can't let the real `os.Stdout` be written to. Three approaches were evaluated.

## Options considered

### Option A: Redirect `os.Stdout` in tests via `os.Pipe()`
- **Pros**: No public API change; `Config` stays minimal.
- **Cons**: Verbose boilerplate in every test that checks output. Not goroutine-safe — tests using `t.Parallel()` would race on `os.Stdout`. Fragile if the file descriptor is inherited by other goroutines.

### Option B: Unexported `write(report, cfg, w io.Writer)` + exported `Write` delegates
- **Pros**: Exported `Config` stays minimal.
- **Cons**: The real `os.Stdout` path (inside `Write`) is never exercised by tests, only the wrapper. Error propagation must be duplicated or the wrapper must be trivial.

### Option C: `Stdout io.Writer` field in `Config`, nil → `os.Stdout`
- **Pros**: Tests pass `&bytes.Buffer{}` — straightforward and goroutine-safe. The real nil-defaults-to-os.Stdout path is clearly documented. No wrapper function needed.
- **Cons**: One extra field on a public struct that production callers won't set. Slightly unusual — the field is primarily a test seam, not a user-facing feature.

## Decision

`Config.Stdout io.Writer` with nil defaulting to `os.Stdout`. The test-seam field is transparent (zero-value works for production use) and makes all paths through `Write` unit-testable without OS-level stdout capture.

This follows the same approach used elsewhere in the project for injectable behaviours — see sleep injection via Config in the zerodha provider.

## Consequences

- All callers in production code leave `Stdout` unset (nil) — no change in behaviour.
- Test code passes `&bytes.Buffer{}` to capture and assert output content.
- The `cfg.Stdout == nil → os.Stdout` branch is intentionally untested in unit tests; it is covered by integration/manual smoke testing.

## Related decisions

- [Sleep injection via Config for rate-limit throttling](../convention/2026-04-08-sleep-injection-via-config.md) — established the pattern of injectable behaviours via Config fields; this decision applies the same pattern to stdout.
