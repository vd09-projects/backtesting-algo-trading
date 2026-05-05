---
name: omitempty pointer pattern for optional JSON blocks
description: Use *T pointer with omitempty for optional JSON objects, not float64 fields with omitempty, when zero is a valid value
type: feedback
---

When adding optional blocks to a JSON output struct, use a pointer-to-struct with `omitempty` rather than individual `float64` (or other numeric) fields with `omitempty`.

**Why:** `omitempty` on `float64` suppresses the field when value is `0.0`. For domain values like `SharpeP5`, `0.0` is a valid and meaningful result — it means the kill-switch threshold is exactly zero, not that the measurement is missing. Using a `*T` pointer makes absent vs zero-result unambiguous: `nil` pointer = block absent (not run), non-nil pointer = block present (ran, all fields serialized including zeros).

Established in TASK-0082 (bootstrap stats in output JSON): `*BootstrapStats json:"bootstrap,omitempty"` in `jsonResult`.

**How to apply:** Any time a new optional block is added to `jsonResult` or similar output structs, check whether any numeric fields could legitimately be zero. If yes, use pointer-to-struct. If all fields are string-only or guaranteed non-zero when present, top-level fields with omitempty are fine.
