# --out-dir existence validated at startup before any engine runs

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-13       |
| Status   | experimental     |
| Category | convention       |
| Tags     | validation, out-dir, startup, cmd/param-search, fail-fast, TASK-0077 |

## Context

`cmd/param-search` writes its output to `--out-dir/param-search-results.csv`. If `--out-dir` doesn't exist, the write fails. The question was when to check: at startup (before engine runs) or at write time (after all variants have run).

## Decision

Check with `os.Stat(f.outDir)` in `buildSearchPipeline`, before any engine runs begin. An error is returned immediately if the directory does not exist. The caller sees the error before any network or compute work is done.

The alternative — checking at write time — would waste all engine run time on a 200-variant grid before failing on a missing directory. This is analogous to the flag validation pattern in all cmd binaries: validate all preconditions at startup, before I/O.

## Consequences

- `--out-dir` must exist at call time; the tool does not create it. This is documented in the flag help text and README.
- The check is `os.Stat`, not `os.MkdirAll`. The tool deliberately does not create the directory — the caller controls the directory lifecycle (matches the existing pattern in `cmd/evaluate` and `cmd/walk-forward` where `makeStageDir` creates a subdirectory inside an existing `--out-dir` that must already exist).
- TASK-0103 (cmd/evaluate `makeStageDir` should use `os.MkdirAll`) is related but separate: that fix addresses a retry-after-error case, not the startup validation question.

## Related decisions

- [cmd/evaluate --out-dir dated subdirectory](../../architecture/2026-05-10-stage-outputs-in-dated-subdirectory-cmd-eval.md) — cmd/evaluate equivalent; same startup validation pattern

## Revisit trigger

If the tool is commonly invoked from scripts that need to create `--out-dir` on first run, add `os.MkdirAll` here (matching TASK-0103 for cmd/evaluate). Current behavior matches the broader cmd pattern where callers manage directories.
