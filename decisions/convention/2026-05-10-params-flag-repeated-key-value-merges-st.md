# --params as repeated key=value flag merging into strategy param map

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | convention       |
| Tags     | CLI, params, flag-design, key-value, repeated-flag, cmd/evaluate, TASK-0073 |

## Context

`cmd/evaluate` needs to let callers override strategy parameters (e.g., use MACD fast=17 instead of the registry default). Options were: a single JSON file (`--params-file`), individual typed flags (as in `cmd/backtest` via `GlobalRegistry.RegisterFlags`), or a repeated key=value flag.

## Options considered

### Option A: Per-strategy typed flags via GlobalRegistry.RegisterFlags
- **Pros**: Type-checked at parse time; autocomplete-friendly.
- **Cons**: Adds 7–13 strategy-specific flags to every invocation of `cmd/evaluate`. Users mostly run with defaults; the flag proliferation is noise.

### Option B: JSON params file (--params-file path.json)
- **Pros**: Structured, version-controllable params.
- **Cons**: Requires a file on disk for a simple override. Awkward for quick CLI use.

### Option C: Repeated --params key=value flag (chosen)
- **Pros**: Ergonomic for CLI use: `--params fast-period=17 --params slow-period=26`. Zero overrides = all defaults. Unknown keys pass through and are ignored by strategy constructors. No extra file needed.
- **Cons**: Values are parsed as float64 strings; non-numeric values produce an error. Keys are not validated against the strategy's known params at flag-parse time (they're validated implicitly when passed to `GlobalRegistry.Build`).

## Decision

`--params key=value` is a repeated flag implemented via `flag.Func`. Each occurrence appends to `f.rawParams []string`. After flag parsing, `parseParams()` converts them to `map[string]float64` and `buildPipeline` merges overrides onto `GlobalRegistry.DefaultParams(stratName)`. Unrecognized keys survive in the map and are silently ignored by strategy constructors.

## Consequences

- Zero `--params` flags means all strategy defaults apply. This is the common case.
- Type errors in param values are caught early (`--params fast-period=abc` → parse error).
- Unknown key silently ignored mirrors the behavior of strategy constructors, which only read their named keys.
- The same pattern is appropriate for `cmd/param-search` (TASK-0077) if it also accepts strategy overrides.
