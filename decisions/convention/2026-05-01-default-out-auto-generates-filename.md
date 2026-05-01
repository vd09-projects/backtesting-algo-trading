# Default --out auto-generates filename when flag is omitted

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-01       |
| Status   | experimental     |
| Category | convention       |
| Tags     | CLI, --out, default-filename, timeframe, cmd/backtest, TASK-0064 |

## Context

Before TASK-0064, `cmd/backtest` defaulted `--out` to `""` (empty string), which meant no JSON output was written unless the caller explicitly supplied a path. TASK-0064's acceptance criteria required the default filename to include timeframe so that daily and intraday runs of the same strategy on the same instrument are distinguishable. The question was: should omitting `--out` produce a default file, or should it still mean "no output"?

## Options considered

### Option A: Keep --out="" as "no output" default; require explicit opt-in
- **Pros**: No behavior change for existing callers
- **Cons**: Violates the AC requirement; users would need to construct the canonical filename manually on every run; the whole point of TASK-0064 is that filenames encode run identity

### Option B: Auto-generate filename when --out is omitted (chosen)
- **Pros**: Fulfills AC; every run produces a self-describing JSON file by default; easy for callers who want no output to pass `--out=/dev/null` or redirect
- **Cons**: Behavior change — callers who previously relied on omitting `--out` to suppress output now get a file written to the working directory

## Decision

When `--out` is omitted (empty string), `cmd/backtest` auto-generates a canonical filename via `cmdutil.DefaultOutPath`. The generated name follows `{strategy}-{instrument}-{timeframe}-{from}-{to}.json` with instrument sanitized for filesystem safety. This is the right default: every backtest run should produce a self-describing artifact. Callers who want to suppress JSON output must supply `--out=/dev/null` explicitly.

The flag help text was updated to document this behavior: "Path for JSON results export; when omitted a default name is generated from the run params".

## Consequences

Any script or tool that previously relied on no JSON file being written when `--out` was omitted will now get an unexpected file in the working directory. This is a conscious behavior change, justified by the AC requirement. The flag semantics are documented.

## Revisit trigger

If there are callers that cannot tolerate unexpected file creation and cannot be updated to pass `--out=/dev/null`. At that point, add a `--no-out` flag to explicitly suppress output.
