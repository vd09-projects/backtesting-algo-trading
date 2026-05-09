# providerFactory receives parsed flags — no global state required

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | convention       |
| Tags     | testability, provider-factory, no-global-state, dry-run, cmd/fetch-history, TASK-0070 |

## Context

`cmd/fetch-history`'s `run()` function needs to construct a provider on non-dry-run paths but must
not construct one on dry-run paths. Tests must be able to inject a mock provider without real
credentials.

An initial implementation used a `var productionMode bool` set by `main()` before calling `run()`.
`resolveProvider()` checked `productionMode` to decide whether to call `buildProductionProvider` or
the injected factory. The quality gate flagged this as a CLAUDE.md violation ("No global state").

## Options considered

### Option A: `productionMode` package-level bool (rejected)
- `main()` sets `productionMode = true`; `run()` checks it in `resolveProvider()`.
- **Cons**: Package-level mutable state. Violates CLAUDE.md. Not testable without resetting the var
  between tests (test isolation risk). Also, comparing function values to detect the "real" factory
  vs. the "test" factory is impossible in Go.

### Option B: `run()` accepts `provider.DataProvider` directly; `main()` builds it first
- `main()` calls `parseFlags()`, builds the real provider, then passes it to `run()`.
- **Cons**: `run()` can no longer be the single testable entry point: `main()` needs to expose
  `parseFlags()` externally, and dry-run paths would still construct a provider unnecessarily.

### Option C: `providerFactory func(fetchFlags) (provider.DataProvider, error)` (chosen)
- `run()` accepts a factory that receives the fully-parsed `fetchFlags`.
- `main()` passes `buildProductionProvider` directly — a top-level function with the right signature.
- Tests inject `func(_ fetchFlags) (provider.DataProvider, error) { return mock, nil }`.
- The factory is only invoked after dry-run check returns false.

## Decision

`run()` has the signature:
```go
func run(args []string, stdout, stderr io.Writer, providerFactory func(fetchFlags) (provider.DataProvider, error)) error
```

`main()` passes `buildProductionProvider` as the factory. The production provider is built from the
parsed flags — no closure over package-level state needed. Tests inject a factory that ignores flags
and returns a mock.

## Consequences

- No package-level mutable state. `productionMode` var and `resolveProvider()` helper removed.
- `buildProductionProvider` is a testable pure function (0% coverage — integration-only — but
  structurally sound).
- The factory receives the full `fetchFlags` struct (120 bytes), triggering a `gocritic hugeParam`
  lint warning. Suppressed with `//nolint:gocritic` — the function is called once per process and
  value semantics at the cmd-layer API boundary is the repo convention.
- This pattern is specific to cmd/ packages that need lazy provider construction. Other cmd/ packages
  (`cmd/walk-forward`, `cmd/monitor`) do not inject a factory — they are not batch tools and do not
  need the dry-run bypass.

## Related decisions

- [run() extraction for testability in cmd/walk-forward](2026-05-03-run-extraction-for-testability-in-cmd-walk-forward.md) — the originating pattern for run() testability.
- [run() extraction for testability in cmd/monitor](2026-05-07-run-extraction-for-testability-in-cmd-monitor.md) — second application; cmd/monitor has no DataProvider injection.

## Revisit trigger

If a second batch CLI needs the same providerFactory pattern, extract a shared type alias or helper
to `internal/cmdutil` rather than repeating the convention per-package.
