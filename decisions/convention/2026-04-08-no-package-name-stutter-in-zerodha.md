# Types in pkg/provider/zerodha must not repeat the package name

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-08       |
| Status   | accepted         |
| Category | convention       |
| Tags     | zerodha, naming, revive, convention, provider, stutter |

## Context

The Zerodha provider struct was originally named `ZerodhaProvider`, making it `zerodha.ZerodhaProvider` at call sites. The `revive` linter flagged this as a package-name stutter — a well-established Go anti-pattern where the package name already provides context, making the type prefix redundant and verbose.

The same pattern extends to the constructor (`NewZerodhaProvider`) and any future types added to this package.

## Options considered

### Option A: Keep `ZerodhaProvider` — suppress the linter
- **Pros**: No rename churn, explicit name at definition site.
- **Cons**: Violates Go naming conventions. Call sites read as `zerodha.ZerodhaProvider`, which is redundant. Revive correctly flags it — suppressing would mean suppressing a correct finding.

### Option B: Rename to `Provider` / `NewProvider`
- **Pros**: Idiomatic Go. Call sites read as `zerodha.Provider` and `zerodha.NewProvider` — clean, no redundancy. Consistent with how the standard library names types (`http.Client`, not `http.HTTPClient`).
- **Cons**: One-time rename cost. External callers (currently only `cmd/providertest`) must update.

## Decision

Renamed `ZerodhaProvider` → `Provider` and `NewZerodhaProvider` → `NewProvider`. The package name provides the "Zerodha" context. This is the Go standard — the package is the namespace.

**Convention going forward**: All exported types and constructors in `pkg/provider/zerodha` must follow this pattern. Do not prefix type names with "Zerodha". Examples:
- `zerodha.Provider` ✓ — not `zerodha.ZerodhaProvider`
- `zerodha.Config` ✓ — not `zerodha.ZerodhaConfig`
- `zerodha.NewProvider` ✓ — not `zerodha.NewZerodhaProvider`

The `revive` linter's `exported` check enforces this automatically.

## Consequences

- Any future types added to this package that inadvertently stutter will be caught by the linter at CI time.
- Existing decision files and documentation that reference `ZerodhaProvider` by name are technically stale but the concept is unchanged — the rename is cosmetic.

## Related decisions

- [`doHTTP` centralizes 401/403 → ErrAuthRequired](../architecture/2026-04-08-dohttp-centralizes-auth-errors.md) — established in the same package; same naming convention applies to that helper.

## Revisit trigger

If `pkg/provider/zerodha` is ever refactored into sub-packages, revisit whether `Provider` is still unambiguous enough as a name.
