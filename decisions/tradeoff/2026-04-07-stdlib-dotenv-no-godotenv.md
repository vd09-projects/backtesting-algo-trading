# stdlib .env parser — no godotenv dependency

| Field    | Value      |
|----------|------------|
| Date     | 2026-04-07 |
| Status   | accepted   |
| Category | tradeoff   |
| Tags     | dependencies, dotenv, credentials, prototype, stdlib, convention |

## Context

`cmd/authtest` needs to read `KITE_API_KEY` and `KITE_API_SECRET` from a `.env` file so the
developer doesn't have to export environment variables manually on every shell session. The
standard Go ecosystem approach is to import `github.com/joho/godotenv`.

CLAUDE.md explicitly requires: **"Do not add new dependencies without discussion. Keep the
dependency footprint minimal."**

## Options considered

### Option A — `github.com/joho/godotenv` (external dependency)
- **Pros:** Handles edge cases (quoted values, multi-line values, escape sequences, variable
  expansion). Battle-tested.
- **Cons:** Adds an external dependency for a ~20-line problem. The prototype is throwaway
  code — even `godotenv`'s full feature set would go unused. Violates the repo's dependency
  footprint rule without justification.

### Option B — stdlib `bufio.Scanner` + `strings.Cut` (chosen)
- **Pros:** Zero new dependencies. Covers 100% of the actual use case: `KEY=value` lines,
  blank lines, and `#` comments. Lines with quoted values or escape sequences are not used in
  this project's `.env` file. Real env vars take precedence (one `os.Getenv` check).
- **Cons:** Does not handle quoted values, multi-line values, or variable expansion — none of
  which this project needs. If `.env` format requirements grow, would need to revisit.

## Decision

Implement `loadDotEnv` in stdlib (~25 lines). The implementation in `cmd/authtest/main.go`
handles:
- Blank lines and `#` comment lines → skipped
- `KEY=value` → `os.Setenv(key, value)` if key not already set
- Missing `.env` file → silently skipped (optional by design)

No external dependency added.

## Consequences

- `.env` files with quoted values (`KEY="value with spaces"`) will be parsed incorrectly —
  the quotes become part of the value. This is a known limitation; document it in `.env.example`.
- If `cmd/authtest` is ever promoted to production code or the Zerodha provider needs `.env`
  loading at startup, reconsider `godotenv` at that point.
- `loadDotEnv` is local to `cmd/authtest/main.go` — it is NOT a shared utility. If another
  package needs `.env` loading, that decision is made independently.

## Revisit trigger

If any `.env` value legitimately needs quoted strings, spaces, or escape sequences, import
`godotenv` at that time. The 25-line implementation is not worth maintaining if requirements
grow beyond simple `KEY=value`.
