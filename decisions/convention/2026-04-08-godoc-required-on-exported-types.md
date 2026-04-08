# Godoc comments are required on all exported types and functions

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-08       |
| Status   | accepted         |
| Category | convention       |
| Tags     | godoc, naming, revive, convention, documentation, exported |

## Context

Go's `revive` linter (enabled in `.golangci.yml`) flags exported identifiers that lack a doc comment. The question arose when looking at the `Provider` type: is the comment above it required, or optional style?

## Decision

Required. Every exported type, function, method, and constant must have a godoc comment beginning with the identifier name. This is enforced by:

1. **`revive` linter** — `exported` rule fires on missing or malformed doc comments and fails CI.
2. **Go convention** — `go doc` and pkg.go.dev surface these comments as the public API documentation. Missing comments mean missing docs.

The comment format must start with the identifier name:
```go
// Provider implements provider.DataProvider using the Kite Connect API.
type Provider struct { ... }

// NewProvider creates a Provider and downloads the instruments CSV.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) { ... }
```

Unexported identifiers do not require comments (linter does not check them), but complex unexported logic should be commented for maintainability.

## Consequences

- Any new exported symbol without a doc comment will fail `golangci-lint run ./...` and block commit/CI.
- No exceptions — the linter is the enforcer, not code review.

## Related decisions

- [Types in pkg/provider/zerodha must not repeat the package name](../convention/2026-04-08-no-package-name-stutter-in-zerodha.md) — the `revive` linter enforces both this and the stutter rule.
