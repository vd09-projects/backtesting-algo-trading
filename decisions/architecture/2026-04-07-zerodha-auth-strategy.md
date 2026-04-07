# Zerodha auth strategy — manual token paste for v1

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | architecture |
| Tags     | zerodha, auth, kite-connect, access-token, provider, session |

## Context

Kite Connect uses a three-step OAuth-like flow to obtain a daily `access_token`:

1. User opens a login URL in the browser:
   ```
   https://kite.zerodha.com/connect/login?v=3&api_key={api_key}
   ```
2. After login, Kite redirects to the app's registered `redirect_url` with a short-lived `request_token` in the query string. The `request_token` is valid for only a few minutes.
3. The app POSTs to `https://api.kite.trade/session/token` with:
   - `api_key`
   - `request_token`
   - `checksum` = `SHA-256(api_key + request_token + api_secret)` hex-encoded

   On success, the response includes an `access_token`.

The `access_token` expires at **6:00 AM IST the next day** (or earlier if the user logs out from Kite Web). The auth header for all subsequent API calls is:
```
Authorization: token {api_key}:{access_token}
X-Kite-Version: 3
```

This is a backtesting CLI tool — single developer, no persistent web server. The question is how to capture the `request_token` without a live redirect endpoint.

## Options considered

### Option A — Manual paste (CLI prints URL, user copies token)

Flow:
1. CLI prints the login URL.
2. User opens it in a browser, logs in.
3. Kite redirects to the configured `redirect_url`. User copies the `request_token` value from the browser's address bar.
4. CLI reads `request_token` from stdin (prompted).
5. CLI exchanges it for `access_token` via `POST /session/token`.
6. `access_token` stored to `~/.config/backtest/token.json` with expiry timestamp.
7. On next run: load token from file. If not expired (before 6 AM IST today), skip login entirely.

- **Pros:** No local server required. Zero infrastructure. Simple to implement and debug. The one manual step per day is acceptable for a tool run during market hours or the prior evening.
- **Cons:** One manual copy-paste step each morning. `redirect_url` in the Kite app settings can be any URL (even `https://localhost` — the browser will show a "connection refused" page but the token is visible in the address bar).

### Option B — Local HTTP server captures redirect automatically

Flow:
1. CLI starts a temporary HTTP server on `localhost:{port}`.
2. `redirect_url` in Kite app settings: `http://localhost:{port}/callback`.
3. User opens the login URL; Kite redirects to the local server.
4. Server parses `request_token` from the query string, stores it, shuts down.
5. CLI proceeds with the exchange and token storage.

- **Pros:** Zero manual steps — fully automated after the browser login click.
- **Cons:** `redirect_url` must be `http://localhost:{port}/callback` — hardcoded in the Kite app settings. Port must be free. More code. The browser still opens; this only eliminates the copy-paste step.

## Decision

**Option A — manual paste for v1.**

This is a backtesting tool, not a production trading system. One manual step per day is a reasonable constraint. Option A keeps the auth implementation under ~60 lines, has no moving parts, and is easy to test (the token exchange can be unit-tested with a mock HTTP server; the login redirect step is inherently manual). Option B can be added later if the friction proves significant.

**Token persistence contract:**
- Storage path: `~/.config/backtest/token.json` (overridable via `BACKTEST_TOKEN_PATH` env var)
- Schema:
  ```json
  { "access_token": "...", "expires_at": "2026-04-08T00:30:00Z" }
  ```
- Expiry logic: token is valid if `time.Now().UTC()` is before `expires_at`. `expires_at` is set to **6:00 AM IST (00:30 UTC)** of the day after the token was issued.
- On expiry: re-run the full login flow.

**Auth is internal to the provider — not exposed through the `DataProvider` interface.** The Zerodha provider constructor (`NewZerodhaProvider`) handles token loading and refresh; callers only interact with `FetchCandles`.

## Consequences

- The `redirect_url` registered in the Kite developer console should be set to `https://localhost` (or any URL that won't serve a real page) — the user reads the token from the browser address bar.
- Token file must not be committed to git. Add `~/.config/backtest/` to gitignore or document that the path is outside the repo.
- If the token file is missing or expired at runtime, the provider returns a typed error (`ErrAuthRequired`) so the caller can prompt re-authentication.
- Daily token refresh is a manual step — no background refresh, no cron job.

## Related decisions

- [context.Context deferred from Run() and DataProvider interface](../tradeoff/2026-04-06-context-parameter-deferred.md) — context is already threaded through `FetchCandles`; cancellation will interrupt in-flight auth requests too

## Revisit trigger

If the tool is ever used in an automated pipeline (CI, scheduled runs) where a human cannot perform the daily login, switch to Option B (local HTTP server) or explore Zerodha's API key refresh token mechanism if one becomes available.
