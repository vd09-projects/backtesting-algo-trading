# calcCommission side parameter: isBuy bool, not an enum

- **Date:** 2026-04-25
- **Status:** experimental
- **Category:** convention
- **Tags:** commission, calcCommission, bool, enum, TASK-0038
- **Task:** TASK-0038

## Decision

The side parameter added to `calcCommission` (unexported method on `*Portfolio`) is `isBuy bool`,
not a dedicated side enum or string constant.

## Rationale

The function has exactly two call sites and the domain has exactly two possible values: buy or sell.
An enum (e.g. `OrderSide`) would require a new exported type in `pkg/model`, godoc, a `String()`
method, and exhaustive switch coverage — all for a function that is unexported and called in two
places. The bool is clear at both call sites (`calcCommission(..., true)` for buy,
`calcCommission(..., false)` for sell) and costs nothing extra.

The rule of thumb: use a bool when there are exactly two states, neither state has domain-level
meaning that needs to be communicated to callers across package boundaries, and the function is not
part of any public interface. All three conditions hold here.

## Rejected alternatives

- **`OrderSide` enum** — correct for a public API with many call sites; overkill for an unexported
  two-call-site helper.
- **String constant `"buy"/"sell"`** — worse than bool: no exhaustive check, string comparison
  overhead, typo risk.

## Revisit trigger

If `calcCommission` becomes exported or gains a third call site with a different side semantic
(e.g. short-sell), migrate to an enum then.
