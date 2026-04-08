# `doHTTP` helper centralizes 401/403 → ErrAuthRequired mapping

| Field    | Value        |
|----------|--------------|
| Date     | 2026-04-08   |
| Status   | accepted     |
| Category | architecture |
| Tags     | zerodha, provider, http, auth, error-handling, ErrAuthRequired, doHTTP, convention |

## Context

The Zerodha provider makes HTTP calls from three places: `ExchangeToken` (session endpoint),
`loadInstrumentsCSV` (instruments master), and `fetchChunk` (historical candles). All three can
receive a 401 or 403 response when the access token is expired or invalid. The question is where
the mapping from `HTTP 4xx` → `ErrAuthRequired` should live.

## Options considered

### Option A — Each caller checks the status code independently

```go
// in fetchChunk:
body, err := doHTTP(client, req)
var httpErr *httpError
if errors.As(err, &httpErr) && (httpErr.Code == 401 || httpErr.Code == 403) {
    return nil, fmt.Errorf("fetch chunk: %w", ErrAuthRequired)
}
```

- **Pros:** Each call site has full control over the error mapping.
- **Cons:** 401/403 → ErrAuthRequired mapping is repeated at every call site. A new call site that
  forgets to check will silently return a raw HTTP error instead of `ErrAuthRequired`, breaking
  `errors.Is` checks in the caller. Violates DRY for a correctness-critical path.

### Option B — Typed `httpError` struct, callers unwrap with `errors.As`

Define `type httpError struct{ Code int; Body string }`. Callers use `errors.As` to inspect status
codes.

- **Pros:** Maximum flexibility — callers can react to any status code.
- **Cons:** The provider has no callers that need to distinguish 401 from 403 from 404 in different
  ways. All non-200 responses except auth errors are fatal and wrapped as generic errors. Adding a
  typed struct for one sentinel mapping is over-engineering.

### Option C — `doHTTP` maps 401/403 → ErrAuthRequired directly (chosen)

```go
switch resp.StatusCode {
case http.StatusOK:
    return body, nil
case http.StatusUnauthorized, http.StatusForbidden:
    return nil, fmt.Errorf("HTTP %d: %w", resp.StatusCode, ErrAuthRequired)
default:
    return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, body)
}
```

- **Pros:** The mapping is defined once, in one place. Every HTTP call in the package automatically
  gets correct auth error handling without any call-site boilerplate. Adding a new HTTP call in the
  future cannot accidentally miss auth error detection. The `%w` wrapper preserves `errors.Is`
  while including the status code in the message.
- **Cons:** `doHTTP` encodes domain knowledge (that 401/403 means "re-auth") rather than being a
  generic HTTP utility. This is intentional — `doHTTP` is a package-private helper, not a general
  library.

## Decision

**Option C — `doHTTP` centralizes the 401/403 → `ErrAuthRequired` mapping.**

`doHTTP` is in `http.go` as a package-private function. It handles three cases:
- 200 → return body, nil
- 401/403 → return nil, `fmt.Errorf("HTTP %d: %w", code, ErrAuthRequired)`
- other → return nil, `fmt.Errorf("HTTP %d: %s", code, body)`

This is the right boundary: the helper knows about the transport protocol (HTTP status codes)
and the package knows about its domain error type (`ErrAuthRequired`). Combining them in the
shared helper is correct because the mapping is uniform across all endpoints.

## Consequences

- **Any new HTTP call in the package automatically handles auth errors correctly** — no checklist
  needed at code review time.
- If Zerodha ever returns a different status code for auth failures (e.g. 419 or a JSON error body
  with `status: "error"`), only `doHTTP` needs updating.
- The `ErrAuthRequired` sentinel can propagate all the way from any HTTP call through FetchCandles
  to the CLI, where it triggers the login flow.
- `doHTTP` is package-private — it cannot be misused by external packages. If it ever needs to
  become shared (e.g. a `CachedProvider` in TASK-0009 makes its own HTTP calls), promote to an
  internal package rather than exporting it from the zerodha package.

## Related decisions

- [Zerodha auth strategy](./2026-04-07-zerodha-auth-strategy.md) — defines ErrAuthRequired and
  the token lifecycle that this mapping serves
- [Zerodha instrument token lookup](./2026-04-07-zerodha-instrument-token-lookup.md) — one of the
  three HTTP call sites that benefit from centralized auth error handling

## Revisit trigger

If TASK-0009 (CachedProvider) or a future provider needs to make its own HTTP calls with
different auth error semantics, extract `doHTTP` to `internal/kite/` or parameterize the
auth status codes.
