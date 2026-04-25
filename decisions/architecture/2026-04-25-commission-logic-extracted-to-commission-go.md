# Full-model commission logic extracted to commission.go, method stays on *Portfolio

- **Date:** 2026-04-25
- **Status:** experimental
- **Category:** architecture
- **Tags:** commission, file-organization, portfolio.go, commission.go, TASK-0038
- **Task:** TASK-0038

## Decision

The full NSE cost computation (`calcZerodhaFullCommission`) lives in a new file
`internal/engine/commission.go`. The `calcCommission` method itself remains on `*Portfolio` in
`portfolio.go`, but it delegates to the pure helper in `commission.go`. Both files are in the same
`engine` package.

## Rationale

`portfolio.go` was already long before TASK-0038. Adding ~80 lines of charge arithmetic inline
would have pushed it further toward unreadable. File extraction within the same package is the right
refactoring unit: it groups related logic (all commission math) without introducing a new package
boundary or changing any method ownership.

The `calcZerodhaFullCommission` function is a pure computation (no struct state needed) that takes
notional and side as inputs and returns a cost breakdown. Placing it in `commission.go` makes it
trivially testable via golden tests without any `Portfolio` setup. The method on `*Portfolio`
remains the dispatch point so no external callers change.

## Rejected alternatives

- **Inline in portfolio.go** — works but makes portfolio.go harder to read and the commission math
  harder to locate.
- **New `internal/engine/commission` sub-package** — adds package boundary overhead (exported
  names, import paths, godoc) for a single helper function. The same-package file split is
  sufficient.

## Consequences

Commission models for future exchange variants (MIS, BSE, etc.) should also live in
`commission.go`, not in `portfolio.go`. The file is the natural extension point.
