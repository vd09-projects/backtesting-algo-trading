# fetch-history uses direct apiKey+accessToken, not OAuth login flow

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-09       |
| Status   | experimental     |
| Category | architecture     |
| Tags     | auth, zerodha, access-token, batch-tooling, BuildProvider, cmd/fetch-history |

## Context

`cmdutil.BuildProvider` — used by `cmd/backtest`, `cmd/universe-sweep`, and `cmd/walk-forward` — loads
the Zerodha access token from a saved file and falls back to an interactive browser-based OAuth login
flow. This is the right design for interactive tools run by a developer who can open a browser.

`cmd/fetch-history` is a different class of tool: a long-running batch job (potentially hours) designed
to drain years of historical data in one shot. It is a natural candidate for scheduled CI/CD runs or
cron jobs where no interactive browser session is available. The OAuth flow would block such runs.

## Options considered

### Option A: Use `cmdutil.BuildProvider` (OAuth flow)
- **Pros**: Consistent with all other cmd/ binaries; reuses existing token management.
- **Cons**: Blocks on interactive browser login when no saved token exists. Unusable in CI/CD.
  The saved-token path would work today but adds a silent dependency on the token file path.

### Option B: Accept `--api-key` and `--access-token` directly (chosen)
- **Pros**: Fully scriptable. Token injected at call site — standard CI/CD pattern (`KITE_ACCESS_TOKEN`
  env var). No dependency on saved token file path. Makes credentials explicit in the invocation.
  Users who run the browser flow can export their token file's value to `KITE_ACCESS_TOKEN` for scripted
  use.
- **Cons**: Slightly different credential UX from other CLIs. Access token must be obtained and refreshed
  externally (Zerodha tokens expire daily). Does not benefit from `cmdutil.BuildProvider`'s lazy-init
  behaviour — but fetch-history always fetches, so lazy init provides no benefit here anyway.

## Decision

`cmd/fetch-history` accepts `--api-key` and `--access-token` flags (env vars: `KITE_API_KEY` and
`KITE_ACCESS_TOKEN`) and passes them directly to `zerodha.NewProvider` via a local
`buildProductionProvider` function. The `cmdutil.BuildProvider` OAuth flow is not used.

## Consequences

- `cmd/fetch-history` is the only cmd binary that does not use `cmdutil.BuildProvider`. This is
  intentional — it is a different class of tool (batch vs. interactive).
- Token lifecycle management (daily renewal, export to env) is the user's responsibility.
- If a second batch tool is added, the same direct-token pattern should apply to it; extract to a
  `cmdutil.BuildDirectProvider` helper if there are three or more batch callers.

## Related decisions

- [lazyProvider: defer Zerodha auth and client init to first cache miss](2026-05-07-lazy-provider-pattern-defer-auth-to-cache-miss.md) — the lazy-auth pattern is for interactive tools; fetch-history always hits the network, so lazy init provides no benefit.

## Revisit trigger

If `cmd/fetch-history` is ever run in an environment where both the access token and the token file
can be provided, or if `cmdutil.BuildProvider` gains a `--access-token` flag bypass mode, revisit
whether to unify the UX.
