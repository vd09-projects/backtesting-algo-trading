# MIS STT rate co-located with CNC STT rate in commission.go constant block

- **Date:** 2026-04-27
- **Status:** experimental
- **Category:** architecture
- **Tags:** commission, MIS, STT, constants, commission.go, TASK-0047
- **Task:** TASK-0047

## Decision

`nseMISSTTRate` (0.025% — MIS intraday STT on sell leg only) is placed alongside `nseSTTRate`
(0.10% — CNC delivery STT on both legs) in the same constant block in
`internal/engine/commission.go`. The constants are not split into separate CNC and MIS sections.

## Rationale

The constant block in `commission.go` is organized by charge type: brokerage, STT, exchange
charges, SEBI charges, stamp duty, GST. Both `nseSTTRate` and `nseMISSTTRate` are STT variants —
grouping them together makes it trivially easy to compare the two rates and verify that the MIS
rate is correctly lower.

Splitting the block by commission model variant (one block for CNC constants, one for MIS) would
fragment conceptually related rates across the file. A reader wanting to know "what is the STT
difference between CNC and MIS?" would have to jump between two blocks; with co-location, the
answer is on adjacent lines.

This is consistent with the established decision (`2026-04-25-nse-charge-rates-as-unexported-constants`)
that NSE statutory rates are package-level constants grouped for readability alongside the
arithmetic that uses them.

## Rejected alternatives

- **Separate MIS constant block** — fragments the STT comparison across the file without adding
  clarity. A reader looking at the MIS arithmetic would have to locate a separate section to find
  the reference CNC rate.
- **Named constant groups using `iota` or struct** — no benefit for a small set of unrelated
  float64 values; adds syntactic overhead.

## Consequences

When a third STT variant is added (e.g. futures or options), it should also be placed in the same
constant block with a comment identifying which product type it applies to. If the constant block
grows unwieldy, consider introducing a named comment separator (`// --- STT rates ---`) rather than
splitting into separate `const` blocks.
