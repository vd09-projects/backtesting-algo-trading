# nolint:gocritic for hugeParam on model.Candle in pkg/strategy functions

| Field    | Value            |
|----------|------------------|
| Date     | 2026-05-07       |
| Status   | experimental     |
| Category | convention       |
| Tags     | gocritic, nolint, model.Candle, value-receiver, hugeParam, convention, TASK-0078, pkg/strategy |

## Context

TASK-0078 added `IsSessionOpen(bar model.Candle) bool` to `pkg/strategy/session.go`. `golangci-lint` (via the `gocritic` linter) flagged `hugeParam: bar is heavy (96 bytes); consider passing it by pointer`. The same finding would apply to any `pkg/strategy/` function that accepts a `model.Candle` parameter by value.

The question: should `pkg/strategy/` functions accept `model.Candle` by pointer (`*model.Candle`) to satisfy the linter, or suppress the finding?

## Options considered

### Option A: Suppress with nolint (selected)

```go
func IsSessionOpen(bar model.Candle) bool { //nolint:gocritic // hugeParam: Candle is always passed by value throughout the codebase; pointer would be an inconsistency
```

- **Pros**: Consistent with all other code that uses `model.Candle` by value. No caller-side changes. No nil-dereference risk. Matches the existing nolint pattern in `model.Candle.Validate()`.
- **Cons**: Suppresses a linter finding that has genuine performance relevance.

### Option B: Accept `*model.Candle`

```go
func IsSessionOpen(bar *model.Candle) bool
```

- **Pros**: Eliminates the 96-byte copy per call.
- **Cons**: Introduces nil-dereference risk. Inconsistent with the entire rest of the codebase where `model.Candle` is always passed by value (engine event loop, analytics, strategies). A single pointer-accepting function in `pkg/strategy/` would require all callers to take the address of their `Candle` values, propagating the inconsistency outward.

## Decision

**Suppress with `//nolint:gocritic`**, using the same comment text as `model.Candle.Validate()`:

```
// hugeParam: Candle is always passed by value throughout the codebase; pointer would be an inconsistency
```

The existing `2026-04-06-value-semantics-for-domain-types.md` decision establishes that `model.Candle` uses value semantics codebase-wide, and that `gocritic hugeParam` is suppressed by convention for domain types. This decision extends that convention explicitly to `pkg/strategy/` function signatures — the prior decision covered method receivers; this covers function parameters.

The 96-byte copy is negligible at the call frequency involved (once per bar per strategy call). The consistency and nil-safety benefits outweigh the trivial copy cost.

## Consequences

- Any future `pkg/strategy/` function accepting `model.Candle` should apply the same nolint.
- The nolint comment must be kept descriptive (not bare `//nolint:gocritic`) so reviewers understand it is intentional and consistent, not lazy suppression.
- If `model.Candle` is ever refactored to a smaller struct (e.g., instrument identifier separated), the hugeParam finding may disappear naturally and the nolints can be removed.

## Related decisions

- [Value semantics for domain types](./2026-04-06-value-semantics-for-domain-types.md) — the parent convention this extends; covers method receivers on `Candle` and `Config`. This decision extends it to `pkg/strategy/` function parameters.
