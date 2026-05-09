---
name: cmd/fetch-history pattern and status
description: Architecture decisions and test patterns established in TASK-0070 for the bulk intraday data fetch CLI
type: project
---

TASK-0070 (cmd/fetch-history) is done as of 2026-05-09.

**Why:** Enables all future intraday backtests to run from local disk cache without a Zerodha token, by draining full history for all instruments in a universe YAML.

Key architectural decisions:
- Does NOT use cmdutil.BuildProvider — accepts --api-key/--access-token directly (or KITE_API_KEY/KITE_ACCESS_TOKEN); batch/CI use case requires direct token injection
- providerFactory is func(fetchFlags)(provider.DataProvider,error) — receives parsed flags so main() passes buildProductionProvider directly; no global productionMode sentinel
- fetch-progress.json in --cache-dir tracks bulk-run completion (resume-after-failure); distinct from TASK-0080's per-instrument incremental manifest in CachedProvider
- Incremental mode (LastCachedTime) stubbed with TODO(TASK-0080) — full range fetch until TASK-0080 ships
- TASK-0083 (*ErrIncompleteData typed error handling) applies to cmd/fetch-history but tracked separately

Follow-up tasks open:
- TASK-0080: CachedProvider incremental manifest (unblocks incremental mode)
- TASK-0083: *ErrIncompleteData handling at cmd/ layer (now actionable)
- TASK-0093: os.MkdirAll guard in fetchAll + 2 missing parseFlags tests
- TASK-0094: fetchOne 11-parameter refactor to fetchState struct

**How to apply:** When building other batch CLIs (non-interactive, CI/CD), use the same direct-token + providerFactory pattern rather than cmdutil.BuildProvider.
