---
name: errcheck in validated closures
description: errcheck fires on `s, _ := constructor()` inside closures even when params were validated before the closure was returned — use explicit error check with panic and a message that identifies it as a post-validation invariant violation
type: feedback
---

When a factory closure calls a constructor with `s, _ := constructor(...)` after prior validation, errcheck still fires. The linter cannot reason about prior validation.

Fix: use explicit `if err != nil { panic(...) }` inside the closure, with a message like `"params validated at startup, unexpected error: %v"`. This satisfies errcheck and makes the invariant explicit to future readers.

**Why:** `s, _ :=` silently discards errors and trips errcheck. The panic is appropriate here because it represents a programming invariant violation (params validated at startup, constructor is deterministic), not a recoverable runtime error.

**How to apply:** Any time a factory closure needs to call a constructor that was already validated before the closure was built — use the explicit panic form, not `_`.
