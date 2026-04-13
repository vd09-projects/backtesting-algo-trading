# Function-parameter injection for testability — complement to Config injection

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-13       |
| Status   | accepted         |
| Category | convention       |
| Tags     | testability, dependency-injection, http.Client, function-parameters, LoginFlow, cmdutil, convention |

## Context

`LoginFlow` in `internal/cmdutil` hardcoded `http.DefaultClient` and `"https://api.kite.trade"`
when calling `zerodha.ExchangeToken`. This made the exchange and token-save paths untestable
without network access. During a quality review pass, `LoginFlow` showed 38.9% coverage because
the untested paths required injected dependencies.

The rest of the codebase already uses two injection patterns:
- **Config-level injection**: `Config.Sleep func(time.Duration)`, `Config.Stdout io.Writer` —
  used for long-lived structs where the dependency is set once.
- **Parameter-level injection**: `zerodha.ExchangeToken(ctx, client, baseURL, ...)` — used for
  pure functions that have no long-lived state.

`LoginFlow` is a stateless function. Config injection would require a new struct just for one
function. Parameter injection is the right fit.

## Options considered

### Option A: Inject `client *http.Client` and `baseURL string` as parameters (chosen)
- **Pros**: Consistent with `ExchangeToken` (which already accepted these); callers in production
  pass `http.DefaultClient, "https://api.kite.trade"` unchanged; tests can pass a `httptest.Server`
  client and URL directly; no new struct needed.
- **Cons**: Slightly longer function signature.

### Option B: Config-level injection (a `LoginFlowConfig` struct)
- **Pros**: Cleaner call site for production code (zero-value Config could default to live client).
- **Cons**: Overkill for a one-shot CLI function; creates a struct that's only ever used in one
  place; hides the injection behind a default-value mechanism.

### Option C: Leave hardcoded, accept untestable paths
- **Pros**: No change.
- **Cons**: 38.9% coverage on `LoginFlow`; can't unit-test exchange error handling or the
  save-failure warning path.

## Decision

Use function-parameter injection when:
1. The function is stateless (no long-lived struct).
2. The callee already accepts the injection (following an established pattern in the same codebase).
3. The injection point is an external I/O boundary (HTTP client, base URL, file path).

Config-level injection remains the pattern for long-lived structs (providers, engine). Function-
parameter injection is the pattern for stateless helpers and pure functions.

The rule for callers: production code always passes real dependencies explicitly
(`http.DefaultClient`, live URL). No "default client" magic — it's visible at the call site.

## Consequences

- `LoginFlow` signature changed from 4 args to 6. Both callers (`cmd/backtest`, `cmd/providertest`)
  updated.
- `LoginFlow` coverage moved from 38.9% to 94.4% — the exchange error, success+save, and
  success+save-failure paths are now exercised.
- Future functions in `cmdutil` (or any package) that call injectable external code should follow
  the same pattern rather than adding a new Config struct.

## Related decisions

- [Sleep injection via Config](2026-04-08-sleep-injection-via-config.md) — the long-lived struct
  counterpart to this pattern; use Config when the dependency is set once for a provider lifetime.
- [io.Writer field in Config for stdout testability](2026-04-09-io-writer-in-config-for-stdout-testability.md)
  — same principle for output.
