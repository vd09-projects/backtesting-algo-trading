---
name: analytics package — no JSON tags, no serialization concerns
description: internal/analytics is a pure computation layer; never add JSON tags or serialization to its types
type: feedback
---

internal/analytics types (KillSwitchThresholds, Report, etc.) must NOT receive JSON tags. The package is a pure computation layer — adding JSON concerns there violates its dependency-free design (confirmed in TASK-0048 build session, 2026-05-07).

**Why:** analytics imports only pkg/model, math, sort, time. Adding JSON tags would signal it is also a serialization type, blurring its responsibility. The prior decision (2026-04-21-kill-switch-analytics-to-montecarlo-boundary.md) established this pattern.

**How to apply:** When a cmd/ binary needs to read or write an analytics type to/from JSON, define a local DTO in the cmd/ package (e.g., thresholdsFile in cmd/monitor). Convert DTO ↔ analytics struct at the cmd boundary.
