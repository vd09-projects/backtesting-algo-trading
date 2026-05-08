# MustGet panics on unknown strategy name — fail-fast at startup, not silently at runtime

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | strategy, registry, mustget, fail-fast, startup-validation, panic, cmdutil, TASK-0079 |

## Context

`StrategyRegistry.MustGet(name)` needed to signal when a strategy name is not found. The signal options were: return an error (caller decides what to do), return a nil factory (caller must check for nil), or panic (process dies immediately with a descriptive message).

The callers of `MustGet` are cmd binaries at flag-parse time — they call it with the value of `--strategy` before any data fetch, network call, or engine run. An unknown strategy name at this point is unambiguously a programming error or a user input error. There is no useful recovery path.

## Options considered

### Option A: Return (factory, error) — two-value return
`MustGet` returns `(func() strategy.Strategy, error)`; callers check the error.
- **Pros**: Standard Go pattern. Composable.
- **Cons**: Callers can ignore the error (and historically have done so with switch-default paths). The `Must*` naming convention in Go (`Must`, `MustCompile`, etc.) explicitly signals panic-on-failure — a two-value return named `MustGet` would be misleading.

### Option B: Return nil factory — single-value return, nil signals failure
`MustGet` returns `func() strategy.Strategy`; nil means not found.
- **Pros**: Simple.
- **Cons**: Callers can ignore nil. A nil factory called at fold time produces a panic inside a goroutine with no useful context — the *worst* failure mode.

### Option C: Panic with descriptive message (chosen)
`MustGet` panics with `"unknown strategy %q; available: sma-crossover, ..."` if the name is not registered.
- **Pros**: Consistent with the `Must*` naming convention in Go standard library (`regexp.MustCompile`, `template.Must`). Fails immediately at program startup with a message that includes the unknown name and all available names. Cannot be silently ignored. All callers invoke `MustGet` before any I/O, so the panic surfaces cleanly in the main goroutine.
- **Cons**: Callers cannot recover from an unknown strategy name without wrapping in `recover()` — but there is no meaningful recovery for an unknown strategy name at startup, so this is not a real con.

## Decision

`MustGet` panics with `fmt.Sprintf("unknown strategy %q; available: %s", name, strings.Join(available, ", "))`. The panic message names the unknown strategy and lists all registered names, which is sufficient information for a user to correct the `--strategy` flag value.

All cmd binaries that call `MustGet` do so at flag-parse time, before any network call, fold run, or engine execution. The panic fires in the main goroutine and produces a clean crash with a useful message — not a goroutine-internal panic with a stack trace pointing into the middle of a fold.

`Register` also panics on duplicate name registration — same reasoning: duplicate registration is a programmer error with no meaningful recovery.

## Consequences

- Existing tests that expected an error return for unknown strategy names (`TestStrategyFactory_UnknownStrategy`, `TestRun_UnknownStrategy` in `cmd/walk-forward`) were updated to use `defer recover()` assertions.
- Any future caller of `MustGet` must understand the fail-fast contract. The `Must` prefix is the signal; callers should not call `MustGet` from a goroutine where a panic would be unrecoverable without explicit `recover()` wrapping.

## Related decisions

- [StrategyRegistry type in internal/cmdutil](../architecture/2026-05-07-strategy-registry-in-internal-cmdutil.md) — the type this function belongs to

## Revisit trigger

If a use case emerges where `MustGet` is called from within a long-running goroutine (e.g., a web server handling per-request strategy dispatch), replace the call site with a checked `Get(name) (factory, bool)` helper. Do not change `MustGet` semantics.
