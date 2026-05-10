# All cmd/evaluate stage outputs written to a dated subdirectory

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-10       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | output-layout, stage-dir, YYYY-MM-DD, cmd/evaluate, dated-subdir, TASK-0073 |

## Context

`cmd/evaluate` writes multiple output files per run: `universe-sweep.csv`, `wf-{instrument}.json`, `wf-{instrument}-folds.csv`, `bootstrap-{instrument}.json`, and `verdict.json`. The question was whether to write them directly into `--out-dir` or into a subdirectory.

## Options considered

### Option A: Write directly to --out-dir
- **Pros**: No nested directory; files are immediately accessible at the root.
- **Cons**: Running the same strategy twice would overwrite the previous run's files. Multiple strategies produce interleaved files in the same directory.

### Option B: Dated subdirectory per run (chosen)
- **Pros**: `--out-dir/YYYY-MM-DD-{strategy}-{timeframe}/` gives each run its own namespace. Multiple runs on the same day produce distinct directories (same prefix, different strategy/timeframe). YYYY-MM-DD prefix gives natural sort order. Root `--out-dir` stays clean across runs.
- **Cons**: One extra level of directory nesting. User must navigate into the dated subdir.

## Decision

All stage outputs are written to `--out-dir/YYYY-MM-DD-{strategy}-{timeframe}/`, created by `makeStageDir` at pipeline startup. Strategy name is sanitized for filesystem safety (`:`, ` `, `/` → `_`). If the same pipeline run fails and is retried on the same day, the existing dated dir already exists — `makeStageDir` currently uses `os.Mkdir` which fails on retry; TASK-0103 tracks changing this to `os.MkdirAll`.

## Consequences

- Users running multiple strategies can see all results side-by-side at the `--out-dir` level, sorted by date prefix.
- Re-running the same strategy/timeframe combination on the same date writes to the same subdirectory (after TASK-0103 fix).
- If `--out-dir` is a shared results directory, each run is isolated without manual naming.

## Revisit trigger

If `cmd/param-search` (TASK-0077) or other pipeline orchestrators also produce dated output directories, consider standardizing the subdirectory naming convention across all eval-related CLIs.
