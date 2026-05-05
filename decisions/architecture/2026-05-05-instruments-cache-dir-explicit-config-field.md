# InstrumentsCacheDir as explicit optional field on zerodha.Config

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-05       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | zerodha, provider, InstrumentsCacheDir, Config, dependency-injection, no-global-state, TASK-0081 |

## Context

The instruments CSV caching feature (TASK-0081) needed a way for callers to opt in to caching without breaking existing consumers that don't need it. Three patterns were considered: explicit Config field, environment variable, and auto-detection from the candle cache path. The choice had to respect the CLAUDE.md no-global-state rule.

## Options considered

### Option A: Explicit optional field — `InstrumentsCacheDir string` on `zerodha.Config`
- **Pros**: No global state; dependency injection; empty = backward-compatible no-op; wired once in `cmdutil.BuildProvider`; testable without env setup.
- **Cons**: Increases Config struct surface by one field.

### Option B: Environment variable — `ZERODHA_INSTRUMENTS_CACHE_DIR`
- **Pros**: No API change.
- **Cons**: Violates no-global-state rule (env var is ambient global state); makes tests environment-dependent; can't be set differently for different cmd/ binaries in the same process.

### Option C: Auto-detect from candle cache path
- **Pros**: Zero API change.
- **Cons**: Requires `NewProvider` to know about the candle cache layout — wrong layer coupling; brittle if cache layout changes.

## Decision

Option A. `InstrumentsCacheDir string` added to `zerodha.Config`. Empty string = original behavior (no cache, network always required). Non-empty = cache-aware path. `cmdutil.BuildProvider` passes `cacheDir` as `InstrumentsCacheDir` at wiring time, making the dependency explicit and visible. No auto-registration, no `init()`, no env vars.

## Consequences

- All existing cmd/ binaries get instruments CSV caching automatically via `cmdutil.BuildProvider` with no per-binary code changes.
- Tests that don't set `InstrumentsCacheDir` continue to exercise the original code path.
- Adding `InstrumentsCacheDir` to Config is a non-breaking change (struct field with zero-value default).

## Related decisions

- [Instruments CSV cached at cacheDir/instruments.csv](2026-05-05-instruments-csv-cache-path.md) — companion decision on where the file lives.
- [buildProvider extracted to cmdutil](2026-04-22-buildprovider-extracted-to-cmdutil.md) — the wiring point; cmdutil.BuildProvider is where InstrumentsCacheDir is set for all CLIs.

## Revisit trigger

If a third cache-related Config field is added (e.g., candle cache TTL), consider promoting cache settings into a dedicated `CacheConfig` sub-struct to avoid Config bloat.
