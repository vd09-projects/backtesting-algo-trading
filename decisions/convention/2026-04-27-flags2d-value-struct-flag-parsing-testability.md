# flags2D value struct for flag parsing testability in cmd/sweep2d

| Field    | Value            |
|----------|------------------|
| Date     | 2026-04-27       |
| Status   | experimental     |
| Category | convention       |
| Tags     | cmd/sweep2d, flags, testability, value-struct, TASK-0044 |

## Context

`cmd/sweep2d/main.go` has 10 flag values that need validation before use (from/to dates, timeframe, strategy name, 4 param axis values). The validation logic needs to be testable without constructing a real `flag.FlagSet`. In `cmd/sweep`, the existing `parseAndValidateFlags` helper takes 7 individual parameters — acceptable for 7 args but unwieldy if the count grows.

## Options considered

### Option A: `flags2D` value struct passed to `parseAndValidateFlags2D` (chosen)

- **Pros**: All flag values are bundled into one named struct, making the validation function signature stable as flags are added. Tests construct a `flags2D` literal and mutate it — readable and compact. Avoids the 8+ parameter smell. Same pattern as `cmd/sweep` but using a struct instead of individual parameters.
- **Cons**: One additional type to define. Tests must import the unexported struct from the same `package main`.

### Option B: Individual parameters (status quo from cmd/sweep)

- **Pros**: No new type. Works fine for the 7-param case in cmd/sweep.
- **Cons**: `parseAndValidateFlags2D` would take 10 individual parameters — unreadable and brittle to add more.

### Option C: Accept `*flag.FlagSet` directly

- **Pros**: No wrapper struct needed.
- **Cons**: Tests must construct and configure a full `flag.FlagSet` — more ceremony. Couples the validation function to the flag package.

## Decision

`flags2D` struct groups the parsed string/float values for `parseAndValidateFlags2D`. `main()` fills the struct from `flag.*` dereferenced values and passes it to the validator. Tests construct a `flags2D` literal directly, mutate one field per case, and call `parseAndValidateFlags2D`. The struct is unexported; tests in `package main` can access it directly.

## Consequences

When new flags are added to `cmd/sweep2d`, they are added to `flags2D` and to the validator — one place for each concern. Test cases add a new `modify` closure, not a new function.

## Related decisions

- [cmd/sweep parseAndValidateFlags](../convention/2026-04-09-io-writer-in-config-for-stdout-testability.md) — the broader testability-via-injection pattern this follows.

## Revisit trigger

If `cmd/sweep` is ever refactored, consider applying the same `flags` struct pattern there for consistency.
