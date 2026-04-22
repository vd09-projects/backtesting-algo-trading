# `buildProvider` extracted from cmd binaries into `internal/cmdutil.BuildProvider`

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-22       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | provider, DRY, cmd, zerodha, cmdutil, refactor, TASK-0035 |

## Context

`cmd/backtest/main.go` and `cmd/sweep/main.go` each had an identical private `buildProvider` function (~30 lines): load env, load/exchange token, construct Zerodha provider, wrap in cache. Adding `cmd/universe-sweep` would have made a third copy.

## Decision

Extracted to `internal/cmdutil.BuildProvider(ctx context.Context) (*cache.CachedProvider, error)`. All three cmd binaries call `cmdutil.BuildProvider(ctx)`. The function is a pure I/O constructor with no business logic; it belongs alongside `MustEnv`, `TokenFilePath`, and `LoginFlow` — the other shared cmd-layer plumbing already in `cmdutil`.

Three copies is the threshold for extraction in this codebase. Two copies are acceptable (cost of abstraction vs. cost of duplication is close). Three copies means the duplication will compound with every new binary.

## Consequences

`cmd/sweep/main.go` had its local `buildProvider` removed and its unused imports (`net/http`, `zerodha`, `cache`) cleaned up in the same commit. `cmd/backtest/main.go` was already calling `cmdutil.BuildProvider` before this session. Any future `cmd/` binary that needs a Zerodha provider calls `cmdutil.BuildProvider` — no copy-paste.
