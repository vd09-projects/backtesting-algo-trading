---
name: forvar loop variable capture Go 1.22+
description: Go 1.22+ fixes loop variable capture; `i := i` copies inside range loops are flagged by the forvar linter as unneeded and must be removed
type: feedback
---

In Go 1.22+, range loop variables are per-iteration (no longer shared). The `i := i` copy pattern that was required for goroutine closure capture in older Go is now flagged by the `forvar` linter.

**Why:** The copy was a workaround for a pre-1.22 Go semantics issue. Go 1.22 changed loop variable semantics — the copy is now both unnecessary and a lint violation.

**How to apply:** Remove all `i := i` (and similar) capture copies in range loops. This project uses go1.25.0, so the fix is always to delete the copy line.
