# fetch-history progress manifest is a cmd-layer concern, not CachedProvider

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | partial-failure, manifest, progress-tracking, cmd-layer, cmd/fetch-history, CachedProvider |

## Context

`cmd/fetch-history` is a bulk fetch tool that may run for hours across dozens of instruments and
multiple timeframes. A crash or network failure mid-run should not force a full restart from `--from`.
Some form of progress tracking is required.

There are two distinct tracking concerns in this space:

1. **Incremental timestamp tracking (TASK-0080):** `CachedProvider` needs a per-instrument
   `LastCandleTime` manifest to know the delta between what's cached and today. This is a
   provider-level concern that all callers benefit from.

2. **Bulk run progress tracking:** `cmd/fetch-history` needs to know which `{instrument × timeframe}`
   pairs from the current run have been fully completed. This is needed so a re-run after partial
   failure can skip already-done pairs without re-fetching.

## Options considered

### Option A: Add progress tracking to `CachedProvider`
- **Pros**: Single manifest file; leverages existing cache infrastructure.
- **Cons**: `CachedProvider` is a DataProvider decorator — its job is caching, not tracking bulk-run
  state. Adding run-state tracking couples a general-purpose cache layer to a specific CLI workflow.
  TASK-0080 is already planned for the per-instrument timestamp; conflating the two concerns would
  produce a muddled API.

### Option B: Separate `fetch-progress.json` in the cmd/ layer (chosen)
- **Pros**: Clean separation. `CachedProvider` tracks "what's cached"; `fetch-progress.json` tracks
  "what the current bulk run has completed". The two manifests serve different lifetimes:
  `fetch-progress.json` is reset when a new run starts; the CachedProvider manifest persists
  indefinitely. cmd-layer state belongs in the cmd/ layer.
- **Cons**: Two manifest files in the cache directory. Slightly more user-visible state.

## Decision

`fetch-progress.json` is a cmd-layer progress tracker written to `{cache-dir}/fetch-progress.json`.
It records completed `{instrument, timeframe}` entries. On re-run, completed pairs are skipped.
Users delete the file to force a full re-fetch.

`CachedProvider` is unchanged. TASK-0080's incremental manifest (`fetch-manifest.json` at
`{cache-dir}/{instrument}/{tf}/`) is a separate file serving a different purpose.

## Consequences

- Two manifest files coexist in the cache directory: `fetch-progress.json` (cmd-layer bulk-run state)
  and `{instrument}/{tf}/fetch-manifest.json` (TASK-0080, per-instrument incremental timestamp).
  Users who inspect the cache directory will see both; the distinction is documented here.
- `fetch-progress.json` is written atomically (write to `.tmp`, then `os.Rename`) — same pattern as
  TASK-0080's manifest, following the established convention.
- When TASK-0080 ships and `LastCachedTime` is wired in, `fetch-progress.json` and the CachedProvider
  manifest will have overlapping but non-redundant semantics: one tracks run completion, the other
  tracks the last candle timestamp.

## Related decisions

- [CachedProvider incremental time-series manifest (TASK-0080)](../architecture/2026-05-05-instruments-cache-dir-explicit-config-field.md) — the complementary per-instrument incremental timestamp manifest; distinct concern.

## Revisit trigger

If `CachedProvider.LastCachedTime` from TASK-0080 is implemented and `cmd/fetch-history` uses it to
compute the delta, the interaction between the two manifests should be reviewed: specifically, if
`fetch-progress.json` marks a pair complete but the cache files are deleted (e.g. `rm -rf .cache/`),
a re-run will skip the pair even though nothing is actually cached. Add a validation step or document
this edge case.
