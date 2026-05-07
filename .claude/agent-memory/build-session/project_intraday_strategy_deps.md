---
name: Intraday strategy infrastructure status (2026-05-07)
description: Status of session-boundary utilities and what intraday strategies are still blocked on after TASK-0078
type: project
---

TASK-0078 (session-boundary utilities) is complete as of 2026-05-07. `pkg/strategy/session.go` now provides `IsSessionOpen` and `PreviousSessionClose` — the shared IST-timezone utilities both ORB and gap-and-go strategies need.

**Why:** Both TASK-0074 (ORB) and TASK-0075 (gap-and-go) were previously blocked on TASK-0078 in addition to Marcus design rules. Now all infrastructure deps are resolved.

**How to apply:** TASK-0074 and TASK-0075 are blocked solely on Marcus ruling on entry/exit rules. When `marcus-design` is invoked for either, the session-boundary utilities are ready — no preflight needed for that dependency. The primary decision files for the methodology are still pending in `decisions/algorithm/`.

Additional pkg/strategy pattern: `var ist = time.FixedZone("IST", 5*3600+30*60)` is the package-level IST timezone var — any new pkg/strategy function that needs IST should reference this var (unexported). The nolint pattern for `model.Candle` hugeParam is `//nolint:gocritic // hugeParam: Candle is always passed by value throughout the codebase; pointer would be an inconsistency`.
