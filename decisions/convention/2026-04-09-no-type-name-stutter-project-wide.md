# No type-name stutter — project-wide convention

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-09       |
| Status   | accepted         |
| Category | convention       |
| Tags     | naming, revive, convention, stutter, output, Config |

## Context

The `revive` linter's `exported` rule flags type names that repeat their package name — e.g., `output.OutputConfig` stutters because the package already provides the "output" context. This was first applied to `pkg/provider/zerodha` (renaming `ZerodhaProvider` → `Provider`). During the TASK-0010 quality review, `output.OutputConfig` was flagged and renamed to `output.Config`, making the same principle apply beyond a single package.

## Decision

No exported type or constructor in this repo may repeat its own package name. The rule is:

- `output.Config` ✓ — not `output.OutputConfig`
- `zerodha.Provider` ✓ — not `zerodha.ZerodhaProvider`
- `engine.Config` ✓ — not `engine.EngineConfig`

The `revive` linter enforces this automatically at CI time. No `//nolint` suppressions are permitted for this rule.

## Consequences

- When acceptance criteria or task descriptions name a type (e.g., "implement `OutputConfig`"), the actual name must still follow the no-stutter convention. The spec is describing intent, not the literal identifier.
- New packages must be designed with this in mind from the start — types named after the package's purpose rather than the package name.

## Related decisions

- [Types in pkg/provider/zerodha must not repeat the package name](./2026-04-08-no-package-name-stutter-in-zerodha.md) — the specific zerodha instance that established this pattern; this decision generalizes it.
